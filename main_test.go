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

func TestExpandEnvVars(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		envVars  map[string]string
		expected string
	}{
		{
			name:     "simple variable substitution",
			input:    "Hello ${WORLD}",
			envVars:  map[string]string{"WORLD": "Go"},
			expected: "Hello Go",
		},
		{
			name:     "escaped dollar sign",
			input:    "Price: \\$100",
			envVars:  map[string]string{},
			expected: "Price: $100",
		},
		{
			name:     "escaped variable syntax",
			input:    "Literal: \\${NOT_A_VAR}",
			envVars:  map[string]string{},
			expected: "Literal: ${NOT_A_VAR}",
		},
		{
			name:     "non-existent variable",
			input:    "Value: ${MISSING}",
			envVars:  map[string]string{},
			expected: "Value: ",
		},
		{
			name:     "multiple variables",
			input:    "${FIRST} and ${SECOND}",
			envVars:  map[string]string{"FIRST": "A", "SECOND": "B"},
			expected: "A and B",
		},
		{
			name:     "no variables",
			input:    "Just plain text",
			envVars:  map[string]string{},
			expected: "Just plain text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExpandEnvVars(tt.input, tt.envVars)
			if result != tt.expected {
				t.Errorf("ExpandEnvVars(%q, envVars) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseEnvFile(t *testing.T) {
	// Create a temporary .env file
	envContent := `# This is a comment
BASIC_VAR=simple_value
COMPLEX_VAR="value with spaces"
QUOTED_VAR='single quoted'
EMPTY_VAR=
# Another comment
WORK_DIR=/tmp/test`

	tmpFile := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(tmpFile, []byte(envContent), 0644); err != nil {
		t.Fatalf("Failed to create temp .env file: %v", err)
	}

	envVars, err := parseEnvFile(tmpFile)
	if err != nil {
		t.Fatalf("parseEnvFile failed: %v", err)
	}

	expectedVars := map[string]string{
		"BASIC_VAR":   "simple_value",
		"COMPLEX_VAR": "value with spaces",
		"QUOTED_VAR":  "single quoted",
		"EMPTY_VAR":   "",
		"WORK_DIR":    "/tmp/test",
	}

	if len(envVars) != len(expectedVars) {
		t.Errorf("Expected %d variables, got %d", len(expectedVars), len(envVars))
	}

	for key, expected := range expectedVars {
		if actual, exists := envVars[key]; !exists {
			t.Errorf("Missing variable: %s", key)
		} else if actual != expected {
			t.Errorf("Variable %s: expected %q, got %q", key, expected, actual)
		}
	}
}

func TestProcessConfigWithEnv(t *testing.T) {
	tempDir := t.TempDir()

	// Create a temporary .env file in the same directory as the config file
	envContent := `BASIC_VAR=basic_value
COMPLEX_VAR="complex value with spaces"
WORK_DIR=/tmp/test`

	envFile := filepath.Join(tempDir, ".env")
	if err := os.WriteFile(envFile, []byte(envContent), 0644); err != nil {
		t.Fatalf("Failed to create temp .env file: %v", err)
	}

	// Create a temporary config file
	configContent := `{
		"name": "Test Config",
		"show": true,
		"commands": [
			{
				"name": "Test Command",
				"command": "echo",
				"args": ["${BASIC_VAR}", "\\${ESCAPED}", "${COMPLEX_VAR}"],
				"workDir": "${WORK_DIR}",
				"env": {
					"LOCAL_VAR": "${BASIC_VAR}_local"
				},
				"show": false
			}
		]
	}`

	configFile := filepath.Join(tempDir, "test.json")
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create temp config file: %v", err)
	}

	// Load and process the config
	config, err := LoadConfigFile(configFile)
	if err != nil {
		t.Fatalf("LoadConfigFile failed: %v", err)
	}

	
	// Verify environment variable expansion
	if len(config.Commands) != 1 {
		t.Fatalf("Expected 1 command, got %d", len(config.Commands))
	}

	cmd := config.Commands[0]

	// Check command arguments
	expectedArgs := []string{"basic_value", "${ESCAPED}", "complex value with spaces"}
	if len(cmd.Args) != len(expectedArgs) {
		t.Errorf("Expected %d args, got %d", len(expectedArgs), len(cmd.Args))
	}
	for i, expected := range expectedArgs {
		if i < len(cmd.Args) && cmd.Args[i] != expected {
			t.Errorf("Arg %d: expected %q, got %q", i, expected, cmd.Args[i])
		}
	}

	// Check work directory expansion
	if cmd.WorkDir != "/tmp/test" {
		t.Errorf("Expected workDir %q, got %q", "/tmp/test", cmd.WorkDir)
	}

	// Check local environment variable expansion
	if localVar, exists := cmd.Env["LOCAL_VAR"]; !exists {
		t.Error("LOCAL_VAR not found in command env")
	} else if localVar != "basic_value_local" {
		t.Errorf("Expected LOCAL_VAR %q, got %q", "basic_value_local", localVar)
	}

	// Check show settings
	if config.Show == nil || *config.Show != true {
		t.Error("Expected file-level show to be true")
	}
	if cmd.Show == nil || *cmd.Show != false {
		t.Error("Expected command-level show to be false")
	}
}
