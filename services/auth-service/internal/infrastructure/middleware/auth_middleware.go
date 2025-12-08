package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"users-service/internal/infrastructure/auth"
)

type AuthMiddleware struct {
	jwtService auth.JWTService
}

func NewAuthMiddleware(jwtService auth.JWTService) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService: jwtService,
	}
}

func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := m.extractTokenFromHeader(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization token required"})
			c.Abort()
			return
		}

		claims, err := m.jwtService.ValidateAccessToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Set user information in context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("token", token)

		c.Next()
	}
}

func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := m.extractTokenFromHeader(c)
		if token != "" {
			claims, err := m.jwtService.ValidateAccessToken(token)
			if err == nil {
				// Set user information in context if token is valid
				c.Set("user_id", claims.UserID)
				c.Set("user_email", claims.Email)
				c.Set("token", token)
			}
		}

		c.Next()
	}
}

func (m *AuthMiddleware) extractTokenFromHeader(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}

	return parts[1]
} 