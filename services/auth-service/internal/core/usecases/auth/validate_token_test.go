package auth

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/giia/giia-core-engine/services/auth-service/internal/core/domain"
	"github.com/giia/giia-core-engine/services/auth-service/internal/core/providers"
)

func TestValidateTokenUseCase_Execute_WithValidToken_ReturnsValidResult(t *testing.T) {
	// Given
	givenUserID := uuid.New()
	givenOrgID := uuid.New()
	givenEmail := "user@example.com"
	givenRoles := []string{"admin", "user"}
	givenToken := "valid_access_token"
	givenExpiresAt := time.Now().Add(1 * time.Hour).Unix()

	givenClaims := &providers.Claims{
		UserID:         givenUserID.String(),
		Email:          givenEmail,
		OrganizationID: givenOrgID.String(),
		Roles:          givenRoles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Unix(givenExpiresAt, 0)),
		},
	}

	givenUser := &domain.User{
		ID:             givenUserID,
		OrganizationID: givenOrgID,
		Email:          givenEmail,
		Status:         domain.UserStatusActive,
	}

	mockUserRepo := new(providers.MockUserRepository)
	mockJWTManager := new(providers.MockJWTManager)
	mockLogger := new(providers.MockLogger)

	useCase := NewValidateTokenUseCase(mockUserRepo, mockJWTManager, mockLogger)

	mockJWTManager.On("ValidateAccessToken", givenToken).Return(givenClaims, nil)
	mockUserRepo.On("GetByID", mock.Anything, givenUserID).Return(givenUser, nil)
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	result, err := useCase.Execute(context.Background(), givenToken)

	// Then
	assert.NoError(t, err)
	assert.True(t, result.Valid)
	assert.Equal(t, givenUserID, result.UserID)
	assert.Equal(t, givenOrgID, result.OrganizationID)
	assert.Equal(t, givenEmail, result.Email)
	assert.Equal(t, givenRoles, result.Roles)
	assert.Equal(t, givenExpiresAt, result.ExpiresAt)
	mockUserRepo.AssertExpectations(t)
	mockJWTManager.AssertExpectations(t)
}

func TestValidateTokenUseCase_Execute_WithEmptyToken_ReturnsBadRequest(t *testing.T) {
	// Given
	mockUserRepo := new(providers.MockUserRepository)
	mockJWTManager := new(providers.MockJWTManager)
	mockLogger := new(providers.MockLogger)

	useCase := NewValidateTokenUseCase(mockUserRepo, mockJWTManager, mockLogger)

	// When
	result, err := useCase.Execute(context.Background(), "")

	// Then
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "token is required")
}

func TestValidateTokenUseCase_Execute_WithInvalidToken_ReturnsInvalidResult(t *testing.T) {
	// Given
	givenInvalidToken := "invalid.jwt.token"

	mockUserRepo := new(providers.MockUserRepository)
	mockJWTManager := new(providers.MockJWTManager)
	mockLogger := new(providers.MockLogger)

	useCase := NewValidateTokenUseCase(mockUserRepo, mockJWTManager, mockLogger)

	mockJWTManager.On("ValidateAccessToken", givenInvalidToken).Return((*providers.Claims)(nil), assert.AnError)
	mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	result, err := useCase.Execute(context.Background(), givenInvalidToken)

	// Then
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Valid)
	mockJWTManager.AssertExpectations(t)
}

func TestValidateTokenUseCase_Execute_WithExpiredToken_ReturnsInvalidResult(t *testing.T) {
	// Given
	givenExpiredToken := "expired.jwt.token"

	mockUserRepo := new(providers.MockUserRepository)
	mockJWTManager := new(providers.MockJWTManager)
	mockLogger := new(providers.MockLogger)

	useCase := NewValidateTokenUseCase(mockUserRepo, mockJWTManager, mockLogger)

	mockJWTManager.On("ValidateAccessToken", givenExpiredToken).Return((*providers.Claims)(nil), assert.AnError)
	mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	result, err := useCase.Execute(context.Background(), givenExpiredToken)

	// Then
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Valid)
	mockJWTManager.AssertExpectations(t)
}

func TestValidateTokenUseCase_Execute_WithInvalidUserIDInToken_ReturnsInvalidResult(t *testing.T) {
	// Given
	givenToken := "token_with_invalid_user_id"

	givenClaims := &providers.Claims{
		UserID:         "invalid-uuid-format",
		Email:          "user@example.com",
		OrganizationID: uuid.New().String(),
		Roles:          []string{"user"},
	}

	mockUserRepo := new(providers.MockUserRepository)
	mockJWTManager := new(providers.MockJWTManager)
	mockLogger := new(providers.MockLogger)

	useCase := NewValidateTokenUseCase(mockUserRepo, mockJWTManager, mockLogger)

	mockJWTManager.On("ValidateAccessToken", givenToken).Return(givenClaims, nil)
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	result, err := useCase.Execute(context.Background(), givenToken)

	// Then
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Valid)
	mockJWTManager.AssertExpectations(t)
}

func TestValidateTokenUseCase_Execute_WithInvalidOrganizationIDInToken_ReturnsInvalidResult(t *testing.T) {
	// Given
	givenToken := "token_with_invalid_org_id"
	givenUserID := uuid.New()

	givenClaims := &providers.Claims{
		UserID:         givenUserID.String(),
		Email:          "user@example.com",
		OrganizationID: "invalid-uuid-format",
		Roles:          []string{"user"},
	}

	mockUserRepo := new(providers.MockUserRepository)
	mockJWTManager := new(providers.MockJWTManager)
	mockLogger := new(providers.MockLogger)

	useCase := NewValidateTokenUseCase(mockUserRepo, mockJWTManager, mockLogger)

	mockJWTManager.On("ValidateAccessToken", givenToken).Return(givenClaims, nil)
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	result, err := useCase.Execute(context.Background(), givenToken)

	// Then
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Valid)
	mockJWTManager.AssertExpectations(t)
}

func TestValidateTokenUseCase_Execute_WithUserNotFound_ReturnsInvalidResult(t *testing.T) {
	// Given
	givenUserID := uuid.New()
	givenOrgID := uuid.New()
	givenToken := "token_for_nonexistent_user"

	givenClaims := &providers.Claims{
		UserID:         givenUserID.String(),
		Email:          "user@example.com",
		OrganizationID: givenOrgID.String(),
		Roles:          []string{"user"},
	}

	mockUserRepo := new(providers.MockUserRepository)
	mockJWTManager := new(providers.MockJWTManager)
	mockLogger := new(providers.MockLogger)

	useCase := NewValidateTokenUseCase(mockUserRepo, mockJWTManager, mockLogger)

	mockJWTManager.On("ValidateAccessToken", givenToken).Return(givenClaims, nil)
	mockUserRepo.On("GetByID", mock.Anything, givenUserID).Return((*domain.User)(nil), assert.AnError)
	mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	result, err := useCase.Execute(context.Background(), givenToken)

	// Then
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Valid)
	mockUserRepo.AssertExpectations(t)
	mockJWTManager.AssertExpectations(t)
}

func TestValidateTokenUseCase_Execute_WithInactiveUser_ReturnsInvalidResult(t *testing.T) {
	// Given
	givenUserID := uuid.New()
	givenOrgID := uuid.New()
	givenToken := "token_for_inactive_user"

	givenClaims := &providers.Claims{
		UserID:         givenUserID.String(),
		Email:          "user@example.com",
		OrganizationID: givenOrgID.String(),
		Roles:          []string{"user"},
	}

	givenUser := &domain.User{
		ID:             givenUserID,
		OrganizationID: givenOrgID,
		Email:          "user@example.com",
		Status:         domain.UserStatusInactive,
	}

	mockUserRepo := new(providers.MockUserRepository)
	mockJWTManager := new(providers.MockJWTManager)
	mockLogger := new(providers.MockLogger)

	useCase := NewValidateTokenUseCase(mockUserRepo, mockJWTManager, mockLogger)

	mockJWTManager.On("ValidateAccessToken", givenToken).Return(givenClaims, nil)
	mockUserRepo.On("GetByID", mock.Anything, givenUserID).Return(givenUser, nil)
	mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	result, err := useCase.Execute(context.Background(), givenToken)

	// Then
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Valid)
	mockUserRepo.AssertExpectations(t)
	mockJWTManager.AssertExpectations(t)
}

func TestValidateTokenUseCase_Execute_WithSuspendedUser_ReturnsInvalidResult(t *testing.T) {
	// Given
	givenUserID := uuid.New()
	givenOrgID := uuid.New()
	givenToken := "token_for_suspended_user"

	givenClaims := &providers.Claims{
		UserID:         givenUserID.String(),
		Email:          "user@example.com",
		OrganizationID: givenOrgID.String(),
		Roles:          []string{"user"},
	}

	givenUser := &domain.User{
		ID:             givenUserID,
		OrganizationID: givenOrgID,
		Email:          "user@example.com",
		Status:         domain.UserStatusSuspended,
	}

	mockUserRepo := new(providers.MockUserRepository)
	mockJWTManager := new(providers.MockJWTManager)
	mockLogger := new(providers.MockLogger)

	useCase := NewValidateTokenUseCase(mockUserRepo, mockJWTManager, mockLogger)

	mockJWTManager.On("ValidateAccessToken", givenToken).Return(givenClaims, nil)
	mockUserRepo.On("GetByID", mock.Anything, givenUserID).Return(givenUser, nil)
	mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	result, err := useCase.Execute(context.Background(), givenToken)

	// Then
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Valid)
	mockUserRepo.AssertExpectations(t)
	mockJWTManager.AssertExpectations(t)
}
