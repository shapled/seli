package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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
	Show        *bool             `json:"show,omitempty" yaml:"show,omitempty" toml:"show,omitempty"`
}

// ConfigFile represents a configuration file containing multiple commands
type ConfigFile struct {
	Name        string          `json:"name" yaml:"name" toml:"name"`
	Description string          `json:"description,omitempty" yaml:"description,omitempty" toml:"description,omitempty"`
	Show        *bool           `json:"show,omitempty" yaml:"show,omitempty" toml:"show,omitempty"`
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

	// Process environment variables
	if err := ProcessConfigWithEnv(&config, path); err != nil {
		return nil, fmt.Errorf("failed to process environment variables: %w", err)
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

// LoadEnvFile loads .env file from the given directory and its parent directories
func LoadEnvFile(configDir string) (map[string]string, error) {
	envVars := make(map[string]string)

	// First, check the config directory itself
	envFile := filepath.Join(configDir, ".env")
	if _, err := os.Stat(envFile); err == nil {
		fileEnvVars, err := parseEnvFile(envFile)
		if err != nil {
			return nil, fmt.Errorf("failed to parse .env file %s: %w", envFile, err)
		}
		for k, v := range fileEnvVars {
			envVars[k] = v
		}
	}

	// Then walk up to home directory for parent .env files
	currentDir := filepath.Dir(configDir)
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	// Walk from parent directory up to home directory (but not beyond)
	for {
		if currentDir == homeDir {
			break
		}

		envFile = filepath.Join(currentDir, ".env")
		if _, err := os.Stat(envFile); err == nil {
			fileEnvVars, err := parseEnvFile(envFile)
			if err != nil {
				return nil, fmt.Errorf("failed to parse .env file %s: %w", envFile, err)
			}
			// Parent directory variables are loaded only if not already set
			for k, v := range fileEnvVars {
				if _, exists := envVars[k]; !exists {
					envVars[k] = v
				}
			}
		}

		// Move to parent directory
		parent := filepath.Dir(currentDir)
		if parent == currentDir {
			break
		}
		currentDir = parent
	}

	return envVars, nil
}

// parseEnvFile parses a .env file and returns environment variables
func parseEnvFile(envFile string) (map[string]string, error) {
	envVars := make(map[string]string)

	file, err := os.Open(envFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse KEY=VALUE format
		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				// Remove quotes if present
				if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
				   (strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
					value = value[1 : len(value)-1]
				}

				envVars[key] = value
			}
		}
	}

	return envVars, scanner.Err()
}

// ExpandEnvVars expands environment variables in a string with escape support
func ExpandEnvVars(input string, envVars map[string]string) string {
	// Regular expression to match ${VAR_NAME} and \$ escape sequences
	re := regexp.MustCompile(`\\\$|\$\{([^}]+)\}`)

	return re.ReplaceAllStringFunc(input, func(match string) string {
		switch match {
		case "\\$":
			return "$" // Unescape
		case "\\${":
			return "${" // Unescape
		default:
			// Match ${VAR_NAME}
			if strings.HasPrefix(match, "${") && strings.HasSuffix(match, "}") {
				varName := match[2 : len(match)-1]
				if value, exists := envVars[varName]; exists {
					return value
				}
				// Fallback to system environment
				return os.Getenv(varName)
			}
			return match
		}
	})
}

// ProcessConfigWithEnv processes configuration file with environment variable expansion
func ProcessConfigWithEnv(config *ConfigFile, configPath string) error {
	// Get directory containing the config file
	configDir := filepath.Dir(configPath)

	// Load .env files
	envVars, err := LoadEnvFile(configDir)
	if err != nil {
		return fmt.Errorf("failed to load .env files: %w", err)
	}

	// Add system environment variables (lower priority)
	for _, env := range os.Environ() {
		if !strings.Contains(env, "=") {
			continue
		}
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			if _, exists := envVars[parts[0]]; !exists {
				envVars[parts[0]] = parts[1]
			}
		}
	}


	// Process environment variable expansion for all commands
	for i := range config.Commands {
		// First, expand env values using global environment variables
		expandedEnv := make(map[string]string)
		for k, v := range config.Commands[i].Env {
			expandedEnv[k] = ExpandEnvVars(v, envVars)
		}

		// Create command-specific environment by merging global env with command env
		// Command env has higher priority
		commandEnv := make(map[string]string)
		for k, v := range envVars {
			commandEnv[k] = v
		}
		for k, v := range expandedEnv {
			commandEnv[k] = v
		}

		// Now expand command fields using the merged environment
		config.Commands[i].Command = ExpandEnvVars(config.Commands[i].Command, commandEnv)

		// Expand args
		for j := range config.Commands[i].Args {
			config.Commands[i].Args[j] = ExpandEnvVars(config.Commands[i].Args[j], commandEnv)
		}

		// Update env with expanded values
		config.Commands[i].Env = expandedEnv

		// Expand workDir
		config.Commands[i].WorkDir = ExpandEnvVars(config.Commands[i].WorkDir, commandEnv)
	}

	return nil
}
