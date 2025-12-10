package client

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthMiddleware struct {
	client *AuthClient
}

func NewAuthMiddleware(client *AuthClient) *AuthMiddleware {
	return &AuthMiddleware{
		client: client,
	}
}

func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(401, gin.H{"error": "authorization header required"})
			c.Abort()
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader {
			c.JSON(401, gin.H{"error": "invalid authorization header format"})
			c.Abort()
			return
		}

		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		resp, err := m.client.ValidateToken(c.Request.Context(), token, requestID)
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to validate token"})
			c.Abort()
			return
		}

		if !resp.Valid {
			c.JSON(401, gin.H{"error": resp.Reason})
			c.Abort()
			return
		}

		c.Set("user_id", resp.User.UserId)
		c.Set("organization_id", resp.User.OrganizationId)
		c.Set("email", resp.User.Email)
		c.Set("roles", resp.User.Roles)

		c.Next()
	}
}

func (m *AuthMiddleware) RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(401, gin.H{"error": "user not authenticated"})
			c.Abort()
			return
		}

		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		resp, err := m.client.CheckPermission(c.Request.Context(), userID.(string), permission, requestID)
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to check permission"})
			c.Abort()
			return
		}

		if !resp.Allowed {
			c.JSON(403, gin.H{"error": resp.Reason})
			c.Abort()
			return
		}

		c.Next()
	}
}

func ExtractUserIDFromContext(ctx context.Context) (string, bool) {
	if ginCtx, ok := ctx.(*gin.Context); ok {
		if userID, exists := ginCtx.Get("user_id"); exists {
			if userIDStr, ok := userID.(string); ok {
				return userIDStr, true
			}
		}
	}
	return "", false
}

func ExtractOrganizationIDFromContext(ctx context.Context) (string, bool) {
	if ginCtx, ok := ctx.(*gin.Context); ok {
		if orgID, exists := ginCtx.Get("organization_id"); exists {
			if orgIDStr, ok := orgID.(string); ok {
				return orgIDStr, true
			}
		}
	}
	return "", false
}
