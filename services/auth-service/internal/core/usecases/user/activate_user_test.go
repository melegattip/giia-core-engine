package user

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	pkgErrors "github.com/melegattip/giia-core-engine/pkg/errors"
	pkgLogger "github.com/melegattip/giia-core-engine/pkg/logger"
	"github.com/melegattip/giia-core-engine/services/auth-service/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/auth-service/internal/core/providers"
)

func TestActivateUserUseCase_Execute_Success(t *testing.T) {
	// Given
	mockUserRepo := new(providers.MockUserRepository)
	mockPermissionRepo := new(providers.MockPermissionRepository)
	mockEventPublisher := new(providers.MockEventPublisher)
	mockTimeManager := new(providers.MockTimeManager)
	logger := pkgLogger.New("test", "error")

	useCase := NewActivateUserUseCase(
		mockUserRepo,
		mockPermissionRepo,
		mockEventPublisher,
		mockTimeManager,
		logger,
	)

	ctx := context.Background()
	adminUserID := uuid.New()
	targetUserID := uuid.New()
	orgID := uuid.New()
	fixedTime := time.Date(2025, 1, 18, 10, 0, 0, 0, time.UTC)

	givenPermissions := []*domain.Permission{
		{
			ID:   uuid.New(),
			Code: "users:activate",
		},
	}

	givenTargetUser := &domain.User{
		ID:             targetUserID,
		Email:          "target@example.com",
		FirstName:      "Target",
		LastName:       "User",
		Status:         domain.UserStatusInactive,
		OrganizationID: orgID,
	}

	mockPermissionRepo.On("GetUserPermissions", ctx, adminUserID).Return(givenPermissions, nil)
	mockUserRepo.On("GetByID", ctx, targetUserID).Return(givenTargetUser, nil)
	mockTimeManager.On("Now").Return(fixedTime)
	mockUserRepo.On("Update", ctx, mock.MatchedBy(func(user *domain.User) bool {
		return user.Status == domain.UserStatusActive && user.VerifiedAt != nil
	})).Return(nil)
	mockEventPublisher.On("PublishAsync", ctx, "auth.user.activated", mock.AnythingOfType("*events.Event")).Return(nil)

	// When
	err := useCase.Execute(ctx, adminUserID, targetUserID)

	// Then
	assert.NoError(t, err)
	mockPermissionRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
	mockTimeManager.AssertExpectations(t)
	mockEventPublisher.AssertExpectations(t)
}

func TestActivateUserUseCase_Execute_AdminUserIDNil_ReturnsError(t *testing.T) {
	// Given
	mockUserRepo := new(providers.MockUserRepository)
	mockPermissionRepo := new(providers.MockPermissionRepository)
	mockEventPublisher := new(providers.MockEventPublisher)
	mockTimeManager := new(providers.MockTimeManager)
	logger := pkgLogger.New("test", "error")

	useCase := NewActivateUserUseCase(
		mockUserRepo,
		mockPermissionRepo,
		mockEventPublisher,
		mockTimeManager,
		logger,
	)

	ctx := context.Background()
	adminUserID := uuid.Nil
	targetUserID := uuid.New()

	// When
	err := useCase.Execute(ctx, adminUserID, targetUserID)

	// Then
	assert.Error(t, err)
	assert.Equal(t, "admin user ID is required", err.(*pkgErrors.CustomError).Message)
	mockPermissionRepo.AssertNotCalled(t, "GetUserPermissions")
}

func TestActivateUserUseCase_Execute_TargetUserIDNil_ReturnsError(t *testing.T) {
	// Given
	mockUserRepo := new(providers.MockUserRepository)
	mockPermissionRepo := new(providers.MockPermissionRepository)
	mockEventPublisher := new(providers.MockEventPublisher)
	mockTimeManager := new(providers.MockTimeManager)
	logger := pkgLogger.New("test", "error")

	useCase := NewActivateUserUseCase(
		mockUserRepo,
		mockPermissionRepo,
		mockEventPublisher,
		mockTimeManager,
		logger,
	)

	ctx := context.Background()
	adminUserID := uuid.New()
	targetUserID := uuid.Nil

	// When
	err := useCase.Execute(ctx, adminUserID, targetUserID)

	// Then
	assert.Error(t, err)
	assert.Equal(t, "target user ID is required", err.(*pkgErrors.CustomError).Message)
	mockPermissionRepo.AssertNotCalled(t, "GetUserPermissions")
}

func TestActivateUserUseCase_Execute_InsufficientPermissions_ReturnsError(t *testing.T) {
	// Given
	mockUserRepo := new(providers.MockUserRepository)
	mockPermissionRepo := new(providers.MockPermissionRepository)
	mockEventPublisher := new(providers.MockEventPublisher)
	mockTimeManager := new(providers.MockTimeManager)
	logger := pkgLogger.New("test", "error")

	useCase := NewActivateUserUseCase(
		mockUserRepo,
		mockPermissionRepo,
		mockEventPublisher,
		mockTimeManager,
		logger,
	)

	ctx := context.Background()
	adminUserID := uuid.New()
	targetUserID := uuid.New()

	givenPermissions := []*domain.Permission{
		{
			ID:   uuid.New(),
			Code: "users:read", // Wrong permission
		},
	}

	mockPermissionRepo.On("GetUserPermissions", ctx, adminUserID).Return(givenPermissions, nil)

	// When
	err := useCase.Execute(ctx, adminUserID, targetUserID)

	// Then
	assert.Error(t, err)
	assert.Equal(t, "insufficient permissions to activate users", err.(*pkgErrors.CustomError).Message)
	mockPermissionRepo.AssertExpectations(t)
	mockUserRepo.AssertNotCalled(t, "GetByID")
}

func TestActivateUserUseCase_Execute_PermissionCheckFails_ReturnsError(t *testing.T) {
	// Given
	mockUserRepo := new(providers.MockUserRepository)
	mockPermissionRepo := new(providers.MockPermissionRepository)
	mockEventPublisher := new(providers.MockEventPublisher)
	mockTimeManager := new(providers.MockTimeManager)
	logger := pkgLogger.New("test", "error")

	useCase := NewActivateUserUseCase(
		mockUserRepo,
		mockPermissionRepo,
		mockEventPublisher,
		mockTimeManager,
		logger,
	)

	ctx := context.Background()
	adminUserID := uuid.New()
	targetUserID := uuid.New()

	mockPermissionRepo.On("GetUserPermissions", ctx, adminUserID).
		Return(nil, pkgErrors.NewInternalServerError("database error"))

	// When
	err := useCase.Execute(ctx, adminUserID, targetUserID)

	// Then
	assert.Error(t, err)
	assert.Equal(t, "failed to verify permissions", err.(*pkgErrors.CustomError).Message)
	mockPermissionRepo.AssertExpectations(t)
	mockUserRepo.AssertNotCalled(t, "GetByID")
}

func TestActivateUserUseCase_Execute_TargetUserNotFound_ReturnsError(t *testing.T) {
	// Given
	mockUserRepo := new(providers.MockUserRepository)
	mockPermissionRepo := new(providers.MockPermissionRepository)
	mockEventPublisher := new(providers.MockEventPublisher)
	mockTimeManager := new(providers.MockTimeManager)
	logger := pkgLogger.New("test", "error")

	useCase := NewActivateUserUseCase(
		mockUserRepo,
		mockPermissionRepo,
		mockEventPublisher,
		mockTimeManager,
		logger,
	)

	ctx := context.Background()
	adminUserID := uuid.New()
	targetUserID := uuid.New()

	givenPermissions := []*domain.Permission{
		{
			ID:   uuid.New(),
			Code: "users:activate",
		},
	}

	mockPermissionRepo.On("GetUserPermissions", ctx, adminUserID).Return(givenPermissions, nil)
	mockUserRepo.On("GetByID", ctx, targetUserID).Return(nil, pkgErrors.NewNotFound("user not found"))

	// When
	err := useCase.Execute(ctx, adminUserID, targetUserID)

	// Then
	assert.Error(t, err)
	assert.Equal(t, "user not found", err.(*pkgErrors.CustomError).Message)
	mockPermissionRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestActivateUserUseCase_Execute_AlreadyActive_ReturnsSuccess(t *testing.T) {
	// Given
	mockUserRepo := new(providers.MockUserRepository)
	mockPermissionRepo := new(providers.MockPermissionRepository)
	mockEventPublisher := new(providers.MockEventPublisher)
	mockTimeManager := new(providers.MockTimeManager)
	logger := pkgLogger.New("test", "error")

	useCase := NewActivateUserUseCase(
		mockUserRepo,
		mockPermissionRepo,
		mockEventPublisher,
		mockTimeManager,
		logger,
	)

	ctx := context.Background()
	adminUserID := uuid.New()
	targetUserID := uuid.New()

	givenPermissions := []*domain.Permission{
		{
			ID:   uuid.New(),
			Code: "users:activate",
		},
	}

	givenTargetUser := &domain.User{
		ID:     targetUserID,
		Email:  "target@example.com",
		Status: domain.UserStatusActive, // Already active
	}

	mockPermissionRepo.On("GetUserPermissions", ctx, adminUserID).Return(givenPermissions, nil)
	mockUserRepo.On("GetByID", ctx, targetUserID).Return(givenTargetUser, nil)

	// When
	err := useCase.Execute(ctx, adminUserID, targetUserID)

	// Then
	assert.NoError(t, err)
	mockPermissionRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
	mockUserRepo.AssertNotCalled(t, "Update")
}

func TestActivateUserUseCase_Execute_UpdateUserFails_ReturnsError(t *testing.T) {
	// Given
	mockUserRepo := new(providers.MockUserRepository)
	mockPermissionRepo := new(providers.MockPermissionRepository)
	mockEventPublisher := new(providers.MockEventPublisher)
	mockTimeManager := new(providers.MockTimeManager)
	logger := pkgLogger.New("test", "error")

	useCase := NewActivateUserUseCase(
		mockUserRepo,
		mockPermissionRepo,
		mockEventPublisher,
		mockTimeManager,
		logger,
	)

	ctx := context.Background()
	adminUserID := uuid.New()
	targetUserID := uuid.New()
	orgID := uuid.New()
	fixedTime := time.Date(2025, 1, 18, 10, 0, 0, 0, time.UTC)

	givenPermissions := []*domain.Permission{
		{
			ID:   uuid.New(),
			Code: "users:activate",
		},
	}

	givenTargetUser := &domain.User{
		ID:             targetUserID,
		Email:          "target@example.com",
		Status:         domain.UserStatusInactive,
		OrganizationID: orgID,
	}

	mockPermissionRepo.On("GetUserPermissions", ctx, adminUserID).Return(givenPermissions, nil)
	mockUserRepo.On("GetByID", ctx, targetUserID).Return(givenTargetUser, nil)
	mockTimeManager.On("Now").Return(fixedTime)
	mockUserRepo.On("Update", ctx, mock.AnythingOfType("*domain.User")).
		Return(pkgErrors.NewInternalServerError("database error"))

	// When
	err := useCase.Execute(ctx, adminUserID, targetUserID)

	// Then
	assert.Error(t, err)
	assert.Equal(t, "failed to activate user", err.(*pkgErrors.CustomError).Message)
	mockPermissionRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
	mockTimeManager.AssertExpectations(t)
	mockEventPublisher.AssertNotCalled(t, "PublishAsync")
}
