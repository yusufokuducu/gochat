package websocket

import (
	"time"

	"gorm.io/gorm"
)

// Message type constants
const (
	MessageTypeText     = "message"
	MessageTypeSystem   = "system"
	MessageTypeUserList = "user_list"
	MessageTypeFile     = "file"
	MessageTypeError    = "error"
)

// Message represents a chat message
type Message struct {
	gorm.Model
	Type        string       `json:"type" gorm:"type:varchar(50)"`
	Content     string       `json:"content" gorm:"type:text"`
	Sender      string       `json:"sender" gorm:"type:varchar(100)"`
	SentAt      time.Time    `json:"sent_at" gorm:"index"`
	Attachments []Attachment `json:"attachments,omitempty" gorm:"foreignKey:MessageID"`
}

// Attachment represents a file attachment to a message
type Attachment struct {
	gorm.Model
	MessageID uint   `json:"message_id" gorm:"index"`
	FileName  string `json:"file_name" gorm:"type:varchar(255)"`
	FileURL   string `json:"file_url" gorm:"type:varchar(255)"`
	FileType  string `json:"file_type" gorm:"type:varchar(50)"`
	FileSize  int64  `json:"file_size"`
	UploadedAt time.Time `json:"uploaded_at"`
}

// UserListMessage represents a message containing the list of online users
type UserListMessage struct {
	Type  string   `json:"type"`
	Users []string `json:"users"`
}

// FileMessage represents a message containing a file attachment
type FileMessage struct {
	Type     string    `json:"type"`
	Content  string    `json:"content"`
	Sender   string    `json:"sender"`
	SentAt   time.Time `json:"sent_at"`
	FileName string    `json:"file_name"`
	FileURL  string    `json:"file_url"`
	FileType string    `json:"file_type"`
	FileSize int64     `json:"file_size"`
}

// ErrorMessage represents an error message sent to clients
type ErrorMessage struct {
	Type    string `json:"type"`
	Content string `json:"content"`
	Sender  string `json:"sender"`
}

// TableName sets the table name for the Message model
func (Message) TableName() string {
	return "messages"
}

// TableName sets the table name for the Attachment model
func (Attachment) TableName() string {
	return "attachments"
}
