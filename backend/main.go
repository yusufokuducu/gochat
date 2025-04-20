package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Global variables
var (
	db        *gorm.DB
	jwtSecret []byte
)

// User model
type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"uniqueIndex;not null" json:"username"`
	Email        string    `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash string    `gorm:"not null" json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

// Friendship model
type Friendship struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null" json:"user_id"`
	FriendID  uint      `gorm:"not null" json:"friend_id"`
	Status    string    `gorm:"not null" json:"status"` // pending, accepted, rejected
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Message model
type Message struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	SenderID   uint      `gorm:"not null" json:"sender_id"`
	ReceiverID uint      `gorm:"not null" json:"receiver_id"`
	Content    string    `gorm:"not null" json:"content"`
	IsRead     bool      `gorm:"default:false" json:"is_read"`
	Timestamp  time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"timestamp"`
}

// JWT Claims struct
type Claims struct {
	UserID uint `json:"user_id"`
	jwt.RegisteredClaims
}

// Login request
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Register request
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// Token response
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

// WebSocket message
type WebSocketMessage struct {
	SenderID   uint   `json:"sender_id"`
	ReceiverID uint   `json:"receiver_id"`
	Content    string `json:"content"`
}

// WebSocket connection manager
type ConnectionManager struct {
	connections map[uint]*websocket.Conn
}

func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		connections: make(map[uint]*websocket.Conn),
	}
}

func (m *ConnectionManager) Connect(userID uint, conn *websocket.Conn) {
	m.connections[userID] = conn
}

func (m *ConnectionManager) Disconnect(userID uint) {
	delete(m.connections, userID)
}

func (m *ConnectionManager) SendPersonalMessage(userID uint, message []byte) error {
	if conn, ok := m.connections[userID]; ok {
		return conn.WriteMessage(websocket.TextMessage, message)
	}
	return fmt.Errorf("user not connected")
}

var manager = NewConnectionManager()
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using default values")
	}

	// Initialize JWT secret
	jwtSecret = []byte(getEnv("SECRET_KEY", "your-secret-key-for-development-only"))

	// Initialize database
	initDB()

	// Set up router
	router := gin.Default()

	// Configure CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Root endpoint
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Welcome to GoChat API",
			"version": "1.0.0",
			"status":  "running",
		})
	})

	// API routes
	api := router.Group("/api")
	{
		// Auth endpoints
		auth := api.Group("/auth")
		{
			auth.POST("/login", loginHandler)
			auth.POST("/register", registerHandler)
		}

		// Protected endpoints
		protected := api.Group("")
		protected.Use(authMiddleware())
		{
			// User endpoints
			protected.GET("/users", listUsersHandler)
			protected.GET("/users/me", getCurrentUserHandler)
			protected.GET("/users/:id", getUserHandler)
			protected.PUT("/users/:id", updateUserHandler)

			// Friendship endpoints
			friendships := protected.Group("/friendships")
			{
				friendships.GET("", listFriendshipsHandler)
				friendships.POST("", createFriendshipHandler)
				friendships.PUT("/:id", updateFriendshipHandler)
				friendships.DELETE("/:id", deleteFriendshipHandler)
			}

			// Message endpoints
			messages := protected.Group("/messages")
			{
				messages.GET("", listMessagesHandler)
				messages.POST("", createMessageHandler)
				messages.PUT("/:id/read", markMessageAsReadHandler)
			}
		}
	}

	// WebSocket endpoint
	router.GET("/api/ws", websocketHandler)

	// Start server
	port := getEnvAsInt("PORT", 8000)
	fmt.Printf("Server running on port %d...\n", port)
	if err := router.Run(fmt.Sprintf(":%d", port)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// Helper function to get environment variables with default values
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// Helper function to get environment variables as integers with default values
func getEnvAsInt(key string, defaultValue int) int {
	if valueStr, exists := os.LookupEnv(key); exists {
		if value, err := strconv.Atoi(valueStr); err == nil {
			return value
		}
	}
	return defaultValue
}

func initDB() {
	var err error
	dsn := getEnv("DATABASE_URL", "postgresql://postgres:postgres@db:5432/gochat")
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Println("Warning: Failed to connect to database:", err)
		log.Println("Running in memory mode without database")
		return
	}

	// Auto migrate the schema
	db.AutoMigrate(&User{}, &Friendship{}, &Message{})

	// Create admin user if not exists
	var count int64
	db.Model(&User{}).Count(&count)
	if count == 0 {
		// Hash password: admin1234
		passwordHash, _ := bcrypt.GenerateFromPassword([]byte("admin1234"), bcrypt.DefaultCost)
		adminUser := User{
			Username:     "admin",
			Email:        "admin@example.com",
			PasswordHash: string(passwordHash),
			CreatedAt:    time.Now(),
		}
		db.Create(&adminUser)
	}
}

// Authentication middleware
func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			return
		}

		// Extract the token
		tokenString := authHeader[7:] // Remove "Bearer " prefix
		claims := &Claims{}

		// Parse the token
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		// Set user ID in context
		c.Set("userID", claims.UserID)
		c.Next()
	}
}

// Login handler
func loginHandler(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find user by username
	var user User
	if db == nil {
		// For testing without database
		if req.Username == "admin" && req.Password == "admin1234" {
			// Generate token
			expirationTime := time.Now().Add(24 * time.Hour)
			claims := &Claims{
				UserID: 1,
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(expirationTime),
				},
			}

			token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
			tokenString, err := token.SignedString(jwtSecret)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
				return
			}

			c.JSON(http.StatusOK, TokenResponse{
				AccessToken: tokenString,
				TokenType:   "bearer",
			})
			return
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	if err := db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	// Generate token
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID: user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
		return
	}

	c.JSON(http.StatusOK, TokenResponse{
		AccessToken: tokenString,
		TokenType:   "bearer",
	})
}

// Register handler
func registerHandler(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database not available"})
		return
	}

	// Check if username already exists
	var existingUser User
	if err := db.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username already registered"})
		return
	}

	// Check if email already exists
	if err := db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email already registered"})
		return
	}

	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Create new user
	user := User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(passwordHash),
		CreatedAt:    time.Now(),
	}

	if err := db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, user)
}

// List users handler
func listUsersHandler(c *gin.Context) {
	if db == nil {
		c.JSON(http.StatusOK, []gin.H{
			{
				"id":         1,
				"username":   "admin",
				"email":      "admin@example.com",
				"created_at": time.Now().AddDate(0, 0, -10),
			},
			{
				"id":         2,
				"username":   "user1",
				"email":      "user1@example.com",
				"created_at": time.Now().AddDate(0, 0, -5),
			},
		})
		return
	}

	var users []User
	if err := db.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}
	c.JSON(http.StatusOK, users)
}

// Get user handler
func getUserHandler(c *gin.Context) {
	id := c.Param("id")

	if db == nil {
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

	var user User
	if err := db.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

// Get current user handler
func getCurrentUserHandler(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	if db == nil {
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

	var user User
	if err := db.First(&user, userID).Error; err != nil {
		log.Printf("Error fetching user with ID %v: %v", userID, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

// Update user handler
func updateUserHandler(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("userID")

	// Convert id to uint for comparison
	var userIDUint uint
	fmt.Sscanf(id, "%d", &userIDUint)

	// Check if user is updating their own profile
	if userID.(uint) != userIDUint {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot update other user's profile"})
		return
	}

	if db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database not available"})
		return
	}

	var user User
	if err := db.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Only allow updating email
	type UpdateUserRequest struct {
		Email string `json:"email" binding:"required,email"`
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user.Email = req.Email
	if err := db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// List friendships handler
func listFriendshipsHandler(c *gin.Context) {
	userID, _ := c.Get("userID")

	if db == nil {
		c.JSON(http.StatusOK, []gin.H{})
		return
	}

	var friendships []Friendship
	if err := db.Where("user_id = ? OR friend_id = ?", userID, userID).Find(&friendships).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch friendships"})
		return
	}

	c.JSON(http.StatusOK, friendships)
}

// Create friendship handler
func createFriendshipHandler(c *gin.Context) {
	userID, _ := c.Get("userID")

	type FriendshipRequest struct {
		FriendID uint `json:"friend_id" binding:"required"`
	}

	var req FriendshipRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database not available"})
		return
	}

	// Check if friendship already exists
	var existingFriendship Friendship
	if err := db.Where(
		"(user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)",
		userID, req.FriendID, req.FriendID, userID,
	).First(&existingFriendship).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Friendship already exists"})
		return
	}

	// Create new friendship
	friendship := Friendship{
		UserID:    userID.(uint),
		FriendID:  req.FriendID,
		Status:    "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := db.Create(&friendship).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create friendship"})
		return
	}

	c.JSON(http.StatusCreated, friendship)
}

// Update friendship handler
func updateFriendshipHandler(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("userID")

	if db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database not available"})
		return
	}

	var friendship Friendship
	if err := db.First(&friendship, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Friendship not found"})
		return
	}

	// Check if user is the friend (can accept/reject)
	if friendship.FriendID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot update this friendship"})
		return
	}

	type UpdateFriendshipRequest struct {
		Status string `json:"status" binding:"required,oneof=accepted rejected"`
	}

	var req UpdateFriendshipRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	friendship.Status = req.Status
	friendship.UpdatedAt = time.Now()

	if err := db.Save(&friendship).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update friendship"})
		return
	}

	c.JSON(http.StatusOK, friendship)
}

// Delete friendship handler
func deleteFriendshipHandler(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("userID")

	if db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database not available"})
		return
	}

	var friendship Friendship
	if err := db.First(&friendship, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Friendship not found"})
		return
	}

	// Check if user is part of the friendship
	if friendship.UserID != userID.(uint) && friendship.FriendID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot delete this friendship"})
		return
	}

	if err := db.Delete(&friendship).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete friendship"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Friendship deleted"})
}

// List messages handler
func listMessagesHandler(c *gin.Context) {
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

	if db == nil {
		// Veritabanı bağlantısı yoksa boş mesaj listesi döndür
		c.JSON(http.StatusOK, []gin.H{})
		return
	}

	var messages []Message
	query := db.Where(
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

// Create message handler
func createMessageHandler(c *gin.Context) {
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

	if db == nil {
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
	message := Message{
		SenderID:   userID.(uint),
		ReceiverID: req.ReceiverID,
		Content:    req.Content,
		IsRead:     false,
		Timestamp:  time.Now(),
	}

	if err := db.Create(&message).Error; err != nil {
		log.Printf("Error creating message: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create message"})
		return
	}

	c.JSON(http.StatusCreated, message)
}

// Mark message as read handler
func markMessageAsReadHandler(c *gin.Context) {
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

	if db == nil {
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

	var message Message
	if err := db.First(&message, id).Error; err != nil {
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
	if err := db.Save(&message).Error; err != nil {
		log.Printf("Error updating message with ID %s: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update message"})
		return
	}

	c.JSON(http.StatusOK, message)
}

// WebSocket handler
func websocketHandler(c *gin.Context) {
	// Token'ı header'dan al
	authHeader := c.GetHeader("Sec-WebSocket-Protocol")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
		return
	}

	// Authenticate user from token
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(authHeader, claims, func(token *jwt.Token) (interface{}, error) {
		// Algoritma kontrolü
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
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
	manager.Connect(userID, conn)
	defer manager.Disconnect(userID)

	log.Printf("User %d connected to WebSocket", userID)

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
		var wsMsg WebSocketMessage
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

		if db == nil {
			// Veritabanı yoksa test mesajı döndür
			conn.WriteJSON(gin.H{"status": "sent", "message_id": 0})
			continue
		}

		// Create message in database
		message := Message{
			SenderID:   wsMsg.SenderID,
			ReceiverID: wsMsg.ReceiverID,
			Content:    wsMsg.Content,
			IsRead:     false,
			Timestamp:  time.Now(),
		}

		if err := db.Create(&message).Error; err != nil {
			log.Printf("Error saving WebSocket message: %v", err)
			conn.WriteJSON(gin.H{"error": "Failed to save message"})
			continue
		}

		// Send to receiver if online
		messageJSON, _ := json.Marshal(message)
		if err := manager.SendPersonalMessage(wsMsg.ReceiverID, messageJSON); err != nil {
			log.Printf("Failed to send message to user %d: %v", wsMsg.ReceiverID, err)
		}

		// Send confirmation to sender
		conn.WriteJSON(gin.H{"status": "sent", "message_id": message.ID})
	}

	log.Printf("User %d disconnected from WebSocket", userID)
}
