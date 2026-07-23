package conversation

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/jetomax/realtime-chat/backend/internal/domain/entity"
	domainerrors "github.com/jetomax/realtime-chat/backend/internal/domain/errors"
	"github.com/jetomax/realtime-chat/backend/internal/domain/repository"
)

type Service struct {
	repository repository.ConversationRepository
}

func (s *Service) CreateDirect(ctx context.Context, creatorID, targetID string) (*entity.ConversationSummary, bool, error) {
	if _, err := uuid.Parse(targetID); err != nil {
		return nil, false, fmt.Errorf("%w: invalid user_id", domainerrors.ErrValidation)
	}
	if creatorID == targetID {
		return nil, false, fmt.Errorf("%w: cannot create a direct conversation with yourself", domainerrors.ErrValidation)
	}
	ids := []string{creatorID, targetID}
	sort.Strings(ids)
	return s.repository.CreateDirect(ctx, creatorID, targetID, ids[0]+":"+ids[1])
}

func (s *Service) CreateGroup(ctx context.Context, creatorID, name, avatarKey string, memberIDs []string) (*entity.ConversationSummary, error) {
	name, avatarKey = strings.TrimSpace(name), strings.TrimSpace(avatarKey)
	if len([]rune(name)) < 2 || len([]rune(name)) > 100 {
		return nil, fmt.Errorf("%w: name must contain 2-100 characters", domainerrors.ErrValidation)
	}
	if len(memberIDs) == 0 {
		return nil, fmt.Errorf("%w: at least one member is required", domainerrors.ErrValidation)
	}
	if len(memberIDs) > 99 {
		return nil, fmt.Errorf("%w: no more than 99 invited members are allowed", domainerrors.ErrValidation)
	}
	unique := make([]string, 0, len(memberIDs))
	seen := map[string]struct{}{creatorID: {}}
	for _, id := range memberIDs {
		if _, err := uuid.Parse(id); err != nil {
			return nil, fmt.Errorf("%w: invalid member_id", domainerrors.ErrValidation)
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		unique = append(unique, id)
	}
	if len(unique) == 0 {
		return nil, fmt.Errorf("%w: at least one other member is required", domainerrors.ErrValidation)
	}
	return s.repository.CreateGroup(ctx, creatorID, entity.CreateGroupInput{Name: name, AvatarKey: avatarKey, MemberIDs: unique})
}

func (s *Service) AddMembers(ctx context.Context, actorID, conversationID string, userIDs []string) ([]string, error) {
	if _, err := uuid.Parse(conversationID); err != nil {
		return nil, fmt.Errorf("%w: invalid conversation id", domainerrors.ErrValidation)
	}
	if len(userIDs) == 0 {
		return nil, fmt.Errorf("%w: at least one user_id is required", domainerrors.ErrValidation)
	}
	if len(userIDs) > 100 {
		return nil, fmt.Errorf("%w: no more than 100 users can be added at once", domainerrors.ErrValidation)
	}
	unique := make([]string, 0, len(userIDs))
	seen := make(map[string]struct{}, len(userIDs))
	for _, id := range userIDs {
		if _, err := uuid.Parse(id); err != nil {
			return nil, fmt.Errorf("%w: invalid user_id", domainerrors.ErrValidation)
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		unique = append(unique, id)
	}
	return s.repository.AddMembers(ctx, conversationID, actorID, unique)
}

func NewService(repository repository.ConversationRepository) *Service {
	return &Service{repository: repository}
}

func (s *Service) List(ctx context.Context, userID string) ([]entity.ConversationSummary, error) {
	return s.repository.ListForUser(ctx, userID, 50)
}
