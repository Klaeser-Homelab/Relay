package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

type REPLSession struct {
	currentProject *Project
	projectManager *ProjectManager
	claudeCLI      *ClaudeCLI
	gitOps         *GitOperations
	issueManager   *IssueManager
	running        bool
	logger         *log.Logger
}

// NewREPLSession creates a new REPL session for the specified project
func NewREPLSession(projectName string) (*REPLSession, error) {
	logger := log.New(os.Stdout, "[REPL] ", log.LstdFlags)

	// Initialize project manager
	pm, err := NewProjectManager()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize project manager: %w", err)
	}

	// Get and open the specified project
	project, err := pm.GetProject(projectName)
	if err != nil {
		pm.Close()
		return nil, fmt.Errorf("failed to get project '%s': %w", projectName, err)
	}

	// Open the project (sets as active and changes directory)
	err = pm.OpenProject(projectName)
	if err != nil {
		pm.Close()
		return nil, fmt.Errorf("failed to open project '%s': %w", projectName, err)
	}

	// Initialize Claude CLI for this project
	claudeCLI, err := NewClaudeCLI(true, project.Path) // Use session mode for REPL
	if err != nil {
		pm.Close()
		return nil, fmt.Errorf("failed to initialize Claude CLI: %w", err)
	}

	// Initialize Git operations
	gitOps, err := NewGitOperations(project.Path)
	if err != nil {
		claudeCLI.Close()
		pm.Close()
		return nil, fmt.Errorf("failed to initialize git operations: %w", err)
	}

	// Initialize Issue Manager
	issueManager, err := NewIssueManager(project.Path)
	if err != nil {
		gitOps.Close()
		claudeCLI.Close()
		pm.Close()
		return nil, fmt.Errorf("failed to initialize issue manager: %w", err)
	}

	return &REPLSession{
		currentProject: project,
		projectManager: pm,
		claudeCLI:      claudeCLI,
		gitOps:         gitOps,
		issueManager:   issueManager,
		running:        true,
		logger:         logger,
	}, nil
}

// Start begins the REPL loop
func (r *REPLSession) Start() error {
	defer r.Close()

	// Print welcome message
	r.printWelcome()

	// Create scanner for user input
	scanner := bufio.NewScanner(os.Stdin)

	// Main REPL loop
	for r.running {
		// Print prompt
		fmt.Printf("[%s]> ", r.currentProject.Name)

		// Read user input
		if !scanner.Scan() {
			break // EOF or error
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		// Process the command
		err := r.processCommand(input)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("input error: %w", err)
	}

	return nil
}

// processCommand handles user input and executes the appropriate action
func (r *REPLSession) processCommand(input string) error {
	// Handle REPL commands (starting with /)
	if strings.HasPrefix(input, "/") {
		return r.handleREPLCommand(input)
	}

	// Handle direct Claude commands
	return r.handleClaudeCommand(input)
}

// handleREPLCommand processes commands that start with /
func (r *REPLSession) handleREPLCommand(input string) error {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}

	command := parts[0]

	switch command {
	case "/help", "/h":
		r.printHelp()
	case "/exit", "/quit", "/q":
		fmt.Println("Goodbye!")
		r.running = false
	case "/status":
		return r.handleStatus()
	case "/commit":
		return r.handleCommit()
	case "/push":
		return r.handlePush()
	case "/commit-push":
		return r.handleCommitPush()
	case "/list":
		return r.handleListProjects()
	case "/switch":
		if len(parts) < 2 {
			return fmt.Errorf("usage: /switch <project-name>")
		}
		return r.handleSwitchProject(parts[1])
	case "/pwd":
		r.handlePwd()
	case "/info":
		r.handleProjectInfo()
	case "/issue":
		return r.handleIssueCommand(parts)
	case "/issues":
		return r.handleListIssues(parts)
	default:
		return fmt.Errorf("unknown command: %s (type /help for available commands)", command)
	}

	return nil
}

// handleClaudeCommand sends input directly to Claude
func (r *REPLSession) handleClaudeCommand(input string) error {
	fmt.Printf("🤖 Sending to Claude: %s\n", input)
	
	response, err := r.claudeCLI.SendCommand(input)
	if err != nil {
		return fmt.Errorf("Claude error: %w", err)
	}

	fmt.Printf("Claude: %s\n", response)
	return nil
}

// REPL command handlers
func (r *REPLSession) handleStatus() error {
	fmt.Printf("Project Status for '%s':\n", r.currentProject.Name)
	fmt.Printf("Path: %s\n", r.currentProject.Path)
	fmt.Printf("Last Opened: %s\n", r.currentProject.LastOpened.Format("2006-01-02 15:04:05"))

	// Get git status through Claude
	response, err := r.claudeCLI.SendCommand("Show me the current git status and a brief summary of any changes.")
	if err != nil {
		return fmt.Errorf("failed to get git status: %w", err)
	}

	fmt.Printf("\nGit Status:\n%s\n", response)
	return nil
}

func (r *REPLSession) handleCommit() error {
	fmt.Println("🚀 Starting smart commit...")
	return r.gitOps.SmartCommit()
}

func (r *REPLSession) handlePush() error {
	fmt.Println("📤 Pushing to remote...")
	return r.gitOps.Push("")
}

func (r *REPLSession) handleCommitPush() error {
	fmt.Println("🚀📤 Starting smart commit and push...")
	return r.gitOps.SmartCommitAndPush()
}

func (r *REPLSession) handleListProjects() error {
	projects, err := r.projectManager.ListProjects()
	if err != nil {
		return fmt.Errorf("failed to list projects: %w", err)
	}

	if len(projects) == 0 {
		fmt.Println("No projects found.")
		return nil
	}

	fmt.Println("Projects:")
	for _, project := range projects {
		marker := "  "
		if project.Name == r.currentProject.Name {
			marker = "* "
		}
		fmt.Printf("%s%s - %s\n", marker, project.Name, project.Path)
	}

	return nil
}

func (r *REPLSession) handleSwitchProject(projectName string) error {
	// Get the new project
	newProject, err := r.projectManager.GetProject(projectName)
	if err != nil {
		return fmt.Errorf("failed to get project '%s': %w", projectName, err)
	}

	// Open the new project
	err = r.projectManager.OpenProject(projectName)
	if err != nil {
		return fmt.Errorf("failed to open project '%s': %w", projectName, err)
	}

	// Update current project reference
	r.currentProject = newProject

	// Update Claude CLI working directory
	r.claudeCLI.workingDir = newProject.Path

	// Update Git operations
	r.gitOps.Close()
	r.gitOps, err = NewGitOperations(newProject.Path)
	if err != nil {
		return fmt.Errorf("failed to initialize git operations for new project: %w", err)
	}

	// Update Issue Manager for new project
	r.issueManager, err = NewIssueManager(newProject.Path)
	if err != nil {
		return fmt.Errorf("failed to initialize issue manager for new project: %w", err)
	}

	fmt.Printf("Switched to project '%s' at %s\n", newProject.Name, newProject.Path)
	return nil
}

func (r *REPLSession) handlePwd() {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Current directory: %s\n", r.currentProject.Path)
	} else {
		fmt.Printf("Current directory: %s\n", cwd)
	}
}

func (r *REPLSession) handleProjectInfo() {
	fmt.Printf("Project Information:\n")
	fmt.Printf("  Name: %s\n", r.currentProject.Name)
	fmt.Printf("  Path: %s\n", r.currentProject.Path)
	fmt.Printf("  Created: %s\n", r.currentProject.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("  Last Opened: %s\n", r.currentProject.LastOpened.Format("2006-01-02 15:04:05"))
	if r.currentProject.Description != "" {
		fmt.Printf("  Description: %s\n", r.currentProject.Description)
	}
}

// Helper methods
func (r *REPLSession) printWelcome() {
	fmt.Printf("🚀 Relay REPL v1.0 - Project: %s\n", r.currentProject.Name)
	fmt.Printf("📁 Working directory: %s\n", r.currentProject.Path)
	fmt.Println("Type /help for commands, /exit to quit")
	fmt.Println()
}

func (r *REPLSession) printHelp() {
	fmt.Println("Available REPL Commands:")
	fmt.Println("  /help, /h           Show this help message")
	fmt.Println("  /exit, /quit, /q    Exit the REPL")
	fmt.Println("  /status             Show current project and git status")
	fmt.Println("  /commit             Smart commit with AI-generated message")
	fmt.Println("  /push               Push to current branch")
	fmt.Println("  /commit-push        Smart commit and push")
	fmt.Println("  /list               List all projects")
	fmt.Println("  /switch <name>      Switch to a different project")
	fmt.Println("  /pwd                Show current working directory")
	fmt.Println("  /info               Show detailed project information")
	fmt.Println()
	fmt.Println("Issue Management:")
	fmt.Println("  /issue <content>    Capture a new development issue")
	fmt.Println("  /issues             Interactive issue browser (↑↓ navigate, Enter select, d delete)")
	fmt.Println("  /issues status <s>  Filter issues by status (captured, in-progress, done, archived)")
	fmt.Println("  /issues category <c> Filter issues by category (feature, bug, refactor, tech-debt)")
	fmt.Println("  /issue show <id>    Show detailed information about an issue")
	fmt.Println("  /issue status <id> <status>  Update issue status")
	fmt.Println("  /issue delete <id>  Delete an issue")
	fmt.Println()
	fmt.Println("Interactive Issue Actions (when issue selected):")
	fmt.Println("  c  Chat about issue with Claude")
	fmt.Println("  r  Rename the issue")
	fmt.Println("  d  Delete the issue")
	fmt.Println("  p  Push issue to GitHub")
	fmt.Println()
	fmt.Println("Direct Claude Commands:")
	fmt.Println("  <any text>          Send directly to Claude AI")
	fmt.Println("  Examples:")
	fmt.Println("    analyze this file")
	fmt.Println("    what changed since last commit?")
	fmt.Println("    explain the git history")
	fmt.Println()
}

// handleIssueCommand handles the /issue command and its subcommands
func (r *REPLSession) handleIssueCommand(parts []string) error {
	if len(parts) < 2 {
		return fmt.Errorf("usage: /issue <content> OR /issue status <id> <status> OR /issue show <id> OR /issue delete <id>")
	}

	subcommand := parts[1]

	switch subcommand {
	case "status":
		return r.handleIssueStatus(parts)
	case "show":
		return r.handleIssueShow(parts)
	case "delete":
		return r.handleIssueDelete(parts)
	default:
		// Everything else is treated as issue content
		content := strings.Join(parts[1:], " ")
		return r.handleAddIssue(content)
	}
}

// handleAddIssue adds a new issue
func (r *REPLSession) handleAddIssue(content string) error {
	issue, err := r.issueManager.AddIssue(content)
	if err != nil {
		return fmt.Errorf("failed to add issue: %w", err)
	}

	categoryEmoji := getCategoryEmoji(issue.Category)
	fmt.Printf("📋 Issue #%d captured: \"%s\" %s [%s]\n", issue.ID, issue.Content, categoryEmoji, issue.Category)
	return nil
}

// handleListIssues displays an interactive list of issues
func (r *REPLSession) handleListIssues(parts []string) error {
	var filterStatus, filterCategory string

	// Parse optional filters
	for i := 1; i < len(parts); i++ {
		arg := parts[i]
		if i+1 < len(parts) {
			switch arg {
			case "status":
				filterStatus = parts[i+1]
				i++ // Skip next argument
			case "category":
				filterCategory = parts[i+1]
				i++ // Skip next argument
			}
		}
	}

	issues := r.issueManager.ListIssues(filterStatus, filterCategory)
	
	if len(issues) == 0 {
		if filterStatus != "" || filterCategory != "" {
			fmt.Println("No issues found matching the specified filters.")
		} else {
			fmt.Println("No issues found. Use '/issue <content>' to add your first issue!")
		}
		return nil
	}

	// Convert issues to menu items
	var menuItems []MenuItem
	for _, issue := range issues {
		statusEmoji := getStatusEmoji(issue.Status)
		relativeTime := formatRelativeTime(issue.Timestamp)
		
		// Format: "#1 🔄 Add Redis caching [feature] (in-progress) - 2m ago"
		content := fmt.Sprintf("#%d %s %s [%s] (%s) - %s",
			issue.ID, statusEmoji, issue.Content, issue.Category, issue.Status, relativeTime)
			
		menuItems = append(menuItems, MenuItem{
			ID:      issue.ID,
			Content: content,
			Data:    issue, // Store the full issue for later use
		})
	}

	// Create and run interactive menu
	title := fmt.Sprintf("📋 Issues for %s (%d)", r.currentProject.Name, len(issues))
	menu := NewInteractiveMenu(title, menuItems)
	
	for {
		selectedItem, action, err := menu.Run()
		if err != nil {
			return fmt.Errorf("menu error: %w", err)
		}
		
		if selectedItem == nil || action == "quit" {
			// User quit, return to REPL
			return nil
		}
		
		issue, ok := selectedItem.Data.(Issue)
		if !ok {
			return fmt.Errorf("invalid issue data")
		}
		
		switch action {
		case "select":
			// Open issue action menu
			err := r.handleIssueActionMenu(&issue)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				fmt.Println("Press any key to continue...")
				// Wait for keypress
				SetSttyRaw()
				reader := bufio.NewReader(os.Stdin)
				reader.ReadByte()
				SetSttyCooked()
			}
			// Refresh the issue list after actions
			issues = r.issueManager.ListIssues(filterStatus, filterCategory)
			if len(issues) == 0 {
				fmt.Println("No issues remaining.")
				return nil
			}
			// Update menu items
			menuItems = nil
			for _, issue := range issues {
				statusEmoji := getStatusEmoji(issue.Status)
				// categoryEmoji removed
				relativeTime := formatRelativeTime(issue.Timestamp)
				
				content := fmt.Sprintf("#%d %s %s [%s] (%s) - %s",
					issue.ID, statusEmoji, issue.Content, issue.Category, issue.Status, relativeTime)
					
				menuItems = append(menuItems, MenuItem{
					ID:      issue.ID,
					Content: content,
					Data:    issue,
				})
			}
			menu.items = menuItems
			// Reset selection if needed
			if menu.selectedIdx >= len(menuItems) {
				menu.selectedIdx = len(menuItems) - 1
			}
			
		case "delete":
			// Confirm deletion
			confirmed, err := ConfirmationDialog(fmt.Sprintf("Delete issue #%d: \"%s\"?", issue.ID, issue.Content))
			if err != nil {
				return fmt.Errorf("confirmation error: %w", err)
			}
			
			if confirmed {
				err := r.issueManager.DeleteIssue(issue.ID)
				if err != nil {
					fmt.Printf("Error deleting issue: %v\n", err)
					fmt.Println("Press any key to continue...")
					SetSttyRaw()
					reader := bufio.NewReader(os.Stdin)
					reader.ReadByte()
					SetSttyCooked()
				} else {
					fmt.Printf("✅ Issue #%d deleted successfully.\n", issue.ID)
				}
				
				// Refresh the issue list
				issues = r.issueManager.ListIssues(filterStatus, filterCategory)
				if len(issues) == 0 {
					fmt.Println("No issues remaining.")
					return nil
				}
				// Update menu items
				menuItems = nil
				for _, issue := range issues {
					statusEmoji := getStatusEmoji(issue.Status)
					relativeTime := formatRelativeTime(issue.Timestamp)
					
					content := fmt.Sprintf("#%d %s %s [%s] (%s) - %s",
						issue.ID, statusEmoji, issue.Content, issue.Category, issue.Status, relativeTime)
						
					menuItems = append(menuItems, MenuItem{
						ID:      issue.ID,
						Content: content,
						Data:    issue,
					})
				}
				menu.items = menuItems
				// Reset selection if needed
				if menu.selectedIdx >= len(menuItems) {
					menu.selectedIdx = len(menuItems) - 1
				}
			}
		}
	}
}

// handleIssueActionMenu displays the action menu for a selected issue
func (r *REPLSession) handleIssueActionMenu(issue *Issue) error {
	actionMenu := NewIssueActionMenu(issue)
	
	for {
		action, err := actionMenu.Run()
		if err != nil {
			return fmt.Errorf("action menu error: %w", err)
		}
		
		switch action {
		case "chat":
			err := r.handleIssueChatWithClaude(issue)
			if err != nil {
				return err
			}
			// Refresh the display after chat
			actionMenu.Display()
			
		case "rename":
			err := r.handleIssueRename(issue)
			if err != nil {
				return err
			}
			// Update the action menu with the new issue content
			actionMenu.issue = issue
			actionMenu.Display()
			
		case "delete":
			confirmed, err := ConfirmationDialog(fmt.Sprintf("Delete issue #%d: \"%s\"?", issue.ID, issue.Content))
			if err != nil {
				return fmt.Errorf("confirmation error: %w", err)
			}
			
			if confirmed {
				err := r.issueManager.DeleteIssue(issue.ID)
				if err != nil {
					return fmt.Errorf("failed to delete issue: %w", err)
				}
				fmt.Printf("✅ Issue #%d deleted successfully.\n", issue.ID)
				return nil // Exit to issue list
			}
			// If not confirmed, redisplay the action menu
			actionMenu.Display()
			
		case "push":
			err := r.handleIssuePushToGitHub(issue)
			if err != nil {
				return err
			}
			// Refresh the display after push
			actionMenu.Display()
			
		case "quit":
			return nil // Return to issue list
		}
	}
}

// handleIssueChatWithClaude starts a chat session about the issue
func (r *REPLSession) handleIssueChatWithClaude(issue *Issue) error {
	SetSttyCooked()
	fmt.Print(ShowCursor)
	clearScreen()
	
	fmt.Printf("%sChat about Issue #%d: %s%s\n", ColorBold, issue.ID, issue.Content, ColorReset)
	fmt.Printf("Status: %s %s | Category: %s %s\n\n", 
		getStatusEmoji(issue.Status), issue.Status,
		getCategoryEmoji(issue.Category), issue.Category)
	
	// Start context with issue information
	contextPrompt := fmt.Sprintf("I want to discuss this development issue:\n\nIssue: %s\nStatus: %s\nCategory: %s\n\nPlease help me think through this issue. What would you like to discuss about it?", 
		issue.Content, issue.Status, issue.Category)
	
	// Send initial context to Claude
	response, err := r.claudeCLI.SendCommand(contextPrompt)
	if err != nil {
		return fmt.Errorf("failed to start chat with Claude: %w", err)
	}
	
	fmt.Printf("%sClaude:%s %s\n\n", ColorGreen, ColorReset, response)
	
	// Interactive chat loop
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("%sYou:%s ", ColorBlue, ColorReset)
		input, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("error reading input: %w", err)
		}
		
		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}
		
		// Check for exit commands
		if input == "/quit" || input == "/exit" || input == "/back" {
			break
		}
		
		// Send to Claude
		response, err := r.claudeCLI.SendCommand(input)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}
		
		fmt.Printf("\n%sClaude:%s %s\n\n", ColorGreen, ColorReset, response)
	}
	
	fmt.Println("Chat ended. Returning to issue menu...")
	fmt.Println("Press any key to continue...")
	SetSttyRaw()
	reader = bufio.NewReader(os.Stdin)
	reader.ReadByte()
	
	return nil
}

// handleIssueRename allows renaming an issue
func (r *REPLSession) handleIssueRename(issue *Issue) error {
	newContent, err := TextInput(fmt.Sprintf("Rename issue #%d (current: \"%s\")", issue.ID, issue.Content))
	if err != nil {
		return fmt.Errorf("failed to get new content: %w", err)
	}
	
	if newContent == "" {
		fmt.Println("Rename cancelled.")
		return nil
	}
	
	err = r.issueManager.UpdateIssueContent(issue.ID, newContent)
	if err != nil {
		return fmt.Errorf("failed to rename issue: %w", err)
	}
	
	// Update the issue struct with new content
	issue.Content = newContent
	issue.Category = categorizeIssue(newContent) // Re-categorize
	
	fmt.Printf("✅ Issue #%d renamed to: \"%s\"\n", issue.ID, newContent)
	fmt.Println("Press any key to continue...")
	SetSttyRaw()
	reader := bufio.NewReader(os.Stdin)
	reader.ReadByte()
	
	return nil
}

// handleIssuePushToGitHub pushes the issue to GitHub
func (r *REPLSession) handleIssuePushToGitHub(issue *Issue) error {
	SetSttyCooked()
	fmt.Print(ShowCursor)
	
	fmt.Printf("%sPushing Issue #%d to GitHub...%s\n", ColorYellow, issue.ID, ColorReset)
	
	// Create GitHub issue via Claude
	githubPrompt := fmt.Sprintf("Create a GitHub issue for this development item. Use the GitHub CLI (gh) to create the issue.\n\nTitle: %s\nCategory: %s\nStatus: %s\n\nPlease create the issue and provide the issue URL.", 
		issue.Content, issue.Category, issue.Status)
	
	response, err := r.claudeCLI.SendCommand(githubPrompt)
	if err != nil {
		return fmt.Errorf("failed to push to GitHub: %w", err)
	}
	
	fmt.Printf("\n%sGitHub Push Result:%s\n%s\n", ColorGreen, ColorReset, response)
	fmt.Println("\nPress any key to continue...")
	SetSttyRaw()
	reader := bufio.NewReader(os.Stdin)
	reader.ReadByte()
	
	return nil
}

// handleIssueStatus updates the status of an issue
func (r *REPLSession) handleIssueStatus(parts []string) error {
	if len(parts) < 4 {
		return fmt.Errorf("usage: /issue status <id> <status>")
	}

	id, err := parseIssueID(parts[2])
	if err != nil {
		return err
	}

	status := parts[3]

	err = r.issueManager.UpdateIssueStatus(id, status)
	if err != nil {
		return err
	}

	statusEmoji := getStatusEmoji(status)
	fmt.Printf("✅ Issue #%d status updated to: %s %s\n", id, statusEmoji, status)
	return nil
}

// handleIssueShow displays detailed information about an issue
func (r *REPLSession) handleIssueShow(parts []string) error {
	if len(parts) < 3 {
		return fmt.Errorf("usage: /issue show <id>")
	}

	id, err := parseIssueID(parts[2])
	if err != nil {
		return err
	}

	issue, err := r.issueManager.GetIssue(id)
	if err != nil {
		return err
	}

	output := r.issueManager.FormatIssueDetails(issue)
	fmt.Print(output)
	return nil
}

// handleIssueDelete deletes an issue
func (r *REPLSession) handleIssueDelete(parts []string) error {
	if len(parts) < 3 {
		return fmt.Errorf("usage: /issue delete <id>")
	}

	id, err := parseIssueID(parts[2])
	if err != nil {
		return err
	}

	// Get issue for confirmation message
	issue, err := r.issueManager.GetIssue(id)
	if err != nil {
		return err
	}

	err = r.issueManager.DeleteIssue(id)
	if err != nil {
		return err
	}

	fmt.Printf("🗑️ Deleted issue #%d: \"%s\"\n", id, issue.Content)
	return nil
}

// Close cleans up the REPL session
func (r *REPLSession) Close() error {
	var errors []error

	if r.gitOps != nil {
		if err := r.gitOps.Close(); err != nil {
			errors = append(errors, fmt.Errorf("git operations close error: %w", err))
		}
	}

	if r.claudeCLI != nil {
		if err := r.claudeCLI.Close(); err != nil {
			errors = append(errors, fmt.Errorf("claude CLI close error: %w", err))
		}
	}

	if r.projectManager != nil {
		if err := r.projectManager.Close(); err != nil {
			errors = append(errors, fmt.Errorf("project manager close error: %w", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("cleanup errors: %v", errors)
	}

	return nil
}