package bootstrap

import (
	"context"
	"fmt"
	"net/http"

	"github.com/jetomax/realtime-chat/backend/internal/config"
	httpdelivery "github.com/jetomax/realtime-chat/backend/internal/delivery/http"
	wsdelivery "github.com/jetomax/realtime-chat/backend/internal/delivery/websocket"
	"github.com/jetomax/realtime-chat/backend/internal/infrastructure/cache"
	"github.com/jetomax/realtime-chat/backend/internal/infrastructure/errorlog"
	"github.com/jetomax/realtime-chat/backend/internal/infrastructure/persistence"
	persistencerepository "github.com/jetomax/realtime-chat/backend/internal/infrastructure/persistence/repository"
	"github.com/jetomax/realtime-chat/backend/internal/infrastructure/security"
	"github.com/jetomax/realtime-chat/backend/internal/infrastructure/storage"
	authusecase "github.com/jetomax/realtime-chat/backend/internal/usecase/auth"
	conversationusecase "github.com/jetomax/realtime-chat/backend/internal/usecase/conversation"
	mediausecase "github.com/jetomax/realtime-chat/backend/internal/usecase/media"
	messageusecase "github.com/jetomax/realtime-chat/backend/internal/usecase/message"
	userusecase "github.com/jetomax/realtime-chat/backend/internal/usecase/user"
)

type Resources struct {
	Database *persistence.Database
	Redis    *cache.Redis
	Errors   *errorlog.Recorder
	Storage  *storage.S3
	Messages *cache.MessageBus
	Realtime *wsdelivery.Hub
}

func Connect(ctx context.Context, cfg config.Config) (*Resources, error) {
	database, err := persistence.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}
	redisClient, err := cache.Open(ctx, cfg.RedisURL)
	if err != nil {
		_ = database.Close()
		return nil, err
	}
	errorRecorder, err := errorlog.Open(database.ORM, cfg.ErrorLogPath)
	if err != nil {
		_ = redisClient.Close()
		_ = database.Close()
		return nil, err
	}
	objectStorage, err := storage.NewS3(ctx, cfg.S3Endpoint, cfg.S3Region, cfg.S3Bucket, cfg.S3AccessKey, cfg.S3SecretKey, cfg.S3UsePathStyle)
	if err != nil {
		_ = errorRecorder.Close()
		_ = redisClient.Close()
		_ = database.Close()
		return nil, err
	}
	messageBus := cache.NewMessageBus(redisClient.Client)
	realtimeHub := wsdelivery.NewHub()
	if err := messageBus.Start(ctx, realtimeHub.BroadcastMessage, realtimeHub.BroadcastRealtime); err != nil {
		_ = errorRecorder.Close()
		_ = redisClient.Close()
		_ = database.Close()
		return nil, fmt.Errorf("start realtime message subscriber: %w", err)
	}
	return &Resources{Database: database, Redis: redisClient, Errors: errorRecorder, Storage: objectStorage, Messages: messageBus, Realtime: realtimeHub}, nil
}

func (r *Resources) Close() error {
	r.Realtime.Close()
	messageBusErr := r.Messages.Close()
	databaseErr := r.Database.Close()
	redisErr := r.Redis.Close()
	errorLogErr := r.Errors.Close()
	if databaseErr != nil || redisErr != nil || errorLogErr != nil || messageBusErr != nil {
		return fmt.Errorf("close resources: database=%v redis=%v error_log=%v message_bus=%v", databaseErr, redisErr, errorLogErr, messageBusErr)
	}
	return nil
}

func NewHTTPServer(cfg config.Config, resources *Resources) *http.Server {
	authRepository := persistencerepository.NewAuth(resources.Database.ORM)
	passwordHasher := security.NewPasswordHasher()
	tokenManager := security.NewTokenManager(cfg.JWTAccessSecret, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)
	authService := authusecase.NewService(authRepository, passwordHasher, tokenManager)
	authHandler := httpdelivery.NewAuthHandler(authService)
	userRepository := persistencerepository.NewUser(resources.Database.ORM)
	userService := userusecase.NewService(userRepository)
	userHandler := httpdelivery.NewUserHandler(userService)
	conversationRepository := persistencerepository.NewConversation(resources.Database.ORM)
	conversationService := conversationusecase.NewService(conversationRepository)
	conversationHandler := httpdelivery.NewConversationHandler(conversationService)
	mediaRepository := persistencerepository.NewMediaUpload(resources.Database.ORM)
	mediaService := mediausecase.NewService(mediaRepository, resources.Storage)
	mediaHandler := httpdelivery.NewMediaHandler(mediaService)
	messageRepository := persistencerepository.NewMessage(resources.Database.ORM)
	messageService := messageusecase.NewService(messageRepository, resources.Messages, mediaService)
	messageHandler := httpdelivery.NewMessageHandler(messageService)
	websocketHandler := wsdelivery.NewHandler(resources.Realtime, messageService, resources.Errors)
	return &http.Server{
		Addr:         cfg.HTTPAddress(),
		Handler:      httpdelivery.NewRouter(cfg.AppEnv, authHandler, userHandler, conversationHandler, messageHandler, mediaHandler, websocketHandler, tokenManager, resources.Errors, resources.Database.Ping, resources.Redis.Ping),
		ReadTimeout:  cfg.HTTPReadTimeout,
		WriteTimeout: cfg.HTTPWriteTimeout,
	}
}
