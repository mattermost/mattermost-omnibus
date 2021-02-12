package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/mattermost/mattermost-omnibus/mmomni/model"
)

func ReconfigureCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "reconfigure",
		Short: "Reconfigures Mattermost Omnibus",
		Long: `Generates and applies a Mattermost Omnibus configuration from the Omnibus configuration file

This command should be run after modifying the /etc/mattermost/mmomni.yml configuration file to apply its changes and restart the platform`,
		Example: `  $ mmomni reconfigure`,
		Args:    cobra.NoArgs,
		Run:     reconfigureCmdF,
	}
}

func reconfigureCmdF(_ *cobra.Command, _ []string) {
	// we read the config from disk to validate it
	config, err := model.ReadConfig(model.CONFIGPATH)
	if err != nil {
		errAndExit(fmt.Errorf("error reading config at %q: %w", model.CONFIGPATH, err))
	}

	// and we save it before running reconfigure in case some defaults
	// using during validation needed to be written
	if err := config.Save(); err != nil {
		errAndExit(fmt.Errorf("error updating configuration at %q: %w", model.CONFIGPATH, err))
	}

	ansibleCmd := exec.Command("ansible-playbook", "/opt/mattermost/mmomni/ansible/playbooks/reconfigure.yml")
	ansibleCmd.Stdout = os.Stdout
	ansibleCmd.Stderr = os.Stderr
	ansibleCmd.Env = append(os.Environ(), "ANSIBLE_LIBRARY=/opt/mattermost/mmomni/ansible/modules")
	if err := ansibleCmd.Run(); err != nil {
		errAndExit(fmt.Errorf("error running reconfigure: %w", err))
	}
}
