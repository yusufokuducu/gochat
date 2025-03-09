package websocket

import (
	"encoding/json"
	"log"
	"sync"
	"time"
)

// Hub maintains the set of active clients and broadcasts messages to the clients
type Hub struct {
	// Registered clients
	Clients map[*Client]bool

	// Inbound messages from the clients
	Broadcast chan *Message

	// Register requests from the clients
	Register chan *Client

	// Unregister requests from clients
	Unregister chan *Client

	// User status tracking
	UserStatus     map[string]bool
	UserStatusLock sync.RWMutex
}

// NewHub creates a new hub instance
func NewHub() *Hub {
	return &Hub{
		// Use buffered channels for better performance
		Broadcast:  make(chan *Message, 256),
		Register:   make(chan *Client, 32),
		Unregister: make(chan *Client, 32),
		Clients:    make(map[*Client]bool),
		UserStatus: make(map[string]bool),
	}
}

// Run starts the hub and handles client connections and messages
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			log.Printf("Registering client: %s (total clients: %d)", client.Username, len(h.Clients)+1)
			h.Clients[client] = true
			
			// Update user status
			h.UserStatusLock.Lock()
			h.UserStatus[client.Username] = true
			h.UserStatusLock.Unlock()
			
			// Broadcast user joined message
			joinMsg := &Message{
				Type:    MessageTypeSystem,
				Content: client.Username + " has joined the chat",
				Sender:  "System",
				SentAt:  time.Now(),
			}
			h.Broadcast <- joinMsg
			
			// Send current online users to the new client
			go h.sendUserStatus(client)
			
		case client := <-h.Unregister:
			if _, ok := h.Clients[client]; ok {
				log.Printf("Unregistering client: %s (remaining clients: %d)", client.Username, len(h.Clients)-1)
				delete(h.Clients, client)
				close(client.Send)
				
				// Update user status
				h.UserStatusLock.Lock()
				delete(h.UserStatus, client.Username)
				h.UserStatusLock.Unlock()
				
				// Broadcast user left message
				leftMsg := &Message{
					Type:    MessageTypeSystem,
					Content: client.Username + " has left the chat",
					Sender:  "System",
					SentAt:  time.Now(),
				}
				h.Broadcast <- leftMsg
			}
			
		case message := <-h.Broadcast:
			// Log the message being broadcast
			msgJSON, _ := json.Marshal(message)
			log.Printf("Broadcasting message: %s to %d clients", string(msgJSON), len(h.Clients))
			
			// Broadcast message to all clients
			for client := range h.Clients {
				select {
				case client.Send <- message:
					// Message sent successfully
				default:
					// Client's send buffer is full, assume they're gone
					log.Printf("Client %s send buffer full, closing connection", client.Username)
					close(client.Send)
					delete(h.Clients, client)
					
					// Update user status
					h.UserStatusLock.Lock()
					delete(h.UserStatus, client.Username)
					h.UserStatusLock.Unlock()
				}
			}
		}
	}
}

// GetOnlineUsers returns a list of currently online users
func (h *Hub) GetOnlineUsers() []UserStatus {
	h.UserStatusLock.RLock()
	defer h.UserStatusLock.RUnlock()
	
	users := make([]UserStatus, 0, len(h.UserStatus))
	for username, online := range h.UserStatus {
		users = append(users, UserStatus{
			Username: username,
			Online:   online,
		})
	}
	
	return users
}

// sendUserStatus sends the current online users to a client
func (h *Hub) sendUserStatus(client *Client) {
	// Get online users
	users := h.GetOnlineUsers()
	
	// Create a message with the user status
	statusMessage := &Message{
		Type:    MessageTypeSystem,
		Content: "userStatus",
		Sender:  "System",
		SentAt:  time.Now(),
		Data:    users, // Add the users data to the message
	}
	
	// Send to the client with timeout
	select {
	case client.Send <- statusMessage:
		log.Println("Sent user status to client:", client.Username)
	case <-time.After(5 * time.Second):
		log.Println("Timeout sending user status to client:", client.Username)
	}
}
