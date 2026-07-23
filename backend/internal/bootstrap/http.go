package bootstrap

import (
	"context"
	"fmt"
	"net/http"

	"github.com/jetomax/realtime-chat/backend/internal/config"
	httpdelivery "github.com/jetomax/realtime-chat/backend/internal/delivery/http"
	"github.com/jetomax/realtime-chat/backend/internal/infrastructure/cache"
	"github.com/jetomax/realtime-chat/backend/internal/infrastructure/errorlog"
	"github.com/jetomax/realtime-chat/backend/internal/infrastructure/persistence"
	persistencerepository "github.com/jetomax/realtime-chat/backend/internal/infrastructure/persistence/repository"
	"github.com/jetomax/realtime-chat/backend/internal/infrastructure/security"
	authusecase "github.com/jetomax/realtime-chat/backend/internal/usecase/auth"
	conversationusecase "github.com/jetomax/realtime-chat/backend/internal/usecase/conversation"
	userusecase "github.com/jetomax/realtime-chat/backend/internal/usecase/user"
)

type Resources struct {
	Database *persistence.Database
	Redis    *cache.Redis
	Errors   *errorlog.Recorder
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
	return &Resources{Database: database, Redis: redisClient, Errors: errorRecorder}, nil
}

func (r *Resources) Close() error {
	databaseErr := r.Database.Close()
	redisErr := r.Redis.Close()
	errorLogErr := r.Errors.Close()
	if databaseErr != nil || redisErr != nil || errorLogErr != nil {
		return fmt.Errorf("close resources: database=%v redis=%v error_log=%v", databaseErr, redisErr, errorLogErr)
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
	return &http.Server{
		Addr:         cfg.HTTPAddress(),
		Handler:      httpdelivery.NewRouter(cfg.AppEnv, authHandler, userHandler, conversationHandler, tokenManager, resources.Errors, resources.Database.Ping, resources.Redis.Ping),
		ReadTimeout:  cfg.HTTPReadTimeout,
		WriteTimeout: cfg.HTTPWriteTimeout,
	}
}
