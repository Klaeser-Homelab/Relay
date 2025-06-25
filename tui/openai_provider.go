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
	openaiAPIURL    = "https://api.openai.com/v1/chat/completions"
	defaultGPTModel = "gpt-4"
)

// OpenAIProvider implements LLMProvider for OpenAI API access
type OpenAIProvider struct {
	config     LLMProviderConfig
	httpClient *http.Client
	logger     *log.Logger
	sessions   map[string]*OpenAISession
	sessionMu  sync.RWMutex
}

// OpenAISession represents a conversation session
type OpenAISession struct {
	ID       string
	Messages []OpenAIMessage
	mu       sync.Mutex
}

// OpenAIMessage represents a message in the conversation
type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAIRequest represents the request structure for OpenAI API
type OpenAIRequest struct {
	Model       string          `json:"model"`
	Messages    []OpenAIMessage `json:"messages"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
	Temperature float64         `json:"temperature,omitempty"`
}

// OpenAIResponse represents the response structure from OpenAI API
type OpenAIResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// NewOpenAIProvider creates a new OpenAI API provider
func NewOpenAIProvider(config LLMProviderConfig) (*OpenAIProvider, error) {
	logger := log.New(os.Stdout, "[OpenAIProvider] ", log.LstdFlags)

	// Use API key from config or environment
	apiKey := config.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("OpenAI API key not provided in config or OPENAI_API_KEY environment variable")
		}
	}

	// Set default model if not specified
	if config.Model == "" {
		config.Model = defaultGPTModel
	}

	// Set default max tokens if not specified
	if config.MaxTokens == 0 {
		config.MaxTokens = 4096
	}

	config.APIKey = apiKey

	httpClient := &http.Client{
		Timeout: 60 * time.Second,
	}

	logger.Printf("OpenAI API provider initialized (model: %s, maxTokens: %d)", config.Model, config.MaxTokens)

	return &OpenAIProvider{
		config:     config,
		httpClient: httpClient,
		logger:     logger,
		sessions:   make(map[string]*OpenAISession),
	}, nil
}

// SendMessage sends a message to OpenAI API
func (p *OpenAIProvider) SendMessage(ctx context.Context, message string) (string, error) {
	messages := []OpenAIMessage{
		{Role: "user", Content: message},
	}

	return p.sendRequest(ctx, messages)
}

// SendMessageWithSession sends a message with session continuity
func (p *OpenAIProvider) SendMessageWithSession(ctx context.Context, message string, sessionID string) (string, error) {
	p.sessionMu.Lock()
	session, exists := p.sessions[sessionID]
	if !exists {
		session = &OpenAISession{
			ID:       sessionID,
			Messages: []OpenAIMessage{},
		}
		p.sessions[sessionID] = session
	}
	p.sessionMu.Unlock()

	session.mu.Lock()
	defer session.mu.Unlock()

	// Add user message to session
	session.Messages = append(session.Messages, OpenAIMessage{
		Role:    "user",
		Content: message,
	})

	// Send request with full conversation history
	response, err := p.sendRequest(ctx, session.Messages)
	if err != nil {
		return "", err
	}

	// Add assistant response to session
	session.Messages = append(session.Messages, OpenAIMessage{
		Role:    "assistant",
		Content: response,
	})

	return response, nil
}

// sendRequest sends a request to OpenAI API
func (p *OpenAIProvider) sendRequest(ctx context.Context, messages []OpenAIMessage) (string, error) {
	request := OpenAIRequest{
		Model:       p.config.Model,
		Messages:    messages,
		MaxTokens:   p.config.MaxTokens,
		Temperature: 0.7,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	p.logger.Printf("Sending request to OpenAI API (model: %s, messages: %d)", p.config.Model, len(messages))

	req, err := http.NewRequestWithContext(ctx, "POST", openaiAPIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.config.APIKey)

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
		return "", fmt.Errorf("OpenAI API error (status %d): %s", resp.StatusCode, string(body))
	}

	var openaiResp OpenAIResponse
	if err := json.Unmarshal(body, &openaiResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(openaiResp.Choices) == 0 {
		return "", fmt.Errorf("empty response choices from OpenAI API")
	}

	response := openaiResp.Choices[0].Message.Content
	p.logger.Printf("Received response from OpenAI API (%d chars, %d total tokens)",
		len(response), openaiResp.Usage.TotalTokens)

	return response, nil
}

// GetProviderName returns the provider name
func (p *OpenAIProvider) GetProviderName() string {
	return "openai"
}

// Close closes the provider and cleans up resources
func (p *OpenAIProvider) Close() error {
	p.sessionMu.Lock()
	defer p.sessionMu.Unlock()

	// Clear all sessions
	p.sessions = make(map[string]*OpenAISession)

	p.logger.Println("OpenAI API provider closed")
	return nil
}
