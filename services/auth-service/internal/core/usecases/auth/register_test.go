package auth

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"

	"github.com/melegattip/giia-core-engine/services/auth-service/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/auth-service/internal/core/providers"
)

func TestRegisterUseCase_Execute_WithValidData_CreatesUser(t *testing.T) {
	// Given
	givenOrgID := uuid.New()
	givenRequest := &domain.RegisterRequest{
		Email:          "user@example.com",
		Password:       "Password123!",
		FirstName:      "John",
		LastName:       "Doe",
		Phone:          "+1234567890",
		OrganizationID: givenOrgID.String(),
	}

	givenOrganization := &domain.Organization{
		ID:   givenOrgID,
		Name: "Test Org",
	}

	mockUserRepo := new(providers.MockUserRepository)
	mockOrgRepo := new(providers.MockOrganizationRepository)
	mockTokenRepo := new(providers.MockTokenRepository)
	mockEventPublisher := new(providers.MockEventPublisher)
	mockTimeManager := new(providers.MockTimeManager)
	mockLogger := new(providers.MockLogger)

	useCase := NewRegisterUseCase(mockUserRepo, mockOrgRepo, mockTokenRepo, mockEventPublisher, mockTimeManager, mockLogger)

	mockOrgRepo.On("GetByID", mock.Anything, givenOrgID).Return(givenOrganization, nil)
	mockUserRepo.On("GetByEmailAndOrg", mock.Anything, givenRequest.Email, givenOrgID).Return((*domain.User)(nil), gorm.ErrRecordNotFound)
	mockUserRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil)
	mockTokenRepo.On("StoreActivationToken", mock.Anything, mock.AnythingOfType("*domain.ActivationToken")).Return(nil)
	mockTimeManager.On("Now").Return(time.Now())
	mockEventPublisher.On("PublishAsync", mock.Anything, "auth.user.created", mock.Anything).Return(nil)
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	err := useCase.Execute(context.Background(), givenRequest)

	// Then
	assert.NoError(t, err)
	mockOrgRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
	mockTokenRepo.AssertExpectations(t)
}

func TestRegisterUseCase_Execute_WithEmptyEmail_ReturnsBadRequest(t *testing.T) {
	// Given
	givenRequest := &domain.RegisterRequest{
		Email:          "",
		Password:       "Password123!",
		FirstName:      "John",
		LastName:       "Doe",
		OrganizationID: uuid.New().String(),
	}

	mockUserRepo := new(providers.MockUserRepository)
	mockOrgRepo := new(providers.MockOrganizationRepository)
	mockTokenRepo := new(providers.MockTokenRepository)
	mockEventPublisher := new(providers.MockEventPublisher)
	mockTimeManager := new(providers.MockTimeManager)
	mockLogger := new(providers.MockLogger)

	useCase := NewRegisterUseCase(mockUserRepo, mockOrgRepo, mockTokenRepo, mockEventPublisher, mockTimeManager, mockLogger)

	// When
	err := useCase.Execute(context.Background(), givenRequest)

	// Then
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "email is required")
}

func TestRegisterUseCase_Execute_WithEmptyPassword_ReturnsBadRequest(t *testing.T) {
	// Given
	givenRequest := &domain.RegisterRequest{
		Email:          "user@example.com",
		Password:       "",
		FirstName:      "John",
		LastName:       "Doe",
		OrganizationID: uuid.New().String(),
	}

	mockUserRepo := new(providers.MockUserRepository)
	mockOrgRepo := new(providers.MockOrganizationRepository)
	mockTokenRepo := new(providers.MockTokenRepository)
	mockEventPublisher := new(providers.MockEventPublisher)
	mockTimeManager := new(providers.MockTimeManager)
	mockLogger := new(providers.MockLogger)

	useCase := NewRegisterUseCase(mockUserRepo, mockOrgRepo, mockTokenRepo, mockEventPublisher, mockTimeManager, mockLogger)

	// When
	err := useCase.Execute(context.Background(), givenRequest)

	// Then
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "password is required")
}

func TestRegisterUseCase_Execute_WithEmptyFirstName_ReturnsBadRequest(t *testing.T) {
	// Given
	givenRequest := &domain.RegisterRequest{
		Email:          "user@example.com",
		Password:       "Password123!",
		FirstName:      "",
		LastName:       "Doe",
		OrganizationID: uuid.New().String(),
	}

	mockUserRepo := new(providers.MockUserRepository)
	mockOrgRepo := new(providers.MockOrganizationRepository)
	mockTokenRepo := new(providers.MockTokenRepository)
	mockEventPublisher := new(providers.MockEventPublisher)
	mockTimeManager := new(providers.MockTimeManager)
	mockLogger := new(providers.MockLogger)

	useCase := NewRegisterUseCase(mockUserRepo, mockOrgRepo, mockTokenRepo, mockEventPublisher, mockTimeManager, mockLogger)

	// When
	err := useCase.Execute(context.Background(), givenRequest)

	// Then
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "first name is required")
}

func TestRegisterUseCase_Execute_WithEmptyLastName_ReturnsBadRequest(t *testing.T) {
	// Given
	givenRequest := &domain.RegisterRequest{
		Email:          "user@example.com",
		Password:       "Password123!",
		FirstName:      "John",
		LastName:       "",
		OrganizationID: uuid.New().String(),
	}

	mockUserRepo := new(providers.MockUserRepository)
	mockOrgRepo := new(providers.MockOrganizationRepository)
	mockTokenRepo := new(providers.MockTokenRepository)
	mockEventPublisher := new(providers.MockEventPublisher)
	mockTimeManager := new(providers.MockTimeManager)
	mockLogger := new(providers.MockLogger)

	useCase := NewRegisterUseCase(mockUserRepo, mockOrgRepo, mockTokenRepo, mockEventPublisher, mockTimeManager, mockLogger)

	// When
	err := useCase.Execute(context.Background(), givenRequest)

	// Then
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "last name is required")
}

func TestRegisterUseCase_Execute_WithEmptyOrganizationID_ReturnsBadRequest(t *testing.T) {
	// Given
	givenRequest := &domain.RegisterRequest{
		Email:          "user@example.com",
		Password:       "Password123!",
		FirstName:      "John",
		LastName:       "Doe",
		OrganizationID: "",
	}

	mockUserRepo := new(providers.MockUserRepository)
	mockOrgRepo := new(providers.MockOrganizationRepository)
	mockTokenRepo := new(providers.MockTokenRepository)
	mockEventPublisher := new(providers.MockEventPublisher)
	mockTimeManager := new(providers.MockTimeManager)
	mockLogger := new(providers.MockLogger)

	useCase := NewRegisterUseCase(mockUserRepo, mockOrgRepo, mockTokenRepo, mockEventPublisher, mockTimeManager, mockLogger)

	// When
	err := useCase.Execute(context.Background(), givenRequest)

	// Then
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "organization ID is required")
}

func TestRegisterUseCase_Execute_WithInvalidEmailFormat_ReturnsBadRequest(t *testing.T) {
	// Given
	givenRequest := &domain.RegisterRequest{
		Email:          "invalid-email",
		Password:       "Password123!",
		FirstName:      "John",
		LastName:       "Doe",
		OrganizationID: uuid.New().String(),
	}

	mockUserRepo := new(providers.MockUserRepository)
	mockOrgRepo := new(providers.MockOrganizationRepository)
	mockTokenRepo := new(providers.MockTokenRepository)
	mockEventPublisher := new(providers.MockEventPublisher)
	mockTimeManager := new(providers.MockTimeManager)
	mockLogger := new(providers.MockLogger)

	useCase := NewRegisterUseCase(mockUserRepo, mockOrgRepo, mockTokenRepo, mockEventPublisher, mockTimeManager, mockLogger)

	// When
	err := useCase.Execute(context.Background(), givenRequest)

	// Then
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid email format")
}

func TestRegisterUseCase_Execute_WithWeakPasswordTooShort_ReturnsBadRequest(t *testing.T) {
	// Given
	givenRequest := &domain.RegisterRequest{
		Email:          "user@example.com",
		Password:       "Pass1!",
		FirstName:      "John",
		LastName:       "Doe",
		OrganizationID: uuid.New().String(),
	}

	mockUserRepo := new(providers.MockUserRepository)
	mockOrgRepo := new(providers.MockOrganizationRepository)
	mockTokenRepo := new(providers.MockTokenRepository)
	mockEventPublisher := new(providers.MockEventPublisher)
	mockTimeManager := new(providers.MockTimeManager)
	mockLogger := new(providers.MockLogger)

	useCase := NewRegisterUseCase(mockUserRepo, mockOrgRepo, mockTokenRepo, mockEventPublisher, mockTimeManager, mockLogger)

	// When
	err := useCase.Execute(context.Background(), givenRequest)

	// Then
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "password must be at least 8 characters long")
}

func TestRegisterUseCase_Execute_WithPasswordMissingUppercase_ReturnsBadRequest(t *testing.T) {
	// Given
	givenRequest := &domain.RegisterRequest{
		Email:          "user@example.com",
		Password:       "password123!",
		FirstName:      "John",
		LastName:       "Doe",
		OrganizationID: uuid.New().String(),
	}

	mockUserRepo := new(providers.MockUserRepository)
	mockOrgRepo := new(providers.MockOrganizationRepository)
	mockTokenRepo := new(providers.MockTokenRepository)
	mockEventPublisher := new(providers.MockEventPublisher)
	mockTimeManager := new(providers.MockTimeManager)
	mockLogger := new(providers.MockLogger)

	useCase := NewRegisterUseCase(mockUserRepo, mockOrgRepo, mockTokenRepo, mockEventPublisher, mockTimeManager, mockLogger)

	// When
	err := useCase.Execute(context.Background(), givenRequest)

	// Then
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "password must contain at least one uppercase letter")
}

func TestRegisterUseCase_Execute_WithPasswordMissingLowercase_ReturnsBadRequest(t *testing.T) {
	// Given
	givenRequest := &domain.RegisterRequest{
		Email:          "user@example.com",
		Password:       "PASSWORD123!",
		FirstName:      "John",
		LastName:       "Doe",
		OrganizationID: uuid.New().String(),
	}

	mockUserRepo := new(providers.MockUserRepository)
	mockOrgRepo := new(providers.MockOrganizationRepository)
	mockTokenRepo := new(providers.MockTokenRepository)
	mockEventPublisher := new(providers.MockEventPublisher)
	mockTimeManager := new(providers.MockTimeManager)
	mockLogger := new(providers.MockLogger)

	useCase := NewRegisterUseCase(mockUserRepo, mockOrgRepo, mockTokenRepo, mockEventPublisher, mockTimeManager, mockLogger)

	// When
	err := useCase.Execute(context.Background(), givenRequest)

	// Then
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "password must contain at least one lowercase letter")
}

func TestRegisterUseCase_Execute_WithPasswordMissingNumber_ReturnsBadRequest(t *testing.T) {
	// Given
	givenRequest := &domain.RegisterRequest{
		Email:          "user@example.com",
		Password:       "Password!",
		FirstName:      "John",
		LastName:       "Doe",
		OrganizationID: uuid.New().String(),
	}

	mockUserRepo := new(providers.MockUserRepository)
	mockOrgRepo := new(providers.MockOrganizationRepository)
	mockTokenRepo := new(providers.MockTokenRepository)
	mockEventPublisher := new(providers.MockEventPublisher)
	mockTimeManager := new(providers.MockTimeManager)
	mockLogger := new(providers.MockLogger)

	useCase := NewRegisterUseCase(mockUserRepo, mockOrgRepo, mockTokenRepo, mockEventPublisher, mockTimeManager, mockLogger)

	// When
	err := useCase.Execute(context.Background(), givenRequest)

	// Then
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "password must contain at least one number")
}

func TestRegisterUseCase_Execute_WithPasswordMissingSpecialChar_ReturnsBadRequest(t *testing.T) {
	// Given
	givenRequest := &domain.RegisterRequest{
		Email:          "user@example.com",
		Password:       "Password123",
		FirstName:      "John",
		LastName:       "Doe",
		OrganizationID: uuid.New().String(),
	}

	mockUserRepo := new(providers.MockUserRepository)
	mockOrgRepo := new(providers.MockOrganizationRepository)
	mockTokenRepo := new(providers.MockTokenRepository)
	mockEventPublisher := new(providers.MockEventPublisher)
	mockTimeManager := new(providers.MockTimeManager)
	mockLogger := new(providers.MockLogger)

	useCase := NewRegisterUseCase(mockUserRepo, mockOrgRepo, mockTokenRepo, mockEventPublisher, mockTimeManager, mockLogger)

	// When
	err := useCase.Execute(context.Background(), givenRequest)

	// Then
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "password must contain at least one special character")
}

func TestRegisterUseCase_Execute_WithInvalidOrganizationIDFormat_ReturnsBadRequest(t *testing.T) {
	// Given
	givenRequest := &domain.RegisterRequest{
		Email:          "user@example.com",
		Password:       "Password123!",
		FirstName:      "John",
		LastName:       "Doe",
		OrganizationID: "invalid-uuid",
	}

	mockUserRepo := new(providers.MockUserRepository)
	mockOrgRepo := new(providers.MockOrganizationRepository)
	mockTokenRepo := new(providers.MockTokenRepository)
	mockEventPublisher := new(providers.MockEventPublisher)
	mockTimeManager := new(providers.MockTimeManager)
	mockLogger := new(providers.MockLogger)

	useCase := NewRegisterUseCase(mockUserRepo, mockOrgRepo, mockTokenRepo, mockEventPublisher, mockTimeManager, mockLogger)

	// When
	err := useCase.Execute(context.Background(), givenRequest)

	// Then
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid organization ID format")
}

func TestRegisterUseCase_Execute_WithNonExistentOrganization_ReturnsBadRequest(t *testing.T) {
	// Given
	givenOrgID := uuid.New()
	givenRequest := &domain.RegisterRequest{
		Email:          "user@example.com",
		Password:       "Password123!",
		FirstName:      "John",
		LastName:       "Doe",
		OrganizationID: givenOrgID.String(),
	}

	mockUserRepo := new(providers.MockUserRepository)
	mockOrgRepo := new(providers.MockOrganizationRepository)
	mockTokenRepo := new(providers.MockTokenRepository)
	mockEventPublisher := new(providers.MockEventPublisher)
	mockTimeManager := new(providers.MockTimeManager)
	mockLogger := new(providers.MockLogger)

	useCase := NewRegisterUseCase(mockUserRepo, mockOrgRepo, mockTokenRepo, mockEventPublisher, mockTimeManager, mockLogger)

	mockOrgRepo.On("GetByID", mock.Anything, givenOrgID).Return((*domain.Organization)(nil), gorm.ErrRecordNotFound)

	// When
	err := useCase.Execute(context.Background(), givenRequest)

	// Then
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "organization not found")
	mockOrgRepo.AssertExpectations(t)
}

func TestRegisterUseCase_Execute_WithDuplicateEmailInOrganization_ReturnsBadRequest(t *testing.T) {
	// Given
	givenOrgID := uuid.New()
	givenEmail := "existing@example.com"
	givenRequest := &domain.RegisterRequest{
		Email:          givenEmail,
		Password:       "Password123!",
		FirstName:      "John",
		LastName:       "Doe",
		OrganizationID: givenOrgID.String(),
	}

	givenOrganization := &domain.Organization{
		ID:   givenOrgID,
		Name: "Test Org",
	}

	givenExistingUser := &domain.User{
		ID:             uuid.New(),
		Email:          givenEmail,
		OrganizationID: givenOrgID,
	}

	mockUserRepo := new(providers.MockUserRepository)
	mockOrgRepo := new(providers.MockOrganizationRepository)
	mockTokenRepo := new(providers.MockTokenRepository)
	mockEventPublisher := new(providers.MockEventPublisher)
	mockTimeManager := new(providers.MockTimeManager)
	mockLogger := new(providers.MockLogger)

	useCase := NewRegisterUseCase(mockUserRepo, mockOrgRepo, mockTokenRepo, mockEventPublisher, mockTimeManager, mockLogger)

	mockOrgRepo.On("GetByID", mock.Anything, givenOrgID).Return(givenOrganization, nil)
	mockUserRepo.On("GetByEmailAndOrg", mock.Anything, givenEmail, givenOrgID).Return(givenExistingUser, nil)

	// When
	err := useCase.Execute(context.Background(), givenRequest)

	// Then
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "email already registered in this organization")
	mockOrgRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestRegisterUseCase_Execute_WhenUserCreationFails_ReturnsError(t *testing.T) {
	// Given
	givenOrgID := uuid.New()
	givenRequest := &domain.RegisterRequest{
		Email:          "user@example.com",
		Password:       "Password123!",
		FirstName:      "John",
		LastName:       "Doe",
		OrganizationID: givenOrgID.String(),
	}

	givenOrganization := &domain.Organization{
		ID:   givenOrgID,
		Name: "Test Org",
	}

	mockUserRepo := new(providers.MockUserRepository)
	mockOrgRepo := new(providers.MockOrganizationRepository)
	mockTokenRepo := new(providers.MockTokenRepository)
	mockEventPublisher := new(providers.MockEventPublisher)
	mockTimeManager := new(providers.MockTimeManager)
	mockLogger := new(providers.MockLogger)

	useCase := NewRegisterUseCase(mockUserRepo, mockOrgRepo, mockTokenRepo, mockEventPublisher, mockTimeManager, mockLogger)

	mockOrgRepo.On("GetByID", mock.Anything, givenOrgID).Return(givenOrganization, nil)
	mockUserRepo.On("GetByEmailAndOrg", mock.Anything, givenRequest.Email, givenOrgID).Return((*domain.User)(nil), gorm.ErrRecordNotFound)
	mockUserRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.User")).Return(assert.AnError)
	mockLogger.On("Error", mock.Anything, assert.AnError, mock.Anything, mock.Anything).Return()

	// When
	err := useCase.Execute(context.Background(), givenRequest)

	// Then
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create user")
	mockUserRepo.AssertExpectations(t)
}

func TestRegisterUseCase_Execute_WhenActivationTokenStorageFails_StillSucceeds(t *testing.T) {
	// Given
	givenOrgID := uuid.New()
	givenRequest := &domain.RegisterRequest{
		Email:          "user@example.com",
		Password:       "Password123!",
		FirstName:      "John",
		LastName:       "Doe",
		OrganizationID: givenOrgID.String(),
	}

	givenOrganization := &domain.Organization{
		ID:   givenOrgID,
		Name: "Test Org",
	}

	mockUserRepo := new(providers.MockUserRepository)
	mockOrgRepo := new(providers.MockOrganizationRepository)
	mockTokenRepo := new(providers.MockTokenRepository)
	mockEventPublisher := new(providers.MockEventPublisher)
	mockTimeManager := new(providers.MockTimeManager)
	mockLogger := new(providers.MockLogger)

	useCase := NewRegisterUseCase(mockUserRepo, mockOrgRepo, mockTokenRepo, mockEventPublisher, mockTimeManager, mockLogger)

	mockOrgRepo.On("GetByID", mock.Anything, givenOrgID).Return(givenOrganization, nil)
	mockUserRepo.On("GetByEmailAndOrg", mock.Anything, givenRequest.Email, givenOrgID).Return((*domain.User)(nil), gorm.ErrRecordNotFound)
	mockUserRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil)
	mockTokenRepo.On("StoreActivationToken", mock.Anything, mock.AnythingOfType("*domain.ActivationToken")).Return(assert.AnError)
	mockTimeManager.On("Now").Return(time.Now())
	mockEventPublisher.On("PublishAsync", mock.Anything, "auth.user.created", mock.Anything).Return(nil)
	mockLogger.On("Error", mock.Anything, assert.AnError, mock.Anything, mock.Anything).Return()
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	err := useCase.Execute(context.Background(), givenRequest)

	// Then
	assert.NoError(t, err)
	mockTokenRepo.AssertExpectations(t)
}
