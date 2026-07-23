package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jetomax/realtime-chat/backend/internal/domain/entity"
	domainerrors "github.com/jetomax/realtime-chat/backend/internal/domain/errors"
	"github.com/jetomax/realtime-chat/backend/internal/infrastructure/persistence/gormmodel"
	"gorm.io/gorm"
)

type Auth struct{ db *gorm.DB }

func NewAuth(db *gorm.DB) *Auth { return &Auth{db: db} }

func (r *Auth) CreateUserWithRefreshToken(ctx context.Context, user *entity.User, token *entity.RefreshToken) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		model := userToModel(user)
		if err := tx.Create(&model).Error; err != nil {
			return translate(err)
		}
		user.ID, user.CreatedAt, user.UpdatedAt = model.ID, model.CreatedAt, model.UpdatedAt
		token.UserID = model.ID
		tokenModel := refreshToModel(token)
		if err := tx.Create(&tokenModel).Error; err != nil {
			return translate(err)
		}
		token.ID, token.CreatedAt = tokenModel.ID, tokenModel.CreatedAt
		return nil
	})
}

func (r *Auth) FindUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	var model gormmodel.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&model).Error; err != nil {
		return nil, translate(err)
	}
	return userToEntity(model), nil
}

func (r *Auth) FindUserByRefreshTokenHash(ctx context.Context, hash string) (*entity.User, *entity.RefreshToken, error) {
	var token gormmodel.RefreshToken
	if err := r.db.WithContext(ctx).Where("token_hash = ? AND revoked_at IS NULL AND expires_at > ?", hash, time.Now()).First(&token).Error; err != nil {
		return nil, nil, translate(err)
	}
	var user gormmodel.User
	if err := r.db.WithContext(ctx).First(&user, "id = ?", token.UserID).Error; err != nil {
		return nil, nil, translate(err)
	}
	return userToEntity(user), refreshToEntity(token), nil
}

func (r *Auth) CreateRefreshToken(ctx context.Context, token *entity.RefreshToken) error {
	model := refreshToModel(token)
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return translate(err)
	}
	token.ID, token.CreatedAt = model.ID, model.CreatedAt
	return nil
}

func (r *Auth) RotateRefreshToken(ctx context.Context, oldHash string, next *entity.RefreshToken) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		result := tx.Model(&gormmodel.RefreshToken{}).
			Where("token_hash = ? AND revoked_at IS NULL AND expires_at > ?", oldHash, now).
			Update("revoked_at", now)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return domainerrors.ErrUnauthorized
		}
		model := refreshToModel(next)
		return translate(tx.Create(&model).Error)
	})
}

func (r *Auth) RevokeRefreshToken(ctx context.Context, hash string) error {
	now := time.Now()
	result := r.db.WithContext(ctx).Model(&gormmodel.RefreshToken{}).
		Where("token_hash = ? AND revoked_at IS NULL", hash).Update("revoked_at", now)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domainerrors.ErrUnauthorized
	}
	return nil
}

func userToModel(user *entity.User) gormmodel.User {
	return gormmodel.User{ID: user.ID, Email: user.Email, PasswordHash: user.PasswordHash, DisplayName: user.DisplayName}
}

func userToEntity(model gormmodel.User) *entity.User {
	avatar := ""
	if model.AvatarKey != nil {
		avatar = *model.AvatarKey
	}
	return &entity.User{ID: model.ID, Email: model.Email, PasswordHash: model.PasswordHash, DisplayName: model.DisplayName, AvatarKey: avatar, CreatedAt: model.CreatedAt, UpdatedAt: model.UpdatedAt}
}

func refreshToModel(token *entity.RefreshToken) gormmodel.RefreshToken {
	return gormmodel.RefreshToken{ID: token.ID, UserID: token.UserID, TokenHash: token.TokenHash, ExpiresAt: token.ExpiresAt, RevokedAt: token.RevokedAt}
}

func refreshToEntity(model gormmodel.RefreshToken) *entity.RefreshToken {
	return &entity.RefreshToken{ID: model.ID, UserID: model.UserID, TokenHash: model.TokenHash, ExpiresAt: model.ExpiresAt, RevokedAt: model.RevokedAt, CreatedAt: model.CreatedAt}
}

func translate(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domainerrors.ErrNotFound
	}
	if err != nil && (errors.Is(err, gorm.ErrDuplicatedKey) || errors.Is(err, gorm.ErrForeignKeyViolated)) {
		return domainerrors.ErrConflict
	}
	return err
}
