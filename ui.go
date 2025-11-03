package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Application states
type state int

const (
	stateBrowsing state = iota
	stateViewingCommands
	stateExecutingCommand
)

// Model represents the application state
type Model struct {
	state         state
	list          list.Model
	viewport      viewport.Model
	configDir     string
	currentPath   string
	configFiles   []ConfigFile
	currentConfig *ConfigFile
	executor      *CommandExecutor
	quitting      bool
	width, height int
}

// Item represents a list item (either file/folder or command)
type Item struct {
	title       string
	description string
	isDir       bool
	isCommand   bool
	command     *CommandConfig
}

func (i Item) Title() string       { return i.title }
func (i Item) Description() string { return i.description }
func (i Item) FilterValue() string { return i.title }

// Styles for the UI
var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#00D75F")).
			Padding(0, 1)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5F87")).
			Bold(true)

	selectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#EE6FF8")).
				Bold(true)
)

// InitialModel creates the initial model
func InitialModel() (Model, error) {
	configDir, entries, err := ScanConfigDir()
	if err != nil {
		return Model{}, err
	}

	// Create list items from directory entries
	var items []list.Item
	var configFiles []string
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() {
			items = append(items, Item{
				title:       name + "/",
				description: "Directory",
				isDir:       true,
			})
		} else if IsConfigFile(name) {
			configFiles = append(configFiles, name)
			items = append(items, Item{
				title:       name,
				description: "Config file",
				isDir:       false,
			})
		}
	}

	// Create the list
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = true
	delegate.Styles.SelectedTitle = selectedItemStyle
	delegate.Styles.SelectedDesc = selectedItemStyle.Copy().Foreground(lipgloss.Color("#A6A3FF"))

	l := list.New(items, delegate, 0, 0)
	l.Title = titleStyle.Render("Seli - Command Launcher")
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle

	model := Model{
		state:       stateBrowsing,
		list:        l,
		configDir:   configDir,
		currentPath: "", // Start at root config directory
		executor:    NewCommandExecutor(),
	}

	// If there's only one config file and no directories, open it directly using the same logic
	if len(configFiles) == 1 && len(items) == 1 {
		updatedModel, _ := model.openConfigFile(configFiles[0])
		return updatedModel, nil
	}

	return model, nil
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles updates
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.quitting = true
			return m, tea.Quit

		case tea.KeyEnter:
			return m.handleEnter()

		case tea.KeyBackspace:
			if m.state == stateViewingCommands {
				return m.goBackToBrowse()
			}

		case tea.KeyRunes:
			if len(msg.Runes) > 0 && msg.Runes[0] == 'q' && m.state == stateViewingCommands {
				return m.goBackToBrowse()
			}

		case tea.KeyUp:
			if m.state == stateBrowsing || m.state == stateViewingCommands {
				// Handle cycling logic BEFORE letting the list process the key
				if len(m.list.Items()) > 0 {
					currentIndex := m.list.Index()
					if currentIndex <= 0 {
						// We're at the first item, cycle to the last item
						return m.handleUp()
					}
				}
			}

		case tea.KeyDown:
			if m.state == stateBrowsing || m.state == stateViewingCommands {
				// Handle cycling logic BEFORE letting the list process the key
				if len(m.list.Items()) > 0 {
					currentIndex := m.list.Index()
					if currentIndex >= len(m.list.Items())-1 {
						// We're at the last item, cycle to the first item
						return m.handleDown()
					}
				}
			}
		}

	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.list.SetSize(msg.Width, msg.Height-4)
	}

	// Update list based on current state
	var cmd tea.Cmd
	if m.state == stateBrowsing || m.state == stateViewingCommands {
		m.list, cmd = m.list.Update(msg)
	}

	return m, cmd
}

// View renders the UI
func (m Model) View() string {
	if m.quitting {
		return ""
	}

	content := m.list.View()

	// Add status bar at bottom
	var status string
	switch m.state {
	case stateBrowsing:
		path := m.configDir
		if m.currentPath != "" {
			path = filepath.Join(m.configDir, m.currentPath)
		}
		status = statusStyle.Render(fmt.Sprintf("Browsing: %s", path))
	case stateViewingCommands:
		status = statusStyle.Render(fmt.Sprintf("Commands: %s", m.currentConfig.Name))
	case stateExecutingCommand:
		status = statusStyle.Render("Executing command...")
	}

	if m.height > 0 {
		return lipgloss.JoinVertical(lipgloss.Left, content, status)
	}
	return content
}

// handleEnter handles Enter key press
func (m Model) handleEnter() (Model, tea.Cmd) {
	selectedItem := m.list.SelectedItem()
	if selectedItem == nil {
		return m, nil
	}

	item := selectedItem.(Item)

	switch m.state {
	case stateBrowsing:
		if item.isDir {
			return m.enterDirectory(item.title)
		} else {
			return m.openConfigFile(item.title)
		}

	case stateViewingCommands:
		if item.isCommand && item.command != nil {
			return m.executeCommand(*item.command)
		}
	}

	return m, nil
}

// enterDirectory enters a subdirectory
func (m Model) enterDirectory(dirName string) (Model, tea.Cmd) {
	newPath := filepath.Join(m.currentPath, strings.TrimSuffix(dirName, "/"))
	fullPath := filepath.Join(m.configDir, newPath)

	entries, err := os.ReadDir(fullPath)
	if err != nil {
		m.list.Title = errorStyle.Render(fmt.Sprintf("Error: %v", err))
		return m, nil
	}

	var items []list.Item
	var configFiles []string
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() {
			items = append(items, Item{
				title:       name + "/",
				description: "Directory",
				isDir:       true,
			})
		} else if IsConfigFile(name) {
			configFiles = append(configFiles, name)
			items = append(items, Item{
				title:       name,
				description: "Config file",
				isDir:       false,
			})
		}
	}

	m.currentPath = newPath

	// If there's only one config file and no directories, open it directly
	if len(configFiles) == 1 && len(items) == 1 {
		return m.openConfigFile(configFiles[0])
	}

	m.list.SetItems(items)
	// Reset selection to first item when entering directory
	if len(items) > 0 {
		m.list.Select(0)
	}
	title := "Seli"
	if newPath != "" {
		title += " - " + newPath
	}
	m.list.Title = titleStyle.Render(title)

	return m, nil
}

// openConfigFile opens a configuration file and shows its commands
func (m Model) openConfigFile(filename string) (Model, tea.Cmd) {
	fullPath := filepath.Join(m.configDir, m.currentPath, filename)

	config, err := LoadConfigFile(fullPath)
	if err != nil {
		m.list.Title = errorStyle.Render(fmt.Sprintf("Error: %v", err))
		return m, nil
	}

	items := createCommandItems(config)

	m.state = stateViewingCommands
	m.currentConfig = config
	m.list.SetItems(items)
	// Reset selection to first command when opening config file
	if len(items) > 0 {
		m.list.Select(0)
	}
	m.list.Title = titleStyle.Render(fmt.Sprintf("Commands in %s", config.Name))

	return m, nil
}

// createCommandItems creates a slice of list.Item from a ConfigFile
func createCommandItems(config *ConfigFile) []list.Item {
	var items []list.Item
	for _, cmd := range config.Commands {
		cmd := cmd // Create a new variable for the current iteration
		description := cmd.Description
		if description == "" {
			description = cmd.Command
		}
		items = append(items, Item{
			title:       cmd.Name,
			description: description,
			isCommand:   true,
			command:     &cmd,
		})
	}
	return items
}

// goBackToBrowse returns to directory browsing
func (m Model) goBackToBrowse() (Model, tea.Cmd) {
	m.state = stateBrowsing
	m.currentConfig = nil

	// Reload directory contents
	var items []list.Item
	fullPath := filepath.Join(m.configDir, m.currentPath)
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		m.list.Title = errorStyle.Render(fmt.Sprintf("Error: %v", err))
		return m, nil
	}

	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() {
			items = append(items, Item{
				title:       name + "/",
				description: "Directory",
				isDir:       true,
			})
		} else if IsConfigFile(name) {
			items = append(items, Item{
				title:       name,
				description: "Config file",
				isDir:       false,
			})
		}
	}

	m.list.SetItems(items)
	title := "Seli - Command Launcher"
	if m.currentPath != "" {
		title += " - " + m.currentPath
	}
	m.list.Title = titleStyle.Render(title)

	return m, nil
}

// handleUp handles up key press with cycling
func (m Model) handleUp() (Model, tea.Cmd) {
	items := m.list.Items()
	if len(items) == 0 {
		return m, nil
	}

	currentIndex := m.list.Index()
	if currentIndex <= 0 {
		// Cycle to the last item
		m.list.Select(len(items) - 1)
	} else {
		m.list.Select(currentIndex - 1)
	}

	return m, nil
}

// handleDown handles down key press with cycling
func (m Model) handleDown() (Model, tea.Cmd) {
	items := m.list.Items()
	if len(items) == 0 {
		return m, nil
	}

	currentIndex := m.list.Index()
	if currentIndex >= len(items)-1 {
		// Cycle to the first item
		m.list.Select(0)
	} else {
		m.list.Select(currentIndex + 1)
	}

	return m, nil
}

// executeCommand executes the selected command
func (m Model) executeCommand(cmd CommandConfig) (Model, tea.Cmd) {
	m.state = stateExecutingCommand
	m.list.Title = statusStyle.Render(fmt.Sprintf("Executing: %s", cmd.Name))

	return m, tea.Quit
}
