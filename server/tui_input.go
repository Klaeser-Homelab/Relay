package main

import (
	"os/exec"
	"runtime"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Text Input Component
type TextInputModel struct {
	prompt      string
	placeholder string
	input       string
	onComplete  func(string) tea.Cmd
	width       int
	height      int
}

func NewTextInputModel(prompt, placeholder string, onComplete func(string) tea.Cmd) TextInputModel {
	return TextInputModel{
		prompt:      prompt,
		placeholder: placeholder,
		input:       "",
		onComplete:  onComplete,
	}
}

func (m TextInputModel) Init() tea.Cmd {
	return nil
}

func (m TextInputModel) Update(msg tea.Msg) (TextInputModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if m.onComplete != nil {
				return m, m.onComplete(m.input)
			}
			return m, BackToPreviousView()
			
		case "esc":
			return m, BackToPreviousView()
			
		case "backspace":
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
			}
			
		case "ctrl+v", "cmd+v":
			// Handle paste
			return m.handlePaste()
			
		case "ctrl+c":
			return m, tea.Quit
			
		default:
			// Add character to input
			if len(msg.String()) == 1 {
				m.input += msg.String()
			}
		}
	}
	
	return m, nil
}

// handlePaste handles clipboard paste operations
func (m TextInputModel) handlePaste() (TextInputModel, tea.Cmd) {
	clipboardText := getClipboardContent()
	if clipboardText != "" {
		// Clean up the clipboard text (remove newlines, trim whitespace)
		clipboardText = strings.ReplaceAll(clipboardText, "\n", " ")
		clipboardText = strings.ReplaceAll(clipboardText, "\r", " ")
		clipboardText = strings.TrimSpace(clipboardText)
		
		m.input += clipboardText
	}
	return m, nil
}

// getClipboardContent gets content from the system clipboard
func getClipboardContent() string {
	var cmd *exec.Cmd
	
	switch runtime.GOOS {
	case "darwin": // macOS
		cmd = exec.Command("pbpaste")
	case "linux":
		// Try xclip first, then xsel
		if _, err := exec.LookPath("xclip"); err == nil {
			cmd = exec.Command("xclip", "-selection", "clipboard", "-o")
		} else if _, err := exec.LookPath("xsel"); err == nil {
			cmd = exec.Command("xsel", "--clipboard", "--output")
		} else {
			return ""
		}
	case "windows":
		cmd = exec.Command("powershell", "-command", "Get-Clipboard")
	default:
		return ""
	}
	
	if cmd == nil {
		return ""
	}
	
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	
	return string(output)
}

func (m TextInputModel) View() string {
	var content strings.Builder
	
	// Title
	title := titleStyle.Render(m.prompt)
	content.WriteString(title + "\n\n")
	
	// Input field with placeholder
	inputDisplay := m.input
	if inputDisplay == "" && m.placeholder != "" {
		inputDisplay = helpStyle.Render(m.placeholder)
	} else {
		inputDisplay = normalStyle.Render(inputDisplay)
	}
	
	// Input box
	inputBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1).
		Width(m.width - 4).
		Render(inputDisplay + "│")
	
	content.WriteString(inputBox + "\n\n")
	
	// Help text
	help := helpStyle.Render("Enter to confirm • Esc to cancel • Ctrl+V/Cmd+V to paste")
	content.WriteString(help)
	
	return content.String()
}

// Confirmation Component
type ConfirmationModel struct {
	message   string
	onConfirm func(bool) tea.Cmd
	selected  bool // true for Yes, false for No
	width     int
	height    int
}

func NewConfirmationModel(message string, onConfirm func(bool) tea.Cmd) ConfirmationModel {
	return ConfirmationModel{
		message:   message,
		onConfirm: onConfirm,
		selected:  true, // Default to Yes
	}
}

func (m ConfirmationModel) Init() tea.Cmd {
	return nil
}

func (m ConfirmationModel) Update(msg tea.Msg) (ConfirmationModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter", " ":
			if m.onConfirm != nil {
				return m, m.onConfirm(m.selected)
			}
			return m, BackToPreviousView()
			
		case "left", "right", "h", "l":
			m.selected = !m.selected
			
		case "y", "Y":
			m.selected = true
			if m.onConfirm != nil {
				return m, m.onConfirm(true)
			}
			return m, BackToPreviousView()
			
		case "n", "N", "esc":
			m.selected = false
			if m.onConfirm != nil {
				return m, m.onConfirm(false)
			}
			return m, BackToPreviousView()
			
		case "ctrl+c":
			return m, tea.Quit
		}
	}
	
	return m, nil
}

func (m ConfirmationModel) View() string {
	var content strings.Builder
	
	// Title
	title := titleStyle.Render("Confirmation")
	content.WriteString(title + "\n\n")
	
	// Message
	content.WriteString(normalStyle.Render(m.message) + "\n\n")
	
	// Buttons
	yesStyle := normalStyle
	noStyle := normalStyle
	
	if m.selected {
		yesStyle = selectedStyle
	} else {
		noStyle = selectedStyle
	}
	
	yesButton := yesStyle.Render("[ Yes ]")
	noButton := noStyle.Render("[ No ]")
	
	buttons := lipgloss.JoinHorizontal(lipgloss.Left, yesButton, "  ", noButton)
	content.WriteString(buttons + "\n\n")
	
	// Help text
	help := helpStyle.Render("Y/N or ←/→ to select • Enter to confirm • Esc to cancel")
	content.WriteString(help)
	
	return content.String()
}