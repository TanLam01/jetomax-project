package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jetomax/realtime-chat/backend/internal/bootstrap"
	"github.com/jetomax/realtime-chat/backend/internal/config"
)

// @title Realtime Chat API
// @version 1.0
// @description REST API for the realtime chat application.
// @BasePath /api/v1
// @schemes http https
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter: Bearer {access_token}

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("load configuration", "error", err)
		os.Exit(1)
	}

	connectCtx, connectCancel := context.WithTimeout(context.Background(), 10*time.Second)
	resources, err := bootstrap.Connect(connectCtx, cfg)
	connectCancel()
	if err != nil {
		slog.Error("connect infrastructure", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := resources.Close(); err != nil {
			slog.Error("close infrastructure", "error", err)
		}
	}()

	server := bootstrap.NewHTTPServer(cfg, resources)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		slog.Info("API server started", "address", server.Addr, "environment", cfg.AppEnv)
		errCh <- server.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("API server failed", "error", err)
			os.Exit(1)
		}
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTPShutdownTimeout)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			slog.Error("API server shutdown", "error", err)
			os.Exit(1)
		}
		slog.Info("API server stopped")
	}
}
