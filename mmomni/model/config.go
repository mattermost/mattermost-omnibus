package model

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	CONFIGPATH      = "/etc/mattermost/mmomni.yml"
	DBUSER          = "mmuser"
	DATADIRECTORY   = "/var/opt/mattermost/data"
	AUTO_BACKUP_DIR = "/var/opt/mattermost/backups"
)

type Config struct {
	Path string `yaml:"-"`

	DBUser              *string `yaml:"db_user"`
	DBPassword          *string `yaml:"db_password"`
	FQDN                *string `yaml:"fqdn"`
	Email               *string `yaml:"email"`
	HTTPS               *bool   `yaml:"https"`
	DataDirectory       *string `yaml:"data_directory"`
	EnablePluginUploads *bool   `yaml:"enable_plugin_uploads"`
	EnableLocalMode     *bool   `yaml:"enable_local_mode"`

	NginxTemplate *string `yaml:"nginx_template,omitempty"`

	MonitoringInstalled *bool   `yaml:"monitoring_installed"`
	MonitoringEnabled   *bool   `yaml:"monitoring_enabled"`
	GrafanaUser         *string `yaml:"grafana_user"`
	GrafanaPassword     *string `yaml:"grafana_password"`

	JitsiInstalled     *bool   `yaml:"jitsi_installed"`
	JitsiEnabled       *bool   `yaml:"jitsi_enabled"`
	JitsiFQDN          *string `yaml:"jitsi_fqdn"`
	JitsiJVBSecret     *string `yaml:"jitsi_jvb_secret"`
	JitsiFocusSecret   *string `yaml:"jitsi_focus_secret"`
	JitsiFocusPassword *string `yaml:"jitsi_focus_password"`
}

func ReadConfig(path string) (*Config, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, err
	}

	fileBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	config := &Config{Path: path}
	if err := yaml.Unmarshal(fileBytes, config); err != nil {
		return nil, err
	}

	config.SetDefaults()
	if err := config.IsValid(); err != nil {
		return nil, err
	}

	return config, nil
}

func (c *Config) SetDefaults() {
	if c.DBUser == nil {
		c.DBUser = NewString(DBUSER)
	}

	if c.DBPassword == nil {
		c.DBPassword = NewString("")
	}

	if c.FQDN == nil {
		c.FQDN = NewString("")
	}

	if c.Email == nil {
		c.Email = NewString("")
	}

	if c.HTTPS == nil {
		c.HTTPS = NewBool(false)
	}

	if c.DataDirectory == nil {
		c.DataDirectory = NewString(DATADIRECTORY)
	}

	if c.EnablePluginUploads == nil {
		c.EnablePluginUploads = NewBool(false)
	}

	if c.EnableLocalMode == nil {
		c.EnableLocalMode = NewBool(true)
	}

	if c.NginxTemplate == nil {
		c.NginxTemplate = NewString("")
	}

	// Monitoring
	if c.MonitoringInstalled == nil {
		c.MonitoringInstalled = NewBool(false)
	}

	if c.MonitoringEnabled == nil {
		c.MonitoringEnabled = NewBool(false)
	}

	if c.GrafanaUser == nil {
		c.GrafanaUser = NewString("")
	}

	if c.GrafanaPassword == nil {
		c.GrafanaPassword = NewString("")
	}

	// Jitsi
	if c.JitsiInstalled == nil {
		c.JitsiInstalled = NewBool(false)
	}

	if c.JitsiEnabled == nil {
		c.JitsiEnabled = NewBool(false)
	}

	if c.JitsiFQDN == nil {
		c.JitsiFQDN = NewString("")
	}

	if c.JitsiJVBSecret == nil {
		c.JitsiJVBSecret = NewString("")
	}

	if c.JitsiFocusSecret == nil {
		c.JitsiFocusSecret = NewString("")
	}

	if c.JitsiFocusPassword == nil {
		c.JitsiFocusPassword = NewString("")
	}
}

func (c *Config) Clone() (*Config, error) {
	configBytes, err := yaml.Marshal(c)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal config for cloning: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal([]byte(configBytes), &cfg); err != nil {
		return nil, fmt.Errorf("cannot unmarshal config for cloning: %w", err)
	}

	return &cfg, nil
}

// PreSave returns a cloned config that removes the pointers to values
// that can be ommited from the config file if empty and have a empty
// value
func (c *Config) PreSave() (*Config, error) {
	cfg, err := c.Clone()
	if err != nil {
		return nil, err
	}

	if *cfg.NginxTemplate == "" {
		cfg.NginxTemplate = nil
	}

	return cfg, nil
}

func (c *Config) IsValid() error {
	if *c.DBUser == "" {
		return fmt.Errorf("database user cannot be empty")
	}

	if *c.HTTPS && (*c.FQDN == "" || *c.Email == "") {
		return fmt.Errorf("fqdn and email must be set if https is enabled")
	}

	if *c.DataDirectory == "" {
		return fmt.Errorf("data_directory cannot be empty")
	}

	if *c.MonitoringEnabled && !*c.MonitoringInstalled {
		return fmt.Errorf("monitoring_enabled cannot be true if monitoring_installed is false")
	}

	if *c.MonitoringEnabled && (*c.GrafanaUser == "" || *c.GrafanaPassword == "") {
		return fmt.Errorf("grafana_user and grafana_password must be set if monitoring_enabled is set to true")
	}

	if *c.JitsiEnabled && !*c.JitsiInstalled {
		return fmt.Errorf("jitsi_enabled cannot be true if jitsi_installed is false")
	}

	if *c.JitsiEnabled && (*c.JitsiFQDN == "" || *c.JitsiJVBSecret == "" || *c.JitsiFocusSecret == "" || *c.JitsiFocusPassword == "") {
		return fmt.Errorf("jitsi_fqdn, jitsi_jvb_secret, jitsi_focus_secret and jitsi_focus_password must be set if jitsi_enabled is set to true")
	}

	if *c.JitsiEnabled && (*c.JitsiFQDN == *c.FQDN) {
		return fmt.Errorf("fqdn and jitsi_fqdn cannot have the same value")
	}

	return nil
}

func (c *Config) WriteToDisk() error {
	cfg, err := c.PreSave()
	if err != nil {
		return fmt.Errorf("cannot prepare config for writting to disk: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(c.Path), 0755); err != nil {
		return err
	}

	configBytes, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(c.Path, configBytes, 0640); err != nil {
		return err
	}

	return nil
}

func (c *Config) Save() error {
	if err := c.IsValid(); err != nil {
		return err
	}

	return c.WriteToDisk()
}
