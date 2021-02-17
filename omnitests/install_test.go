package main

import (
	"io/ioutil"
	"os"
	"path/filepath"

	mmoModel "github.com/mattermost/mattermost-omnibus/mmomni/model"
	mmModel "github.com/mattermost/mattermost-server/v5/model"
)

func (s *OmnibusTestSuite) TestInstall() {
	s.Run("Should install latest version", func() {
		s.installOmnibus(s.fqdn, DEFAULT_EMAIL)
	})

	var client *mmModel.Client4
	var team *mmModel.Team

	s.Run("Should create the initial user and team", func() {
		s.createInitialUser("john_doe@example.com", "john_doe", "Sys@dmin-sample1")
		client = s.client("john_doe", "Sys@dmin-sample1")

		var resp *mmModel.Response
		team, resp = client.CreateTeam(&mmModel.Team{DisplayName: "Test Team", Name: "test-team", Type: "O"})
		s.Require().Nil(resp.Error)
		s.Require().Equal("Test Team", team.DisplayName)

		teams, resp := client.GetAllTeams("", 0, 10)
		s.Require().Nil(resp.Error)
		s.Require().Len(teams, 1)
	})

	s.Run("Should be able to post in the default channel", func() {
		defaultChannelName := "town-square"
		channel, resp := client.GetChannelByName(defaultChannelName, team.Id, "")
		s.Require().Nil(resp.Error)
		s.Require().Equal(defaultChannelName, channel.Name)

		_, resp = client.CreatePost(&mmModel.Post{ChannelId: channel.Id, Message: "Hello world!"})
		s.Require().Nil(resp.Error)
	})

	s.Run("Should be able to reconfigure Omnibus with mmomni", func() {
		var config *mmoModel.Config
		var serverConfig *mmModel.Config
		var resp *mmModel.Response

		s.Run("Get both server and Omnibus configurations and check that both are equal", func() {
			config = s.config()
			serverConfig, resp = client.GetConfig()
			s.Require().Nil(resp.Error)
			s.Require().Equal(*config.EnablePluginUploads, *serverConfig.PluginSettings.EnableUploads)
			s.Require().False(*config.EnablePluginUploads)
		})

		s.Run("Update enable plugin uploads value, save configuration and reconfigure", func() {
			config.EnablePluginUploads = mmoModel.NewBool(true)
			s.Require().NoError(config.Save())
			s.reconfigure()
		})

		s.Run("Get again both configurations, check that they are equal and correspond to the new value", func() {
			config = s.config()
			serverConfig, resp = client.GetConfig()
			s.Require().Nil(resp.Error)

			s.Require().Equal(*config.EnablePluginUploads, *serverConfig.PluginSettings.EnableUploads)
			s.Require().True(*config.EnablePluginUploads)
		})
	})

	s.Run("Should not allow to modify config variables that are set in mmomni.yml", func() {
		var config *mmoModel.Config
		var serverConfig *mmModel.Config
		var resp *mmModel.Response
		var initialValue bool

		s.Run("Get initial value from the mmomni config and assert that it corresponds to the server value", func() {
			config = s.config()
			initialValue = *config.EnablePluginUploads

			serverConfig, resp = client.GetConfig()
			s.Require().Nil(resp.Error)
			s.Require().Equal(initialValue, *serverConfig.PluginSettings.EnableUploads)
		})

		s.Run("Update the configuration value through the API", func() {
			serverConfig.PluginSettings.EnableUploads = mmModel.NewBool(!initialValue)
			newConfig, resp := client.PatchConfig(serverConfig)
			s.Require().Nil(resp.Error)
			s.Require().Equal(initialValue, *newConfig.PluginSettings.EnableUploads)
		})

		s.Run("Get config from the server and check that the value hasn't changed", func() {
			serverConfig, resp = client.GetConfig()
			s.Require().Nil(resp.Error)
			s.Require().Equal(initialValue, *serverConfig.PluginSettings.EnableUploads)
		})
	})

	s.Run("Should be able to purge the package", func() {
		s.purgeOmnibus()
	})
}

func (s *OmnibusTestSuite) TestUpdate() {
	s.Run("Install a previous Omnibus version", func() {
		s.installOmnibus("localhost", DEFAULT_EMAIL, "5.28.0-0")
		s.checkAPIVersion("5.28.0")
	})

	s.Run("Ensure backup directory doesn't exist", func() {
		s.Require().NoError(os.RemoveAll(mmoModel.AUTO_BACKUP_DIR))
	})

	s.Run("Update to a newer Omnibus version", func() {
		s.run("apt-get install -y mattermost=5.28.1-0 mattermost-omnibus=5.28.1-0")
		s.checkURL(s.URL(), 200, "<title>Mattermost</title>")
		s.checkAPIVersion("5.28.1")
	})

	s.Run("Check that an automated backup was created during the update", func() {
		s.directoryExists(mmoModel.AUTO_BACKUP_DIR)
		files, err := ioutil.ReadDir(mmoModel.AUTO_BACKUP_DIR)
		s.Require().NoError(err)
		s.Require().Len(files, 1)
		s.Require().Contains(files[0].Name(), "mmobackup_")
	})
}

func (s *OmnibusTestSuite) TestReinstall() {
	s.Run("Install the last version of Omnibus", func() {
		s.installOmnibus(s.fqdn, DEFAULT_EMAIL)
	})

	s.Run("Remove Omnibus", func() {
		s.removeOmnibus()
	})

	s.Run("Install Omnibus again", func() {
		s.installOmnibus(s.fqdn, DEFAULT_EMAIL)
	})

	s.Run("Run reconfigure to ensure the process works", func() {
		s.reconfigure()
		s.checkURL("http://localhost", 200)
	})
}

func (s *OmnibusTestSuite) TestBackupAndRestore() {
	var dataDir string
	var backup string
	pluginID := "com.mattermost.wrangler"
	pluginURL := "https://github.com/gabrieljackson/mattermost-plugin-wrangler/releases/download/v0.6.0/com.mattermost.wrangler-0.6.0.tar.gz"

	s.Run("Install the last version of Omnibus and enable plugin uploads", func() {
		s.installOmnibus(s.fqdn, DEFAULT_EMAIL)

		config := s.config()
		dataDir = *config.DataDirectory
		config.EnablePluginUploads = mmoModel.NewBool(true)
		s.Require().NoError(config.Save())
		s.reconfigure()
	})

	s.Run("Create a set of test data", func() {
		s.createInitialUser("john_doe@example.com", "john_doe", "Sys@dmin-sample1")
		client := s.client("john_doe", "Sys@dmin-sample1")

		team, resp := client.CreateTeam(&mmModel.Team{DisplayName: "Test Team", Name: "test-team", Type: "O"})
		s.Require().Nil(resp.Error)

		channel, resp := client.GetChannelByName("town-square", team.Id, "")
		s.Require().Nil(resp.Error)

		_, resp = client.CreatePost(&mmModel.Post{ChannelId: channel.Id, Message: "Hello world!"})
		s.Require().Nil(resp.Error)

		_, resp = client.InstallPluginFromUrl(pluginURL, false)
		s.Require().Nil(resp.Error)
	})

	s.Run("Take a backup of the server's current state", func() {
		backup = s.backup()
	})
	defer os.Remove(backup)

	s.Run("Reset server state", func() {
		s.reset()
		s.pathNotExists(dataDir)
	})

	s.Run("Install mattermost again", func() {
		s.installOmnibus(s.fqdn, DEFAULT_EMAIL)
	})

	s.Run("Restore the backup and reconfigure", func() {
		s.restore(backup)
		s.reconfigure()
	})

	s.Run("Ensure that the original test data has been correctly restored", func() {
		var team *mmModel.Team
		var channel *mmModel.Channel
		var resp *mmModel.Response

		client := s.client("john_doe", "Sys@dmin-sample1")

		s.Run("Check team", func() {
			teams, resp := client.GetAllTeams("", 0, 10)
			s.Require().Nil(resp.Error)
			s.Require().Len(teams, 1)
			team = teams[0]
			s.Require().Equal("test-team", team.Name)
		})

		s.Run("Check channel", func() {
			channel, resp = client.GetChannelByName("town-square", team.Id, "")
			s.Require().Nil(resp.Error)
			s.Require().Equal("town-square", channel.Name)
		})

		s.Run("Check post", func() {
			postList, resp := client.GetPostsForChannel(channel.Id, 0, 10, "")
			s.Require().Nil(resp.Error)
			posts := postList.ToSlice()
			s.Require().Equal("Hello world!", posts[0].Message)
		})

		s.Run("Check plugin", func() {
			found := false
			pluginStatuses, resp := client.GetPluginStatuses()
			s.Require().Nil(resp.Error)
			for _, p := range pluginStatuses {
				if p.PluginId == pluginID {
					found = true
				}
			}
			s.Require().True(found, "Wrangler plugin could not be found after restoring the backup")
		})
	})
}

func (s *OmnibusTestSuite) TestCustomNginxTemplate() {
	comment := "# test comment for custom template"
	nginxConfigPath := "/etc/nginx/conf.d/mattermost.conf"
	nginxTemplatePath := "/opt/mattermost/mmomni/ansible/playbooks/mattermost.conf"

	// install omnibus
	s.installOmnibus(s.fqdn, DEFAULT_EMAIL)

	// ensure that the comment is not present
	s.fileNotContains(nginxConfigPath, comment)

	// copy the nginx template to a temporary location and modify it adding a comment
	tmpDir, err := ioutil.TempDir("", "omnitests_*")
	s.Require().NoError(err)
	defer os.RemoveAll(tmpDir)
	modifiedTemplatePath := filepath.Join(tmpDir, "mattermost_modified.conf")

	fileBytes, err := ioutil.ReadFile(nginxTemplatePath)
	s.Require().NoError(err)

	// append comment to the file
	err = ioutil.WriteFile(modifiedTemplatePath, []byte(comment+"\n"+string(fileBytes)), 0755)
	s.Require().NoError(err)

	// assert that the template now contains the comment
	s.fileContains(modifiedTemplatePath, comment)

	// update the configuration to point to the custom template
	config := s.config()
	config.NginxTemplate = mmModel.NewString(modifiedTemplatePath)
	s.Require().NoError(config.Save())

	// reconfigure to use the new template
	s.reconfigure()

	// assert the configuration file is using the new template
	s.fileContains(nginxConfigPath, comment)
}
