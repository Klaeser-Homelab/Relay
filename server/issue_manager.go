package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Issue represents a development issue captured by the user
type Issue struct {
	ID        int       `json:"id"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Status    string    `json:"status"`    // "captured", "in-progress", "done", "archived"
	Category  string    `json:"category"`  // "feature", "bug", "refactor", "tech-debt"
}

// IssueManager manages issues for a specific project
type IssueManager struct {
	issues      []Issue
	projectPath string
	nextID      int
	dataFile    string
}

// NewIssueManager creates a new IssueManager for the specified project
func NewIssueManager(projectPath string) (*IssueManager, error) {
	// Create .relay directory if it doesn't exist
	relayDir := filepath.Join(projectPath, ".relay")
	if err := os.MkdirAll(relayDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create .relay directory: %w", err)
	}

	dataFile := filepath.Join(relayDir, "issues.json")

	im := &IssueManager{
		issues:      []Issue{},
		projectPath: projectPath,
		nextID:      1,
		dataFile:    dataFile,
	}

	// Load existing issues from file
	if err := im.loadIssues(); err != nil {
		return nil, fmt.Errorf("failed to load existing issues: %w", err)
	}

	return im, nil
}

// loadIssues loads issues from the JSON file
func (im *IssueManager) loadIssues() error {
	// Check if file exists
	if _, err := os.Stat(im.dataFile); os.IsNotExist(err) {
		// File doesn't exist, start fresh
		return nil
	}

	// Read file
	data, err := ioutil.ReadFile(im.dataFile)
	if err != nil {
		return fmt.Errorf("failed to read issues file: %w", err)
	}

	// Handle empty file
	if len(data) == 0 {
		return nil
	}

	// Parse JSON
	var issues []Issue
	if err := json.Unmarshal(data, &issues); err != nil {
		return fmt.Errorf("failed to parse issues JSON: %w", err)
	}

	im.issues = issues

	// Set nextID to be greater than the highest existing ID
	if len(issues) > 0 {
		maxID := 0
		for _, issue := range issues {
			if issue.ID > maxID {
				maxID = issue.ID
			}
		}
		im.nextID = maxID + 1
	}

	return nil
}

// saveIssues saves issues to the JSON file
func (im *IssueManager) saveIssues() error {
	// Convert to JSON
	data, err := json.MarshalIndent(im.issues, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal issues to JSON: %w", err)
	}

	// Write to temporary file first for atomic operation
	tempFile := im.dataFile + ".tmp"
	if err := ioutil.WriteFile(tempFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write temporary issues file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tempFile, im.dataFile); err != nil {
		os.Remove(tempFile) // Clean up temp file
		return fmt.Errorf("failed to save issues file: %w", err)
	}

	return nil
}

// AddIssue adds a new issue
func (im *IssueManager) AddIssue(content string) (*Issue, error) {
	// Validate content
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, fmt.Errorf("issue content cannot be empty")
	}

	if len(content) > 1000 {
		return nil, fmt.Errorf("issue content too long (max 1000 characters)")
	}

	// Create new issue
	issue := Issue{
		ID:        im.nextID,
		Content:   content,
		Timestamp: time.Now(),
		Status:    "captured",
		Category:  categorizeIssue(content),
	}

	// Add to issues slice
	im.issues = append(im.issues, issue)
	im.nextID++

	// Save to file
	if err := im.saveIssues(); err != nil {
		// Rollback in-memory change
		im.issues = im.issues[:len(im.issues)-1]
		im.nextID--
		return nil, fmt.Errorf("failed to save issue: %w", err)
	}

	return &issue, nil
}

// GetIssue retrieves an issue by ID
func (im *IssueManager) GetIssue(id int) (*Issue, error) {
	for i := range im.issues {
		if im.issues[i].ID == id {
			return &im.issues[i], nil
		}
	}
	return nil, fmt.Errorf("issue with ID %d not found", id)
}

// UpdateIssueContent updates the content of an issue
func (im *IssueManager) UpdateIssueContent(id int, content string) error {
	// Validate content
	content = strings.TrimSpace(content)
	if content == "" {
		return fmt.Errorf("issue content cannot be empty")
	}

	if len(content) > 1000 {
		return fmt.Errorf("issue content too long (max 1000 characters)")
	}

	// Find and update issue
	for i := range im.issues {
		if im.issues[i].ID == id {
			oldContent := im.issues[i].Content
			im.issues[i].Content = content
			// Re-categorize the issue based on new content
			im.issues[i].Category = categorizeIssue(content)

			// Save to file
			if err := im.saveIssues(); err != nil {
				// Rollback change
				im.issues[i].Content = oldContent
				return fmt.Errorf("failed to save issue content update: %w", err)
			}

			return nil
		}
	}

	return fmt.Errorf("issue with ID %d not found", id)
}

// UpdateIssueStatus updates the status of an issue
func (im *IssueManager) UpdateIssueStatus(id int, status string) error {
	// Validate status
	validStatuses := []string{"captured", "in-progress", "done", "archived"}
	isValid := false
	for _, validStatus := range validStatuses {
		if status == validStatus {
			isValid = true
			break
		}
	}

	if !isValid {
		return fmt.Errorf("invalid status '%s'. Valid statuses: %s", status, strings.Join(validStatuses, ", "))
	}

	// Find and update issue
	for i := range im.issues {
		if im.issues[i].ID == id {
			oldStatus := im.issues[i].Status
			im.issues[i].Status = status

			// Save to file
			if err := im.saveIssues(); err != nil {
				// Rollback change
				im.issues[i].Status = oldStatus
				return fmt.Errorf("failed to save issue status update: %w", err)
			}

			return nil
		}
	}

	return fmt.Errorf("issue with ID %d not found", id)
}

// DeleteIssue removes an issue by ID
func (im *IssueManager) DeleteIssue(id int) error {
	for i, issue := range im.issues {
		if issue.ID == id {
			// Remove from slice
			im.issues = append(im.issues[:i], im.issues[i+1:]...)

			// Save to file
			if err := im.saveIssues(); err != nil {
				// Rollback change
				im.issues = append(im.issues[:i], append([]Issue{issue}, im.issues[i:]...)...)
				return fmt.Errorf("failed to save after deleting issue: %w", err)
			}

			return nil
		}
	}

	return fmt.Errorf("issue with ID %d not found", id)
}

// ListIssues returns all issues, optionally filtered by status or category
func (im *IssueManager) ListIssues(filterStatus, filterCategory string) []Issue {
	var filteredIssues []Issue

	for _, issue := range im.issues {
		// Apply status filter
		if filterStatus != "" && issue.Status != filterStatus {
			continue
		}

		// Apply category filter
		if filterCategory != "" && issue.Category != filterCategory {
			continue
		}

		filteredIssues = append(filteredIssues, issue)
	}

	// Sort by timestamp (newest first)
	sort.Slice(filteredIssues, func(i, j int) bool {
		return filteredIssues[i].Timestamp.After(filteredIssues[j].Timestamp)
	})

	return filteredIssues
}

// GetStats returns statistics about issues
func (im *IssueManager) GetStats() map[string]int {
	stats := map[string]int{
		"total":       len(im.issues),
		"captured":    0,
		"in-progress": 0,
		"done":        0,
		"archived":    0,
		"feature":     0,
		"bug":         0,
		"refactor":    0,
		"tech-debt":   0,
	}

	for _, issue := range im.issues {
		stats[issue.Status]++
		stats[issue.Category]++
	}

	return stats
}

// categorizeIssue automatically categorizes an issue based on content keywords
func categorizeIssue(content string) string {
	content = strings.ToLower(content)

	// Bug-related keywords
	bugKeywords := []string{"bug", "fix", "error", "issue", "problem", "crash", "fail", "broken", "exception"}
	if containsAny(content, bugKeywords) {
		return "bug"
	}

	// Refactoring keywords
	refactorKeywords := []string{"refactor", "cleanup", "improve", "optimize", "reorganize", "restructure", "simplify"}
	if containsAny(content, refactorKeywords) {
		return "refactor"
	}

	// Technical debt keywords
	debtKeywords := []string{"debt", "technical", "tech debt", "legacy", "deprecated", "upgrade", "migrate"}
	if containsAny(content, debtKeywords) {
		return "tech-debt"
	}

	// Feature keywords (also default)
	featureKeywords := []string{"add", "implement", "create", "feature", "enhance", "extend", "support"}
	if containsAny(content, featureKeywords) {
		return "feature"
	}

	// Default to feature
	return "feature"
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

// getStatusEmoji returns an emoji for the given status
func getStatusEmoji(status string) string {
	switch status {
	case "captured":
		return "💡"
	case "in-progress":
		return "🔄"
	case "done":
		return "✅"
	case "archived":
		return "📦"
	default:
		return "❓"
	}
}

// getCategoryEmoji returns an emoji for the given category
func getCategoryEmoji(category string) string {
	switch category {
	case "feature":
		return "✨"
	case "bug":
		return "🐛"
	case "refactor":
		return "🔧"
	case "tech-debt":
		return "⚠️"
	default:
		return "📝"
	}
}

// FormatIssueList formats a list of issues for display
func (im *IssueManager) FormatIssueList(issues []Issue, showDetails bool) string {
	if len(issues) == 0 {
		return "No issues found."
	}

	var output strings.Builder
	output.WriteString(fmt.Sprintf("📋 Issues (%d):\n", len(issues)))

	for _, issue := range issues {
		statusEmoji := getStatusEmoji(issue.Status)
		categoryEmoji := getCategoryEmoji(issue.Category)
		relativeTime := formatRelativeTime(issue.Timestamp)

		if showDetails {
			output.WriteString(fmt.Sprintf("  #%d %s [%s] %s (%s) - %s\n",
				issue.ID, categoryEmoji, issue.Category, issue.Content, issue.Status, relativeTime))
		} else {
			// Truncate long content for list view
			content := issue.Content
			if len(content) > 60 {
				content = content[:57] + "..."
			}
			output.WriteString(fmt.Sprintf("  #%d %s %s [%s] (%s) - %s\n",
				issue.ID, statusEmoji, content, issue.Category, issue.Status, relativeTime))
		}
	}

	return output.String()
}

// FormatIssueDetails formats detailed information about a single issue
func (im *IssueManager) FormatIssueDetails(issue *Issue) string {
	statusEmoji := getStatusEmoji(issue.Status)
	categoryEmoji := getCategoryEmoji(issue.Category)

	var output strings.Builder
	output.WriteString(fmt.Sprintf("📋 Issue #%d\n", issue.ID))
	output.WriteString(fmt.Sprintf("Content: %s\n", issue.Content))
	output.WriteString(fmt.Sprintf("Status: %s %s\n", statusEmoji, issue.Status))
	output.WriteString(fmt.Sprintf("Category: %s %s\n", categoryEmoji, issue.Category))
	output.WriteString(fmt.Sprintf("Created: %s (%s)\n", issue.Timestamp.Format("2006-01-02 15:04:05"), formatRelativeTime(issue.Timestamp)))

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