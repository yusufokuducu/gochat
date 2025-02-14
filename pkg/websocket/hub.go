package websocket

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"gochat/pkg/database"
)

type Hub struct {
	clients    map[*Client]bool
	Broadcast  chan []byte
	Register   chan *Client
	Unregister chan *Client
	mu         sync.RWMutex
	db         *database.Database
}

func NewHub(db *database.Database) *Hub {
	return &Hub{
		Broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		db:         db,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()

			// Yeni kullanıcıya son mesajları gönder
			messages, err := h.db.GetRecentMessages(50)
			if err != nil {
				log.Printf("Error getting recent messages: %v", err)
			} else {
				// Sort messages in ascending order by timestamp
				sort.Slice(messages, func(i, j int) bool {
					return messages[i].Timestamp.Before(messages[j].Timestamp)
				})
				for _, msg := range messages {
					var messageData Message
					if msg.Type == "file" {
						messageData = Message{
							Type:      "file",
							Sender:    msg.Sender,
							Timestamp: msg.Timestamp,
							FileInfo: &FileInfo{
								FileURL: msg.FileURL,
							},
						}
					} else {
						messageData = Message{
							Type:      msg.Type,
							Content:   msg.Content,
							Sender:    msg.Sender,
							Timestamp: msg.Timestamp,
						}
					}

					messageJSON, err := json.Marshal(messageData)
					if err != nil {
						log.Printf("Error marshaling message: %v", err)
						continue
					}

					select {
					case client.Send <- messageJSON:
					default:
						log.Printf("Failed to send message to client %s", client.ID)
					}
				}
			}

			h.broadcastUserList()
			h.broadcastSystemMessage(client.Name + " katıldı")

		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
			}
			h.mu.Unlock()
			h.broadcastUserList()
			h.broadcastSystemMessage(client.Name + " ayrıldı")

		case message := <-h.Broadcast:
			var msg Message
			if err := json.Unmarshal(message, &msg); err != nil {
				log.Printf("Error unmarshaling message: %v", err)
				continue
			}

			switch msg.Type {
			case "file":
				if err := h.handleFileMessage(&msg); err != nil {
					log.Printf("Error handling file message: %v", err)
					continue
				}
				// Yeni mesaj JSON'ını oluştur
				var err error
				message, err = json.Marshal(msg)
				if err != nil {
					log.Printf("Error marshaling file message: %v", err)
					continue
				}
			case "emoji":
				log.Printf("Received emoji message: %s", msg.Content)
				newJSON, err := json.Marshal(msg)
				if err != nil {
					log.Printf("Error marshaling emoji message: %v", err)
					continue
				}
				message = newJSON
			default:
				// Normal text message, no extra processing
			}

			// Mesajı veritabanına kaydet
			dbMsg := &database.ChatMessage{
				Content:   msg.Content,
				Sender:    msg.Sender,
				Timestamp: msg.Timestamp,
				Type:      msg.Type,
			}

			if msg.Type == "file" && msg.FileInfo != nil {
				dbMsg.FileURL = msg.FileInfo.FileURL
				dbMsg.Content = fmt.Sprintf("Dosya paylaşıldı: %s (%s)", msg.FileInfo.FileName, msg.FileInfo.FileType)
			}

			if err := h.db.SaveMessage(dbMsg); err != nil {
				log.Printf("Error saving message: %v", err)
			}

			// Mesajı tüm istemcilere gönder
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) broadcastUserList() {
	h.mu.RLock()
	users := make([]map[string]interface{}, 0)
	for client := range h.clients {
		users = append(users, map[string]interface{}{
			"id":     client.ID,
			"name":   client.Name,
			"online": client.IsOnline(),
		})
	}
	h.mu.RUnlock()

	message := Message{
		Type: "user_list",
		Content: func() string {
			data, err := json.Marshal(map[string]interface{}{
				"users": users,
			})
			if err != nil {
				log.Printf("Error marshaling user list: %v", err)
				return ""
			}
			return string(data)
		}(),
		Timestamp: getNow(),
	}

	messageJSON, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	h.Broadcast <- messageJSON
}

func (h *Hub) broadcastSystemMessage(content string) {
	message := Message{
		Type:      "system",
		Content:   content,
		Timestamp: getNow(),
	}

	// Sistem mesajını veritabanına kaydet
	dbMsg := &database.ChatMessage{
		Content:   content,
		Sender:    "System",
		Timestamp: message.Timestamp,
		Type:      "system",
	}
	if err := h.db.SaveMessage(dbMsg); err != nil {
		log.Printf("Error saving system message: %v", err)
	}

	messageJSON, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling system message: %v", err)
		return
	}

	h.Broadcast <- messageJSON
}

func (h *Hub) handleFileMessage(msg *Message) error {
	if msg.FileInfo == nil {
		return fmt.Errorf("missing file info in file message")
	}
	// Base64'ten dosyayı decode et
	data, err := base64.StdEncoding.DecodeString(msg.Content)
	if err != nil {
		return fmt.Errorf("base64 decode error: %v", err)
	}

	// Dosya adını oluştur
	fileName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), msg.FileInfo.FileName)
	filePath := filepath.Join("uploads", fileName)

	// uploads klasörünü oluştur
	if err := os.MkdirAll("uploads", 0755); err != nil {
		return fmt.Errorf("failed to create uploads directory: %v", err)
	}

	// Dosyayı kaydet
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to save file: %v", err)
	}

	// FileURL'i güncelle
	msg.FileInfo.FileURL = "/uploads/" + fileName
	msg.Content = "" // Base64 içeriğini temizle

	return nil
}

func getNow() time.Time {
	return time.Now()
} 