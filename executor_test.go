package main

import (
	"testing"
)

func TestNewCommandExecutor(t *testing.T) {
	executor := NewCommandExecutor()
	if executor == nil {
		t.Error("NewCommandExecutor should return a non-nil executor")
	}
}

func TestCommandExecutor_ExecuteCommand(t *testing.T) {
	executor := NewCommandExecutor()

	tests := []struct {
		name        string
		command     CommandConfig
		expectError bool
	}{
		{
			name: "Simple echo command",
			command: CommandConfig{
				Name:    "Test Echo",
				Command: "echo",
				Args:    []string{"hello", "world"},
			},
			expectError: false,
		},
		{
			name: "Command with environment variable",
			command: CommandConfig{
				Name:    "Test Env",
				Command: "sh",
				Args:    []string{"-c", "echo $TEST_VAR"},
				Env:     map[string]string{"TEST_VAR": "test_value"},
			},
			expectError: false,
		},
		{
			name: "Command without args",
			command: CommandConfig{
				Name:    "Test Date",
				Command: "date",
			},
			expectError: false,
		},
		{
			name: "Non-existent command",
			command: CommandConfig{
				Name:    "Invalid Command",
				Command: "nonexistentcommand12345",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := executor.ExecuteCommand(tt.command)
			if (err != nil) != tt.expectError {
				t.Errorf("ExecuteCommand() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestCommandExecutor_ExecuteCommandInBackground(t *testing.T) {
	executor := NewCommandExecutor()

	command := CommandConfig{
		Name:    "Background Sleep",
		Command: "sleep",
		Args:    []string{"0.1"},
	}

	cmd, err := executor.ExecuteCommandInBackground(command)
	if err != nil {
		t.Fatalf("ExecuteCommandInBackground() error = %v", err)
	}

	if cmd == nil {
		t.Error("ExecuteCommandInBackground() should return a non-nil cmd")
	}

	// Wait for the command to complete
	if err := cmd.Wait(); err != nil {
		t.Errorf("background command failed: %v", err)
	}
}

func TestCommandExecutor_WithEmptyCommand(t *testing.T) {
	executor := NewCommandExecutor()

	command := CommandConfig{
		Name:    "Empty Command",
		Command: "",
	}

	err := executor.ExecuteCommand(command)
	if err == nil {
		t.Error("expected error for empty command")
	}
}
