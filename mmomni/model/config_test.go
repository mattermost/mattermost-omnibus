package model

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfigPreSave(t *testing.T) {
	t.Run("nginx_template pointer should be nil after PreSave if its empty string", func(t *testing.T) {
		baseConfig := &Config{
			NginxTemplate: NewString("/some/path.conf"),
		}

		cfg, err := baseConfig.PreSave()
		require.NoError(t, err)
		require.NotNil(t, cfg.NginxTemplate)
		require.Equal(t, *baseConfig.NginxTemplate, *cfg.NginxTemplate)

		baseConfig.NginxTemplate = NewString("")
		cfg, err = baseConfig.PreSave()
		require.NoError(t, err)
		require.Nil(t, cfg.NginxTemplate)
	})
}

func TestConfigIsValid(t *testing.T) {
	basicConfig := &Config{}
	basicConfig.SetDefaults()

	t.Run("Database User empty", func(t *testing.T) {
		config, err := basicConfig.Clone()
		require.NoError(t, err)
		config.DBUser = NewString("")

		require.EqualError(t, config.IsValid(), "database user cannot be empty")
	})

	t.Run("HTTPS enabled and FQDN not set", func(t *testing.T) {
		config, err := basicConfig.Clone()
		require.NoError(t, err)
		config.HTTPS = NewBool(true)
		config.Email = NewString("john_doe@example.com")

		require.EqualError(t, config.IsValid(), "fqdn and email must be set if https is enabled")
	})

	t.Run("HTTPS enabled and email not set", func(t *testing.T) {
		config, err := basicConfig.Clone()
		require.NoError(t, err)
		config.HTTPS = NewBool(true)
		config.FQDN = NewString("localhost")

		require.EqualError(t, config.IsValid(), "fqdn and email must be set if https is enabled")
	})

	t.Run("Empty data directory", func(t *testing.T) {
		config, err := basicConfig.Clone()
		require.NoError(t, err)
		config.DataDirectory = NewString("")

		require.EqualError(t, config.IsValid(), "data_directory cannot be empty")
	})

	t.Run("Monitoring enabled without being installed", func(t *testing.T) {
		config, err := basicConfig.Clone()
		require.NoError(t, err)
		config.MonitoringInstalled = NewBool(false)
		config.MonitoringEnabled = NewBool(true)

		require.EqualError(t, config.IsValid(), "monitoring_enabled cannot be true if monitoring_installed is false")
	})

	t.Run("Monitoring enabled without grafana user", func(t *testing.T) {
		config, err := basicConfig.Clone()
		require.NoError(t, err)
		config.MonitoringInstalled = NewBool(true)
		config.MonitoringEnabled = NewBool(true)
		config.GrafanaUser = NewString("root")

		require.EqualError(t, config.IsValid(), "grafana_user and grafana_password must be set if monitoring_enabled is set to true")
	})

	t.Run("Monitoring enabled without grafana password", func(t *testing.T) {
		config, err := basicConfig.Clone()
		require.NoError(t, err)
		config.MonitoringInstalled = NewBool(true)
		config.MonitoringEnabled = NewBool(true)
		config.GrafanaPassword = NewString("secret")

		require.EqualError(t, config.IsValid(), "grafana_user and grafana_password must be set if monitoring_enabled is set to true")
	})

	t.Run("Monitoring fully configured", func(t *testing.T) {
		config, err := basicConfig.Clone()
		require.NoError(t, err)
		config.MonitoringInstalled = NewBool(true)
		config.MonitoringEnabled = NewBool(true)
		config.GrafanaUser = NewString("root")
		config.GrafanaPassword = NewString("secret")

		require.NoError(t, config.IsValid())
	})

	t.Run("Jitsi enabled without being installed", func(t *testing.T) {
		config, err := basicConfig.Clone()
		require.NoError(t, err)
		config.JitsiInstalled = NewBool(false)
		config.JitsiEnabled = NewBool(true)

		require.EqualError(t, config.IsValid(), "jitsi_enabled cannot be true if jitsi_installed is false")
	})

	t.Run("Jitsi enabled without all secrets", func(t *testing.T) {
		config, err := basicConfig.Clone()
		require.NoError(t, err)
		config.JitsiInstalled = NewBool(true)
		config.JitsiEnabled = NewBool(true)
		config.JitsiFQDN = NewString("jitsi.localhost")
		config.JitsiJVBSecret = NewString("jvbsecret")

		require.EqualError(t, config.IsValid(), "jitsi_fqdn, jitsi_jvb_secret, jitsi_focus_secret and jitsi_focus_password must be set if jitsi_enabled is set to true")
	})

	t.Run("Jitsi enabled with the same FQDN as Mattermost", func(t *testing.T) {
		config, err := basicConfig.Clone()
		require.NoError(t, err)
		config.JitsiInstalled = NewBool(true)
		config.JitsiEnabled = NewBool(true)
		config.FQDN = NewString("localhost")
		config.JitsiFQDN = NewString("localhost")
		config.JitsiJVBSecret = NewString("jvbsecret")
		config.JitsiFocusSecret = NewString("focussecret")
		config.JitsiFocusPassword = NewString("focuspassword")

		require.EqualError(t, config.IsValid(), "fqdn and jitsi_fqdn cannot have the same value")
	})

	t.Run("Jitsi fully configured", func(t *testing.T) {
		config, err := basicConfig.Clone()
		require.NoError(t, err)
		config.JitsiInstalled = NewBool(true)
		config.JitsiEnabled = NewBool(true)
		config.JitsiFQDN = NewString("jitsi.localhost")
		config.JitsiJVBSecret = NewString("jvbsecret")
		config.JitsiFocusSecret = NewString("focussecret")
		config.JitsiFocusPassword = NewString("focuspassword")

		require.NoError(t, config.IsValid())
	})
}
