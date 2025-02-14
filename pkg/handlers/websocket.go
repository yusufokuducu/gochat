package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	gorilla "github.com/gorilla/websocket"
	"gochat/pkg/websocket"
)

var upgrader = gorilla.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	HandshakeTimeout: 10 * time.Second,
}

type WebSocketHandler struct {
	hub *websocket.Hub
}

func NewWebSocketHandler(hub *websocket.Hub) *WebSocketHandler {
	return &WebSocketHandler{
		hub: hub,
	}
}

func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	// WebSocket bağlantısını yükselt
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket yükseltme hatası: %v", err)
		return
	}

	// Bağlantı ayarlarını yapılandır
	conn.SetReadLimit(32768) // 32KB
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

	// Kullanıcı adını query parametresinden al
	username := c.Query("username")
	if username == "" {
		username = "Anonim"
	}

	// Yeni istemci oluştur
	client := websocket.NewClient(conn, h.hub, username)

	// İstemciyi hub'a kaydet
	h.hub.Register <- client

	// Goroutine'leri başlat
	go client.WritePump()
	go client.ReadPump()
} 