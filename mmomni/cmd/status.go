package cmd

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

func StatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "status",
		Short:   "Shows the Omnibus status",
		Long:    "Shows the status of all the Mattermost Omnibus components",
		Example: `  $ mmomni status`,
		Args:    cobra.NoArgs,
		Run:     statusCmdF,
	}
}

func getServiceStatus(name string) (string, int, error) {
	cmd := exec.Command("systemctl", "show", name, "--property", "SubState")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", 0, err
	}
	state := strings.Split(string(out), "=")[1]

	cmd = exec.Command("systemctl", "show", name, "--property", "MainPID")
	out, err = cmd.CombinedOutput()
	if err != nil {
		return "", 0, err
	}
	pidStr := strings.TrimSpace(strings.Split(string(out), "=")[1])
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return "", 0, err
	}

	return state, pid, nil
}

func statusCmdF(_ *cobra.Command, _ []string) {
	services := []string{"nginx", "postgresql@12-main", "mattermost"}

	for _, service := range services {
		state, pid, err := getServiceStatus(service)
		if err != nil {
			errAndExit(fmt.Errorf("error checking service %q status: %w", service, err))
		}

		fmt.Printf("[%d] %s: %s", pid, service, state)
	}
}
