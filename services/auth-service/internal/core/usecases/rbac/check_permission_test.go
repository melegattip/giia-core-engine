package rbac

import (
	"context"
	"testing"

	"github.com/giia/giia-core-engine/services/auth-service/internal/core/providers"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCheckPermissionUseCase_Execute_WithExactMatchPermission_ReturnsTrue(t *testing.T) {
	// Given
	givenUserID := uuid.New()
	givenRequiredPermission := "catalog:products:read"
	givenUserPermissions := []string{"catalog:products:read", "catalog:products:write"}

	mockRoleRepo := new(providers.MockRoleRepository)
	mockCache := new(providers.MockPermissionCache)
	mockLogger := new(providers.MockLogger)

	mockResolveInheritance := new(MockResolveInheritanceUseCase)

	getUserPermsUC := NewGetUserPermissionsUseCase(mockRoleRepo, mockResolveInheritance, mockCache, mockLogger)
	useCase := NewCheckPermissionUseCase(getUserPermsUC, mockLogger)

	// Mock GetUserPermissions behavior
	mockRoleRepo.On("GetUserRoles", mock.Anything, givenUserID).Return([]*domain.Role{}, nil)
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

	mockGetUserPerms := new(MockGetUserPermissionsUseCase)
	mockLogger := new(providers.MockLogger)

	useCase := NewCheckPermissionUseCase(mockGetUserPerms, mockLogger)

	mockGetUserPerms.On("Execute", mock.Anything, givenUserID).Return(givenUserPermissions, nil)

	// When
	allowed, err := useCase.Execute(context.Background(), givenUserID, givenRequiredPermission)

	// Then
	assert.NoError(t, err)
	assert.True(t, allowed)

	mockGetUserPerms.AssertExpectations(t)
}

func TestCheckPermissionUseCase_Execute_WithServiceWildcardPermission_ReturnsTrue(t *testing.T) {
	// Given
	givenUserID := uuid.New()
	givenRequiredPermission := "catalog:products:read"
	givenUserPermissions := []string{"catalog:*:*"} // All catalog permissions

	mockGetUserPerms := new(MockGetUserPermissionsUseCase)
	mockLogger := new(providers.MockLogger)

	useCase := NewCheckPermissionUseCase(mockGetUserPerms, mockLogger)

	mockGetUserPerms.On("Execute", mock.Anything, givenUserID).Return(givenUserPermissions, nil)

	// When
	allowed, err := useCase.Execute(context.Background(), givenUserID, givenRequiredPermission)

	// Then
	assert.NoError(t, err)
	assert.True(t, allowed)

	mockGetUserPerms.AssertExpectations(t)
}

func TestCheckPermissionUseCase_Execute_WithResourceWildcardPermission_ReturnsTrue(t *testing.T) {
	// Given
	givenUserID := uuid.New()
	givenRequiredPermission := "catalog:products:read"
	givenUserPermissions := []string{"catalog:products:*"} // All actions on products

	mockGetUserPerms := new(MockGetUserPermissionsUseCase)
	mockLogger := new(providers.MockLogger)

	useCase := NewCheckPermissionUseCase(mockGetUserPerms, mockLogger)

	mockGetUserPerms.On("Execute", mock.Anything, givenUserID).Return(givenUserPermissions, nil)

	// When
	allowed, err := useCase.Execute(context.Background(), givenUserID, givenRequiredPermission)

	// Then
	assert.NoError(t, err)
	assert.True(t, allowed)

	mockGetUserPerms.AssertExpectations(t)
}

func TestCheckPermissionUseCase_Execute_WithoutPermission_ReturnsFalse(t *testing.T) {
	// Given
	givenUserID := uuid.New()
	givenRequiredPermission := "catalog:products:write"
	givenUserPermissions := []string{"catalog:products:read"} // Only read, not write

	mockGetUserPerms := new(MockGetUserPermissionsUseCase)
	mockLogger := new(providers.MockLogger)

	useCase := NewCheckPermissionUseCase(mockGetUserPerms, mockLogger)

	mockGetUserPerms.On("Execute", mock.Anything, givenUserID).Return(givenUserPermissions, nil)
	mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	allowed, err := useCase.Execute(context.Background(), givenUserID, givenRequiredPermission)

	// Then
	assert.NoError(t, err)
	assert.False(t, allowed)

	mockGetUserPerms.AssertExpectations(t)
}

func TestCheckPermissionUseCase_Execute_WithNoPermissions_ReturnsFalse(t *testing.T) {
	// Given
	givenUserID := uuid.New()
	givenRequiredPermission := "catalog:products:read"
	givenUserPermissions := []string{} // No permissions

	mockGetUserPerms := new(MockGetUserPermissionsUseCase)
	mockLogger := new(providers.MockLogger)

	useCase := NewCheckPermissionUseCase(mockGetUserPerms, mockLogger)

	mockGetUserPerms.On("Execute", mock.Anything, givenUserID).Return(givenUserPermissions, nil)
	mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	allowed, err := useCase.Execute(context.Background(), givenUserID, givenRequiredPermission)

	// Then
	assert.NoError(t, err)
	assert.False(t, allowed)

	mockGetUserPerms.AssertExpectations(t)
}

func TestCheckPermissionUseCase_Execute_WhenGetPermissionsFails_ReturnsError(t *testing.T) {
	// Given
	givenUserID := uuid.New()
	givenRequiredPermission := "catalog:products:read"
	givenError := assert.AnError

	mockGetUserPerms := new(MockGetUserPermissionsUseCase)
	mockLogger := new(providers.MockLogger)

	useCase := NewCheckPermissionUseCase(mockGetUserPerms, mockLogger)

	mockGetUserPerms.On("Execute", mock.Anything, givenUserID).Return([]string(nil), givenError)
	mockLogger.On("Error", mock.Anything, givenError, mock.Anything, mock.Anything).Return()

	// When
	allowed, err := useCase.Execute(context.Background(), givenUserID, givenRequiredPermission)

	// Then
	assert.Error(t, err)
	assert.False(t, allowed)
	assert.Equal(t, givenError, err)

	mockGetUserPerms.AssertExpectations(t)
}

func TestCheckPermissionUseCase_Execute_WithMultipleWildcards_ChoosesMostSpecific(t *testing.T) {
	// Given
	givenUserID := uuid.New()
	givenRequiredPermission := "catalog:products:read"
	givenUserPermissions := []string{
		"*:*:*",               // Admin
		"catalog:*:*",         // Catalog admin
		"catalog:products:*",  // Product admin
		"catalog:products:read", // Specific read
	}

	mockGetUserPerms := new(MockGetUserPermissionsUseCase)
	mockLogger := new(providers.MockLogger)

	useCase := NewCheckPermissionUseCase(mockGetUserPerms, mockLogger)

	mockGetUserPerms.On("Execute", mock.Anything, givenUserID).Return(givenUserPermissions, nil)

	// When
	allowed, err := useCase.Execute(context.Background(), givenUserID, givenRequiredPermission)

	// Then
	assert.NoError(t, err)
	assert.True(t, allowed)

	mockGetUserPerms.AssertExpectations(t)
}

// MockGetUserPermissionsUseCase is a mock for testing
type MockGetUserPermissionsUseCase struct {
	mock.Mock
}

func (m *MockGetUserPermissionsUseCase) Execute(ctx context.Context, userID uuid.UUID) ([]string, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}
