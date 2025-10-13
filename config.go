package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"
)

// CommandConfig represents a single command configuration
type CommandConfig struct {
	Name        string            `json:"name" yaml:"name" toml:"name"`
	Description string            `json:"description" yaml:"description" toml:"description"`
	Command     string            `json:"command" yaml:"command" toml:"command"`
	Args        []string          `json:"args,omitempty" yaml:"args,omitempty" toml:"args,omitempty"`
	Env         map[string]string `json:"env,omitempty" yaml:"env,omitempty" toml:"env,omitempty"`
	WorkDir     string            `json:"workDir,omitempty" yaml:"workDir,omitempty" toml:"workDir,omitempty"`
}

// ConfigFile represents a configuration file containing multiple commands
type ConfigFile struct {
	Name        string          `json:"name" yaml:"name" toml:"name"`
	Description string          `json:"description,omitempty" yaml:"description,omitempty" toml:"description,omitempty"`
	Commands    []CommandConfig `json:"commands" yaml:"commands" toml:"commands"`
}

// LoadConfigFile loads a configuration file from the given path
func LoadConfigFile(path string) (*ConfigFile, error) {
	ext := strings.ToLower(filepath.Ext(path))

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", path, err)
	}

	var config ConfigFile

	switch ext {
	case ".json":
		err = json.Unmarshal(data, &config)
	case ".yaml", ".yml":
		err = yaml.Unmarshal(data, &config)
	case ".toml":
		_, err = toml.Decode(string(data), &config)
	default:
		return nil, fmt.Errorf("unsupported file format: %s", ext)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", path, err)
	}

	// Set default name from filename if not provided
	if config.Name == "" {
		config.Name = strings.TrimSuffix(filepath.Base(path), ext)
	}

	return &config, nil
}

// ScanConfigDir scans ~/.seli/ directory for configuration files
func ScanConfigDir() (string, []os.DirEntry, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".seli")

	// Create config directory if it doesn't exist
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return "", nil, fmt.Errorf("failed to create config directory %s: %w", configDir, err)
		}
	}

	entries, err := os.ReadDir(configDir)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read config directory %s: %w", configDir, err)
	}

	return configDir, entries, nil
}

// IsConfigFile checks if a file is a supported configuration file
func IsConfigFile(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	return ext == ".json" || ext == ".yaml" || ext == ".yml" || ext == ".toml"
}
