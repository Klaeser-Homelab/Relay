package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// ANSI color codes
const (
	ColorReset  = "\033[0m"
	ColorBlue   = "\033[34m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorRed    = "\033[31m"
	ColorBold   = "\033[1m"
	ColorCyan   = "\033[36m"
	ColorGray   = "\033[90m"
)

// ANSI control codes
const (
	ClearScreen = "\033[2J"
	CursorHome  = "\033[H"
	HideCursor  = "\033[?25l"
	ShowCursor  = "\033[?25h"
	CursorUp    = "\033[A"
	CursorDown  = "\033[B"
)

// Layout constants for consistent alignment
const (
	LabelWidth    = 12  // Width for labels like "Title:", "Status:"
	ActionWidth   = 2   // Width for action keys like "c", "r"
	StatusPadding = 20  // Padding for status display
)

// MenuItem represents an item in an interactive menu
type MenuItem struct {
	ID      int
	Content string
	Data    interface{} // Can store additional data like Issue struct
}

// InteractiveMenu handles terminal-based interactive menus
type InteractiveMenu struct {
	title       string
	items       []MenuItem
	selectedIdx int
	showHelp    bool
}

// NewInteractiveMenu creates a new interactive menu
func NewInteractiveMenu(title string, items []MenuItem) *InteractiveMenu {
	return &InteractiveMenu{
		title:       title,
		items:       items,
		selectedIdx: 0,
		showHelp:    true,
	}
}

// SetSttyRaw sets terminal to raw mode for character input
func SetSttyRaw() error {
	cmd := exec.Command("stty", "raw", "-echo")
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// SetSttyCooked restores terminal to normal mode
func SetSttyCooked() error {
	cmd := exec.Command("stty", "cooked", "echo")
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// clearScreen clears the terminal screen
func clearScreen() {
	fmt.Print(ClearScreen + CursorHome)
}

// Display renders the menu with proper column alignment
func (m *InteractiveMenu) Display() {
	clearScreen()
	fmt.Print(HideCursor)
	
	// Title section with consistent formatting
	fmt.Printf("%s%s%s\n", ColorBold, m.title, ColorReset)
	fmt.Printf("%s\n\n", strings.Repeat("=", len(m.title)))
	
	// Menu items with consistent left alignment and selection indicator
	for i, item := range m.items {
		if i == m.selectedIdx {
			// Selected item with clear visual indicator
			fmt.Printf("%s> %s%s\n", 
				ColorBlue, 
				item.Content,
				ColorReset)
		} else {
			// Unselected item with consistent spacing
			fmt.Printf("  %s\n", item.Content)
		}
	}
	
	// Help section with aligned columns
	if m.showHelp {
		fmt.Printf("\n%sControls:%s\n", ColorBold, ColorReset)
		fmt.Printf("%s\n", strings.Repeat("-", 9))
		fmt.Printf("%-8s %s\n", "↑/↓", "Navigate")
		fmt.Printf("%-8s %s\n", "Enter", "Select item")
		fmt.Printf("%-8s %s\n", "c", "Close item")
		fmt.Printf("%-8s %s\n", "d", "Delete item")
		fmt.Printf("%-8s %s\n", "q", "Quit")
	}
}

// Run starts the interactive menu loop
func (m *InteractiveMenu) Run() (*MenuItem, string, error) {
	if len(m.items) == 0 {
		return nil, "", fmt.Errorf("no items to display")
	}
	
	// Set terminal to raw mode
	if err := SetSttyRaw(); err != nil {
		return nil, "", fmt.Errorf("failed to set terminal to raw mode: %w", err)
	}
	defer func() {
		SetSttyCooked()
		fmt.Print(ShowCursor)
		clearScreen()
	}()
	
	// Initial display
	m.Display()
	
	// Input loop
	reader := bufio.NewReader(os.Stdin)
	for {
		char, err := reader.ReadByte()
		if err != nil {
			return nil, "", fmt.Errorf("error reading input: %w", err)
		}
		
		switch char {
		case 27: // ESC sequence (arrow keys)
			// Read next two bytes for arrow key detection
			seq1, _ := reader.ReadByte()
			seq2, _ := reader.ReadByte()
			
			if seq1 == '[' {
				switch seq2 {
				case 'A': // Up arrow
					if m.selectedIdx > 0 {
						m.selectedIdx--
						m.Display()
					}
				case 'B': // Down arrow
					if m.selectedIdx < len(m.items)-1 {
						m.selectedIdx++
						m.Display()
					}
				}
			}
		case 13: // Enter
			selected := &m.items[m.selectedIdx]
			return selected, "select", nil
		case 'c', 'C': // Close
			selected := &m.items[m.selectedIdx]
			return selected, "close", nil
		case 'd', 'D': // Delete (backward compatibility)
			selected := &m.items[m.selectedIdx]
			return selected, "delete", nil
		case 'q', 'Q': // Quit
			return nil, "quit", nil
		}
	}
}

// ConfirmationDialog shows a Y/n confirmation dialog
func ConfirmationDialog(message string) (bool, error) {
	// Set terminal to raw mode
	if err := SetSttyRaw(); err != nil {
		return false, fmt.Errorf("failed to set terminal to raw mode: %w", err)
	}
	defer SetSttyCooked()
	
	fmt.Printf("\n%s%s%s [Y/n]: ", ColorYellow, message, ColorReset)
	
	reader := bufio.NewReader(os.Stdin)
	char, err := reader.ReadByte()
	if err != nil {
		return false, fmt.Errorf("error reading input: %w", err)
	}
	
	fmt.Println() // New line after input
	
	switch char {
	case 'Y', 'y', 13: // Y, y, or Enter (default to Yes)
		return true, nil
	case 'N', 'n':
		return false, nil
	default:
		return false, nil // Default to No for any other input
	}
}

// IssueActionMenu shows the action menu for a selected issue
type IssueActionMenu struct {
	issue *Issue
}

// NewIssueActionMenu creates a new issue action menu
func NewIssueActionMenu(issue *Issue) *IssueActionMenu {
	return &IssueActionMenu{issue: issue}
}

// getStatusDisplay returns a clean text representation of status
func getStatusDisplay(status string) string {
	switch strings.ToLower(status) {
	case "captured":
		return ColorGreen + "[CAPTURED]" + ColorReset
	case "in-progress":
		return ColorYellow + "[IN-PROGRESS]" + ColorReset
	case "done":
		return ColorBlue + "[DONE]" + ColorReset
	case "archived":
		return ColorCyan + "[ARCHIVED]" + ColorReset
	default:
		return ColorCyan + "[" + strings.ToUpper(status) + "]" + ColorReset
	}
}

// getLabelsDisplay returns a clean text representation of labels
func getLabelsDisplay(labels []string) string {
	if len(labels) == 0 {
		return "" // Return empty string instead of "[no labels]"
	}
	
	var displayLabels []string
	for _, label := range labels {
		switch strings.ToLower(label) {
		case "enhancement":
			displayLabels = append(displayLabels, ColorGreen + "[ENHANCEMENT]" + ColorReset)
		case "bug":
			displayLabels = append(displayLabels, ColorRed + "[BUG]" + ColorReset)
		default:
			displayLabels = append(displayLabels, ColorCyan + "[" + strings.ToUpper(label) + "]" + ColorReset)
		}
	}
	
	return strings.Join(displayLabels, " ")
}

// Display renders the issue action menu with perfect column alignment
func (m *IssueActionMenu) Display() {
	clearScreen()
	fmt.Print(HideCursor)
	
	// Header section
	fmt.Printf("%sIssue #%d%s\n", ColorBold, m.issue.ID, ColorReset)
	fmt.Printf("%s\n\n", strings.Repeat("=", 15))
	
	// Issue details with fixed-width left column for labels
	fmt.Printf("%-*s %s\n", LabelWidth, "Title:", m.issue.Content)
	fmt.Printf("%-*s %s\n", LabelWidth, "Status:", getStatusDisplay(m.issue.Status))
	
	// Only show labels line if labels exist
	if len(m.issue.Labels) > 0 {
		fmt.Printf("%-*s %s\n", LabelWidth, "Labels:", getLabelsDisplay(m.issue.Labels))
	}
	
	fmt.Printf("%-*s %s\n", LabelWidth, "Created:", formatRelativeTime(m.issue.Timestamp))
	
	// Actions section with aligned columns
	fmt.Printf("\n%sActions:%s\n", ColorBold, ColorReset)
	fmt.Printf("%s\n", strings.Repeat("-", 8))
	
	// Action menu items with consistent key/description alignment
	fmt.Printf("%-*s %s\n", ActionWidth+1, "c", "Chat about this issue with Claude")
	fmt.Printf("%-*s %s\n", ActionWidth+1, "r", "Rename this issue")
	fmt.Printf("%-*s %s\n", ActionWidth+1, "o", "Close this issue")
	fmt.Printf("%-*s %s\n", ActionWidth+1, "d", "Delete this issue")
	fmt.Printf("%-*s %s\n", ActionWidth+1, "p", "Push this issue to GitHub")
	fmt.Printf("%-*s %s\n", ActionWidth+1, "q", "Back to issue list")
}

// Run starts the issue action menu loop
func (m *IssueActionMenu) Run() (string, error) {
	// Set terminal to raw mode
	if err := SetSttyRaw(); err != nil {
		return "", fmt.Errorf("failed to set terminal to raw mode: %w", err)
	}
	defer func() {
		SetSttyCooked()
		fmt.Print(ShowCursor)
	}()
	
	// Initial display
	m.Display()
	
	// Input loop
	reader := bufio.NewReader(os.Stdin)
	for {
		char, err := reader.ReadByte()
		if err != nil {
			return "", fmt.Errorf("error reading input: %w", err)
		}
		
		switch char {
		case 'c', 'C':
			return "chat", nil
		case 'r', 'R':
			return "rename", nil
		case 'o', 'O':
			return "close", nil
		case 'd', 'D':
			return "delete", nil
		case 'p', 'P':
			return "push", nil
		case 'q', 'Q', 27: // q, Q, or ESC
			return "quit", nil
		}
	}
}

// TextInput prompts for text input and returns the entered text
func TextInput(prompt string) (string, error) {
	// Restore normal terminal mode for text input
	SetSttyCooked()
	fmt.Print(ShowCursor)
	
	fmt.Printf("%s%s%s: ", ColorBold, prompt, ColorReset)
	
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("error reading input: %w", err)
	}
	
	return strings.TrimSpace(input), nil
}

// CloseReasonDialog shows a menu to select close reason
func CloseReasonDialog() (string, error) {
	reasons := []MenuItem{
		{ID: 1, Content: "Close as completed - Issue was successfully resolved"},
		{ID: 2, Content: "Close as not planned - Issue will not be implemented"},
		{ID: 3, Content: "Close as duplicate - Issue is a duplicate of another"},
	}
	
	menu := NewInteractiveMenu("How do you want to close this issue?", reasons)
	menu.showHelp = false // Hide navigation help for simplicity
	
	selected, action, err := menu.Run()
	if err != nil {
		return "", fmt.Errorf("close reason selection error: %w", err)
	}
	
	if selected == nil || action == "quit" {
		return "", fmt.Errorf("close cancelled")
	}
	
	switch selected.ID {
	case 1:
		return "completed", nil
	case 2:
		return "not planned", nil
	case 3:
		return "duplicate", nil
	default:
		return "", fmt.Errorf("invalid selection")
	}
}
