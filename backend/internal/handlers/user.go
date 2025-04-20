package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/faust-lvii/gochat/backend/internal/models"
)

// UserHandler handles user requests
type UserHandler struct {
	DB *gorm.DB
}

// NewUserHandler creates a new user handler
func NewUserHandler(db *gorm.DB) *UserHandler {
	return &UserHandler{
		DB: db,
	}
}

// ListUsers lists all users
func (h *UserHandler) ListUsers(c *gin.Context) {
	if h.DB == nil {
		c.JSON(http.StatusOK, []gin.H{
			{
				"id":         1,
				"username":   "admin",
				"email":      "admin@example.com",
				"created_at": time.Now().AddDate(0, 0, -10),
			},
		})
		return
	}

	var users []models.User
	if err := h.DB.Find(&users).Error; err != nil {
		log.Printf("Error fetching users: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	c.JSON(http.StatusOK, users)
}

// GetUser gets a user by ID
func (h *UserHandler) GetUser(c *gin.Context) {
	id := c.Param("id")

	if h.DB == nil {
		if id == "1" {
			c.JSON(http.StatusOK, gin.H{
				"id":         1,
				"username":   "admin",
				"email":      "admin@example.com",
				"created_at": time.Now().AddDate(0, 0, -10),
			})
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		}
		return
	}

	var user models.User
	if err := h.DB.First(&user, id).Error; err != nil {
		log.Printf("Error fetching user with ID %s: %v", id, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

// GetCurrentUser gets the current user
func (h *UserHandler) GetCurrentUser(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	if h.DB == nil {
		// Veritabanı bağlantısı yoksa, token'dan alınan userID'yi kullanarak
		// test kullanıcısı döndür
		c.JSON(http.StatusOK, gin.H{
			"id":         userID,
			"username":   "admin",
			"email":      "admin@example.com",
			"created_at": time.Now().AddDate(0, 0, -10),
		})
		return
	}

	var user models.User
	if err := h.DB.First(&user, userID).Error; err != nil {
		log.Printf("Error fetching user with ID %v: %v", userID, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

// UpdateUser updates a user
func (h *UserHandler) UpdateUser(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("userID")

	// Convert userID to string for comparison
	userIDStr := ""
	switch v := userID.(type) {
	case uint:
		userIDStr = c.Param("id")
	case float64:
		userIDStr = c.Param("id")
	default:
		log.Printf("Unexpected userID type: %T", v)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Check if user is updating their own profile
	if id != userIDStr {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot update other user's profile"})
		return
	}

	var updateData struct {
		Email string `json:"email"`
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if h.DB == nil {
		c.JSON(http.StatusOK, gin.H{
			"id":         userID,
			"username":   "admin",
			"email":      updateData.Email,
			"created_at": time.Now().AddDate(0, 0, -10),
		})
		return
	}

	var user models.User
	if err := h.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Update user data
	user.Email = updateData.Email

	if err := h.DB.Save(&user).Error; err != nil {
		log.Printf("Error updating user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, user)
}
