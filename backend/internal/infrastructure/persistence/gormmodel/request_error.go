package gormmodel

import "time"

type RequestError struct {
	ID        string `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	RequestID string `gorm:"not null;index"`
	Method    string `gorm:"not null"`
	Path      string `gorm:"not null"`
	Status    int    `gorm:"not null;index"`
	ErrorCode string `gorm:"not null"`
	Message   string `gorm:"not null"`
	ClientIP  string
	UserAgent string
	CreatedAt time.Time `gorm:"not null"`
}

func (RequestError) TableName() string { return "errors" }
