package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	pkgErrors "github.com/melegattip/giia-core-engine/pkg/errors"
	pkgLogger "github.com/melegattip/giia-core-engine/pkg/logger"
	"github.com/melegattip/giia-core-engine/services/auth-service/internal/infrastructure/adapters/jwt"
)

func AuthMiddleware(logger pkgLogger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, pkgErrors.ToHTTPResponse(
				pkgErrors.NewUnauthorized("missing authorization header"),
			))
			c.Abort()
			return
		}

		tokenString := authHeader
		if len(authHeader) > 7 && strings.HasPrefix(authHeader, "Bearer ") {
			tokenString = authHeader[7:]
		}

		jwtManager := c.MustGet("jwtManager").(*jwt.JWTManager)

		claims, err := jwtManager.ValidateAccessToken(tokenString)
		if err != nil {
			logger.Warn(c.Request.Context(), "Invalid or expired token", pkgLogger.Tags{
				"error": err.Error(),
			})
			c.JSON(http.StatusUnauthorized, pkgErrors.ToHTTPResponse(
				pkgErrors.NewUnauthorized("invalid or expired token"),
			))
			c.Abort()
			return
		}

		orgID, err := uuid.Parse(claims.OrganizationID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, pkgErrors.ToHTTPResponse(
				pkgErrors.NewUnauthorized("invalid organization ID in token"),
			))
			c.Abort()
			return
		}

		userID, err := uuid.Parse(claims.UserID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, pkgErrors.ToHTTPResponse(
				pkgErrors.NewUnauthorized("invalid user ID in token"),
			))
			c.Abort()
			return
		}

		c.Set(string(OrganizationIDKey), orgID)
		c.Set(string(UserIDKey), userID)
		c.Set("email", claims.Email)
		c.Set("roles", claims.Roles)

		c.Next()
	}
}
