package handlers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	gorillaws "github.com/gorilla/websocket"
	ws "gochat/internal/websocket"
	"gorm.io/gorm"
)

// Database instance
var DB *gorm.DB

// WebSocket upgrader
var upgrader = gorillaws.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for now
	},
}

// Hub instance
var Hub *ws.Hub

// SetupHandlers initializes the handlers with the database and hub
func SetupHandlers(db *gorm.DB, hub *ws.Hub) {
	DB = db
	Hub = hub
}

// HomeHandler serves the main page
func HomeHandler(c *gin.Context) {
	c.File("./static/index.html")
}

// WebSocketHandler handles WebSocket connections
func WebSocketHandler(c *gin.Context) {
	// Get username from query parameter
	username := c.Query("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username is required"})
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Failed to upgrade connection:", err)
		return
	}

	// Create a new client
	client := ws.NewClient(Hub, conn, username)

	// Register client with the hub
	Hub.Register <- client

	// Start goroutines for reading and writing
	go client.ReadPump()
	go client.WritePump()

	// Send message history to the client
	go sendMessageHistory(client)
}

// UploadHandler handles file uploads
func UploadHandler(c *gin.Context) {
	// Get username from form
	username := c.PostForm("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username is required"})
		return
	}

	// Get file from form
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	// Check file size
	maxSize, _ := strconv.ParseInt(os.Getenv("UPLOAD_MAX_SIZE"), 10, 64)
	if maxSize == 0 {
		maxSize = 10 * 1024 * 1024 // Default 10MB
	}
	if file.Size > maxSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File too large"})
		return
	}

	// Create uploads directory if it doesn't exist
	if err := os.MkdirAll("uploads", 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
		return
	}

	// Generate unique filename
	filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), file.Filename)
	filepath := filepath.Join("uploads", filename)

	// Save the file
	if err := c.SaveUploadedFile(file, filepath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	// Create attachment record
	attachment := ws.Attachment{
		FileName:   file.Filename,
		FileSize:   file.Size,
		FileType:   file.Header.Get("Content-Type"),
		FilePath:   filename,
		UploadedAt: time.Now(),
	}

	// Create message with attachment
	message := ws.Message{
		Type:        ws.MessageTypeFile,
		Content:     fmt.Sprintf("%s shared a file: %s", username, file.Filename),
		Sender:      username,
		SentAt:      time.Now(),
		Attachments: []ws.Attachment{attachment},
	}

	// Save message to database if connected
	if DB != nil {
		if err := DB.Create(&message).Error; err != nil {
			log.Println("Failed to save message:", err)
		}
	} else {
		log.Println("Database not connected, skipping message persistence")
	}

	// Broadcast message
	Hub.Broadcast <- &message

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"file":    file.Filename,
		"filePath": filename,
	})
}

// GetUsersHandler returns a list of online users
func GetUsersHandler(c *gin.Context) {
	users := Hub.GetOnlineUsers()
	c.JSON(http.StatusOK, users)
}

// sendMessageHistory sends the message history to a new client
func sendMessageHistory(client *ws.Client) {
	// If database is not connected, skip loading message history
	if DB == nil {
		log.Println("Database not connected, skipping message history")
		return
	}
	
	var messages []ws.Message
	
	// Get the last 50 messages
	DB.Preload("Attachments").Order("sent_at desc").Limit(50).Find(&messages)

	// Reverse the order to send oldest first
	for i := len(messages) - 1; i >= 0; i-- {
		client.Send <- &messages[i]
	}
}
