package rbac

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/melegattip/giia-core-engine/services/auth-service/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/auth-service/internal/core/providers"
)

func TestBatchCheckPermissionsUseCase_Execute_WithAllPermissionsGranted_ReturnsAllTrue(t *testing.T) {
	// Given
	givenUserID := uuid.New()
	givenPermissions := []string{"user:read", "user:write", "post:read"}
	givenUserPermissions := []string{"user:read", "user:write", "post:read", "post:write"}

	mockRoleRepo := new(providers.MockRoleRepository)
	mockPermRepo := new(providers.MockPermissionRepository)
	mockCache := new(providers.MockPermissionCache)
	mockLogger := new(providers.MockLogger)

	resolveInheritanceUC := NewResolveInheritanceUseCase(mockRoleRepo, mockPermRepo, mockLogger)
	getUserPermsUC := NewGetUserPermissionsUseCase(mockRoleRepo, resolveInheritanceUC, mockCache, mockLogger)
	checkPermUC := NewCheckPermissionUseCase(getUserPermsUC, mockLogger)
	useCase := NewBatchCheckPermissionsUseCase(checkPermUC, mockLogger)

	mockCache.On("GetUserPermissions", mock.Anything, givenUserID.String()).Return(givenUserPermissions, nil)
	mockLogger.On("Debug", mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	results, err := useCase.Execute(context.Background(), givenUserID, givenPermissions)

	// Then
	assert.NoError(t, err)
	assert.Equal(t, 3, len(results))
	assert.True(t, results["user:read"])
	assert.True(t, results["user:write"])
	assert.True(t, results["post:read"])
	mockCache.AssertExpectations(t)
}

func TestBatchCheckPermissionsUseCase_Execute_WithSomePermissionsDenied_ReturnsPartialResults(t *testing.T) {
	// Given
	givenUserID := uuid.New()
	givenPermissions := []string{"user:read", "user:delete", "post:read"}
	givenUserPermissions := []string{"user:read", "post:read"}

	mockRoleRepo := new(providers.MockRoleRepository)
	mockPermRepo := new(providers.MockPermissionRepository)
	mockCache := new(providers.MockPermissionCache)
	mockLogger := new(providers.MockLogger)

	resolveInheritanceUC := NewResolveInheritanceUseCase(mockRoleRepo, mockPermRepo, mockLogger)
	getUserPermsUC := NewGetUserPermissionsUseCase(mockRoleRepo, resolveInheritanceUC, mockCache, mockLogger)
	checkPermUC := NewCheckPermissionUseCase(getUserPermsUC, mockLogger)
	useCase := NewBatchCheckPermissionsUseCase(checkPermUC, mockLogger)

	mockCache.On("GetUserPermissions", mock.Anything, givenUserID.String()).Return(givenUserPermissions, nil)
	mockLogger.On("Debug", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	results, err := useCase.Execute(context.Background(), givenUserID, givenPermissions)

	// Then
	assert.NoError(t, err)
	assert.Equal(t, 3, len(results))
	assert.True(t, results["user:read"])
	assert.False(t, results["user:delete"])
	assert.True(t, results["post:read"])
	mockCache.AssertExpectations(t)
}

func TestBatchCheckPermissionsUseCase_Execute_WithNilUserID_ReturnsBadRequest(t *testing.T) {
	// Given
	givenPermissions := []string{"user:read"}

	mockLogger := new(providers.MockLogger)

	useCase := NewBatchCheckPermissionsUseCase(nil, mockLogger)

	// When
	results, err := useCase.Execute(context.Background(), uuid.Nil, givenPermissions)

	// Then
	assert.Error(t, err)
	assert.Nil(t, results)
	assert.Contains(t, err.Error(), "user ID cannot be empty")
}

func TestBatchCheckPermissionsUseCase_Execute_WithEmptyPermissionsList_ReturnsBadRequest(t *testing.T) {
	// Given
	givenUserID := uuid.New()

	mockLogger := new(providers.MockLogger)

	useCase := NewBatchCheckPermissionsUseCase(nil, mockLogger)

	// When
	results, err := useCase.Execute(context.Background(), givenUserID, []string{})

	// Then
	assert.Error(t, err)
	assert.Nil(t, results)
	assert.Contains(t, err.Error(), "permissions list cannot be empty")
}

func TestBatchCheckPermissionsUseCase_Execute_WithWildcardPermission_ReturnsAllTrue(t *testing.T) {
	// Given
	givenUserID := uuid.New()
	givenPermissions := []string{"user:read", "user:write", "admin:delete"}
	givenUserPermissions := []string{"*:*:*"}

	mockRoleRepo := new(providers.MockRoleRepository)
	mockPermRepo := new(providers.MockPermissionRepository)
	mockCache := new(providers.MockPermissionCache)
	mockLogger := new(providers.MockLogger)

	resolveInheritanceUC := NewResolveInheritanceUseCase(mockRoleRepo, mockPermRepo, mockLogger)
	getUserPermsUC := NewGetUserPermissionsUseCase(mockRoleRepo, resolveInheritanceUC, mockCache, mockLogger)
	checkPermUC := NewCheckPermissionUseCase(getUserPermsUC, mockLogger)
	useCase := NewBatchCheckPermissionsUseCase(checkPermUC, mockLogger)

	mockCache.On("GetUserPermissions", mock.Anything, givenUserID.String()).Return(givenUserPermissions, nil)
	mockLogger.On("Debug", mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	results, err := useCase.Execute(context.Background(), givenUserID, givenPermissions)

	// Then
	assert.NoError(t, err)
	assert.Equal(t, 3, len(results))
	assert.True(t, results["user:read"])
	assert.True(t, results["user:write"])
	assert.True(t, results["admin:delete"])
	mockCache.AssertExpectations(t)
}

func TestBatchCheckPermissionsUseCase_Execute_WhenCheckPermissionFails_ReturnsError(t *testing.T) {
	// Given
	givenUserID := uuid.New()
	givenPermissions := []string{"user:read"}

	mockRoleRepo := new(providers.MockRoleRepository)
	mockPermRepo := new(providers.MockPermissionRepository)
	mockCache := new(providers.MockPermissionCache)
	mockLogger := new(providers.MockLogger)

	resolveInheritanceUC := NewResolveInheritanceUseCase(mockRoleRepo, mockPermRepo, mockLogger)
	getUserPermsUC := NewGetUserPermissionsUseCase(mockRoleRepo, resolveInheritanceUC, mockCache, mockLogger)
	checkPermUC := NewCheckPermissionUseCase(getUserPermsUC, mockLogger)
	useCase := NewBatchCheckPermissionsUseCase(checkPermUC, mockLogger)

	mockCache.On("GetUserPermissions", mock.Anything, givenUserID.String()).Return(([]string)(nil), assert.AnError)
	mockRoleRepo.On("GetUserRoles", mock.Anything, givenUserID).Return(([]*domain.Role)(nil), assert.AnError)
	mockLogger.On("Debug", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	results, err := useCase.Execute(context.Background(), givenUserID, givenPermissions)

	// Then
	assert.Error(t, err)
	assert.Nil(t, results)
	mockRoleRepo.AssertExpectations(t)
}

func TestBatchCheckPermissionsUseCase_Execute_WithSinglePermission_ReturnsSingleResult(t *testing.T) {
	// Given
	givenUserID := uuid.New()
	givenPermissions := []string{"user:read"}
	givenUserPermissions := []string{"user:read", "user:write"}

	mockRoleRepo := new(providers.MockRoleRepository)
	mockPermRepo := new(providers.MockPermissionRepository)
	mockCache := new(providers.MockPermissionCache)
	mockLogger := new(providers.MockLogger)

	resolveInheritanceUC := NewResolveInheritanceUseCase(mockRoleRepo, mockPermRepo, mockLogger)
	getUserPermsUC := NewGetUserPermissionsUseCase(mockRoleRepo, resolveInheritanceUC, mockCache, mockLogger)
	checkPermUC := NewCheckPermissionUseCase(getUserPermsUC, mockLogger)
	useCase := NewBatchCheckPermissionsUseCase(checkPermUC, mockLogger)

	mockCache.On("GetUserPermissions", mock.Anything, givenUserID.String()).Return(givenUserPermissions, nil)
	mockLogger.On("Debug", mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	results, err := useCase.Execute(context.Background(), givenUserID, givenPermissions)

	// Then
	assert.NoError(t, err)
	assert.Equal(t, 1, len(results))
	assert.True(t, results["user:read"])
	mockCache.AssertExpectations(t)
}

func TestBatchCheckPermissionsUseCase_Execute_WithDuplicatePermissions_ReturnsDeduplicatedResults(t *testing.T) {
	// Given
	givenUserID := uuid.New()
	givenPermissions := []string{"user:read", "user:read", "user:write"}
	givenUserPermissions := []string{"user:read", "user:write"}

	mockRoleRepo := new(providers.MockRoleRepository)
	mockPermRepo := new(providers.MockPermissionRepository)
	mockCache := new(providers.MockPermissionCache)
	mockLogger := new(providers.MockLogger)

	resolveInheritanceUC := NewResolveInheritanceUseCase(mockRoleRepo, mockPermRepo, mockLogger)
	getUserPermsUC := NewGetUserPermissionsUseCase(mockRoleRepo, resolveInheritanceUC, mockCache, mockLogger)
	checkPermUC := NewCheckPermissionUseCase(getUserPermsUC, mockLogger)
	useCase := NewBatchCheckPermissionsUseCase(checkPermUC, mockLogger)

	mockCache.On("GetUserPermissions", mock.Anything, givenUserID.String()).Return(givenUserPermissions, nil)
	mockLogger.On("Debug", mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	results, err := useCase.Execute(context.Background(), givenUserID, givenPermissions)

	// Then
	assert.NoError(t, err)
	assert.Equal(t, 2, len(results))
	assert.True(t, results["user:read"])
	assert.True(t, results["user:write"])
	mockCache.AssertExpectations(t)
}
