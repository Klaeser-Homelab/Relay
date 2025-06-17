package shared

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"sort"
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

// FileChange represents a file change in git
type FileChange struct {
	Status   string // A=added, M=modified, D=deleted, R=renamed
	FilePath string
	NewPath  string // for renamed files
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

// CreatePullRequest creates a pull request using gh CLI
func (gs *GitHubService) CreatePullRequest(title, body string) (string, error) {
	cmd := exec.Command("gh", "pr", "create",
		"--repo", gs.repository,
		"--title", title,
		"--body", body)
	cmd.Dir = gs.workingDir

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to create pull request: %w", err)
	}
	
	return strings.TrimSpace(string(output)), nil
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

// GetCommitFileChanges gets the list of changed files between main and current branch
func GetCommitFileChanges(workingDir, baseBranch string) ([]FileChange, error) {
	if baseBranch == "" {
		baseBranch = "main"
	}

	cmd := exec.Command("git", "diff", "--name-status", baseBranch+"...HEAD")
	cmd.Dir = workingDir

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get file changes: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var changes []FileChange

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		change := FileChange{
			Status:   parts[0],
			FilePath: parts[1],
		}

		// Handle renamed files
		if strings.HasPrefix(parts[0], "R") && len(parts) >= 3 {
			change.NewPath = parts[2]
		}

		changes = append(changes, change)
	}

	return changes, nil
}

// GetCommitMessages gets commit messages for the current branch
func GetCommitMessages(workingDir, baseBranch string) ([]string, error) {
	if baseBranch == "" {
		baseBranch = "main"
	}

	cmd := exec.Command("git", "log", "--pretty=format:%s", baseBranch+"..HEAD")
	cmd.Dir = workingDir

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get commit messages: %w", err)
	}

	messages := strings.Split(strings.TrimSpace(string(output)), "\n")
	var result []string

	for _, msg := range messages {
		if msg != "" {
			result = append(result, msg)
		}
	}

	return result, nil
}

// GenerateChangeSummary creates a human-readable summary of file changes
func GenerateChangeSummary(changes []FileChange) string {
	if len(changes) == 0 {
		return "No changes detected"
	}

	// Categorize changes
	categories := map[string][]FileChange{
		"Added":    {},
		"Modified": {},
		"Deleted":  {},
		"Renamed":  {},
	}

	for _, change := range changes {
		switch {
		case strings.HasPrefix(change.Status, "A"):
			categories["Added"] = append(categories["Added"], change)
		case strings.HasPrefix(change.Status, "M"):
			categories["Modified"] = append(categories["Modified"], change)
		case strings.HasPrefix(change.Status, "D"):
			categories["Deleted"] = append(categories["Deleted"], change)
		case strings.HasPrefix(change.Status, "R"):
			categories["Renamed"] = append(categories["Renamed"], change)
		}
	}

	var summary strings.Builder

	// Generate summary based on file types and patterns
	summary.WriteString("## Changes\n\n")

	for category, files := range categories {
		if len(files) == 0 {
			continue
		}

		summary.WriteString(fmt.Sprintf("### %s Files\n", category))

		// Group files by type
		fileGroups := groupFilesByType(files)
		
		// Sort groups for consistent output
		var groupNames []string
		for groupName := range fileGroups {
			groupNames = append(groupNames, groupName)
		}
		sort.Strings(groupNames)

		for _, groupName := range groupNames {
			groupFiles := fileGroups[groupName]
			summary.WriteString(fmt.Sprintf("- **%s**\n", groupName))
			
			for _, file := range groupFiles {
				if category == "Renamed" && file.NewPath != "" {
					summary.WriteString(fmt.Sprintf("  - %s → %s\n", file.FilePath, file.NewPath))
				} else {
					summary.WriteString(fmt.Sprintf("  - %s\n", file.FilePath))
				}
			}
		}
		summary.WriteString("\n")
	}

	return summary.String()
}

// groupFilesByType groups files by their type/purpose
func groupFilesByType(files []FileChange) map[string][]FileChange {
	groups := make(map[string][]FileChange)

	for _, file := range files {
		filePath := file.FilePath
		if file.NewPath != "" {
			filePath = file.NewPath
		}

		groupName := categorizeFile(filePath)
		groups[groupName] = append(groups[groupName], file)
	}

	return groups
}

// categorizeFile determines the category/type of a file
func categorizeFile(filePath string) string {
	lower := strings.ToLower(filePath)

	// Test files
	if strings.Contains(lower, "test") || strings.Contains(lower, "spec") {
		return "Tests"
	}

	// Documentation
	if strings.HasSuffix(lower, ".md") || strings.HasSuffix(lower, ".txt") || strings.Contains(lower, "readme") || strings.Contains(lower, "doc") {
		return "Documentation"
	}

	// Configuration files
	if strings.HasSuffix(lower, ".json") || strings.HasSuffix(lower, ".yaml") || strings.HasSuffix(lower, ".yml") || 
	   strings.HasSuffix(lower, ".toml") || strings.HasSuffix(lower, ".ini") || strings.Contains(lower, "config") {
		return "Configuration"
	}

	// Database/Schema files
	if strings.Contains(lower, "migration") || strings.Contains(lower, "schema") || strings.HasSuffix(lower, ".sql") {
		return "Database"
	}

	// Frontend files
	if strings.HasSuffix(lower, ".tsx") || strings.HasSuffix(lower, ".jsx") || strings.Contains(lower, "component") {
		return "React Components"
	}

	if strings.HasSuffix(lower, ".css") || strings.HasSuffix(lower, ".scss") || strings.HasSuffix(lower, ".sass") {
		return "Styles"
	}

	// Backend files
	if strings.HasSuffix(lower, ".go") {
		return "Go Code"
	}

	if strings.HasSuffix(lower, ".js") || strings.HasSuffix(lower, ".ts") {
		return "JavaScript/TypeScript"
	}

	if strings.HasSuffix(lower, ".py") {
		return "Python Code"
	}

	// Other source files
	if strings.HasSuffix(lower, ".java") || strings.HasSuffix(lower, ".cpp") || strings.HasSuffix(lower, ".c") || 
	   strings.HasSuffix(lower, ".h") || strings.HasSuffix(lower, ".hpp") {
		return "Source Code"
	}

	// Assets
	if strings.HasSuffix(lower, ".png") || strings.HasSuffix(lower, ".jpg") || strings.HasSuffix(lower, ".jpeg") || 
	   strings.HasSuffix(lower, ".gif") || strings.HasSuffix(lower, ".svg") {
		return "Images"
	}

	// Default
	return "Other"
}

// GitAdd stages all changes for commit
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

// GitCommit commits staged changes with a message
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