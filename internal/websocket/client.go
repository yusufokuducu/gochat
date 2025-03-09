package websocket

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512 * 1024 // 512KB
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

// Client is a middleman between the websocket connection and the hub
type Client struct {
	Hub *Hub

	// The websocket connection
	Conn *websocket.Conn

	// Buffered channel of outbound messages
	Send chan *Message

	// User information
	Username string
}

// NewClient creates a new client instance
func NewClient(hub *Hub, conn *websocket.Conn, username string) *Client {
	return &Client{
		Hub:      hub,
		Conn:     conn,
		Send:     make(chan *Message, 256),
		Username: username,
	}
}

// ReadPump pumps messages from the websocket connection to the hub
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
		log.Printf("ReadPump stopped for client: %s", c.Username)
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		log.Printf("Received pong from client: %s", c.Username)
		return nil
	})

	// Send a test message to confirm connection
	testMsg := &Message{
		Type:    MessageTypeSystem,
		Content: "Connection established",
		Sender:  "System",
		SentAt:  time.Now(),
	}
	c.Send <- testMsg
	log.Printf("Sent test message to client: %s", c.Username)

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Unexpected close error from client %s: %v", c.Username, err)
			} else {
				log.Printf("Normal close from client %s: %v", c.Username, err)
			}
			break
		}

		log.Printf("Received message from client %s: %s", c.Username, string(message))

		// Parse the message
		var clientMsg ClientMessage
		if err := json.Unmarshal(message, &clientMsg); err != nil {
			log.Printf("Error parsing message from client %s: %v", c.Username, err)
			continue
		}

		// Create a new message to broadcast
		msg := &Message{
			Type:    clientMsg.Type,
			Content: clientMsg.Content,
			Sender:  c.Username,
			SentAt:  time.Now(),
		}

		// Broadcast the message
		c.Hub.Broadcast <- msg
		log.Printf("Message from client %s queued for broadcast", c.Username)
	}
}

// WritePump pumps messages from the hub to the websocket connection
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
		log.Printf("WritePump stopped for client: %s", c.Username)
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel
				log.Printf("Hub closed channel for client: %s", c.Username)
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Marshal the message to JSON
			jsonMessage, err := json.Marshal(message)
			if err != nil {
				log.Printf("Error marshaling message for client %s: %v", c.Username, err)
				continue
			}

			// Write the message as JSON
			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				log.Printf("Error getting writer for client %s: %v", c.Username, err)
				return
			}

			w.Write(jsonMessage)
			log.Printf("Message sent to client %s: %s", c.Username, string(jsonMessage))

			// Add queued messages to the current websocket message
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				nextMsg := <-c.Send
				nextJSON, _ := json.Marshal(nextMsg)
				w.Write(nextJSON)
				log.Printf("Additional message sent to client %s: %s", c.Username, string(nextJSON))
			}

			if err := w.Close(); err != nil {
				log.Printf("Error closing writer for client %s: %v", c.Username, err)
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("Error sending ping to client %s: %v", c.Username, err)
				return
			}
			log.Printf("Ping sent to client: %s", c.Username)
		}
	}
}
