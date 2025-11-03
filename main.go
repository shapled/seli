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
				// Execute the command (show details will be handled inside ExecuteCommand)
				err := model.executor.ExecuteCommand(*item.command, model.currentConfig.Show)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error executing command: %v\n", err)
					os.Exit(1)
				}
			}
		}
	}
}
