package main

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type REPLContext struct {
	Type        string      // "issues" or "issue"
	Data        interface{} // Issues list or single Issue
	DisplayName string      // For prompt display
}

type REPLModel struct {
	replSession *REPLSession
	input       string
	history     []string
	output      []string
	cursor      int
	width       int
	height      int
	maxHistory  int
	context     *REPLContext // Current context
}

func NewREPLModel(replSession *REPLSession) REPLModel {
	return REPLModel{
		replSession: replSession,
		input:       "",
		history:     []string{},
		output:      []string{},
		cursor:      0,
		maxHistory:  50,
		context:     nil,
	}
}

// SetContext sets the current REPL context and clears input
func (m *REPLModel) SetContext(context *REPLContext) {
	m.context = context
	m.input = ""      // Clear any existing input
	m.cursor = 0      // Reset history cursor
	
	// Add context message to output
	if context != nil {
		m.output = append(m.output, fmt.Sprintf("📋 Context loaded: %s", context.DisplayName))
	}
}

// ClearContext clears the current REPL context
func (m *REPLModel) ClearContext() {
	m.context = nil
}

func (m REPLModel) Init() tea.Cmd {
	return nil
}

func (m REPLModel) Update(msg tea.Msg) (REPLModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if m.input == "" {
				return m, nil
			}
			
			// Add to history
			m.history = append(m.history, m.input)
			if len(m.history) > m.maxHistory {
				m.history = m.history[1:]
			}
			
			// Add input to output
			m.output = append(m.output, fmt.Sprintf("> %s", m.input))
			
			// Process the command
			return m.processCommand(m.input)
			
		case "backspace":
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
			}
			
		case "ctrl+c":
			return m, tea.Quit
			
		case "up":
			if len(m.history) > 0 && m.cursor < len(m.history) {
				m.cursor++
				m.input = m.history[len(m.history)-m.cursor]
			}
			
		case "down":
			if m.cursor > 1 {
				m.cursor--
				m.input = m.history[len(m.history)-m.cursor]
			} else if m.cursor == 1 {
				m.cursor = 0
				m.input = ""
			}
			
		default:
			// Add character to input
			if len(msg.String()) == 1 {
				m.input += msg.String()
				m.cursor = 0 // Reset history cursor when typing
			}
		}
	}
	
	return m, nil
}

func (m REPLModel) processCommand(input string) (REPLModel, tea.Cmd) {
	// Handle REPL commands (starting with /)
	if strings.HasPrefix(input, "/") {
		return m.handleREPLCommand(input)
	}
	
	// Handle direct Claude commands
	return m.handleClaudeCommand(input)
}

func (m REPLModel) handleREPLCommand(input string) (REPLModel, tea.Cmd) {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		m.output = append(m.output, "Error: empty command")
		m.input = ""
		return m, nil
	}
	
	command := parts[0]
	
	switch command {
	case "/help", "/h":
		m.output = append(m.output, m.getHelpText())
		
	case "/exit", "/quit", "/q":
		return m, tea.Quit
		
	case "/issues":
		// Switch to issues view
		m.input = ""
		return m, SwitchToView(ViewIssueList, nil)
		
	case "/issue":
		if len(parts) < 2 {
			m.output = append(m.output, "Error: usage: /issue <content>")
		} else {
			content := strings.Join(parts[1:], " ")
			return m.handleAddIssue(content)
		}
		
	case "/status":
		return m.handleStatus()
		
	case "/commit":
		m.output = append(m.output, "🚀 Starting smart commit...")
		return m.handleCommit()
		
	case "/push":
		m.output = append(m.output, "📤 Pushing to remote...")
		return m.handlePush()
		
	case "/commit-push":
		m.output = append(m.output, "🚀📤 Starting smart commit and push...")
		return m.handleCommitPush()
		
	case "/list":
		return m.handleListProjects()
		
	case "/pwd":
		m.handlePwd()
		
	case "/info":
		m.handleProjectInfo()
		
	case "/config":
		// Switch to config view
		m.input = ""
		return m, SwitchToView(ViewConfig, nil)
		
	default:
		m.output = append(m.output, fmt.Sprintf("Error: unknown command: %s (type /help for available commands)", command))
	}
	
	m.input = ""
	return m, nil
}

func (m REPLModel) handleClaudeCommand(input string) (REPLModel, tea.Cmd) {
	// Build context-aware command
	var contextualInput string
	if m.context != nil {
		switch m.context.Type {
		case "issues":
			if issues, ok := m.context.Data.([]Issue); ok {
				contextualInput = m.buildIssuesContext(input, issues)
			}
		case "issue":
			if issue, ok := m.context.Data.(Issue); ok {
				contextualInput = m.buildIssueContext(input, issue)
			}
		}
	}
	
	if contextualInput == "" {
		contextualInput = input
	}
	
	m.output = append(m.output, fmt.Sprintf("🤖 Sending to Claude: %s", input))
	
	// In a real implementation, this would be async
	response, err := m.replSession.llmManager.GetExecutingProvider().SendMessage(context.Background(), contextualInput)
	if err != nil {
		m.output = append(m.output, fmt.Sprintf("Claude error: %v", err))
	} else {
		m.output = append(m.output, fmt.Sprintf("Claude: %s", response))
	}
	
	m.input = ""
	return m, nil
}

func (m REPLModel) buildIssuesContext(input string, issues []Issue) string {
	var contextBuilder strings.Builder
	contextBuilder.WriteString("Here are the current issues in this project:\n\n")
	
	for _, issue := range issues {
		var issueDesc string
		if len(issue.Labels) > 0 {
			labelsStr := strings.Join(issue.Labels, ", ")
			issueDesc = fmt.Sprintf("Issue #%d: %s [%s] (%s)\n", 
				issue.ID, issue.Content, labelsStr, issue.Status)
		} else {
			issueDesc = fmt.Sprintf("Issue #%d: %s (%s)\n", 
				issue.ID, issue.Content, issue.Status)
		}
		contextBuilder.WriteString(issueDesc)
	}
	
	contextBuilder.WriteString(fmt.Sprintf("\nUser question about these issues: %s", input))
	return contextBuilder.String()
}

func (m REPLModel) buildIssueContext(input string, issue Issue) string {
	var contextBuilder strings.Builder
	contextBuilder.WriteString("I'm working on this specific issue:\n\n")
	contextBuilder.WriteString(fmt.Sprintf("Issue #%d: %s\n", issue.ID, issue.Content))
	
	if len(issue.Labels) > 0 {
		labelsStr := strings.Join(issue.Labels, ", ")
		contextBuilder.WriteString(fmt.Sprintf("Labels: %s\n", labelsStr))
	}
	
	contextBuilder.WriteString(fmt.Sprintf("Status: %s\n", issue.Status))
	contextBuilder.WriteString(fmt.Sprintf("Created: %s\n", formatRelativeTime(issue.Timestamp)))
	contextBuilder.WriteString(fmt.Sprintf("\nUser question about this issue: %s", input))
	
	return contextBuilder.String()
}

func (m REPLModel) handleAddIssue(content string) (REPLModel, tea.Cmd) {
	issue, err := m.replSession.issueManager.AddIssue(content)
	if err != nil {
		m.output = append(m.output, fmt.Sprintf("Error adding issue: %v", err))
	} else {
		var issueMsg string
		if len(issue.Labels) > 0 {
			labelsStr := strings.Join(issue.Labels, ", ")
			issueMsg = fmt.Sprintf("📋 Issue #%d captured: \"%s\" [%s]", issue.ID, issue.Content, labelsStr)
		} else {
			issueMsg = fmt.Sprintf("📋 Issue #%d captured: \"%s\"", issue.ID, issue.Content)
		}
		m.output = append(m.output, issueMsg)
	}
	
	m.input = ""
	return m, nil
}

func (m REPLModel) handleStatus() (REPLModel, tea.Cmd) {
	m.output = append(m.output, fmt.Sprintf("Project Status for '%s':", m.replSession.currentProject.Name))
	m.output = append(m.output, fmt.Sprintf("Path: %s", m.replSession.currentProject.Path))
	m.output = append(m.output, fmt.Sprintf("Last Opened: %s", m.replSession.currentProject.LastOpened.Format("2006-01-02 15:04:05")))
	
	// Get git status through Claude
	response, err := m.replSession.llmManager.GetExecutingProvider().SendMessage(context.Background(), "Show me the current git status and a brief summary of any changes.")
	if err != nil {
		m.output = append(m.output, fmt.Sprintf("Failed to get git status: %v", err))
	} else {
		m.output = append(m.output, fmt.Sprintf("Git Status:\n%s", response))
	}
	
	m.input = ""
	return m, nil
}

func (m REPLModel) handleCommit() (REPLModel, tea.Cmd) {
	err := m.replSession.gitOps.SmartCommit()
	if err != nil {
		m.output = append(m.output, fmt.Sprintf("Commit failed: %v", err))
	} else {
		m.output = append(m.output, "✅ Smart commit completed successfully")
	}
	
	m.input = ""
	return m, nil
}

func (m REPLModel) handlePush() (REPLModel, tea.Cmd) {
	err := m.replSession.gitOps.Push("")
	if err != nil {
		m.output = append(m.output, fmt.Sprintf("Push failed: %v", err))
	} else {
		m.output = append(m.output, "✅ Push completed successfully")
	}
	
	m.input = ""
	return m, nil
}

func (m REPLModel) handleCommitPush() (REPLModel, tea.Cmd) {
	err := m.replSession.gitOps.SmartCommitAndPush()
	if err != nil {
		m.output = append(m.output, fmt.Sprintf("Commit and push failed: %v", err))
	} else {
		m.output = append(m.output, "✅ Smart commit and push completed successfully")
	}
	
	m.input = ""
	return m, nil
}

func (m REPLModel) handleListProjects() (REPLModel, tea.Cmd) {
	projects, err := m.replSession.projectManager.ListProjects()
	if err != nil {
		m.output = append(m.output, fmt.Sprintf("Failed to list projects: %v", err))
		m.input = ""
		return m, nil
	}
	
	if len(projects) == 0 {
		m.output = append(m.output, "No projects found.")
		m.input = ""
		return m, nil
	}
	
	m.output = append(m.output, "Projects:")
	for _, project := range projects {
		marker := "  "
		if project.Name == m.replSession.currentProject.Name {
			marker = "* "
		}
		m.output = append(m.output, fmt.Sprintf("%s%s - %s", marker, project.Name, project.Path))
	}
	
	m.input = ""
	return m, nil
}

func (m REPLModel) handlePwd() {
	m.output = append(m.output, fmt.Sprintf("Current directory: %s", m.replSession.currentProject.Path))
	m.input = ""
}

func (m REPLModel) handleProjectInfo() {
	m.output = append(m.output, "Project Information:")
	m.output = append(m.output, fmt.Sprintf("  Name: %s", m.replSession.currentProject.Name))
	m.output = append(m.output, fmt.Sprintf("  Path: %s", m.replSession.currentProject.Path))
	m.output = append(m.output, fmt.Sprintf("  Created: %s", m.replSession.currentProject.CreatedAt.Format("2006-01-02 15:04:05")))
	m.output = append(m.output, fmt.Sprintf("  Last Opened: %s", m.replSession.currentProject.LastOpened.Format("2006-01-02 15:04:05")))
	if m.replSession.currentProject.Description != "" {
		m.output = append(m.output, fmt.Sprintf("  Description: %s", m.replSession.currentProject.Description))
	}
	m.input = ""
}

func (m REPLModel) getHelpText() string {
	return `Available REPL Commands:
  /help, /h           Show this help message. Help me trapped.
  /exit, /quit, /q    Exit the REPL
  /status             Show current project and git status
  /commit             Smart commit with AI-generated message
  /push               Push to current branch
  /commit-push        Smart commit and push
  /list               List all projects
  /pwd                Show current working directory
  /info               Show detailed project information
  /config             Open configuration menu

Issue Management:
  /issue <content>    Capture a new development issue
  /issues             Interactive issue browser

Direct Claude Commands:
  <any text>          Send directly to Claude AI
  Examples:
    analyze this file
    what changed since last commit?
    explain the git history`
}

func (m REPLModel) View() string {
	var content strings.Builder
	
	// Title
	title := titleStyle.Render(fmt.Sprintf("🚀 Relay REPL - Project: %s", m.replSession.currentProject.Name))
	content.WriteString(title + "\n")
	content.WriteString(fmt.Sprintf("📁 Working directory: %s\n\n", m.replSession.currentProject.Path))
	
	// Output history (scrollable)
	maxLines := m.height - 6 // Reserve space for title, input, and help
	startIdx := 0
	if len(m.output) > maxLines {
		startIdx = len(m.output) - maxLines
	}
	
	for i := startIdx; i < len(m.output); i++ {
		line := m.output[i]
		// Apply gray styling to past commands (lines starting with ">")
		if strings.HasPrefix(line, "> ") {
			content.WriteString(historyStyle.Render(line) + "\n")
		} else {
			content.WriteString(line + "\n")
		}
	}
	
	// Current input prompt with context
	var promptPrefix string
	if m.context != nil {
		promptPrefix = fmt.Sprintf("[%s] hi there", m.context.DisplayName)
	} else {
		promptPrefix = ""
	}
	
	var prompt string
	if promptPrefix != "" {
		prompt = fmt.Sprintf("%s> %s", promptPrefix, m.input)
	} else {
		prompt = fmt.Sprintf("> %s", m.input)
	}
	
	// Create full-width prompt box with text wrapping
	promptStyle := selectedStyle.
		Width(m.width - 2) // Account for border
		//Align(lipgloss.Left)
	
	content.WriteString("\n" + promptStyle.Render(prompt))
	
	// Help text
	help := helpStyle.Render("Type /help for commands, /issues for issue management, Ctrl+C to quit")
	content.WriteString("\n\n" + help)
	
	return content.String()
}