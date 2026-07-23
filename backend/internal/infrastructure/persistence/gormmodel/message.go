package gormmodel

import "time"

type Message struct {
	ID              string `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	ConversationID  string `gorm:"type:uuid;not null;index"`
	SenderID        string `gorm:"type:uuid;not null"`
	Type            string `gorm:"type:message_type;not null"`
	Text            *string
	ClientMessageID string `gorm:"not null"`
	CreatedAt       time.Time
}

func (Message) TableName() string { return "messages" }
