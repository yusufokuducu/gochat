package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"gochat/pkg/handlers"
	"gochat/pkg/websocket"
)

func main() {
	// Gin router'ı oluştur
	router := gin.Default()

	// WebSocket hub'ını başlat
	hub := websocket.NewHub()
	go hub.Run()

	// WebSocket handler'ını oluştur
	wsHandler := handlers.NewWebSocketHandler(hub)

	// Statik dosyalar için klasör ayarı
	router.Static("/static", "./static")
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