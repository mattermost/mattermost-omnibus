package main

import (
	"github.com/mattermost/mattermost-omnibus/mmomni/model"
)

func (s *OmnibusTestSuite) checkJitsiIsInstalled() {
	s.checkPackageIsInstalled("jitsi-meet-web")
	s.checkPackageIsInstalled("jitsi-meet-prosody")
	s.checkPackageIsInstalled("jicofo")
	s.checkPackageIsInstalled("coturn")
	s.checkPackageIsInstalled("dnsutils")
	s.checkPackageIsInstalled("jitsi-videobridge2")
}

func (s *OmnibusTestSuite) checkJitsiIsNotInstalled() {
	s.checkPackageIsNotInstalled("jitsi-meet-web")
	s.checkPackageIsNotInstalled("jitsi-meet-prosody")
	s.checkPackageIsNotInstalled("jicofo")
	s.checkPackageIsNotInstalled("coturn")
	s.checkPackageIsNotInstalled("dnsutils")
	s.checkPackageIsNotInstalled("jitsi-videobridge2")
}

func (s *OmnibusTestSuite) TestJitsiLifecycle() {
	jitsiFQDN := "jitsi.localhost"

	s.Run("Should be able to install Jitsi", func() {
		s.installOmnibus(s.fqdn, DEFAULT_EMAIL)

		config := s.config()
		config.JitsiInstalled = model.NewBool(true)
		s.Require().NoError(config.Save())
		s.reconfigure()

		s.Run("Jitsi should not respond yet as it's not yet enabled", func() {
			s.checkURL("http://"+jitsiFQDN, 200, "<title>Mattermost</title>")
		})

		s.Run("Jitsi packages should have been installed", func() {
			s.checkJitsiIsInstalled()
		})
	})

	s.Run("Should be able to reconfigure after installing Jitsi", func() {
		s.reconfigure()
	})

	s.Run("Should be able to enable Jitsi", func() {
		config := s.config()
		config.JitsiEnabled = model.NewBool(true)
		config.JitsiFQDN = model.NewString(jitsiFQDN)
		config.JitsiJVBSecret = model.NewString("JVBSecret")
		config.JitsiFocusSecret = model.NewString("FocusSecret")
		config.JitsiFocusPassword = model.NewString("FocusPassword")
		s.Require().NoError(config.Save())
		s.reconfigure()

		s.Run("Jitsi should respond now that is enabled", func() {
			s.checkURL("http://"+jitsiFQDN, 200, "<title>Jitsi Meet</title>")
		})
	})

	s.Run("Should be able to reconfigure after enabling Jitsi", func() {
		s.reconfigure()
	})

	s.Run("Should be able to disable Jitsi", func() {
		config := s.config()
		config.JitsiEnabled = model.NewBool(false)
		s.Require().NoError(config.Save())
		s.reconfigure()

		s.Run("Jitsi should not respond now that is disabled", func() {
			s.checkURL("http://"+jitsiFQDN, 200, "<title>Mattermost</title>")
		})

		s.Run("Jitsi packages should still be installed", func() {
			s.checkJitsiIsInstalled()
		})
	})

	s.Run("Should be able to reconfigure after disabling Jitsi", func() {
		s.reconfigure()
	})

	s.Run("Should be able to uninstall Jitsi", func() {
		config := s.config()
		config.JitsiInstalled = model.NewBool(false)
		s.Require().NoError(config.Save())
		s.reconfigure()

		s.Run("Jitsi should now have been removed", func() {
			s.checkJitsiIsNotInstalled()
		})
	})

	s.Run("Should be able to reconfigure after uninstalling Jitsi", func() {
		s.reconfigure()
	})
}
