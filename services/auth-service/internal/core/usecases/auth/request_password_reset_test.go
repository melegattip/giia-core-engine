package auth

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	pkgErrors "github.com/giia/giia-core-engine/pkg/errors"
	pkgLogger "github.com/giia/giia-core-engine/pkg/logger"
	"github.com/giia/giia-core-engine/services/auth-service/internal/core/domain"
	"github.com/giia/giia-core-engine/services/auth-service/internal/core/providers"
)

type MockEmailService struct {
	mock.Mock
}

func (m *MockEmailService) SendActivationEmail(ctx context.Context, to, token, userName string) error {
	args := m.Called(ctx, to, token, userName)
	return args.Error(0)
}

func (m *MockEmailService) SendPasswordResetEmail(ctx context.Context, to, token, userName string) error {
	args := m.Called(ctx, to, token, userName)
	return args.Error(0)
}

func (m *MockEmailService) SendWelcomeEmail(ctx context.Context, to, userName string) error {
	args := m.Called(ctx, to, userName)
	return args.Error(0)
}

func TestRequestPasswordResetUseCase_Execute_Success(t *testing.T) {
	// Given
	mockUserRepo := new(providers.MockUserRepository)
	mockTokenRepo := new(providers.MockTokenRepository)
	mockEmailService := new(MockEmailService)
	logger := pkgLogger.New("test", "error")

	useCase := NewRequestPasswordResetUseCase(
		mockUserRepo,
		mockTokenRepo,
		mockEmailService,
		logger,
	)

	ctx := context.Background()
	email := "test@example.com"
	orgID := uuid.New()
	userID := uuid.New()

	givenUser := &domain.User{
		ID:             userID,
		Email:          email,
		FirstName:      "Test",
		LastName:       "User",
		OrganizationID: orgID,
	}

	mockUserRepo.On("GetByEmailAndOrg", ctx, email, orgID).Return(givenUser, nil)
	mockTokenRepo.On("StorePasswordResetToken", ctx, mock.AnythingOfType("*domain.PasswordResetToken")).Return(nil)
	mockEmailService.On("SendPasswordResetEmail", ctx, email, mock.AnythingOfType("string"), "Test").Return(nil)

	// When
	err := useCase.Execute(ctx, email, orgID)

	// Then
	assert.NoError(t, err)
	mockUserRepo.AssertExpectations(t)
	mockTokenRepo.AssertExpectations(t)
	mockEmailService.AssertExpectations(t)
}

func TestRequestPasswordResetUseCase_Execute_EmptyEmail_ReturnsError(t *testing.T) {
	// Given
	mockUserRepo := new(providers.MockUserRepository)
	mockTokenRepo := new(providers.MockTokenRepository)
	mockEmailService := new(MockEmailService)
	logger := pkgLogger.New("test", "error")

	useCase := NewRequestPasswordResetUseCase(
		mockUserRepo,
		mockTokenRepo,
		mockEmailService,
		logger,
	)

	ctx := context.Background()
	email := ""
	orgID := uuid.New()

	// When
	err := useCase.Execute(ctx, email, orgID)

	// Then
	assert.Error(t, err)
	assert.Equal(t, "email is required", err.(*pkgErrors.CustomError).Message)
	mockUserRepo.AssertNotCalled(t, "GetByEmailAndOrg")
}

func TestRequestPasswordResetUseCase_Execute_InvalidEmail_ReturnsError(t *testing.T) {
	// Given
	mockUserRepo := new(providers.MockUserRepository)
	mockTokenRepo := new(providers.MockTokenRepository)
	mockEmailService := new(MockEmailService)
	logger := pkgLogger.New("test", "error")

	useCase := NewRequestPasswordResetUseCase(
		mockUserRepo,
		mockTokenRepo,
		mockEmailService,
		logger,
	)

	ctx := context.Background()
	email := "invalid-email"
	orgID := uuid.New()

	// When
	err := useCase.Execute(ctx, email, orgID)

	// Then
	assert.Error(t, err)
	assert.Equal(t, "invalid email format", err.(*pkgErrors.CustomError).Message)
	mockUserRepo.AssertNotCalled(t, "GetByEmailAndOrg")
}

func TestRequestPasswordResetUseCase_Execute_UserNotFound_ReturnsSuccessForSecurity(t *testing.T) {
	// Given
	mockUserRepo := new(providers.MockUserRepository)
	mockTokenRepo := new(providers.MockTokenRepository)
	mockEmailService := new(MockEmailService)
	logger := pkgLogger.New("test", "error")

	useCase := NewRequestPasswordResetUseCase(
		mockUserRepo,
		mockTokenRepo,
		mockEmailService,
		logger,
	)

	ctx := context.Background()
	email := "nonexistent@example.com"
	orgID := uuid.New()

	mockUserRepo.On("GetByEmailAndOrg", ctx, email, orgID).Return(nil, pkgErrors.NewNotFound("user not found"))

	// When
	err := useCase.Execute(ctx, email, orgID)

	// Then - Should return success to avoid email enumeration
	assert.NoError(t, err)
	mockUserRepo.AssertExpectations(t)
	mockTokenRepo.AssertNotCalled(t, "StorePasswordResetToken")
	mockEmailService.AssertNotCalled(t, "SendPasswordResetEmail")
}

func TestRequestPasswordResetUseCase_Execute_TokenStorageFails_ReturnsError(t *testing.T) {
	// Given
	mockUserRepo := new(providers.MockUserRepository)
	mockTokenRepo := new(providers.MockTokenRepository)
	mockEmailService := new(MockEmailService)
	logger := pkgLogger.New("test", "error")

	useCase := NewRequestPasswordResetUseCase(
		mockUserRepo,
		mockTokenRepo,
		mockEmailService,
		logger,
	)

	ctx := context.Background()
	email := "test@example.com"
	orgID := uuid.New()
	userID := uuid.New()

	givenUser := &domain.User{
		ID:             userID,
		Email:          email,
		FirstName:      "Test",
		LastName:       "User",
		OrganizationID: orgID,
	}

	mockUserRepo.On("GetByEmailAndOrg", ctx, email, orgID).Return(givenUser, nil)
	mockTokenRepo.On("StorePasswordResetToken", ctx, mock.AnythingOfType("*domain.PasswordResetToken")).
		Return(pkgErrors.NewInternalServerError("database error"))

	// When
	err := useCase.Execute(ctx, email, orgID)

	// Then
	assert.Error(t, err)
	assert.Equal(t, "failed to initiate password reset", err.(*pkgErrors.CustomError).Message)
	mockUserRepo.AssertExpectations(t)
	mockTokenRepo.AssertExpectations(t)
	mockEmailService.AssertNotCalled(t, "SendPasswordResetEmail")
}

func TestRequestPasswordResetUseCase_Execute_EmailSendFails_StillReturnsSuccess(t *testing.T) {
	// Given
	mockUserRepo := new(providers.MockUserRepository)
	mockTokenRepo := new(providers.MockTokenRepository)
	mockEmailService := new(MockEmailService)
	logger := pkgLogger.New("test", "error")

	useCase := NewRequestPasswordResetUseCase(
		mockUserRepo,
		mockTokenRepo,
		mockEmailService,
		logger,
	)

	ctx := context.Background()
	email := "test@example.com"
	orgID := uuid.New()
	userID := uuid.New()

	givenUser := &domain.User{
		ID:             userID,
		Email:          email,
		FirstName:      "Test",
		LastName:       "User",
		OrganizationID: orgID,
	}

	mockUserRepo.On("GetByEmailAndOrg", ctx, email, orgID).Return(givenUser, nil)
	mockTokenRepo.On("StorePasswordResetToken", ctx, mock.AnythingOfType("*domain.PasswordResetToken")).Return(nil)
	mockEmailService.On("SendPasswordResetEmail", ctx, email, mock.AnythingOfType("string"), "Test").
		Return(pkgErrors.NewInternalServerError("SMTP error"))

	// When
	err := useCase.Execute(ctx, email, orgID)

	// Then - Should still return success, email failure is logged but not blocking
	assert.NoError(t, err)
	mockUserRepo.AssertExpectations(t)
	mockTokenRepo.AssertExpectations(t)
	mockEmailService.AssertExpectations(t)
}
