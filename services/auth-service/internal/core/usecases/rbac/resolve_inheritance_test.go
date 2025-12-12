package rbac

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/giia/giia-core-engine/services/auth-service/internal/core/domain"
	"github.com/giia/giia-core-engine/services/auth-service/internal/core/providers"
)

func TestResolveInheritanceUseCase_Execute_WithSingleRole_ReturnsRolePermissions(t *testing.T) {
	// Given
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
	mockLogger := new(providers.MockLogger)

	useCase := NewResolveInheritanceUseCase(mockRoleRepo, mockPermRepo, mockLogger)

	mockRoleRepo.On("GetByID", mock.Anything, givenRoleID).Return(givenRole, nil)
	mockPermRepo.On("GetRolePermissions", mock.Anything, givenRoleID).Return(givenPermissions, nil)
	mockLogger.On("Debug", mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	permissions, err := useCase.Execute(context.Background(), givenRoleID)

	// Then
	assert.NoError(t, err)
	assert.Equal(t, 2, len(permissions))
	assert.Contains(t, getPermissionCodes(permissions), "user:read")
	assert.Contains(t, getPermissionCodes(permissions), "user:write")
	mockRoleRepo.AssertExpectations(t)
	mockPermRepo.AssertExpectations(t)
}

func TestResolveInheritanceUseCase_Execute_WithNilRoleID_ReturnsBadRequest(t *testing.T) {
	// Given
	mockRoleRepo := new(providers.MockRoleRepository)
	mockPermRepo := new(providers.MockPermissionRepository)
	mockLogger := new(providers.MockLogger)

	useCase := NewResolveInheritanceUseCase(mockRoleRepo, mockPermRepo, mockLogger)

	// When
	permissions, err := useCase.Execute(context.Background(), uuid.Nil)

	// Then
	assert.Error(t, err)
	assert.Nil(t, permissions)
	assert.Contains(t, err.Error(), "role ID cannot be empty")
}

func TestResolveInheritanceUseCase_Execute_WithParentRole_ReturnsInheritedPermissions(t *testing.T) {
	// Given
	givenChildRoleID := uuid.New()
	givenParentRoleID := uuid.New()

	givenChildRole := &domain.Role{
		ID:           givenChildRoleID,
		Name:         "editor",
		ParentRoleID: &givenParentRoleID,
	}

	givenParentRole := &domain.Role{
		ID:   givenParentRoleID,
		Name: "viewer",
	}

	givenChildPermissions := []*domain.Permission{
		{ID: uuid.New(), Code: "user:write", Description: "Write Users"},
	}

	givenParentPermissions := []*domain.Permission{
		{ID: uuid.New(), Code: "user:read", Description: "Read Users"},
	}

	mockRoleRepo := new(providers.MockRoleRepository)
	mockPermRepo := new(providers.MockPermissionRepository)
	mockLogger := new(providers.MockLogger)

	useCase := NewResolveInheritanceUseCase(mockRoleRepo, mockPermRepo, mockLogger)

	mockRoleRepo.On("GetByID", mock.Anything, givenChildRoleID).Return(givenChildRole, nil)
	mockRoleRepo.On("GetByID", mock.Anything, givenParentRoleID).Return(givenParentRole, nil)
	mockPermRepo.On("GetRolePermissions", mock.Anything, givenChildRoleID).Return(givenChildPermissions, nil)
	mockPermRepo.On("GetRolePermissions", mock.Anything, givenParentRoleID).Return(givenParentPermissions, nil)
	mockLogger.On("Debug", mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	permissions, err := useCase.Execute(context.Background(), givenChildRoleID)

	// Then
	assert.NoError(t, err)
	assert.Equal(t, 2, len(permissions))
	assert.Contains(t, getPermissionCodes(permissions), "user:read")
	assert.Contains(t, getPermissionCodes(permissions), "user:write")
	mockRoleRepo.AssertExpectations(t)
	mockPermRepo.AssertExpectations(t)
}

func TestResolveInheritanceUseCase_Execute_WithMultiLevelHierarchy_ReturnsAllPermissions(t *testing.T) {
	// Given
	givenGrandchildRoleID := uuid.New()
	givenChildRoleID := uuid.New()
	givenParentRoleID := uuid.New()

	givenGrandchildRole := &domain.Role{
		ID:           givenGrandchildRoleID,
		Name:         "admin",
		ParentRoleID: &givenChildRoleID,
	}

	givenChildRole := &domain.Role{
		ID:           givenChildRoleID,
		Name:         "editor",
		ParentRoleID: &givenParentRoleID,
	}

	givenParentRole := &domain.Role{
		ID:   givenParentRoleID,
		Name: "viewer",
	}

	givenGrandchildPermissions := []*domain.Permission{
		{ID: uuid.New(), Code: "user:delete", Description: "Delete Users"},
	}

	givenChildPermissions := []*domain.Permission{
		{ID: uuid.New(), Code: "user:write", Description: "Write Users"},
	}

	givenParentPermissions := []*domain.Permission{
		{ID: uuid.New(), Code: "user:read", Description: "Read Users"},
	}

	mockRoleRepo := new(providers.MockRoleRepository)
	mockPermRepo := new(providers.MockPermissionRepository)
	mockLogger := new(providers.MockLogger)

	useCase := NewResolveInheritanceUseCase(mockRoleRepo, mockPermRepo, mockLogger)

	mockRoleRepo.On("GetByID", mock.Anything, givenGrandchildRoleID).Return(givenGrandchildRole, nil)
	mockRoleRepo.On("GetByID", mock.Anything, givenChildRoleID).Return(givenChildRole, nil)
	mockRoleRepo.On("GetByID", mock.Anything, givenParentRoleID).Return(givenParentRole, nil)
	mockPermRepo.On("GetRolePermissions", mock.Anything, givenGrandchildRoleID).Return(givenGrandchildPermissions, nil)
	mockPermRepo.On("GetRolePermissions", mock.Anything, givenChildRoleID).Return(givenChildPermissions, nil)
	mockPermRepo.On("GetRolePermissions", mock.Anything, givenParentRoleID).Return(givenParentPermissions, nil)
	mockLogger.On("Debug", mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	permissions, err := useCase.Execute(context.Background(), givenGrandchildRoleID)

	// Then
	assert.NoError(t, err)
	assert.Equal(t, 3, len(permissions))
	assert.Contains(t, getPermissionCodes(permissions), "user:read")
	assert.Contains(t, getPermissionCodes(permissions), "user:write")
	assert.Contains(t, getPermissionCodes(permissions), "user:delete")
	mockRoleRepo.AssertExpectations(t)
	mockPermRepo.AssertExpectations(t)
}

func TestResolveInheritanceUseCase_Execute_WithCircularDependency_ReturnsError(t *testing.T) {
	// Given
	givenRole1ID := uuid.New()
	givenRole2ID := uuid.New()

	givenRole1 := &domain.Role{
		ID:           givenRole1ID,
		Name:         "role1",
		ParentRoleID: &givenRole2ID,
	}

	givenRole2 := &domain.Role{
		ID:           givenRole2ID,
		Name:         "role2",
		ParentRoleID: &givenRole1ID,
	}

	givenPermissions := []*domain.Permission{
		{ID: uuid.New(), Code: "user:read", Description: "Read Users"},
	}

	mockRoleRepo := new(providers.MockRoleRepository)
	mockPermRepo := new(providers.MockPermissionRepository)
	mockLogger := new(providers.MockLogger)

	useCase := NewResolveInheritanceUseCase(mockRoleRepo, mockPermRepo, mockLogger)

	mockRoleRepo.On("GetByID", mock.Anything, givenRole1ID).Return(givenRole1, nil)
	mockRoleRepo.On("GetByID", mock.Anything, givenRole2ID).Return(givenRole2, nil)
	mockPermRepo.On("GetRolePermissions", mock.Anything, givenRole1ID).Return(givenPermissions, nil)
	mockPermRepo.On("GetRolePermissions", mock.Anything, givenRole2ID).Return(givenPermissions, nil)

	// When
	permissions, err := useCase.Execute(context.Background(), givenRole1ID)

	// Then
	assert.Error(t, err)
	assert.Nil(t, permissions)
	assert.Contains(t, err.Error(), "circular role hierarchy detected")
	mockRoleRepo.AssertExpectations(t)
	mockPermRepo.AssertExpectations(t)
}

func TestResolveInheritanceUseCase_Execute_WithDuplicatePermissions_ReturnsDeduplicatedList(t *testing.T) {
	// Given
	givenChildRoleID := uuid.New()
	givenParentRoleID := uuid.New()

	givenChildRole := &domain.Role{
		ID:           givenChildRoleID,
		Name:         "editor",
		ParentRoleID: &givenParentRoleID,
	}

	givenParentRole := &domain.Role{
		ID:   givenParentRoleID,
		Name: "viewer",
	}

	givenSharedPermission := &domain.Permission{
		ID:          uuid.New(),
		Code:        "user:read",
		Description: "Read Users",
	}

	givenChildPermissions := []*domain.Permission{
		givenSharedPermission,
		{ID: uuid.New(), Code: "user:write", Description: "Write Users"},
	}

	givenParentPermissions := []*domain.Permission{
		givenSharedPermission,
	}

	mockRoleRepo := new(providers.MockRoleRepository)
	mockPermRepo := new(providers.MockPermissionRepository)
	mockLogger := new(providers.MockLogger)

	useCase := NewResolveInheritanceUseCase(mockRoleRepo, mockPermRepo, mockLogger)

	mockRoleRepo.On("GetByID", mock.Anything, givenChildRoleID).Return(givenChildRole, nil)
	mockRoleRepo.On("GetByID", mock.Anything, givenParentRoleID).Return(givenParentRole, nil)
	mockPermRepo.On("GetRolePermissions", mock.Anything, givenChildRoleID).Return(givenChildPermissions, nil)
	mockPermRepo.On("GetRolePermissions", mock.Anything, givenParentRoleID).Return(givenParentPermissions, nil)
	mockLogger.On("Debug", mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	permissions, err := useCase.Execute(context.Background(), givenChildRoleID)

	// Then
	assert.NoError(t, err)
	assert.Equal(t, 2, len(permissions))
	codes := getPermissionCodes(permissions)
	assert.Contains(t, codes, "user:read")
	assert.Contains(t, codes, "user:write")
	mockRoleRepo.AssertExpectations(t)
	mockPermRepo.AssertExpectations(t)
}

func TestResolveInheritanceUseCase_Execute_WhenGetRoleFails_ReturnsError(t *testing.T) {
	// Given
	givenRoleID := uuid.New()

	mockRoleRepo := new(providers.MockRoleRepository)
	mockPermRepo := new(providers.MockPermissionRepository)
	mockLogger := new(providers.MockLogger)

	useCase := NewResolveInheritanceUseCase(mockRoleRepo, mockPermRepo, mockLogger)

	mockRoleRepo.On("GetByID", mock.Anything, givenRoleID).Return((*domain.Role)(nil), assert.AnError)
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	permissions, err := useCase.Execute(context.Background(), givenRoleID)

	// Then
	assert.Error(t, err)
	assert.Nil(t, permissions)
	assert.Contains(t, err.Error(), "failed to resolve role hierarchy")
	mockRoleRepo.AssertExpectations(t)
}

func TestResolveInheritanceUseCase_Execute_WhenGetPermissionsFails_ReturnsError(t *testing.T) {
	// Given
	givenRoleID := uuid.New()

	givenRole := &domain.Role{
		ID:   givenRoleID,
		Name: "user",
	}

	mockRoleRepo := new(providers.MockRoleRepository)
	mockPermRepo := new(providers.MockPermissionRepository)
	mockLogger := new(providers.MockLogger)

	useCase := NewResolveInheritanceUseCase(mockRoleRepo, mockPermRepo, mockLogger)

	mockRoleRepo.On("GetByID", mock.Anything, givenRoleID).Return(givenRole, nil)
	mockPermRepo.On("GetRolePermissions", mock.Anything, givenRoleID).Return(([]*domain.Permission)(nil), assert.AnError)
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	permissions, err := useCase.Execute(context.Background(), givenRoleID)

	// Then
	assert.Error(t, err)
	assert.Nil(t, permissions)
	assert.Contains(t, err.Error(), "failed to get role permissions")
	mockRoleRepo.AssertExpectations(t)
	mockPermRepo.AssertExpectations(t)
}

func TestResolveInheritanceUseCase_Execute_WithNoPermissions_ReturnsEmptyList(t *testing.T) {
	// Given
	givenRoleID := uuid.New()

	givenRole := &domain.Role{
		ID:   givenRoleID,
		Name: "empty",
	}

	mockRoleRepo := new(providers.MockRoleRepository)
	mockPermRepo := new(providers.MockPermissionRepository)
	mockLogger := new(providers.MockLogger)

	useCase := NewResolveInheritanceUseCase(mockRoleRepo, mockPermRepo, mockLogger)

	mockRoleRepo.On("GetByID", mock.Anything, givenRoleID).Return(givenRole, nil)
	mockPermRepo.On("GetRolePermissions", mock.Anything, givenRoleID).Return([]*domain.Permission{}, nil)
	mockLogger.On("Debug", mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	permissions, err := useCase.Execute(context.Background(), givenRoleID)

	// Then
	assert.NoError(t, err)
	assert.NotNil(t, permissions)
	assert.Equal(t, 0, len(permissions))
	mockRoleRepo.AssertExpectations(t)
	mockPermRepo.AssertExpectations(t)
}

func getPermissionCodes(permissions []*domain.Permission) []string {
	codes := make([]string, len(permissions))
	for i, perm := range permissions {
		codes[i] = perm.Code
	}
	return codes
}
