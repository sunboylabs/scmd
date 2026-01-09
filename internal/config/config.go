// Package config provides configuration management for scmd
package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// Config represents scmd configuration
type Config struct {
	Version        string         `mapstructure:"version"`
	Backends       BackendsConfig `mapstructure:"backends"`
	UI             UIConfig       `mapstructure:"ui"`
	Models         ModelsConfig   `mapstructure:"models"`
	SetupCompleted bool           `mapstructure:"setup_completed"`
}

// BackendsConfig for LLM backends
type BackendsConfig struct {
	Default string             `mapstructure:"default"`
	Local   LocalBackendConfig `mapstructure:"local"`
}

// LocalBackendConfig for local llama.cpp
type LocalBackendConfig struct {
	Model         string `mapstructure:"model"`
	ModelPath     string `mapstructure:"model_path"`
	ContextLength int    `mapstructure:"context_length"`
	GPULayers     int    `mapstructure:"gpu_layers"`
	Threads       int    `mapstructure:"threads"`
}

// UIConfig for UI preferences
type UIConfig struct {
	Streaming bool `mapstructure:"streaming"`
	Colors    bool `mapstructure:"colors"`
	Verbose   bool `mapstructure:"verbose"`
}

// ModelsConfig for model management
type ModelsConfig struct {
	Directory    string `mapstructure:"directory"`
	AutoDownload bool   `mapstructure:"auto_download"`
}

// DataDir returns the scmd data directory
func DataDir() string {
	// Check for environment variable first (useful for testing)
	if dir := os.Getenv("SCMD_DATA_DIR"); dir != "" {
		return dir
	}

	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".scmd")
}

// ConfigPath returns the config file path
func ConfigPath() string {
	return filepath.Join(DataDir(), "config.yaml")
}

// GetString returns a string config value
func (c *Config) GetString(key string) string {
	switch key {
	case "backends.default":
		return c.Backends.Default
	case "backends.local.model":
		return c.Backends.Local.Model
	case "models.directory":
		return c.Models.Directory
	default:
		return ""
	}
}

// GetBool returns a bool config value
func (c *Config) GetBool(key string) bool {
	switch key {
	case "ui.streaming":
		return c.UI.Streaming
	case "ui.colors":
		return c.UI.Colors
	case "ui.verbose":
		return c.UI.Verbose
	case "models.auto_download":
		return c.Models.AutoDownload
	case "setup_completed":
		return c.SetupCompleted
	default:
		return false
	}
}

// GetInt returns an int config value
func (c *Config) GetInt(key string) int {
	switch key {
	case "backends.local.context_length":
		return c.Backends.Local.ContextLength
	case "backends.local.gpu_layers":
		return c.Backends.Local.GPULayers
	case "backends.local.threads":
		return c.Backends.Local.Threads
	default:
		return 0
	}
}

// Set sets a configuration value
func (c *Config) Set(key string, value interface{}) error {
	switch key {
	case "backends.default":
		if v, ok := value.(string); ok {
			c.Backends.Default = v
			return nil
		}
		return fmt.Errorf("value must be a string")
	case "backends.local.model":
		if v, ok := value.(string); ok {
			c.Backends.Local.Model = v
			return nil
		}
		return fmt.Errorf("value must be a string")
	case "backends.local.context_length":
		if v, ok := value.(int); ok {
			c.Backends.Local.ContextLength = v
			return nil
		}
		return fmt.Errorf("value must be an integer")
	case "ui.streaming":
		if v, ok := value.(bool); ok {
			c.UI.Streaming = v
			return nil
		}
		return fmt.Errorf("value must be a boolean")
	case "ui.colors":
		if v, ok := value.(bool); ok {
			c.UI.Colors = v
			return nil
		}
		return fmt.Errorf("value must be a boolean")
	case "ui.verbose":
		if v, ok := value.(bool); ok {
			c.UI.Verbose = v
			return nil
		}
		return fmt.Errorf("value must be a boolean")
	case "models.auto_download":
		if v, ok := value.(bool); ok {
			c.Models.AutoDownload = v
			return nil
		}
		return fmt.Errorf("value must be a boolean")
	case "setup_completed":
		if v, ok := value.(bool); ok {
			c.SetupCompleted = v
			return nil
		}
		return fmt.Errorf("value must be a boolean")
	default:
		return fmt.Errorf("unknown configuration key: %s", key)
	}
}
