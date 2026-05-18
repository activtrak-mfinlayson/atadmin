// Package config provides configuration loading for atadmin.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration values for atadmin.
type Config struct {
	Token   string
	BaseURL string
	Format  string
	Timeout time.Duration
}

// Load initializes a fresh Viper instance and returns a populated Config.
// Config file is searched in XDG and legacy paths. Missing config files are
// not an error. Values may be overridden via ATADMIN_* environment variables.
func Load() (*Config, error) {
	v := viper.New()

	v.SetConfigName("config")
	v.SetConfigType("yaml")

	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		v.AddConfigPath(filepath.Join(xdg, "atadmin"))
	}
	v.AddConfigPath(filepath.Join(os.Getenv("HOME"), ".config", "atadmin"))
	v.AddConfigPath(filepath.Join(os.Getenv("HOME"), ".atadmin"))

	v.SetDefault("format", "table")
	v.SetDefault("base_url", "https://api.activtrak.com")
	v.SetDefault("timeout", "30s")

	if err := v.BindEnv("token", "ATADMIN_TOKEN"); err != nil {
		return nil, fmt.Errorf("binding token env: %w", err)
	}
	if err := v.BindEnv("base_url", "ATADMIN_BASE_URL"); err != nil {
		return nil, fmt.Errorf("binding base_url env: %w", err)
	}
	if err := v.BindEnv("format", "ATADMIN_FORMAT"); err != nil {
		return nil, fmt.Errorf("binding format env: %w", err)
	}
	if err := v.BindEnv("timeout", "ATADMIN_TIMEOUT"); err != nil {
		return nil, fmt.Errorf("binding timeout env: %w", err)
	}

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("reading config file: %w", err)
		}
	}

	timeout, err := time.ParseDuration(v.GetString("timeout"))
	if err != nil {
		timeout = 30 * time.Second
	}

	return &Config{
		Token:   v.GetString("token"),
		BaseURL: v.GetString("base_url"),
		Format:  v.GetString("format"),
		Timeout: timeout,
	}, nil
}
