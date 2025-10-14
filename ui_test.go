package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/charmbracelet/bubbles/list"
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

func TestEnterDirectoryWithSingleConfigFile(t *testing.T) {
	// Create temporary directory structure
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, ".seli")
	testDir := filepath.Join(configDir, "single-config")

	// Create directories
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Create a single config file
	configContent := `{
  "name": "Test Commands",
  "commands": [
    {
      "name": "test-cmd",
      "description": "Test command",
      "command": "echo test"
    }
  ]
}`
	configFile := filepath.Join(testDir, "commands.json")
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Create a model pointing to test directory
	model := Model{
		state:       stateBrowsing,
		configDir:   configDir,
		currentPath: "",
		executor:    NewCommandExecutor(),
	}

	// Initialize list with root directory content
	items := []list.Item{
		Item{title: "single-config/", description: "Directory", isDir: true},
	}
	model.list = list.New(items, list.NewDefaultDelegate(), 0, 0)

	// Enter the directory that contains only one config file
	updatedModel, _ := model.enterDirectory("single-config/")

	// Verify that we're now in the command viewing state
	if updatedModel.state != stateViewingCommands {
		t.Errorf("Expected state to be stateViewingCommands, got %v", updatedModel.state)
	}

	// Verify that currentConfig is set
	if updatedModel.currentConfig == nil {
		t.Error("Expected currentConfig to be set")
	}

	// Verify the config name
	if updatedModel.currentConfig.Name != "Test Commands" {
		t.Errorf("Expected config name 'Test Commands', got '%s'", updatedModel.currentConfig.Name)
	}

	// Verify the list contains commands, not directories
	if len(updatedModel.list.Items()) == 0 {
		t.Error("Expected list to contain commands")
	}

	firstItem := updatedModel.list.Items()[0].(Item)
	if !firstItem.isCommand {
		t.Error("Expected first item to be a command")
	}

	if firstItem.title != "test-cmd" {
		t.Errorf("Expected first command to be 'test-cmd', got '%s'", firstItem.title)
	}
}

func TestEnterDirectoryWithMultipleItems(t *testing.T) {
	// Create temporary directory structure
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, ".seli")
	testDir := filepath.Join(configDir, "multiple-items")

	// Create directories
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Create multiple config files
	configContent1 := `{"name": "Test 1", "commands": []}`
	configContent2 := `{"name": "Test 2", "commands": []}`

	if err := os.WriteFile(filepath.Join(testDir, "config1.json"), []byte(configContent1), 0644); err != nil {
		t.Fatalf("Failed to create config file 1: %v", err)
	}
	if err := os.WriteFile(filepath.Join(testDir, "config2.json"), []byte(configContent2), 0644); err != nil {
		t.Fatalf("Failed to create config file 2: %v", err)
	}

	// Create a model pointing to test directory
	model := Model{
		state:       stateBrowsing,
		configDir:   configDir,
		currentPath: "",
		executor:    NewCommandExecutor(),
	}

	// Initialize list with root directory content
	items := []list.Item{
		Item{title: "multiple-items/", description: "Directory", isDir: true},
	}
	model.list = list.New(items, list.NewDefaultDelegate(), 0, 0)

	// Enter the directory that contains multiple items
	updatedModel, _ := model.enterDirectory("multiple-items/")

	// Verify that we're still in browsing state (should not auto-open)
	if updatedModel.state != stateBrowsing {
		t.Errorf("Expected state to remain stateBrowsing, got %v", updatedModel.state)
	}

	// Verify that currentConfig is not set
	if updatedModel.currentConfig != nil {
		t.Error("Expected currentConfig to be nil")
	}

	// Verify the list contains the config files, not commands
	if len(updatedModel.list.Items()) != 2 {
		t.Errorf("Expected 2 items in list, got %d", len(updatedModel.list.Items()))
	}

	// Verify items are config files, not commands
	for _, item := range updatedModel.list.Items() {
		listItem := item.(Item)
		if listItem.isCommand {
			t.Error("Expected items to be config files, not commands")
		}
	}
}

func TestEnterDirectoryWithMixedItems(t *testing.T) {
	// Create temporary directory structure
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, ".seli")
	testDir := filepath.Join(configDir, "mixed-items")

	// Create directories
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Create one config file and one subdirectory
	configContent := `{"name": "Test Config", "commands": []}`
	if err := os.WriteFile(filepath.Join(testDir, "config.json"), []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	subDir := filepath.Join(testDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	// Create a model pointing to test directory
	model := Model{
		state:       stateBrowsing,
		configDir:   configDir,
		currentPath: "",
		executor:    NewCommandExecutor(),
	}

	// Initialize list with root directory content
	items := []list.Item{
		Item{title: "mixed-items/", description: "Directory", isDir: true},
	}
	model.list = list.New(items, list.NewDefaultDelegate(), 0, 0)

	// Enter the directory that contains mixed items
	updatedModel, _ := model.enterDirectory("mixed-items/")

	// Verify that we're still in browsing state (should not auto-open when mixed with directories)
	if updatedModel.state != stateBrowsing {
		t.Errorf("Expected state to remain stateBrowsing, got %v", updatedModel.state)
	}

	// Verify that currentConfig is not set
	if updatedModel.currentConfig != nil {
		t.Error("Expected currentConfig to be nil")
	}

	// Verify the list contains both items
	if len(updatedModel.list.Items()) != 2 {
		t.Errorf("Expected 2 items in list, got %d", len(updatedModel.list.Items()))
	}
}

func TestHandleUpWithCycling(t *testing.T) {
	// Create a model with some items
	items := []list.Item{
		Item{title: "Item 1", description: "First item"},
		Item{title: "Item 2", description: "Second item"},
		Item{title: "Item 3", description: "Third item"},
	}

	model := Model{
		state: stateBrowsing,
		list:  list.New(items, list.NewDefaultDelegate(), 0, 0),
	}

	// Test cycling from first to last
	model.list.Select(0) // Select first item
	updatedModel, _ := model.handleUp()

	if updatedModel.list.Index() != 2 {
		t.Errorf("Expected to cycle to index 2, got %d", updatedModel.list.Index())
	}

	// Test normal upward movement from middle
	model.list.Select(1) // Select second item
	updatedModel, _ = model.handleUp()

	if updatedModel.list.Index() != 0 {
		t.Errorf("Expected to move to index 0, got %d", updatedModel.list.Index())
	}
}

func TestHandleDownWithCycling(t *testing.T) {
	// Create a model with some items
	items := []list.Item{
		Item{title: "Item 1", description: "First item"},
		Item{title: "Item 2", description: "Second item"},
		Item{title: "Item 3", description: "Third item"},
	}

	model := Model{
		state: stateBrowsing,
		list:  list.New(items, list.NewDefaultDelegate(), 0, 0),
	}

	// Test cycling from last to first
	model.list.Select(2) // Select last item
	updatedModel, _ := model.handleDown()

	if updatedModel.list.Index() != 0 {
		t.Errorf("Expected to cycle to index 0, got %d", updatedModel.list.Index())
	}

	// Test normal downward movement from middle
	model.list.Select(1) // Select second item
	updatedModel, _ = model.handleDown()

	if updatedModel.list.Index() != 2 {
		t.Errorf("Expected to move to index 2, got %d", updatedModel.list.Index())
	}
}

func TestHandleUpDownWithEmptyList(t *testing.T) {
	// Create a model with no items
	var items []list.Item

	model := Model{
		state: stateBrowsing,
		list:  list.New(items, list.NewDefaultDelegate(), 0, 0),
	}

	// Test up with empty list
	updatedModel, _ := model.handleUp()
	// For empty lists, the index should remain unchanged or be at a safe default
	if len(updatedModel.list.Items()) != 0 {
		t.Errorf("Expected list to remain empty, got %d items", len(updatedModel.list.Items()))
	}

	// Test down with empty list
	updatedModel, _ = model.handleDown()
	if len(updatedModel.list.Items()) != 0 {
		t.Errorf("Expected list to remain empty, got %d items", len(updatedModel.list.Items()))
	}
}

func TestInitialModelWithSingleConfigFile(t *testing.T) {
	// Create temporary directory structure with single config file in root
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, ".seli")

	// Create config directory
	if err := os.Mkdir(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	// Create a single config file
	configContent := `{
  "name": "Root Commands",
  "commands": [
    {
      "name": "root-cmd",
      "description": "Root command",
      "command": "echo root"
    }
  ]
}`
	configFile := filepath.Join(configDir, "commands.json")
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Note: Since we can't override the ScanConfigDir function directly in Go,
	// we'll test the behavior by creating a test config directory structure
	// and using the existing ScanConfigDir logic with the HOME directory override

	// Temporarily override HOME directory to point to our test directory
	originalHome := os.Getenv("HOME")
	testHome := tempDir
	os.Setenv("HOME", testHome)
	defer func() {
		os.Setenv("HOME", originalHome)
	}()

	// Test InitialModel with single config file
	model, err := InitialModel()
	if err != nil {
		t.Fatalf("InitialModel failed: %v", err)
	}

	// Verify that we're directly in command viewing state
	if model.state != stateViewingCommands {
		t.Errorf("Expected state to be stateViewingCommands, got %v", model.state)
	}

	// Verify that currentConfig is set
	if model.currentConfig == nil {
		t.Error("Expected currentConfig to be set")
	}

	// Verify the config name
	if model.currentConfig.Name != "Root Commands" {
		t.Errorf("Expected config name 'Root Commands', got '%s'", model.currentConfig.Name)
	}

	// Verify the list contains commands, not config files
	if len(model.list.Items()) == 0 {
		t.Error("Expected list to contain commands")
	}

	firstItem := model.list.Items()[0].(Item)
	if !firstItem.isCommand {
		t.Error("Expected first item to be a command")
	}

	if firstItem.title != "root-cmd" {
		t.Errorf("Expected first command to be 'root-cmd', got '%s'", firstItem.title)
	}
}