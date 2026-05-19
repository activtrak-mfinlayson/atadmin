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

// configPaths returns the search paths for the config file.
func configPaths() []string {
	var paths []string
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		paths = append(paths, filepath.Join(xdg, "atadmin"))
	}
	paths = append(paths, filepath.Join(os.Getenv("HOME"), ".config", "atadmin"))
	paths = append(paths, filepath.Join(os.Getenv("HOME"), ".atadmin"))
	return paths
}

// newBaseViper creates a Viper instance loaded from disk.
func newBaseViper() (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	for _, p := range configPaths() {
		v.AddConfigPath(p)
	}
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("reading config file: %w", err)
		}
	}
	return v, nil
}

// Load loads the default profile (backward-compatible).
func Load() (*Config, error) {
	return LoadProfile("default")
}

// LoadProfile loads configuration for the named profile.
// Values may be overridden via ATADMIN_* environment variables.
// Returns an error listing available profiles when the profile is not found.
func LoadProfile(name string) (*Config, error) {
	base, err := newBaseViper()
	if err != nil {
		return nil, err
	}

	// Environment variable overrides always win.
	for _, pair := range []struct{ key, env string }{
		{"token", "ATADMIN_TOKEN"},
		{"base_url", "ATADMIN_BASE_URL"},
		{"format", "ATADMIN_FORMAT"},
		{"timeout", "ATADMIN_TIMEOUT"},
	} {
		if err := base.BindEnv(pair.key, pair.env); err != nil {
			return nil, fmt.Errorf("binding %s env: %w", pair.key, err)
		}
	}

	prefix := "profiles." + name + "."

	// Determine whether a profiles section exists at all.
	profiles := base.GetStringMap("profiles")
	hasProfiles := len(profiles) > 0

	if hasProfiles {
		if _, ok := profiles[name]; !ok {
			names, _ := ListProfiles()
			return nil, fmt.Errorf("profile %q not found; available profiles: %v", name, names)
		}
	}

	token := coalesce(base.GetString("token"), base.GetString(prefix+"token"))
	baseURL := coalesce(base.GetString("base_url"), base.GetString(prefix+"base_url"), "https://api.activtrak.com")
	format := coalesce(base.GetString("format"), base.GetString(prefix+"format"), "table")
	timeoutStr := coalesce(base.GetString("timeout"), base.GetString(prefix+"timeout"), "30s")

	timeout, err := time.ParseDuration(timeoutStr)
	if err != nil {
		timeout = 30 * time.Second
	}

	return &Config{
		Token:   token,
		BaseURL: baseURL,
		Format:  format,
		Timeout: timeout,
	}, nil
}

// ListProfiles returns the names of all profiles defined in the config file.
func ListProfiles() ([]string, error) {
	base, err := newBaseViper()
	if err != nil {
		return nil, err
	}
	profiles := base.GetStringMap("profiles")
	names := make([]string, 0, len(profiles))
	for k := range profiles {
		names = append(names, k)
	}
	return names, nil
}

// SaveProfile writes cfg under profiles.<name> in the config file with 0600 permissions.
func SaveProfile(name string, cfg *Config) error {
	path, err := configFilePath()
	if err != nil {
		return err
	}

	base, err := newBaseViper()
	if err != nil {
		return err
	}

	base.Set("profiles."+name+".token", cfg.Token)
	base.Set("profiles."+name+".base_url", cfg.BaseURL)
	base.Set("profiles."+name+".format", cfg.Format)
	base.Set("profiles."+name+".timeout", cfg.Timeout.String())

	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}

	if err := base.WriteConfigAs(path); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	return os.Chmod(path, 0600)
}

// configFilePath returns the path to the config file, preferring XDG.
func configFilePath() (string, error) {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "atadmin", "config.yaml"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting home dir: %w", err)
	}
	return filepath.Join(home, ".config", "atadmin", "config.yaml"), nil
}

func coalesce(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
