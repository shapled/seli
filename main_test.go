package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanConfigDir(t *testing.T) {
	configDir, entries, err := ScanConfigDir()
	if err != nil {
		t.Fatalf("ScanConfigDir failed: %v", err)
	}

	if configDir == "" {
		t.Fatal("configDir should not be empty")
	}

	if len(entries) == 0 {
		t.Fatal("should have at least some entries in config directory")
	}

	// Verify config directory exists
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		t.Fatalf("config directory %s does not exist", configDir)
	}
}

func TestIsConfigFile(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{"config.json", true},
		{"settings.yaml", true},
		{"data.yml", true},
		{"config.toml", true},
		{"readme.txt", false},
		{"script.sh", false},
		{".gitignore", false},
		{"", false},
	}

	for _, tt := range tests {
		result := IsConfigFile(tt.name)
		if result != tt.expected {
			t.Errorf("IsConfigFile(%s) = %v; want %v", tt.name, result, tt.expected)
		}
	}
}

func TestLoadConfigFile(t *testing.T) {
	configDir, _, err := ScanConfigDir()
	if err != nil {
		t.Fatalf("ScanConfigDir failed: %v", err)
	}

	// Test JSON config
	jsonPath := filepath.Join(configDir, "development.json")
	if _, err := os.Stat(jsonPath); err == nil {
		config, err := LoadConfigFile(jsonPath)
		if err != nil {
			t.Fatalf("LoadConfigFile(%s) failed: %v", jsonPath, err)
		}

		if config.Name == "" {
			t.Error("config name should not be empty")
		}

		if len(config.Commands) == 0 {
			t.Error("should have at least one command")
		}

		for i, cmd := range config.Commands {
			if cmd.Name == "" {
				t.Errorf("command %d should have a name", i)
			}
			if cmd.Command == "" {
				t.Errorf("command %d should have a command", i)
			}
		}
	}

	// Test YAML config
	yamlPath := filepath.Join(configDir, "system.yaml")
	if _, err := os.Stat(yamlPath); err == nil {
		config, err := LoadConfigFile(yamlPath)
		if err != nil {
			t.Fatalf("LoadConfigFile(%s) failed: %v", yamlPath, err)
		}

		if config.Name == "" {
			t.Error("config name should not be empty")
		}
	}

	// Test TOML config
	tomlPath := filepath.Join(configDir, "docker.toml")
	if _, err := os.Stat(tomlPath); err == nil {
		config, err := LoadConfigFile(tomlPath)
		if err != nil {
			t.Fatalf("LoadConfigFile(%s) failed: %v", tomlPath, err)
		}

		if config.Name == "" {
			t.Error("config name should not be empty")
		}
	}
}

func TestLoadConfigFileWithInvalidFormat(t *testing.T) {
	// Create a temporary file with unsupported extension
	tmpFile := filepath.Join(t.TempDir(), "config.txt")
	if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	_, err := LoadConfigFile(tmpFile)
	if err == nil {
		t.Error("expected error for unsupported file format")
	}
}

func TestLoadConfigFileWithNonExistentFile(t *testing.T) {
	nonExistentPath := filepath.Join(t.TempDir(), "nonexistent.json")

	_, err := LoadConfigFile(nonExistentPath)
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}
