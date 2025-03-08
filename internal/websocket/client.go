package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"
	"unicode"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second
	
	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second
	
	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10
	
	// Maximum message size allowed from peer
	maxMessageSize = 512 * 1024 // 512KB
	
	// Maximum number of reconnection attempts
	maxRetryCount = 5
	
	// Interval between reconnection attempts
	retryInterval = 3 * time.Second
)

// Upgrader upgrades HTTP connections to WebSocket connections
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins in development environment
		return true
	},
}

// Client represents a connected chat user
type Client struct {
	hub         *Hub
	conn        *websocket.Conn
	send        chan []byte
	Name        string
	retryCount  int
	isConnected bool
	lastActivity time.Time
}

// NewClient creates a new client instance
func NewClient(conn *websocket.Conn, hub *Hub, username string) *Client {
	return &Client{
		hub:         hub,
		conn:        conn,
		send:        make(chan []byte, 256),
		Name:        username,
		retryCount:  0,
		isConnected: true,
		lastActivity: time.Now(),
	}
}

// ReadPump pumps messages from the WebSocket connection to the hub
func (c *Client) ReadPump() {
	defer func() {
		log.Printf("Client %s disconnected, cleaning up", c.Name)
		c.hub.unregister <- c
		c.conn.Close()
		c.isConnected = false
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		c.lastActivity = time.Now()
		return nil
	})

	for {
		_, rawMessage, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error for user %s: %v", c.Name, err)
			}
			break
		}
		
		c.lastActivity = time.Now()

		// Parse and validate message
		var msg Message
		if err := json.Unmarshal(rawMessage, &msg); err != nil {
			log.Printf("Error parsing message from user %s: %v", c.Name, err)
			errorMsg := Message{
				Type:    MessageTypeError,
				Content: "Invalid message format",
				Sender:  "System",
				SentAt:  time.Now(),
			}
			errorJSON, _ := json.Marshal(errorMsg)
			c.send <- errorJSON
			continue
		}

		// Set sender to client's name to prevent spoofing
		msg.Sender = c.Name
		msg.SentAt = time.Now()

		// Handle different message types
		switch msg.Type {
		case "get_history":
			log.Printf("User %s requested message history", c.Name)
			messages, err := c.hub.GetMessageHistory()
			if err != nil {
				log.Printf("Error getting message history for user %s: %v", c.Name, err)
				errorMsg := Message{
					Type:    MessageTypeError,
					Content: "Failed to retrieve message history",
					Sender:  "System",
					SentAt:  time.Now(),
				}
				errorJSON, _ := json.Marshal(errorMsg)
				c.send <- errorJSON
				continue
			}

			// Send each message to the client
			for _, historyMsg := range messages {
				msgJSON, _ := json.Marshal(historyMsg)
				c.send <- msgJSON
			}

		case MessageTypeText:
			// Validate message content
			if strings.TrimSpace(msg.Content) == "" {
				continue // Ignore empty messages
			}

			// Sanitize message content
			msg.Content = sanitizeMessage(msg.Content)

			// Save message to database
			if err := c.hub.db.Create(&msg).Error; err != nil {
				log.Printf("Error saving message from user %s: %v", c.Name, err)
			}

			// Broadcast message to all clients
			msgJSON, _ := json.Marshal(msg)
			c.hub.broadcast <- msgJSON

		case MessageTypeFile:
			// File messages are handled separately in the upload handler
			// We just broadcast the message here
			msgJSON, _ := json.Marshal(msg)
			c.hub.broadcast <- msgJSON
		}
	}
}

// WritePump pumps messages from the hub to the WebSocket connection
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current websocket message
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
			
			// Check for client inactivity
			if time.Since(c.lastActivity) > pongWait*2 {
				log.Printf("Client %s inactive for too long, disconnecting", c.Name)
				return
			}
		}
	}
}

// sanitizeMessage cleans up message content
func sanitizeMessage(message string) string {
	// Trim whitespace
	message = strings.TrimSpace(message)
	
	// Limit message length to prevent abuse
	const maxLength = 2000
	if len(message) > maxLength {
		message = message[:maxLength]
	}
	
	// Remove control characters
	message = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) && r != '\n' && r != '\t' {
			return -1
		}
		return r
	}, message)
	
	return message
}

// ServeWs handles WebSocket requests from clients
func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	// Validate username
	username := r.URL.Query().Get("username")
	username = strings.TrimSpace(username)
	if username == "" {
		log.Println("Username is required")
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}

	// Validate username format
	if !isValidUsername(username) {
		log.Printf("Invalid username format: %s", username)
		http.Error(w, "Invalid username format. Use only letters, numbers, and underscores.", http.StatusBadRequest)
		return
	}

	// Check if username is already taken
	if hub.IsUsernameTaken(username) {
		log.Printf("Username already taken: %s", username)
		http.Error(w, "Username already taken", http.StatusConflict)
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading connection: %v", err)
		return
	}

	// Create new client
	client := NewClient(conn, hub, username)

	// Register client with hub
	hub.Register <- client

	// Start client goroutines
	go client.WritePump()
	go client.ReadPump()
}

// isValidUsername checks if a username contains only valid characters
func isValidUsername(username string) bool {
	if len(username) < 3 || len(username) > 20 {
		return false
	}
	
	for _, r := range username {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
			return false
		}
	}
	
	return true
}
