package main

import (
	"github.com/mattermost/mattermost-omnibus/mmomni/model"
)

func (s *OmnibusTestSuite) checkMonitoringIsInstalled() {
	s.checkPackageIsInstalled("prometheus")
	s.checkPackageIsInstalled("grafana")
}

func (s *OmnibusTestSuite) checkMonitoringIsNotInstalled() {
	s.checkPackageIsNotInstalled("prometheus")
	s.checkPackageIsNotInstalled("grafana")
}

func (s *OmnibusTestSuite) TestMonitoringLifecycle() {
	s.Run("Should be able to install Monitoring", func() {
		s.installOmnibus(s.fqdn, DEFAULT_EMAIL)

		config := s.config()
		config.MonitoringInstalled = model.NewBool(true)
		s.Require().NoError(config.Save())
		s.reconfigure()

		s.Run("Monitoring should not respond yet as it's not yet enabled", func() {
			s.checkURL(s.URL()+"/monitoring", 200, "<title>Mattermost</title>")
		})

		s.Run("Monitoring packages should have been installed", func() {
			s.checkMonitoringIsInstalled()
		})
	})

	s.Run("Should be able to reconfigure after installing Monitoring", func() {
		s.reconfigure()
	})

	s.Run("Should be able to enable Monitoring", func() {
		config := s.config()
		config.MonitoringEnabled = model.NewBool(true)
		config.GrafanaUser = model.NewString("root")
		config.GrafanaPassword = model.NewString("toor")
		s.Require().NoError(config.Save())
		s.reconfigure()

		s.Run("Assert that Monitoring responds now that it is enabled", func() {
			s.checkURL(s.URL()+"/monitoring", 200, "<title>Grafana</title>")
		})
	})

	s.Run("Should be able to reconfigure after enabling Monitoring", func() {
		s.reconfigure()
	})

	s.Run("Should be able to disable Monitoring", func() {
		config := s.config()
		config.MonitoringEnabled = model.NewBool(false)
		s.Require().NoError(config.Save())
		s.reconfigure()

		s.Run("Monitoring should not respond now that is disabled", func() {
			s.checkURL(s.URL()+"/monitoring", 200, "<title>Mattermost</title>")
		})

		s.Run("Monitoring packages should still be installed", func() {
			s.checkMonitoringIsInstalled()
		})
	})

	s.Run("Should be able to reconfigure after disabling Monitoring", func() {
		s.reconfigure()
	})

	s.Run("Should be able to uninstall Monitoring", func() {
		config := s.config()
		config.MonitoringInstalled = model.NewBool(false)
		s.Require().NoError(config.Save())
		s.reconfigure()

		s.Run("Monitoring should now have been removed", func() {
			s.checkMonitoringIsNotInstalled()
		})
	})

	s.Run("Should be able to reconfigure after uninstalling Monitoring", func() {
		s.reconfigure()
	})
}
