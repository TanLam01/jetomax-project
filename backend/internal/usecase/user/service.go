package user

import (
	"context"
	"fmt"
	"strings"

	"github.com/jetomax/realtime-chat/backend/internal/domain/entity"
	domainerrors "github.com/jetomax/realtime-chat/backend/internal/domain/errors"
	"github.com/jetomax/realtime-chat/backend/internal/domain/repository"
)

type Service struct{ repository repository.UserRepository }

func NewService(repository repository.UserRepository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Me(ctx context.Context, userID string) (*entity.User, error) {
	return s.repository.FindByID(ctx, userID)
}

func (s *Service) Search(ctx context.Context, userID, query string) ([]entity.User, error) {
	query = strings.TrimSpace(query)
	if len([]rune(query)) < 2 || len([]rune(query)) > 100 {
		return nil, fmt.Errorf("%w: query must contain 2-100 characters", domainerrors.ErrValidation)
	}
	return s.repository.Search(ctx, query, userID, 20)
}
