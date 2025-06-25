package main

import (
	"context"
	"fmt"
)

// LLMProvider defines the interface for all LLM providers
type LLMProvider interface {
	// SendMessage sends a message to the LLM and returns the response
	SendMessage(ctx context.Context, message string) (string, error)

	// SendMessageWithSession sends a message with session continuity
	SendMessageWithSession(ctx context.Context, message string, sessionID string) (string, error)

	// GetProviderName returns the name of the provider
	GetProviderName() string

	// Close cleans up any resources used by the provider
	Close() error
}

// LLMProviderConfig holds configuration for LLM providers
type LLMProviderConfig struct {
	Type      string            `json:"type"`       // "claude", "openai", "local", etc.
	APIKey    string            `json:"api_key"`    // API key for the provider
	BaseURL   string            `json:"base_url"`   // Custom base URL (optional)
	Model     string            `json:"model"`      // Model name (e.g., "claude-3-5-sonnet-20241022")
	MaxTokens int               `json:"max_tokens"` // Maximum tokens in response
	Options   map[string]string `json:"options"`    // Provider-specific options
}

// ProviderFactory creates LLM providers based on configuration
type ProviderFactory struct {
	workingDir string
}

// NewProviderFactory creates a new provider factory
func NewProviderFactory(workingDir string) *ProviderFactory {
	return &ProviderFactory{
		workingDir: workingDir,
	}
}

// CreateProvider creates an LLM provider based on the configuration
func (f *ProviderFactory) CreateProvider(config LLMProviderConfig) (LLMProvider, error) {
	switch config.Type {
	case "claude":
		return NewClaudeProvider(config, f.workingDir)
	case "claude-cli":
		// Backwards compatibility with existing CLI implementation
		return NewClaudeCLIProvider(f.workingDir)
	case "openai":
		return NewOpenAIProvider(config)
	default:
		return nil, fmt.Errorf("unsupported LLM provider type: %s", config.Type)
	}
}

// LLMManager manages multiple LLM providers for different use cases
type LLMManager struct {
	planningProvider  LLMProvider
	executingProvider LLMProvider
	factory           *ProviderFactory
}

// NewLLMManager creates a new LLM manager with the specified providers
func NewLLMManager(planningConfig, executingConfig LLMProviderConfig, workingDir string) (*LLMManager, error) {
	factory := NewProviderFactory(workingDir)

	planningProvider, err := factory.CreateProvider(planningConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create planning provider: %w", err)
	}

	executingProvider, err := factory.CreateProvider(executingConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create executing provider: %w", err)
	}

	return &LLMManager{
		planningProvider:  planningProvider,
		executingProvider: executingProvider,
		factory:           factory,
	}, nil
}

// GetPlanningProvider returns the provider for planning tasks
func (m *LLMManager) GetPlanningProvider() LLMProvider {
	return m.planningProvider
}

// GetExecutingProvider returns the provider for executing tasks
func (m *LLMManager) GetExecutingProvider() LLMProvider {
	return m.executingProvider
}

// Close closes all providers
func (m *LLMManager) Close() error {
	var errs []error

	if err := m.planningProvider.Close(); err != nil {
		errs = append(errs, fmt.Errorf("planning provider close error: %w", err))
	}

	if err := m.executingProvider.Close(); err != nil {
		errs = append(errs, fmt.Errorf("executing provider close error: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing providers: %v", errs)
	}

	return nil
}
