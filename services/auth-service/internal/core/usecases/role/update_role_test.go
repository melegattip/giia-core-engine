package role

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/giia/giia-core-engine/services/auth-service/internal/core/domain"
	"github.com/giia/giia-core-engine/services/auth-service/internal/core/providers"
)

func TestUpdateRoleUseCase_Execute_WithValidRequest_UpdatesRoleSuccessfully(t *testing.T) {
	// Given
	givenRoleID := uuid.New()
	givenRequest := &domain.UpdateRoleRequest{
		Name:        "updated-editor",
		Description: "Updated description",
	}

	givenRole := &domain.Role{
		ID:          givenRoleID,
		Name:        "editor",
		Description: "Old description",
		IsSystem:    false,
	}

	mockRoleRepo := new(providers.MockRoleRepository)
	mockPermRepo := new(providers.MockPermissionRepository)
	mockCache := new(providers.MockPermissionCache)
	mockLogger := new(providers.MockLogger)

	useCase := NewUpdateRoleUseCase(mockRoleRepo, mockPermRepo, mockCache, mockLogger)

	mockRoleRepo.On("GetByID", mock.Anything, givenRoleID).Return(givenRole, nil)
	mockRoleRepo.On("Update", mock.Anything, mock.MatchedBy(func(role *domain.Role) bool {
		return role.ID == givenRoleID &&
			role.Name == givenRequest.Name &&
			role.Description == givenRequest.Description
	})).Return(nil)
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	role, err := useCase.Execute(context.Background(), givenRoleID, givenRequest)

	// Then
	assert.NoError(t, err)
	assert.NotNil(t, role)
	assert.Equal(t, givenRequest.Name, role.Name)
	assert.Equal(t, givenRequest.Description, role.Description)
	mockRoleRepo.AssertExpectations(t)
}

func TestUpdateRoleUseCase_Execute_WithNilRoleID_ReturnsBadRequest(t *testing.T) {
	// Given
	givenRequest := &domain.UpdateRoleRequest{
		Name: "updated-editor",
	}

	mockRoleRepo := new(providers.MockRoleRepository)
	mockPermRepo := new(providers.MockPermissionRepository)
	mockCache := new(providers.MockPermissionCache)
	mockLogger := new(providers.MockLogger)

	useCase := NewUpdateRoleUseCase(mockRoleRepo, mockPermRepo, mockCache, mockLogger)

	// When
	role, err := useCase.Execute(context.Background(), uuid.Nil, givenRequest)

	// Then
	assert.Error(t, err)
	assert.Nil(t, role)
	assert.Contains(t, err.Error(), "role ID cannot be empty")
}

func TestUpdateRoleUseCase_Execute_WithNonExistentRole_ReturnsNotFound(t *testing.T) {
	// Given
	givenRoleID := uuid.New()
	givenRequest := &domain.UpdateRoleRequest{
		Name: "updated-editor",
	}

	mockRoleRepo := new(providers.MockRoleRepository)
	mockPermRepo := new(providers.MockPermissionRepository)
	mockCache := new(providers.MockPermissionCache)
	mockLogger := new(providers.MockLogger)

	useCase := NewUpdateRoleUseCase(mockRoleRepo, mockPermRepo, mockCache, mockLogger)

	mockRoleRepo.On("GetByID", mock.Anything, givenRoleID).Return((*domain.Role)(nil), assert.AnError)
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	role, err := useCase.Execute(context.Background(), givenRoleID, givenRequest)

	// Then
	assert.Error(t, err)
	assert.Nil(t, role)
	assert.Contains(t, err.Error(), "role not found")
	mockRoleRepo.AssertExpectations(t)
}

func TestUpdateRoleUseCase_Execute_WithSystemRole_ReturnsBadRequest(t *testing.T) {
	// Given
	givenRoleID := uuid.New()
	givenRequest := &domain.UpdateRoleRequest{
		Name: "updated-admin",
	}

	givenRole := &domain.Role{
		ID:       givenRoleID,
		Name:     "admin",
		IsSystem: true,
	}

	mockRoleRepo := new(providers.MockRoleRepository)
	mockPermRepo := new(providers.MockPermissionRepository)
	mockCache := new(providers.MockPermissionCache)
	mockLogger := new(providers.MockLogger)

	useCase := NewUpdateRoleUseCase(mockRoleRepo, mockPermRepo, mockCache, mockLogger)

	mockRoleRepo.On("GetByID", mock.Anything, givenRoleID).Return(givenRole, nil)

	// When
	role, err := useCase.Execute(context.Background(), givenRoleID, givenRequest)

	// Then
	assert.Error(t, err)
	assert.Nil(t, role)
	assert.Contains(t, err.Error(), "cannot update system roles")
	mockRoleRepo.AssertExpectations(t)
}

func TestUpdateRoleUseCase_Execute_WithInvalidParentRoleID_ReturnsBadRequest(t *testing.T) {
	// Given
	givenRoleID := uuid.New()
	givenInvalidParentID := "invalid-uuid"

	givenRequest := &domain.UpdateRoleRequest{
		Name:         "updated-editor",
		ParentRoleID: &givenInvalidParentID,
	}

	givenRole := &domain.Role{
		ID:       givenRoleID,
		Name:     "editor",
		IsSystem: false,
	}

	mockRoleRepo := new(providers.MockRoleRepository)
	mockPermRepo := new(providers.MockPermissionRepository)
	mockCache := new(providers.MockPermissionCache)
	mockLogger := new(providers.MockLogger)

	useCase := NewUpdateRoleUseCase(mockRoleRepo, mockPermRepo, mockCache, mockLogger)

	mockRoleRepo.On("GetByID", mock.Anything, givenRoleID).Return(givenRole, nil)

	// When
	role, err := useCase.Execute(context.Background(), givenRoleID, givenRequest)

	// Then
	assert.Error(t, err)
	assert.Nil(t, role)
	assert.Contains(t, err.Error(), "invalid parent role ID format")
	mockRoleRepo.AssertExpectations(t)
}

func TestUpdateRoleUseCase_Execute_WithNonExistentParentRole_ReturnsNotFound(t *testing.T) {
	// Given
	givenRoleID := uuid.New()
	givenParentRoleID := uuid.New()
	givenParentRoleIDString := givenParentRoleID.String()

	givenRequest := &domain.UpdateRoleRequest{
		Name:         "updated-editor",
		ParentRoleID: &givenParentRoleIDString,
	}

	givenRole := &domain.Role{
		ID:       givenRoleID,
		Name:     "editor",
		IsSystem: false,
	}

	mockRoleRepo := new(providers.MockRoleRepository)
	mockPermRepo := new(providers.MockPermissionRepository)
	mockCache := new(providers.MockPermissionCache)
	mockLogger := new(providers.MockLogger)

	useCase := NewUpdateRoleUseCase(mockRoleRepo, mockPermRepo, mockCache, mockLogger)

	mockRoleRepo.On("GetByID", mock.Anything, givenRoleID).Return(givenRole, nil)
	mockRoleRepo.On("GetByID", mock.Anything, givenParentRoleID).Return((*domain.Role)(nil), assert.AnError)

	// When
	role, err := useCase.Execute(context.Background(), givenRoleID, givenRequest)

	// Then
	assert.Error(t, err)
	assert.Nil(t, role)
	assert.Contains(t, err.Error(), "parent role not found")
	mockRoleRepo.AssertExpectations(t)
}

func TestUpdateRoleUseCase_Execute_WithSystemParentRoleInOrgRole_ReturnsBadRequest(t *testing.T) {
	// Given
	givenRoleID := uuid.New()
	givenOrgID := uuid.New()
	givenParentRoleID := uuid.New()
	givenParentRoleIDString := givenParentRoleID.String()

	givenRequest := &domain.UpdateRoleRequest{
		Name:         "updated-org-admin",
		ParentRoleID: &givenParentRoleIDString,
	}

	givenRole := &domain.Role{
		ID:             givenRoleID,
		Name:           "org-admin",
		OrganizationID: &givenOrgID,
		IsSystem:       false,
	}

	givenParentRole := &domain.Role{
		ID:       givenParentRoleID,
		Name:     "system-admin",
		IsSystem: true,
	}

	mockRoleRepo := new(providers.MockRoleRepository)
	mockPermRepo := new(providers.MockPermissionRepository)
	mockCache := new(providers.MockPermissionCache)
	mockLogger := new(providers.MockLogger)

	useCase := NewUpdateRoleUseCase(mockRoleRepo, mockPermRepo, mockCache, mockLogger)

	mockRoleRepo.On("GetByID", mock.Anything, givenRoleID).Return(givenRole, nil)
	mockRoleRepo.On("GetByID", mock.Anything, givenParentRoleID).Return(givenParentRole, nil)

	// When
	role, err := useCase.Execute(context.Background(), givenRoleID, givenRequest)

	// Then
	assert.Error(t, err)
	assert.Nil(t, role)
	assert.Contains(t, err.Error(), "cannot inherit from system role in organization-specific role")
	mockRoleRepo.AssertExpectations(t)
}

func TestUpdateRoleUseCase_Execute_WhenUpdateFails_ReturnsInternalServerError(t *testing.T) {
	// Given
	givenRoleID := uuid.New()
	givenRequest := &domain.UpdateRoleRequest{
		Name: "updated-editor",
	}

	givenRole := &domain.Role{
		ID:       givenRoleID,
		Name:     "editor",
		IsSystem: false,
	}

	mockRoleRepo := new(providers.MockRoleRepository)
	mockPermRepo := new(providers.MockPermissionRepository)
	mockCache := new(providers.MockPermissionCache)
	mockLogger := new(providers.MockLogger)

	useCase := NewUpdateRoleUseCase(mockRoleRepo, mockPermRepo, mockCache, mockLogger)

	mockRoleRepo.On("GetByID", mock.Anything, givenRoleID).Return(givenRole, nil)
	mockRoleRepo.On("Update", mock.Anything, mock.Anything).Return(assert.AnError)
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	role, err := useCase.Execute(context.Background(), givenRoleID, givenRequest)

	// Then
	assert.Error(t, err)
	assert.Nil(t, role)
	assert.Contains(t, err.Error(), "failed to update role")
	mockRoleRepo.AssertExpectations(t)
}

func TestUpdateRoleUseCase_Execute_WithPermissions_ReplacesPermissionsAndInvalidatesCache(t *testing.T) {
	// Given
	givenRoleID := uuid.New()
	givenPermID1 := uuid.New()
	givenPermID2 := uuid.New()
	givenUserID1 := uuid.New()
	givenUserID2 := uuid.New()

	givenRequest := &domain.UpdateRoleRequest{
		Name: "updated-editor",
		PermissionIDs: []string{
			givenPermID1.String(),
			givenPermID2.String(),
		},
	}

	givenRole := &domain.Role{
		ID:       givenRoleID,
		Name:     "editor",
		IsSystem: false,
	}

	mockRoleRepo := new(providers.MockRoleRepository)
	mockPermRepo := new(providers.MockPermissionRepository)
	mockCache := new(providers.MockPermissionCache)
	mockLogger := new(providers.MockLogger)

	useCase := NewUpdateRoleUseCase(mockRoleRepo, mockPermRepo, mockCache, mockLogger)

	mockRoleRepo.On("GetByID", mock.Anything, givenRoleID).Return(givenRole, nil)
	mockRoleRepo.On("Update", mock.Anything, mock.Anything).Return(nil)
	mockPermRepo.On("ReplaceRolePermissions", mock.Anything, givenRoleID, mock.MatchedBy(func(perms []uuid.UUID) bool {
		return len(perms) == 2 && perms[0] == givenPermID1 && perms[1] == givenPermID2
	})).Return(nil)
	mockRoleRepo.On("GetUsersWithRole", mock.Anything, givenRoleID).Return([]uuid.UUID{givenUserID1, givenUserID2}, nil)
	mockCache.On("InvalidateUsersWithRole", mock.Anything, mock.MatchedBy(func(userIDs []string) bool {
		return len(userIDs) == 2
	})).Return(nil)
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	role, err := useCase.Execute(context.Background(), givenRoleID, givenRequest)

	// Then
	assert.NoError(t, err)
	assert.NotNil(t, role)
	mockRoleRepo.AssertExpectations(t)
	mockPermRepo.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestUpdateRoleUseCase_Execute_WithInvalidPermissionID_ReturnsBadRequest(t *testing.T) {
	// Given
	givenRoleID := uuid.New()

	givenRequest := &domain.UpdateRoleRequest{
		Name: "updated-editor",
		PermissionIDs: []string{
			"invalid-uuid",
		},
	}

	givenRole := &domain.Role{
		ID:       givenRoleID,
		Name:     "editor",
		IsSystem: false,
	}

	mockRoleRepo := new(providers.MockRoleRepository)
	mockPermRepo := new(providers.MockPermissionRepository)
	mockCache := new(providers.MockPermissionCache)
	mockLogger := new(providers.MockLogger)

	useCase := NewUpdateRoleUseCase(mockRoleRepo, mockPermRepo, mockCache, mockLogger)

	mockRoleRepo.On("GetByID", mock.Anything, givenRoleID).Return(givenRole, nil)
	mockRoleRepo.On("Update", mock.Anything, mock.Anything).Return(nil)
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	role, err := useCase.Execute(context.Background(), givenRoleID, givenRequest)

	// Then
	assert.Error(t, err)
	assert.Nil(t, role)
	assert.Contains(t, err.Error(), "invalid permission ID format")
}

func TestUpdateRoleUseCase_Execute_WhenReplacePermissionsFails_ReturnsInternalServerError(t *testing.T) {
	// Given
	givenRoleID := uuid.New()
	givenPermID := uuid.New()

	givenRequest := &domain.UpdateRoleRequest{
		Name: "updated-editor",
		PermissionIDs: []string{
			givenPermID.String(),
		},
	}

	givenRole := &domain.Role{
		ID:       givenRoleID,
		Name:     "editor",
		IsSystem: false,
	}

	mockRoleRepo := new(providers.MockRoleRepository)
	mockPermRepo := new(providers.MockPermissionRepository)
	mockCache := new(providers.MockPermissionCache)
	mockLogger := new(providers.MockLogger)

	useCase := NewUpdateRoleUseCase(mockRoleRepo, mockPermRepo, mockCache, mockLogger)

	mockRoleRepo.On("GetByID", mock.Anything, givenRoleID).Return(givenRole, nil)
	mockRoleRepo.On("Update", mock.Anything, mock.Anything).Return(nil)
	mockPermRepo.On("ReplaceRolePermissions", mock.Anything, givenRoleID, mock.Anything).Return(assert.AnError)
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	role, err := useCase.Execute(context.Background(), givenRoleID, givenRequest)

	// Then
	assert.Error(t, err)
	assert.Nil(t, role)
	assert.Contains(t, err.Error(), "failed to update role permissions")
	mockRoleRepo.AssertExpectations(t)
	mockPermRepo.AssertExpectations(t)
}

func TestUpdateRoleUseCase_Execute_WhenCacheInvalidationFails_StillReturnsSuccess(t *testing.T) {
	// Given
	givenRoleID := uuid.New()
	givenPermID := uuid.New()
	givenUserID := uuid.New()

	givenRequest := &domain.UpdateRoleRequest{
		Name: "updated-editor",
		PermissionIDs: []string{
			givenPermID.String(),
		},
	}

	givenRole := &domain.Role{
		ID:       givenRoleID,
		Name:     "editor",
		IsSystem: false,
	}

	mockRoleRepo := new(providers.MockRoleRepository)
	mockPermRepo := new(providers.MockPermissionRepository)
	mockCache := new(providers.MockPermissionCache)
	mockLogger := new(providers.MockLogger)

	useCase := NewUpdateRoleUseCase(mockRoleRepo, mockPermRepo, mockCache, mockLogger)

	mockRoleRepo.On("GetByID", mock.Anything, givenRoleID).Return(givenRole, nil)
	mockRoleRepo.On("Update", mock.Anything, mock.Anything).Return(nil)
	mockPermRepo.On("ReplaceRolePermissions", mock.Anything, givenRoleID, mock.Anything).Return(nil)
	mockRoleRepo.On("GetUsersWithRole", mock.Anything, givenRoleID).Return([]uuid.UUID{givenUserID}, nil)
	mockCache.On("InvalidateUsersWithRole", mock.Anything, mock.Anything).Return(assert.AnError)
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	role, err := useCase.Execute(context.Background(), givenRoleID, givenRequest)

	// Then
	assert.NoError(t, err)
	assert.NotNil(t, role)
	mockRoleRepo.AssertExpectations(t)
	mockPermRepo.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}
