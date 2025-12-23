package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	pkgErrors "github.com/melegattip/giia-core-engine/pkg/errors"
	pkgLogger "github.com/melegattip/giia-core-engine/pkg/logger"
	"github.com/melegattip/giia-core-engine/services/auth-service/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/auth-service/internal/core/usecases/auth"
	"github.com/melegattip/giia-core-engine/services/auth-service/internal/infrastructure/entrypoints/http/middleware"
)

type AuthHandler struct {
	loginUseCase                  *auth.LoginUseCase
	registerUseCase               *auth.RegisterUseCase
	refreshTokenUseCase           *auth.RefreshTokenUseCase
	logoutUseCase                 *auth.LogoutUseCase
	activateAccountUseCase        *auth.ActivateAccountUseCase
	requestPasswordResetUseCase   *auth.RequestPasswordResetUseCase
	confirmPasswordResetUseCase   *auth.ConfirmPasswordResetUseCase
	logger                        pkgLogger.Logger
}

func NewAuthHandler(
	loginUseCase *auth.LoginUseCase,
	registerUseCase *auth.RegisterUseCase,
	refreshTokenUseCase *auth.RefreshTokenUseCase,
	logoutUseCase *auth.LogoutUseCase,
	activateAccountUseCase *auth.ActivateAccountUseCase,
	requestPasswordResetUseCase *auth.RequestPasswordResetUseCase,
	confirmPasswordResetUseCase *auth.ConfirmPasswordResetUseCase,
	logger pkgLogger.Logger,
) *AuthHandler {
	return &AuthHandler{
		loginUseCase:                  loginUseCase,
		registerUseCase:               registerUseCase,
		refreshTokenUseCase:           refreshTokenUseCase,
		logoutUseCase:                 logoutUseCase,
		activateAccountUseCase:        activateAccountUseCase,
		requestPasswordResetUseCase:   requestPasswordResetUseCase,
		confirmPasswordResetUseCase:   confirmPasswordResetUseCase,
		logger:                        logger,
	}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req domain.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkgErrors.ToHTTPResponse(
			pkgErrors.NewBadRequest("invalid request body"),
		))
		return
	}

	response, err := h.loginUseCase.Execute(c.Request.Context(), &req)
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

	c.SetCookie(
		"refresh_token",
		response.RefreshToken,
		int(7*24*60*60),
		"/",
		"",
		false,
		true,
	)

	c.JSON(http.StatusOK, gin.H{
		"access_token": response.AccessToken,
		"expires_in":   response.ExpiresIn,
		"user":         response.User,
	})
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req domain.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkgErrors.ToHTTPResponse(
			pkgErrors.NewBadRequest("invalid request body"),
		))
		return
	}

	if err := h.registerUseCase.Execute(c.Request.Context(), &req); err != nil {
		if customErr, ok := err.(*pkgErrors.CustomError); ok {
			c.JSON(customErr.HTTPStatus, pkgErrors.ToHTTPResponse(err))
		} else {
			c.JSON(http.StatusInternalServerError, pkgErrors.ToHTTPResponse(
				pkgErrors.NewInternalServerError("internal server error"),
			))
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully. Please check your email for activation instructions.",
	})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil || refreshToken == "" {
		var req domain.RefreshTokenRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, pkgErrors.ToHTTPResponse(
				pkgErrors.NewBadRequest("refresh token is required"),
			))
			return
		}
		refreshToken = req.RefreshToken
	}

	accessToken, err := h.refreshTokenUseCase.Execute(c.Request.Context(), refreshToken)
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

	c.JSON(http.StatusOK, gin.H{
		"access_token": accessToken,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, pkgErrors.ToHTTPResponse(err))
		return
	}

	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusBadRequest, pkgErrors.ToHTTPResponse(
			pkgErrors.NewBadRequest("authorization header is required"),
		))
		return
	}

	accessToken := authHeader
	if len(authHeader) > 7 && strings.HasPrefix(authHeader, "Bearer ") {
		accessToken = authHeader[7:]
	}

	if err := h.logoutUseCase.Execute(c.Request.Context(), accessToken, userID); err != nil {
		if customErr, ok := err.(*pkgErrors.CustomError); ok {
			c.JSON(customErr.HTTPStatus, pkgErrors.ToHTTPResponse(err))
		} else {
			c.JSON(http.StatusInternalServerError, pkgErrors.ToHTTPResponse(
				pkgErrors.NewInternalServerError("internal server error"),
			))
		}
		return
	}

	c.SetCookie(
		"refresh_token",
		"",
		-1,
		"/",
		"",
		false,
		true,
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}

func (h *AuthHandler) Activate(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		var req domain.ActivateAccountRequest
		if err := c.ShouldBindJSON(&req); err != nil || req.Token == "" {
			c.JSON(http.StatusBadRequest, pkgErrors.ToHTTPResponse(
				pkgErrors.NewBadRequest("activation token is required"),
			))
			return
		}
		token = req.Token
	}

	if err := h.activateAccountUseCase.Execute(c.Request.Context(), token); err != nil {
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
		"message": "Account activated successfully. You can now log in.",
	})
}

func (h *AuthHandler) RequestPasswordReset(c *gin.Context) {
	var req domain.PasswordResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkgErrors.ToHTTPResponse(
			pkgErrors.NewBadRequest("invalid request body"),
		))
		return
	}

	orgID, err := middleware.GetOrganizationID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, pkgErrors.ToHTTPResponse(
			pkgErrors.NewBadRequest("organization ID is required"),
		))
		return
	}

	if err := h.requestPasswordResetUseCase.Execute(c.Request.Context(), req.Email, orgID); err != nil {
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
		"message": "If the email exists, a password reset link has been sent.",
	})
}

func (h *AuthHandler) ConfirmPasswordReset(c *gin.Context) {
	var req domain.PasswordResetComplete
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkgErrors.ToHTTPResponse(
			pkgErrors.NewBadRequest("invalid request body"),
		))
		return
	}

	if err := h.confirmPasswordResetUseCase.Execute(c.Request.Context(), req.Token, req.NewPassword); err != nil {
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
		"message": "Password has been reset successfully. You can now log in with your new password.",
	})
}
