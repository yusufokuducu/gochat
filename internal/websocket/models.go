package websocket

import (
	"time"
	
	"gorm.io/gorm"
)

type Message struct {
	gorm.Model
	Content     string     `gorm:"type:text" json:"content"`
	Sender      string     `gorm:"type:varchar(100)" json:"sender"`
	RoomID      string     `gorm:"type:varchar(100);index" json:"room_id"`
	Attachments []Attachment `gorm:"foreignKey:MessageID" json:"attachments"`
	Timestamp   time.Time  `gorm:"index" json:"timestamp"`
}

type Attachment struct {
	gorm.Model
	MessageID   uint   `gorm:"index" json:"message_id"`
	FileName    string `gorm:"type:varchar(255)" json:"file_name"`
	FileType    string `gorm:"type:varchar(100)" json:"file_type"`
	FileSize    int64  `json:"file_size"`
	StoragePath string `gorm:"type:text" json:"storage_path"`
}
