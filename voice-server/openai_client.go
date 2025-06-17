package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// OpenAI Realtime API client
type OpenAIRealtimeClient struct {
	apiKey      string
	baseURL     string
	conn        *websocket.Conn
	context     context.Context
	cancel      context.CancelFunc
	tools       []OpenAITool
	mu          sync.RWMutex
	closed      bool
	isConnected bool
	
	// Response channel
	responses chan *OpenAIResponse
}

type OpenAITool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

type OpenAIResponse struct {
	Type         string        `json:"type"`
	AudioData    []byte        `json:"audio_data,omitempty"`
	Text         string        `json:"text,omitempty"`
	FunctionCall *FunctionCall `json:"function_call,omitempty"`
	Error        string        `json:"error,omitempty"`
}

type FunctionCall struct {
	CallID    string                 `json:"call_id"`
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// OpenAI Realtime API message types
type RealtimeMessage struct {
	Type   string      `json:"type"`
	EventID string     `json:"event_id,omitempty"`
	Data   interface{} `json:",inline"`
}

type SessionUpdate struct {
	Type    string `json:"type"`
	Session struct {
		Model         string        `json:"model"`
		Instructions  string        `json:"instructions"`
		Voice         string        `json:"voice"`
		InputFormat   string        `json:"input_audio_format"`
		OutputFormat  string        `json:"output_audio_format"`
		Tools         []OpenAITool  `json:"tools"`
	} `json:"session"`
}

type AudioAppend struct {
	Type  string `json:"type"`
	Audio string `json:"audio"` // base64 encoded
}

type ResponseCreate struct {
	Type     string `json:"type"`
	Response struct {
		Modalities []string `json:"modalities"`
	} `json:"response"`
}

func NewOpenAIRealtimeClient(apiKey string) (*OpenAIRealtimeClient, error) {
	ctx, cancel := context.WithCancel(context.Background())
	
	client := &OpenAIRealtimeClient{
		apiKey:    apiKey,
		baseURL:   "wss://api.openai.com/v1/realtime",
		context:   ctx,
		cancel:    cancel,
		responses: make(chan *OpenAIResponse, 100),
	}

	return client, nil
}

func (c *OpenAIRealtimeClient) StartSession() error {
	c.mu.Lock()
	
	// Check if already connected
	if c.isConnected && c.conn != nil && !c.closed {
		c.mu.Unlock()
		log.Printf("OpenAI session already connected")
		return nil
	}
	
	// Reset connection state
	c.closed = false
	c.isConnected = false
	c.mu.Unlock()

	// Build WebSocket URL with auth
	u, err := url.Parse(c.baseURL)
	if err != nil {
		return fmt.Errorf("invalid base URL: %w", err)
	}

	// Add model parameter
	q := u.Query()
	q.Set("model", "gpt-4o-realtime-preview-2024-10-01")
	u.RawQuery = q.Encode()

	// Create WebSocket connection with auth headers
	header := http.Header{}
	header.Set("Authorization", "Bearer "+c.apiKey)
	header.Set("OpenAI-Beta", "realtime=v1")

	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	log.Printf("Connecting to OpenAI Realtime API...")
	conn, _, err := dialer.Dial(u.String(), header)
	if err != nil {
		return fmt.Errorf("failed to connect to OpenAI Realtime API: %w", err)
	}

	c.mu.Lock()
	c.conn = conn
	c.isConnected = true
	c.mu.Unlock()

	log.Printf("WebSocket connection established")

	// Start message handler
	go c.handleMessages()

	// Initialize session with basic configuration
	err = c.initializeSession()
	if err != nil {
		c.Close()
		return fmt.Errorf("failed to initialize session: %w", err)
	}

	return nil
}

func (c *OpenAIRealtimeClient) initializeSession() error {
	// Start with minimal session configuration - no tools initially
	sessionUpdate := SessionUpdate{
		Type: "session.update",
	}
	
	sessionUpdate.Session.Model = "gpt-4o-realtime-preview-2024-10-01"
	sessionUpdate.Session.Instructions = `You are a voice assistant for the Relay development tool. 
	Help users manage their development projects through voice commands.
	Convert natural language requests into appropriate function calls.
	Always confirm what actions you're taking and provide helpful feedback.`
	sessionUpdate.Session.Voice = "alloy"
	sessionUpdate.Session.InputFormat = "pcm16"
	sessionUpdate.Session.OutputFormat = "pcm16"
	// Don't set tools initially - will be configured later
	sessionUpdate.Session.Tools = []OpenAITool{}

	return c.sendMessage(sessionUpdate)
}

func (c *OpenAIRealtimeClient) ConfigureSessionTools() error {
	c.mu.RLock()
	conn := c.conn
	closed := c.closed
	connected := c.isConnected
	tools := c.tools
	c.mu.RUnlock()

	if conn == nil || closed || !connected {
		return fmt.Errorf("connection not established")
	}

	// Send updated session configuration with tools
	sessionUpdate := SessionUpdate{
		Type: "session.update",
	}
	sessionUpdate.Session.Tools = tools

	return c.sendMessage(sessionUpdate)
}

func (c *OpenAIRealtimeClient) ConfigureTools(tools []OpenAITool) error {
	c.mu.Lock()
	c.tools = tools
	c.mu.Unlock()

	// If session is already started, update it
	if c.conn != nil {
		return c.initializeSession()
	}

	return nil
}

func (c *OpenAIRealtimeClient) SendAudio(audioData []byte) error {
	c.mu.RLock()
	conn := c.conn
	closed := c.closed
	connected := c.isConnected
	c.mu.RUnlock()

	if conn == nil || closed || !connected {
		return fmt.Errorf("connection not established")
	}

	// Convert audio to base64
	audioBase64 := base64.StdEncoding.EncodeToString(audioData)

	audioAppend := AudioAppend{
		Type:  "input_audio_buffer.append",
		Audio: audioBase64,
	}

	err := c.sendMessage(audioAppend)
	if err != nil {
		return err
	}

	// Don't commit and request response for every chunk - let them accumulate
	// Only commit when we have a substantial amount or on user action
	return nil
}

func (c *OpenAIRealtimeClient) CommitAudio() error {
	c.mu.RLock()
	conn := c.conn
	closed := c.closed
	connected := c.isConnected
	c.mu.RUnlock()

	if conn == nil || closed || !connected {
		return fmt.Errorf("connection not established")
	}

	// Commit the audio buffer and request response
	commitMsg := map[string]interface{}{
		"type": "input_audio_buffer.commit",
	}

	err := c.sendMessage(commitMsg)
	if err != nil {
		return err
	}

	// Request response generation
	responseCreate := ResponseCreate{
		Type: "response.create",
	}
	responseCreate.Response.Modalities = []string{"text", "audio"}

	return c.sendMessage(responseCreate)
}

func (c *OpenAIRealtimeClient) SendFunctionResult(callID string, result interface{}, errorMsg string) error {
	response := map[string]interface{}{
		"type": "conversation.item.create",
		"item": map[string]interface{}{
			"type":    "function_call_output",
			"call_id": callID,
		},
	}

	if errorMsg != "" {
		response["item"].(map[string]interface{})["output"] = fmt.Sprintf("Error: %s", errorMsg)
	} else {
		output, _ := json.Marshal(result)
		response["item"].(map[string]interface{})["output"] = string(output)
	}

	err := c.sendMessage(response)
	if err != nil {
		return err
	}

	// Request response after function result
	responseCreate := ResponseCreate{
		Type: "response.create",
	}
	responseCreate.Response.Modalities = []string{"text", "audio"}

	return c.sendMessage(responseCreate)
}

func (c *OpenAIRealtimeClient) UpdateContext(instructions string) error {
	sessionUpdate := SessionUpdate{
		Type: "session.update",
	}
	sessionUpdate.Session.Instructions = instructions

	return c.sendMessage(sessionUpdate)
}

func (c *OpenAIRealtimeClient) ReceiveResponse() (*OpenAIResponse, error) {
	select {
	case response := <-c.responses:
		return response, nil
	case <-c.context.Done():
		return nil, io.EOF
	}
}

func (c *OpenAIRealtimeClient) handleMessages() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("OpenAI message handler recovered from panic: %v\n", r)
		}
		c.Close()
	}()

	for {
		select {
		case <-c.context.Done():
			return
		default:
			c.mu.RLock()
			conn := c.conn
			closed := c.closed
			c.mu.RUnlock()

			if conn == nil || closed {
				return
			}

			var msg map[string]interface{}
			err := conn.ReadJSON(&msg)
			if err != nil {
				// Mark as disconnected
				c.mu.Lock()
				c.isConnected = false
				c.mu.Unlock()
				
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					fmt.Printf("OpenAI WebSocket error: %v\n", err)
				} else {
					fmt.Printf("OpenAI WebSocket closed: %v\n", err)
				}
				return
			}

			c.processMessage(msg)
		}
	}
}

func (c *OpenAIRealtimeClient) processMessage(msg map[string]interface{}) {
	msgType, ok := msg["type"].(string)
	if !ok {
		return
	}

	switch msgType {
	case "response.audio.delta":
		c.handleAudioDelta(msg)
	case "response.text.delta":
		c.handleTextDelta(msg)
	case "response.function_call_arguments.delta":
		c.handleFunctionCallDelta(msg)
	case "response.function_call_arguments.done":
		c.handleFunctionCallDone(msg)
	case "conversation.item.input_audio_transcription.completed":
		c.handleTranscriptionCompleted(msg)
	case "response.audio_transcript.delta":
		c.handleAudioTranscriptDelta(msg)
	case "response.audio_transcript.done":
		c.handleAudioTranscriptDone(msg)
	case "response.output_item.added":
		c.handleOutputItemAdded(msg)
	case "conversation.item.created":
		c.handleConversationItemCreated(msg)
	case "response.content_part.added":
		c.handleContentPartAdded(msg)
	case "response.audio.done":
		c.handleAudioDone(msg)
	case "response.content_part.done":
		c.handleContentPartDone(msg)
	case "response.output_item.done":
		c.handleOutputItemDone(msg)
	case "rate_limits.updated":
		c.handleRateLimitsUpdated(msg)
	case "error":
		c.handleError(msg)
	case "session.created":
		fmt.Println("OpenAI session created successfully")
	case "session.updated":
		fmt.Println("OpenAI session updated")
	case "response.created":
		fmt.Println("OpenAI response created")
	case "response.done":
		fmt.Println("OpenAI response completed")
	default:
		// Log unknown message types for debugging (less verbosely)
		fmt.Printf("Unhandled OpenAI message: %s\n", msgType)
	}
}

func (c *OpenAIRealtimeClient) handleAudioDelta(msg map[string]interface{}) {
	if delta, ok := msg["delta"].(string); ok {
		// Decode base64 audio
		audioData, err := base64.StdEncoding.DecodeString(delta)
		if err != nil {
			fmt.Printf("Failed to decode audio delta: %v\n", err)
			return
		}

		response := &OpenAIResponse{
			Type:      "audio",
			AudioData: audioData,
		}

		select {
		case c.responses <- response:
		default:
			// Channel full, skip this audio chunk
		case <-c.context.Done():
			return
		}
	}
}

func (c *OpenAIRealtimeClient) handleTextDelta(msg map[string]interface{}) {
	if delta, ok := msg["delta"].(string); ok {
		response := &OpenAIResponse{
			Type: "text",
			Text: delta,
		}

		select {
		case c.responses <- response:
		default:
			// Channel full, skip
		}
	}
}

func (c *OpenAIRealtimeClient) handleFunctionCallDelta(msg map[string]interface{}) {
	// For function calls, we'll wait for the "done" event
	// This just accumulates the arguments
}

func (c *OpenAIRealtimeClient) handleFunctionCallDone(msg map[string]interface{}) {
	callID, _ := msg["call_id"].(string)
	name, _ := msg["name"].(string)
	argumentsStr, _ := msg["arguments"].(string)

	var arguments map[string]interface{}
	if argumentsStr != "" {
		json.Unmarshal([]byte(argumentsStr), &arguments)
	}

	functionCall := &FunctionCall{
		CallID:    callID,
		Name:      name,
		Arguments: arguments,
	}

	response := &OpenAIResponse{
		Type:         "function_call",
		FunctionCall: functionCall,
	}

	select {
	case c.responses <- response:
	default:
		// Channel full
	}
}

func (c *OpenAIRealtimeClient) handleTranscriptionCompleted(msg map[string]interface{}) {
	if transcript, ok := msg["transcript"].(string); ok {
		response := &OpenAIResponse{
			Type: "transcription",
			Text: transcript,
		}

		select {
		case c.responses <- response:
		default:
			// Channel full
		}
	}
}

func (c *OpenAIRealtimeClient) handleError(msg map[string]interface{}) {
	errorMsg := "Unknown error"
	if err, ok := msg["error"].(map[string]interface{}); ok {
		if message, ok := err["message"].(string); ok {
			errorMsg = message
		}
	}

	response := &OpenAIResponse{
		Type:  "error",
		Error: errorMsg,
	}

	select {
	case c.responses <- response:
	default:
		// Channel full
	}
}

func (c *OpenAIRealtimeClient) sendMessage(message interface{}) error {
	c.mu.RLock()
	conn := c.conn
	closed := c.closed
	connected := c.isConnected
	c.mu.RUnlock()

	if conn == nil || closed || !connected {
		return fmt.Errorf("connection not established")
	}

	return conn.WriteJSON(message)
}

func (c *OpenAIRealtimeClient) handleAudioTranscriptDelta(msg map[string]interface{}) {
	if delta, ok := msg["delta"].(string); ok {
		fmt.Printf("Audio transcript: %s\n", delta)
		
		response := &OpenAIResponse{
			Type: "transcription",
			Text: delta,
		}

		select {
		case c.responses <- response:
		default:
			// Channel full
		}
	}
}

func (c *OpenAIRealtimeClient) handleAudioTranscriptDone(msg map[string]interface{}) {
	if transcript, ok := msg["transcript"].(string); ok {
		fmt.Printf("Audio transcript complete: %s\n", transcript)
		
		response := &OpenAIResponse{
			Type: "transcription",
			Text: transcript,
		}

		select {
		case c.responses <- response:
		default:
			// Channel full
		}
	}
}

func (c *OpenAIRealtimeClient) handleOutputItemAdded(msg map[string]interface{}) {
	// Handle output item added - usually indicates response structure
	fmt.Println("OpenAI output item added")
}

func (c *OpenAIRealtimeClient) handleConversationItemCreated(msg map[string]interface{}) {
	// Handle conversation item created
	fmt.Println("OpenAI conversation item created")
}

func (c *OpenAIRealtimeClient) handleContentPartAdded(msg map[string]interface{}) {
	// Handle content part added - part of response structure
	fmt.Println("OpenAI content part added")
}

func (c *OpenAIRealtimeClient) handleAudioDone(msg map[string]interface{}) {
	// Handle audio response completion
	fmt.Println("OpenAI audio response complete")
}

func (c *OpenAIRealtimeClient) handleContentPartDone(msg map[string]interface{}) {
	// Handle content part completion
	fmt.Println("OpenAI content part complete")
}

func (c *OpenAIRealtimeClient) handleOutputItemDone(msg map[string]interface{}) {
	// Handle output item completion
	fmt.Println("OpenAI output item complete")
}

func (c *OpenAIRealtimeClient) handleRateLimitsUpdated(msg map[string]interface{}) {
	// Handle rate limit updates - can be useful for monitoring
	if limits, ok := msg["rate_limits"].([]interface{}); ok {
		fmt.Printf("OpenAI rate limits updated: %d limits\n", len(limits))
	}
}

func (c *OpenAIRealtimeClient) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if c.closed {
		return // Already closed
	}
	
	c.closed = true
	c.isConnected = false
	c.cancel()
	
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}

	// Close channel safely
	select {
	case <-c.responses:
		// Channel already closed or empty
	default:
		close(c.responses)
	}
}
