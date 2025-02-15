package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kyusu/gochat/internal/database"
	"github.com/kyusu/gochat/internal/websocket"
)

func main() {
	// Gerekli dizinleri oluştur
	for _, dir := range []string{"data", "uploads"} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatal("Failed to create directory:", dir, err)
		}
	}

	// Database configuration
	dbConfig := database.Config{
		Host:     "localhost",
		Port:     "5432", 
		User:     "postgres",
		Password: "your_db_password",
		DBName:   "gochat",
		SSLMode:  "disable",
	}

	// Initialize database with retry logic
	var db *gorm.DB
	var err error
	
	for retries := 0; retries < 5; retries++ {
		db, err = database.NewConnection(dbConfig)
		if err == nil {
			break
		}
		log.Printf("Database connection attempt %d failed: %v", retries+1, err)
		time.Sleep(2 * time.Second)
	}
	
	if err != nil {
		log.Fatal("Failed to connect to database after 5 attempts: ", err)
	}

	// Perform health check
	if err := database.HealthCheck(db); err != nil {
		log.Fatal("Database health check failed: ", err)
	}

	// Auto-migrate models
	err = db.AutoMigrate(
		&websocket.Message{},
		&websocket.Attachment{},
	)
	if err != nil {
		log.Fatal("Failed to auto-migrate database models: ", err)
	}

	// Gin router'ı oluştur
	router := gin.Default()

	// WebSocket hub'ını başlat
	hub := websocket.NewHub(db)
	go hub.Run()

	// WebSocket handler'ını oluştur
	wsHandler := websocket.ServeWs(hub)

	// Statik dosyalar için klasör ayarları
	router.Static("/static", "./static")
	router.Static("/uploads", "./uploads")
	router.LoadHTMLGlob("static/*.html")

	// Ana sayfa
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	// WebSocket endpoint'i
	router.GET("/ws", wsHandler)

	// Sunucuyu başlat
	log.Println("Server starting at :8080...")
	if err := router.Run(":8080"); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}