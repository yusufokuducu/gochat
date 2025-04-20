package websocket

import (
	"fmt"
	"log"

	"github.com/gorilla/websocket"
)

// ConnectionManager manages WebSocket connections
type ConnectionManager struct {
	connections map[uint]*websocket.Conn
}

// NewConnectionManager creates a new connection manager
func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		connections: make(map[uint]*websocket.Conn),
	}
}

// Connect adds a new connection to the manager
func (m *ConnectionManager) Connect(userID uint, conn *websocket.Conn) {
	m.connections[userID] = conn
	log.Printf("User %d connected to WebSocket", userID)
}

// Disconnect removes a connection from the manager
func (m *ConnectionManager) Disconnect(userID uint) {
	delete(m.connections, userID)
	log.Printf("User %d disconnected from WebSocket", userID)
}

// SendPersonalMessage sends a message to a specific user
func (m *ConnectionManager) SendPersonalMessage(userID uint, message []byte) error {
	if conn, ok := m.connections[userID]; ok {
		return conn.WriteMessage(websocket.TextMessage, message)
	}
	return fmt.Errorf("user not connected")
}

// GetConnections returns all active connections
func (m *ConnectionManager) GetConnections() map[uint]*websocket.Conn {
	return m.connections
}

// IsConnected checks if a user is connected
func (m *ConnectionManager) IsConnected(userID uint) bool {
	_, ok := m.connections[userID]
	return ok
}
