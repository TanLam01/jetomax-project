package gormmodel

import "time"

type MediaUpload struct {
	ID           string `gorm:"type:uuid;primaryKey"`
	UserID       string `gorm:"type:uuid;not null;index"`
	ObjectKey    string `gorm:"not null;uniqueIndex"`
	OriginalName string `gorm:"not null"`
	MIMEType     string `gorm:"not null"`
	Size         int64  `gorm:"not null"`
	Status       string `gorm:"not null"`
	ExpiresAt    time.Time
	CreatedAt    time.Time
}

func (MediaUpload) TableName() string { return "media_uploads" }
