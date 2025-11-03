package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// CommandExecutor handles command execution with environment variables
type CommandExecutor struct{}

// NewCommandExecutor creates a new command executor
func NewCommandExecutor() *CommandExecutor {
	return &CommandExecutor{}
}

// ExecuteCommand executes a command with the given configuration
func (e *CommandExecutor) ExecuteCommand(config CommandConfig, fileShow *bool) error {
	// Determine if we should show command details
	shouldShow := false
	if config.Show != nil {
		shouldShow = *config.Show
	} else if fileShow != nil {
		shouldShow = *fileShow
	}

	// Prepare the command and arguments
	var cmd *exec.Cmd
	var displayArgs []string

	if len(config.Args) > 0 {
		// Command with arguments
		displayArgs = append([]string{config.Command}, config.Args...)
		cmd = exec.Command(displayArgs[0], displayArgs[1:]...)
	} else {
		// Simple command (may contain spaces, need to split)
		parts := strings.Fields(config.Command)
		if len(parts) == 0 {
			return fmt.Errorf("empty command")
		}
		displayArgs = parts
		cmd = exec.Command(parts[0], parts[1:]...)
	}

	// Show command details if requested
	if shouldShow {
		fmt.Printf("\nExecuting command: %s\n", config.Name)
		fmt.Printf("Executing: %s", displayArgs[0])
		for i := 1; i < len(displayArgs); i++ {
			fmt.Printf(" %q", displayArgs[i])
		}
		fmt.Println()

		if len(config.Env) > 0 {
			fmt.Println("Environment variables:")
			for k, v := range config.Env {
				fmt.Printf("  %s=%q\n", k, v)
			}
		}

		if config.WorkDir != "" {
			fmt.Printf("Working directory: %q\n", config.WorkDir)
		}

		fmt.Println()
	}

	// Set working directory if specified
	if config.WorkDir != "" {
		cmd.Dir = config.WorkDir
	}

	// Set environment variables
	if len(config.Env) > 0 {
		env := os.Environ()
		for key, value := range config.Env {
			env = append(env, fmt.Sprintf("%s=%s", key, value))
		}
		cmd.Env = env
	}

	// Set standard input/output to current terminal
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Execute the command
	return cmd.Run()
}

// ExecuteCommandInBackground executes a command in background (for future use)
func (e *CommandExecutor) ExecuteCommandInBackground(config CommandConfig) (*exec.Cmd, error) {
	var cmd *exec.Cmd

	if len(config.Args) > 0 {
		args := append([]string{config.Command}, config.Args...)
		cmd = exec.Command(args[0], args[1:]...)
	} else {
		parts := strings.Fields(config.Command)
		if len(parts) == 0 {
			return nil, fmt.Errorf("empty command")
		}
		cmd = exec.Command(parts[0], parts[1:]...)
	}

	if config.WorkDir != "" {
		cmd.Dir = config.WorkDir
	}

	if len(config.Env) > 0 {
		env := os.Environ()
		for key, value := range config.Env {
			env = append(env, fmt.Sprintf("%s=%s", key, value))
		}
		cmd.Env = env
	}

	// Start the command in background
	err := cmd.Start()
	return cmd, err
}
