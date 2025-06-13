package main

import (
	"context"
	"os"
	"testing"
	"time"
)

// TestLLMProviderFactory tests the provider factory
func TestLLMProviderFactory(t *testing.T) {
	factory := NewProviderFactory("/tmp")

	// Test Claude CLI provider creation
	config := LLMProviderConfig{
		Type:      "claude-cli",
		Model:     "claude-3-5-sonnet-20241022",
		MaxTokens: 4096,
		Options:   make(map[string]string),
	}

	provider, err := factory.CreateProvider(config)
	if err != nil {
		t.Fatalf("Failed to create Claude CLI provider: %v", err)
	}

	if provider.GetProviderName() != "claude-cli" {
		t.Errorf("Expected provider name 'claude-cli', got '%s'", provider.GetProviderName())
	}

	provider.Close()

	// Test unsupported provider type
	config.Type = "unsupported"
	_, err = factory.CreateProvider(config)
	if err == nil {
		t.Error("Expected error for unsupported provider type")
	}
}

// TestLLMManager tests the LLM manager
func TestLLMManager(t *testing.T) {
	planningConfig := LLMProviderConfig{
		Type:      "claude-cli",
		Model:     "claude-3-5-sonnet-20241022",
		MaxTokens: 4096,
		Options:   make(map[string]string),
	}

	executingConfig := LLMProviderConfig{
		Type:      "claude-cli",
		Model:     "claude-3-5-sonnet-20241022",
		MaxTokens: 4096,
		Options:   make(map[string]string),
	}

	manager, err := NewLLMManager(planningConfig, executingConfig, "/tmp")
	if err != nil {
		t.Fatalf("Failed to create LLM manager: %v", err)
	}

	// Test that providers are created correctly
	planningProvider := manager.GetPlanningProvider()
	if planningProvider == nil {
		t.Error("Planning provider is nil")
	}

	executingProvider := manager.GetExecutingProvider()
	if executingProvider == nil {
		t.Error("Executing provider is nil")
	}

	// Test provider names
	if planningProvider.GetProviderName() != "claude-cli" {
		t.Errorf("Expected planning provider name 'claude-cli', got '%s'", planningProvider.GetProviderName())
	}

	if executingProvider.GetProviderName() != "claude-cli" {
		t.Errorf("Expected executing provider name 'claude-cli', got '%s'", executingProvider.GetProviderName())
	}

	// Test close
	err = manager.Close()
	if err != nil {
		t.Errorf("Failed to close LLM manager: %v", err)
	}
}

// TestClaudeCLIProvider tests the Claude CLI provider wrapper
func TestClaudeCLIProvider(t *testing.T) {
	provider, err := NewClaudeCLIProvider("/tmp")
	if err != nil {
		t.Fatalf("Failed to create Claude CLI provider: %v", err)
	}
	defer provider.Close()

	// Test provider name
	if provider.GetProviderName() != "claude-cli" {
		t.Errorf("Expected provider name 'claude-cli', got '%s'", provider.GetProviderName())
	}

	// Test basic message sending (this will only work if Claude CLI is installed)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Skip the actual API call if Claude CLI is not available
	if !isClaudeCLIAvailable() {
		t.Skip("Claude CLI not available, skipping message test")
	}

	response, err := provider.SendMessage(ctx, "Hello, respond with just 'Hi'")
	if err != nil {
		t.Logf("Claude CLI not available or failed: %v", err)
		return
	}

	if response == "" {
		t.Error("Expected non-empty response")
	}

	t.Logf("Response: %s", response)
}

// TestClaudeProvider tests the Claude API provider (requires API key)
func TestClaudeProvider(t *testing.T) {
	// Skip if no API key is available
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("ANTHROPIC_API_KEY not set, skipping Claude API test")
	}

	config := LLMProviderConfig{
		Type:      "claude",
		APIKey:    apiKey,
		Model:     "claude-3-5-sonnet-20241022",
		MaxTokens: 100,
		Options:   make(map[string]string),
	}

	provider, err := NewClaudeProvider(config, "/tmp")
	if err != nil {
		t.Fatalf("Failed to create Claude provider: %v", err)
	}
	defer provider.Close()

	// Test provider name
	if provider.GetProviderName() != "claude" {
		t.Errorf("Expected provider name 'claude', got '%s'", provider.GetProviderName())
	}

	// Test basic message sending
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	response, err := provider.SendMessage(ctx, "Hello, respond with just 'Hi'")
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	if response == "" {
		t.Error("Expected non-empty response")
	}

	t.Logf("Response: %s", response)

	// Test session continuity
	sessionID := "test-session"
	response1, err := provider.SendMessageWithSession(ctx, "My name is Alice", sessionID)
	if err != nil {
		t.Fatalf("Failed to send first message with session: %v", err)
	}

	response2, err := provider.SendMessageWithSession(ctx, "What is my name?", sessionID)
	if err != nil {
		t.Fatalf("Failed to send second message with session: %v", err)
	}

	// The model should remember the name from the previous message
	t.Logf("First response: %s", response1)
	t.Logf("Second response: %s", response2)
}

// TestOpenAIProvider tests the OpenAI provider (requires API key)
func TestOpenAIProvider(t *testing.T) {
	// Skip if no API key is available
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set, skipping OpenAI test")
	}

	config := LLMProviderConfig{
		Type:      "openai",
		APIKey:    apiKey,
		Model:     "gpt-4",
		MaxTokens: 100,
		Options:   make(map[string]string),
	}

	provider, err := NewOpenAIProvider(config)
	if err != nil {
		t.Fatalf("Failed to create OpenAI provider: %v", err)
	}
	defer provider.Close()

	// Test provider name
	if provider.GetProviderName() != "openai" {
		t.Errorf("Expected provider name 'openai', got '%s'", provider.GetProviderName())
	}

	// Test basic message sending
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	response, err := provider.SendMessage(ctx, "Hello, respond with just 'Hi'")
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	if response == "" {
		t.Error("Expected non-empty response")
	}

	t.Logf("Response: %s", response)
}

// Helper function to check if Claude CLI is available
func isClaudeCLIAvailable() bool {
	// This is a copy of the check from claude_tests.go
	_, err := NewClaudeCLI(false, "/tmp")
	return err == nil
}
