package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	pkgErrors "github.com/giia/giia-core-engine/pkg/errors"
	pkgLogger "github.com/giia/giia-core-engine/pkg/logger"
	"github.com/giia/giia-core-engine/services/auth-service/internal/core/usecases/rbac"
	"github.com/google/uuid"
)

type PermissionMiddleware struct {
	checkPermissionUseCase *rbac.CheckPermissionUseCase
	logger                 pkgLogger.Logger
}

func NewPermissionMiddleware(
	checkPermissionUseCase *rbac.CheckPermissionUseCase,
	logger pkgLogger.Logger,
) *PermissionMiddleware {
	return &PermissionMiddleware{
		checkPermissionUseCase: checkPermissionUseCase,
		logger:                 logger,
	}
}

func (m *PermissionMiddleware) RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDValue, exists := c.Get(string(UserIDKey))
		if !exists {
			c.JSON(http.StatusUnauthorized, pkgErrors.ToHTTPResponse(
				pkgErrors.NewUnauthorized("user not authenticated"),
			))
			c.Abort()
			return
		}

		userID, ok := userIDValue.(uuid.UUID)
		if !ok {
			c.JSON(http.StatusInternalServerError, pkgErrors.ToHTTPResponse(
				pkgErrors.NewInternalServerError("invalid user ID in context"),
			))
			c.Abort()
			return
		}

		allowed, err := m.checkPermissionUseCase.Execute(c.Request.Context(), userID, permission)
		if err != nil {
			m.logger.Error(c.Request.Context(), err, "Permission check failed", pkgLogger.Tags{
				"user_id":    userID.String(),
				"permission": permission,
			})
			c.JSON(http.StatusInternalServerError, pkgErrors.ToHTTPResponse(
				pkgErrors.NewInternalServerError("permission check failed"),
			))
			c.Abort()
			return
		}

		if !allowed {
			m.logger.Warn(c.Request.Context(), "Permission denied", pkgLogger.Tags{
				"user_id":    userID.String(),
				"permission": permission,
				"method":     c.Request.Method,
				"path":       c.Request.URL.Path,
			})
			c.JSON(http.StatusForbidden, pkgErrors.ToHTTPResponse(
				pkgErrors.NewForbidden("insufficient permissions"),
			))
			c.Abort()
			return
		}

		m.logger.Debug(c.Request.Context(), "Permission granted", pkgLogger.Tags{
			"user_id":    userID.String(),
			"permission": permission,
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
		})

		c.Next()
	}
}

func (m *PermissionMiddleware) RequireAnyPermission(permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDValue, exists := c.Get(string(UserIDKey))
		if !exists {
			c.JSON(http.StatusUnauthorized, pkgErrors.ToHTTPResponse(
				pkgErrors.NewUnauthorized("user not authenticated"),
			))
			c.Abort()
			return
		}

		userID, ok := userIDValue.(uuid.UUID)
		if !ok {
			c.JSON(http.StatusInternalServerError, pkgErrors.ToHTTPResponse(
				pkgErrors.NewInternalServerError("invalid user ID in context"),
			))
			c.Abort()
			return
		}

		hasPermission := false
		for _, permission := range permissions {
			allowed, err := m.checkPermissionUseCase.Execute(c.Request.Context(), userID, permission)
			if err != nil {
				m.logger.Error(c.Request.Context(), err, "Permission check failed", pkgLogger.Tags{
					"user_id":    userID.String(),
					"permission": permission,
				})
				continue
			}

			if allowed {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			m.logger.Warn(c.Request.Context(), "No matching permission found", pkgLogger.Tags{
				"user_id":     userID.String(),
				"permissions": permissions,
				"method":      c.Request.Method,
				"path":        c.Request.URL.Path,
			})
			c.JSON(http.StatusForbidden, pkgErrors.ToHTTPResponse(
				pkgErrors.NewForbidden("insufficient permissions"),
			))
			c.Abort()
			return
		}

		c.Next()
	}
}
