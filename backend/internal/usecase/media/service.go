package media

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jetomax/realtime-chat/backend/internal/domain/entity"
	domainerrors "github.com/jetomax/realtime-chat/backend/internal/domain/errors"
	"github.com/jetomax/realtime-chat/backend/internal/domain/repository"
)

const maxImageSize int64 = 10 * 1024 * 1024

var imageExtensions = map[string]string{"image/jpeg": ".jpg", "image/png": ".png", "image/webp": ".webp"}

type Service struct {
	repository repository.MediaUploadRepository
	signer     repository.UploadSigner
	now        func() time.Time
}

type Upload struct {
	entity.MediaUpload
	UploadURL string
	Headers   map[string]string
}

func NewService(repository repository.MediaUploadRepository, signer repository.UploadSigner) *Service {
	return &Service{repository: repository, signer: signer, now: time.Now}
}

func (s *Service) CreateUpload(ctx context.Context, userID, fileName, mimeType string, size int64) (*Upload, error) {
	fileName, mimeType = filepath.Base(strings.TrimSpace(fileName)), strings.ToLower(strings.TrimSpace(mimeType))
	extension, allowed := imageExtensions[mimeType]
	if !allowed {
		return nil, fmt.Errorf("%w: unsupported image MIME type", domainerrors.ErrValidation)
	}
	if fileName == "." || fileName == "" || len(fileName) > 255 {
		return nil, fmt.Errorf("%w: invalid file name", domainerrors.ErrValidation)
	}
	if size <= 0 || size > maxImageSize {
		return nil, fmt.Errorf("%w: image size must be between 1 byte and 10 MB", domainerrors.ErrValidation)
	}
	now, ttl := s.now().UTC(), 15*time.Minute
	id := uuid.NewString()
	objectKey := fmt.Sprintf("users/%s/images/%s%s", userID, id, extension)
	url, headers, err := s.signer.PresignPut(ctx, objectKey, mimeType, size, ttl)
	if err != nil {
		return nil, err
	}
	upload := entity.MediaUpload{ID: id, UserID: userID, ObjectKey: objectKey, OriginalName: fileName,
		MIMEType: mimeType, Size: size, Status: "pending", ExpiresAt: now.Add(ttl)}
	if err := s.repository.Create(ctx, &upload); err != nil {
		return nil, err
	}
	return &Upload{MediaUpload: upload, UploadURL: url, Headers: headers}, nil
}
