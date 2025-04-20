package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/faust-lvii/gochat/backend/internal/models"
)

// FriendshipHandler handles friendship requests
type FriendshipHandler struct {
	DB *gorm.DB
}

// NewFriendshipHandler creates a new friendship handler
func NewFriendshipHandler(db *gorm.DB) *FriendshipHandler {
	return &FriendshipHandler{
		DB: db,
	}
}

// ListFriendships lists all friendships for the current user
func (h *FriendshipHandler) ListFriendships(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	if h.DB == nil {
		// Veritabanı bağlantısı yoksa boş liste döndür
		c.JSON(http.StatusOK, []gin.H{})
		return
	}

	var friendships []models.Friendship
	if err := h.DB.Where("user_id = ? OR friend_id = ?", userID, userID).Find(&friendships).Error; err != nil {
		log.Printf("Error fetching friendships: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch friendships"})
		return
	}

	c.JSON(http.StatusOK, friendships)
}

// CreateFriendship creates a new friendship
func (h *FriendshipHandler) CreateFriendship(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	type FriendshipRequest struct {
		FriendID uint `json:"friend_id" binding:"required"`
	}

	var req FriendshipRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if h.DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database not available"})
		return
	}

	// Check if friendship already exists
	var existingFriendship models.Friendship
	if err := h.DB.Where(
		"(user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)",
		userID, req.FriendID, req.FriendID, userID,
	).First(&existingFriendship).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Friendship already exists"})
		return
	}

	// Create new friendship
	friendship := models.Friendship{
		UserID:    userID.(uint),
		FriendID:  req.FriendID,
		Status:    "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := h.DB.Create(&friendship).Error; err != nil {
		log.Printf("Error creating friendship: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create friendship"})
		return
	}

	c.JSON(http.StatusCreated, friendship)
}

// UpdateFriendship updates a friendship
func (h *FriendshipHandler) UpdateFriendship(c *gin.Context) {
	id := c.Param("id")
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	type FriendshipUpdateRequest struct {
		Status string `json:"status" binding:"required"`
	}

	var req FriendshipUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate status
	if req.Status != "accepted" && req.Status != "rejected" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status, must be 'accepted' or 'rejected'"})
		return
	}

	if h.DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database not available"})
		return
	}

	var friendship models.Friendship
	if err := h.DB.First(&friendship, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Friendship not found"})
		return
	}

	// Check if user is the recipient of the friendship request
	if friendship.FriendID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot update this friendship"})
		return
	}

	// Update friendship status
	friendship.Status = req.Status
	friendship.UpdatedAt = time.Now()

	if err := h.DB.Save(&friendship).Error; err != nil {
		log.Printf("Error updating friendship: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update friendship"})
		return
	}

	c.JSON(http.StatusOK, friendship)
}

// DeleteFriendship deletes a friendship
func (h *FriendshipHandler) DeleteFriendship(c *gin.Context) {
	id := c.Param("id")
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	if h.DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database not available"})
		return
	}

	var friendship models.Friendship
	if err := h.DB.First(&friendship, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Friendship not found"})
		return
	}

	// Check if user is part of the friendship
	if friendship.UserID != userID.(uint) && friendship.FriendID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot delete this friendship"})
		return
	}

	if err := h.DB.Delete(&friendship).Error; err != nil {
		log.Printf("Error deleting friendship: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete friendship"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Friendship deleted"})
}
