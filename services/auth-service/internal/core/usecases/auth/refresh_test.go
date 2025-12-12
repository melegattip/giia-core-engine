package auth

import (
	"context"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/giia/giia-core-engine/services/auth-service/internal/core/domain"
	"github.com/giia/giia-core-engine/services/auth-service/internal/core/providers"
)

func TestRefreshTokenUseCase_Execute_WithValidToken_ReturnsNewAccessToken(t *testing.T) {
	// Given
	givenUserID := uuid.New()
	givenOrgID := uuid.New()
	givenEmail := "user@example.com"
	givenRefreshToken := "valid_refresh_token"
	givenTokenHash := hashToken(givenRefreshToken)

	givenClaims := &jwt.RegisteredClaims{
		Subject: givenUserID.String(),
	}

	givenStoredToken := &domain.RefreshToken{
		TokenHash: givenTokenHash,
		UserID:    givenUserID,
		Revoked:   false,
	}

	givenUser := &domain.User{
		ID:             givenUserID,
		Email:          givenEmail,
		Status:         domain.UserStatusActive,
		OrganizationID: givenOrgID,
	}

	mockUserRepo := new(providers.MockUserRepository)
	mockTokenRepo := new(providers.MockTokenRepository)
	mockJWTManager := new(providers.MockJWTManager)
	mockLogger := new(providers.MockLogger)

	useCase := NewRefreshTokenUseCase(mockUserRepo, mockTokenRepo, mockJWTManager, mockLogger)

	mockJWTManager.On("ValidateRefreshToken", givenRefreshToken).Return(givenClaims, nil)
	mockTokenRepo.On("GetRefreshToken", mock.Anything, givenTokenHash).Return(givenStoredToken, nil)
	mockUserRepo.On("GetByID", mock.Anything, givenUserID).Return(givenUser, nil)
	mockJWTManager.On("GenerateAccessToken", givenUserID, givenOrgID, givenEmail, mock.Anything).Return("new_access_token", nil)
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	accessToken, err := useCase.Execute(context.Background(), givenRefreshToken)

	// Then
	assert.NoError(t, err)
	assert.Equal(t, "new_access_token", accessToken)
	mockUserRepo.AssertExpectations(t)
	mockTokenRepo.AssertExpectations(t)
	mockJWTManager.AssertExpectations(t)
}

func TestRefreshTokenUseCase_Execute_WithEmptyToken_ReturnsBadRequest(t *testing.T) {
	// Given
	mockUserRepo := new(providers.MockUserRepository)
	mockTokenRepo := new(providers.MockTokenRepository)
	mockJWTManager := new(providers.MockJWTManager)
	mockLogger := new(providers.MockLogger)

	useCase := NewRefreshTokenUseCase(mockUserRepo, mockTokenRepo, mockJWTManager, mockLogger)

	// When
	accessToken, err := useCase.Execute(context.Background(), "")

	// Then
	assert.Error(t, err)
	assert.Empty(t, accessToken)
	assert.Contains(t, err.Error(), "refresh token is required")
}

func TestRefreshTokenUseCase_Execute_WithInvalidToken_ReturnsUnauthorized(t *testing.T) {
	// Given
	givenInvalidToken := "invalid_token"

	mockUserRepo := new(providers.MockUserRepository)
	mockTokenRepo := new(providers.MockTokenRepository)
	mockJWTManager := new(providers.MockJWTManager)
	mockLogger := new(providers.MockLogger)

	useCase := NewRefreshTokenUseCase(mockUserRepo, mockTokenRepo, mockJWTManager, mockLogger)

	mockJWTManager.On("ValidateRefreshToken", givenInvalidToken).Return((*jwt.RegisteredClaims)(nil), assert.AnError)
	mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	accessToken, err := useCase.Execute(context.Background(), givenInvalidToken)

	// Then
	assert.Error(t, err)
	assert.Empty(t, accessToken)
	assert.Contains(t, err.Error(), "invalid or expired refresh token")
	mockJWTManager.AssertExpectations(t)
}

func TestRefreshTokenUseCase_Execute_WithExpiredToken_ReturnsUnauthorized(t *testing.T) {
	// Given
	givenExpiredToken := "expired_token"
	givenTokenHash := hashToken(givenExpiredToken)
	givenUserID := uuid.New()

	givenClaims := &jwt.RegisteredClaims{
		Subject: givenUserID.String(),
	}

	mockUserRepo := new(providers.MockUserRepository)
	mockTokenRepo := new(providers.MockTokenRepository)
	mockJWTManager := new(providers.MockJWTManager)
	mockLogger := new(providers.MockLogger)

	useCase := NewRefreshTokenUseCase(mockUserRepo, mockTokenRepo, mockJWTManager, mockLogger)

	mockJWTManager.On("ValidateRefreshToken", givenExpiredToken).Return(givenClaims, nil)
	mockTokenRepo.On("GetRefreshToken", mock.Anything, givenTokenHash).Return((*domain.RefreshToken)(nil), assert.AnError)
	mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	accessToken, err := useCase.Execute(context.Background(), givenExpiredToken)

	// Then
	assert.Error(t, err)
	assert.Empty(t, accessToken)
	assert.Contains(t, err.Error(), "refresh token not found or expired")
	mockJWTManager.AssertExpectations(t)
	mockTokenRepo.AssertExpectations(t)
}

func TestRefreshTokenUseCase_Execute_WithRevokedToken_ReturnsUnauthorized(t *testing.T) {
	// Given
	givenUserID := uuid.New()
	givenRevokedToken := "revoked_token"
	givenTokenHash := hashToken(givenRevokedToken)

	givenClaims := &jwt.RegisteredClaims{
		Subject: givenUserID.String(),
	}

	givenStoredToken := &domain.RefreshToken{
		TokenHash: givenTokenHash,
		UserID:    givenUserID,
		Revoked:   true,
	}

	mockUserRepo := new(providers.MockUserRepository)
	mockTokenRepo := new(providers.MockTokenRepository)
	mockJWTManager := new(providers.MockJWTManager)
	mockLogger := new(providers.MockLogger)

	useCase := NewRefreshTokenUseCase(mockUserRepo, mockTokenRepo, mockJWTManager, mockLogger)

	mockJWTManager.On("ValidateRefreshToken", givenRevokedToken).Return(givenClaims, nil)
	mockTokenRepo.On("GetRefreshToken", mock.Anything, givenTokenHash).Return(givenStoredToken, nil)
	mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	accessToken, err := useCase.Execute(context.Background(), givenRevokedToken)

	// Then
	assert.Error(t, err)
	assert.Empty(t, accessToken)
	assert.Contains(t, err.Error(), "refresh token has been revoked")
	mockJWTManager.AssertExpectations(t)
	mockTokenRepo.AssertExpectations(t)
}

func TestRefreshTokenUseCase_Execute_WithInvalidUserIDInToken_ReturnsUnauthorized(t *testing.T) {
	// Given
	givenInvalidToken := "token_with_invalid_user_id"
	givenTokenHash := hashToken(givenInvalidToken)

	givenClaims := &jwt.RegisteredClaims{
		Subject: "invalid-uuid",
	}

	givenStoredToken := &domain.RefreshToken{
		TokenHash: givenTokenHash,
		UserID:    uuid.New(),
		Revoked:   false,
	}

	mockUserRepo := new(providers.MockUserRepository)
	mockTokenRepo := new(providers.MockTokenRepository)
	mockJWTManager := new(providers.MockJWTManager)
	mockLogger := new(providers.MockLogger)

	useCase := NewRefreshTokenUseCase(mockUserRepo, mockTokenRepo, mockJWTManager, mockLogger)

	mockJWTManager.On("ValidateRefreshToken", givenInvalidToken).Return(givenClaims, nil)
	mockTokenRepo.On("GetRefreshToken", mock.Anything, givenTokenHash).Return(givenStoredToken, nil)

	// When
	accessToken, err := useCase.Execute(context.Background(), givenInvalidToken)

	// Then
	assert.Error(t, err)
	assert.Empty(t, accessToken)
	assert.Contains(t, err.Error(), "invalid user ID in token")
	mockJWTManager.AssertExpectations(t)
	mockTokenRepo.AssertExpectations(t)
}

func TestRefreshTokenUseCase_Execute_WithNonExistentUser_ReturnsUnauthorized(t *testing.T) {
	// Given
	givenUserID := uuid.New()
	givenRefreshToken := "valid_token_nonexistent_user"
	givenTokenHash := hashToken(givenRefreshToken)

	givenClaims := &jwt.RegisteredClaims{
		Subject: givenUserID.String(),
	}

	givenStoredToken := &domain.RefreshToken{
		TokenHash: givenTokenHash,
		UserID:    givenUserID,
		Revoked:   false,
	}

	mockUserRepo := new(providers.MockUserRepository)
	mockTokenRepo := new(providers.MockTokenRepository)
	mockJWTManager := new(providers.MockJWTManager)
	mockLogger := new(providers.MockLogger)

	useCase := NewRefreshTokenUseCase(mockUserRepo, mockTokenRepo, mockJWTManager, mockLogger)

	mockJWTManager.On("ValidateRefreshToken", givenRefreshToken).Return(givenClaims, nil)
	mockTokenRepo.On("GetRefreshToken", mock.Anything, givenTokenHash).Return(givenStoredToken, nil)
	mockUserRepo.On("GetByID", mock.Anything, givenUserID).Return((*domain.User)(nil), assert.AnError)
	mockLogger.On("Error", mock.Anything, assert.AnError, mock.Anything, mock.Anything).Return()

	// When
	accessToken, err := useCase.Execute(context.Background(), givenRefreshToken)

	// Then
	assert.Error(t, err)
	assert.Empty(t, accessToken)
	assert.Contains(t, err.Error(), "user not found")
	mockJWTManager.AssertExpectations(t)
	mockTokenRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestRefreshTokenUseCase_Execute_WithInactiveUser_ReturnsForbidden(t *testing.T) {
	// Given
	givenUserID := uuid.New()
	givenOrgID := uuid.New()
	givenRefreshToken := "valid_token_inactive_user"
	givenTokenHash := hashToken(givenRefreshToken)

	givenClaims := &jwt.RegisteredClaims{
		Subject: givenUserID.String(),
	}

	givenStoredToken := &domain.RefreshToken{
		TokenHash: givenTokenHash,
		UserID:    givenUserID,
		Revoked:   false,
	}

	givenUser := &domain.User{
		ID:             givenUserID,
		Email:          "user@example.com",
		Status:         domain.UserStatusInactive,
		OrganizationID: givenOrgID,
	}

	mockUserRepo := new(providers.MockUserRepository)
	mockTokenRepo := new(providers.MockTokenRepository)
	mockJWTManager := new(providers.MockJWTManager)
	mockLogger := new(providers.MockLogger)

	useCase := NewRefreshTokenUseCase(mockUserRepo, mockTokenRepo, mockJWTManager, mockLogger)

	mockJWTManager.On("ValidateRefreshToken", givenRefreshToken).Return(givenClaims, nil)
	mockTokenRepo.On("GetRefreshToken", mock.Anything, givenTokenHash).Return(givenStoredToken, nil)
	mockUserRepo.On("GetByID", mock.Anything, givenUserID).Return(givenUser, nil)
	mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	accessToken, err := useCase.Execute(context.Background(), givenRefreshToken)

	// Then
	assert.Error(t, err)
	assert.Empty(t, accessToken)
	assert.Contains(t, err.Error(), "account is not active")
	mockJWTManager.AssertExpectations(t)
	mockTokenRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestRefreshTokenUseCase_Execute_WhenAccessTokenGenerationFails_ReturnsInternalServerError(t *testing.T) {
	// Given
	givenUserID := uuid.New()
	givenOrgID := uuid.New()
	givenEmail := "user@example.com"
	givenRefreshToken := "valid_token"
	givenTokenHash := hashToken(givenRefreshToken)

	givenClaims := &jwt.RegisteredClaims{
		Subject: givenUserID.String(),
	}

	givenStoredToken := &domain.RefreshToken{
		TokenHash: givenTokenHash,
		UserID:    givenUserID,
		Revoked:   false,
	}

	givenUser := &domain.User{
		ID:             givenUserID,
		Email:          givenEmail,
		Status:         domain.UserStatusActive,
		OrganizationID: givenOrgID,
	}

	mockUserRepo := new(providers.MockUserRepository)
	mockTokenRepo := new(providers.MockTokenRepository)
	mockJWTManager := new(providers.MockJWTManager)
	mockLogger := new(providers.MockLogger)

	useCase := NewRefreshTokenUseCase(mockUserRepo, mockTokenRepo, mockJWTManager, mockLogger)

	mockJWTManager.On("ValidateRefreshToken", givenRefreshToken).Return(givenClaims, nil)
	mockTokenRepo.On("GetRefreshToken", mock.Anything, givenTokenHash).Return(givenStoredToken, nil)
	mockUserRepo.On("GetByID", mock.Anything, givenUserID).Return(givenUser, nil)
	mockJWTManager.On("GenerateAccessToken", givenUserID, givenOrgID, givenEmail, mock.Anything).Return("", assert.AnError)
	mockLogger.On("Error", mock.Anything, assert.AnError, mock.Anything, mock.Anything).Return()

	// When
	accessToken, err := useCase.Execute(context.Background(), givenRefreshToken)

	// Then
	assert.Error(t, err)
	assert.Empty(t, accessToken)
	assert.Contains(t, err.Error(), "failed to generate access token")
	mockJWTManager.AssertExpectations(t)
	mockTokenRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}
