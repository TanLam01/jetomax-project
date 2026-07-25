package repository

import (
	"context"

	"github.com/jetomax/realtime-chat/backend/internal/domain/entity"
	"github.com/jetomax/realtime-chat/backend/internal/infrastructure/persistence/gormmodel"
	"gorm.io/gorm"
)

type MediaUpload struct{ db *gorm.DB }

func NewMediaUpload(db *gorm.DB) *MediaUpload { return &MediaUpload{db: db} }

func (r *MediaUpload) Create(ctx context.Context, upload *entity.MediaUpload) error {
	model := gormmodel.MediaUpload{ID: upload.ID, UserID: upload.UserID, ObjectKey: upload.ObjectKey,
		OriginalName: upload.OriginalName, MIMEType: upload.MIMEType, Size: upload.Size,
		Status: upload.Status, ExpiresAt: upload.ExpiresAt}
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return translate(err)
	}
	upload.CreatedAt = model.CreatedAt
	return nil
}
