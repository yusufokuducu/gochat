package websocket

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"gorm.io/gorm"
)

// Hub maintains the set of active clients and broadcasts messages to them
type Hub struct {
	// Registered clients
	clients    map[*Client]bool
	
	// Inbound messages from the clients
	broadcast  chan []byte
	
	// Register requests from the clients
	Register   chan *Client
	
	// Unregister requests from clients
	unregister chan *Client
	
	// Database connection
	db         *gorm.DB
	
	// Mutex for thread-safe access to the clients map
	usersMutex sync.RWMutex
}

// NewHub creates a new hub instance
func NewHub(db *gorm.DB) *Hub {
	return &Hub{
		broadcast:  make(chan []byte, 256),
		Register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		db:         db,
		usersMutex: sync.RWMutex{},
	}
}

// GetMessageHistory retrieves the last 50 messages from the database
func (h *Hub) GetMessageHistory() ([]Message, error) {
	var messages []Message
	
	// Query messages with their attachments
	err := h.db.Order("sent_at desc").Limit(50).
		Preload("Attachments").
		Find(&messages).Error
	
	if err != nil {
		return nil, err
	}
	
	// Reverse the order so oldest messages come first
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}
	
	return messages, nil
}

// GetOnlineUsers returns a list of all online users
func (h *Hub) GetOnlineUsers() []string {
	h.usersMutex.RLock()
	defer h.usersMutex.RUnlock()
	
	users := make([]string, 0, len(h.clients))
	for client := range h.clients {
		users = append(users, client.Name)
	}
	return users
}

// IsUsernameTaken checks if a username is already in use
func (h *Hub) IsUsernameTaken(username string) bool {
	h.usersMutex.RLock()
	defer h.usersMutex.RUnlock()
	
	for client := range h.clients {
		if client.Name == username {
			return true
		}
	}
	return false
}

// broadcastUserList sends the current user list to all clients
func (h *Hub) broadcastUserList() {
	users := h.GetOnlineUsers()
	userListMsg := struct {
		Type  string   `json:"type"`
		Users []string `json:"users"`
	}{
		Type:  "user_list",
		Users: users,
	}
	
	messageJSON, err := json.Marshal(userListMsg)
	if err != nil {
		log.Printf("Error marshaling user list: %v", err)
		return
	}
	
	h.broadcast <- messageJSON
}

// broadcastSystemMessage sends a system message to all clients
func (h *Hub) broadcastSystemMessage(content string) {
	systemMsg := Message{
		Type:    MessageTypeSystem,
		Content: content,
		Sender:  "System",
		SentAt:  time.Now(),
	}
	
	// Save system message to database
	if err := h.db.Create(&systemMsg).Error; err != nil {
		log.Printf("Error saving system message: %v", err)
	}
	
	messageJSON, err := json.Marshal(systemMsg)
	if err != nil {
		log.Printf("Error marshaling system message: %v", err)
		return
	}
	
	h.broadcast <- messageJSON
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	log.Println("Starting WebSocket hub")
	
	for {
		select {
		case client := <-h.Register:
			// Register new client
			h.usersMutex.Lock()
			h.clients[client] = true
			h.usersMutex.Unlock()
			
			log.Printf("Client registered: %s (total: %d)", client.Name, len(h.clients))
			
			// Send welcome message to the new client
			welcomeMsg := Message{
				Type:    MessageTypeSystem,
				Content: "Welcome to the chat!",
				Sender:  "System",
				SentAt:  time.Now(),
			}
			welcomeJSON, _ := json.Marshal(welcomeMsg)
			client.send <- welcomeJSON
			
			// Send user list to the new client
			h.sendUserListToClient(client)
			
			// Broadcast user joined message
			h.broadcastSystemMessage(client.Name + " has joined the chat")
			
			// Broadcast updated user list to all clients
			h.broadcastUserList()
			
		case client := <-h.unregister:
			// Unregister client
			h.usersMutex.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Printf("Client unregistered: %s (total: %d)", client.Name, len(h.clients))
				
				// Broadcast user left message
				h.broadcastSystemMessage(client.Name + " has left the chat")
				
				// Broadcast updated user list
				h.broadcastUserList()
			}
			h.usersMutex.Unlock()
			
		case message := <-h.broadcast:
			// Broadcast message to all clients
			h.usersMutex.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
					// Message sent successfully
				default:
					// Client's send buffer is full, unregister client
					h.usersMutex.RUnlock()
					h.usersMutex.Lock()
					close(client.send)
					delete(h.clients, client)
					h.usersMutex.Unlock()
					h.usersMutex.RLock()
					
					log.Printf("Client buffer full, unregistered: %s", client.Name)
				}
			}
			h.usersMutex.RUnlock()
		}
	}
}

// sendUserListToClient sends the current user list to a specific client
func (h *Hub) sendUserListToClient(client *Client) {
	users := h.GetOnlineUsers()
	userListMsg := struct {
		Type  string   `json:"type"`
		Users []string `json:"users"`
	}{
		Type:  "user_list",
		Users: users,
	}
	
	messageJSON, err := json.Marshal(userListMsg)
	if err != nil {
		log.Printf("Error marshaling user list for client %s: %v", client.Name, err)
		return
	}
	
	client.send <- messageJSON
}
