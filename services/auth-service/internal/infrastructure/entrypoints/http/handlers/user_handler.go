package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	pkgErrors "github.com/melegattip/giia-core-engine/pkg/errors"
	pkgLogger "github.com/melegattip/giia-core-engine/pkg/logger"
	"github.com/melegattip/giia-core-engine/services/auth-service/internal/core/usecases/user"
	"github.com/melegattip/giia-core-engine/services/auth-service/internal/infrastructure/entrypoints/http/middleware"
)

type UserHandler struct {
	activateUserUseCase   *user.ActivateUserUseCase
	deactivateUserUseCase *user.DeactivateUserUseCase
	logger                pkgLogger.Logger
}

func NewUserHandler(
	activateUserUseCase *user.ActivateUserUseCase,
	deactivateUserUseCase *user.DeactivateUserUseCase,
	logger pkgLogger.Logger,
) *UserHandler {
	return &UserHandler{
		activateUserUseCase:   activateUserUseCase,
		deactivateUserUseCase: deactivateUserUseCase,
		logger:                logger,
	}
}

func (h *UserHandler) ActivateUser(c *gin.Context) {
	adminUserID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, pkgErrors.ToHTTPResponse(
			pkgErrors.NewUnauthorized("authentication required"),
		))
		return
	}

	targetUserIDStr := c.Param("id")
	if targetUserIDStr == "" {
		c.JSON(http.StatusBadRequest, pkgErrors.ToHTTPResponse(
			pkgErrors.NewBadRequest("user ID is required"),
		))
		return
	}

	targetUserID, err := strconv.Atoi(targetUserIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, pkgErrors.ToHTTPResponse(
			pkgErrors.NewBadRequest("invalid user ID format"),
		))
		return
	}

	if err := h.activateUserUseCase.Execute(c.Request.Context(), adminUserID, targetUserID); err != nil {
		if customErr, ok := err.(*pkgErrors.CustomError); ok {
			c.JSON(customErr.HTTPStatus, pkgErrors.ToHTTPResponse(err))
		} else {
			c.JSON(http.StatusInternalServerError, pkgErrors.ToHTTPResponse(
				pkgErrors.NewInternalServerError("internal server error"),
			))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User activated successfully",
	})
}

func (h *UserHandler) DeactivateUser(c *gin.Context) {
	adminUserID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, pkgErrors.ToHTTPResponse(
			pkgErrors.NewUnauthorized("authentication required"),
		))
		return
	}

	targetUserIDStr := c.Param("id")
	if targetUserIDStr == "" {
		c.JSON(http.StatusBadRequest, pkgErrors.ToHTTPResponse(
			pkgErrors.NewBadRequest("user ID is required"),
		))
		return
	}

	targetUserID, err := strconv.Atoi(targetUserIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, pkgErrors.ToHTTPResponse(
			pkgErrors.NewBadRequest("invalid user ID format"),
		))
		return
	}

	if err := h.deactivateUserUseCase.Execute(c.Request.Context(), adminUserID, targetUserID); err != nil {
		if customErr, ok := err.(*pkgErrors.CustomError); ok {
			c.JSON(customErr.HTTPStatus, pkgErrors.ToHTTPResponse(err))
		} else {
			c.JSON(http.StatusInternalServerError, pkgErrors.ToHTTPResponse(
				pkgErrors.NewInternalServerError("internal server error"),
			))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User deactivated successfully",
	})
}
