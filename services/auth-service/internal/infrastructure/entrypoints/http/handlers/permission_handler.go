package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	pkgErrors "github.com/giia/giia-core-engine/pkg/errors"
	pkgLogger "github.com/giia/giia-core-engine/pkg/logger"
	"github.com/giia/giia-core-engine/services/auth-service/internal/core/domain"
	"github.com/giia/giia-core-engine/services/auth-service/internal/core/usecases/rbac"
	"github.com/google/uuid"
)

type PermissionHandler struct {
	checkPermissionUseCase *rbac.CheckPermissionUseCase
	batchCheckUseCase      *rbac.BatchCheckPermissionsUseCase
	logger                 pkgLogger.Logger
}

func NewPermissionHandler(
	checkPermissionUseCase *rbac.CheckPermissionUseCase,
	batchCheckUseCase *rbac.BatchCheckPermissionsUseCase,
	logger pkgLogger.Logger,
) *PermissionHandler {
	return &PermissionHandler{
		checkPermissionUseCase: checkPermissionUseCase,
		batchCheckUseCase:      batchCheckUseCase,
		logger:                 logger,
	}
}

func (h *PermissionHandler) CheckPermission(c *gin.Context) {
	var req domain.CheckPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkgErrors.ToHTTPResponse(
			pkgErrors.NewBadRequest("invalid request body"),
		))
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, pkgErrors.ToHTTPResponse(
			pkgErrors.NewBadRequest("invalid user ID format"),
		))
		return
	}

	allowed, err := h.checkPermissionUseCase.Execute(c.Request.Context(), userID, req.Permission)
	if err != nil {
		if customErr, ok := err.(*pkgErrors.CustomError); ok {
			c.JSON(customErr.HTTPStatus, pkgErrors.ToHTTPResponse(err))
		} else {
			c.JSON(http.StatusInternalServerError, pkgErrors.ToHTTPResponse(
				pkgErrors.NewInternalServerError("internal server error"),
			))
		}
		return
	}

	c.JSON(http.StatusOK, domain.CheckPermissionResponse{
		Allowed: allowed,
	})
}

func (h *PermissionHandler) BatchCheckPermissions(c *gin.Context) {
	var req domain.BatchCheckPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkgErrors.ToHTTPResponse(
			pkgErrors.NewBadRequest("invalid request body"),
		))
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, pkgErrors.ToHTTPResponse(
			pkgErrors.NewBadRequest("invalid user ID format"),
		))
		return
	}

	results, err := h.batchCheckUseCase.Execute(c.Request.Context(), userID, req.Permissions)
	if err != nil {
		if customErr, ok := err.(*pkgErrors.CustomError); ok {
			c.JSON(customErr.HTTPStatus, pkgErrors.ToHTTPResponse(err))
		} else {
			c.JSON(http.StatusInternalServerError, pkgErrors.ToHTTPResponse(
				pkgErrors.NewInternalServerError("internal server error"),
			))
		}
		return
	}

	c.JSON(http.StatusOK, domain.BatchCheckPermissionResponse{
		Results: results,
	})
}
