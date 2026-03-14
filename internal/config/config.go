package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"gopkg.in/yaml.v3"
)

// Config is the top-level snowctl configuration.
type Config struct {
	APIVersion     string            `yaml:"apiVersion"`
	CurrentContext string            `yaml:"current-context"`
	Contexts       []ContextConfig   `yaml:"contexts"`
	Defaults       DefaultsConfig    `yaml:"defaults,omitempty"`
	Aliases        map[string]string `yaml:"aliases,omitempty"`
}

// ContextConfig defines a single ServiceNow instance connection.
type ContextConfig struct {
	Name     string         `yaml:"name"`
	Instance string         `yaml:"instance"`
	Auth     AuthConfig     `yaml:"auth"`
	Defaults DefaultsConfig `yaml:"defaults,omitempty"`
}

// AuthConfig holds authentication configuration.
type AuthConfig struct {
	Type         string `yaml:"type"`                    // "basic" or "oauth"
	Username     string `yaml:"username,omitempty"`       // basic
	Password     string `yaml:"password,omitempty"`       // basic (not recommended)
	ClientID     string `yaml:"client-id,omitempty"`      // oauth
	ClientSecret string `yaml:"client-secret,omitempty"` // oauth (not recommended)
	TokenURL     string `yaml:"token-url,omitempty"`      // oauth
}

// DefaultsConfig holds default settings.
type DefaultsConfig struct {
	Limit         int    `yaml:"limit,omitempty"`
	DisplayValue  string `yaml:"display-value,omitempty"`
	Output        string `yaml:"output,omitempty"`
	Editor        string `yaml:"editor,omitempty"`
	WatchInterval string `yaml:"watch-interval,omitempty"`
}

// DefaultConfigDir returns the platform-appropriate config directory.
func DefaultConfigDir() string {
	if v := os.Getenv("SNOWCTL_CONFIG_DIR"); v != "" {
		return v
	}
	switch runtime.GOOS {
	case "darwin":
		home, _ := os.UserHomeDir()
		return filepath.Join(home, "Library", "Application Support", "snowctl")
	case "windows":
		return filepath.Join(os.Getenv("LOCALAPPDATA"), "snowctl")
	default:
		if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
			return filepath.Join(xdg, "snowctl")
		}
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".config", "snowctl")
	}
}

// DefaultConfigPath returns the path to config.yaml.
func DefaultConfigPath() string {
	return filepath.Join(DefaultConfigDir(), "config.yaml")
}

// Load reads the config file from the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return defaultConfig(), nil
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}
	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	cfg.applyDefaults()
	return cfg, nil
}

// Save writes the config to the given path.
func Save(cfg *Config, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}
	return os.WriteFile(path, data, 0600)
}

// ActiveContext returns the currently selected context configuration.
func (c *Config) ActiveContext() (*ContextConfig, error) {
	if c.CurrentContext == "" {
		return nil, fmt.Errorf("no context configured; run 'snowctl config set-context <name>'")
	}
	for i := range c.Contexts {
		if c.Contexts[i].Name == c.CurrentContext {
			return &c.Contexts[i], nil
		}
	}
	return nil, fmt.Errorf("context %q not found in config", c.CurrentContext)
}

// EffectiveLimit returns the limit to use, considering context-level and global defaults.
func (c *Config) EffectiveLimit(ctx *ContextConfig) int {
	if ctx != nil && ctx.Defaults.Limit > 0 {
		return ctx.Defaults.Limit
	}
	if c.Defaults.Limit > 0 {
		return c.Defaults.Limit
	}
	return 50
}

// EffectiveOutput returns the output format to use.
func (c *Config) EffectiveOutput(ctx *ContextConfig) string {
	if ctx != nil && ctx.Defaults.Output != "" {
		return ctx.Defaults.Output
	}
	if c.Defaults.Output != "" {
		return c.Defaults.Output
	}
	return "table"
}

func defaultConfig() *Config {
	return &Config{
		APIVersion: "v1",
		Defaults: DefaultsConfig{
			Limit:         50,
			DisplayValue:  "true",
			Output:        "table",
			WatchInterval: "5s",
		},
	}
}

func (c *Config) applyDefaults() {
	if c.APIVersion == "" {
		c.APIVersion = "v1"
	}
	if c.Defaults.Limit == 0 {
		c.Defaults.Limit = 50
	}
	if c.Defaults.Output == "" {
		c.Defaults.Output = "table"
	}
	if c.Defaults.DisplayValue == "" {
		c.Defaults.DisplayValue = "true"
	}
	if c.Defaults.WatchInterval == "" {
		c.Defaults.WatchInterval = "5s"
	}
}
