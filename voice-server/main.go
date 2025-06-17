package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/websocket/v2"
)

type VoiceServer struct {
	openaiAPIKey string
	relayManager *RelayManager
	sessions     map[string]*VoiceSession
}

func main() {
	// Get OpenAI API key from environment
	openaiAPIKey := os.Getenv("OPENAI_API_KEY")
	if openaiAPIKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	// Initialize the voice server
	server, err := NewVoiceServer(openaiAPIKey)
	if err != nil {
		log.Fatalf("Failed to initialize voice server: %v", err)
	}

	// Create Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			log.Printf("Error: %v", err)
			return c.Status(500).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Middleware
	app.Use(logger.New())
	app.Use(cors.New())

	// Health check endpoint
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "healthy",
			"service": "relay-voice-server",
		})
	})

	// WebSocket upgrade middleware
	app.Use("/voice", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	// Voice WebSocket endpoint
	app.Get("/voice", websocket.New(server.handleVoiceConnection))

	// API endpoints for project management
	api := app.Group("/api")
	
	api.Get("/projects", server.handleListProjects)
	api.Post("/projects/:name/select", server.handleSelectProject)
	api.Get("/projects/:name/status", server.handleProjectStatus)

	// Serve static files (for the web client)
	app.Static("/", "./web/build", fiber.Static{
		Index: "index.html",
	})

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Voice server starting on port %s", port)
	log.Fatal(app.Listen(":" + port))
}

func NewVoiceServer(openaiAPIKey string) (*VoiceServer, error) {
	// Initialize Relay manager
	relayManager, err := NewRelayManager()
	if err != nil {
		return nil, err
	}

	return &VoiceServer{
		openaiAPIKey: openaiAPIKey,
		relayManager: relayManager,
		sessions:     make(map[string]*VoiceSession),
	}, nil
}

func (vs *VoiceServer) handleListProjects(c *fiber.Ctx) error {
	projects, err := vs.relayManager.ListProjects()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to list projects",
		})
	}

	return c.JSON(fiber.Map{
		"projects": projects,
	})
}

func (vs *VoiceServer) handleSelectProject(c *fiber.Ctx) error {
	projectName := c.Params("name")
	
	err := vs.relayManager.SelectProject(projectName)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Failed to select project",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Project selected successfully",
		"project": projectName,
	})
}

func (vs *VoiceServer) handleProjectStatus(c *fiber.Ctx) error {
	projectName := c.Params("name")
	
	status, err := vs.relayManager.GetProjectStatus(projectName)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Failed to get project status",
		})
	}

	return c.JSON(status)
}

func (vs *VoiceServer) handleVoiceConnection(c *websocket.Conn) {
	defer c.Close()

	// Generate session ID
	sessionID := generateSessionID()
	log.Printf("New voice connection: %s", sessionID)

	// Create voice session
	session, err := NewVoiceSession(sessionID, c, vs.openaiAPIKey, vs.relayManager)
	if err != nil {
		log.Printf("Failed to create voice session: %v", err)
		return
	}

	// Store session
	vs.sessions[sessionID] = session

	// Handle session
	session.Start()

	// Cleanup
	delete(vs.sessions, sessionID)
	log.Printf("Voice session ended: %s", sessionID)
}

func generateSessionID() string {
	// Simple session ID generation
	// In production, use proper UUID
	return "session_" + randomString(8)
}

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[i%len(charset)] // Better distribution
	}
	return string(b)
}
