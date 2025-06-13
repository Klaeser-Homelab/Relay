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
	issueManager  *IssueManager
	configManager *ConfigManager
	projectName   string
	projectPath   string
	issues        []Issue
	selected      int
	width         int
	height        int
	filterStatus  string
	filterLabel   string
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
		width:         80, // Default width
		height:        24, // Default height
	}
}

func (m IssueListModel) Init() tea.Cmd {
	return nil
}

func (m IssueListModel) Update(msg tea.Msg) (IssueListModel, tea.Cmd) {
	switch msg := msg.(type) {
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
			// Close issue with reason selection
			if len(m.issues) > 0 {
				selectedIssue := m.issues[m.selected]
				closeData := CloseReasonData{
					IssueID:    selectedIssue.Number,
					IssueTitle: selectedIssue.Title,
					OnConfirm: func(reason string) tea.Cmd {
						err := m.issueManager.CloseIssue(selectedIssue.Number, reason)
						if err != nil {
							// Handle error - could show error message
							return BackToPreviousView()
						}
						// Return to issue list
						return SwitchToView(ViewIssueList, nil)
					},
				}
				return m, SwitchToView(ViewCloseReason, closeData)
			}

		case "o":
			// Chat with all issues in context
			context := &REPLContext{
				Type:        "issues",
				Data:        m.issues,
				DisplayName: "Issues",
			}
			return m, SwitchToView(ViewREPL, context)



		case "f":
			// Finish in-progress issue
			if len(m.issues) > 0 {
				selectedIssue := m.issues[m.selected]
				if selectedIssue.State != "closed" && isIssueInProgress(m.projectName, selectedIssue.Number) {
					return m, m.finishIssue(selectedIssue)
				}
			}

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
				// Find issue with matching GitHub number
				for i, issue := range m.issues {
					if issue.Number == issueNum {
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


func (m IssueListModel) View() string {
	var content strings.Builder

	// Title
	title := titleStyle.Render(fmt.Sprintf("📋 Issues"))
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
			relativeTime := formatRelativeTime(issue.CreatedAt)
			isClosed := issue.State == "closed"
			inProgress := !isClosed && isIssueInProgress(m.projectName, issue.Number)

			// Add "in progress" indicator to time if applicable
			timeWithStatus := relativeTime
			if inProgress {
				blueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("4"))
				timeWithStatus = relativeTime + " " + blueStyle.Render("(in progress)")
			}

			// Format labels with colors (only if labels exist)
			var line string
			if len(issue.Labels) == 0 {
				// No labels - don't show label section or status
				line = fmt.Sprintf("#%-2d %s - %s",
					issue.Number, issue.Title, timeWithStatus)
			} else {
				var labelParts []string
				for _, label := range issue.Labels {
					if isClosed {
						// For closed issues, render labels in plain text (will be grayed out below)
						labelParts = append(labelParts, label)
					} else {
						switch label {
						case "bug":
							labelParts = append(labelParts, lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true).Render(label))
						case "enhancement":
							labelParts = append(labelParts, lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true).Render(label))
						default:
							labelParts = append(labelParts, label)
						}
					}
				}
				styledLabels := strings.Join(labelParts, ", ")
				line = fmt.Sprintf("#%-2d %s [%s] - %s",
					issue.Number, issue.Title, styledLabels, timeWithStatus)
			}

			if i == m.selected {
				if isClosed {
					// Apply gray styling to closed issues, even when selected
					graySelectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Bold(true)
					content.WriteString(graySelectedStyle.Render("> "+line) + "\n")
				} else {
					content.WriteString(selectedIssueStyle.Render("> "+line) + "\n")
				}
			} else {
				if isClosed {
					// Apply gray styling to closed issues
					grayStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
					content.WriteString(grayStyle.Render("  "+line) + "\n")
				} else {
					content.WriteString(unselectedIssueStyle.Render("  "+line) + "\n")
				}
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
	deleteStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)  // Red for delete
	createStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Bold(true) // Yellow for new/create
	backStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Bold(true)    // Gray for back

	// Action options displayed horizontally
	chatStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true) // Blue for chat
	finishStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true) // Green for finish

	actionOptions := []string{
		chatStyle.Render("o") + " Chat",
		deleteStyle.Render("c") + " Close",
		createStyle.Render("n") + " New",
		backStyle.Render("q") + " Back",
	}

	// Add finish option if selected issue is in progress
	if len(m.issues) > 0 && m.selected < len(m.issues) {
		selectedIssue := m.issues[m.selected]
		if selectedIssue.State != "closed" && isIssueInProgress(m.projectName, selectedIssue.Number) {
			// Insert finish option before "New"
			actionOptions = []string{
				chatStyle.Render("o") + " Chat",
				deleteStyle.Render("c") + " Close",
				finishStyle.Render("f") + " Finish",
				createStyle.Render("n") + " New",
				backStyle.Render("q") + " Back",
			}
		}
	}

	// Join actions with bullet separators
	optionsLine := strings.Join(actionOptions, "  •  ")
	content.WriteString(optionsLine + "\n\n")

	// Controls section (minimal)
	content.WriteString(helpStyle.Render("Controls: Use keys above to interact, or press issue number to select") + "\n")

	return content.String()
}

// finishIssue handles finishing an in-progress issue
func (m IssueListModel) finishIssue(issue Issue) tea.Cmd {
	return func() tea.Msg {
		// Get the worktree path
		worktreeName := generateWorktreeName(m.projectName, issue.Number)
		branchName := generateBranchName(issue.Number)
		worktreeAbsPath := fmt.Sprintf("%s/%s", m.projectPath, worktreeName)
		
		// Stage all changes
		cmd := exec.Command("git", "add", ".")
		cmd.Dir = worktreeAbsPath
		if err := cmd.Run(); err != nil {
			fmt.Printf("Failed to stage changes: %v\n", err)
			return BackToPreviousView()
		}
		
		// Check if there are changes to commit
		cmd = exec.Command("git", "diff", "--cached", "--quiet")
		cmd.Dir = worktreeAbsPath
		if cmd.Run() == nil {
			fmt.Printf("No changes to commit\n")
			return BackToPreviousView()
		}
		
		// Commit changes
		commitMsg := fmt.Sprintf("feat: resolve issue #%d - %s", issue.Number, issue.Title)
		cmd = exec.Command("git", "commit", "-m", commitMsg)
		cmd.Dir = worktreeAbsPath
		if err := cmd.Run(); err != nil {
			fmt.Printf("Failed to commit changes: %v\n", err)
			return BackToPreviousView()
		}
		
		// Push branch
		cmd = exec.Command("git", "push", "-u", "origin", branchName)
		cmd.Dir = worktreeAbsPath
		if err := cmd.Run(); err != nil {
			fmt.Printf("Failed to push branch: %v\n", err)
			return BackToPreviousView()
		}
		
		// Create pull request using gh CLI
		prTitle := fmt.Sprintf("feat: resolve issue #%d - %s", issue.Number, issue.Title)
		prBody := fmt.Sprintf("Closes #%d\n\n%s", issue.Number, issue.Body)
		cmd = exec.Command("gh", "pr", "create", "--title", prTitle, "--body", prBody)
		cmd.Dir = worktreeAbsPath
		output, err := cmd.Output()
		if err != nil {
			fmt.Printf("Failed to create PR: %v\n", err)
			return BackToPreviousView()
		}
		
		fmt.Printf("✅ Issue #%d finished successfully!\nPR created: %s\n", issue.Number, string(output))
		return BackToPreviousView()
	}
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
		"Body",
		"State",
		"Labels",
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
			// Close issue shortcut
			return m.handleClose()

		case "s":
			// Start with Claude Code shortcut
			return m.handleOpenInClaudeCode()

		case "o":
			// Close issue shortcut
			return m.handleClose()

		case "f":
			// Finish in-progress issue
			if m.issue.State != "closed" && isIssueInProgress(m.replSession.currentProject.Name, m.issue.Number) {
				return m, m.finishIssueDetail()
			}

		case "d":
			// Chat with single issue in context
			context := &REPLContext{
				Type:        "issue",
				Data:        m.issue,
				DisplayName: fmt.Sprintf("Issue %d: %s", m.issue.Number, m.issue.Title),
			}
			return m, SwitchToView(ViewREPL, context)
		}
	}

	return m, nil
}

func (m IssueDetailModel) editSelectedField() (IssueDetailModel, tea.Cmd) {
	switch m.selected {
	case 0: // Title
		return m.handleRename()
	case 1: // Body
		return m.handleEditBody()
	case 2: // State
		return m.handleEditStatus()
	case 3: // Labels
		return m.handleEditLabels()
	}
	return m, nil
}

func (m IssueDetailModel) handleOpenInClaudeCode() (IssueDetailModel, tea.Cmd) {
	return m, m.openClaudeCodeTerminal()
}

// generateWorktreeName creates a descriptive worktree directory name
func generateWorktreeName(projectName string, issueID int) string {
	return fmt.Sprintf("../%s-issue-%d", projectName, issueID)
}

// generateBranchName creates a feature branch name
func generateBranchName(issueID int) string {
	return fmt.Sprintf("feature/issue-%d", issueID)
}

// checkFeatureBranchExists checks if a feature branch exists for the given issue
func checkFeatureBranchExists(issueID int) bool {
	branchName := generateBranchName(issueID)
	cmd := exec.Command("git", "branch", "--list", branchName)
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(output)) != ""
}

// checkWorktreeExists checks if a worktree exists for the given issue
func checkWorktreeExists(projectName string, issueID int) bool {
	worktreeSuffix := fmt.Sprintf("%s-issue-%d", projectName, issueID)
	cmd := exec.Command("git", "worktree", "list")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(output), worktreeSuffix)
}

// isIssueInProgress checks if an issue has both a feature branch and worktree
func isIssueInProgress(projectName string, issueID int) bool {
	return checkFeatureBranchExists(issueID) && checkWorktreeExists(projectName, issueID)
}

func (m IssueDetailModel) openClaudeCodeTerminal() tea.Cmd {
	return func() tea.Msg {
		// Generate worktree and branch names
		worktreeName := generateWorktreeName(m.replSession.currentProject.Name, m.issue.Number)
		branchName := generateBranchName(m.issue.Number)

		// Build the prompt for Claude Code
		var promptBuilder strings.Builder

		// Add issue information
		promptBuilder.WriteString(fmt.Sprintf("Issue #%d: %s\n", m.issue.Number, m.issue.Title))
		promptBuilder.WriteString(fmt.Sprintf("Labels: %s\n", strings.Join(m.issue.Labels, ", ")))
		promptBuilder.WriteString(fmt.Sprintf("Status: %s\n", m.issue.State))
		promptBuilder.WriteString(fmt.Sprintf("Created: %s\n", formatRelativeTime(m.issue.CreatedAt)))

		// Add issue body if exists
		if m.issue.Body != "" {
			promptBuilder.WriteString(fmt.Sprintf("\nDescription:\n%s\n", m.issue.Body))
		}

		// Add planning instruction
		promptBuilder.WriteString(fmt.Sprintf("\nWhat follows is a GitHub issue. Read the Issue and Description and create a plan for implementing it then ask me what I think of the plan. Do not start implementing the plan or make any changes.\n"))

		// Add worktree information to the prompt
		promptBuilder.WriteString(fmt.Sprintf("\nWorktree Setup:\n"))
		promptBuilder.WriteString(fmt.Sprintf("- Working in isolated worktree: %s\n", worktreeName))
		promptBuilder.WriteString(fmt.Sprintf("- Feature branch: %s\n", branchName))
		promptBuilder.WriteString(fmt.Sprintf("\nWhen you're done:\n"))
		promptBuilder.WriteString(fmt.Sprintf("1. Push: git push -u origin %s\n", branchName))
		promptBuilder.WriteString(fmt.Sprintf("2. Return to main: cd %s\n", m.replSession.currentProject.Path))
		promptBuilder.WriteString(fmt.Sprintf("3. Merge: git checkout main && git pull origin main && git merge %s && git push\n", branchName))
		promptBuilder.WriteString(fmt.Sprintf("4. Cleanup: git worktree remove %s && git branch -d %s\n", worktreeName, branchName))

		prompt := promptBuilder.String()

		// Build claude command - escape quotes properly
		claudeCmd := fmt.Sprintf("claude \"%s\"", strings.ReplaceAll(prompt, "\"", "\\\""))

		// Check if worktree and branch already exist
		branchExists := checkFeatureBranchExists(m.issue.Number)
		worktreeExists := checkWorktreeExists(m.replSession.currentProject.Name, m.issue.Number)

		var worktreeSetupCmd string
		if branchExists && worktreeExists {
			// Both exist - just navigate to worktree and start Claude
			// Need to get the absolute path to the existing worktree
			worktreeAbsPath := fmt.Sprintf("%s/%s", m.replSession.currentProject.Path, worktreeName)
			worktreeSetupCmd = fmt.Sprintf("cd \"%s\" && %s",
				worktreeAbsPath, claudeCmd)
		} else {
			// Create worktree and branch from latest main, then start Claude
			worktreeSetupCmd = fmt.Sprintf("cd \"%s\" && git fetch origin main && git checkout main && git pull origin main && git worktree add %s -b %s && cd %s && %s",
				m.replSession.currentProject.Path, worktreeName, branchName, worktreeName, claudeCmd)
		}

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


func (m IssueDetailModel) handleEditBody() (IssueDetailModel, tea.Cmd) {
	inputData := TextInputData{
		Prompt:      fmt.Sprintf("Edit body for issue #%d", m.issue.Number),
		Placeholder: m.issue.Body,
		OnComplete: func(body string) tea.Cmd {
			if body != m.issue.Body {
				err := m.replSession.issueManager.UpdateIssueBody(m.issue.Number, body)
				if err == nil {
					m.issue.Body = body
				}
			}
			return BackToPreviousView()
		},
	}
	return m, SwitchToView(ViewTextInput, inputData)
}

func (m IssueDetailModel) handleRename() (IssueDetailModel, tea.Cmd) {
	inputData := TextInputData{
		Prompt:      fmt.Sprintf("Rename issue #%d", m.issue.Number),
		Placeholder: m.issue.Title,
		OnComplete: func(newTitle string) tea.Cmd {
			if newTitle != "" {
				err := m.replSession.issueManager.UpdateIssueTitle(m.issue.Number, newTitle)
				if err == nil {
					m.issue.Title = newTitle
				}
			}
			return BackToPreviousView()
		},
	}
	return m, SwitchToView(ViewTextInput, inputData)
}

func (m IssueDetailModel) handleEditStatus() (IssueDetailModel, tea.Cmd) {
	inputData := TextInputData{
		Prompt:      fmt.Sprintf("Edit state for issue #%d (open/closed)", m.issue.Number),
		Placeholder: m.issue.State,
		OnComplete: func(newState string) tea.Cmd {
			if newState != "" && newState != m.issue.State && (newState == "open" || newState == "closed") {
				err := m.replSession.issueManager.UpdateIssueStatus(m.issue.Number, newState)
				if err == nil {
					m.issue.State = newState
				}
			} else if newState != "open" && newState != "closed" && newState != "" {
				// Show error for invalid state
				fmt.Printf("💡 GitHub issues only support 'open' and 'closed' states.\n")
			}
			return BackToPreviousView()
		},
	}
	return m, SwitchToView(ViewTextInput, inputData)
}

func (m IssueDetailModel) handleEditLabels() (IssueDetailModel, tea.Cmd) {
	labelData := LabelEditorData{
		IssueID:       m.issue.Number,
		CurrentLabels: append([]string(nil), m.issue.Labels...), // Copy slice
		OnComplete: func(newLabels []string) tea.Cmd {
			// Update the issue labels
			err := m.replSession.issueManager.UpdateIssueLabels(m.issue.Number, newLabels)
			if err == nil {
				m.issue.Labels = newLabels
			}
			return BackToPreviousView()
		},
	}
	return m, SwitchToView(ViewLabelEditor, labelData)
}

func (m IssueDetailModel) handleDelete() (IssueDetailModel, tea.Cmd) {
	confirmData := ConfirmationData{
		Message: fmt.Sprintf("Delete issue #%d: \"%s\"? (This will close the issue on GitHub)", m.issue.Number, m.issue.Title),
		OnConfirm: func(confirmed bool) tea.Cmd {
			if confirmed {
				err := m.replSession.issueManager.DeleteIssue(m.issue.Number)
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
	closeData := CloseReasonData{
		IssueID:    m.issue.Number,
		IssueTitle: m.issue.Title,
		OnConfirm: func(reason string) tea.Cmd {
			err := m.replSession.issueManager.CloseIssue(m.issue.Number, reason)
			if err != nil {
				// Handle error - could show error message
				return BackToPreviousView()
			}
			// Return to issue list
			return SwitchToView(ViewIssueList, nil)
		},
	}
	return m, SwitchToView(ViewCloseReason, closeData)
}

// finishIssueDetail handles finishing an in-progress issue from detail view
func (m IssueDetailModel) finishIssueDetail() tea.Cmd {
	return func() tea.Msg {
		// Get the worktree path
		worktreeName := generateWorktreeName(m.replSession.currentProject.Name, m.issue.Number)
		branchName := generateBranchName(m.issue.Number)
		worktreeAbsPath := fmt.Sprintf("%s/%s", m.replSession.currentProject.Path, worktreeName)
		
		// Stage all changes
		cmd := exec.Command("git", "add", ".")
		cmd.Dir = worktreeAbsPath
		if err := cmd.Run(); err != nil {
			fmt.Printf("Failed to stage changes: %v\n", err)
			return BackToPreviousView()
		}
		
		// Check if there are changes to commit
		cmd = exec.Command("git", "diff", "--cached", "--quiet")
		cmd.Dir = worktreeAbsPath
		if cmd.Run() == nil {
			fmt.Printf("No changes to commit\n")
			return BackToPreviousView()
		}
		
		// Commit changes
		commitMsg := fmt.Sprintf("feat: resolve issue #%d - %s", m.issue.Number, m.issue.Title)
		cmd = exec.Command("git", "commit", "-m", commitMsg)
		cmd.Dir = worktreeAbsPath
		if err := cmd.Run(); err != nil {
			fmt.Printf("Failed to commit changes: %v\n", err)
			return BackToPreviousView()
		}
		
		// Push branch
		cmd = exec.Command("git", "push", "-u", "origin", branchName)
		cmd.Dir = worktreeAbsPath
		if err := cmd.Run(); err != nil {
			fmt.Printf("Failed to push branch: %v\n", err)
			return BackToPreviousView()
		}
		
		// Create pull request using gh CLI
		prTitle := fmt.Sprintf("feat: resolve issue #%d - %s", m.issue.Number, m.issue.Title)
		prBody := fmt.Sprintf("Closes #%d\n\n%s", m.issue.Number, m.issue.Body)
		cmd = exec.Command("gh", "pr", "create", "--title", prTitle, "--body", prBody)
		cmd.Dir = worktreeAbsPath
		output, err := cmd.Output()
		if err != nil {
			fmt.Printf("Failed to create PR: %v\n", err)
			return BackToPreviousView()
		}
		
		fmt.Printf("✅ Issue #%d finished successfully!\nPR created: %s\n", m.issue.Number, string(output))
		return BackToPreviousView()
	}
}

func (m IssueDetailModel) View() string {
	var content strings.Builder

	// Header
	title := titleStyle.Render(fmt.Sprintf("Issue #%d", m.issue.Number))
	content.WriteString(title + "\n")
	content.WriteString(strings.Repeat("=", 15) + "\n\n")

	// Format labels for display - only show if labels exist
	var labelsStr string
	if len(m.issue.Labels) > 0 {
		labelsStr = strings.Join(m.issue.Labels, ", ")
	} else {
		labelsStr = "<press Enter to add labels>"
	}

	fields := []struct {
		label string
		value string
	}{
		{"Title", m.issue.Title},
		{"Body", func() string {
			if m.issue.Body != "" {
				return m.issue.Body
			}
			return "<empty - press Enter to set>"
		}()},
		{"State", func() string {
			if m.issue.State == "closed" && m.issue.ClosedAt != nil {
				return fmt.Sprintf("%s (%s)", m.issue.State, formatRelativeTime(*m.issue.ClosedAt))
			}
			return m.issue.State
		}()},
		{"Labels", labelsStr},
	}

	for i, field := range fields {
		line := fmt.Sprintf("%s: %s", field.label, field.value)
		if i == m.selected {
			content.WriteString(selectedIssueStyle.Render("> "+line) + "\n")
		} else {
			content.WriteString(unselectedIssueStyle.Render("  "+line) + "\n")
		}
	}

	content.WriteString("\n")

	// Created timestamp in gray below selection area
	grayStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Faint(true)
	content.WriteString(grayStyle.Render(fmt.Sprintf("Created: %s", formatRelativeTime(m.issue.CreatedAt))) + "\n")
	content.WriteString(grayStyle.Render(fmt.Sprintf("URL: %s", m.issue.URL)) + "\n\n")

	// Define color styles for different action types
	chatStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true)  // Blue for chat
	openStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)  // Green for start
	deleteStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true) // Red for delete
	backStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Bold(true)   // Gray for back

	// Check if issue is in progress to show appropriate action text
	var startAction string
	var actionData []string
	
	if isIssueInProgress(m.replSession.currentProject.Name, m.issue.Number) {
		startAction = openStyle.Render("s") + " Continue"
		finishStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true) // Green for finish
		actionData = []string{
			chatStyle.Render("d") + " Chat",
			startAction,
			finishStyle.Render("f") + " Finish",
			deleteStyle.Render("c") + " Close",
			backStyle.Render("q") + " Back",
		}
	} else {
		startAction = openStyle.Render("s") + " Start"
		actionData = []string{
			chatStyle.Render("d") + " Chat",
			startAction,
			deleteStyle.Render("c") + " Close",
			backStyle.Render("q") + " Back",
		}
	}

	// Join all actions with proper spacing
	actionsLine := strings.Join(actionData, "  •  ")
	content.WriteString(actionsLine + "\n")

	return content.String()
}

