package handlers

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/giia/giia-core-engine/services/auth-service/internal/domain"
	"github.com/giia/giia-core-engine/services/auth-service/internal/usecases"
	"github.com/giia/giia-core-engine/services/auth-service/pkg/imageprocessor"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService usecases.UserService
}

func NewUserHandler(userService usecases.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) Register(c *gin.Context) {
	var req domain.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, tokens, err := h.userService.Register(c.Request.Context(), &req)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		if strings.Contains(err.Error(), "password") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"user":          user,
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"expires_at":    tokens.ExpiresAt.Unix(),
		"token_type":    tokens.TokenType,
	})
}

func (h *UserHandler) Login(c *gin.Context) {
	var req domain.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, tokens, err := h.userService.Login(c.Request.Context(), &req)
	if err != nil {
		if strings.Contains(err.Error(), "invalid email or password") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}
		if strings.Contains(err.Error(), "locked") {
			c.JSON(http.StatusLocked, gin.H{"error": err.Error()})
			return
		}
		if strings.Contains(err.Error(), "2FA") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error(), "requires_2fa": true})
			return
		}
		if strings.Contains(err.Error(), "deactivated") {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Login failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":          user,
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"expires_at":    tokens.ExpiresAt.Unix(),
		"token_type":    tokens.TokenType,
	})
}

func (h *UserHandler) Logout(c *gin.Context) {
	userID := h.getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	token := h.getTokenFromHeader(c)
	if err := h.userService.Logout(c.Request.Context(), userID, token); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Logout failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

func (h *UserHandler) Refresh(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tokens, err := h.userService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"expires_at":    tokens.ExpiresAt.Unix(),
		"token_type":    tokens.TokenType,
	})
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := h.getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	user, err := h.userService.GetProfile(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID := h.getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var update domain.User
	if err := c.ShouldBindJSON(&update); err != nil {
		log.Printf("❌ [UpdateProfile] Error binding JSON for user %d: %v", userID, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Log the received data
	log.Printf("🔧 [UpdateProfile] Received update request for user %d: %+v", userID, update)

	user, err := h.userService.UpdateProfile(c.Request.Context(), userID, &update)
	if err != nil {
		log.Printf("❌ [UpdateProfile] Error updating profile for user %d: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	log.Printf("✅ [UpdateProfile] Successfully updated profile for user %d", userID)
	c.JSON(http.StatusOK, gin.H{"user": user})
}

func (h *UserHandler) GetPreferences(c *gin.Context) {
	userID := h.getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	prefs, err := h.userService.GetPreferences(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Preferences not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"preferences": prefs})
}

func (h *UserHandler) UpdatePreferences(c *gin.Context) {
	userID := h.getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var prefs domain.Preferences
	if err := c.ShouldBindJSON(&prefs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.userService.UpdatePreferences(c.Request.Context(), userID, &prefs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update preferences"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Preferences updated successfully"})
}

func (h *UserHandler) GetNotifications(c *gin.Context) {
	userID := h.getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	notifs, err := h.userService.GetNotifications(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Notification settings not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"notifications": notifs})
}

func (h *UserHandler) UpdateNotifications(c *gin.Context) {
	userID := h.getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var notifs domain.NotificationSettings
	if err := c.ShouldBindJSON(&notifs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.userService.UpdateNotifications(c.Request.Context(), userID, &notifs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update notifications"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification settings updated successfully"})
}

func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID := h.getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req domain.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.userService.ChangePassword(c.Request.Context(), userID, &req); err != nil {
		if strings.Contains(err.Error(), "incorrect") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		if strings.Contains(err.Error(), "password") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to change password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}

func (h *UserHandler) Setup2FA(c *gin.Context) {
	userID := h.getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	setup, err := h.userService.Setup2FA(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to setup 2FA"})
		return
	}

	// Generate QR code image as base64
	qrCodeBytes, err := h.userService.GenerateQRCode(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate QR code"})
		return
	}

	// Convert to base64
	qrCodeBase64 := base64.StdEncoding.EncodeToString(qrCodeBytes)

	c.JSON(http.StatusOK, gin.H{
		"secret":        setup.Secret,
		"qr_code_url":   setup.QRCodeURL,
		"qr_code_image": qrCodeBase64,
		"backup_codes":  setup.BackupCodes,
	})
}

func (h *UserHandler) Enable2FA(c *gin.Context) {
	userID := h.getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req struct {
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.userService.Enable2FA(c.Request.Context(), userID, req.Code); err != nil {
		if strings.Contains(err.Error(), "invalid") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to enable 2FA"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "2FA enabled successfully"})
}

func (h *UserHandler) Disable2FA(c *gin.Context) {
	userID := h.getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req struct {
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.userService.Disable2FA(c.Request.Context(), userID, req.Password); err != nil {
		if strings.Contains(err.Error(), "incorrect") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to disable 2FA"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "2FA disabled successfully"})
}

func (h *UserHandler) Verify2FA(c *gin.Context) {
	userID := h.getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req struct {
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.userService.Verify2FA(c.Request.Context(), userID, req.Code); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "2FA verification successful"})
}

func (h *UserHandler) Check2FA(c *gin.Context) {
	var req domain.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Intentar hacer login para verificar credenciales
	user, _, err := h.userService.Login(c.Request.Context(), &req)
	if err != nil {
		if strings.Contains(err.Error(), "2FA code required") {
			// Si requiere 2FA, las credenciales son válidas
			c.JSON(http.StatusOK, gin.H{
				"requires_2fa": true,
				"user_id": 0, // No tenemos el user_id aquí, pero no es necesario
			})
			return
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Si llegamos aquí, el login fue exitoso sin 2FA
	c.JSON(http.StatusOK, gin.H{
		"requires_2fa": false,
		"user_id": user.ID,
	})
}

func (h *UserHandler) VerifyEmail(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token is required"})
		return
	}

	if err := h.userService.VerifyEmailWithToken(c.Request.Context(), token); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Email verified successfully"})
}

func (h *UserHandler) RequestPasswordReset(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.userService.RequestPasswordReset(c.Request.Context(), req.Email); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send reset email"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password reset email sent"})
}

func (h *UserHandler) ResetPassword(c *gin.Context) {
	var req struct {
		Token       string `json:"token" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.userService.ResetPassword(c.Request.Context(), req.Token, req.NewPassword); err != nil {
		if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "expired") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
}

func (h *UserHandler) ExportData(c *gin.Context) {
	userID := h.getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	data, err := h.userService.ExportData(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to export data"})
		return
	}

	c.Header("Content-Type", "text/plain")
	c.Header("Content-Disposition", "attachment; filename=user_data.txt")
	c.String(http.StatusOK, data)
}

func (h *UserHandler) DeleteAccount(c *gin.Context) {
	userID := h.getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req struct {
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.userService.DeleteAccount(c.Request.Context(), userID, req.Password); err != nil {
		if strings.Contains(err.Error(), "incorrect") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete account"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Account deleted successfully"})
}

func (h *UserHandler) UploadAvatar(c *gin.Context) {
	userID := h.getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	log.Printf("🔧 [UploadAvatar] Starting avatar upload for user %d", userID)

	file, err := c.FormFile("avatar")
	if err != nil {
		log.Printf("❌ [UploadAvatar] Failed to get file for user %d: %v", userID, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get file"})
		return
	}

	log.Printf("🔧 [UploadAvatar] File received: %s, size: %d bytes (%.1fKB)",
		file.Filename, file.Size, float64(file.Size)/1024)

	// Verificar límite de archivo inicial (antes de procesamiento)
	if file.Size > 10*1024*1024 { // 10MB límite inicial para archivos sin procesar
		c.JSON(http.StatusBadRequest, gin.H{"error": "File too large (maximum 10MB)"})
		return
	}

	// Verificar formato de archivo
	if !imageprocessor.IsValidImageFormat(file.Filename) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type. Supported: JPG, PNG, GIF"})
		return
	}

	// Generar nombre de archivo (siempre será .jpeg después del procesamiento)
	filename := fmt.Sprintf("%d_%d.jpeg", userID, time.Now().Unix())
	uploadPath := filepath.Join("uploads", filename)

	// Path web para guardar en la base de datos (debe empezar con /)
	webPath := fmt.Sprintf("/uploads/%s", filename)

	log.Printf("🔧 [UploadAvatar] Processing image with compression...")

	// Abrir el archivo subido
	fileReader, err := file.Open()
	if err != nil {
		log.Printf("❌ [UploadAvatar] Failed to open uploaded file for user %d: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process file"})
		return
	}
	defer fileReader.Close()

	// Procesar la imagen (redimensionar y comprimir)
	if err := imageprocessor.ProcessUploadedImage(fileReader, uploadPath, imageprocessor.DefaultAvatarConfig); err != nil {
		log.Printf("❌ [UploadAvatar] Failed to process image for user %d: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process image"})
		return
	}

	log.Printf("✅ [UploadAvatar] Image processed and saved successfully at: %s", uploadPath)

	if err := h.userService.UpdateAvatar(c.Request.Context(), userID, webPath); err != nil {
		log.Printf("❌ [UploadAvatar] Failed to update avatar in database for user %d: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update avatar"})
		return
	}

	log.Printf("✅ [UploadAvatar] Avatar updated successfully for user %d: %s", userID, webPath)
	c.JSON(http.StatusOK, gin.H{"message": "Avatar uploaded successfully", "avatar_url": webPath})
}

func (h *UserHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   "users-service",
		"timestamp": gin.H{"now": "ok"},
	})
}

// Helper methods
func (h *UserHandler) getUserID(c *gin.Context) uint {
	userIDValue, exists := c.Get("user_id")
	if !exists {
		return 0
	}

	switch v := userIDValue.(type) {
	case uint:
		return v
	case float64:
		return uint(v)
	case string:
		if id, err := strconv.ParseUint(v, 10, 32); err == nil {
			return uint(id)
		}
	}

	return 0
}

func (h *UserHandler) getTokenFromHeader(c *gin.Context) string {
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
