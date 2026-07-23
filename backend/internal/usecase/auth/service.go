package auth

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/jetomax/realtime-chat/backend/internal/domain/entity"
	domainerrors "github.com/jetomax/realtime-chat/backend/internal/domain/errors"
	"github.com/jetomax/realtime-chat/backend/internal/domain/repository"
	"github.com/jetomax/realtime-chat/backend/internal/infrastructure/security"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repository repository.AuthRepository
	hasher     security.PasswordHasher
	tokens     security.TokenManager
	now        func() time.Time
}

type Session struct {
	User             entity.User
	AccessToken      string
	RefreshToken     string
	AccessExpiresAt  time.Time
	RefreshExpiresAt time.Time
}

func NewService(repo repository.AuthRepository, hasher security.PasswordHasher, tokens security.TokenManager) *Service {
	return &Service{repository: repo, hasher: hasher, tokens: tokens, now: time.Now}
}

func (s *Service) Register(ctx context.Context, email, displayName, password string) (*Session, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	displayName = strings.TrimSpace(displayName)
	if err := validate(email, displayName, password); err != nil {
		return nil, err
	}
	hash, err := s.hasher.Hash(password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}
	now := s.now()
	plain, tokenHash, refreshExpiry, err := s.tokens.RefreshToken(now)
	if err != nil {
		return nil, err
	}
	user := entity.User{Email: email, DisplayName: displayName, PasswordHash: hash}
	refresh := entity.RefreshToken{TokenHash: tokenHash, ExpiresAt: refreshExpiry}
	if err := s.repository.CreateUserWithRefreshToken(ctx, &user, &refresh); err != nil {
		return nil, err
	}
	return s.session(user, plain, refreshExpiry, now)
}

func (s *Service) Login(ctx context.Context, email, password string) (*Session, error) {
	user, err := s.repository.FindUserByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
	if err != nil {
		return nil, domainerrors.ErrUnauthorized
	}
	if err := s.hasher.Compare(user.PasswordHash, password); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return nil, domainerrors.ErrUnauthorized
		}
		return nil, err
	}
	now := s.now()
	plain, hash, expiry, err := s.tokens.RefreshToken(now)
	if err != nil {
		return nil, err
	}
	if err := s.repository.CreateRefreshToken(ctx, &entity.RefreshToken{UserID: user.ID, TokenHash: hash, ExpiresAt: expiry}); err != nil {
		return nil, err
	}
	return s.session(*user, plain, expiry, now)
}

func (s *Service) Refresh(ctx context.Context, oldToken string) (*Session, error) {
	if oldToken == "" {
		return nil, domainerrors.ErrUnauthorized
	}
	oldHash := security.HashRefreshToken(oldToken)
	user, _, err := s.repository.FindUserByRefreshTokenHash(ctx, oldHash)
	if err != nil {
		return nil, domainerrors.ErrUnauthorized
	}
	now := s.now()
	plain, hash, expiry, err := s.tokens.RefreshToken(now)
	if err != nil {
		return nil, err
	}
	next := entity.RefreshToken{UserID: user.ID, TokenHash: hash, ExpiresAt: expiry}
	if err := s.repository.RotateRefreshToken(ctx, oldHash, &next); err != nil {
		return nil, domainerrors.ErrUnauthorized
	}
	return s.session(*user, plain, expiry, now)
}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return domainerrors.ErrUnauthorized
	}
	return s.repository.RevokeRefreshToken(ctx, security.HashRefreshToken(refreshToken))
}

func (s *Service) session(user entity.User, refresh string, refreshExpiry, now time.Time) (*Session, error) {
	access, accessExpiry, err := s.tokens.AccessToken(user.ID, now)
	if err != nil {
		return nil, err
	}
	user.PasswordHash = ""
	return &Session{User: user, AccessToken: access, RefreshToken: refresh, AccessExpiresAt: accessExpiry, RefreshExpiresAt: refreshExpiry}, nil
}

func validate(email, name, password string) error {
	if _, err := mail.ParseAddress(email); err != nil || len(email) > 254 {
		return fmt.Errorf("%w: invalid email", domainerrors.ErrValidation)
	}
	if len(name) < 2 || len(name) > 100 {
		return fmt.Errorf("%w: display_name must contain 2-100 characters", domainerrors.ErrValidation)
	}
	if len(password) < 8 || len(password) > 72 {
		return fmt.Errorf("%w: password must contain 8-72 characters", domainerrors.ErrValidation)
	}
	return nil
}
