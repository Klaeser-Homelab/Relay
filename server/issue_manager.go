package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Issue represents a GitHub issue
type Issue struct {
	Number    int        `json:"number"`     // GitHub issue number (primary ID)
	Title     string     `json:"title"`      // GitHub issue title
	Body      string     `json:"body"`       // GitHub issue body/description
	State     string     `json:"state"`      // "open" or "closed"
	Labels    []string   `json:"labels"`     // GitHub labels
	CreatedAt time.Time  `json:"created_at"` // GitHub creation timestamp
	UpdatedAt time.Time  `json:"updated_at"` // GitHub last update timestamp
	ClosedAt  *time.Time `json:"closed_at"`  // GitHub closure timestamp (null for open issues)
	URL       string     `json:"html_url"`   // GitHub issue URL
}

// IssueManager manages GitHub issues for a specific project
type IssueManager struct {
	githubService *GitHubService
	configManager *ConfigManager
	projectPath   string
}

// NewIssueManager creates a new IssueManager for the specified project
func NewIssueManager(projectPath string, configManager *ConfigManager) (*IssueManager, error) {
	githubService := NewGitHubService(configManager, projectPath)
	
	// Verify GitHub authentication
	authenticated, err := githubService.IsAuthenticated()
	if err != nil {
		return nil, fmt.Errorf("failed to check GitHub authentication: %w", err)
	}
	if !authenticated {
		return nil, fmt.Errorf("GitHub CLI is not authenticated. Run 'gh auth login' first")
	}

	// Auto-detect and configure GitHub repository if not set
	githubConfig := configManager.GetGitHubConfig()
	if githubConfig.Repository == "" {
		repo, err := githubService.DetectRepository()
		if err != nil {
			return nil, fmt.Errorf("failed to detect repository: %w", err)
		}
		err = configManager.UpdateGitHubRepository(repo)
		if err != nil {
			return nil, fmt.Errorf("failed to update repository config: %w", err)
		}
	}

	return &IssueManager{
		githubService: githubService,
		configManager: configManager,
		projectPath:   projectPath,
	}, nil
}

// ListIssues returns GitHub issues for the repository (open issues + closed issues from last 24 hours)
func (im *IssueManager) ListIssues(filterStatus, filterLabel string) []Issue {
	issues, err := im.githubService.ListIssues()
	if err != nil {
		fmt.Printf("Error fetching issues from GitHub: %v\n", err)
		return []Issue{}
	}

	var filteredIssues []Issue

	for _, issue := range issues {
		// Apply status filter (GitHub uses "open"/"closed")
		if filterStatus != "" {
			if filterStatus == "open" && issue.State != "open" {
				continue
			}
			if filterStatus == "closed" && issue.State != "closed" {
				continue
			}
		}

		// Apply label filter
		if filterLabel != "" {
			hasLabel := false
			for _, label := range issue.Labels {
				if label == filterLabel {
					hasLabel = true
					break
				}
			}
			if !hasLabel {
				continue
			}
		}

		filteredIssues = append(filteredIssues, issue)
	}

	// Sort issues: open issues first (highest number first), then closed issues (highest number first)
	sort.Slice(filteredIssues, func(i, j int) bool {
		isClosed_i := filteredIssues[i].State == "closed"
		isClosed_j := filteredIssues[j].State == "closed"

		// Open issues come first
		if isClosed_i != isClosed_j {
			return !isClosed_i // Open issues (false) come before closed issues (true)
		}

		// Within the same state group, sort by issue number (highest first)
		return filteredIssues[i].Number > filteredIssues[j].Number
	})

	return filteredIssues
}

// AddIssue creates a new GitHub issue
func (im *IssueManager) AddIssue(title string) (*Issue, error) {
	// Validate title
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, fmt.Errorf("issue title cannot be empty")
	}

	if len(title) > 256 {
		return nil, fmt.Errorf("issue title too long (max 256 characters)")
	}

	// Auto-categorize and create labels
	labels := categorizeIssue(title)

	// Create issue on GitHub
	issueNumber, err := im.githubService.CreateIssue(title, "", labels)
	if err != nil {
		return nil, fmt.Errorf("failed to create GitHub issue: %w", err)
	}

	// Fetch the created issue to get complete data
	issue, err := im.GetIssue(issueNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch created issue: %w", err)
	}

	return issue, nil
}

// GetIssue retrieves a GitHub issue by number
func (im *IssueManager) GetIssue(number int) (*Issue, error) {
	issue, err := im.githubService.GetIssue(number)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch issue #%d from GitHub: %w", number, err)
	}
	return issue, nil
}

// UpdateIssueTitle updates the title of a GitHub issue
func (im *IssueManager) UpdateIssueTitle(number int, title string) error {
	// Validate title
	title = strings.TrimSpace(title)
	if title == "" {
		return fmt.Errorf("issue title cannot be empty")
	}

	if len(title) > 256 {
		return fmt.Errorf("issue title too long (max 256 characters)")
	}

	// Update issue on GitHub
	err := im.githubService.UpdateIssue(number, title, "", "", nil)
	if err != nil {
		return fmt.Errorf("failed to update GitHub issue #%d: %w", number, err)
	}

	return nil
}

// UpdateIssueBody updates the body/description of a GitHub issue
func (im *IssueManager) UpdateIssueBody(number int, body string) error {
	// Update issue on GitHub
	err := im.githubService.UpdateIssue(number, "", body, "", nil)
	if err != nil {
		return fmt.Errorf("failed to update GitHub issue #%d body: %w", number, err)
	}

	return nil
}

// UpdateIssueStatus updates the status of a GitHub issue (open/closed)
func (im *IssueManager) UpdateIssueStatus(number int, state string) error {
	// Validate state (GitHub only supports "open" and "closed")
	if state != "open" && state != "closed" {
		return fmt.Errorf("invalid state '%s'. Valid states: open, closed", state)
	}

	// Update issue on GitHub
	err := im.githubService.UpdateIssue(number, "", "", state, nil)
	if err != nil {
		return fmt.Errorf("failed to update GitHub issue #%d state: %w", number, err)
	}

	return nil
}

// UpdateIssueLabels updates the labels of a GitHub issue
func (im *IssueManager) UpdateIssueLabels(number int, labels []string) error {
	// Update issue on GitHub
	err := im.githubService.UpdateIssue(number, "", "", "", labels)
	if err != nil {
		return fmt.Errorf("failed to update GitHub issue #%d labels: %w", number, err)
	}

	return nil
}

// DeleteIssue deletes a GitHub issue
func (im *IssueManager) DeleteIssue(number int) error {
	// Note: GitHub doesn't support deleting issues via API, so we close it instead
	err := im.UpdateIssueStatus(number, "closed")
	if err != nil {
		return fmt.Errorf("failed to close issue #%d (GitHub doesn't support deletion): %w", number, err)
	}
	
	// Add a comment indicating this was meant to be deleted
	commentErr := im.githubService.AddComment(number, "This issue was marked for deletion and has been closed instead.")
	if commentErr != nil {
		fmt.Printf("Warning: Failed to add deletion comment to GitHub issue #%d: %v\n", number, commentErr)
	}
	
	return nil
}

// CloseIssue closes a GitHub issue with a specific completion status
func (im *IssueManager) CloseIssue(number int, closeReason string) error {
	// Validate close reason
	validReasons := []string{"completed", "not planned", "duplicate"}
	isValid := false
	for _, validReason := range validReasons {
		if closeReason == validReason {
			isValid = true
			break
		}
	}

	if !isValid {
		return fmt.Errorf("invalid close reason '%s'. Valid reasons: %s", closeReason, strings.Join(validReasons, ", "))
	}

	// Close the issue on GitHub
	err := im.githubService.UpdateIssue(number, "", "", "closed", nil)
	if err != nil {
		return fmt.Errorf("failed to close GitHub issue #%d: %w", number, err)
	}

	// Add a comment with the close reason
	comment := fmt.Sprintf("Closed as: %s", closeReason)
	commentErr := im.githubService.AddComment(number, comment)
	if commentErr != nil {
		// Log error but don't fail the close operation
		fmt.Printf("Warning: Failed to add close reason comment to GitHub issue #%d: %v\n", number, commentErr)
	}

	return nil
}

// GetSyncStatus returns sync status (always "Synced" since we're using GitHub as source of truth)
func (im *IssueManager) GetSyncStatus() string {
	return "Synced"
}


// GetStats returns statistics about GitHub issues
func (im *IssueManager) GetStats() map[string]int {
	issues := im.ListIssues("", "")
	
	stats := map[string]int{
		"total":  len(issues),
		"open":   0,
		"closed": 0,
		"bug":    0,
		"enhancement": 0,
	}

	for _, issue := range issues {
		stats[issue.State]++
		// Count each label
		for _, label := range issue.Labels {
			if _, exists := stats[label]; !exists {
				stats[label] = 0
			}
			stats[label]++
		}
	}

	return stats
}

// categorizeIssue automatically categorizes an issue based on content keywords
func categorizeIssue(content string) []string {
	content = strings.ToLower(content)

	// Bug-related keywords
	bugKeywords := []string{"bug", "fix", "error", "issue", "problem", "crash", "fail", "broken", "exception"}
	if containsAny(content, bugKeywords) {
		return []string{"bug"}
	}

	// Default to enhancement for everything else
	return []string{"enhancement"}
}

// containsAny checks if the text contains any of the given keywords
func containsAny(text string, keywords []string) bool {
	for _, keyword := range keywords {
		if strings.Contains(text, keyword) {
			return true
		}
	}
	return false
}

// formatRelativeTime formats a timestamp as relative time (e.g., "2m ago", "1d ago")
func formatRelativeTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	if diff < time.Minute {
		return "just now"
	} else if diff < time.Hour {
		minutes := int(diff.Minutes())
		return fmt.Sprintf("%dm ago", minutes)
	} else if diff < 24*time.Hour {
		hours := int(diff.Hours())
		return fmt.Sprintf("%dh ago", hours)
	} else if diff < 7*24*time.Hour {
		days := int(diff.Hours() / 24)
		return fmt.Sprintf("%dd ago", days)
	} else {
		return t.Format("Jan 2")
	}
}

// getStatusEmoji returns an emoji for the given GitHub status
func getStatusEmoji(state string) string {
	switch state {
	case "open":
		return "🔓"
	case "closed":
		return "✅"
	default:
		return "❓"
	}
}

// getLabelEmoji returns an emoji for the given label
func getLabelEmoji(label string) string {
	switch label {
	case "enhancement":
		return "✨"
	case "bug":
		return "🐛"
	default:
		return "📝"
	}
}

// FormatIssueList formats a list of GitHub issues for display
func (im *IssueManager) FormatIssueList(issues []Issue, showDetails bool) string {
	if len(issues) == 0 {
		return "No issues found."
	}

	var output strings.Builder
	output.WriteString(fmt.Sprintf("📋 Issues (%d):\n", len(issues)))

	for _, issue := range issues {
		statusEmoji := getStatusEmoji(issue.State)
		relativeTime := formatRelativeTime(issue.CreatedAt)

		// Format labels
		labelsStr := strings.Join(issue.Labels, ", ")
		if labelsStr == "" {
			labelsStr = "no labels"
		}

		if showDetails {
			output.WriteString(fmt.Sprintf("  #%d [%s] %s (%s) - %s\n",
				issue.Number, labelsStr, issue.Title, issue.State, relativeTime))
		} else {
			// Truncate long title for list view
			title := issue.Title
			if len(title) > 60 {
				title = title[:57] + "..."
			}
			output.WriteString(fmt.Sprintf("  #%d %s %s [%s] (%s) - %s\n",
				issue.Number, statusEmoji, title, labelsStr, issue.State, relativeTime))
		}
	}

	return output.String()
}

// FormatIssueDetails formats detailed information about a single GitHub issue
func (im *IssueManager) FormatIssueDetails(issue *Issue) string {
	statusEmoji := getStatusEmoji(issue.State)

	// Format labels
	labelsStr := strings.Join(issue.Labels, ", ")
	if labelsStr == "" {
		labelsStr = "no labels"
	}

	var output strings.Builder
	output.WriteString(fmt.Sprintf("📋 Issue #%d\n", issue.Number))
	output.WriteString(fmt.Sprintf("Title: %s\n", issue.Title))
	if issue.Body != "" {
		output.WriteString(fmt.Sprintf("Description: %s\n", issue.Body))
	}
	output.WriteString(fmt.Sprintf("Status: %s %s\n", statusEmoji, issue.State))
	output.WriteString(fmt.Sprintf("Labels: %s\n", labelsStr))
	output.WriteString(fmt.Sprintf("Created: %s (%s)\n", issue.CreatedAt.Format("2006-01-02 15:04:05"), formatRelativeTime(issue.CreatedAt)))
	if issue.ClosedAt != nil {
		output.WriteString(fmt.Sprintf("Closed: %s (%s)\n", issue.ClosedAt.Format("2006-01-02 15:04:05"), formatRelativeTime(*issue.ClosedAt)))
	}
	output.WriteString(fmt.Sprintf("URL: %s\n", issue.URL))

	return output.String()
}

// parseIssueID parses an issue ID from a string, with helpful error messages
func parseIssueID(idStr string) (int, error) {
	if idStr == "" {
		return 0, fmt.Errorf("issue ID is required")
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, fmt.Errorf("invalid issue ID '%s' (must be a number)", idStr)
	}

	if id <= 0 {
		return 0, fmt.Errorf("issue ID must be positive")
	}

	return id, nil
}
