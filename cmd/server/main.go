package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"gochat/pkg/database"
	"gochat/pkg/handlers"
	"gochat/pkg/websocket"
)

func main() {
	// Gerekli dizinleri oluştur
	for _, dir := range []string{"data", "uploads"} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatal("Failed to create directory:", dir, err)
		}
	}

	// Veritabanı bağlantısını oluştur
	db, err := database.NewDatabase(filepath.Join("data", "chat.json"))
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	// Gin router'ı oluştur
	router := gin.Default()

	// WebSocket hub'ını başlat
	hub := websocket.NewHub(db)
	go hub.Run()

	// WebSocket handler'ını oluştur
	wsHandler := handlers.NewWebSocketHandler(hub)

	// Statik dosyalar için klasör ayarları
	router.Static("/static", "./static")
	router.Static("/uploads", "./uploads")
	router.LoadHTMLGlob("static/*.html")

	// Ana sayfa
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	// WebSocket endpoint'i
	router.GET("/ws", wsHandler.HandleWebSocket)

	// Sunucuyu başlat
	log.Println("Server starting at :8080...")
	if err := router.Run(":8080"); err != nil {
		log.Fatal("Server failed to start:", err)
	}
} 