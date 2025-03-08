package handlers

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	gorilla "github.com/gorilla/websocket"
	"gochat/internal/websocket"
)

// Configuration constants
const (
	readBufferSize    = 4096
	writeBufferSize   = 4096
	handshakeTimeout  = 10 * time.Second
	maxMessageSize    = 32 * 1024 // 32KB
	pongWaitTime      = 60 * time.Second
	writeWaitTime     = 10 * time.Second
	defaultUsername   = "Anonymous"
	minUsernameLength = 3
	maxUsernameLength = 20
)

// WebSocket upgrader with security settings
var upgrader = gorilla.Upgrader{
	ReadBufferSize:  readBufferSize,
	WriteBufferSize: writeBufferSize,
	CheckOrigin: func(r *http.Request) bool {
		// In production, you should implement proper origin checking
		// For development, we allow all origins
		return true
	},
	HandshakeTimeout: handshakeTimeout,
}

// WebSocketHandler handles WebSocket connections
type WebSocketHandler struct {
	hub *websocket.Hub
}

// NewWebSocketHandler creates a new WebSocketHandler instance
func NewWebSocketHandler(hub *websocket.Hub) *WebSocketHandler {
	return &WebSocketHandler{
		hub: hub,
	}
}

// ServeWs upgrades HTTP connection to WebSocket and manages the connection
func (h *WebSocketHandler) ServeWs(c *gin.Context) {
	// Validate username before upgrading connection
	username := sanitizeUsername(c.Query("username"))
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid username",
		})
		return
	}

	// Check if username is already taken
	if h.hub.IsUsernameTaken(username) {
		c.JSON(http.StatusConflict, gin.H{
			"error": "Username is already taken. Please choose another one.",
		})
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Error upgrading to WebSocket: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not establish WebSocket connection",
		})
		return
	}

	// Configure connection settings
	conn.SetReadLimit(maxMessageSize)
	conn.SetReadDeadline(time.Now().Add(pongWaitTime))
	conn.SetWriteDeadline(time.Now().Add(writeWaitTime))

	// Log connection
	log.Printf("New WebSocket connection established for user: %s from IP: %s", 
		username, c.ClientIP())

	// Create new client
	client := websocket.NewClient(conn, h.hub, username)

	// Register client to hub
	h.hub.Register <- client

	// Start goroutines for reading and writing
	go client.ReadPump()
	go client.WritePump()
}

// sanitizeUsername validates and sanitizes the username
func sanitizeUsername(username string) string {
	// Trim whitespace
	username = strings.TrimSpace(username)
	
	// If empty or too short, use default
	if len(username) < minUsernameLength {
		return defaultUsername
	}
	
	// If too long, truncate
	if len(username) > maxUsernameLength {
		username = username[:maxUsernameLength]
	}
	
	// Remove any potentially harmful characters
	username = strings.Map(func(r rune) rune {
		if r < 32 || r > 126 {
			return -1 // Remove control characters and non-ASCII
		}
		return r
	}, username)
	
	return username
}
