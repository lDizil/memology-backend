package middleware

import (
	"net/http"
	"strings"

	"memology-backend/internal/services"

	"github.com/gin-gonic/gin"
)

func JWTAuth(authService services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var token string

		token, err := c.Cookie("access_token")
		if err != nil || token == "" {
			authHeader := c.GetHeader("Authorization")
			if authHeader == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization required"})
				c.Abort()
				return
			}

			if strings.HasPrefix(authHeader, "Bearer ") {
				token = strings.TrimPrefix(authHeader, "Bearer ")
			} else {
				token = authHeader
			}
		}

		claims, err := authService.ValidateToken(c.Request.Context(), token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Next()
	}
}
