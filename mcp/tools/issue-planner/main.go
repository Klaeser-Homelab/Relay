package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"shared"
)

// MCP Protocol Types
type MCPRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
}

type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type ToolInfo struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

type UpdateIssuePlanParams struct {
	Plan        string `json:"plan"`
	WorkingDir  string `json:"workingDir,omitempty"`
	IssueNumber int    `json:"issueNumber,omitempty"`
}

type UpdateIssuePlanResult struct {
	Success     bool   `json:"success"`
	IssueNumber int    `json:"issueNumber"`
	Message     string `json:"message"`
}

type FinishParams struct {
	CommitMessage string `json:"commitMessage,omitempty"`
	PrTitle       string `json:"prTitle,omitempty"`
	PrBody        string `json:"prBody,omitempty"`
	WorkingDir    string `json:"workingDir,omitempty"`
	IssueNumber   int    `json:"issueNumber,omitempty"`
}

type FinishResult struct {
	Success     bool   `json:"success"`
	Message     string `json:"message"`
	PullRequest string `json:"pullRequest,omitempty"`
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	
	for scanner.Scan() {
		line := scanner.Text()
		
		var request MCPRequest
		if err := json.Unmarshal([]byte(line), &request); err != nil {
			sendError(request.ID, -32700, "Parse error")
			continue
		}
		
		handleRequest(request)
	}
}

func handleRequest(request MCPRequest) {
	switch request.Method {
	case "initialize":
		handleInitialize(request)
	case "tools/list":
		handleToolsList(request)
	case "tools/call":
		handleToolsCall(request)
	default:
		sendError(request.ID, -32601, "Method not found")
	}
}

func handleInitialize(request MCPRequest) {
	result := map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities": map[string]interface{}{
			"tools": map[string]interface{}{},
		},
		"serverInfo": map[string]interface{}{
			"name":    "issue-planner",
			"version": "1.0.0",
		},
	}
	
	sendResponse(request.ID, result)
}

func handleToolsList(request MCPRequest) {
	tools := []ToolInfo{
		{
			Name:        "update_issue_plan",
			Description: "Updates a GitHub issue with a plan extracted from conversation context",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"plan": map[string]interface{}{
						"type":        "string",
						"description": "The plan content to add to the issue",
					},
					"workingDir": map[string]interface{}{
						"type":        "string",
						"description": "Working directory (optional, defaults to current directory)",
					},
					"issueNumber": map[string]interface{}{
						"type":        "integer",
						"description": "Issue number (optional, will be auto-detected from branch if not provided)",
					},
				},
				"required": []string{"plan"},
			},
		},
		{
			Name:        "finish",
			Description: "Stages, commits, pushes changes and creates a merge request for the current work",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"commitMessage": map[string]interface{}{
						"type":        "string",
						"description": "Commit message for the changes (optional, auto-generated from issue if not provided)",
					},
					"prTitle": map[string]interface{}{
						"type":        "string",
						"description": "Title for the pull request (optional, auto-generated from issue if not provided)",
					},
					"prBody": map[string]interface{}{
						"type":        "string",
						"description": "Body/description for the pull request (optional, auto-generated from context if not provided)",
					},
					"workingDir": map[string]interface{}{
						"type":        "string",
						"description": "Working directory (optional, defaults to current directory)",
					},
					"issueNumber": map[string]interface{}{
						"type":        "integer",
						"description": "Issue number (optional, will be auto-detected from branch if not provided)",
					},
				},
				"required": []string{},
			},
		},
	}
	
	result := map[string]interface{}{
		"tools": tools,
	}
	
	sendResponse(request.ID, result)
}

func handleToolsCall(request MCPRequest) {
	paramsMap, ok := request.Params.(map[string]interface{})
	if !ok {
		sendError(request.ID, -32602, "Invalid params")
		return
	}
	
	toolName, ok := paramsMap["name"].(string)
	if !ok {
		sendError(request.ID, -32602, "Tool name is required")
		return
	}
	
	switch toolName {
	case "update_issue_plan":
		handleUpdateIssuePlan(request, paramsMap)
	case "finish":
		handleFinish(request, paramsMap)
	default:
		sendError(request.ID, -32601, "Tool not found")
	}
}

func handleUpdateIssuePlan(request MCPRequest, paramsMap map[string]interface{}) {
	// Parse arguments
	argsMap, ok := paramsMap["arguments"].(map[string]interface{})
	if !ok {
		sendError(request.ID, -32602, "Invalid arguments")
		return
	}
	
	var params UpdateIssuePlanParams
	
	// Extract plan (required)
	if plan, ok := argsMap["plan"].(string); ok {
		params.Plan = plan
	} else {
		sendError(request.ID, -32602, "Plan is required")
		return
	}
	
	// Extract working directory (optional)
	if workingDir, ok := argsMap["workingDir"].(string); ok {
		params.WorkingDir = workingDir
	} else {
		// Default to current directory
		var err error
		params.WorkingDir, err = os.Getwd()
		if err != nil {
			sendError(request.ID, -32603, fmt.Sprintf("Failed to get working directory: %v", err))
			return
		}
	}
	
	// Extract issue number (optional)
	if issueNum, ok := argsMap["issueNumber"].(float64); ok {
		params.IssueNumber = int(issueNum)
	}
	
	// Execute the tool
	result, err := updateIssuePlan(params)
	if err != nil {
		sendError(request.ID, -32603, err.Error())
		return
	}
	
	sendResponse(request.ID, result)
}

func handleFinish(request MCPRequest, paramsMap map[string]interface{}) {
	// Parse arguments
	argsMap, ok := paramsMap["arguments"].(map[string]interface{})
	if !ok {
		sendError(request.ID, -32602, "Invalid arguments")
		return
	}
	
	var params FinishParams
	
	// Extract commit message (optional)
	if commitMessage, ok := argsMap["commitMessage"].(string); ok {
		params.CommitMessage = commitMessage
	}
	
	// Extract PR title (optional)
	if prTitle, ok := argsMap["prTitle"].(string); ok {
		params.PrTitle = prTitle
	}
	
	// Extract PR body (optional)
	if prBody, ok := argsMap["prBody"].(string); ok {
		params.PrBody = prBody
	}
	
	// Extract issue number (optional)
	if issueNum, ok := argsMap["issueNumber"].(float64); ok {
		params.IssueNumber = int(issueNum)
	}
	
	// Extract working directory (optional)
	if workingDir, ok := argsMap["workingDir"].(string); ok {
		params.WorkingDir = workingDir
	} else {
		// Default to current directory
		var err error
		params.WorkingDir, err = os.Getwd()
		if err != nil {
			sendError(request.ID, -32603, fmt.Sprintf("Failed to get working directory: %v", err))
			return
		}
	}
	
	// Execute the tool
	result, err := finish(params)
	if err != nil {
		sendError(request.ID, -32603, err.Error())
		return
	}
	
	sendResponse(request.ID, result)
}

func updateIssuePlan(params UpdateIssuePlanParams) (*UpdateIssuePlanResult, error) {
	// Determine issue number
	issueNumber := params.IssueNumber
	
	if issueNumber == 0 {
		// Auto-detect from branch
		branchName, err := shared.GetCurrentBranch(params.WorkingDir)
		if err != nil {
			return nil, fmt.Errorf("failed to get current branch: %w", err)
		}
		
		issueNumber, err = shared.ExtractIssueNumberFromBranch(branchName)
		if err != nil {
			return nil, fmt.Errorf("failed to extract issue number from branch '%s': %w", branchName, err)
		}
	}
	
	// Initialize GitHub service
	githubService, err := shared.NewGitHubService(params.WorkingDir)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize GitHub service: %w", err)
	}
	
	// Get current issue
	issue, err := githubService.GetIssue(issueNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue #%d: %w", issueNumber, err)
	}
	
	// Create updated body with plan
	updatedBody := createUpdatedBody(issue.Body, params.Plan)
	
	// Update the issue
	err = githubService.UpdateIssueBody(issueNumber, updatedBody)
	if err != nil {
		return nil, fmt.Errorf("failed to update issue #%d: %w", issueNumber, err)
	}
	
	return &UpdateIssuePlanResult{
		Success:     true,
		IssueNumber: issueNumber,
		Message:     fmt.Sprintf("Successfully updated issue #%d with plan", issueNumber),
	}, nil
}

func createUpdatedBody(existingBody, plan string) string {
	// If there's no existing body, just add the plan
	if existingBody == "" {
		return fmt.Sprintf("## Plan\n\n%s", plan)
	}
	
	// Check if there's already a plan section
	if strings.Contains(strings.ToLower(existingBody), "## plan") {
		// Replace existing plan section
		lines := strings.Split(existingBody, "\n")
		var newLines []string
		inPlanSection := false
		planAdded := false
		
		for _, line := range lines {
			if strings.HasPrefix(strings.ToLower(strings.TrimSpace(line)), "## plan") {
				inPlanSection = true
				newLines = append(newLines, "## Plan")
				newLines = append(newLines, "")
				newLines = append(newLines, plan)
				planAdded = true
				continue
			}
			
			if inPlanSection && strings.HasPrefix(strings.TrimSpace(line), "##") {
				inPlanSection = false
			}
			
			if !inPlanSection {
				newLines = append(newLines, line)
			}
		}
		
		if !planAdded {
			newLines = append(newLines, "", "## Plan", "", plan)
		}
		
		return strings.Join(newLines, "\n")
	} else {
		// Add plan section at the end
		return fmt.Sprintf("%s\n\n## Plan\n\n%s", existingBody, plan)
	}
}

func finish(params FinishParams) (*FinishResult, error) {
	// Determine issue number if not provided
	issueNumber := params.IssueNumber
	if issueNumber == 0 {
		branchName, err := shared.GetCurrentBranch(params.WorkingDir)
		if err != nil {
			return &FinishResult{
				Success: false,
				Message: fmt.Sprintf("Failed to get current branch: %v", err),
			}, nil
		}
		
		issueNumber, err = shared.ExtractIssueNumberFromBranch(branchName)
		if err != nil {
			return &FinishResult{
				Success: false,
				Message: fmt.Sprintf("Failed to extract issue number from branch '%s': %v", branchName, err),
			}, nil
		}
	}
	
	// Initialize GitHub service
	githubService, err := shared.NewGitHubService(params.WorkingDir)
	if err != nil {
		return &FinishResult{
			Success: false,
			Message: fmt.Sprintf("Failed to initialize GitHub service: %v", err),
		}, nil
	}
	
	// Get issue details for auto-generation
	issue, err := githubService.GetIssue(issueNumber)
	if err != nil {
		return &FinishResult{
			Success: false,
			Message: fmt.Sprintf("Failed to get issue #%d: %v", issueNumber, err),
		}, nil
	}
	
	// Auto-generate commit message if not provided
	commitMessage := params.CommitMessage
	if commitMessage == "" {
		commitMessage = fmt.Sprintf("feat: %s\n\nCloses #%d", issue.Title, issueNumber)
	}
	
	// Auto-generate PR title if not provided
	prTitle := params.PrTitle
	if prTitle == "" {
		prTitle = fmt.Sprintf("feat: %s", issue.Title)
	}
	
	// Auto-generate PR body if not provided
	prBody := params.PrBody
	if prBody == "" {
		prBody = fmt.Sprintf("## Summary\n\n%s\n\n## Changes\n\n- Implemented solution for issue #%d\n\nCloses #%d", issue.Title, issueNumber, issueNumber)
	}
	
	// Stage all changes
	if err := shared.GitAdd(params.WorkingDir); err != nil {
		return &FinishResult{
			Success: false,
			Message: fmt.Sprintf("Failed to stage changes: %v", err),
		}, nil
	}
	
	// Commit changes
	if err := shared.GitCommit(params.WorkingDir, commitMessage); err != nil {
		return &FinishResult{
			Success: false,
			Message: fmt.Sprintf("Failed to commit changes: %v", err),
		}, nil
	}
	
	// Push changes
	if err := shared.GitPush(params.WorkingDir); err != nil {
		return &FinishResult{
			Success: false,
			Message: fmt.Sprintf("Failed to push changes: %v", err),
		}, nil
	}
	
	// Create pull request
	prURL, err := githubService.CreatePullRequest(prTitle, prBody)
	if err != nil {
		return &FinishResult{
			Success: false,
			Message: fmt.Sprintf("Failed to create pull request: %v", err),
		}, nil
	}
	
	return &FinishResult{
		Success:     true,
		Message:     fmt.Sprintf("Successfully staged, committed, pushed changes and created pull request for issue #%d", issueNumber),
		PullRequest: prURL,
	}, nil
}

func sendResponse(id interface{}, result interface{}) {
	response := MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	
	jsonData, _ := json.Marshal(response)
	fmt.Println(string(jsonData))
}

func sendError(id interface{}, code int, message string) {
	response := MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &MCPError{
			Code:    code,
			Message: message,
		},
	}
	
	jsonData, _ := json.Marshal(response)
	fmt.Println(string(jsonData))
}