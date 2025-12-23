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

func TestCheckPermissionUseCase_Execute_WithExactMatchPermission_ReturnsTrue(t *testing.T) {
	// Given
	givenUserID := uuid.New()
	givenRequiredPermission := "catalog:products:read"
	givenUserPermissions := []string{"catalog:products:read", "catalog:products:write"}

	mockRoleRepo := new(providers.MockRoleRepository)
	mockPermRepo := new(providers.MockPermissionRepository)
	mockCache := new(providers.MockPermissionCache)
	mockLogger := new(providers.MockLogger)

	resolveInheritanceUC := NewResolveInheritanceUseCase(mockRoleRepo, mockPermRepo, mockLogger)
	getUserPermsUC := NewGetUserPermissionsUseCase(mockRoleRepo, resolveInheritanceUC, mockCache, mockLogger)
	useCase := NewCheckPermissionUseCase(getUserPermsUC, mockLogger)

	mockCache.On("GetUserPermissions", mock.Anything, givenUserID.String()).Return(givenUserPermissions, nil)
	mockLogger.On("Debug", mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	allowed, err := useCase.Execute(context.Background(), givenUserID, givenRequiredPermission)

	// Then
	assert.NoError(t, err)
	assert.True(t, allowed)
}

func TestCheckPermissionUseCase_Execute_WithWildcardAllPermission_ReturnsTrue(t *testing.T) {
	// Given
	givenUserID := uuid.New()
	givenRequiredPermission := "catalog:products:read"
	givenUserPermissions := []string{"*:*:*"} // Admin wildcard

	mockRoleRepo := new(providers.MockRoleRepository)
	mockPermRepo := new(providers.MockPermissionRepository)
	mockCache := new(providers.MockPermissionCache)
	mockLogger := new(providers.MockLogger)

	resolveInheritanceUC := NewResolveInheritanceUseCase(mockRoleRepo, mockPermRepo, mockLogger)
	getUserPermsUC := NewGetUserPermissionsUseCase(mockRoleRepo, resolveInheritanceUC, mockCache, mockLogger)
	useCase := NewCheckPermissionUseCase(getUserPermsUC, mockLogger)

	mockCache.On("GetUserPermissions", mock.Anything, givenUserID.String()).Return(givenUserPermissions, nil)
	mockLogger.On("Debug", mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	allowed, err := useCase.Execute(context.Background(), givenUserID, givenRequiredPermission)

	// Then
	assert.NoError(t, err)
	assert.True(t, allowed)

	mockCache.AssertExpectations(t)
}

func TestCheckPermissionUseCase_Execute_WithServiceWildcardPermission_ReturnsTrue(t *testing.T) {
	// Given
	givenUserID := uuid.New()
	givenRequiredPermission := "catalog:products:read"
	givenUserPermissions := []string{"catalog:*:*"} // All catalog permissions

	mockRoleRepo := new(providers.MockRoleRepository)
	mockPermRepo := new(providers.MockPermissionRepository)
	mockCache := new(providers.MockPermissionCache)
	mockLogger := new(providers.MockLogger)

	resolveInheritanceUC := NewResolveInheritanceUseCase(mockRoleRepo, mockPermRepo, mockLogger)
	getUserPermsUC := NewGetUserPermissionsUseCase(mockRoleRepo, resolveInheritanceUC, mockCache, mockLogger)
	useCase := NewCheckPermissionUseCase(getUserPermsUC, mockLogger)

	mockCache.On("GetUserPermissions", mock.Anything, givenUserID.String()).Return(givenUserPermissions, nil)
	mockLogger.On("Debug", mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	allowed, err := useCase.Execute(context.Background(), givenUserID, givenRequiredPermission)

	// Then
	assert.NoError(t, err)
	assert.True(t, allowed)

	mockCache.AssertExpectations(t)
}

func TestCheckPermissionUseCase_Execute_WithResourceWildcardPermission_ReturnsTrue(t *testing.T) {
	// Given
	givenUserID := uuid.New()
	givenRequiredPermission := "catalog:products:read"
	givenUserPermissions := []string{"catalog:products:*"} // All actions on products

	mockRoleRepo := new(providers.MockRoleRepository)
	mockPermRepo := new(providers.MockPermissionRepository)
	mockCache := new(providers.MockPermissionCache)
	mockLogger := new(providers.MockLogger)

	resolveInheritanceUC := NewResolveInheritanceUseCase(mockRoleRepo, mockPermRepo, mockLogger)
	getUserPermsUC := NewGetUserPermissionsUseCase(mockRoleRepo, resolveInheritanceUC, mockCache, mockLogger)
	useCase := NewCheckPermissionUseCase(getUserPermsUC, mockLogger)

	mockCache.On("GetUserPermissions", mock.Anything, givenUserID.String()).Return(givenUserPermissions, nil)
	mockLogger.On("Debug", mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	allowed, err := useCase.Execute(context.Background(), givenUserID, givenRequiredPermission)

	// Then
	assert.NoError(t, err)
	assert.True(t, allowed)

	mockCache.AssertExpectations(t)
}

func TestCheckPermissionUseCase_Execute_WithoutPermission_ReturnsFalse(t *testing.T) {
	// Given
	givenUserID := uuid.New()
	givenRequiredPermission := "catalog:products:write"
	givenUserPermissions := []string{"catalog:products:read"} // Only read, not write

	mockRoleRepo := new(providers.MockRoleRepository)
	mockPermRepo := new(providers.MockPermissionRepository)
	mockCache := new(providers.MockPermissionCache)
	mockLogger := new(providers.MockLogger)

	resolveInheritanceUC := NewResolveInheritanceUseCase(mockRoleRepo, mockPermRepo, mockLogger)
	getUserPermsUC := NewGetUserPermissionsUseCase(mockRoleRepo, resolveInheritanceUC, mockCache, mockLogger)
	useCase := NewCheckPermissionUseCase(getUserPermsUC, mockLogger)

	mockCache.On("GetUserPermissions", mock.Anything, givenUserID.String()).Return(givenUserPermissions, nil)
	mockLogger.On("Debug", mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	allowed, err := useCase.Execute(context.Background(), givenUserID, givenRequiredPermission)

	// Then
	assert.NoError(t, err)
	assert.False(t, allowed)

	mockCache.AssertExpectations(t)
}

func TestCheckPermissionUseCase_Execute_WithNoPermissions_ReturnsFalse(t *testing.T) {
	// Given
	givenUserID := uuid.New()
	givenRequiredPermission := "catalog:products:read"
	givenUserPermissions := []string{} // No permissions

	mockRoleRepo := new(providers.MockRoleRepository)
	mockPermRepo := new(providers.MockPermissionRepository)
	mockCache := new(providers.MockPermissionCache)
	mockLogger := new(providers.MockLogger)

	resolveInheritanceUC := NewResolveInheritanceUseCase(mockRoleRepo, mockPermRepo, mockLogger)
	getUserPermsUC := NewGetUserPermissionsUseCase(mockRoleRepo, resolveInheritanceUC, mockCache, mockLogger)
	useCase := NewCheckPermissionUseCase(getUserPermsUC, mockLogger)

	mockCache.On("GetUserPermissions", mock.Anything, givenUserID.String()).Return(givenUserPermissions, nil)
	mockLogger.On("Debug", mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	allowed, err := useCase.Execute(context.Background(), givenUserID, givenRequiredPermission)

	// Then
	assert.NoError(t, err)
	assert.False(t, allowed)

	mockCache.AssertExpectations(t)
}

func TestCheckPermissionUseCase_Execute_WhenGetPermissionsFails_ReturnsError(t *testing.T) {
	// Given
	givenUserID := uuid.New()
	givenRequiredPermission := "catalog:products:read"
	givenError := assert.AnError

	mockRoleRepo := new(providers.MockRoleRepository)
	mockPermRepo := new(providers.MockPermissionRepository)
	mockCache := new(providers.MockPermissionCache)
	mockLogger := new(providers.MockLogger)

	resolveInheritanceUC := NewResolveInheritanceUseCase(mockRoleRepo, mockPermRepo, mockLogger)
	getUserPermsUC := NewGetUserPermissionsUseCase(mockRoleRepo, resolveInheritanceUC, mockCache, mockLogger)
	useCase := NewCheckPermissionUseCase(getUserPermsUC, mockLogger)

	mockCache.On("GetUserPermissions", mock.Anything, givenUserID.String()).Return([]string(nil), givenError)
	mockRoleRepo.On("GetUserRoles", mock.Anything, givenUserID).Return([]*domain.Role(nil), givenError)
	mockLogger.On("Debug", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, givenError, mock.Anything, mock.Anything).Return()

	// When
	allowed, err := useCase.Execute(context.Background(), givenUserID, givenRequiredPermission)

	// Then
	assert.Error(t, err)
	assert.False(t, allowed)

	mockCache.AssertExpectations(t)
}

func TestCheckPermissionUseCase_Execute_WithMultipleWildcards_ChoosesMostSpecific(t *testing.T) {
	// Given
	givenUserID := uuid.New()
	givenRequiredPermission := "catalog:products:read"
	givenUserPermissions := []string{
		"*:*:*",                 // Admin
		"catalog:*:*",           // Catalog admin
		"catalog:products:*",    // Product admin
		"catalog:products:read", // Specific read
	}

	mockRoleRepo := new(providers.MockRoleRepository)
	mockPermRepo := new(providers.MockPermissionRepository)
	mockCache := new(providers.MockPermissionCache)
	mockLogger := new(providers.MockLogger)

	resolveInheritanceUC := NewResolveInheritanceUseCase(mockRoleRepo, mockPermRepo, mockLogger)
	getUserPermsUC := NewGetUserPermissionsUseCase(mockRoleRepo, resolveInheritanceUC, mockCache, mockLogger)
	useCase := NewCheckPermissionUseCase(getUserPermsUC, mockLogger)

	mockCache.On("GetUserPermissions", mock.Anything, givenUserID.String()).Return(givenUserPermissions, nil)
	mockLogger.On("Debug", mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	allowed, err := useCase.Execute(context.Background(), givenUserID, givenRequiredPermission)

	// Then
	assert.NoError(t, err)
	assert.True(t, allowed)

	mockCache.AssertExpectations(t)
}
