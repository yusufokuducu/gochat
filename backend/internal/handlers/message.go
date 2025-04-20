package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/faust-lvii/gochat/backend/internal/models"
)

// MessageHandler handles message requests
type MessageHandler struct {
	DB *gorm.DB
}

// NewMessageHandler creates a new message handler
func NewMessageHandler(db *gorm.DB) *MessageHandler {
	return &MessageHandler{
		DB: db,
	}
}

// ListMessages lists all messages between the current user and a friend
func (h *MessageHandler) ListMessages(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	friendID := c.Query("friend_id")
	if friendID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "friend_id parameter is required"})
		return
	}

	if h.DB == nil {
		// Veritabanı bağlantısı yoksa boş mesaj listesi döndür
		c.JSON(http.StatusOK, []gin.H{})
		return
	}

	var messages []models.Message
	query := h.DB.Where(
		"(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)",
		userID, friendID, friendID, userID,
	).Order("timestamp asc")

	if err := query.Find(&messages).Error; err != nil {
		log.Printf("Error fetching messages: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch messages"})
		return
	}

	c.JSON(http.StatusOK, messages)
}

// CreateMessage creates a new message
func (h *MessageHandler) CreateMessage(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	type MessageRequest struct {
		ReceiverID uint   `json:"receiver_id" binding:"required"`
		Content    string `json:"content" binding:"required"`
	}

	var req MessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Message content cannot be empty"})
		return
	}

	if h.DB == nil {
		// Veritabanı bağlantısı yoksa test mesajı döndür
		c.JSON(http.StatusOK, gin.H{
			"id":         1,
			"sender_id":  userID,
			"receiver_id": req.ReceiverID,
			"content":    req.Content,
			"is_read":    false,
			"timestamp":  time.Now(),
		})
		return
	}

	// Create new message
	message := models.Message{
		SenderID:   userID.(uint),
		ReceiverID: req.ReceiverID,
		Content:    req.Content,
		IsRead:     false,
		Timestamp:  time.Now(),
	}

	if err := h.DB.Create(&message).Error; err != nil {
		log.Printf("Error creating message: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create message"})
		return
	}

	c.JSON(http.StatusCreated, message)
}

// MarkMessageAsRead marks a message as read
func (h *MessageHandler) MarkMessageAsRead(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Message ID is required"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	if h.DB == nil {
		// Veritabanı bağlantısı yoksa başarılı yanıt döndür
		c.JSON(http.StatusOK, gin.H{
			"id":         id,
			"sender_id":  1,
			"receiver_id": userID,
			"content":    "Test message",
			"is_read":    true,
			"timestamp":  time.Now(),
		})
		return
	}

	var message models.Message
	if err := h.DB.First(&message, id).Error; err != nil {
		log.Printf("Error finding message with ID %s: %v", id, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Message not found"})
		return
	}

	// Check if user is the receiver
	if message.ReceiverID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot mark this message as read"})
		return
	}

	message.IsRead = true
	if err := h.DB.Save(&message).Error; err != nil {
		log.Printf("Error updating message with ID %s: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update message"})
		return
	}

	c.JSON(http.StatusOK, message)
}
