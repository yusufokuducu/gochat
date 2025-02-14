package websocket

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	ID       string
	Name     string
	Conn     *websocket.Conn
	Hub      *Hub
	Send     chan []byte
	mu       sync.Mutex
	isOnline bool
}

type Message struct {
	Type      string    `json:"type"`
	Content   string    `json:"content"`
	Sender    string    `json:"sender"`
	Timestamp time.Time `json:"timestamp"`
	FileInfo  *FileInfo `json:"file_info,omitempty"`
}

type FileInfo struct {
	FileName string `json:"fileName"`
	FileSize int64  `json:"fileSize"`
	FileType string `json:"fileType"`
	FileURL  string `json:"fileURL"`
}

func NewClient(conn *websocket.Conn, hub *Hub, name string) *Client {
	return &Client{
		ID:       generateID(),
		Name:     name,
		Conn:     conn,
		Hub:      hub,
		Send:     make(chan []byte, 256),
		isOnline: true,
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("error unmarshaling message: %v", err)
			continue
		}

		// Yazıyor durumu mesajlarını özel olarak işle
		if msg.Type == "typing" {
			// Yazıyor durumunu diğer kullanıcılara ilet
			typingMsg := Message{
				Type:      "typing",
				Sender:    c.Name,
				Content:   msg.Content,
				Timestamp: time.Now(),
			}
			
			typingJSON, err := json.Marshal(typingMsg)
			if err != nil {
				log.Printf("error marshaling typing message: %v", err)
				continue
			}
			
			c.Hub.Broadcast <- typingJSON
			continue
		}

		msg.Timestamp = time.Now()
		msg.Sender = c.Name

		// Dosya mesajlarını işle
		if msg.Type == "file" && msg.FileInfo != nil {
			if err := c.handleFileMessage(&msg); err != nil {
				log.Printf("error handling file message: %v", err)
				continue
			}
		}

		messageJSON, err := json.Marshal(msg)
		if err != nil {
			log.Printf("error marshaling message: %v", err)
			continue
		}

		c.Hub.Broadcast <- messageJSON
	}
}

func (c *Client) handleFileMessage(msg *Message) error {
	// Base64 kodlu dosya içeriğini çöz
	fileData, err := base64.StdEncoding.DecodeString(msg.Content)
	if err != nil {
		log.Printf("error decoding file data: %v", err)
		return err
	}

	// Güvenli dosya adı oluştur
	safeFileName := generateSafeFileName(msg.FileInfo.FileName)
	fileName := filepath.Join("uploads", safeFileName)

	// Dosyayı kaydet
	err = os.WriteFile(fileName, fileData, 0644)
	if err != nil {
		log.Printf("error saving file: %v", err)
		return err
	}

	// Dosya URL'sini güncelle
	msg.FileInfo.FileURL = "/uploads/" + safeFileName
	msg.Content = "" // Base64 içeriğini temizle
	return nil
}

func generateSafeFileName(originalName string) string {
	// Dosya uzantısını al
	ext := filepath.Ext(originalName)
	// Dosya adının ana kısmını al
	baseName := strings.TrimSuffix(originalName, ext)
	// Güvenli bir dosya adı oluştur
	safeBaseName := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		return '-'
	}, baseName)
	// Benzersiz bir isim oluştur
	timestamp := time.Now().Format("20060102150405")
	return fmt.Sprintf("%s_%s%s", safeBaseName, timestamp, ext)
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) SetOnline(status bool) {
	c.mu.Lock()
	c.isOnline = status
	c.mu.Unlock()
}

func (c *Client) IsOnline() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.isOnline
}

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 10 * 1024 * 1024 // 10MB
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

func generateID() string {
	return time.Now().Format("20060102150405") + "_" + randomString(6)
}

func randomString(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[time.Now().UnixNano()%int64(len(letterBytes))]
	}
	return string(b)
} 