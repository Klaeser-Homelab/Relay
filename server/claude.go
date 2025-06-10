package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

type ClaudeCLI struct {
	logger         *log.Logger
	useSession     bool
	sessionStarted bool
	workingDir     string
}

type ClaudeResponse struct {
	Content string `json:"content"`
	Role    string `json:"role"`
}

func NewClaudeCLI(useSession bool, workingDir string) (*ClaudeCLI, error) {
	logger := log.New(os.Stdout, "[ClaudeCLI] ", log.LstdFlags)

	logger.Printf("Claude CLI client initialized (session: %v, workingDir: %s)", useSession, workingDir)

	return &ClaudeCLI{
		logger:         logger,
		useSession:     useSession,
		sessionStarted: false,
		workingDir:     workingDir,
	}, nil
}

func (c *ClaudeCLI) SendCommand(command string) (string, error) {
	c.logger.Printf("Sending command: %s", command)

	var cmd *exec.Cmd

	if c.useSession && c.sessionStarted {
		cmd = exec.Command("claude", "--print", "--output-format", "json", "--continue", command)
	} else {
		cmd = exec.Command("claude", "--print", "--output-format", "json", command)
		if c.useSession {
			c.sessionStarted = true
		}
	}

	// Set working directory if specified
	if c.workingDir != "" {
		cmd.Dir = c.workingDir
	}

	output, err := cmd.Output()
	if err != nil {
		// If the command failed, try to get stderr for better error messages
		if exitError, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("claude command failed: %s (stderr: %s)", err.Error(), string(exitError.Stderr))
		}
		return "", fmt.Errorf("failed to execute claude command: %w", err)
	}

	responseText := strings.TrimSpace(string(output))
	c.logger.Printf("Received raw response (%d chars): %s", len(responseText), responseText)

	return c.parseResponse(responseText)
}

func (c *ClaudeCLI) parseResponse(responseText string) (string, error) {
	if responseText == "" {
		return "", fmt.Errorf("empty response from claude")
	}

	// Try to parse as JSON first
	var claudeResp ClaudeResponse
	err := json.Unmarshal([]byte(responseText), &claudeResp)
	if err == nil {
		c.logger.Printf("Parsed JSON response: %s", claudeResp.Content)
		return claudeResp.Content, nil
	}

	// If JSON parsing fails, return raw text
	c.logger.Printf("Using raw text response (JSON parse failed: %v)", err)
	return responseText, nil
}

func (c *ClaudeCLI) SendCommandInProject(command string, projectPath string) (string, error) {
	// Temporarily change working directory for this command
	originalDir := c.workingDir
	c.workingDir = projectPath

	result, err := c.SendCommand(command)

	// Restore original working directory
	c.workingDir = originalDir

	return result, err
}

func (c *ClaudeCLI) Close() error {
	c.logger.Println("Claude CLI client closed")
	return nil
}
