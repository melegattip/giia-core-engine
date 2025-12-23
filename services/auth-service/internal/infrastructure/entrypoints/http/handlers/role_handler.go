package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	pkgErrors "github.com/melegattip/giia-core-engine/pkg/errors"
	pkgLogger "github.com/melegattip/giia-core-engine/pkg/logger"
	"github.com/melegattip/giia-core-engine/services/auth-service/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/auth-service/internal/core/usecases/role"
	"github.com/melegattip/giia-core-engine/services/auth-service/internal/infrastructure/entrypoints/http/middleware"
)

type RoleHandler struct {
	assignRoleUseCase *role.AssignRoleUseCase
	createRoleUseCase *role.CreateRoleUseCase
	updateRoleUseCase *role.UpdateRoleUseCase
	deleteRoleUseCase *role.DeleteRoleUseCase
	logger            pkgLogger.Logger
}

func NewRoleHandler(
	assignRoleUseCase *role.AssignRoleUseCase,
	createRoleUseCase *role.CreateRoleUseCase,
	updateRoleUseCase *role.UpdateRoleUseCase,
	deleteRoleUseCase *role.DeleteRoleUseCase,
	logger pkgLogger.Logger,
) *RoleHandler {
	return &RoleHandler{
		assignRoleUseCase: assignRoleUseCase,
		createRoleUseCase: createRoleUseCase,
		updateRoleUseCase: updateRoleUseCase,
		deleteRoleUseCase: deleteRoleUseCase,
		logger:            logger,
	}
}

func (h *RoleHandler) CreateRole(c *gin.Context) {
	var req domain.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkgErrors.ToHTTPResponse(
			pkgErrors.NewBadRequest("invalid request body"),
		))
		return
	}

	role, err := h.createRoleUseCase.Execute(c.Request.Context(), &req)
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

	c.JSON(http.StatusCreated, role.ToResponse())
}

func (h *RoleHandler) UpdateRole(c *gin.Context) {
	roleIDStr := c.Param("roleId")
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, pkgErrors.ToHTTPResponse(
			pkgErrors.NewBadRequest("invalid role ID format"),
		))
		return
	}

	var req domain.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkgErrors.ToHTTPResponse(
			pkgErrors.NewBadRequest("invalid request body"),
		))
		return
	}

	role, err := h.updateRoleUseCase.Execute(c.Request.Context(), roleID, &req)
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

	c.JSON(http.StatusOK, role.ToResponse())
}

func (h *RoleHandler) DeleteRole(c *gin.Context) {
	roleIDStr := c.Param("roleId")
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, pkgErrors.ToHTTPResponse(
			pkgErrors.NewBadRequest("invalid role ID format"),
		))
		return
	}

	if err := h.deleteRoleUseCase.Execute(c.Request.Context(), roleID); err != nil {
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
		"message": "Role deleted successfully",
	})
}

func (h *RoleHandler) AssignRole(c *gin.Context) {
	var req domain.AssignRoleRequest
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

	roleID, err := uuid.Parse(req.RoleID)
	if err != nil {
		c.JSON(http.StatusBadRequest, pkgErrors.ToHTTPResponse(
			pkgErrors.NewBadRequest("invalid role ID format"),
		))
		return
	}

	assignedByID, exists := c.Get(string(middleware.UserIDKey))
	if !exists {
		c.JSON(http.StatusUnauthorized, pkgErrors.ToHTTPResponse(
			pkgErrors.NewUnauthorized("user not authenticated"),
		))
		return
	}

	assignedByUUID, ok := assignedByID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, pkgErrors.ToHTTPResponse(
			pkgErrors.NewInternalServerError("invalid user ID in context"),
		))
		return
	}

	if err := h.assignRoleUseCase.Execute(c.Request.Context(), userID, roleID, assignedByUUID); err != nil {
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
		"message": "Role assigned successfully",
	})
}
