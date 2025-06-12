package main

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type IssueListModel struct {
	issueManager   *IssueManager
	configManager  *ConfigManager
	projectName    string
	projectPath    string
	issues         []Issue
	selected       int
	width          int
	height         int
	filterStatus   string
	filterLabel string
	syncing        bool // Track if sync is in progress
}

func NewIssueListModel(issueManager *IssueManager, configManager *ConfigManager, projectName string, projectPath string) IssueListModel {
	issues := issueManager.ListIssues("", "")
	return IssueListModel{
		issueManager:  issueManager,
		configManager: configManager,
		projectName:   projectName,
		projectPath:   projectPath,
		issues:        issues,
		selected:      0,
		width:         80,  // Default width
		height:        24,  // Default height  
		syncing:       false,
	}
}

func (m IssueListModel) Init() tea.Cmd {
	return nil
}

func (m IssueListModel) Update(msg tea.Msg) (IssueListModel, tea.Cmd) {
	switch msg := msg.(type) {
	case SyncResultMsg:
		// Handle sync result
		m.syncing = false
		if msg.Success {
			// Refresh issue list after successful sync
			m.issues = m.issueManager.ListIssues(m.filterStatus, m.filterLabel)
			if m.selected >= len(m.issues) && len(m.issues) > 0 {
				m.selected = len(m.issues) - 1
			}
		}
		// TODO: Show sync status/errors to user
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			// Return to REPL
			return m, SwitchToView(ViewREPL, nil)
			
		case "up", "k":
			if m.selected > 0 {
				m.selected--
			}
			
		case "down", "j":
			if m.selected < len(m.issues)-1 {
				m.selected++
			}
			
		case "enter", " ":
			// Select issue for detailed view
			if len(m.issues) > 0 {
				selectedIssue := m.issues[m.selected]
				return m, SwitchToView(ViewIssueDetail, selectedIssue)
			}
			
		case "c":
			// Chat with all issues in context
			context := &REPLContext{
				Type:        "issues",
				Data:        m.issues,
				DisplayName: "Issues",
			}
			return m, SwitchToView(ViewREPL, context)
			
		case "o":
			// Close issue (default to completed for now)
			if len(m.issues) > 0 {
				selectedIssue := m.issues[m.selected]
				err := m.issueManager.CloseIssue(selectedIssue.ID, "completed")
				if err != nil {
					// Handle error
					return m, nil
				}
				// Refresh issue list
				m.issues = m.issueManager.ListIssues(m.filterStatus, m.filterLabel)
				if m.selected >= len(m.issues) && len(m.issues) > 0 {
					m.selected = len(m.issues) - 1
				}
				return m, nil
			}
			
		case "d":
			// Delete issue with confirmation (backward compatibility)
			if len(m.issues) > 0 {
				selectedIssue := m.issues[m.selected]
				confirmData := ConfirmationData{
					Message: fmt.Sprintf("Delete issue #%d: \"%s\"?", selectedIssue.ID, selectedIssue.Content),
					OnConfirm: func(confirmed bool) tea.Cmd {
						if confirmed {
							err := m.issueManager.DeleteIssue(selectedIssue.ID)
							if err != nil {
								// Handle error
								return nil
							}
							// Refresh issue list
							m.issues = m.issueManager.ListIssues(m.filterStatus, m.filterLabel)
							if m.selected >= len(m.issues) && len(m.issues) > 0 {
								m.selected = len(m.issues) - 1
							}
						}
						return BackToPreviousView()
					},
				}
				return m, SwitchToView(ViewConfirmation, confirmData)
			}
			
		case "s":
			// Sync with GitHub
			return m.handleSync()
			
		case "n":
			// New issue
			inputData := TextInputData{
				Prompt:      "New Issue",
				Placeholder: "Enter issue description...",
				OnComplete: func(content string) tea.Cmd {
					if content != "" {
						_, err := m.issueManager.AddIssue(content)
						if err == nil {
							// Refresh issue list
							m.issues = m.issueManager.ListIssues(m.filterStatus, m.filterLabel)
						}
					}
					return BackToPreviousView()
				},
			}
			return m, SwitchToView(ViewTextInput, inputData)
			
		default:
			// Check for number keys (0-9) to select issues directly
			if len(msg.String()) == 1 && msg.String() >= "0" && msg.String() <= "9" {
				issueNum := int(msg.String()[0] - '0')
				// Find issue with matching ID
				for i, issue := range m.issues {
					if issue.ID == issueNum {
						// Select this issue and open detail view
						m.selected = i
						return m, SwitchToView(ViewIssueDetail, issue)
					}
				}
			}
		}
	}
	
	return m, nil
}

// handleSync performs GitHub synchronization
func (m IssueListModel) handleSync() (IssueListModel, tea.Cmd) {
	if m.syncing {
		// Already syncing, ignore
		return m, nil
	}

	// Mark as syncing
	m.syncing = true

	// Return a command that performs the sync operation
	return m, func() tea.Msg {
		return performSync(m.issueManager, m.configManager, m.projectPath)
	}
}

func (m IssueListModel) View() string {
	var content strings.Builder
	
	// Title
	title := titleStyle.Render(fmt.Sprintf("📋 Issues",  ))
	content.WriteString(title + "\n")
	
	if len(m.issues) == 0 {
		content.WriteString("No issues found. Press 'n' to add your first issue!\n")
	} else {
		// Issue list
		// Use a sensible default if height is not set
		height := m.height
		if height <= 0 {
			height = 24 // Default terminal height
		}
		maxLines := height - 8 // Reserve space for title and help
		if maxLines < 5 {
			maxLines = 5 // Minimum visible lines
		}
		
		startIdx := 0
		endIdx := len(m.issues)
		
		// Calculate visible range
		if len(m.issues) > maxLines {
			if m.selected >= maxLines/2 {
				startIdx = m.selected - maxLines/2
				endIdx = startIdx + maxLines
				if endIdx > len(m.issues) {
					endIdx = len(m.issues)
					startIdx = endIdx - maxLines
				}
			} else {
				endIdx = maxLines
			}
		}
		
		for i := startIdx; i < endIdx; i++ {
			issue := m.issues[i]
			relativeTime := formatRelativeTime(issue.Timestamp)
			
			// Format labels with colors (only if labels exist)
			var line string
			if len(issue.Labels) == 0 {
				// No labels - don't show label section
				line = fmt.Sprintf("#%-2d %s (%s) - %s",
					issue.ID, issue.Content, issue.Status, relativeTime)
			} else {
				var labelParts []string
				for _, label := range issue.Labels {
					switch label {
					case "bug":
						labelParts = append(labelParts, lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true).Render(label))
					case "enhancement":
						labelParts = append(labelParts, lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true).Render(label))
					default:
						labelParts = append(labelParts, label)
					}
				}
				styledLabels := strings.Join(labelParts, ", ")
				line = fmt.Sprintf("#%-2d %s [%s] (%s) - %s",
					issue.ID, issue.Content, styledLabels, issue.Status, relativeTime)
			}
			
			if i == m.selected {
				content.WriteString(selectedIssueStyle.Render("> " + line) + "\n")
			} else {
				content.WriteString(unselectedIssueStyle.Render("  " + line) + "\n")
			}
		}
		
		// Show scroll indicator if needed
		if len(m.issues) > maxLines {
			content.WriteString(helpStyle.Render(fmt.Sprintf("  ... showing %d-%d of %d issues", startIdx+1, endIdx, len(m.issues))) + "\n")
		}
	}
	
	content.WriteString("\n")
	
	content.WriteString("\n")
	
	// Define color styles for different action types
	grayStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Faint(true)
	deleteStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)   // Red for delete
	createStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Bold(true)  // Yellow for new/create
	syncStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("13")).Bold(true)    // Magenta for sync
	backStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Bold(true)     // Gray for back
	
	// Action options displayed horizontally
	chatStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true)     // Blue for chat
	
	// Get sync status for display in parentheses with colored circles
	var syncStatusText string
	var statusCircle string
	if m.syncing {
		syncStatusText = "syncing..."
		statusCircle = lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Render("●") // Yellow circle
	} else {
		switch m.issueManager.GetSyncStatus() {
		case "Synced":
			syncStatusText = "synced"
			statusCircle = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render("●") // Green circle
		case "Sync Error":
			syncStatusText = "error"
			statusCircle = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render("●")  // Red circle
		default:
			syncStatusText = "not synced"
			statusCircle = lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render("●")  // Gray circle
		}
	}
	
	// Build action options with sync status in parentheses including colored circle
	syncAction := syncStyle.Render("s") + " Sync (" + statusCircle + " " + grayStyle.Render(syncStatusText) + ")"

	actionOptions := []string{
		chatStyle.Render("c") + " Chat",
		chatStyle.Render("o") + " Close",
		deleteStyle.Render("d") + " Delete",
		createStyle.Render("n") + " New",
		syncAction,
		backStyle.Render("q") + " Back",
	}
	
	// Join actions with bullet separators
	optionsLine := strings.Join(actionOptions, "  •  ")
	content.WriteString(optionsLine + "\n\n")
	
	// Controls section (minimal)
	content.WriteString(helpStyle.Render("Controls: Use keys above to interact, or press issue number to select") + "\n")
	
	return content.String()
}

// Issue Detail View
type IssueDetailModel struct {
	issue       Issue
	replSession *REPLSession
	selected    int
	width       int
	height      int
	fields      []string
}

func NewIssueDetailModel(issue Issue, replSession *REPLSession) IssueDetailModel {
	fields := []string{
		"Title",
		"Status",
		"Labels",
		"Prompt",
	}
	
	return IssueDetailModel{
		issue:       issue,
		replSession: replSession,
		selected:    0,
		fields:      fields,
	}
}

func (m IssueDetailModel) Init() tea.Cmd {
	return nil
}

func (m IssueDetailModel) Update(msg tea.Msg) (IssueDetailModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			// Back to issue list
			return m, SwitchToView(ViewIssueList, nil)
			
		case "up", "k":
			if m.selected > 0 {
				m.selected--
			}
			
		case "down", "j":
			if m.selected < len(m.fields)-1 {
				m.selected++
			}
			
		case "enter", " ":
			return m.editSelectedField()
			
		case "c":
			// Chat with single issue in context
			context := &REPLContext{
				Type:        "issue", 
				Data:        m.issue,
				DisplayName: fmt.Sprintf("Issue %d: %s", m.issue.ID, m.issue.Content),
			}
			return m, SwitchToView(ViewREPL, context)
			
		case "s":
			// Start with Claude Code shortcut
			return m.handleOpenInClaudeCode()
			
		case "o":
			// Close issue shortcut
			return m.handleClose()
		
		case "d":
			// Delete shortcut
			return m.handleDelete()
		}
	}
	
	return m, nil
}

func (m IssueDetailModel) editSelectedField() (IssueDetailModel, tea.Cmd) {
	switch m.selected {
	case 0: // Title
		return m.handleRename()
	case 1: // Status
		return m.handleEditStatus()
	case 2: // Labels
		return m.handleEditLabels()
	case 3: // Prompt
		return m.handleEditPrompt()
	}
	return m, nil
}

func (m IssueDetailModel) handleOpenInClaudeCode() (IssueDetailModel, tea.Cmd) {
	return m, m.openClaudeCodeTerminal()
}

// generateWorktreeName creates a descriptive worktree directory name
func generateWorktreeName(projectName string, issueID int, issueContent string) string {
	// Create brief description from issue content (max 30 chars, no spaces)
	briefDesc := strings.ReplaceAll(strings.ToLower(issueContent), " ", "-")
	if len(briefDesc) > 30 {
		briefDesc = briefDesc[:30]
	}
	// Remove special characters
	briefDesc = strings.ReplaceAll(briefDesc, "'", "")
	briefDesc = strings.ReplaceAll(briefDesc, "\"", "")
	briefDesc = strings.ReplaceAll(briefDesc, "/", "-")
	briefDesc = strings.ReplaceAll(briefDesc, "\\", "-")
	
	return fmt.Sprintf("../%s-issue-%d-%s", projectName, issueID, briefDesc)
}

// generateBranchName creates a feature branch name
func generateBranchName(issueID int, issueContent string) string {
	// Create brief description from issue content (max 30 chars, no spaces)
	briefDesc := strings.ReplaceAll(strings.ToLower(issueContent), " ", "-")
	if len(briefDesc) > 30 {
		briefDesc = briefDesc[:30]
	}
	// Remove special characters
	briefDesc = strings.ReplaceAll(briefDesc, "'", "")
	briefDesc = strings.ReplaceAll(briefDesc, "\"", "")
	briefDesc = strings.ReplaceAll(briefDesc, "/", "-")
	briefDesc = strings.ReplaceAll(briefDesc, "\\", "-")
	
	return fmt.Sprintf("feature/issue-%d-%s", issueID, briefDesc)
}

func (m IssueDetailModel) openClaudeCodeTerminal() tea.Cmd {
	return func() tea.Msg {
		// Generate worktree and branch names
		worktreeName := generateWorktreeName(m.replSession.currentProject.Name, m.issue.ID, m.issue.Content)
		branchName := generateBranchName(m.issue.ID, m.issue.Content)
		
		// Build the prompt for Claude Code
		var promptBuilder strings.Builder
		
		// Add issue information
		promptBuilder.WriteString(fmt.Sprintf("Issue #%d: %s\n", m.issue.ID, m.issue.Content))
		promptBuilder.WriteString(fmt.Sprintf("Labels: %s\n", strings.Join(m.issue.Labels, ", ")))
		promptBuilder.WriteString(fmt.Sprintf("Status: %s\n", m.issue.Status))
		promptBuilder.WriteString(fmt.Sprintf("Created: %s\n", formatRelativeTime(m.issue.Timestamp)))
		
		// Add custom prompt if exists
		if m.issue.Prompt != "" {
			promptBuilder.WriteString(fmt.Sprintf("\nCustom Prompt:\n%s\n", m.issue.Prompt))
		} else {
			promptBuilder.WriteString("\nPlease help me work on this issue.")
		}
		
		// Add worktree information to the prompt
		promptBuilder.WriteString(fmt.Sprintf("\n\nWorktree Setup:\n"))
		promptBuilder.WriteString(fmt.Sprintf("- Working in isolated worktree: %s\n", worktreeName))
		promptBuilder.WriteString(fmt.Sprintf("- Feature branch: %s\n", branchName))
		promptBuilder.WriteString(fmt.Sprintf("\nWhen you're done:\n"))
		promptBuilder.WriteString(fmt.Sprintf("1. Push: git push -u origin %s\n", branchName))
		promptBuilder.WriteString(fmt.Sprintf("2. Return to main: cd %s\n", m.replSession.currentProject.Path))
		promptBuilder.WriteString(fmt.Sprintf("3. Merge: git checkout main && git merge %s && git push\n", branchName))
		promptBuilder.WriteString(fmt.Sprintf("4. Cleanup: git worktree remove %s && git branch -d %s\n", worktreeName, branchName))
		
		prompt := promptBuilder.String()
		
		// Build claude command - escape quotes properly
		claudeCmd := fmt.Sprintf("claude \"%s\"", strings.ReplaceAll(prompt, "\"", "\\\""))
		
		// Create the worktree setup commands
		worktreeSetupCmd := fmt.Sprintf("cd \"%s\" && git worktree add %s -b %s && cd %s && %s", 
			m.replSession.currentProject.Path, worktreeName, branchName, worktreeName, claudeCmd)
		
		// Open new terminal with worktree setup and claude command based on OS
		var cmd *exec.Cmd
		
		// Detect OS and use appropriate terminal command
		switch runtime.GOOS {
		case "darwin": // macOS
			// Use Terminal.app to open a new window with the worktree setup
			script := fmt.Sprintf("tell application \"Terminal\" to do script \"%s\"", 
				strings.ReplaceAll(worktreeSetupCmd, "\"", "\\\""))
			cmd = exec.Command("osascript", "-e", script)
			
		case "linux":
			// Try common Linux terminals
			terminals := []string{"gnome-terminal", "konsole", "xterm"}
			for _, terminal := range terminals {
				if _, err := exec.LookPath(terminal); err == nil {
					fullCmd := fmt.Sprintf("%s; exec bash", worktreeSetupCmd)
					if terminal == "gnome-terminal" {
						cmd = exec.Command(terminal, "--", "bash", "-c", fullCmd)
					} else if terminal == "konsole" {
						cmd = exec.Command(terminal, "-e", "bash", "-c", fullCmd)
					} else {
						cmd = exec.Command(terminal, "-e", "bash", "-c", fullCmd)
					}
					break
				}
			}
			
		case "windows":
			// Use cmd.exe to open new window with worktree setup
			cmd = exec.Command("cmd", "/c", "start", "cmd", "/k", worktreeSetupCmd)
			
		default:
			// Fallback - just run the setup commands directly
			setupCmd := exec.Command("bash", "-c", worktreeSetupCmd)
			setupCmd.Start()
		}
		
		if cmd != nil {
			err := cmd.Start()
			if err != nil {
				// If opening terminal fails, fallback to running setup directly
				fallbackCmd := exec.Command("bash", "-c", worktreeSetupCmd)
				fallbackCmd.Start()
			}
		}
		
		return nil
	}
}

// triggerSyncAfterUpdate creates a command to sync the issue after a field update
func (m IssueDetailModel) triggerSyncAfterUpdate() tea.Cmd {
	return func() tea.Msg {
		// Mark the specific issue as needing sync
		err := m.replSession.issueManager.UpdateIssueSyncStatus(m.issue.ID, "Syncing")
		if err != nil {
			return nil // Ignore sync errors for now
		}
		
		// Trigger sync in background
		return performSync(m.replSession.issueManager, m.replSession.configManager, m.replSession.currentProject.Path)
	}
}

func (m IssueDetailModel) handleEditPrompt() (IssueDetailModel, tea.Cmd) {
	inputData := TextInputData{
		Prompt:      fmt.Sprintf("Edit prompt for issue #%d", m.issue.ID),
		Placeholder: m.issue.Prompt,
		OnComplete: func(prompt string) tea.Cmd {
			if prompt != m.issue.Prompt {
				err := m.replSession.issueManager.UpdateIssuePrompt(m.issue.ID, prompt)
				if err == nil {
					m.issue.Prompt = prompt
					// Trigger sync after successful update
					go func() {
						performSync(m.replSession.issueManager, m.replSession.configManager, m.replSession.currentProject.Path)
					}()
				}
			}
			return BackToPreviousView()
		},
	}
	return m, SwitchToView(ViewTextInput, inputData)
}

func (m IssueDetailModel) handleRename() (IssueDetailModel, tea.Cmd) {
	inputData := TextInputData{
		Prompt:      fmt.Sprintf("Rename issue #%d", m.issue.ID),
		Placeholder: m.issue.Content,
		OnComplete: func(newContent string) tea.Cmd {
			if newContent != "" {
				err := m.replSession.issueManager.UpdateIssueContent(m.issue.ID, newContent)
				if err == nil {
					m.issue.Content = newContent
					m.issue.Labels = categorizeIssue(newContent)
					// Trigger sync after successful update
					go func() {
						performSync(m.replSession.issueManager, m.replSession.configManager, m.replSession.currentProject.Path)
					}()
				}
			}
			return BackToPreviousView()
		},
	}
	return m, SwitchToView(ViewTextInput, inputData)
}

func (m IssueDetailModel) handleEditStatus() (IssueDetailModel, tea.Cmd) {
	inputData := TextInputData{
		Prompt:      fmt.Sprintf("Edit status for issue #%d", m.issue.ID),
		Placeholder: m.issue.Status,
		OnComplete: func(newStatus string) tea.Cmd {
			if newStatus != "" && newStatus != m.issue.Status {
				err := m.replSession.issueManager.UpdateIssueStatus(m.issue.ID, newStatus)
				if err == nil {
					m.issue.Status = newStatus
					// Trigger sync after successful update
					go func() {
						performSync(m.replSession.issueManager, m.replSession.configManager, m.replSession.currentProject.Path)
					}()
				}
			}
			return BackToPreviousView()
		},
	}
	return m, SwitchToView(ViewTextInput, inputData)
}

func (m IssueDetailModel) handleEditLabels() (IssueDetailModel, tea.Cmd) {
	labelData := LabelEditorData{
		IssueID:       m.issue.ID,
		CurrentLabels: append([]string(nil), m.issue.Labels...), // Copy slice
		OnComplete: func(newLabels []string) tea.Cmd {
			// Update the issue labels
			err := m.replSession.issueManager.UpdateIssueLabels(m.issue.ID, newLabels)
			if err == nil {
				m.issue.Labels = newLabels
				// Trigger sync after successful update
				go func() {
					performSync(m.replSession.issueManager, m.replSession.configManager, m.replSession.currentProject.Path)
				}()
			}
			return BackToPreviousView()
		},
	}
	return m, SwitchToView(ViewLabelEditor, labelData)
}

func (m IssueDetailModel) handleDelete() (IssueDetailModel, tea.Cmd) {
	confirmData := ConfirmationData{
		Message: fmt.Sprintf("Delete issue #%d: \"%s\"?", m.issue.ID, m.issue.Content),
		OnConfirm: func(confirmed bool) tea.Cmd {
			if confirmed {
				err := m.replSession.issueManager.DeleteIssue(m.issue.ID)
				if err == nil {
					return SwitchToView(ViewIssueList, nil)
				}
			}
			return BackToPreviousView()
		},
	}
	return m, SwitchToView(ViewConfirmation, confirmData)
}

func (m IssueDetailModel) handleClose() (IssueDetailModel, tea.Cmd) {
	// For now, default to "completed". In the future, this could show a dialog for close reason selection
	err := m.replSession.issueManager.CloseIssue(m.issue.ID, "completed")
	if err == nil {
		return m, SwitchToView(ViewIssueList, nil)
	}
	return m, BackToPreviousView()
}

func (m IssueDetailModel) View() string {
	var content strings.Builder
	
	// Header
	title := titleStyle.Render(fmt.Sprintf("Issue #%d", m.issue.ID))
	content.WriteString(title + "\n")
	content.WriteString(strings.Repeat("=", 15) + "\n\n")
	
	// Format labels for display - only show if labels exist
	var labelsStr string
	if len(m.issue.Labels) > 0 {
		labelsStr = strings.Join(m.issue.Labels, ", ")
	} else {
		labelsStr = "<press Enter to add labels>"
	}

	// Editable fields with selection highlighting
	fields := []struct{
		label string
		value string
	}{
		{"Title", m.issue.Content},
		{"Status", m.issue.Status},
		{"Labels", labelsStr},
		{"Prompt", func() string {
			if m.issue.Prompt != "" {
				return m.issue.Prompt
			}
			return "<empty - press Enter to set>"
		}()},
	}
	
	for i, field := range fields {
		line := fmt.Sprintf("%s: %s", field.label, field.value)
		if i == m.selected {
			content.WriteString(selectedIssueStyle.Render("> " + line) + "\n")
		} else {
			content.WriteString(unselectedIssueStyle.Render("  " + line) + "\n")
		}
	}
	
	content.WriteString("\n")
	
	// Created timestamp in gray below selection area
	grayStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Faint(true)
	content.WriteString(grayStyle.Render(fmt.Sprintf("Created: %s", formatRelativeTime(m.issue.Timestamp))) + "\n\n")
	
	// Define color styles for different action types
	chatStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true)     // Blue for chat
	openStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)     // Green for start
	deleteStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)    // Red for delete
	backStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Bold(true)      // Gray for back
	
	// Actions displayed on single horizontal line with colored letters
	actionData := []string{
		chatStyle.Render("c") + " Chat",
		openStyle.Render("s") + " Start",
		chatStyle.Render("o") + " Close",
		deleteStyle.Render("d") + " Delete",
		backStyle.Render("q") + " Back",
	}
	
	// Join all actions with proper spacing
	actionsLine := strings.Join(actionData, "  •  ")
	content.WriteString(actionsLine + "\n")
	
	return content.String()
}

// SyncResultMsg represents the result of a sync operation
type SyncResultMsg struct {
	Success bool
	Result  *SyncResult
	Error   error
}

// performSync executes the GitHub synchronization
func performSync(issueManager *IssueManager, configManager *ConfigManager, projectPath string) SyncResultMsg {
	// Create GitHub service and sync manager
	githubService := NewGitHubService(configManager, projectPath)
	syncManager := NewGitHubSyncManager(issueManager, githubService, configManager)

	// Auto-detect and configure GitHub repository if not set
	githubConfig := configManager.GetGitHubConfig()
	if githubConfig.Repository == "" {
		repo, err := githubService.DetectRepository()
		if err != nil {
			return SyncResultMsg{Success: false, Error: fmt.Errorf("failed to detect repository: %w", err)}
		}
		
		err = configManager.UpdateGitHubRepository(repo)
		if err != nil {
			return SyncResultMsg{Success: false, Error: fmt.Errorf("failed to update repository config: %w", err)}
		}
	}

	// Perform bidirectional sync
	result, err := syncManager.SyncBidirectional()
	if err != nil {
		return SyncResultMsg{Success: false, Error: err}
	}

	return SyncResultMsg{Success: true, Result: result}
}