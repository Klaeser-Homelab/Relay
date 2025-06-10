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

// Display renders the menu
func (m *InteractiveMenu) Display() {
	clearScreen()
	fmt.Print(HideCursor)
	
	// Title
	fmt.Printf("%s%s%s\n\n", ColorBold, m.title, ColorReset)
	
	// Menu items
	for i, item := range m.items {
		prefix := "  "
		color := ""
		
		if i == m.selectedIdx {
			prefix = "> "
			color = ColorBlue
		}
		
		fmt.Printf("%s%s%s%s%s\n", prefix, color, item.Content, ColorReset, "")
	}
	
	// Help text
	if m.showHelp {
		fmt.Printf("\n%sControls:%s\n", ColorBold, ColorReset)
		fmt.Printf("  ↑/↓  Navigate\n")
		fmt.Printf("  Enter Select item\n")
		fmt.Printf("  d     Delete item\n")
		fmt.Printf("  q     Quit\n")
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
		case 'd', 'D': // Delete
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

// Display renders the issue action menu
func (m *IssueActionMenu) Display() {
	clearScreen()
	fmt.Print(HideCursor)
	
	// Issue details
	statusEmoji := getStatusEmoji(m.issue.Status)
	categoryEmoji := getCategoryEmoji(m.issue.Category)
	
	fmt.Printf("%sIssue: %s%s\n\n", ColorBold, m.issue.Content, ColorReset)
	fmt.Printf("ID: #%d\n", m.issue.ID)
	fmt.Printf("Status: %s %s\n", statusEmoji, m.issue.Status)
	fmt.Printf("Category: %s %s\n", categoryEmoji, m.issue.Category)
	fmt.Printf("Created: %s\n", formatRelativeTime(m.issue.Timestamp))
	
	// Action menu
	fmt.Printf("\n%sActions:%s\n", ColorBold, ColorReset)
	fmt.Printf("  %sc%s  Chat about this issue with Claude\n", ColorGreen, ColorReset)
	fmt.Printf("  %sr%s  Rename this issue\n", ColorYellow, ColorReset)
	fmt.Printf("  %sd%s  Delete this issue\n", ColorRed, ColorReset)
	fmt.Printf("  %sp%s  Push this issue to GitHub\n", ColorBlue, ColorReset)
	fmt.Printf("  %sq%s  Back to issue list\n", ColorReset, ColorReset)
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