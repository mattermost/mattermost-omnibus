package cmd

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/mattermost/mattermost-omnibus/mmomni/model"
)

func BackupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backup",
		Short: "Creates a backup",
		Long:  "Creates a backup of the Mattermost Omnibus configuration files, the database and the data directory, which includes attachments, installed plugins, etc",
		Example: `  # if we don't provide an output path, mmomni will generate one using the current timestamp
  $ mmomni backup

  # we can specify that we only want to store the database and config files in the backup
  $ mmomni backup --dbonly --output my-custom-file.tgz

  # we can run as well an automatic backup, which only includes the database and stores
  # the resulting tarball in /var/opt/mattermost/backups
  $ mmomni backup --auto`,
		Args: cobra.NoArgs,
		Run:  backupCmdF,
	}

	cmd.Flags().StringP("output", "o", "", "The path of the backup file")
	cmd.Flags().BoolP("dbonly", "d", false, "Backup database only, excluding data directory")
	cmd.Flags().StringP("config", "c", model.CONFIGPATH, "The path of the configuration file")
	cmd.Flags().BoolP("auto", "a", false, "Run the automatic backup process")

	return cmd
}

func getAutoBackupPath(base string, t time.Time) string {
	return filepath.Join(
		base,
		"mmobackup_"+t.Format("20060102_150405")+".tgz",
	)
}

func backupCmdF(cmd *cobra.Command, _ []string) {
	output, _ := cmd.Flags().GetString("output")
	dbonly, _ := cmd.Flags().GetBool("dbonly")
	configPath, _ := cmd.Flags().GetString("config")
	auto, _ := cmd.Flags().GetBool("auto")

	config, err := model.ReadConfig(configPath)
	if err != nil {
		errAndExit(fmt.Errorf("error reading configuration file in %q: %w", configPath, err))
	}

	if auto {
		if err := os.MkdirAll(model.AUTO_BACKUP_DIR, 0700); err != nil {
			errAndExit(fmt.Errorf("error creating automatic backup path %q: %w", model.AUTO_BACKUP_DIR, err))
		}

		output = getAutoBackupPath(model.AUTO_BACKUP_DIR, time.Now())
		dbonly = true
	} else if output == "" {
		output = fmt.Sprintf("mmomni-backup_%s.tgz", time.Now().Format("200601021504"))
	}

	// Creates tmpdir
	dir, err := ioutil.TempDir(os.TempDir(), "mmomni_")
	if err != nil {
		errAndExit(fmt.Errorf("error creating temp directory: %w", err))
	}
	defer os.RemoveAll(dir)

	// Runs pgdump
	dumpFilepath := filepath.Join(dir, "database.dump")
	pgDumpCmd := exec.Command("pg_dump", "mattermost", "-Fc", "-f", dumpFilepath, "-w", "-U", *config.DBUser, "-h", "localhost")
	pgDumpCmd.Env = append(pgDumpCmd.Env, "PGPASSWORD="+*config.DBPassword)
	pgDumpCmd.Stdout = os.Stdout
	pgDumpCmd.Stderr = os.Stderr
	if err := pgDumpCmd.Run(); err != nil {
		errAndExit(fmt.Errorf("error running database backup command: %w", err))
	}

	// Copies config to tmp directory
	fileBytes, err := ioutil.ReadFile(configPath)
	if err != nil {
		errAndExit(fmt.Errorf("error reading configuration file in %q: %w", configPath, err))
	}

	configFilepath := filepath.Join(dir, filepath.Base(model.CONFIGPATH))
	if err := ioutil.WriteFile(configFilepath, fileBytes, 0600); err != nil {
		errAndExit(fmt.Errorf("error writing configuration to %q: %w", configFilepath, err))
	}

	// Creates tarball with config, dump and data folder
	tarball, err := os.Create(output)
	if err != nil {
		errAndExit(fmt.Errorf("error creating tarball file %q: %w", output, err))
	}
	defer tarball.Close()

	gw := gzip.NewWriter(tarball)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	// Adds basic files to the tarball's root path
	if err := addFilesToTarball(tw, []string{configFilepath, dumpFilepath}, ""); err != nil {
		errAndExit(fmt.Errorf("error adding files to tarball: %w", err))
	}

	// If datadir is included, adds its contents under the "data" directory
	if !dbonly {
		files, err := ioutil.ReadDir(*config.DataDirectory)
		if err != nil {
			errAndExit(fmt.Errorf("error listing files in data directory %q: %w", *config.DataDirectory, err))
		}

		datadirFiles := make([]string, len(files))
		for i, file := range files {
			datadirFiles[i] = filepath.Join(*config.DataDirectory, file.Name())
		}

		if err := addFilesToTarball(tw, datadirFiles, "data"); err != nil {
			errAndExit(fmt.Errorf("error adding data files to tarball: %w", err))
		}
	}

	fmt.Printf("Backup created at %q\n", output)
}

func addFilesToTarball(w *tar.Writer, filePaths []string, basePath string) error {
	for _, path := range filePaths {
		// this is broken into a separate function to allow the file handle to
		// be closed inside the loop, whilst still retaining auto-closing
		// utility of defer
		err := addFileToTarball(w, path, basePath)

		if err != nil {
			return err
		}
	}

	return nil
}

func addFileToTarball(w *tar.Writer, path string, basePath string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	if stat.IsDir() {
		files, err := ioutil.ReadDir(path)
		if err != nil {
			return err
		}

		dirFilePaths := make([]string, len(files))
		for i, file := range files {
			dirFilePaths[i] = filepath.Join(path, file.Name())
		}

		if err := addFilesToTarball(w, dirFilePaths, filepath.Join(basePath, stat.Name())); err != nil {
			return err
		}
	} else {
		header := &tar.Header{
			Name:    filepath.Join(basePath, stat.Name()),
			Size:    stat.Size(),
			Mode:    int64(stat.Mode()),
			ModTime: stat.ModTime(),
		}

		if err := w.WriteHeader(header); err != nil {
			return err
		}

		if _, err := io.Copy(w, file); err != nil {
			return err
		}
	}

	return nil
}
