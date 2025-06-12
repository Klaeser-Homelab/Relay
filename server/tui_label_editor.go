package main

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// LabelEditorModel handles interactive label editing
type LabelEditorModel struct {
	issueID          int
	currentLabels    []string
	availableLabels  []string
	selected         int
	width            int
	height           int
	onComplete       func([]string) tea.Cmd
}

// LabelEditorData contains data for the label editor
type LabelEditorData struct {
	IssueID       int
	CurrentLabels []string
	OnComplete    func([]string) tea.Cmd
}

// NewLabelEditorModel creates a new label editor model
func NewLabelEditorModel(data LabelEditorData) LabelEditorModel {
	availableLabels := []string{"bug", "enhancement"}
	
	return LabelEditorModel{
		issueID:         data.IssueID,
		currentLabels:   append([]string(nil), data.CurrentLabels...), // Copy slice
		availableLabels: availableLabels,
		selected:        0,
		onComplete:      data.OnComplete,
	}
}

func (m LabelEditorModel) Init() tea.Cmd {
	return nil
}

func (m LabelEditorModel) Update(msg tea.Msg) (LabelEditorModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			// Cancel and go back
			return m, BackToPreviousView()
			
		case "up", "k":
			if m.selected > 0 {
				m.selected--
			}
			
		case "down", "j":
			if m.selected < len(m.availableLabels)-1 {
				m.selected++
			}
			
		case "enter", " ":
			// Toggle selected label
			selectedLabel := m.availableLabels[m.selected]
			if m.hasLabel(selectedLabel) {
				m.removeLabel(selectedLabel)
			} else {
				m.addLabel(selectedLabel)
			}
			
		case "s":
			// Save and go back
			if m.onComplete != nil {
				return m, m.onComplete(m.currentLabels)
			}
			return m, BackToPreviousView()
		}
	}
	
	return m, nil
}

func (m LabelEditorModel) View() string {
	var content strings.Builder
	
	// Title
	title := titleStyle.Render("🏷️  Edit Labels")
	content.WriteString(title + "\n")
	content.WriteString(strings.Repeat("=", 15) + "\n\n")
	
	// Instructions
	content.WriteString("Use ↑↓ to navigate, Enter to toggle, 's' to save, 'q' to cancel\n\n")
	
	// Available labels with checkmarks
	for i, label := range m.availableLabels {
		var line string
		
		// Add checkmark for selected labels
		if m.hasLabel(label) {
			line = "✓ "
		} else {
			line = "  "
		}
		
		// Add styled label
		if label == "bug" {
			line += lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true).Render(label)
		} else if label == "enhancement" {
			line += lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true).Render(label)
		} else {
			line += label
		}
		
		// Highlight selected item
		if i == m.selected {
			line = selectedIssueStyle.Render("> " + line)
		} else {
			line = unselectedIssueStyle.Render("  " + line)
		}
		
		content.WriteString(line + "\n")
	}
	
	content.WriteString("\n")
	
	// Current selection
	currentLabelsStr := strings.Join(m.currentLabels, ", ")
	if currentLabelsStr == "" {
		currentLabelsStr = "none"
	}
	
	grayStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Faint(true)
	content.WriteString(grayStyle.Render("Current labels: " + currentLabelsStr) + "\n\n")
	
	// Help
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	content.WriteString(helpStyle.Render("↑↓ Navigate  •  Enter Toggle  •  s Save  •  q Cancel") + "\n")
	
	return content.String()
}

// hasLabel checks if a label is currently selected
func (m LabelEditorModel) hasLabel(label string) bool {
	for _, l := range m.currentLabels {
		if l == label {
			return true
		}
	}
	return false
}

// addLabel adds a label to the current selection
func (m *LabelEditorModel) addLabel(label string) {
	if !m.hasLabel(label) {
		m.currentLabels = append(m.currentLabels, label)
	}
}

// removeLabel removes a label from the current selection
func (m *LabelEditorModel) removeLabel(label string) {
	for i, l := range m.currentLabels {
		if l == label {
			m.currentLabels = append(m.currentLabels[:i], m.currentLabels[i+1:]...)
			break
		}
	}
}