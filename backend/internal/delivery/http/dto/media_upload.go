package dto

import (
	"time"

	mediausecase "github.com/jetomax/realtime-chat/backend/internal/usecase/media"
)

type CreateMediaUploadRequest struct {
	FileName string `json:"file_name" binding:"required"`
	MIMEType string `json:"mime_type" binding:"required"`
	Size     int64  `json:"size" binding:"required"`
}

type CreateMediaUploadResponse struct {
	UploadID  string            `json:"upload_id"`
	ObjectKey string            `json:"object_key"`
	UploadURL string            `json:"upload_url"`
	Method    string            `json:"method"`
	Headers   map[string]string `json:"headers"`
	ExpiresAt string            `json:"expires_at"`
}

func NewCreateMediaUploadResponse(upload *mediausecase.Upload) CreateMediaUploadResponse {
	return CreateMediaUploadResponse{UploadID: upload.ID, ObjectKey: upload.ObjectKey,
		UploadURL: upload.UploadURL, Method: "PUT", Headers: upload.Headers,
		ExpiresAt: upload.ExpiresAt.Format(time.RFC3339)}
}
