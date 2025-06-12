package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

const (
	claudeAPIURL = "https://api.anthropic.com/v1/messages"
	defaultModel = "claude-3-5-sonnet-20241022"
)

// ClaudeProvider implements LLMProvider for direct Claude API access
type ClaudeProvider struct {
	config     LLMProviderConfig
	httpClient *http.Client
	logger     *log.Logger
	workingDir string
	sessions   map[string]*ClaudeSession
	sessionMu  sync.RWMutex
}

// ClaudeSession represents a conversation session
type ClaudeSession struct {
	ID       string
	Messages []ClaudeMessage
	mu       sync.Mutex
}

// ClaudeMessage represents a message in the conversation
type ClaudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ClaudeRequest represents the request structure for Claude API
type ClaudeRequest struct {
	Model     string          `json:"model"`
	MaxTokens int             `json:"max_tokens"`
	Messages  []ClaudeMessage `json:"messages"`
}

// ClaudeAPIResponse represents the response structure from Claude API
type ClaudeAPIResponse struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Role    string `json:"role"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Model        string `json:"model"`
	StopReason   string `json:"stop_reason"`
	StopSequence string `json:"stop_sequence"`
	Usage        struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// NewClaudeProvider creates a new Claude API provider
func NewClaudeProvider(config LLMProviderConfig, workingDir string) (*ClaudeProvider, error) {
	logger := log.New(os.Stdout, "[ClaudeProvider] ", log.LstdFlags)

	// Use API key from config or environment
	apiKey := config.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("Claude API key not provided in config or ANTHROPIC_API_KEY environment variable")
		}
	}

	// Set default model if not specified
	if config.Model == "" {
		config.Model = defaultModel
	}

	// Set default max tokens if not specified
	if config.MaxTokens == 0 {
		config.MaxTokens = 4096
	}

	config.APIKey = apiKey

	httpClient := &http.Client{
		Timeout: 60 * time.Second,
	}

	logger.Printf("Claude API provider initialized (model: %s, maxTokens: %d)", config.Model, config.MaxTokens)

	return &ClaudeProvider{
		config:     config,
		httpClient: httpClient,
		logger:     logger,
		workingDir: workingDir,
		sessions:   make(map[string]*ClaudeSession),
	}, nil
}

// SendMessage sends a message to Claude API
func (p *ClaudeProvider) SendMessage(ctx context.Context, message string) (string, error) {
	messages := []ClaudeMessage{
		{Role: "user", Content: message},
	}

	return p.sendRequest(ctx, messages)
}

// SendMessageWithSession sends a message with session continuity
func (p *ClaudeProvider) SendMessageWithSession(ctx context.Context, message string, sessionID string) (string, error) {
	p.sessionMu.Lock()
	session, exists := p.sessions[sessionID]
	if !exists {
		session = &ClaudeSession{
			ID:       sessionID,
			Messages: []ClaudeMessage{},
		}
		p.sessions[sessionID] = session
	}
	p.sessionMu.Unlock()

	session.mu.Lock()
	defer session.mu.Unlock()

	// Add user message to session
	session.Messages = append(session.Messages, ClaudeMessage{
		Role:    "user",
		Content: message,
	})

	// Send request with full conversation history
	response, err := p.sendRequest(ctx, session.Messages)
	if err != nil {
		return "", err
	}

	// Add assistant response to session
	session.Messages = append(session.Messages, ClaudeMessage{
		Role:    "assistant",
		Content: response,
	})

	return response, nil
}

// sendRequest sends a request to Claude API
func (p *ClaudeProvider) sendRequest(ctx context.Context, messages []ClaudeMessage) (string, error) {
	request := ClaudeRequest{
		Model:     p.config.Model,
		MaxTokens: p.config.MaxTokens,
		Messages:  messages,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	p.logger.Printf("Sending request to Claude API (model: %s, messages: %d)", p.config.Model, len(messages))

	req, err := http.NewRequestWithContext(ctx, "POST", claudeAPIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.config.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Claude API error (status %d): %s", resp.StatusCode, string(body))
	}

	var claudeResp ClaudeAPIResponse
	if err := json.Unmarshal(body, &claudeResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(claudeResp.Content) == 0 {
		return "", fmt.Errorf("empty response content from Claude API")
	}

	response := claudeResp.Content[0].Text
	p.logger.Printf("Received response from Claude API (%d chars, %d input tokens, %d output tokens)",
		len(response), claudeResp.Usage.InputTokens, claudeResp.Usage.OutputTokens)

	return response, nil
}

// GetProviderName returns the provider name
func (p *ClaudeProvider) GetProviderName() string {
	return "claude"
}

// Close closes the provider and cleans up resources
func (p *ClaudeProvider) Close() error {
	p.sessionMu.Lock()
	defer p.sessionMu.Unlock()

	// Clear all sessions
	p.sessions = make(map[string]*ClaudeSession)

	p.logger.Println("Claude API provider closed")
	return nil
}
