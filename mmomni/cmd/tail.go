package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

func TailCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "tail",
		Short:   "Tails Omnibus components logs",
		Long:    "Shows the logs for all the Mattermost Omnibus components. To close the command, press CTRL-c",
		Example: `  $ mmomni tail`,
		Args:    cobra.NoArgs,
		Run:     tailCmdF,
	}
}

func tailCmdF(_ *cobra.Command, _ []string) {
	tailArgs := []string{"-f"}
	logDirs := []string{"/var/log/nginx", "/var/log/postgresql", "/var/log/mattermost"}

	for _, dir := range logDirs {
		logfiles, err := filepath.Glob(dir + "/*.log")
		if err != nil {
			errAndExit(fmt.Errorf("error expanding glob %q: %w", dir+"/*.log", err))
		}
		tailArgs = append(tailArgs, logfiles...)
	}

	tailCmd := exec.Command("tail", tailArgs...)
	tailCmd.Stdout = os.Stdout
	tailCmd.Stderr = os.Stderr

	if err := tailCmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error running tail: %s\n", err)
		os.Exit(1)
	}
}
