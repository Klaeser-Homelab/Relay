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
	logger          *log.Logger
	useSession      bool
	sessionStarted  bool
}

type ClaudeResponse struct {
	Content string `json:"content"`
	Role    string `json:"role"`
}

func NewClaudeCLI(useSession bool) (*ClaudeCLI, error) {
	logger := log.New(os.Stdout, "[ClaudeCLI] ", log.LstdFlags)
	
	logger.Println("Claude CLI client initialized")
	
	return &ClaudeCLI{
		logger:          logger,
		useSession:      useSession,
		sessionStarted:  false,
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
	
	output, err := cmd.Output()
	if err != nil {
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

func (c *ClaudeCLI) Close() error {
	c.logger.Println("Claude CLI client closed")
	return nil
}

func testClaudeIntegration() string {
	claude, err := NewClaudeCLI(true) // Enable session continuity
	if err != nil {
		log.Printf("Failed to create Claude CLI: %v", err)
		return "failed"
	}
	defer claude.Close()
	
	response, err := claude.SendCommand("hello world")
	if err != nil {
		log.Printf("Failed to send command: %v", err)
		return "failed"
	}
	
	fmt.Printf("Claude Response:\n%s\n", response)
	
	return "done"
}

func main() {
	fmt.Println("Starting Relay Server...")
	
	result := testClaudeIntegration()
	fmt.Printf("Claude integration test result: %s\n", result)
}