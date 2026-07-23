package repository

import (
	"context"
	"strings"

	"github.com/jetomax/realtime-chat/backend/internal/domain/entity"
	"github.com/jetomax/realtime-chat/backend/internal/infrastructure/persistence/gormmodel"
	"gorm.io/gorm"
)

type User struct{ db *gorm.DB }

func NewUser(db *gorm.DB) *User { return &User{db: db} }

func (r *User) FindByID(ctx context.Context, id string) (*entity.User, error) {
	var model gormmodel.User
	if err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error; err != nil {
		return nil, translate(err)
	}
	return userToEntity(model), nil
}

func (r *User) Search(ctx context.Context, query, excludeUserID string, limit int) ([]entity.User, error) {
	pattern := "%" + strings.ToLower(query) + "%"
	var models []gormmodel.User
	err := r.db.WithContext(ctx).
		Where("id <> ? AND (LOWER(display_name) LIKE ? OR LOWER(email) LIKE ?)", excludeUserID, pattern, pattern).
		Order("display_name ASC, id ASC").Limit(limit).Find(&models).Error
	if err != nil {
		return nil, err
	}
	users := make([]entity.User, 0, len(models))
	for _, model := range models {
		users = append(users, *userToEntity(model))
	}
	return users, nil
}
