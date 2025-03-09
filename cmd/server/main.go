package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"gochat/internal/handlers"
	"gochat/internal/websocket"
)

func main() {
	// Create a new hub
	hub := websocket.NewHub()
	go hub.Run()

	// Setup handlers without database
	handlers.SetupHandlers(nil, hub)
	log.Println("Running in memory-only mode. Messages will not be persisted.")

	// Setup router
	router := gin.Default()

	// Serve static files
	router.Static("/static", "./static")
	router.StaticFile("/", "./static/index.html")

	// Setup routes
	setupRoutes(router)

	// Get server port from environment variable
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8082"
	}

	// Start server
	log.Printf("Server starting on port %s...", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// setupRoutes configures the HTTP routes
func setupRoutes(router *gin.Engine) {
	// WebSocket endpoint
	router.GET("/ws", handlers.WebSocketHandler)

	// File upload endpoint
	router.POST("/upload", handlers.UploadHandler)

	// Get users endpoint
	router.GET("/users", handlers.GetUsersHandler)
}
