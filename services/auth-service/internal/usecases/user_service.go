package usecases

import (
	"context"
	"fmt"
	"log"
	"time"

	"users-service/internal/domain"
	"users-service/internal/infrastructure/auth"
	"users-service/internal/repository"
)

type UserService interface {
	Register(ctx context.Context, req *domain.RegisterRequest) (*domain.UserResponse, *auth.TokenPair, error)
	Login(ctx context.Context, req *domain.LoginRequest) (*domain.UserResponse, *auth.TokenPair, error)
	Logout(ctx context.Context, userID uint, token string) error
	RefreshToken(ctx context.Context, refreshToken string) (*auth.TokenPair, error)
	GetProfile(ctx context.Context, userID uint) (*domain.UserResponse, error)
	UpdateProfile(ctx context.Context, userID uint, update *domain.User) (*domain.UserResponse, error)
	UpdateAvatar(ctx context.Context, userID uint, avatarPath string) error
	GetPreferences(ctx context.Context, userID uint) (*domain.Preferences, error)
	UpdatePreferences(ctx context.Context, userID uint, prefs *domain.Preferences) error
	GetNotifications(ctx context.Context, userID uint) (*domain.NotificationSettings, error)
	UpdateNotifications(ctx context.Context, userID uint, notif *domain.NotificationSettings) error
	ChangePassword(ctx context.Context, userID uint, req *domain.ChangePasswordRequest) error
	Setup2FA(ctx context.Context, userID uint) (*auth.TwoFASetup, error)
	GenerateQRCode(ctx context.Context, userID uint) ([]byte, error)
	Enable2FA(ctx context.Context, userID uint, code string) error
	Disable2FA(ctx context.Context, userID uint, password string) error
	Verify2FA(ctx context.Context, userID uint, code string) error
	VerifyEmailWithToken(ctx context.Context, token string) error
	RequestPasswordReset(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token, newPassword string) error
	ExportData(ctx context.Context, userID uint) (string, error)
	DeleteAccount(ctx context.Context, userID uint, password string) error
}

type userService struct {
	repo             repository.UserRepository
	jwtService       auth.JWTService
	passwordService  auth.PasswordService
	twoFAService     auth.TwoFAService
	maxLoginAttempts int
	lockoutDuration  time.Duration
}

func NewUserService(
	repo repository.UserRepository,
	jwtService auth.JWTService,
	passwordService auth.PasswordService,
	twoFAService auth.TwoFAService,
	maxLoginAttempts int,
	lockoutDuration time.Duration,
) UserService {
	return &userService{
		repo:             repo,
		jwtService:       jwtService,
		passwordService:  passwordService,
		twoFAService:     twoFAService,
		maxLoginAttempts: maxLoginAttempts,
		lockoutDuration:  lockoutDuration,
	}
}

func (s *userService) Register(ctx context.Context, req *domain.RegisterRequest) (*domain.UserResponse, *auth.TokenPair, error) {
	// Check if user already exists
	if existingUser, _ := s.repo.GetByEmail(ctx, req.Email); existingUser != nil {
		return nil, nil, fmt.Errorf("user with email %s already exists", req.Email)
	}

	// Hash password
	hashedPassword, err := s.passwordService.HashPassword(req.Password)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &domain.User{
		Email:      req.Email,
		Password:   hashedPassword,
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		Phone:      req.Phone,
		IsActive:   true,
		IsVerified: false, // Will be verified via email
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate email verification token
	verificationToken, err := s.jwtService.GenerateEmailVerificationToken(user.ID, user.Email)
	if err != nil {
		log.Printf("Warning: Failed to generate email verification token: %v", err)
	} else {
		expires := time.Now().Add(24 * time.Hour)
		if err := s.repo.SetEmailVerificationToken(ctx, user.ID, verificationToken, expires); err != nil {
			log.Printf("Warning: Failed to set email verification token: %v", err)
		}
	}

	// Generate JWT tokens
	tokens, err := s.jwtService.GenerateTokens(user.ID, user.Email)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Convert to response format
	userResponse := &domain.UserResponse{
		ID:         user.ID,
		Email:      user.Email,
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		Phone:      user.Phone,
		IsActive:   user.IsActive,
		IsVerified: user.IsVerified,
		CreatedAt:  user.CreatedAt,
	}

	return userResponse, tokens, nil
}

func (s *userService) Login(ctx context.Context, req *domain.LoginRequest) (*domain.UserResponse, *auth.TokenPair, error) {
	log.Printf("🔍 [Login] Attempting login for email: %s", req.Email)

	// Get user by email
	user, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		log.Printf("❌ [Login] User not found for email: %s, error: %v", req.Email, err)
		return nil, nil, fmt.Errorf("invalid email or password")
	}

	log.Printf("✅ [Login] User found: ID=%d, Email=%s, IsActive=%v, IsVerified=%v",
		user.ID, user.Email, user.IsActive, user.IsVerified)

	// Check if account is active
	if !user.IsActive {
		return nil, nil, fmt.Errorf("account is deactivated")
	}

	// Check if account is locked
	if user.LockedUntil != nil && user.LockedUntil.After(time.Now()) {
		return nil, nil, fmt.Errorf("account is locked until %v", user.LockedUntil.Format(time.RFC3339))
	}

	// Verify password
	log.Printf("🔍 [Login] Verifying password for user: %s", user.Email)
	if err := s.passwordService.VerifyPassword(user.Password, req.Password); err != nil {
		log.Printf("❌ [Login] Password verification failed for user: %s, error: %v", user.Email, err)

		// Increment failed login attempts
		if err := s.repo.IncrementFailedLoginAttempts(ctx, user.ID); err != nil {
			log.Printf("Warning: Failed to increment login attempts: %v", err)
		}

		// Check if we should lock the account
		if user.FailedLoginAttempts+1 >= s.maxLoginAttempts {
			lockUntil := time.Now().Add(s.lockoutDuration)
			if err := s.repo.SetAccountLocked(ctx, user.ID, lockUntil); err != nil {
				log.Printf("Warning: Failed to lock account: %v", err)
			}
			return nil, nil, fmt.Errorf("account locked due to too many failed login attempts")
		}

		return nil, nil, fmt.Errorf("invalid email or password")
	}

	log.Printf("✅ [Login] Password verification successful for user: %s", user.Email)

	// Check 2FA if enabled
	if req.TwoFACode != "" {
		twoFA, err := s.repo.Get2FA(ctx, user.ID)
		if err == nil && twoFA.Enabled {
			if !s.twoFAService.ValidateCode(twoFA.Secret, req.TwoFACode) {
				// Check backup codes
				newBackupCodes, isBackupCode := s.twoFAService.ValidateBackupCode(twoFA.BackupCodes, req.TwoFACode)
				if !isBackupCode {
					return nil, nil, fmt.Errorf("invalid 2FA code")
				}
				// Update backup codes if a backup code was used
				twoFA.BackupCodes = newBackupCodes
				twoFA.LastUsedCode = req.TwoFACode
				if err := s.repo.Update2FA(ctx, twoFA); err != nil {
					log.Printf("Warning: Failed to update 2FA backup codes: %v", err)
				}
			}
		}
	} else {
		// Check if 2FA is required
		if twoFA, err := s.repo.Get2FA(ctx, user.ID); err == nil && twoFA.Enabled {
			return nil, nil, fmt.Errorf("2FA code required")
		}
	}

	// Reset failed login attempts on successful login
	if err := s.repo.ResetFailedLoginAttempts(ctx, user.ID); err != nil {
		log.Printf("Warning: Failed to reset login attempts: %v", err)
	}

	// Update last login time
	if err := s.repo.UpdateLastLogin(ctx, user.ID); err != nil {
		log.Printf("Warning: Failed to update last login: %v", err)
	}

	// Generate JWT tokens
	tokens, err := s.jwtService.GenerateTokens(user.ID, user.Email)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Convert to response format
	userResponse := &domain.UserResponse{
		ID:         user.ID,
		Email:      user.Email,
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		Phone:      user.Phone,
		IsActive:   user.IsActive,
		IsVerified: user.IsVerified,
		LastLogin:  user.LastLogin,
		CreatedAt:  user.CreatedAt,
	}

	return userResponse, tokens, nil
}

func (s *userService) Logout(ctx context.Context, userID uint, token string) error {
	// In a production system, you would add the token to a blacklist
	// For now, we'll just log the logout
	log.Printf("User %d logged out", userID)
	return nil
}

func (s *userService) RefreshToken(ctx context.Context, refreshToken string) (*auth.TokenPair, error) {
	// Validate refresh token
	claims, err := s.jwtService.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Check if user still exists and is active
	user, err := s.repo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	if !user.IsActive {
		return nil, fmt.Errorf("user account is deactivated")
	}

	// Generate new tokens
	tokens, err := s.jwtService.GenerateTokens(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	return tokens, nil
}

func (s *userService) GetProfile(ctx context.Context, userID uint) (*domain.UserResponse, error) {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	return &domain.UserResponse{
		ID:         user.ID,
		Email:      user.Email,
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		Phone:      user.Phone,
		Avatar:     user.Avatar,
		IsActive:   user.IsActive,
		IsVerified: user.IsVerified,
		LastLogin:  user.LastLogin,
		CreatedAt:  user.CreatedAt,
	}, nil
}

func (s *userService) UpdateProfile(ctx context.Context, userID uint, update *domain.User) (*domain.UserResponse, error) {
	// Get current user
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Update fields
	user.FirstName = update.FirstName
	user.LastName = update.LastName
	user.Phone = update.Phone

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return &domain.UserResponse{
		ID:         user.ID,
		Email:      user.Email,
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		Phone:      user.Phone,
		Avatar:     user.Avatar,
		IsActive:   user.IsActive,
		IsVerified: user.IsVerified,
		LastLogin:  user.LastLogin,
		CreatedAt:  user.CreatedAt,
	}, nil
}

func (s *userService) UpdateAvatar(ctx context.Context, userID uint, avatarPath string) error {
	log.Printf("🔧 [UpdateAvatar] Updating avatar for user %d to: %s", userID, avatarPath)

	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		log.Printf("❌ [UpdateAvatar] Failed to get user %d: %v", userID, err)
		return fmt.Errorf("failed to get user: %w", err)
	}

	log.Printf("🔧 [UpdateAvatar] User found: %d, current avatar: %s", user.ID, user.Avatar)

	user.Avatar = avatarPath
	if err := s.repo.Update(ctx, user); err != nil {
		log.Printf("❌ [UpdateAvatar] Failed to update user %d in database: %v", userID, err)
		return fmt.Errorf("failed to update avatar: %w", err)
	}

	log.Printf("✅ [UpdateAvatar] Avatar updated successfully for user %d", userID)
	return nil
}

func (s *userService) GetPreferences(ctx context.Context, userID uint) (*domain.Preferences, error) {
	return s.repo.GetPreferences(ctx, userID)
}

func (s *userService) UpdatePreferences(ctx context.Context, userID uint, prefs *domain.Preferences) error {
	prefs.UserID = userID
	return s.repo.UpdatePreferences(ctx, prefs)
}

func (s *userService) GetNotifications(ctx context.Context, userID uint) (*domain.NotificationSettings, error) {
	return s.repo.GetNotifications(ctx, userID)
}

func (s *userService) UpdateNotifications(ctx context.Context, userID uint, notif *domain.NotificationSettings) error {
	notif.UserID = userID
	return s.repo.UpdateNotifications(ctx, notif)
}

func (s *userService) ChangePassword(ctx context.Context, userID uint, req *domain.ChangePasswordRequest) error {
	// Get current user
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Verify current password
	if err := s.passwordService.VerifyPassword(user.Password, req.CurrentPassword); err != nil {
		return fmt.Errorf("current password is incorrect")
	}

	// Hash new password
	hashedPassword, err := s.passwordService.HashPassword(req.NewPassword)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	// Update password
	if err := s.repo.UpdatePassword(ctx, userID, hashedPassword); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

func (s *userService) Setup2FA(ctx context.Context, userID uint) (*auth.TwoFASetup, error) {
	// Get user
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Generate 2FA setup
	setup, err := s.twoFAService.GenerateSecret(user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate 2FA secret: %w", err)
	}

	// Save 2FA secret (disabled until user verifies)
	twoFA := &domain.TwoFA{
		UserID:      userID,
		Secret:      setup.Secret,
		Enabled:     false,
		BackupCodes: setup.BackupCodes,
	}

	if err := s.repo.Update2FA(ctx, twoFA); err != nil {
		return nil, fmt.Errorf("failed to save 2FA setup: %w", err)
	}

	return setup, nil
}

func (s *userService) GenerateQRCode(ctx context.Context, userID uint) ([]byte, error) {
	// Get user
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Get 2FA settings
	twoFA, err := s.repo.Get2FA(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("2FA not set up: %w", err)
	}

	// Generate QR code image
	qrCodeBytes, err := s.twoFAService.GenerateQRCode(twoFA.Secret, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate QR code: %w", err)
	}

	return qrCodeBytes, nil
}

func (s *userService) Enable2FA(ctx context.Context, userID uint, code string) error {
	// Get 2FA settings
	twoFA, err := s.repo.Get2FA(ctx, userID)
	if err != nil {
		return fmt.Errorf("2FA not set up: %w", err)
	}

	// Validate code
	if !s.twoFAService.ValidateCode(twoFA.Secret, code) {
		return fmt.Errorf("invalid verification code")
	}

	// Enable 2FA
	twoFA.Enabled = true
	if err := s.repo.Update2FA(ctx, twoFA); err != nil {
		return fmt.Errorf("failed to enable 2FA: %w", err)
	}

	return nil
}

func (s *userService) Disable2FA(ctx context.Context, userID uint, password string) error {
	// Get user and verify password
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if err := s.passwordService.VerifyPassword(user.Password, password); err != nil {
		return fmt.Errorf("password is incorrect")
	}

	// Get 2FA settings
	twoFA, err := s.repo.Get2FA(ctx, userID)
	if err != nil {
		return fmt.Errorf("2FA not found: %w", err)
	}

	// Disable 2FA
	twoFA.Enabled = false
	if err := s.repo.Update2FA(ctx, twoFA); err != nil {
		return fmt.Errorf("failed to disable 2FA: %w", err)
	}

	return nil
}

func (s *userService) Verify2FA(ctx context.Context, userID uint, code string) error {
	twoFA, err := s.repo.Get2FA(ctx, userID)
	if err != nil {
		return fmt.Errorf("2FA not found: %w", err)
	}

	if !twoFA.Enabled {
		return fmt.Errorf("2FA is not enabled")
	}

	if !s.twoFAService.ValidateCode(twoFA.Secret, code) {
		// Check backup codes
		newBackupCodes, isBackupCode := s.twoFAService.ValidateBackupCode(twoFA.BackupCodes, code)
		if !isBackupCode {
			return fmt.Errorf("invalid 2FA code")
		}
		// Update backup codes
		twoFA.BackupCodes = newBackupCodes
		twoFA.LastUsedCode = code
		if err := s.repo.Update2FA(ctx, twoFA); err != nil {
			log.Printf("Warning: Failed to update 2FA backup codes: %v", err)
		}
	}

	return nil
}

func (s *userService) VerifyEmailWithToken(ctx context.Context, token string) error {
	// Validate token
	claims, err := s.jwtService.ValidateEmailVerificationToken(token)
	if err != nil {
		return fmt.Errorf("invalid verification token: %w", err)
	}

	// Get user by token (also checks expiry)
	user, err := s.repo.GetByEmailVerificationToken(ctx, token)
	if err != nil {
		return fmt.Errorf("invalid or expired verification token: %w", err)
	}

	// Verify user ID matches
	if user.ID != claims.UserID {
		return fmt.Errorf("token user ID mismatch")
	}

	// Set email as verified
	if err := s.repo.SetEmailVerified(ctx, user.ID, true); err != nil {
		return fmt.Errorf("failed to verify email: %w", err)
	}

	return nil
}

func (s *userService) RequestPasswordReset(ctx context.Context, email string) error {
	// Check if user exists
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		// Don't reveal if email exists or not for security
		return nil
	}

	// Generate password reset token
	resetToken, err := s.jwtService.GeneratePasswordResetToken(user.ID, user.Email)
	if err != nil {
		return fmt.Errorf("failed to generate reset token: %w", err)
	}

	// Set reset token with 1-hour expiry
	expires := time.Now().Add(1 * time.Hour)
	if err := s.repo.SetPasswordResetToken(ctx, user.ID, resetToken, expires); err != nil {
		return fmt.Errorf("failed to set reset token: %w", err)
	}

	// In a real implementation, you would send an email here
	log.Printf("Password reset token for user %s: %s", email, resetToken)

	return nil
}

func (s *userService) ResetPassword(ctx context.Context, token, newPassword string) error {
	// Validate token
	claims, err := s.jwtService.ValidatePasswordResetToken(token)
	if err != nil {
		return fmt.Errorf("invalid reset token: %w", err)
	}

	// Get user by token (also checks expiry)
	user, err := s.repo.GetByPasswordResetToken(ctx, token)
	if err != nil {
		return fmt.Errorf("invalid or expired reset token: %w", err)
	}

	// Verify user ID matches
	if user.ID != claims.UserID {
		return fmt.Errorf("token user ID mismatch")
	}

	// Hash new password
	hashedPassword, err := s.passwordService.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password and clear reset token
	if err := s.repo.UpdatePassword(ctx, user.ID, hashedPassword); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Clear reset token
	if err := s.repo.SetPasswordResetToken(ctx, user.ID, "", time.Now()); err != nil {
		log.Printf("Warning: Failed to clear reset token: %v", err)
	}

	return nil
}

func (s *userService) ExportData(ctx context.Context, userID uint) (string, error) {
	// Get user data
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("failed to get user: %w", err)
	}

	// Get preferences
	prefs, _ := s.repo.GetPreferences(ctx, userID)

	// Get notification settings
	notifs, _ := s.repo.GetNotifications(ctx, userID)

	// In a real implementation, you would format this as JSON/CSV
	data := fmt.Sprintf("User Data Export for %s\n", user.Email)
	data += fmt.Sprintf("Name: %s %s\n", user.FirstName, user.LastName)
	data += fmt.Sprintf("Phone: %s\n", user.Phone)
	data += fmt.Sprintf("Created: %s\n", user.CreatedAt.Format(time.RFC3339))

	if prefs != nil {
		data += fmt.Sprintf("Currency: %s\n", prefs.Currency)
		data += fmt.Sprintf("Language: %s\n", prefs.Language)
		data += fmt.Sprintf("Theme: %s\n", prefs.Theme)
	}

	if notifs != nil {
		data += fmt.Sprintf("Email Notifications: %t\n", notifs.EmailNotifications)
		data += fmt.Sprintf("Push Notifications: %t\n", notifs.PushNotifications)
	}

	return data, nil
}

func (s *userService) DeleteAccount(ctx context.Context, userID uint, password string) error {
	// Get user and verify password
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if err := s.passwordService.VerifyPassword(user.Password, password); err != nil {
		return fmt.Errorf("password is incorrect")
	}

	// Delete user (cascade will handle related data)
	if err := s.repo.Delete(ctx, userID); err != nil {
		return fmt.Errorf("failed to delete account: %w", err)
	}

	log.Printf("Account deleted for user: %s", user.Email)
	return nil
}
