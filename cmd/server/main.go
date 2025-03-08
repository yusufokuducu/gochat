package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gochat/internal/handlers"
	"gochat/internal/websocket"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Configure database connection
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "host=localhost user=postgres password=1122 dbname=gochat port=5432 sslmode=disable TimeZone=Europe/Istanbul"
	}

	// Configure GORM logger
	gormLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	// Connect to database
	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto migrate database schema
	err = db.AutoMigrate(&websocket.Message{}, &websocket.Attachment{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	log.Println("Database connection established and migrations completed")

	// Create required directories
	createRequiredDirectories()

	// Create a new hub
	hub := websocket.NewHub(db)
	go hub.Run()

	// Create a new WebSocket handler
	wsHandler := handlers.NewWebSocketHandler(hub)

	// Set up Gin router
	router := gin.Default()

	// Serve static files
	router.Static("/css", "./static/css")
	router.Static("/js", "./static/js")
	router.Static("/uploads", "./uploads")
	router.StaticFile("/", "./static/index.html")

	// WebSocket endpoint
	router.GET("/ws", func(c *gin.Context) {
		wsHandler.ServeWs(c)
	})

	// File upload endpoint
	router.POST("/upload", func(c *gin.Context) {
		handleFileUpload(c, db)
	})

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on http://localhost:%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// createRequiredDirectories creates necessary directories for the application
func createRequiredDirectories() {
	directories := []string{
		"./uploads",
		"./static/js",
		"./static/css",
	}
	
	for _, dir := range directories {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}
}

// handleFileUpload processes file uploads and saves them to the server
func handleFileUpload(c *gin.Context, db *gorm.DB) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "No file provided"})
		return
	}

	// Generate unique filename
	filename := time.Now().Format("20060102150405") + "_" + file.Filename

	// Save file to uploads directory
	if err := c.SaveUploadedFile(file, "./uploads/"+filename); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to save file"})
		return
	}

	// Create attachment record
	attachment := websocket.Attachment{
		FileName: filename,
	}

	if err := db.Create(&attachment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": "Failed to save attachment to database",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"filename":      filename,
		"attachment_id": attachment.ID,
	})
}
