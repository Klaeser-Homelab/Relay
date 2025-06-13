package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type CloseReasonModel struct {
	issueID    int
	issueTitle string
	options    []string
	selected   int
	onConfirm  func(string) tea.Cmd
	width      int
	height     int
}

func NewCloseReasonModel(data CloseReasonData) CloseReasonModel {
	return CloseReasonModel{
		issueID:    data.IssueID,
		issueTitle: data.IssueTitle,
		options: []string{
			"Close as completed - Issue was successfully resolved",
			"Close as not planned - Issue will not be implemented",
			"Close as duplicate - Issue is a duplicate of another",
		},
		selected:  0, // Default to "completed"
		onConfirm: data.OnConfirm,
		width:     80,
		height:    24,
	}
}

func (m CloseReasonModel) Init() tea.Cmd {
	return nil
}

func (m CloseReasonModel) Update(msg tea.Msg) (CloseReasonModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, BackToPreviousView()

		case "up", "k":
			if m.selected > 0 {
				m.selected--
			}

		case "down", "j":
			if m.selected < len(m.options)-1 {
				m.selected++
			}

		case "enter", " ":
			// Map selection to close reason
			var reason string
			switch m.selected {
			case 0:
				reason = "completed"
			case 1:
				reason = "not planned"
			case 2:
				reason = "duplicate"
			}

			if m.onConfirm != nil {
				return m, m.onConfirm(reason)
			}
			return m, BackToPreviousView()
		}
	}

	return m, nil
}

func (m CloseReasonModel) View() string {
	var content strings.Builder

	// Styles
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	// Header
	title := fmt.Sprintf("Close Issue #%d", m.issueID)
	content.WriteString(titleStyle.Render(title) + "\n")
	content.WriteString(strings.Repeat("=", len(title)) + "\n\n")

	// Issue title (truncated if too long)
	issueTitle := m.issueTitle
	if len(issueTitle) > 60 {
		issueTitle = issueTitle[:57] + "..."
	}
	content.WriteString(fmt.Sprintf("Issue: %s\n\n", issueTitle))

	// Question
	content.WriteString("How do you want to close this issue?\n\n")

	// Options
	for i, option := range m.options {
		prefix := "  "
		style := normalStyle

		if i == m.selected {
			prefix = "> "
			style = selectedStyle
		}

		content.WriteString(prefix + style.Render(option) + "\n")
	}

	// Help text
	content.WriteString("\n" + helpStyle.Render("↑/↓ Navigate • Enter Select • Esc Cancel"))

	return content.String()
}
