package user

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	pkgErrors "github.com/giia/giia-core-engine/pkg/errors"
	pkgLogger "github.com/giia/giia-core-engine/pkg/logger"
	"github.com/giia/giia-core-engine/services/auth-service/internal/core/domain"
	"github.com/giia/giia-core-engine/services/auth-service/internal/core/providers"
)

func TestDeactivateUserUseCase_Execute_Success(t *testing.T) {
	// Given
	mockUserRepo := new(providers.MockUserRepository)
	mockPermissionRepo := new(providers.MockPermissionRepository)
	mockEventPublisher := new(providers.MockEventPublisher)
	mockTimeManager := new(providers.MockTimeManager)
	logger := pkgLogger.New("test", "error")

	useCase := NewDeactivateUserUseCase(
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
			Code: "users:deactivate",
		},
	}

	givenTargetUser := &domain.User{
		ID:             targetUserID,
		Email:          "target@example.com",
		FirstName:      "Target",
		LastName:       "User",
		Status:         domain.UserStatusActive,
		OrganizationID: orgID,
	}

	mockPermissionRepo.On("GetUserPermissions", ctx, adminUserID).Return(givenPermissions, nil)
	mockUserRepo.On("GetByID", ctx, targetUserID).Return(givenTargetUser, nil)
	mockUserRepo.On("Update", ctx, mock.MatchedBy(func(user *domain.User) bool {
		return user.Status == domain.UserStatusInactive
	})).Return(nil)
	mockTimeManager.On("Now").Return(fixedTime)
	mockEventPublisher.On("PublishAsync", ctx, "auth.user.deactivated", mock.AnythingOfType("*events.Event")).Return(nil)

	// When
	err := useCase.Execute(ctx, adminUserID, targetUserID)

	// Then
	assert.NoError(t, err)
	mockPermissionRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
	mockTimeManager.AssertExpectations(t)
	mockEventPublisher.AssertExpectations(t)
}

func TestDeactivateUserUseCase_Execute_AdminUserIDNil_ReturnsError(t *testing.T) {
	// Given
	mockUserRepo := new(providers.MockUserRepository)
	mockPermissionRepo := new(providers.MockPermissionRepository)
	mockEventPublisher := new(providers.MockEventPublisher)
	mockTimeManager := new(providers.MockTimeManager)
	logger := pkgLogger.New("test", "error")

	useCase := NewDeactivateUserUseCase(
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

func TestDeactivateUserUseCase_Execute_TargetUserIDNil_ReturnsError(t *testing.T) {
	// Given
	mockUserRepo := new(providers.MockUserRepository)
	mockPermissionRepo := new(providers.MockPermissionRepository)
	mockEventPublisher := new(providers.MockEventPublisher)
	mockTimeManager := new(providers.MockTimeManager)
	logger := pkgLogger.New("test", "error")

	useCase := NewDeactivateUserUseCase(
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

func TestDeactivateUserUseCase_Execute_SelfDeactivation_ReturnsError(t *testing.T) {
	// Given
	mockUserRepo := new(providers.MockUserRepository)
	mockPermissionRepo := new(providers.MockPermissionRepository)
	mockEventPublisher := new(providers.MockEventPublisher)
	mockTimeManager := new(providers.MockTimeManager)
	logger := pkgLogger.New("test", "error")

	useCase := NewDeactivateUserUseCase(
		mockUserRepo,
		mockPermissionRepo,
		mockEventPublisher,
		mockTimeManager,
		logger,
	)

	ctx := context.Background()
	sameUserID := uuid.New()

	// When
	err := useCase.Execute(ctx, sameUserID, sameUserID)

	// Then
	assert.Error(t, err)
	assert.Equal(t, "cannot deactivate your own account", err.(*pkgErrors.CustomError).Message)
	mockPermissionRepo.AssertNotCalled(t, "GetUserPermissions")
}

func TestDeactivateUserUseCase_Execute_InsufficientPermissions_ReturnsError(t *testing.T) {
	// Given
	mockUserRepo := new(providers.MockUserRepository)
	mockPermissionRepo := new(providers.MockPermissionRepository)
	mockEventPublisher := new(providers.MockEventPublisher)
	mockTimeManager := new(providers.MockTimeManager)
	logger := pkgLogger.New("test", "error")

	useCase := NewDeactivateUserUseCase(
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
	assert.Equal(t, "insufficient permissions to deactivate users", err.(*pkgErrors.CustomError).Message)
	mockPermissionRepo.AssertExpectations(t)
	mockUserRepo.AssertNotCalled(t, "GetByID")
}

func TestDeactivateUserUseCase_Execute_PermissionCheckFails_ReturnsError(t *testing.T) {
	// Given
	mockUserRepo := new(providers.MockUserRepository)
	mockPermissionRepo := new(providers.MockPermissionRepository)
	mockEventPublisher := new(providers.MockEventPublisher)
	mockTimeManager := new(providers.MockTimeManager)
	logger := pkgLogger.New("test", "error")

	useCase := NewDeactivateUserUseCase(
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

func TestDeactivateUserUseCase_Execute_TargetUserNotFound_ReturnsError(t *testing.T) {
	// Given
	mockUserRepo := new(providers.MockUserRepository)
	mockPermissionRepo := new(providers.MockPermissionRepository)
	mockEventPublisher := new(providers.MockEventPublisher)
	mockTimeManager := new(providers.MockTimeManager)
	logger := pkgLogger.New("test", "error")

	useCase := NewDeactivateUserUseCase(
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
			Code: "users:deactivate",
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

func TestDeactivateUserUseCase_Execute_AlreadyInactive_ReturnsSuccess(t *testing.T) {
	// Given
	mockUserRepo := new(providers.MockUserRepository)
	mockPermissionRepo := new(providers.MockPermissionRepository)
	mockEventPublisher := new(providers.MockEventPublisher)
	mockTimeManager := new(providers.MockTimeManager)
	logger := pkgLogger.New("test", "error")

	useCase := NewDeactivateUserUseCase(
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
			Code: "users:deactivate",
		},
	}

	givenTargetUser := &domain.User{
		ID:     targetUserID,
		Email:  "target@example.com",
		Status: domain.UserStatusInactive, // Already inactive
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

func TestDeactivateUserUseCase_Execute_UpdateUserFails_ReturnsError(t *testing.T) {
	// Given
	mockUserRepo := new(providers.MockUserRepository)
	mockPermissionRepo := new(providers.MockPermissionRepository)
	mockEventPublisher := new(providers.MockEventPublisher)
	mockTimeManager := new(providers.MockTimeManager)
	logger := pkgLogger.New("test", "error")

	useCase := NewDeactivateUserUseCase(
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

	givenPermissions := []*domain.Permission{
		{
			ID:   uuid.New(),
			Code: "users:deactivate",
		},
	}

	givenTargetUser := &domain.User{
		ID:             targetUserID,
		Email:          "target@example.com",
		Status:         domain.UserStatusActive,
		OrganizationID: orgID,
	}

	mockPermissionRepo.On("GetUserPermissions", ctx, adminUserID).Return(givenPermissions, nil)
	mockUserRepo.On("GetByID", ctx, targetUserID).Return(givenTargetUser, nil)
	mockUserRepo.On("Update", ctx, mock.AnythingOfType("*domain.User")).
		Return(pkgErrors.NewInternalServerError("database error"))

	// When
	err := useCase.Execute(ctx, adminUserID, targetUserID)

	// Then
	assert.Error(t, err)
	assert.Equal(t, "failed to deactivate user", err.(*pkgErrors.CustomError).Message)
	mockPermissionRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
	mockEventPublisher.AssertNotCalled(t, "PublishAsync")
}

func TestDeactivateUserUseCase_Execute_EventPublishFails_StillReturnsSuccess(t *testing.T) {
	// Given
	mockUserRepo := new(providers.MockUserRepository)
	mockPermissionRepo := new(providers.MockPermissionRepository)
	mockEventPublisher := new(providers.MockEventPublisher)
	mockTimeManager := new(providers.MockTimeManager)
	logger := pkgLogger.New("test", "error")

	useCase := NewDeactivateUserUseCase(
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
			Code: "users:deactivate",
		},
	}

	givenTargetUser := &domain.User{
		ID:             targetUserID,
		Email:          "target@example.com",
		Status:         domain.UserStatusActive,
		OrganizationID: orgID,
	}

	mockPermissionRepo.On("GetUserPermissions", ctx, adminUserID).Return(givenPermissions, nil)
	mockUserRepo.On("GetByID", ctx, targetUserID).Return(givenTargetUser, nil)
	mockUserRepo.On("Update", ctx, mock.AnythingOfType("*domain.User")).Return(nil)
	mockTimeManager.On("Now").Return(fixedTime)
	mockEventPublisher.On("PublishAsync", ctx, "auth.user.deactivated", mock.AnythingOfType("*events.Event")).
		Return(pkgErrors.NewInternalServerError("NATS error"))

	// When
	err := useCase.Execute(ctx, adminUserID, targetUserID)

	// Then - Should still return success, event publishing failure is logged
	assert.NoError(t, err)
	mockPermissionRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
	mockTimeManager.AssertExpectations(t)
	mockEventPublisher.AssertExpectations(t)
}
