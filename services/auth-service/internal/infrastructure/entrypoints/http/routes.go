package http

import (
	"github.com/gin-gonic/gin"

	pkgLogger "github.com/melegattip/giia-core-engine/pkg/logger"
	"github.com/melegattip/giia-core-engine/services/auth-service/internal/infrastructure/entrypoints/http/handlers"
	"github.com/melegattip/giia-core-engine/services/auth-service/internal/infrastructure/entrypoints/http/middleware"
)

type RouteConfig struct {
	AuthHandler       *handlers.AuthHandler
	UserHandler       *handlers.UserHandler
	RoleHandler       *handlers.RoleHandler
	PermissionHandler *handlers.PermissionHandler
	Logger            pkgLogger.Logger
}

func SetupRoutes(router *gin.Engine, config *RouteConfig) {
	v1 := router.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/register", config.AuthHandler.Register)
			auth.POST("/login", config.AuthHandler.Login)
			auth.POST("/refresh", config.AuthHandler.Refresh)
			auth.POST("/verify", config.AuthHandler.Activate)
			auth.POST("/reset-password", config.AuthHandler.RequestPasswordReset)
			auth.POST("/confirm-reset", config.AuthHandler.ConfirmPasswordReset)

			authenticated := auth.Group("")
			authenticated.Use(middleware.AuthMiddleware(config.Logger))
			{
				authenticated.POST("/logout", config.AuthHandler.Logout)
			}
		}

		users := v1.Group("/users")
		users.Use(middleware.AuthMiddleware(config.Logger))
		{
			users.PUT("/:id/activate", config.UserHandler.ActivateUser)
			users.PUT("/:id/deactivate", config.UserHandler.DeactivateUser)
		}

		roles := v1.Group("/roles")
		roles.Use(middleware.AuthMiddleware(config.Logger))
		{
			roles.POST("", config.RoleHandler.CreateRole)
			roles.PUT("/:id", config.RoleHandler.UpdateRole)
			roles.DELETE("/:id", config.RoleHandler.DeleteRole)
		}

		permissions := v1.Group("/permissions")
		permissions.Use(middleware.AuthMiddleware(config.Logger))
		{
			permissions.POST("/check", config.PermissionHandler.CheckPermission)
			permissions.POST("/batch-check", config.PermissionHandler.BatchCheckPermissions)
		}
	}

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"service": "auth-service",
		})
	})
}
