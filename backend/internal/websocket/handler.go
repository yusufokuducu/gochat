package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"

	"github.com/faust-lvii/gochat/backend/internal/models"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins for development
		},
	}
)

// Handler handles WebSocket connections
type Handler struct {
	Manager    *ConnectionManager
	DB         *gorm.DB
	JWTSecret  []byte
}

// NewHandler creates a new WebSocket handler
func NewHandler(db *gorm.DB, jwtSecret []byte) *Handler {
	return &Handler{
		Manager:    NewConnectionManager(),
		DB:         db,
		JWTSecret:  jwtSecret,
	}
}

// HandleConnection handles a WebSocket connection
func (h *Handler) HandleConnection(c *gin.Context) {
	// Token'ı header'dan al
	authHeader := c.GetHeader("Sec-WebSocket-Protocol")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
		return
	}

	// Authenticate user from token
	claims := &models.Claims{}
	token, err := jwt.ParseWithClaims(authHeader, claims, func(token *jwt.Token) (interface{}, error) {
		// Algoritma kontrolü
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return h.JWTSecret, nil
	})

	if err != nil || !token.Valid {
		log.Printf("WebSocket authentication error: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
		return
	}

	// Upgrade to WebSocket with token as protocol
	upgrader.Subprotocols = []string{authHeader}
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade to WebSocket: %v", err)
		return
	}
	defer conn.Close()

	// Register connection
	userID := claims.UserID
	h.Manager.Connect(userID, conn)
	defer h.Manager.Disconnect(userID)

	// Ping/Pong handler for connection keep-alive
	conn.SetPingHandler(func(message string) error {
		err := conn.WriteControl(websocket.PongMessage, []byte(message), time.Now().Add(time.Second*5))
		if err != nil {
			log.Printf("WebSocket ping error: %v", err)
		}
		return err
	})

	// Handle WebSocket messages
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket unexpected close error: %v", err)
			}
			break
		}

		// Parse message
		var wsMsg models.WebSocketMessage
		if err := json.Unmarshal(msg, &wsMsg); err != nil {
			log.Printf("WebSocket message parse error: %v", err)
			conn.WriteJSON(gin.H{"error": "Invalid message format"})
			continue
		}

		// Validate sender
		if wsMsg.SenderID != userID {
			conn.WriteJSON(gin.H{"error": "Sender ID does not match authenticated user"})
			continue
		}

		// Validate message content
		if wsMsg.Content == "" {
			conn.WriteJSON(gin.H{"error": "Message content cannot be empty"})
			continue
		}

		if h.DB == nil {
			// Veritabanı yoksa test mesajı döndür
			conn.WriteJSON(gin.H{"status": "sent", "message_id": 0})
			continue
		}

		// Create message in database
		message := models.Message{
			SenderID:   wsMsg.SenderID,
			ReceiverID: wsMsg.ReceiverID,
			Content:    wsMsg.Content,
			IsRead:     false,
			Timestamp:  time.Now(),
		}

		if err := h.DB.Create(&message).Error; err != nil {
			log.Printf("Error saving WebSocket message: %v", err)
			conn.WriteJSON(gin.H{"error": "Failed to save message"})
			continue
		}

		// Send to receiver if online
		messageJSON, _ := json.Marshal(message)
		if err := h.Manager.SendPersonalMessage(wsMsg.ReceiverID, messageJSON); err != nil {
			log.Printf("Failed to send message to user %d: %v", wsMsg.ReceiverID, err)
		}

		// Send confirmation to sender
		conn.WriteJSON(gin.H{"status": "sent", "message_id": message.ID})
	}
}
