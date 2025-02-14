package database

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

type Database struct {
	messages []ChatMessage
	filePath string
	mu       sync.RWMutex
}

type ChatMessage struct {
	ID        int64     `json:"id"`
	Content   string    `json:"content"`
	Sender    string    `json:"sender"`
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"`      // "message", "file" veya "system"
	FileURL   string    `json:"file_url"`  // Dosya mesajları için
}

func NewDatabase(dbPath string) (*Database, error) {
	if dbPath == "" {
		dbPath = "data/chat.json"
	}

	// Dizini oluştur
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %v", err)
	}

	db := &Database{
		messages: make([]ChatMessage, 0),
		filePath: dbPath,
	}

	// Dosya varsa, mesajları yükle
	if _, err := os.Stat(dbPath); err == nil {
		data, err := os.ReadFile(dbPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read database file: %v", err)
		}

		if len(data) > 0 {
			if err := json.Unmarshal(data, &db.messages); err != nil {
				return nil, fmt.Errorf("failed to unmarshal messages: %v", err)
			}
		}
	}

	return db, nil
}

func (d *Database) SaveMessage(msg *ChatMessage) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	msg.ID = time.Now().UnixNano()
	d.messages = append(d.messages, *msg)

	// Mesaj sayısı 1000'i geçerse en eski mesajları sil
	if len(d.messages) > 1000 {
		d.messages = d.messages[len(d.messages)-1000:]
	}

	// Mesajları dosyaya kaydet
	data, err := json.Marshal(d.messages)
	if err != nil {
		return fmt.Errorf("failed to marshal messages: %v", err)
	}

	if err := os.WriteFile(d.filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write database file: %v", err)
	}

	return nil
}

func (d *Database) GetRecentMessages(limit int) ([]ChatMessage, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if limit <= 0 {
		limit = 50
	}

	// Mesajları tarihe göre sırala (en yeni en sonda)
	messages := make([]ChatMessage, len(d.messages))
	copy(messages, d.messages)

	// Tarihe göre sırala
	sort.Slice(messages, func(i, j int) bool {
		return messages[i].Timestamp.Before(messages[j].Timestamp)
	})

	// Son mesajları al
	start := 0
	if len(messages) > limit {
		start = len(messages) - limit
	}

	result := make([]ChatMessage, len(messages[start:]))
	copy(result, messages[start:])

	return result, nil
}

func (d *Database) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Son kez mesajları kaydet
	data, err := json.Marshal(d.messages)
	if err != nil {
		return fmt.Errorf("failed to marshal messages: %v", err)
	}

	if err := os.WriteFile(d.filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write database file: %v", err)
	}

	return nil
} 