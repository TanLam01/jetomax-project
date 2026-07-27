package http

import (
	"context"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
	wsdelivery "github.com/jetomax/realtime-chat/backend/internal/delivery/websocket"
	"github.com/jetomax/realtime-chat/backend/internal/domain/repository"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/jetomax/realtime-chat/backend/docs"
)

type HealthCheck func(context.Context) error

func NewRouter(environment string, authHandler *AuthHandler, userHandler *UserHandler, conversationHandler *ConversationHandler, messageHandler *MessageHandler, mediaHandler *MediaHandler, websocketHandler *wsdelivery.Handler, verifier AccessTokenVerifier, errorRecorder repository.ErrorRecorder, checks ...HealthCheck) *gin.Engine {
	if environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Logger())
	if errorRecorder != nil {
		router.Use(ErrorAudit(errorRecorder))
	}
	router.Use(gin.CustomRecovery(func(c *gin.Context, recovered any) {
		slog.Error("HTTP handler panic", "error", recovered, "stack", string(debug.Stack()))
		if !c.Writer.Written() {
			respondError(c, http.StatusInternalServerError, "handler_panic", "handler panic")
		}
	}))
	router.GET("/health/ready", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		for _, check := range checks {
			if err := check(ctx); err != nil {
				respondError(c, http.StatusServiceUnavailable, "readiness_check_failed", err.Error())
				return
			}
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.GET("/ws", RequireAuth(verifier), websocketHandler.Connect)

	v1 := router.Group("/api/v1")
	v1.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"service": "realtime-chat-api", "version": "v1"})
	})
	auth := v1.Group("/auth")
	auth.POST("/register", authHandler.Register)
	auth.POST("/login", authHandler.Login)
	auth.POST("/refresh", authHandler.Refresh)
	auth.POST("/logout", authHandler.Logout)
	protected := v1.Group("")
	protected.Use(RequireAuth(verifier))
	users := protected.Group("/users")
	users.GET("/me", userHandler.Me)
	users.GET("", userHandler.Search)
	protected.GET("/conversations", conversationHandler.List)
	protected.POST("/conversations/direct", conversationHandler.CreateDirect)
	protected.POST("/conversations/groups", conversationHandler.CreateGroup)
	protected.POST("/conversations/:id/members", conversationHandler.AddMembers)
	protected.GET("/conversations/:id/members", conversationHandler.ListMembers)
	protected.DELETE("/conversations/:id/members/:userId", conversationHandler.RemoveMember)
	protected.PATCH("/conversations/:id/members/:userId/role", conversationHandler.UpdateMemberRole)
	protected.POST("/conversations/:id/ownership", conversationHandler.TransferOwnership)
	protected.GET("/conversations/:id/messages", messageHandler.List)
	protected.POST("/media/uploads", mediaHandler.CreateUpload)
	protected.GET("/media/attachments/:id/download", mediaHandler.CreateDownload)
	router.NoRoute(func(c *gin.Context) {
		respondError(c, http.StatusNotFound, "route_not_found", "route not found")
	})
	return router
}
