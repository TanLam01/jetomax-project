package message

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jetomax/realtime-chat/backend/internal/domain/entity"
	domainerrors "github.com/jetomax/realtime-chat/backend/internal/domain/errors"
	"github.com/jetomax/realtime-chat/backend/internal/domain/repository"
)

type Cursor struct {
	CreatedAt time.Time
	MessageID string
}

type Page struct {
	Messages []entity.Message
	HasMore  bool
}

type Service struct{ repository repository.MessageRepository }

func NewService(repository repository.MessageRepository) *Service {
	return &Service{repository: repository}
}

func (s *Service) List(ctx context.Context, userID, conversationID string, cursor *Cursor, limit int) (*Page, error) {
	if _, err := uuid.Parse(conversationID); err != nil {
		return nil, fmt.Errorf("%w: invalid conversation id", domainerrors.ErrValidation)
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		return nil, fmt.Errorf("%w: limit must not exceed 100", domainerrors.ErrValidation)
	}
	var before *time.Time
	beforeID := ""
	if cursor != nil {
		if cursor.CreatedAt.IsZero() {
			return nil, fmt.Errorf("%w: invalid cursor timestamp", domainerrors.ErrValidation)
		}
		if _, err := uuid.Parse(cursor.MessageID); err != nil {
			return nil, fmt.Errorf("%w: invalid cursor message id", domainerrors.ErrValidation)
		}
		before, beforeID = &cursor.CreatedAt, cursor.MessageID
	}
	messages, err := s.repository.ListForMember(ctx, conversationID, userID, before, beforeID, limit+1)
	if err != nil {
		return nil, err
	}
	hasMore := len(messages) > limit
	if hasMore {
		messages = messages[:limit]
	}
	return &Page{Messages: messages, HasMore: hasMore}, nil
}
