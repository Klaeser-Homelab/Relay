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
	Number    int       `json:"number"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	State     string    `json:"state"` // "open" or "closed"
	Labels    []string  `json:"labels"`
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
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

// FetchIssues retrieves all issues from the GitHub repository
func (gs *GitHubService) FetchIssues() ([]GitHubIssue, error) {
	config := gs.configManager.GetGitHubConfig()
	if config.Repository == "" {
		return nil, fmt.Errorf("GitHub repository not configured")
	}

	// Fetch both open and closed issues
	cmd := exec.Command("gh", "issue", "list",
		"--repo", config.Repository,
		"--state", "all",
		"--json", "number,title,body,state,labels,url,createdAt,updatedAt",
		"--limit", "1000") // Adjust limit as needed
	cmd.Dir = gs.projectPath

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch GitHub issues: %w", err)
	}

	var rawIssues []map[string]interface{}
	if err := json.Unmarshal(output, &rawIssues); err != nil {
		return nil, fmt.Errorf("failed to parse GitHub issues JSON: %w", err)
	}

	var issues []GitHubIssue
	for _, raw := range rawIssues {
		issue, err := gs.parseGitHubIssue(raw)
		if err != nil {
			continue // Skip malformed issues
		}
		issues = append(issues, issue)
	}

	return issues, nil
}

// CreateIssue creates a new issue on GitHub
func (gs *GitHubService) CreateIssue(title, body string, labels []string) (*GitHubIssue, error) {
	config := gs.configManager.GetGitHubConfig()
	if config.Repository == "" {
		return nil, fmt.Errorf("GitHub repository not configured")
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
		return nil, fmt.Errorf("failed to create GitHub issue: %s", string(output))
	}

	// The output from gh issue create is just the URL
	issueURL := strings.TrimSpace(string(output))

	// Extract issue number from URL (e.g., https://github.com/owner/repo/issues/123)
	parts := strings.Split(issueURL, "/")
	if len(parts) < 1 {
		return nil, fmt.Errorf("invalid issue URL format: %s", issueURL)
	}

	issueNumberStr := parts[len(parts)-1]
	issueNumber := 0
	if _, err := fmt.Sscanf(issueNumberStr, "%d", &issueNumber); err != nil {
		return nil, fmt.Errorf("failed to parse issue number from URL: %s", issueURL)
	}

	// Return basic issue info - we could fetch full details with another call if needed
	issue := GitHubIssue{
		Number: issueNumber,
		Title:  title,
		Body:   body,
		State:  "open",
		Labels: labels,
		URL:    issueURL,
	}

	return &issue, nil
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
		issue.State = state
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

	return issue, nil
}

// MapLocalStatusToGitHub converts local issue status to GitHub state
func (gs *GitHubService) MapLocalStatusToGitHub(localStatus string) string {
	switch localStatus {
	case "done", "archived":
		return "closed"
	default:
		return "open"
	}
}

// MapGitHubStateToLocal converts GitHub state to local issue status
func (gs *GitHubService) MapGitHubStateToLocal(githubState string) string {
	switch githubState {
	case "closed":
		return "done"
	default:
		return "captured"
	}
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
