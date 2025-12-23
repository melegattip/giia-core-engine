package auth

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"

	pkgErrors "github.com/giia/giia-core-engine/pkg/errors"
	pkgLogger "github.com/giia/giia-core-engine/pkg/logger"
	"github.com/giia/giia-core-engine/services/auth-service/internal/core/domain"
	"github.com/giia/giia-core-engine/services/auth-service/internal/core/providers"
)

func TestConfirmPasswordResetUseCase_Execute_Success(t *testing.T) {
	// Given
	mockUserRepo := new(providers.MockUserRepository)
	mockTokenRepo := new(providers.MockTokenRepository)
	logger := pkgLogger.New("test", "error")

	useCase := NewConfirmPasswordResetUseCase(
		mockUserRepo,
		mockTokenRepo,
		logger,
	)

	ctx := context.Background()
	token := "reset-token-123"
	newPassword := "NewSecureP@ss123"
	userID := uuid.New()

	givenResetToken := &domain.PasswordResetToken{
		TokenHash: hashActivationToken(token),
		UserID:    userID,
		ExpiresAt: time.Now().Add(1 * time.Hour),
		Used:      false,
	}

	givenUser := &domain.User{
		ID:       userID,
		Email:    "test@example.com",
		Password: "old-hashed-password",
	}

	mockTokenRepo.On("GetPasswordResetToken", ctx, hashActivationToken(token)).Return(givenResetToken, nil)
	mockUserRepo.On("GetByID", ctx, userID).Return(givenUser, nil)
	mockUserRepo.On("Update", ctx, mock.AnythingOfType("*domain.User")).Return(nil)
	mockTokenRepo.On("MarkPasswordResetTokenUsed", ctx, hashActivationToken(token)).Return(nil)

	// When
	err := useCase.Execute(ctx, token, newPassword)

	// Then
	assert.NoError(t, err)
	mockTokenRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)

	// Verify password was updated with bcrypt
	mockUserRepo.AssertCalled(t, "Update", ctx, mock.MatchedBy(func(user *domain.User) bool {
		err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(newPassword))
		return err == nil
	}))
}

func TestConfirmPasswordResetUseCase_Execute_EmptyToken_ReturnsError(t *testing.T) {
	// Given
	mockUserRepo := new(providers.MockUserRepository)
	mockTokenRepo := new(providers.MockTokenRepository)
	logger := pkgLogger.New("test", "error")

	useCase := NewConfirmPasswordResetUseCase(
		mockUserRepo,
		mockTokenRepo,
		logger,
	)

	ctx := context.Background()
	token := ""
	newPassword := "NewSecureP@ss123"

	// When
	err := useCase.Execute(ctx, token, newPassword)

	// Then
	assert.Error(t, err)
	assert.Equal(t, "reset token is required", err.(*pkgErrors.CustomError).Message)
	mockTokenRepo.AssertNotCalled(t, "GetPasswordResetToken")
}

func TestConfirmPasswordResetUseCase_Execute_EmptyPassword_ReturnsError(t *testing.T) {
	// Given
	mockUserRepo := new(providers.MockUserRepository)
	mockTokenRepo := new(providers.MockTokenRepository)
	logger := pkgLogger.New("test", "error")

	useCase := NewConfirmPasswordResetUseCase(
		mockUserRepo,
		mockTokenRepo,
		logger,
	)

	ctx := context.Background()
	token := "reset-token-123"
	newPassword := ""

	// When
	err := useCase.Execute(ctx, token, newPassword)

	// Then
	assert.Error(t, err)
	assert.Equal(t, "new password is required", err.(*pkgErrors.CustomError).Message)
	mockTokenRepo.AssertNotCalled(t, "GetPasswordResetToken")
}

func TestConfirmPasswordResetUseCase_Execute_WeakPassword_ReturnsError(t *testing.T) {
	testCases := []struct {
		name            string
		givenPassword   string
		expectedMessage string
	}{
		{
			name:            "Too short",
			givenPassword:   "Short1!",
			expectedMessage: "password must be at least 8 characters long",
		},
		{
			name:            "No uppercase",
			givenPassword:   "nouppercase1!",
			expectedMessage: "password must contain at least one uppercase letter",
		},
		{
			name:            "No lowercase",
			givenPassword:   "NOLOWERCASE1!",
			expectedMessage: "password must contain at least one lowercase letter",
		},
		{
			name:            "No number",
			givenPassword:   "NoNumber!",
			expectedMessage: "password must contain at least one number",
		},
		{
			name:            "No special character",
			givenPassword:   "NoSpecial1",
			expectedMessage: "password must contain at least one special character",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			mockUserRepo := new(providers.MockUserRepository)
			mockTokenRepo := new(providers.MockTokenRepository)
			logger := pkgLogger.New("test", "error")

			useCase := NewConfirmPasswordResetUseCase(
				mockUserRepo,
				mockTokenRepo,
				logger,
			)

			ctx := context.Background()
			token := "reset-token-123"

			// When
			err := useCase.Execute(ctx, token, tc.givenPassword)

			// Then
			assert.Error(t, err)
			assert.Equal(t, tc.expectedMessage, err.(*pkgErrors.CustomError).Message)
			mockTokenRepo.AssertNotCalled(t, "GetPasswordResetToken")
		})
	}
}

func TestConfirmPasswordResetUseCase_Execute_InvalidToken_ReturnsError(t *testing.T) {
	// Given
	mockUserRepo := new(providers.MockUserRepository)
	mockTokenRepo := new(providers.MockTokenRepository)
	logger := pkgLogger.New("test", "error")

	useCase := NewConfirmPasswordResetUseCase(
		mockUserRepo,
		mockTokenRepo,
		logger,
	)

	ctx := context.Background()
	token := "invalid-token"
	newPassword := "NewSecureP@ss123"

	mockTokenRepo.On("GetPasswordResetToken", ctx, hashActivationToken(token)).
		Return(nil, pkgErrors.NewNotFound("token not found"))

	// When
	err := useCase.Execute(ctx, token, newPassword)

	// Then
	assert.Error(t, err)
	assert.Equal(t, "invalid or expired reset token", err.(*pkgErrors.CustomError).Message)
	mockTokenRepo.AssertExpectations(t)
	mockUserRepo.AssertNotCalled(t, "GetByID")
}

func TestConfirmPasswordResetUseCase_Execute_AlreadyUsedToken_ReturnsError(t *testing.T) {
	// Given
	mockUserRepo := new(providers.MockUserRepository)
	mockTokenRepo := new(providers.MockTokenRepository)
	logger := pkgLogger.New("test", "error")

	useCase := NewConfirmPasswordResetUseCase(
		mockUserRepo,
		mockTokenRepo,
		logger,
	)

	ctx := context.Background()
	token := "used-token"
	newPassword := "NewSecureP@ss123"
	userID := uuid.New()

	givenResetToken := &domain.PasswordResetToken{
		TokenHash: hashActivationToken(token),
		UserID:    userID,
		ExpiresAt: time.Now().Add(1 * time.Hour),
		Used:      true, // Token already used
	}

	mockTokenRepo.On("GetPasswordResetToken", ctx, hashActivationToken(token)).Return(givenResetToken, nil)

	// When
	err := useCase.Execute(ctx, token, newPassword)

	// Then
	assert.Error(t, err)
	assert.Equal(t, "reset token has already been used", err.(*pkgErrors.CustomError).Message)
	mockTokenRepo.AssertExpectations(t)
	mockUserRepo.AssertNotCalled(t, "GetByID")
}

func TestConfirmPasswordResetUseCase_Execute_UserNotFound_ReturnsError(t *testing.T) {
	// Given
	mockUserRepo := new(providers.MockUserRepository)
	mockTokenRepo := new(providers.MockTokenRepository)
	logger := pkgLogger.New("test", "error")

	useCase := NewConfirmPasswordResetUseCase(
		mockUserRepo,
		mockTokenRepo,
		logger,
	)

	ctx := context.Background()
	token := "reset-token-123"
	newPassword := "NewSecureP@ss123"
	userID := uuid.New()

	givenResetToken := &domain.PasswordResetToken{
		TokenHash: hashActivationToken(token),
		UserID:    userID,
		ExpiresAt: time.Now().Add(1 * time.Hour),
		Used:      false,
	}

	mockTokenRepo.On("GetPasswordResetToken", ctx, hashActivationToken(token)).Return(givenResetToken, nil)
	mockUserRepo.On("GetByID", ctx, userID).Return(nil, pkgErrors.NewNotFound("user not found"))

	// When
	err := useCase.Execute(ctx, token, newPassword)

	// Then
	assert.Error(t, err)
	assert.Equal(t, "failed to reset password", err.(*pkgErrors.CustomError).Message)
	mockTokenRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestConfirmPasswordResetUseCase_Execute_UpdateUserFails_ReturnsError(t *testing.T) {
	// Given
	mockUserRepo := new(providers.MockUserRepository)
	mockTokenRepo := new(providers.MockTokenRepository)
	logger := pkgLogger.New("test", "error")

	useCase := NewConfirmPasswordResetUseCase(
		mockUserRepo,
		mockTokenRepo,
		logger,
	)

	ctx := context.Background()
	token := "reset-token-123"
	newPassword := "NewSecureP@ss123"
	userID := uuid.New()

	givenResetToken := &domain.PasswordResetToken{
		TokenHash: hashActivationToken(token),
		UserID:    userID,
		ExpiresAt: time.Now().Add(1 * time.Hour),
		Used:      false,
	}

	givenUser := &domain.User{
		ID:       userID,
		Email:    "test@example.com",
		Password: "old-hashed-password",
	}

	mockTokenRepo.On("GetPasswordResetToken", ctx, hashActivationToken(token)).Return(givenResetToken, nil)
	mockUserRepo.On("GetByID", ctx, userID).Return(givenUser, nil)
	mockUserRepo.On("Update", ctx, mock.AnythingOfType("*domain.User")).
		Return(pkgErrors.NewInternalServerError("database error"))

	// When
	err := useCase.Execute(ctx, token, newPassword)

	// Then
	assert.Error(t, err)
	assert.Equal(t, "failed to reset password", err.(*pkgErrors.CustomError).Message)
	mockTokenRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
	mockTokenRepo.AssertNotCalled(t, "MarkPasswordResetTokenUsed")
}

func TestConfirmPasswordResetUseCase_Execute_MarkTokenUsedFails_StillReturnsSuccess(t *testing.T) {
	// Given
	mockUserRepo := new(providers.MockUserRepository)
	mockTokenRepo := new(providers.MockTokenRepository)
	logger := pkgLogger.New("test", "error")

	useCase := NewConfirmPasswordResetUseCase(
		mockUserRepo,
		mockTokenRepo,
		logger,
	)

	ctx := context.Background()
	token := "reset-token-123"
	newPassword := "NewSecureP@ss123"
	userID := uuid.New()

	givenResetToken := &domain.PasswordResetToken{
		TokenHash: hashActivationToken(token),
		UserID:    userID,
		ExpiresAt: time.Now().Add(1 * time.Hour),
		Used:      false,
	}

	givenUser := &domain.User{
		ID:       userID,
		Email:    "test@example.com",
		Password: "old-hashed-password",
	}

	mockTokenRepo.On("GetPasswordResetToken", ctx, hashActivationToken(token)).Return(givenResetToken, nil)
	mockUserRepo.On("GetByID", ctx, userID).Return(givenUser, nil)
	mockUserRepo.On("Update", ctx, mock.AnythingOfType("*domain.User")).Return(nil)
	mockTokenRepo.On("MarkPasswordResetTokenUsed", ctx, hashActivationToken(token)).
		Return(pkgErrors.NewInternalServerError("database error"))

	// When
	err := useCase.Execute(ctx, token, newPassword)

	// Then - Password was reset, marking token as used is logged but not critical
	assert.NoError(t, err)
	mockTokenRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}
