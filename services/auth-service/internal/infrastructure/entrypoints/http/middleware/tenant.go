package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	pkgErrors "github.com/giia/giia-core-engine/pkg/errors"
	"github.com/giia/giia-core-engine/services/auth-service/internal/infrastructure/adapters/jwt"
	"github.com/google/uuid"
)

type contextKey string

const (
	OrganizationIDKey contextKey = "organization_id"
	UserIDKey         contextKey = "user_id"
)

type TenantMiddleware struct {
	jwtManager *jwt.JWTManager
}

func NewTenantMiddleware(jwtManager *jwt.JWTManager) *TenantMiddleware {
	return &TenantMiddleware{
		jwtManager: jwtManager,
	}
}

func (m *TenantMiddleware) ExtractTenantContext() gin.HandlerFunc {
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
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			tokenString = authHeader[7:]
		}

		claims, err := m.jwtManager.ValidateAccessToken(tokenString)
		if err != nil {
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

func GetOrganizationID(c *gin.Context) (uuid.UUID, error) {
	orgID, exists := c.Get(string(OrganizationIDKey))
	if !exists {
		return uuid.Nil, pkgErrors.NewUnauthorized("organization ID not found in context")
	}

	id, ok := orgID.(uuid.UUID)
	if !ok {
		return uuid.Nil, pkgErrors.NewInternalServerError("invalid organization ID type in context")
	}

	return id, nil
}

func GetUserID(c *gin.Context) (uuid.UUID, error) {
	userID, exists := c.Get(string(UserIDKey))
	if !exists {
		return uuid.Nil, pkgErrors.NewUnauthorized("user ID not found in context")
	}

	id, ok := userID.(uuid.UUID)
	if !ok {
		return uuid.Nil, pkgErrors.NewInternalServerError("invalid user ID type in context")
	}

	return id, nil
}
