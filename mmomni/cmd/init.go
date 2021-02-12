package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/mattermost/mattermost-omnibus/mmomni/model"
)

func InitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "init",
		Hidden: true,
		Short:  "Initializes the Mattermost Omnibus config file",
		Long: `Initializes the Mattermos Omnibus config file in the /etc/mattermost directory and populates it with the initial configuration

This command is expected to be called from the package installation scripts, not intended to be run by the user`,
		Example: `  # most of the time init will be invoked with fqdn and email
  $ mmomni init --fqdn my.domain.com --email contact@example.com

  # https can be disabled if the server is not reachable from the internet,
  # so the SSL certificate cannot be generated
  $ mmomni init --fqdn my.domain.com --email contact@example.com --https false`,
		Args: cobra.NoArgs,
		Run:  initCmdF,
	}

	cmd.Flags().String("fqdn", "", "Mattermost domain name")
	_ = cmd.MarkFlagRequired("fqdn")
	cmd.Flags().String("email", "", "Letsencrypt contact email address")
	_ = cmd.MarkFlagRequired("email")
	cmd.Flags().Bool("https", true, "Enable to configure the SSL certificate")
	_ = viper.BindPFlag("https", cmd.Flags().Lookup("https"))

	return cmd
}

func initCmdF(cmd *cobra.Command, _ []string) {
	fqdn, _ := cmd.Flags().GetString("fqdn")
	email, _ := cmd.Flags().GetString("email")
	https := viper.GetBool("https")

	var config *model.Config
	if _, err := os.Stat(model.CONFIGPATH); !os.IsNotExist(err) {
		var err error
		config, err = model.ReadConfig(model.CONFIGPATH)
		if err != nil {
			errAndExit(fmt.Errorf("error reading configuration file in %q: %w", model.CONFIGPATH, err))
		}
	} else {
		config = &model.Config{Path: model.CONFIGPATH}
		config.SetDefaults()
		config.DBPassword = model.NewString(CreatePGPassword())
		config.HTTPS = model.NewBool(https)
	}

	config.FQDN = model.NewString(ParseFQDN(fqdn))
	config.Email = model.NewString(email)
	config.EnableLocalMode = model.NewBool(true)

	// when initializing the configuration, we save it without
	// validating it, so the file gets written to disk and the user
	// can change it later
	if err := config.WriteToDisk(); err != nil {
		errAndExit(fmt.Errorf("error saving configuration: %w", err))
	}

	fmt.Printf("config file %q successfully saved\n", model.CONFIGPATH)
}
