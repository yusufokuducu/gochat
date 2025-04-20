package main

import (
	"fmt"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"time"

	"github.com/faust-lvii/gochat/backend/internal/database"
	"github.com/faust-lvii/gochat/backend/internal/handlers"
	"github.com/faust-lvii/gochat/backend/internal/middleware"
	"github.com/faust-lvii/gochat/backend/internal/utils"
	"github.com/faust-lvii/gochat/backend/internal/websocket"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using default values")
	}

	// Initialize JWT secret
	jwtSecret := []byte(utils.GetEnv("SECRET_KEY", "your-secret-key-for-development-only"))

	// Initialize database
	dsn := utils.GetEnv("DATABASE_URL", "postgresql://postgres:postgres@db:5432/gochat")
	err := database.Initialize(dsn)
	if err != nil {
		log.Println("Warning: Failed to connect to database:", err)
		log.Println("Running in memory mode without database")
	}

	// Initialize handlers
	db := database.GetDB()
	authHandler := handlers.NewAuthHandler(db, jwtSecret)
	userHandler := handlers.NewUserHandler(db)
	friendshipHandler := handlers.NewFriendshipHandler(db)
	messageHandler := handlers.NewMessageHandler(db)
	wsHandler := websocket.NewHandler(db, jwtSecret)

	// Initialize middleware
	authMiddleware := middleware.NewAuth(jwtSecret)

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
		c.JSON(200, gin.H{
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
			auth.POST("/login", authHandler.Login)
			auth.POST("/register", authHandler.Register)
		}

		// Protected endpoints
		protected := api.Group("")
		protected.Use(authMiddleware.Middleware())
		{
			// User endpoints
			protected.GET("/users", userHandler.ListUsers)
			protected.GET("/users/me", userHandler.GetCurrentUser)
			protected.GET("/users/:id", userHandler.GetUser)
			protected.PUT("/users/:id", userHandler.UpdateUser)

			// Friendship endpoints
			friendships := protected.Group("/friendships")
			{
				friendships.GET("", friendshipHandler.ListFriendships)
				friendships.POST("", friendshipHandler.CreateFriendship)
				friendships.PUT("/:id", friendshipHandler.UpdateFriendship)
				friendships.DELETE("/:id", friendshipHandler.DeleteFriendship)
			}

			// Message endpoints
			messages := protected.Group("/messages")
			{
				messages.GET("", messageHandler.ListMessages)
				messages.POST("", messageHandler.CreateMessage)
				messages.PUT("/:id/read", messageHandler.MarkMessageAsRead)
			}
		}
	}

	// WebSocket endpoint
	router.GET("/api/ws", wsHandler.HandleConnection)

	// Start server
	port := utils.GetEnvAsInt("PORT", 8000)
	fmt.Printf("Server running on port %d...\n", port)
	if err := router.Run(fmt.Sprintf(":%d", port)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
