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

	"github.com/spf13/cobra"

	"github.com/mattermost/mattermost-omnibus/mmomni/model"
)

func RestoreCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "restore",
		Short: "Restores a backup",
		Long: `Restores a backup, replacing the existing database information and data directory contents

To successfully restore a backup into a new installation of Omnibus, first we need to install the Omnibus package and run an initial configuration successfully, and then we can run the restore command. Restoring the backup will reuse the database, replacing its contents with the ones from the backup`,
		Example: `  $ mmomni restore my-backup-file.tgz`,
		Args:    cobra.ExactArgs(1),
		Run:     restoreCmdF,
	}
}

func restoreCmdF(cmd *cobra.Command, args []string) {
	backupFile := args[0]
	tarball, err := os.Open(backupFile)
	if err != nil {
		errAndExit(fmt.Errorf("error opening backup file %q: %w", backupFile, err))
	}
	defer tarball.Close()

	gr, err := gzip.NewReader(tarball)
	if err != nil {
		errAndExit(fmt.Errorf("cannot open tarball: %w ", err))
	}
	defer gr.Close()

	tr := tar.NewReader(gr)

	dir, err := ioutil.TempDir(os.TempDir(), "mmomni_")
	if err != nil {
		errAndExit(fmt.Errorf("cannot create temporal directory: %w", err))
	}
	defer os.RemoveAll(dir)

	// Uncompress tarball in the temp directory
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			errAndExit(fmt.Errorf("cannot read entry from the tarball: %w", err))
		}

		switch header.Typeflag {
		case tar.TypeDir:
			destDir := filepath.Join(dir, header.Name)
			if err := os.MkdirAll(destDir, os.FileMode(header.Mode)); err != nil {
				errAndExit(fmt.Errorf("cannot create directory path %q: %w", destDir, err))
			}
		case tar.TypeReg:
			destDir := filepath.Join(dir, filepath.Dir(header.Name))
			if _, err := os.Stat(destDir); os.IsNotExist(err) {
				if err := os.MkdirAll(destDir, 0750); err != nil {
					errAndExit(fmt.Errorf("cannot create directory path %q: %w", destDir, err))
				}
			} else if err != nil {
				errAndExit(fmt.Errorf("cannot stat the directory %q: %w", destDir, err))
			}

			destFile := filepath.Join(dir, header.Name)
			file, err := os.Create(destFile)
			if err != nil {
				errAndExit(fmt.Errorf("cannot create file %q: %w", destFile, err))
			}
			if err := file.Chmod(os.FileMode(header.Mode)); err != nil {
				errAndExit(fmt.Errorf("cannot change permissions for file %q: %w", destFile, err))
			}

			if _, err := io.Copy(file, tr); err != nil {
				errAndExit(fmt.Errorf("cannot copy contents to the file %q: %w", destFile, err))
			}

			file.Close()
		default:
			errAndExit(fmt.Errorf("unknown file %q with type %q found in the tarball", header.Name, header.Typeflag))
		}
	}

	fmt.Printf("Backup extracted into temporal directory %q\n", dir)

	oldConfig, err := model.ReadConfig(model.CONFIGPATH)
	if err != nil {
		errAndExit(fmt.Errorf("error reading existing Omnibus configuration at %q: %w", model.CONFIGPATH, err))
	}

	tmpConfigPath := filepath.Join(dir, filepath.Base(model.CONFIGPATH))
	config, err := model.ReadConfig(tmpConfigPath)
	if err != nil {
		errAndExit(fmt.Errorf("error reading extracted Omnibus configuration at %q: %w", tmpConfigPath, err))
	}

	config.DBUser = oldConfig.DBUser
	config.DBPassword = oldConfig.DBPassword
	// update the configuration in case we're restoring a backup from
	// an older Omnibus version
	config.SetDefaults()

	config.Path = model.CONFIGPATH
	if err := config.Save(); err != nil {
		errAndExit(fmt.Errorf("error restoring configuration: %w", err))
	}

	fmt.Printf("Configuration restored in %q\n", model.CONFIGPATH)

	// Import pgdump
	dumpFilePath := filepath.Join(dir, "database.dump")
	pgRestoreCmd := exec.Command("pg_restore", "-Fc", "-c", "-d", "mattermost", dumpFilePath, "-w", "-U", *config.DBUser, "-h", "localhost")
	pgRestoreCmd.Env = append(pgRestoreCmd.Env, "PGPASSWORD="+*config.DBPassword)
	pgRestoreCmd.Stdout = os.Stdout
	pgRestoreCmd.Stderr = os.Stderr
	if err := pgRestoreCmd.Run(); err != nil {
		errAndExit(fmt.Errorf("error restoring database backup %q: %w", dumpFilePath, err))
	}

	fmt.Printf("Database backup restored\n")

	// If data directory exists, move data directory to its final destination
	tmpDataDir := filepath.Join(dir, "data")
	if _, err := os.Stat(tmpDataDir); os.IsNotExist(err) {
		fmt.Println("Backup doesn't contain a data directory, skipping...")
	} else if err != nil {
		errAndExit(fmt.Errorf("error checking the data directory %q: %w", tmpDataDir, err))
	} else {
		if _, err := os.Stat(*config.DataDirectory); err != nil && !os.IsNotExist(err) {
			errAndExit(fmt.Errorf("error checking the data directory destination path %q: %w", *config.DataDirectory, err))
		} else if err == nil {
			if err := os.RemoveAll(*config.DataDirectory); err != nil {
				errAndExit(fmt.Errorf("error restoring data directory on %q: %w", *config.DataDirectory, err))
			}
		}

		if err := os.Rename(tmpDataDir, *config.DataDirectory); err != nil {
			errAndExit(fmt.Errorf("error restoring data directory on %q: %w", tmpDataDir, err))
		}

		fmt.Printf("Data directory restored in %q\n", *config.DataDirectory)
		fmt.Println("\nPlease run \"mmomni reconfigure\" to apply the restored configuration")
	}
}
