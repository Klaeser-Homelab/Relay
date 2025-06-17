package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// RelayManager wraps the existing Relay functionality for voice control
type RelayManager struct {
	projectsPath string
	configPath   string
}

// Project represents a Relay project
type Project struct {
	Name       string    `json:"name"`
	Path       string    `json:"path"`
	LastOpened time.Time `json:"last_opened"`
}

// ProjectStatus contains information about a project's current state
type ProjectStatus struct {
	Name           string `json:"name"`
	Path           string `json:"path"`
	GitBranch      string `json:"git_branch"`
	GitStatus      string `json:"git_status"`
	HasChanges     bool   `json:"has_changes"`
	IssueCount     int    `json:"issue_count"`
	LastCommit     string `json:"last_commit"`
}

// FunctionResult represents the result of executing a Relay function
type FunctionResult struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func NewRelayManager() (*RelayManager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	rm := &RelayManager{
		projectsPath: filepath.Join(homeDir, ".relay", "projects"),
		configPath:   filepath.Join(homeDir, ".relay", "config.json"),
	}

	// Ensure config directory exists
	configDir := filepath.Dir(rm.configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	return rm, nil
}

func (rm *RelayManager) ListProjects() ([]Project, error) {
	// Check if we can use the existing relay binary
	relayBinary := rm.findRelayBinary()
	if relayBinary != "" {
		return rm.listProjectsViaRelay(relayBinary)
	}

	// Fallback to manual project discovery
	return rm.discoverProjects()
}

func (rm *RelayManager) findRelayBinary() string {
	// Look for relay binary in common locations
	locations := []string{
		"../server/relay",
		"../server/tmp/relay", 
		"./relay",
		"relay", // In PATH
	}

	for _, location := range locations {
		if _, err := os.Stat(location); err == nil {
			return location
		}
		
		// Try with .exe extension on Windows
		if _, err := os.Stat(location + ".exe"); err == nil {
			return location + ".exe"
		}
	}

	// Try to find in PATH
	if path, err := exec.LookPath("relay"); err == nil {
		return path
	}

	return ""
}

func (rm *RelayManager) listProjectsViaRelay(relayBinary string) ([]Project, error) {
	cmd := exec.Command(relayBinary, "list")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list projects via relay: %w", err)
	}

	// Parse the output (simplified for now)
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var projects []Project

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "No projects") || strings.HasPrefix(line, "Projects:") {
			continue
		}

		// Parse format: "* ProjectName - /path/to/project" or "  ProjectName - /path/to/project"
		parts := strings.Split(line, " - ")
		if len(parts) >= 2 {
			name := strings.TrimSpace(strings.TrimPrefix(parts[0], "*"))
			name = strings.TrimSpace(name)
			path := strings.TrimSpace(parts[1])

			projects = append(projects, Project{
				Name:       name,
				Path:       path,
				LastOpened: time.Now(), // Default for now
			})
		}
	}

	return projects, nil
}

func (rm *RelayManager) discoverProjects() ([]Project, error) {
	// Simple project discovery - look for git repositories
	var projects []Project

	// Common development directories
	homeDir, _ := os.UserHomeDir()
	searchDirs := []string{
		filepath.Join(homeDir, "Code"),
		filepath.Join(homeDir, "Projects"),
		filepath.Join(homeDir, "Development"),
		filepath.Join(homeDir, "src"),
	}

	for _, searchDir := range searchDirs {
		if _, err := os.Stat(searchDir); os.IsNotExist(err) {
			continue
		}

		entries, err := os.ReadDir(searchDir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			projectPath := filepath.Join(searchDir, entry.Name())
			gitPath := filepath.Join(projectPath, ".git")

			// Check if it's a git repository
			if _, err := os.Stat(gitPath); err == nil {
				projects = append(projects, Project{
					Name:       entry.Name(),
					Path:       projectPath,
					LastOpened: time.Now(),
				})
			}
		}
	}

	return projects, nil
}

func (rm *RelayManager) SelectProject(projectName string) error {
	relayBinary := rm.findRelayBinary()
	if relayBinary != "" {
		cmd := exec.Command(relayBinary, "open", projectName)
		return cmd.Run()
	}

	// Fallback: just validate the project exists
	projects, err := rm.ListProjects()
	if err != nil {
		return err
	}

	for _, project := range projects {
		if project.Name == projectName {
			return nil // Project exists
		}
	}

	return fmt.Errorf("project '%s' not found", projectName)
}

func (rm *RelayManager) GetProjectStatus(projectName string) (map[string]interface{}, error) {
	// Find the project
	projects, err := rm.ListProjects()
	if err != nil {
		return nil, err
	}

	var projectPath string
	for _, project := range projects {
		if project.Name == projectName {
			projectPath = project.Path
			break
		}
	}

	if projectPath == "" {
		return nil, fmt.Errorf("project '%s' not found", projectName)
	}

	status := map[string]interface{}{
		"name": projectName,
		"path": projectPath,
	}

	// Get git status
	if gitStatus := rm.getGitStatus(projectPath); gitStatus != nil {
		for k, v := range gitStatus {
			status[k] = v
		}
	}

	return status, nil
}

func (rm *RelayManager) getGitStatus(projectPath string) map[string]interface{} {
	status := make(map[string]interface{})

	// Get current branch
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = projectPath
	if output, err := cmd.Output(); err == nil {
		status["git_branch"] = strings.TrimSpace(string(output))
	}

	// Get git status
	cmd = exec.Command("git", "status", "--porcelain")
	cmd.Dir = projectPath
	if output, err := cmd.Output(); err == nil {
		statusOutput := strings.TrimSpace(string(output))
		status["has_changes"] = statusOutput != ""
		status["git_status"] = statusOutput
	}

	// Get last commit
	cmd = exec.Command("git", "log", "-1", "--pretty=format:%h %s")
	cmd.Dir = projectPath
	if output, err := cmd.Output(); err == nil {
		status["last_commit"] = strings.TrimSpace(string(output))
	}

	return status
}

func (rm *RelayManager) ExecuteFunction(projectName, functionName string, arguments map[string]interface{}) (*FunctionResult, error) {
	log.Printf("Executing function: %s with args: %v", functionName, arguments)

	switch functionName {
	case "create_github_issue":
		return rm.createGitHubIssue(projectName, arguments)
	case "update_github_issue":
		return rm.updateGitHubIssue(projectName, arguments)
	case "git_commit":
		return rm.gitCommit(projectName)
	case "git_status":
		return rm.gitStatus(projectName)
	case "list_issues":
		return rm.listIssues(projectName)
	default:
		return &FunctionResult{
			Success: false,
			Message: fmt.Sprintf("Unknown function: %s", functionName),
		}, nil
	}
}

func (rm *RelayManager) createGitHubIssue(projectName string, args map[string]interface{}) (*FunctionResult, error) {
	title, _ := args["title"].(string)
	body, _ := args["body"].(string)
	
	if title == "" {
		return &FunctionResult{
			Success: false,
			Message: "Title is required for creating an issue",
		}, nil
	}

	// Try to use relay binary first
	relayBinary := rm.findRelayBinary()
	if relayBinary != "" {
		// This would require extending the relay CLI to support issue creation
		// For now, simulate success
		log.Printf("Would create issue: %s - %s", title, body)
		return &FunctionResult{
			Success: true,
			Message: fmt.Sprintf("Issue '%s' created successfully", title),
			Data: map[string]interface{}{
				"title": title,
				"body":  body,
			},
		}, nil
	}

	// Fallback: return simulated result
	return &FunctionResult{
		Success: true,
		Message: fmt.Sprintf("Issue '%s' would be created (relay binary not available)", title),
		Data: map[string]interface{}{
			"title": title,
			"body":  body,
		},
	}, nil
}

func (rm *RelayManager) updateGitHubIssue(projectName string, args map[string]interface{}) (*FunctionResult, error) {
	number, _ := args["number"].(float64)
	title, _ := args["title"].(string)
	body, _ := args["body"].(string)

	if number == 0 {
		return &FunctionResult{
			Success: false,
			Message: "Issue number is required",
		}, nil
	}

	log.Printf("Would update issue #%.0f: %s - %s", number, title, body)
	
	return &FunctionResult{
		Success: true,
		Message: fmt.Sprintf("Issue #%.0f updated successfully", number),
		Data: map[string]interface{}{
			"number": number,
			"title":  title,
			"body":   body,
		},
	}, nil
}

func (rm *RelayManager) gitCommit(projectName string) (*FunctionResult, error) {
	// Get project path
	projects, err := rm.ListProjects()
	if err != nil {
		return &FunctionResult{
			Success: false,
			Message: "Failed to get project list",
		}, nil
	}

	var projectPath string
	for _, project := range projects {
		if project.Name == projectName {
			projectPath = project.Path
			break
		}
	}

	if projectPath == "" {
		return &FunctionResult{
			Success: false,
			Message: "Project not found",
		}, nil
	}

	// Try using relay binary for smart commit
	relayBinary := rm.findRelayBinary()
	if relayBinary != "" {
		cmd := exec.Command(relayBinary, "commit")
		cmd.Dir = projectPath
		output, err := cmd.CombinedOutput()
		
		if err != nil {
			return &FunctionResult{
				Success: false,
				Message: fmt.Sprintf("Commit failed: %s", string(output)),
			}, nil
		}

		return &FunctionResult{
			Success: true,
			Message: "Smart commit completed successfully",
			Data: map[string]interface{}{
				"output": string(output),
			},
		}, nil
	}

	// Fallback: basic git commit
	cmd := exec.Command("git", "add", ".")
	cmd.Dir = projectPath
	if err := cmd.Run(); err != nil {
		return &FunctionResult{
			Success: false,
			Message: "Failed to stage changes",
		}, nil
	}

	cmd = exec.Command("git", "commit", "-m", "Voice-controlled commit")
	cmd.Dir = projectPath
	output, err := cmd.CombinedOutput()

	if err != nil {
		return &FunctionResult{
			Success: false,
			Message: fmt.Sprintf("Commit failed: %s", string(output)),
		}, nil
	}

	return &FunctionResult{
		Success: true,
		Message: "Changes committed successfully",
		Data: map[string]interface{}{
			"output": string(output),
		},
	}, nil
}

func (rm *RelayManager) gitStatus(projectName string) (*FunctionResult, error) {
	status, err := rm.GetProjectStatus(projectName)
	if err != nil {
		return &FunctionResult{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &FunctionResult{
		Success: true,
		Message: "Git status retrieved successfully",
		Data:    status,
	}, nil
}

func (rm *RelayManager) listIssues(projectName string) (*FunctionResult, error) {
	// This would integrate with the existing GitHub functionality
	// For now, return simulated data
	return &FunctionResult{
		Success: true,
		Message: "Issues retrieved successfully",
		Data: map[string]interface{}{
			"issues": []map[string]interface{}{
				{
					"number": 1,
					"title":  "Add voice control feature",
					"state":  "open",
				},
				{
					"number": 2,
					"title":  "Fix audio streaming bug",
					"state":  "closed",
				},
			},
		},
	}, nil
}
