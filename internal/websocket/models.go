package websocket

import (
	"time"
)

// MessageType defines the type of message being sent
type MessageType string

const (
	// MessageTypeText represents a regular text message
	MessageTypeText MessageType = "message"
	// MessageTypeSystem represents a system notification
	MessageTypeSystem MessageType = "system"
	// MessageTypeFile represents a file attachment message
	MessageTypeFile MessageType = "file"
)

// Message represents a chat message
type Message struct {
	ID          uint         `json:"id" gorm:"primaryKey"`
	Type        MessageType  `json:"type"`
	Content     string       `json:"content"`
	Sender      string       `json:"sender"`
	SentAt      time.Time    `json:"sent_at"`
	Attachments []Attachment `json:"attachments,omitempty" gorm:"foreignKey:MessageID"`
	Data        interface{}  `json:"data,omitempty" gorm:"-"` // For additional data like user status, not stored in DB
}

// Attachment represents a file attachment in a message
type Attachment struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	MessageID uint      `json:"message_id"`
	FileName  string    `json:"file_name"`
	FileSize  int64     `json:"file_size"`
	FileType  string    `json:"file_type"`
	FilePath  string    `json:"file_path"`
	UploadedAt time.Time `json:"uploaded_at"`
}

// UserStatus represents the online status of a user
type UserStatus struct {
	Username string `json:"username"`
	Online   bool   `json:"online"`
}

// ClientMessage represents a message sent from the client
type ClientMessage struct {
	Type    MessageType `json:"type"`
	Content string      `json:"content"`
	Sender  string      `json:"sender"`
}
