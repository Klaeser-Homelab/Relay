package shared

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// GitHubIssue represents a GitHub issue
type GitHubIssue struct {
	Number    int        `json:"number"`
	Title     string     `json:"title"`
	Body      string     `json:"body"`
	State     string     `json:"state"`
	Labels    []string   `json:"labels"`
	URL       string     `json:"url"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	ClosedAt  *time.Time `json:"closed_at"`
}

// GitHubService handles GitHub operations via gh CLI
type GitHubService struct {
	workingDir string
	repository string
}

// NewGitHubService creates a new GitHub service
func NewGitHubService(workingDir string) (*GitHubService, error) {
	service := &GitHubService{
		workingDir: workingDir,
	}
	
	// Auto-detect repository
	repo, err := service.DetectRepository()
	if err != nil {
		return nil, fmt.Errorf("failed to detect repository: %w", err)
	}
	service.repository = repo
	
	return service, nil
}

// DetectRepository detects the GitHub repository from git remotes
func (gs *GitHubService) DetectRepository() (string, error) {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = gs.workingDir

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get git remote: %w", err)
	}

	remoteURL := strings.TrimSpace(string(output))

	// Parse GitHub repository from various URL formats
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

// GetIssue retrieves a GitHub issue by number
func (gs *GitHubService) GetIssue(number int) (*GitHubIssue, error) {
	cmd := exec.Command("gh", "issue", "view", strconv.Itoa(number),
		"--repo", gs.repository,
		"--json", "number,title,body,state,labels,url,createdAt,updatedAt,closedAt")
	cmd.Dir = gs.workingDir

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch GitHub issue #%d: %w", number, err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(output, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse GitHub issue JSON: %w", err)
	}

	issue, err := gs.parseGitHubIssue(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to parse GitHub issue: %w", err)
	}

	return &issue, nil
}

// UpdateIssueBody updates the body of a GitHub issue
func (gs *GitHubService) UpdateIssueBody(number int, newBody string) error {
	cmd := exec.Command("gh", "issue", "edit", strconv.Itoa(number),
		"--repo", gs.repository,
		"--body", newBody)
	cmd.Dir = gs.workingDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to update issue #%d body: %s", number, string(output))
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

// GitAdd stages all changes in the git repository
func GitAdd(workingDir string) error {
	// First find the git repository root
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = workingDir
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to find git root: %v", err)
	}
	
	gitRoot := strings.TrimSpace(string(output))
	
	// Stage all changes from git root
	cmd = exec.Command("git", "add", ".")
	cmd.Dir = gitRoot
	
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to stage changes: %s", string(output))
	}
	
	return nil
}

// GitCommit creates a commit with the given message
func GitCommit(workingDir, message string) error {
	cmd := exec.Command("git", "commit", "-m", message)
	cmd.Dir = workingDir
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to commit changes: %s", string(output))
	}
	
	return nil
}

// GitPush pushes the current branch to origin
func GitPush(workingDir string) error {
	cmd := exec.Command("git", "push", "origin", "HEAD")
	cmd.Dir = workingDir
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to push changes: %s", string(output))
	}
	
	return nil
}

// CreatePullRequest creates a pull request using gh CLI
func (gs *GitHubService) CreatePullRequest(title, body string) (string, error) {
	cmd := exec.Command("gh", "pr", "create",
		"--repo", gs.repository,
		"--title", title,
		"--body", body)
	cmd.Dir = gs.workingDir
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to create pull request: %s", string(output))
	}
	
	return strings.TrimSpace(string(output)), nil
}