package auth

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"

	"github.com/melegattip/giia-core-engine/services/auth-service/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/auth-service/internal/core/providers"
)

func TestLoginUseCase_Execute_WithValidCredentials_ReturnsTokens(t *testing.T) {
	// Given
	givenEmail := "user@example.com"
	givenPassword := "password123"
	givenUserID := 1
	givenOrgID := uuid.New()
	givenHashedPassword, _ := bcrypt.GenerateFromPassword([]byte(givenPassword), bcrypt.DefaultCost)

	givenUser := &domain.User{
		ID:             givenUserID,
		Email:          givenEmail,
		Password:       string(givenHashedPassword),
		Status:         domain.UserStatusActive,
		OrganizationID: givenOrgID,
	}

	givenRequest := &domain.LoginRequest{
		Email:    givenEmail,
		Password: givenPassword,
	}

	mockUserRepo := new(providers.MockUserRepository)
	mockTokenRepo := new(providers.MockTokenRepository)
	mockJWTManager := new(providers.MockJWTManager)
	mockEventPublisher := new(providers.MockEventPublisher)
	mockTimeManager := new(providers.MockTimeManager)
	mockLogger := new(providers.MockLogger)

	useCase := NewLoginUseCase(mockUserRepo, mockTokenRepo, mockJWTManager, mockEventPublisher, mockTimeManager, mockLogger)

	mockUserRepo.On("GetByEmail", mock.Anything, givenEmail).Return(givenUser, nil)
	mockJWTManager.On("GenerateAccessToken", givenUserID, givenOrgID, givenEmail, mock.Anything).Return("access_token", nil)
	mockJWTManager.On("GenerateRefreshToken", givenUserID).Return("refresh_token", nil)
	mockJWTManager.On("GetRefreshExpiry").Return(7 * 24 * time.Hour)
	mockJWTManager.On("GetAccessExpiry").Return(15 * time.Minute)
	mockTokenRepo.On("StoreRefreshToken", mock.Anything, mock.AnythingOfType("*domain.RefreshToken")).Return(nil)
	mockUserRepo.On("UpdateLastLogin", mock.Anything, givenUserID).Return(nil)
	mockTimeManager.On("Now").Return(time.Now())
	mockEventPublisher.On("PublishAsync", mock.Anything, "auth.user.login.succeeded", mock.Anything).Return(nil)
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	response, err := useCase.Execute(context.Background(), givenRequest)

	// Then
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "access_token", response.AccessToken)
	assert.Equal(t, "refresh_token", response.RefreshToken)
	assert.Equal(t, 900, response.ExpiresIn)
	mockUserRepo.AssertExpectations(t)
	mockTokenRepo.AssertExpectations(t)
	mockJWTManager.AssertExpectations(t)
}

func TestLoginUseCase_Execute_WithEmptyEmail_ReturnsBadRequest(t *testing.T) {
	// Given
	givenRequest := &domain.LoginRequest{
		Email:    "",
		Password: "password123",
	}

	mockUserRepo := new(providers.MockUserRepository)
	mockTokenRepo := new(providers.MockTokenRepository)
	mockJWTManager := new(providers.MockJWTManager)
	mockEventPublisher := new(providers.MockEventPublisher)
	mockTimeManager := new(providers.MockTimeManager)
	mockLogger := new(providers.MockLogger)

	useCase := NewLoginUseCase(mockUserRepo, mockTokenRepo, mockJWTManager, mockEventPublisher, mockTimeManager, mockLogger)

	// When
	response, err := useCase.Execute(context.Background(), givenRequest)

	// Then
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "email is required")
}

func TestLoginUseCase_Execute_WithEmptyPassword_ReturnsBadRequest(t *testing.T) {
	// Given
	givenRequest := &domain.LoginRequest{
		Email:    "user@example.com",
		Password: "",
	}

	mockUserRepo := new(providers.MockUserRepository)
	mockTokenRepo := new(providers.MockTokenRepository)
	mockJWTManager := new(providers.MockJWTManager)
	mockEventPublisher := new(providers.MockEventPublisher)
	mockTimeManager := new(providers.MockTimeManager)
	mockLogger := new(providers.MockLogger)

	useCase := NewLoginUseCase(mockUserRepo, mockTokenRepo, mockJWTManager, mockEventPublisher, mockTimeManager, mockLogger)

	// When
	response, err := useCase.Execute(context.Background(), givenRequest)

	// Then
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "password is required")
}

func TestLoginUseCase_Execute_WithNonExistentUser_ReturnsUnauthorized(t *testing.T) {
	// Given
	givenEmail := "nonexistent@example.com"
	givenRequest := &domain.LoginRequest{
		Email:    givenEmail,
		Password: "password123",
	}

	mockUserRepo := new(providers.MockUserRepository)
	mockTokenRepo := new(providers.MockTokenRepository)
	mockJWTManager := new(providers.MockJWTManager)
	mockEventPublisher := new(providers.MockEventPublisher)
	mockTimeManager := new(providers.MockTimeManager)
	mockLogger := new(providers.MockLogger)

	useCase := NewLoginUseCase(mockUserRepo, mockTokenRepo, mockJWTManager, mockEventPublisher, mockTimeManager, mockLogger)

	mockUserRepo.On("GetByEmail", mock.Anything, givenEmail).Return((*domain.User)(nil), assert.AnError)
	mockTimeManager.On("Now").Return(time.Now())
	mockEventPublisher.On("PublishAsync", mock.Anything, "auth.user.login.failed", mock.Anything).Return(nil)
	mockLogger.On("Error", mock.Anything, assert.AnError, mock.Anything, mock.Anything).Return()

	// When
	response, err := useCase.Execute(context.Background(), givenRequest)

	// Then
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "invalid email or password")
	mockUserRepo.AssertExpectations(t)
}

func TestLoginUseCase_Execute_WithInvalidPassword_ReturnsUnauthorized(t *testing.T) {
	// Given
	givenEmail := "user@example.com"
	givenCorrectPassword := "correct_password"
	givenWrongPassword := "wrong_password"
	givenUserID := 1
	givenOrgID := uuid.New()
	givenHashedPassword, _ := bcrypt.GenerateFromPassword([]byte(givenCorrectPassword), bcrypt.DefaultCost)

	givenUser := &domain.User{
		ID:             givenUserID,
		Email:          givenEmail,
		Password:       string(givenHashedPassword),
		Status:         domain.UserStatusActive,
		OrganizationID: givenOrgID,
	}

	givenRequest := &domain.LoginRequest{
		Email:    givenEmail,
		Password: givenWrongPassword,
	}

	mockUserRepo := new(providers.MockUserRepository)
	mockTokenRepo := new(providers.MockTokenRepository)
	mockJWTManager := new(providers.MockJWTManager)
	mockEventPublisher := new(providers.MockEventPublisher)
	mockTimeManager := new(providers.MockTimeManager)
	mockLogger := new(providers.MockLogger)

	useCase := NewLoginUseCase(mockUserRepo, mockTokenRepo, mockJWTManager, mockEventPublisher, mockTimeManager, mockLogger)

	mockUserRepo.On("GetByEmail", mock.Anything, givenEmail).Return(givenUser, nil)
	mockTimeManager.On("Now").Return(time.Now())
	mockEventPublisher.On("PublishAsync", mock.Anything, "auth.user.login.failed", mock.Anything).Return(nil)
	mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	response, err := useCase.Execute(context.Background(), givenRequest)

	// Then
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "invalid email or password")
	mockUserRepo.AssertExpectations(t)
}

func TestLoginUseCase_Execute_WithInactiveUser_ReturnsForbidden(t *testing.T) {
	// Given
	givenEmail := "user@example.com"
	givenPassword := "password123"
	givenUserID := 1
	givenOrgID := uuid.New()
	givenHashedPassword, _ := bcrypt.GenerateFromPassword([]byte(givenPassword), bcrypt.DefaultCost)

	givenUser := &domain.User{
		ID:             givenUserID,
		Email:          givenEmail,
		Password:       string(givenHashedPassword),
		Status:         domain.UserStatusInactive,
		OrganizationID: givenOrgID,
	}

	givenRequest := &domain.LoginRequest{
		Email:    givenEmail,
		Password: givenPassword,
	}

	mockUserRepo := new(providers.MockUserRepository)
	mockTokenRepo := new(providers.MockTokenRepository)
	mockJWTManager := new(providers.MockJWTManager)
	mockEventPublisher := new(providers.MockEventPublisher)
	mockTimeManager := new(providers.MockTimeManager)
	mockLogger := new(providers.MockLogger)

	useCase := NewLoginUseCase(mockUserRepo, mockTokenRepo, mockJWTManager, mockEventPublisher, mockTimeManager, mockLogger)

	mockUserRepo.On("GetByEmail", mock.Anything, givenEmail).Return(givenUser, nil)
	mockTimeManager.On("Now").Return(time.Now())
	mockEventPublisher.On("PublishAsync", mock.Anything, "auth.user.login.failed", mock.Anything).Return(nil)
	mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	response, err := useCase.Execute(context.Background(), givenRequest)

	// Then
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "account is not active")
	mockUserRepo.AssertExpectations(t)
}

func TestLoginUseCase_Execute_WithSuspendedUser_ReturnsForbidden(t *testing.T) {
	// Given
	givenEmail := "user@example.com"
	givenPassword := "password123"
	givenUserID := 1
	givenOrgID := uuid.New()
	givenHashedPassword, _ := bcrypt.GenerateFromPassword([]byte(givenPassword), bcrypt.DefaultCost)

	givenUser := &domain.User{
		ID:             givenUserID,
		Email:          givenEmail,
		Password:       string(givenHashedPassword),
		Status:         domain.UserStatusSuspended,
		OrganizationID: givenOrgID,
	}

	givenRequest := &domain.LoginRequest{
		Email:    givenEmail,
		Password: givenPassword,
	}

	mockUserRepo := new(providers.MockUserRepository)
	mockTokenRepo := new(providers.MockTokenRepository)
	mockJWTManager := new(providers.MockJWTManager)
	mockEventPublisher := new(providers.MockEventPublisher)
	mockTimeManager := new(providers.MockTimeManager)
	mockLogger := new(providers.MockLogger)

	useCase := NewLoginUseCase(mockUserRepo, mockTokenRepo, mockJWTManager, mockEventPublisher, mockTimeManager, mockLogger)

	mockUserRepo.On("GetByEmail", mock.Anything, givenEmail).Return(givenUser, nil)
	mockTimeManager.On("Now").Return(time.Now())
	mockEventPublisher.On("PublishAsync", mock.Anything, "auth.user.login.failed", mock.Anything).Return(nil)
	mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	response, err := useCase.Execute(context.Background(), givenRequest)

	// Then
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "account is not active")
	mockUserRepo.AssertExpectations(t)
}

func TestLoginUseCase_Execute_WhenAccessTokenGenerationFails_ReturnsError(t *testing.T) {
	// Given
	givenEmail := "user@example.com"
	givenPassword := "password123"
	givenUserID := 1
	givenOrgID := uuid.New()
	givenHashedPassword, _ := bcrypt.GenerateFromPassword([]byte(givenPassword), bcrypt.DefaultCost)

	givenUser := &domain.User{
		ID:             givenUserID,
		Email:          givenEmail,
		Password:       string(givenHashedPassword),
		Status:         domain.UserStatusActive,
		OrganizationID: givenOrgID,
	}

	givenRequest := &domain.LoginRequest{
		Email:    givenEmail,
		Password: givenPassword,
	}

	mockUserRepo := new(providers.MockUserRepository)
	mockTokenRepo := new(providers.MockTokenRepository)
	mockJWTManager := new(providers.MockJWTManager)
	mockEventPublisher := new(providers.MockEventPublisher)
	mockTimeManager := new(providers.MockTimeManager)
	mockLogger := new(providers.MockLogger)

	useCase := NewLoginUseCase(mockUserRepo, mockTokenRepo, mockJWTManager, mockEventPublisher, mockTimeManager, mockLogger)

	mockUserRepo.On("GetByEmail", mock.Anything, givenEmail).Return(givenUser, nil)
	mockJWTManager.On("GenerateAccessToken", givenUserID, givenOrgID, givenEmail, mock.Anything).Return("", assert.AnError)
	mockLogger.On("Error", mock.Anything, assert.AnError, mock.Anything, mock.Anything).Return()

	// When
	response, err := useCase.Execute(context.Background(), givenRequest)

	// Then
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "failed to generate access token")
	mockUserRepo.AssertExpectations(t)
	mockJWTManager.AssertExpectations(t)
}

func TestLoginUseCase_Execute_WhenRefreshTokenGenerationFails_ReturnsError(t *testing.T) {
	// Given
	givenEmail := "user@example.com"
	givenPassword := "password123"
	givenUserID := 1
	givenOrgID := uuid.New()
	givenHashedPassword, _ := bcrypt.GenerateFromPassword([]byte(givenPassword), bcrypt.DefaultCost)

	givenUser := &domain.User{
		ID:             givenUserID,
		Email:          givenEmail,
		Password:       string(givenHashedPassword),
		Status:         domain.UserStatusActive,
		OrganizationID: givenOrgID,
	}

	givenRequest := &domain.LoginRequest{
		Email:    givenEmail,
		Password: givenPassword,
	}

	mockUserRepo := new(providers.MockUserRepository)
	mockTokenRepo := new(providers.MockTokenRepository)
	mockJWTManager := new(providers.MockJWTManager)
	mockEventPublisher := new(providers.MockEventPublisher)
	mockTimeManager := new(providers.MockTimeManager)
	mockLogger := new(providers.MockLogger)

	useCase := NewLoginUseCase(mockUserRepo, mockTokenRepo, mockJWTManager, mockEventPublisher, mockTimeManager, mockLogger)

	mockUserRepo.On("GetByEmail", mock.Anything, givenEmail).Return(givenUser, nil)
	mockJWTManager.On("GenerateAccessToken", givenUserID, givenOrgID, givenEmail, mock.Anything).Return("access_token", nil)
	mockJWTManager.On("GenerateRefreshToken", givenUserID).Return("", assert.AnError)
	mockLogger.On("Error", mock.Anything, assert.AnError, mock.Anything, mock.Anything).Return()

	// When
	response, err := useCase.Execute(context.Background(), givenRequest)

	// Then
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "failed to generate refresh token")
	mockUserRepo.AssertExpectations(t)
	mockJWTManager.AssertExpectations(t)
}

func TestLoginUseCase_Execute_WhenStoreRefreshTokenFails_ReturnsError(t *testing.T) {
	// Given
	givenEmail := "user@example.com"
	givenPassword := "password123"
	givenUserID := 1
	givenOrgID := uuid.New()
	givenHashedPassword, _ := bcrypt.GenerateFromPassword([]byte(givenPassword), bcrypt.DefaultCost)

	givenUser := &domain.User{
		ID:             givenUserID,
		Email:          givenEmail,
		Password:       string(givenHashedPassword),
		Status:         domain.UserStatusActive,
		OrganizationID: givenOrgID,
	}

	givenRequest := &domain.LoginRequest{
		Email:    givenEmail,
		Password: givenPassword,
	}

	mockUserRepo := new(providers.MockUserRepository)
	mockTokenRepo := new(providers.MockTokenRepository)
	mockJWTManager := new(providers.MockJWTManager)
	mockEventPublisher := new(providers.MockEventPublisher)
	mockTimeManager := new(providers.MockTimeManager)
	mockLogger := new(providers.MockLogger)

	useCase := NewLoginUseCase(mockUserRepo, mockTokenRepo, mockJWTManager, mockEventPublisher, mockTimeManager, mockLogger)

	mockUserRepo.On("GetByEmail", mock.Anything, givenEmail).Return(givenUser, nil)
	mockJWTManager.On("GenerateAccessToken", givenUserID, givenOrgID, givenEmail, mock.Anything).Return("access_token", nil)
	mockJWTManager.On("GenerateRefreshToken", givenUserID).Return("refresh_token", nil)
	mockJWTManager.On("GetRefreshExpiry").Return(7 * 24 * time.Hour)
	mockTokenRepo.On("StoreRefreshToken", mock.Anything, mock.AnythingOfType("*domain.RefreshToken")).Return(assert.AnError)
	mockLogger.On("Error", mock.Anything, assert.AnError, mock.Anything, mock.Anything).Return()

	// When
	response, err := useCase.Execute(context.Background(), givenRequest)

	// Then
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "failed to store refresh token")
	mockUserRepo.AssertExpectations(t)
	mockJWTManager.AssertExpectations(t)
	mockTokenRepo.AssertExpectations(t)
}
