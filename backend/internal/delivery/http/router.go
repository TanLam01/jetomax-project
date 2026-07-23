package http

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jetomax/realtime-chat/backend/internal/delivery/http/dto"
	"github.com/jetomax/realtime-chat/backend/internal/domain/repository"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/jetomax/realtime-chat/backend/docs"
)

type HealthCheck func(context.Context) error

func NewRouter(environment string, authHandler *AuthHandler, userHandler *UserHandler, conversationHandler *ConversationHandler, verifier AccessTokenVerifier, errorRecorder repository.ErrorRecorder, checks ...HealthCheck) *gin.Engine {
	if environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Logger())
	if errorRecorder != nil {
		router.Use(ErrorAudit(errorRecorder))
	}
	router.Use(gin.Recovery())
	router.GET("/health/ready", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		for _, check := range checks {
			if err := check(ctx); err != nil {
				c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not_ready"})
				return
			}
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

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
	router.NoRoute(func(c *gin.Context) {
		setSafeError(c, "not_found", "route not found")
		c.JSON(http.StatusNotFound, dto.NewErrorResponse("not_found", "route not found"))
	})
	return router
}
