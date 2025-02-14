package main

import (
	"log"
	"net/http"
	"github.com/gorilla/websocket"
	"github.com/gin-gonic/gin"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Geliştirme aşamasında tüm originlere izin veriyoruz
	},
}

func main() {
	router := gin.Default()

	// Static dosyalar için klasör ayarı
	router.Static("/static", "./static")
	router.LoadHTMLGlob("static/*.html")

	// Ana sayfa
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	// WebSocket endpoint'i
	router.GET("/ws", handleWebSocket)

	log.Println("Server starting at :8080...")
	if err := router.Run(":8080"); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

func handleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}
	defer conn.Close()

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}

		// Echo the message back for now
		if err := conn.WriteMessage(messageType, message); err != nil {
			log.Printf("Error writing message: %v", err)
			break
		}
	}
} 