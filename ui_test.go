package main

import (
	"testing"
)

func TestCreateCommandItems(t *testing.T) {
	// Create a dummy ConfigFile with multiple commands
	config := &ConfigFile{
		Name: "Test Config",
		Commands: []CommandConfig{
			{Name: "Command 1", Command: "echo 1"},
			{Name: "Command 2", Command: "echo 2"},
			{Name: "Command 3", Command: "echo 3"},
		},
	}

	// Create items using the new function
	items := createCommandItems(config)

	// Check if the command reference of the first item is correct
	if len(items) > 0 {
		firstItem := items[0].(Item)
		firstCommand := config.Commands[0]
		lastCommand := config.Commands[len(config.Commands)-1]

		// The bug is that firstItem.command points to the last command
		if firstItem.command.Command == lastCommand.Command && firstItem.command.Name == lastCommand.Name {
			t.Errorf("Bug reproduced: The first item's command ('%s') incorrectly points to the last command ('%s')", firstItem.command.Name, lastCommand.Name)
		}

		// Also check if it's pointing to the correct command
		if firstItem.command.Command != firstCommand.Command || firstItem.command.Name != firstCommand.Name {
			t.Errorf("The first item's command ('%s') should point to the first command ('%s'), but it does not.", firstItem.command.Name, firstCommand.Name)
		}
	} else {
		t.Fatal("createCommandItems returned no items")
	}
}