package rbac

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/melegattip/giia-core-engine/services/auth-service/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/auth-service/internal/core/providers"
)

func TestGetUserPermissionsUseCase_Execute_WithCachedPermissions_ReturnsCachedResults(t *testing.T) {
	// Given
	givenUserID := uuid.New()
	givenCachedPermissions := []string{"user:read", "user:write", "post:read"}

	mockRoleRepo := new(providers.MockRoleRepository)
	mockPermRepo := new(providers.MockPermissionRepository)
	mockCache := new(providers.MockPermissionCache)
	mockLogger := new(providers.MockLogger)

	resolveInheritanceUC := NewResolveInheritanceUseCase(mockRoleRepo, mockPermRepo, mockLogger)
	useCase := NewGetUserPermissionsUseCase(mockRoleRepo, resolveInheritanceUC, mockCache, mockLogger)

	mockCache.On("GetUserPermissions", mock.Anything, givenUserID.String()).Return(givenCachedPermissions, nil)
	mockLogger.On("Debug", mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	permissions, err := useCase.Execute(context.Background(), givenUserID)

	// Then
	assert.NoError(t, err)
	assert.Equal(t, givenCachedPermissions, permissions)
	assert.Equal(t, 3, len(permissions))
	mockCache.AssertExpectations(t)
	mockRoleRepo.AssertNotCalled(t, "GetUserRoles")
}

func TestGetUserPermissionsUseCase_Execute_WithNilUserID_ReturnsBadRequest(t *testing.T) {
	// Given
	mockRoleRepo := new(providers.MockRoleRepository)
	mockPermRepo := new(providers.MockPermissionRepository)
	mockCache := new(providers.MockPermissionCache)
	mockLogger := new(providers.MockLogger)

	resolveInheritanceUC := NewResolveInheritanceUseCase(mockRoleRepo, mockPermRepo, mockLogger)
	useCase := NewGetUserPermissionsUseCase(mockRoleRepo, resolveInheritanceUC, mockCache, mockLogger)

	// When
	permissions, err := useCase.Execute(context.Background(), uuid.Nil)

	// Then
	assert.Error(t, err)
	assert.Nil(t, permissions)
	assert.Contains(t, err.Error(), "user ID cannot be empty")
}

func TestGetUserPermissionsUseCase_Execute_WithCacheMiss_RetrievesFromDatabase(t *testing.T) {
	// Given
	givenUserID := uuid.New()
	givenRoleID := uuid.New()

	givenRole := &domain.Role{
		ID:   givenRoleID,
		Name: "user",
	}

	givenPermissions := []*domain.Permission{
		{ID: uuid.New(), Code: "user:read", Description: "Read Users"},
		{ID: uuid.New(), Code: "user:write", Description: "Write Users"},
	}

	mockRoleRepo := new(providers.MockRoleRepository)
	mockPermRepo := new(providers.MockPermissionRepository)
	mockCache := new(providers.MockPermissionCache)
	mockLogger := new(providers.MockLogger)

	resolveInheritanceUC := NewResolveInheritanceUseCase(mockRoleRepo, mockPermRepo, mockLogger)
	useCase := NewGetUserPermissionsUseCase(mockRoleRepo, resolveInheritanceUC, mockCache, mockLogger)

	mockCache.On("GetUserPermissions", mock.Anything, givenUserID.String()).Return(([]string)(nil), assert.AnError)
	mockRoleRepo.On("GetUserRoles", mock.Anything, givenUserID).Return([]*domain.Role{givenRole}, nil)
	mockRoleRepo.On("GetByID", mock.Anything, givenRoleID).Return(givenRole, nil)
	mockPermRepo.On("GetRolePermissions", mock.Anything, givenRoleID).Return(givenPermissions, nil)
	mockCache.On("SetUserPermissions", mock.Anything, givenUserID.String(), mock.AnythingOfType("[]string"), 5*time.Minute).Return(nil)
	mockLogger.On("Debug", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	permissions, err := useCase.Execute(context.Background(), givenUserID)

	// Then
	assert.NoError(t, err)
	assert.Equal(t, 2, len(permissions))
	assert.Contains(t, permissions, "user:read")
	assert.Contains(t, permissions, "user:write")
	mockRoleRepo.AssertExpectations(t)
	mockPermRepo.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestGetUserPermissionsUseCase_Execute_WithNoRoles_ReturnsEmptyList(t *testing.T) {
	// Given
	givenUserID := uuid.New()

	mockRoleRepo := new(providers.MockRoleRepository)
	mockPermRepo := new(providers.MockPermissionRepository)
	mockCache := new(providers.MockPermissionCache)
	mockLogger := new(providers.MockLogger)

	resolveInheritanceUC := NewResolveInheritanceUseCase(mockRoleRepo, mockPermRepo, mockLogger)
	useCase := NewGetUserPermissionsUseCase(mockRoleRepo, resolveInheritanceUC, mockCache, mockLogger)

	mockCache.On("GetUserPermissions", mock.Anything, givenUserID.String()).Return(([]string)(nil), assert.AnError)
	mockRoleRepo.On("GetUserRoles", mock.Anything, givenUserID).Return([]*domain.Role{}, nil)
	mockLogger.On("Debug", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	permissions, err := useCase.Execute(context.Background(), givenUserID)

	// Then
	assert.NoError(t, err)
	assert.NotNil(t, permissions)
	assert.Equal(t, 0, len(permissions))
	mockRoleRepo.AssertExpectations(t)
}

func TestGetUserPermissionsUseCase_Execute_WhenGetUserRolesFails_ReturnsError(t *testing.T) {
	// Given
	givenUserID := uuid.New()

	mockRoleRepo := new(providers.MockRoleRepository)
	mockPermRepo := new(providers.MockPermissionRepository)
	mockCache := new(providers.MockPermissionCache)
	mockLogger := new(providers.MockLogger)

	resolveInheritanceUC := NewResolveInheritanceUseCase(mockRoleRepo, mockPermRepo, mockLogger)
	useCase := NewGetUserPermissionsUseCase(mockRoleRepo, resolveInheritanceUC, mockCache, mockLogger)

	mockCache.On("GetUserPermissions", mock.Anything, givenUserID.String()).Return(([]string)(nil), assert.AnError)
	mockRoleRepo.On("GetUserRoles", mock.Anything, givenUserID).Return(([]*domain.Role)(nil), assert.AnError)
	mockLogger.On("Debug", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	permissions, err := useCase.Execute(context.Background(), givenUserID)

	// Then
	assert.Error(t, err)
	assert.Nil(t, permissions)
	assert.Contains(t, err.Error(), "failed to get user roles")
	mockRoleRepo.AssertExpectations(t)
}

func TestGetUserPermissionsUseCase_Execute_WithWildcardPermission_ReturnsOnlyWildcard(t *testing.T) {
	// Given
	givenUserID := uuid.New()
	givenRoleID := uuid.New()

	givenRole := &domain.Role{
		ID:   givenRoleID,
		Name: "superadmin",
	}

	givenPermissions := []*domain.Permission{
		{ID: uuid.New(), Code: "*:*:*", Description: "All Permissions"},
		{ID: uuid.New(), Code: "user:read", Description: "Read Users"},
	}

	mockRoleRepo := new(providers.MockRoleRepository)
	mockPermRepo := new(providers.MockPermissionRepository)
	mockCache := new(providers.MockPermissionCache)
	mockLogger := new(providers.MockLogger)

	resolveInheritanceUC := NewResolveInheritanceUseCase(mockRoleRepo, mockPermRepo, mockLogger)
	useCase := NewGetUserPermissionsUseCase(mockRoleRepo, resolveInheritanceUC, mockCache, mockLogger)

	mockCache.On("GetUserPermissions", mock.Anything, givenUserID.String()).Return(([]string)(nil), assert.AnError)
	mockRoleRepo.On("GetUserRoles", mock.Anything, givenUserID).Return([]*domain.Role{givenRole}, nil)
	mockRoleRepo.On("GetByID", mock.Anything, givenRoleID).Return(givenRole, nil)
	mockPermRepo.On("GetRolePermissions", mock.Anything, givenRoleID).Return(givenPermissions, nil)
	mockCache.On("SetUserPermissions", mock.Anything, givenUserID.String(), []string{"*:*:*"}, 5*time.Minute).Return(nil)
	mockLogger.On("Debug", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	permissions, err := useCase.Execute(context.Background(), givenUserID)

	// Then
	assert.NoError(t, err)
	assert.Equal(t, 1, len(permissions))
	assert.Contains(t, permissions, "*:*:*")
	mockRoleRepo.AssertExpectations(t)
	mockPermRepo.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestGetUserPermissionsUseCase_Execute_WithMultipleRoles_DeduplicatesPermissions(t *testing.T) {
	// Given
	givenUserID := uuid.New()
	givenRole1ID := uuid.New()
	givenRole2ID := uuid.New()

	givenRole1 := &domain.Role{
		ID:   givenRole1ID,
		Name: "editor",
	}

	givenRole2 := &domain.Role{
		ID:   givenRole2ID,
		Name: "viewer",
	}

	givenPermissionsRole1 := []*domain.Permission{
		{ID: uuid.New(), Code: "user:read", Description: "Read Users"},
		{ID: uuid.New(), Code: "user:write", Description: "Write Users"},
	}

	givenPermissionsRole2 := []*domain.Permission{
		{ID: uuid.New(), Code: "user:read", Description: "Read Users"},
		{ID: uuid.New(), Code: "post:read", Description: "Read Posts"},
	}

	mockRoleRepo := new(providers.MockRoleRepository)
	mockPermRepo := new(providers.MockPermissionRepository)
	mockCache := new(providers.MockPermissionCache)
	mockLogger := new(providers.MockLogger)

	resolveInheritanceUC := NewResolveInheritanceUseCase(mockRoleRepo, mockPermRepo, mockLogger)
	useCase := NewGetUserPermissionsUseCase(mockRoleRepo, resolveInheritanceUC, mockCache, mockLogger)

	mockCache.On("GetUserPermissions", mock.Anything, givenUserID.String()).Return(([]string)(nil), assert.AnError)
	mockRoleRepo.On("GetUserRoles", mock.Anything, givenUserID).Return([]*domain.Role{givenRole1, givenRole2}, nil)
	mockRoleRepo.On("GetByID", mock.Anything, givenRole1ID).Return(givenRole1, nil)
	mockRoleRepo.On("GetByID", mock.Anything, givenRole2ID).Return(givenRole2, nil)
	mockPermRepo.On("GetRolePermissions", mock.Anything, givenRole1ID).Return(givenPermissionsRole1, nil)
	mockPermRepo.On("GetRolePermissions", mock.Anything, givenRole2ID).Return(givenPermissionsRole2, nil)
	mockCache.On("SetUserPermissions", mock.Anything, givenUserID.String(), mock.AnythingOfType("[]string"), 5*time.Minute).Return(nil)
	mockLogger.On("Debug", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	permissions, err := useCase.Execute(context.Background(), givenUserID)

	// Then
	assert.NoError(t, err)
	assert.Equal(t, 3, len(permissions))
	assert.Contains(t, permissions, "user:read")
	assert.Contains(t, permissions, "user:write")
	assert.Contains(t, permissions, "post:read")
	mockRoleRepo.AssertExpectations(t)
	mockPermRepo.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestGetUserPermissionsUseCase_Execute_WhenResolveInheritanceFails_ReturnsError(t *testing.T) {
	// Given
	givenUserID := uuid.New()
	givenRoleID := uuid.New()

	givenRole := &domain.Role{
		ID:   givenRoleID,
		Name: "user",
	}

	mockRoleRepo := new(providers.MockRoleRepository)
	mockPermRepo := new(providers.MockPermissionRepository)
	mockCache := new(providers.MockPermissionCache)
	mockLogger := new(providers.MockLogger)

	resolveInheritanceUC := NewResolveInheritanceUseCase(mockRoleRepo, mockPermRepo, mockLogger)
	useCase := NewGetUserPermissionsUseCase(mockRoleRepo, resolveInheritanceUC, mockCache, mockLogger)

	mockCache.On("GetUserPermissions", mock.Anything, givenUserID.String()).Return(([]string)(nil), assert.AnError)
	mockRoleRepo.On("GetUserRoles", mock.Anything, givenUserID).Return([]*domain.Role{givenRole}, nil)
	mockRoleRepo.On("GetByID", mock.Anything, givenRoleID).Return((*domain.Role)(nil), assert.AnError)
	mockLogger.On("Debug", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	permissions, err := useCase.Execute(context.Background(), givenUserID)

	// Then
	assert.Error(t, err)
	assert.Nil(t, permissions)
	assert.Contains(t, err.Error(), "failed to resolve permissions")
	mockRoleRepo.AssertExpectations(t)
}

func TestGetUserPermissionsUseCase_Execute_WhenCacheSetFails_StillReturnsPermissions(t *testing.T) {
	// Given
	givenUserID := uuid.New()
	givenRoleID := uuid.New()

	givenRole := &domain.Role{
		ID:   givenRoleID,
		Name: "user",
	}

	givenPermissions := []*domain.Permission{
		{ID: uuid.New(), Code: "user:read", Description: "Read Users"},
	}

	mockRoleRepo := new(providers.MockRoleRepository)
	mockPermRepo := new(providers.MockPermissionRepository)
	mockCache := new(providers.MockPermissionCache)
	mockLogger := new(providers.MockLogger)

	resolveInheritanceUC := NewResolveInheritanceUseCase(mockRoleRepo, mockPermRepo, mockLogger)
	useCase := NewGetUserPermissionsUseCase(mockRoleRepo, resolveInheritanceUC, mockCache, mockLogger)

	mockCache.On("GetUserPermissions", mock.Anything, givenUserID.String()).Return(([]string)(nil), assert.AnError)
	mockRoleRepo.On("GetUserRoles", mock.Anything, givenUserID).Return([]*domain.Role{givenRole}, nil)
	mockRoleRepo.On("GetByID", mock.Anything, givenRoleID).Return(givenRole, nil)
	mockPermRepo.On("GetRolePermissions", mock.Anything, givenRoleID).Return(givenPermissions, nil)
	mockCache.On("SetUserPermissions", mock.Anything, givenUserID.String(), mock.AnythingOfType("[]string"), 5*time.Minute).Return(assert.AnError)
	mockLogger.On("Debug", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	permissions, err := useCase.Execute(context.Background(), givenUserID)

	// Then
	assert.NoError(t, err)
	assert.Equal(t, 1, len(permissions))
	assert.Contains(t, permissions, "user:read")
	mockCache.AssertExpectations(t)
}
