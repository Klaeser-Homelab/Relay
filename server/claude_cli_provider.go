package main

import (
	"context"
)

// ClaudeCLIProvider wraps the existing ClaudeCLI to implement LLMProvider interface
type ClaudeCLIProvider struct {
	cli *ClaudeCLI
}

// NewClaudeCLIProvider creates a new Claude CLI provider
func NewClaudeCLIProvider(workingDir string) (*ClaudeCLIProvider, error) {
	cli, err := NewClaudeCLI(true, workingDir)
	if err != nil {
		return nil, err
	}
	
	return &ClaudeCLIProvider{
		cli: cli,
	}, nil
}

// SendMessage sends a message to Claude via CLI
func (p *ClaudeCLIProvider) SendMessage(ctx context.Context, message string) (string, error) {
	return p.cli.SendCommand(message)
}

// SendMessageWithSession sends a message with session continuity via CLI
func (p *ClaudeCLIProvider) SendMessageWithSession(ctx context.Context, message string, sessionID string) (string, error) {
	// CLI implementation already handles sessions internally
	return p.cli.SendCommand(message)
}

// GetProviderName returns the provider name
func (p *ClaudeCLIProvider) GetProviderName() string {
	return "claude-cli"
}

// Close closes the CLI provider
func (p *ClaudeCLIProvider) Close() error {
	return p.cli.Close()
}

// SendCommandInProject sends a command in a specific project context
func (p *ClaudeCLIProvider) SendCommandInProject(command string, projectPath string) (string, error) {
	return p.cli.SendCommandInProject(command, projectPath)
}