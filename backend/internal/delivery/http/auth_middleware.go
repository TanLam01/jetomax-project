package http

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const authenticatedUserIDKey = "authenticated_user_id"

type AccessTokenVerifier interface{ VerifyAccessToken(string) (string, error) }

func RequireAuth(verifier AccessTokenVerifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		parts := strings.Fields(c.GetHeader("Authorization"))
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			respondError(c, http.StatusUnauthorized, "unauthorized", "missing or invalid bearer token")
			c.Abort()
			return
		}
		userID, err := verifier.VerifyAccessToken(parts[1])
		if err != nil {
			respondError(c, http.StatusUnauthorized, "unauthorized", "invalid or expired access token")
			c.Abort()
			return
		}
		c.Set(authenticatedUserIDKey, userID)
		c.Next()
	}
}

func authenticatedUserID(c *gin.Context) string { return c.GetString(authenticatedUserIDKey) }
