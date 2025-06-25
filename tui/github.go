package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// GitHubService handles GitHub API operations via GitHub CLI
type GitHubService struct {
	configManager *ConfigManager
	projectPath   string
}

// GitHubIssue represents a GitHub issue from the API
type GitHubIssue struct {
	Number    int        `json:"number"`
	Title     string     `json:"title"`
	Body      string     `json:"body"`
	State     string     `json:"state"` // "open" or "closed"
	Labels    []string   `json:"labels"`
	URL       string     `json:"url"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	ClosedAt  *time.Time `json:"closed_at"`
}

// NewGitHubService creates a new GitHub service instance
func NewGitHubService(configManager *ConfigManager, projectPath string) *GitHubService {
	return &GitHubService{
		configManager: configManager,
		projectPath:   projectPath,
	}
}

// IsAuthenticated checks if GitHub CLI is authenticated
func (gs *GitHubService) IsAuthenticated() (bool, error) {
	cmd := exec.Command("gh", "auth", "status")
	cmd.Dir = gs.projectPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, nil // Not authenticated
	}

	// Check if output contains "Logged in"
	return strings.Contains(string(output), "Logged in"), nil
}

// DetectRepository attempts to detect the GitHub repository from git remotes
func (gs *GitHubService) DetectRepository() (string, error) {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = gs.projectPath

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get git remote: %w", err)
	}

	remoteURL := strings.TrimSpace(string(output))

	// Parse GitHub repository from various URL formats
	// SSH: git@github.com:owner/repo.git
	// HTTPS: https://github.com/owner/repo.git
	var repo string

	sshRegex := regexp.MustCompile(`git@github\.com:([^/]+/[^/]+)(?:\.git)?$`)
	httpsRegex := regexp.MustCompile(`https://github\.com/([^/]+/[^/]+)(?:\.git)?$`)

	if matches := sshRegex.FindStringSubmatch(remoteURL); len(matches) > 1 {
		repo = matches[1]
	} else if matches := httpsRegex.FindStringSubmatch(remoteURL); len(matches) > 1 {
		repo = matches[1]
	} else {
		return "", fmt.Errorf("could not parse GitHub repository from remote URL: %s", remoteURL)
	}

	// Remove .git suffix if present
	repo = strings.TrimSuffix(repo, ".git")

	return repo, nil
}

// ValidateRepository checks if the repository exists and is accessible
func (gs *GitHubService) ValidateRepository(repo string) error {
	cmd := exec.Command("gh", "repo", "view", repo)
	cmd.Dir = gs.projectPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("repository validation failed: %s", string(output))
	}

	return nil
}

// ListIssues retrieves all issues from the GitHub repository
func (gs *GitHubService) ListIssues() ([]Issue, error) {
	config := gs.configManager.GetGitHubConfig()
	if config.Repository == "" {
		return nil, fmt.Errorf("GitHub repository not configured")
	}

	// Calculate date for 24 hours ago for filtering closed issues
	oneDayAgo := time.Now().Add(-24 * time.Hour).Format("2006-01-02")
	
	// Fetch open issues and closed issues from last 24 hours
	// We need to make two separate calls since GitHub search syntax doesn't support OR for state
	var allIssues []Issue
	
	// First, get all open issues
	openCmd := exec.Command("gh", "issue", "list",
		"--repo", config.Repository,
		"--state", "open",
		"--json", "number,title,body,state,labels,url,createdAt,updatedAt,closedAt",
		"--limit", "1000")
	openCmd.Dir = gs.projectPath
	
	openOutput, err := openCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch open GitHub issues: %w", err)
	}
	
	var openIssues []map[string]interface{}
	if err := json.Unmarshal(openOutput, &openIssues); err != nil {
		return nil, fmt.Errorf("failed to parse open GitHub issues JSON: %w", err)
	}
	
	// Convert open issues
	for _, raw := range openIssues {
		githubIssue, err := gs.parseGitHubIssue(raw)
		if err != nil {
			continue // Skip malformed issues
		}
		issue := Issue{
			Number:    githubIssue.Number,
			Title:     githubIssue.Title,
			Body:      githubIssue.Body,
			State:     githubIssue.State,
			Labels:    githubIssue.Labels,
			CreatedAt: githubIssue.CreatedAt,
			UpdatedAt: githubIssue.UpdatedAt,
			ClosedAt:  githubIssue.ClosedAt,
			URL:       githubIssue.URL,
		}
		allIssues = append(allIssues, issue)
	}
	
	// Second, get closed issues from last 24 hours using search
	closedCmd := exec.Command("gh", "issue", "list",
		"--repo", config.Repository,
		"--state", "closed",
		"--search", fmt.Sprintf("closed:>%s", oneDayAgo),
		"--json", "number,title,body,state,labels,url,createdAt,updatedAt,closedAt",
		"--limit", "1000")
	closedCmd.Dir = gs.projectPath

	closedOutput, err := closedCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch closed GitHub issues: %w", err)
	}

	var closedIssues []map[string]interface{}
	if err := json.Unmarshal(closedOutput, &closedIssues); err != nil {
		return nil, fmt.Errorf("failed to parse closed GitHub issues JSON: %w", err)
	}

	// Convert closed issues
	for _, raw := range closedIssues {
		githubIssue, err := gs.parseGitHubIssue(raw)
		if err != nil {
			continue // Skip malformed issues
		}
		issue := Issue{
			Number:    githubIssue.Number,
			Title:     githubIssue.Title,
			Body:      githubIssue.Body,
			State:     githubIssue.State,
			Labels:    githubIssue.Labels,
			CreatedAt: githubIssue.CreatedAt,
			UpdatedAt: githubIssue.UpdatedAt,
			ClosedAt:  githubIssue.ClosedAt,
			URL:       githubIssue.URL,
		}
		allIssues = append(allIssues, issue)
	}

	return allIssues, nil
}

// CreateIssue creates a new issue on GitHub and returns the issue number
func (gs *GitHubService) CreateIssue(title, body string, labels []string) (int, error) {
	config := gs.configManager.GetGitHubConfig()
	if config.Repository == "" {
		return 0, fmt.Errorf("GitHub repository not configured")
	}

	args := []string{"issue", "create", "--repo", config.Repository, "--title", title}

	// Always provide body, even if empty (required by gh CLI in non-interactive mode)
	if body == "" {
		body = " " // Use single space instead of empty string
	}
	args = append(args, "--body", body)

	if len(labels) > 0 {
		args = append(args, "--label", strings.Join(labels, ","))
	}

	cmd := exec.Command("gh", args...)
	cmd.Dir = gs.projectPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("failed to create GitHub issue: %s", string(output))
	}

	// The output from gh issue create is just the URL
	issueURL := strings.TrimSpace(string(output))

	// Extract issue number from URL (e.g., https://github.com/owner/repo/issues/123)
	parts := strings.Split(issueURL, "/")
	if len(parts) < 1 {
		return 0, fmt.Errorf("invalid issue URL format: %s", issueURL)
	}

	issueNumberStr := parts[len(parts)-1]
	issueNumber := 0
	if _, err := fmt.Sscanf(issueNumberStr, "%d", &issueNumber); err != nil {
		return 0, fmt.Errorf("failed to parse issue number from URL: %s", issueURL)
	}

	return issueNumber, nil
}

// UpdateIssue updates an existing GitHub issue
func (gs *GitHubService) UpdateIssue(number int, title, body string, state string, labels []string) error {
	config := gs.configManager.GetGitHubConfig()
	if config.Repository == "" {
		return fmt.Errorf("GitHub repository not configured")
	}

	issueStr := strconv.Itoa(number)

	// Update title and body
	if title != "" || body != "" {
		args := []string{"issue", "edit", issueStr, "--repo", config.Repository}
		if title != "" {
			args = append(args, "--title", title)
		}
		if body != "" {
			args = append(args, "--body", body)
		}

		cmd := exec.Command("gh", args...)
		cmd.Dir = gs.projectPath

		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to update issue %d: %s", number, string(output))
		}
	}

	// Update labels
	if len(labels) > 0 {
		// First, remove all existing labels by getting current labels and removing them
		args := []string{"issue", "edit", issueStr, "--repo", config.Repository, "--remove-label", "*"}
		cmd := exec.Command("gh", args...)
		cmd.Dir = gs.projectPath
		if _, err := cmd.CombinedOutput(); err != nil {
			// Ignore errors for removing labels as it might fail if no labels exist
		}

		// Then add the new labels
		args = []string{"issue", "edit", issueStr, "--repo", config.Repository, "--add-label", strings.Join(labels, ",")}
		cmd = exec.Command("gh", args...)
		cmd.Dir = gs.projectPath
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to update labels for issue %d: %s", number, string(output))
		}
	} else {
		// Remove all labels if none specified
		args := []string{"issue", "edit", issueStr, "--repo", config.Repository, "--remove-label", "*"}
		cmd := exec.Command("gh", args...)
		cmd.Dir = gs.projectPath
		if _, err := cmd.CombinedOutput(); err != nil {
			// Ignore errors for removing labels as it might fail if no labels exist
		}
	}

	// Update state (close/reopen)
	if state == "closed" {
		cmd := exec.Command("gh", "issue", "close", issueStr, "--repo", config.Repository)
		cmd.Dir = gs.projectPath
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to close issue %d: %s", number, string(output))
		}
	} else if state == "open" {
		cmd := exec.Command("gh", "issue", "reopen", issueStr, "--repo", config.Repository)
		cmd.Dir = gs.projectPath
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to reopen issue %d: %s", number, string(output))
		}
	}

	return nil
}

// GetIssue retrieves a single GitHub issue by number
func (gs *GitHubService) GetIssue(number int) (*Issue, error) {
	config := gs.configManager.GetGitHubConfig()
	if config.Repository == "" {
		return nil, fmt.Errorf("GitHub repository not configured")
	}

	cmd := exec.Command("gh", "issue", "view", strconv.Itoa(number),
		"--repo", config.Repository,
		"--json", "number,title,body,state,labels,url,createdAt,updatedAt,closedAt")
	cmd.Dir = gs.projectPath

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch GitHub issue #%d: %w", number, err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(output, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse GitHub issue JSON: %w", err)
	}

	githubIssue, err := gs.parseGitHubIssue(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to parse GitHub issue: %w", err)
	}

	// Convert GitHubIssue to Issue
	issue := &Issue{
		Number:    githubIssue.Number,
		Title:     githubIssue.Title,
		Body:      githubIssue.Body,
		State:     githubIssue.State,
		Labels:    githubIssue.Labels,
		CreatedAt: githubIssue.CreatedAt,
		UpdatedAt: githubIssue.UpdatedAt,
		ClosedAt:  githubIssue.ClosedAt,
		URL:       githubIssue.URL,
	}

	return issue, nil
}

// AddComment adds a comment to a GitHub issue
func (gs *GitHubService) AddComment(number int, comment string) error {
	config := gs.configManager.GetGitHubConfig()
	if config.Repository == "" {
		return fmt.Errorf("GitHub repository not configured")
	}

	cmd := exec.Command("gh", "issue", "comment", strconv.Itoa(number),
		"--repo", config.Repository,
		"--body", comment)
	cmd.Dir = gs.projectPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to add comment to GitHub issue #%d: %s", number, string(output))
	}

	return nil
}

// parseGitHubIssue converts raw JSON data to GitHubIssue struct
func (gs *GitHubService) parseGitHubIssue(raw map[string]interface{}) (GitHubIssue, error) {
	issue := GitHubIssue{}

	// Parse number
	if num, ok := raw["number"].(float64); ok {
		issue.Number = int(num)
	} else {
		return issue, fmt.Errorf("invalid issue number")
	}

	// Parse title
	if title, ok := raw["title"].(string); ok {
		issue.Title = title
	}

	// Parse body
	if body, ok := raw["body"].(string); ok {
		issue.Body = body
	}

	// Parse state
	if state, ok := raw["state"].(string); ok {
		issue.State = strings.ToLower(state)
	}

	// Parse URL
	if url, ok := raw["url"].(string); ok {
		issue.URL = url
	}

	// Parse labels
	if labelsRaw, ok := raw["labels"].([]interface{}); ok {
		for _, labelRaw := range labelsRaw {
			if labelMap, ok := labelRaw.(map[string]interface{}); ok {
				if name, ok := labelMap["name"].(string); ok {
					issue.Labels = append(issue.Labels, name)
				}
			}
		}
	}

	// Parse timestamps
	if createdStr, ok := raw["createdAt"].(string); ok {
		if created, err := time.Parse(time.RFC3339, createdStr); err == nil {
			issue.CreatedAt = created
		}
	}

	if updatedStr, ok := raw["updatedAt"].(string); ok {
		if updated, err := time.Parse(time.RFC3339, updatedStr); err == nil {
			issue.UpdatedAt = updated
		}
	}

	// Parse closedAt (can be null for open issues)
	if closedStr, ok := raw["closedAt"].(string); ok && closedStr != "" {
		if closed, err := time.Parse(time.RFC3339, closedStr); err == nil {
			issue.ClosedAt = &closed
		}
	}

	return issue, nil
}

// MapLocalStatusToGitHub converts any status to GitHub state (legacy compatibility)
func (gs *GitHubService) MapLocalStatusToGitHub(status string) string {
	switch status {
	case "closed":
		return "closed"
	default:
		return "open"
	}
}

// MapGitHubStateToLocal returns GitHub state as-is (no longer mapping to local statuses)
func (gs *GitHubService) MapGitHubStateToLocal(githubState string) string {
	return githubState // Return as-is: "open" or "closed"
}

// MapLocalLabelsToGitHub converts local labels to GitHub labels (1:1 mapping)
func (gs *GitHubService) MapLocalLabelsToGitHub(localLabels []string) []string {
	var githubLabels []string
	for _, label := range localLabels {
		switch label {
		case "bug":
			githubLabels = append(githubLabels, "bug")
		case "enhancement":
			githubLabels = append(githubLabels, "enhancement")
		default:
			githubLabels = append(githubLabels, label) // Pass through unknown labels
		}
	}
	return githubLabels
}

// MapGitHubLabelsToLocal converts GitHub labels to local labels
func (gs *GitHubService) MapGitHubLabelsToLocal(githubLabels []string) []string {
	var localLabels []string
	for _, label := range githubLabels {
		switch strings.ToLower(label) {
		case "bug":
			localLabels = append(localLabels, "bug")
		case "enhancement", "feature":
			localLabels = append(localLabels, "enhancement")
		default:
			// Pass through unknown labels as-is to preserve them
			localLabels = append(localLabels, label)
		}
	}

	return localLabels
}
