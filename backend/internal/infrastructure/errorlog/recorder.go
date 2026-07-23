package errorlog

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/jetomax/realtime-chat/backend/internal/domain/entity"
	"github.com/jetomax/realtime-chat/backend/internal/infrastructure/persistence/gormmodel"
	"gorm.io/gorm"
)

type Recorder struct {
	db   *gorm.DB
	file *os.File
	mu   sync.Mutex
}

func Open(db *gorm.DB, path string) (*Recorder, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return nil, fmt.Errorf("create error log directory: %w", err)
	}
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o640)
	if err != nil {
		return nil, fmt.Errorf("open error log: %w", err)
	}
	return &Recorder{db: db, file: file}, nil
}

func (r *Recorder) Record(ctx context.Context, event entity.RequestError) error {
	model := gormmodel.RequestError{
		RequestID: event.RequestID, Method: event.Method, Path: event.Path,
		Status: event.Status, ErrorCode: event.Code, Message: event.Message,
		ClientIP: event.ClientIP, UserAgent: event.UserAgent, CreatedAt: event.CreatedAt,
	}
	databaseErr := r.db.WithContext(ctx).Create(&model).Error
	event.ID = model.ID
	encoded, encodingErr := json.Marshal(event)
	if encodingErr != nil {
		return fmt.Errorf("encode request error: %w", encodingErr)
	}
	r.mu.Lock()
	_, fileErr := r.file.Write(append(encoded, '\n'))
	r.mu.Unlock()
	if databaseErr != nil || fileErr != nil {
		return fmt.Errorf("record request error: database=%v file=%v", databaseErr, fileErr)
	}
	return nil
}

func (r *Recorder) Close() error { return r.file.Close() }
