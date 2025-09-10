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
	DBHost              *string `yaml:"db_host"`
	DBUseSSL            *string `yaml:"db_use_ssl"`
	FQDN                *string `yaml:"fqdn"`
	Email               *string `yaml:"email"`
	HTTPS               *bool   `yaml:"https"`
	DataDirectory       *string `yaml:"data_directory"`
	EnablePluginUploads *bool   `yaml:"enable_plugin_uploads"`
	EnableLocalMode     *bool   `yaml:"enable_local_mode"`
	ClientMaxBodySize   *string `yaml:"client_max_body_size"`

	NginxTemplate *string `yaml:"nginx_template,omitempty"`
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

	if c.DBHost == nil {
		c.DBHost = NewString("localhost")
	}

	if c.DBUseSSL == nil {
		c.DBUseSSL = NewString("disable")
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

	if c.ClientMaxBodySize == nil {
		c.ClientMaxBodySize = NewString("50M")
	}

	if c.NginxTemplate == nil {
		c.NginxTemplate = NewString("")
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

	if *c.DBUseSSL != "disable" && *c.DBUseSSL != "require" {
		return fmt.Errorf("db_use_ssl must be either 'disable' or 'require'")
	}

	if *c.HTTPS && (*c.FQDN == "" || *c.Email == "") {
		return fmt.Errorf("fqdn and email must be set if https is enabled")
	}

	if *c.DataDirectory == "" {
		return fmt.Errorf("data_directory cannot be empty")
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
