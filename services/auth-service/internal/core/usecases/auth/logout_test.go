package auth

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/melegattip/giia-core-engine/services/auth-service/internal/core/providers"
)

func TestLogoutUseCase_Execute_WithValidToken_LogsOutSuccessfully(t *testing.T) {
	// Given
	givenAccessToken := "valid_access_token"
	givenUserID := uuid.New()
	givenTTL := 15 * time.Minute

	mockTokenRepo := new(providers.MockTokenRepository)
	mockJWTManager := new(providers.MockJWTManager)
	mockLogger := new(providers.MockLogger)

	useCase := NewLogoutUseCase(mockTokenRepo, mockJWTManager, mockLogger)

	mockJWTManager.On("GetAccessExpiry").Return(givenTTL)
	mockTokenRepo.On("BlacklistToken", mock.Anything, givenAccessToken, givenTTL).Return(nil)
	mockTokenRepo.On("RevokeAllUserTokens", mock.Anything, givenUserID).Return(nil)
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	err := useCase.Execute(context.Background(), givenAccessToken, givenUserID)

	// Then
	assert.NoError(t, err)
	mockTokenRepo.AssertExpectations(t)
	mockJWTManager.AssertExpectations(t)
}

func TestLogoutUseCase_Execute_WithEmptyToken_ReturnsBadRequest(t *testing.T) {
	// Given
	givenUserID := uuid.New()

	mockTokenRepo := new(providers.MockTokenRepository)
	mockJWTManager := new(providers.MockJWTManager)
	mockLogger := new(providers.MockLogger)

	useCase := NewLogoutUseCase(mockTokenRepo, mockJWTManager, mockLogger)

	// When
	err := useCase.Execute(context.Background(), "", givenUserID)

	// Then
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "access token is required")
}

func TestLogoutUseCase_Execute_WhenBlacklistFails_ReturnsInternalServerError(t *testing.T) {
	// Given
	givenAccessToken := "valid_access_token"
	givenUserID := uuid.New()
	givenTTL := 15 * time.Minute

	mockTokenRepo := new(providers.MockTokenRepository)
	mockJWTManager := new(providers.MockJWTManager)
	mockLogger := new(providers.MockLogger)

	useCase := NewLogoutUseCase(mockTokenRepo, mockJWTManager, mockLogger)

	mockJWTManager.On("GetAccessExpiry").Return(givenTTL)
	mockTokenRepo.On("BlacklistToken", mock.Anything, givenAccessToken, givenTTL).Return(assert.AnError)
	mockLogger.On("Error", mock.Anything, assert.AnError, mock.Anything, mock.Anything).Return()

	// When
	err := useCase.Execute(context.Background(), givenAccessToken, givenUserID)

	// Then
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to blacklist token")
	mockTokenRepo.AssertExpectations(t)
	mockJWTManager.AssertExpectations(t)
}

func TestLogoutUseCase_Execute_WhenRevokeTokensFails_ContinuesSuccessfully(t *testing.T) {
	// Given
	givenAccessToken := "valid_access_token"
	givenUserID := uuid.New()
	givenTTL := 15 * time.Minute

	mockTokenRepo := new(providers.MockTokenRepository)
	mockJWTManager := new(providers.MockJWTManager)
	mockLogger := new(providers.MockLogger)

	useCase := NewLogoutUseCase(mockTokenRepo, mockJWTManager, mockLogger)

	mockJWTManager.On("GetAccessExpiry").Return(givenTTL)
	mockTokenRepo.On("BlacklistToken", mock.Anything, givenAccessToken, givenTTL).Return(nil)
	mockTokenRepo.On("RevokeAllUserTokens", mock.Anything, givenUserID).Return(assert.AnError)
	mockLogger.On("Error", mock.Anything, assert.AnError, mock.Anything, mock.Anything).Return()
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	err := useCase.Execute(context.Background(), givenAccessToken, givenUserID)

	// Then
	assert.NoError(t, err)
	mockTokenRepo.AssertExpectations(t)
	mockJWTManager.AssertExpectations(t)
}
