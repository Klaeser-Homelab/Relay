package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Config Menu Models
type ConfigMenuModel struct {
	configManager *ConfigManager
	selected      int
	width         int
	height        int
	menuItems     []string
}

type LLMConfigModel struct {
	configManager *ConfigManager
	selected      int
	width         int
	height        int
	menuItems     []string
	llmOptions    []string
}

func NewConfigMenuModel(configManager *ConfigManager) ConfigMenuModel {
	return ConfigMenuModel{
		configManager: configManager,
		selected:      0,
		menuItems:     []string{"Issue Tracker", "LLMs"},
	}
}

func NewLLMConfigModel(configManager *ConfigManager) LLMConfigModel {
	return LLMConfigModel{
		configManager: configManager,
		selected:      0,
		menuItems:     []string{"Planning", "Executing"},
		llmOptions:    []string{"claude", "openai", "local", "anthropic", "chatgpt"},
	}
}

// Config Menu Implementation
func (m ConfigMenuModel) Init() tea.Cmd {
	return nil
}

func (m ConfigMenuModel) Update(msg tea.Msg) (ConfigMenuModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			return m, SwitchToView(ViewREPL, nil)
			
		case "up", "k":
			if m.selected > 0 {
				m.selected--
			}
			
		case "down", "j":
			if m.selected < len(m.menuItems)-1 {
				m.selected++
			}
			
		case "enter", " ":
			return m.executeAction()
		}
	}
	
	return m, nil
}

func (m ConfigMenuModel) executeAction() (ConfigMenuModel, tea.Cmd) {
	switch m.selected {
	case 0: // Issue Tracker
		return m, SwitchToView(ViewIssueTrackerConfig, nil)
	case 1: // LLMs
		return m, SwitchToView(ViewLLMConfig, nil)
	}
	return m, nil
}

func (m ConfigMenuModel) View() string {
	var content strings.Builder
	
	// Title
	title := titleStyle.Render("⚙️  Configuration")
	content.WriteString(title + "\n")
	content.WriteString(strings.Repeat("=", 14) + "\n\n")
	
	// Menu items
	for i, item := range m.menuItems {
		var line string
		if i == m.selected {
			line = selectedIssueStyle.Render("> " + item)
		} else {
			line = unselectedIssueStyle.Render("  " + item)
		}
		content.WriteString(line + "\n")
	}
	
	content.WriteString("\n")
	
	// Current config display
	config := m.configManager.GetConfig()
	content.WriteString(helpStyle.Render("Current Settings:") + "\n")
	content.WriteString(helpStyle.Render(fmt.Sprintf("Issue Tracker: %s", config.IssueTracker.Provider)) + "\n")
	content.WriteString(helpStyle.Render(fmt.Sprintf("Planning LLM: %s", config.LLMs.Planning.Type)) + "\n")
	content.WriteString(helpStyle.Render(fmt.Sprintf("Executing LLM: %s", config.LLMs.Executing.Type)) + "\n\n")
	
	// Define color styles for different action types (matching issues page)
	backStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Bold(true)     // Gray for back
	
	// Help with colored action keys
	actionOptions := []string{
		backStyle.Render("q") + " Back",
	}
	
	help := helpStyle.Render("Controls: " + strings.Join(actionOptions, " • "))
	content.WriteString(help)
	
	return content.String()
}

// LLM Config Implementation
func (m LLMConfigModel) Init() tea.Cmd {
	return nil
}

func (m LLMConfigModel) Update(msg tea.Msg) (LLMConfigModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			return m, SwitchToView(ViewConfig, nil)
			
		case "up", "k":
			if m.selected > 0 {
				m.selected--
			}
			
		case "down", "j":
			if m.selected < len(m.menuItems)-1 {
				m.selected++
			}
			
		case "enter", " ":
			return m.executeAction()
			
		case "e":
			// Edit shortcut
			if m.selected < len(m.menuItems) { // Only for Planning and Executing
				return m.editLLMSetting()
			}
		}
	}
	
	return m, nil
}

func (m LLMConfigModel) executeAction() (LLMConfigModel, tea.Cmd) {
	switch m.selected {
	case 0: // Planning
		return m.editLLMSetting()
	case 1: // Executing
		return m.editLLMSetting()
	}
	return m, nil
}

func (m LLMConfigModel) editLLMSetting() (LLMConfigModel, tea.Cmd) {
	var settingName string
	var currentValue string
	config := m.configManager.GetConfig()
	
	switch m.selected {
	case 0: // Planning
		settingName = "Planning LLM"
		currentValue = config.LLMs.Planning.Type
	case 1: // Executing
		settingName = "Executing LLM"
		currentValue = config.LLMs.Executing.Type
	default:
		return m, nil
	}
	
	inputData := TextInputData{
		Prompt:      fmt.Sprintf("Edit %s", settingName),
		Placeholder: currentValue,
		OnComplete: func(newValue string) tea.Cmd {
			if newValue != "" && newValue != currentValue {
				var err error
				switch m.selected {
				case 0: // Planning
					err = m.configManager.UpdateLLMPlanningType(newValue)
				case 1: // Executing
					err = m.configManager.UpdateLLMExecutingType(newValue)
				}
				
				if err != nil {
					// Handle error - in a real implementation, show error message
				}
			}
			return BackToPreviousView()
		},
	}
	
	return m, SwitchToView(ViewTextInput, inputData)
}

func (m LLMConfigModel) View() string {
	var content strings.Builder
	
	// Title
	title := titleStyle.Render("🤖 LLM Configuration")
	content.WriteString(title + "\n")
	content.WriteString(strings.Repeat("=", 18) + "\n\n")
	
	// Current config
	config := m.configManager.GetConfig()
	
	// Menu items with current values
	menuItemsWithValues := []string{
		fmt.Sprintf("Planning: %s", config.LLMs.Planning.Type),
		fmt.Sprintf("Executing: %s", config.LLMs.Executing.Type),
	}
	
	for i, item := range menuItemsWithValues {
		var line string
		if i == m.selected {
			line = selectedIssueStyle.Render("> " + item)
		} else {
			line = unselectedIssueStyle.Render("  " + item)
		}
		content.WriteString(line + "\n")
	}
	
	content.WriteString("\n")
	
	// Available options
	content.WriteString(helpStyle.Render("Available LLM Options:") + "\n")
	content.WriteString(helpStyle.Render(strings.Join(m.llmOptions, ", ")) + "\n\n")
	
	// Define color styles for different action types (matching issues page)
	editStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Bold(true)    // Yellow for edit
	backStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Bold(true)     // Gray for back
	
	// Help with colored action keys
	actionOptions := []string{
		editStyle.Render("e") + " Edit",
		backStyle.Render("q") + " Back",
	}
	
	help := helpStyle.Render("Controls: " + strings.Join(actionOptions, " • "))
	content.WriteString(help)
	
	return content.String()
}

// Issue Tracker Config Model
type IssueTrackerConfigModel struct {
	configManager *ConfigManager
	selected      int
	width         int
	height        int
	menuItems     []string
	providerOptions []string
}

func NewIssueTrackerConfigModel(configManager *ConfigManager) IssueTrackerConfigModel {
	return IssueTrackerConfigModel{
		configManager: configManager,
		selected:      0,
		menuItems:     []string{"Provider"},
		providerOptions: []string{"local", "github"},
	}
}

func (m IssueTrackerConfigModel) Init() tea.Cmd {
	return nil
}

func (m IssueTrackerConfigModel) Update(msg tea.Msg) (IssueTrackerConfigModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			return m, SwitchToView(ViewConfig, nil)
			
		case "up", "k":
			if m.selected > 0 {
				m.selected--
			}
			
		case "down", "j":
			if m.selected < len(m.menuItems)-1 {
				m.selected++
			}
			
		case "enter", " ":
			return m.executeAction()
			
		case "e":
			// Edit shortcut
			if m.selected < len(m.menuItems) { // Only for Provider
				return m.editProviderSetting()
			}
		}
	}
	
	return m, nil
}

func (m IssueTrackerConfigModel) executeAction() (IssueTrackerConfigModel, tea.Cmd) {
	switch m.selected {
	case 0: // Provider
		return m.editProviderSetting()
	}
	return m, nil
}

func (m IssueTrackerConfigModel) editProviderSetting() (IssueTrackerConfigModel, tea.Cmd) {
	config := m.configManager.GetConfig()
	currentValue := config.IssueTracker.Provider
	
	inputData := TextInputData{
		Prompt:      "Edit Issue Tracker Provider",
		Placeholder: currentValue,
		OnComplete: func(newValue string) tea.Cmd {
			if newValue != "" && newValue != currentValue {
				err := m.configManager.UpdateIssueTracker(newValue)
				if err != nil {
					// Handle error - in a real implementation, show error message
				}
			}
			return BackToPreviousView()
		},
	}
	
	return m, SwitchToView(ViewTextInput, inputData)
}

func (m IssueTrackerConfigModel) View() string {
	var content strings.Builder
	
	// Title
	title := titleStyle.Render("📋 Issue Tracker Configuration")
	content.WriteString(title + "\n")
	content.WriteString(strings.Repeat("=", 28) + "\n\n")
	
	// Current config
	config := m.configManager.GetConfig()
	
	// Menu items with current values
	menuItemsWithValues := []string{
		fmt.Sprintf("Provider: %s", config.IssueTracker.Provider),
	}
	
	for i, item := range menuItemsWithValues {
		var line string
		if i == m.selected {
			line = selectedIssueStyle.Render("> " + item)
		} else {
			line = unselectedIssueStyle.Render("  " + item)
		}
		content.WriteString(line + "\n")
	}
	
	content.WriteString("\n")
	
	// Available options
	content.WriteString(helpStyle.Render("Available Providers:") + "\n")
	content.WriteString(helpStyle.Render(strings.Join(m.providerOptions, ", ")) + "\n\n")
	
	// Define color styles for different action types (matching issues page)
	editStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Bold(true)    // Yellow for edit
	backStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Bold(true)     // Gray for back
	
	// Help with colored action keys
	actionOptions := []string{
		editStyle.Render("e") + " Edit",
		backStyle.Render("q") + " Back",
	}
	
	help := helpStyle.Render("Controls: " + strings.Join(actionOptions, " • "))
	content.WriteString(help)
	
	return content.String()
}