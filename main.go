package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Create initial model
	initialModel, err := InitialModel()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing application: %v\n", err)
		os.Exit(1)
	}

	// Start the bubble tea program
	p := tea.NewProgram(initialModel, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}

	// Handle command execution after TUI exits
	model := finalModel.(Model)
	if model.state == stateExecutingCommand && model.currentConfig != nil {
		selectedItem := model.list.SelectedItem()
		if selectedItem != nil {
			item := selectedItem.(Item)
			if item.isCommand && item.command != nil {
				fmt.Printf("\nExecuting command: %s\n", item.command.Name)
				fmt.Printf("Command: %s\n", item.command.Command)
				if len(item.command.Args) > 0 {
					fmt.Printf("Args: %v\n", item.command.Args)
				}
				if len(item.command.Env) > 0 {
					fmt.Printf("Environment variables:\n")
					for k, v := range item.command.Env {
						fmt.Printf("  %s=%s\n", k, v)
					}
				}
				if item.command.WorkDir != "" {
					fmt.Printf("Working directory: %s\n", item.command.WorkDir)
				}
				fmt.Println()

				// Execute the command
				err := model.executor.ExecuteCommand(*item.command, model.currentConfig.Show)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error executing command: %v\n", err)
					os.Exit(1)
				}
			}
		}
	}
}
