package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func errAndExit(err error) {
	fmt.Fprintf(os.Stderr, "ERROR: %s\n", err.Error())
	os.Exit(1)
}

func RootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:              "mmomni",
		Short:            "Manage the Mattermost Omnibus platform",
		PersistentPreRun: rootPreCheck,
	}

	cmd.AddCommand(
		BackupCmd(),
		DocsCmd(),
		InitCmd(),
		ReconfigureCmd(),
		RestoreCmd(),
		StatusCmd(),
		TailCmd(),
	)

	return cmd
}

func rootPreCheck(_ *cobra.Command, _ []string) {
	if os.Getuid() != 0 {
		errAndExit(fmt.Errorf("this program needs to run as root"))
	}
}

func Execute() {
	viper.SetEnvPrefix("mmo")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	if err := RootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
