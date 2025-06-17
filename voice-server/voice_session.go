package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/gofiber/websocket/v2"
)

// VoiceSession represents a single voice session with a client
type VoiceSession struct {
	ID           string
	WebSocket    *websocket.Conn
	OpenAIConn   *OpenAIRealtimeClient
	RelayManager *RelayManager
	Context      context.Context
	Cancel       context.CancelFunc
	mu           sync.RWMutex
	
	// Session state
	CurrentProject string
	IsRecording    bool
	LastActivity   time.Time
}

// Message types for WebSocket communication
type WSMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data,omitempty"`
}

type AudioMessage struct {
	Type      string `json:"type"`
	AudioData []byte `json:"audio_data"`
}

type StatusMessage struct {
	Type    string `json:"type"`
	Status  string `json:"status"`
	Message string `json:"message"`
	Project string `json:"project,omitempty"`
}

func NewVoiceSession(id string, ws *websocket.Conn, openaiAPIKey string, relayManager *RelayManager) (*VoiceSession, error) {
	ctx, cancel := context.WithCancel(context.Background())
	
	// Initialize OpenAI Realtime client
	openaiClient, err := NewOpenAIRealtimeClient(openaiAPIKey)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize OpenAI client: %w", err)
	}

	session := &VoiceSession{
		ID:           id,
		WebSocket:    ws,
		OpenAIConn:   openaiClient,
		RelayManager: relayManager,
		Context:      ctx,
		Cancel:       cancel,
		LastActivity: time.Now(),
	}

	return session, nil
}

func (vs *VoiceSession) Start() {
	defer vs.Cancel()

	// Start message handlers
	var wg sync.WaitGroup
	
	wg.Add(2)
	go func() {
		defer wg.Done()
		vs.handleWebSocketMessages()
	}()
	
	go func() {
		defer wg.Done()
		vs.handleOpenAIMessages()
	}()

	// Send welcome message
	vs.sendStatusMessage("connected", "Voice session started. Press record to begin.", "")

	wg.Wait()
}

func (vs *VoiceSession) handleWebSocketMessages() {
	for {
		select {
		case <-vs.Context.Done():
			return
		default:
			var msg WSMessage
			err := vs.WebSocket.ReadJSON(&msg)
			if err != nil {
				log.Printf("WebSocket error: %v", err)
				return
			}

			vs.updateActivity()
			vs.handleClientMessage(msg)
		}
	}
}

func (vs *VoiceSession) handleClientMessage(msg WSMessage) {
	switch msg.Type {
	case "audio":
		vs.handleAudioMessage(msg)
	case "start_recording":
		vs.handleStartRecording()
	case "stop_recording":
		vs.handleStopRecording()
	case "select_project":
		vs.handleSelectProject(msg)
	default:
		log.Printf("Unknown message type: %s", msg.Type)
	}
}

func (vs *VoiceSession) handleAudioMessage(msg WSMessage) {
	var audioData []byte
	var err error

	// Extract audio data from message
	if rawBytes, ok := msg.Data.([]byte); ok {
		// Direct byte array
		audioData = rawBytes
	} else if dataMap, ok := msg.Data.(map[string]interface{}); ok {
		if audioStr, ok := dataMap["audio_data"].(string); ok {
			// Base64 encoded string
			log.Printf("Received base64 audio data: %d chars", len(audioStr))
			audioData, err = base64.StdEncoding.DecodeString(audioStr)
			if err != nil {
				log.Printf("Failed to decode base64 audio: %v", err)
				vs.sendStatusMessage("error", "Failed to decode audio data", "")
				return
			}
		} else {
			log.Printf("Invalid audio data format in map")
			return
		}
	} else if audioStr, ok := msg.Data.(string); ok {
		// Direct base64 string
		log.Printf("Received base64 audio string: %d chars", len(audioStr))
		audioData, err = base64.StdEncoding.DecodeString(audioStr)
		if err != nil {
			log.Printf("Failed to decode base64 audio: %v", err)
			vs.sendStatusMessage("error", "Failed to decode audio data", "")
			return
		}
	} else {
		log.Printf("Invalid audio data format: %T", msg.Data)
		return
	}

	log.Printf("Decoded audio data: %d bytes", len(audioData))

	// Send audio to OpenAI
	err = vs.OpenAIConn.SendAudio(audioData)
	if err != nil {
		log.Printf("Failed to send audio to OpenAI: %v", err)
		vs.sendStatusMessage("error", "Failed to process audio", "")
	}
}

func (vs *VoiceSession) handleStartRecording() {
	vs.mu.Lock()
	vs.IsRecording = true
	vs.mu.Unlock()
	
	vs.sendStatusMessage("connecting", "Connecting to voice assistant...", vs.CurrentProject)
	
	// Step 1: Start basic OpenAI session (no tools)
	err := vs.OpenAIConn.StartSession()
	if err != nil {
		log.Printf("Failed to start OpenAI session: %v", err)
		vs.sendStatusMessage("error", "Failed to connect to voice assistant", "")
		vs.mu.Lock()
		vs.IsRecording = false
		vs.mu.Unlock()
		return
	}
	
	log.Printf("OpenAI session started successfully")
	
	// Step 2: Configure tools after connection is established
	err = vs.configureOpenAITools()
	if err != nil {
		log.Printf("Failed to configure OpenAI tools: %v", err)
		vs.sendStatusMessage("error", "Failed to configure voice features", "")
		vs.mu.Lock()
		vs.IsRecording = false
		vs.mu.Unlock()
		return
	}
	
	// Step 3: Apply tools configuration to the session
	err = vs.OpenAIConn.ConfigureSessionTools()
	if err != nil {
		log.Printf("Failed to apply tools configuration: %v", err)
		vs.sendStatusMessage("error", "Failed to configure voice features", "")
		vs.mu.Lock()
		vs.IsRecording = false
		vs.mu.Unlock()
		return
	}
	
	log.Printf("OpenAI tools configured successfully")
	vs.sendStatusMessage("recording", "Recording started - speak now", vs.CurrentProject)
}

func (vs *VoiceSession) handleStopRecording() {
	vs.mu.Lock()
	vs.IsRecording = false
	vs.mu.Unlock()
	
	vs.sendStatusMessage("processing", "Processing voice command...", vs.CurrentProject)
	
	// Commit the accumulated audio and request OpenAI response
	err := vs.OpenAIConn.CommitAudio()
	if err != nil {
		log.Printf("Failed to commit audio: %v", err)
		vs.sendStatusMessage("error", "Failed to process audio", "")
	}
}

func (vs *VoiceSession) handleSelectProject(msg WSMessage) {
	projectName, ok := msg.Data.(string)
	if !ok {
		if dataMap, ok := msg.Data.(map[string]interface{}); ok {
			projectName, _ = dataMap["project"].(string)
		}
	}

	if projectName == "" {
		vs.sendStatusMessage("error", "Project name required", "")
		return
	}

	err := vs.RelayManager.SelectProject(projectName)
	if err != nil {
		vs.sendStatusMessage("error", fmt.Sprintf("Failed to select project: %v", err), "")
		return
	}

	vs.mu.Lock()
	vs.CurrentProject = projectName
	vs.mu.Unlock()

	vs.sendStatusMessage("project_selected", fmt.Sprintf("Selected project: %s", projectName), projectName)
	
	// Update OpenAI context with project information
	vs.updateOpenAIContext()
}

func (vs *VoiceSession) handleOpenAIMessages() {
	for {
		select {
		case <-vs.Context.Done():
			return
		default:
			response, err := vs.OpenAIConn.ReceiveResponse()
			if err != nil {
				if err != io.EOF {
					log.Printf("OpenAI response error: %v", err)
				}
				return
			}

			vs.handleOpenAIResponse(response)
		}
	}
}

func (vs *VoiceSession) handleOpenAIResponse(response *OpenAIResponse) {
	switch response.Type {
	case "audio":
		// Send audio response back to client
		audioMsg := WSMessage{
			Type: "audio_response",
			Data: map[string]interface{}{
				"audio_data": response.AudioData,
			},
		}
		vs.sendMessage(audioMsg)

	case "transcription":
		// Send transcription to client
		transcriptionMsg := WSMessage{
			Type: "transcription",
			Data: map[string]interface{}{
				"text": response.Text,
			},
		}
		vs.sendMessage(transcriptionMsg)

	case "function_call":
		// Execute Relay function
		vs.executeFunctionCall(response.FunctionCall)

	case "error":
		vs.sendStatusMessage("error", response.Error, vs.CurrentProject)

	default:
		log.Printf("Unknown OpenAI response type: %s", response.Type)
	}
}

func (vs *VoiceSession) executeFunctionCall(funcCall *FunctionCall) {
	vs.sendStatusMessage("executing", fmt.Sprintf("Executing: %s", funcCall.Name), vs.CurrentProject)

	result, err := vs.RelayManager.ExecuteFunction(vs.CurrentProject, funcCall.Name, funcCall.Arguments)
	if err != nil {
		log.Printf("Function execution error: %v", err)
		vs.sendStatusMessage("error", fmt.Sprintf("Failed to execute %s: %v", funcCall.Name, err), vs.CurrentProject)
		
		// Send error back to OpenAI
		vs.OpenAIConn.SendFunctionResult(funcCall.CallID, nil, err.Error())
		return
	}

	// Send success status
	vs.sendStatusMessage("completed", fmt.Sprintf("Completed: %s", funcCall.Name), vs.CurrentProject)

	// Send result back to OpenAI
	vs.OpenAIConn.SendFunctionResult(funcCall.CallID, result, "")

	// Send result to client
	resultMsg := WSMessage{
		Type: "function_result",
		Data: map[string]interface{}{
			"function": funcCall.Name,
			"result":   result,
		},
	}
	vs.sendMessage(resultMsg)
}

func (vs *VoiceSession) configureOpenAITools() error {
	tools := []OpenAITool{
		{
			Name:        "create_github_issue",
			Description: "Create a new GitHub issue",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"title": map[string]interface{}{
						"type":        "string",
						"description": "The title of the issue",
					},
					"body": map[string]interface{}{
						"type":        "string",
						"description": "The body/description of the issue",
					},
					"labels": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Labels to add to the issue",
					},
				},
				"required": []string{"title"},
			},
		},
		{
			Name:        "update_github_issue",
			Description: "Update an existing GitHub issue",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"number": map[string]interface{}{
						"type":        "number",
						"description": "The issue number to update",
					},
					"title": map[string]interface{}{
						"type":        "string",
						"description": "New title for the issue",
					},
					"body": map[string]interface{}{
						"type":        "string",
						"description": "New body for the issue",
					},
				},
				"required": []string{"number"},
			},
		},
		{
			Name:        "git_commit",
			Description: "Create a smart git commit",
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "git_status",
			Description: "Get git repository status",
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "list_issues",
			Description: "List GitHub issues for the current project",
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
	}

	return vs.OpenAIConn.ConfigureTools(tools)
}

func (vs *VoiceSession) updateOpenAIContext() {
	if vs.CurrentProject == "" {
		return
	}

	status, err := vs.RelayManager.GetProjectStatus(vs.CurrentProject)
	if err != nil {
		log.Printf("Failed to get project status: %v", err)
		return
	}

	contextUpdate := fmt.Sprintf(`
You are controlling Relay for project: %s
Current working directory: %s
Available commands: create_github_issue, update_github_issue, git_commit, git_status, list_issues

When the user gives voice commands, convert them to appropriate function calls.
Examples:
- "Create an issue for adding user authentication" -> create_github_issue
- "Update issue 5 to mark it as completed" -> update_github_issue  
- "Commit my changes" -> git_commit
- "What's the git status?" -> git_status
- "Show me the open issues" -> list_issues

Always be helpful and confirm what actions you're taking.
	`, vs.CurrentProject, status["path"])

	vs.OpenAIConn.UpdateContext(contextUpdate)
}

func (vs *VoiceSession) sendMessage(msg WSMessage) {
	vs.mu.RLock()
	defer vs.mu.RUnlock()

	if err := vs.WebSocket.WriteJSON(msg); err != nil {
		log.Printf("Failed to send WebSocket message: %v", err)
	}
}

func (vs *VoiceSession) sendStatusMessage(status, message, project string) {
	statusMsg := WSMessage{
		Type: "status",
		Data: StatusMessage{
			Type:    "status",
			Status:  status,
			Message: message,
			Project: project,
		},
	}
	vs.sendMessage(statusMsg)
}

func (vs *VoiceSession) updateActivity() {
	vs.mu.Lock()
	vs.LastActivity = time.Now()
	vs.mu.Unlock()
}

func (vs *VoiceSession) Close() {
	vs.Cancel()
	if vs.OpenAIConn != nil {
		vs.OpenAIConn.Close()
	}
}
