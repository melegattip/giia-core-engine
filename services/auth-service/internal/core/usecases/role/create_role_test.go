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

func TestCreateRoleUseCase_Execute_WithValidRequest_ReturnsCreatedRole(t *testing.T) {
	// Given
	givenRequest := &domain.CreateRoleRequest{
		Name:        "editor",
		Description: "Editor role",
	}

	mockRoleRepo := new(providers.MockRoleRepository)
	mockPermRepo := new(providers.MockPermissionRepository)
	mockLogger := new(providers.MockLogger)

	useCase := NewCreateRoleUseCase(mockRoleRepo, mockPermRepo, mockLogger)

	mockRoleRepo.On("GetByName", mock.Anything, givenRequest.Name, (*uuid.UUID)(nil)).Return((*domain.Role)(nil), assert.AnError)
	mockRoleRepo.On("Create", mock.Anything, mock.MatchedBy(func(role *domain.Role) bool {
		return role.Name == givenRequest.Name &&
			role.Description == givenRequest.Description &&
			role.IsSystem == false
	})).Return(nil)
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	role, err := useCase.Execute(context.Background(), givenRequest)

	// Then
	assert.NoError(t, err)
	assert.NotNil(t, role)
	assert.Equal(t, givenRequest.Name, role.Name)
	assert.Equal(t, givenRequest.Description, role.Description)
	assert.False(t, role.IsSystem)
	mockRoleRepo.AssertExpectations(t)
}

func TestCreateRoleUseCase_Execute_WithEmptyName_ReturnsBadRequest(t *testing.T) {
	// Given
	givenRequest := &domain.CreateRoleRequest{
		Name:        "",
		Description: "Test role",
	}

	mockRoleRepo := new(providers.MockRoleRepository)
	mockPermRepo := new(providers.MockPermissionRepository)
	mockLogger := new(providers.MockLogger)

	useCase := NewCreateRoleUseCase(mockRoleRepo, mockPermRepo, mockLogger)

	// When
	role, err := useCase.Execute(context.Background(), givenRequest)

	// Then
	assert.Error(t, err)
	assert.Nil(t, role)
	assert.Contains(t, err.Error(), "role name is required")
}

func TestCreateRoleUseCase_Execute_WithInvalidOrganizationID_ReturnsBadRequest(t *testing.T) {
	// Given
	givenInvalidOrgID := "invalid-uuid"
	givenRequest := &domain.CreateRoleRequest{
		Name:           "editor",
		Description:    "Editor role",
		OrganizationID: &givenInvalidOrgID,
	}

	mockRoleRepo := new(providers.MockRoleRepository)
	mockPermRepo := new(providers.MockPermissionRepository)
	mockLogger := new(providers.MockLogger)

	useCase := NewCreateRoleUseCase(mockRoleRepo, mockPermRepo, mockLogger)

	// When
	role, err := useCase.Execute(context.Background(), givenRequest)

	// Then
	assert.Error(t, err)
	assert.Nil(t, role)
	assert.Contains(t, err.Error(), "invalid organization ID format")
}

func TestCreateRoleUseCase_Execute_WithDuplicateName_ReturnsBadRequest(t *testing.T) {
	// Given
	givenOrgID := uuid.New()
	givenOrgIDString := givenOrgID.String()
	givenRequest := &domain.CreateRoleRequest{
		Name:           "editor",
		Description:    "Editor role",
		OrganizationID: &givenOrgIDString,
	}

	givenExistingRole := &domain.Role{
		ID:   uuid.New(),
		Name: givenRequest.Name,
	}

	mockRoleRepo := new(providers.MockRoleRepository)
	mockPermRepo := new(providers.MockPermissionRepository)
	mockLogger := new(providers.MockLogger)

	useCase := NewCreateRoleUseCase(mockRoleRepo, mockPermRepo, mockLogger)

	mockRoleRepo.On("GetByName", mock.Anything, givenRequest.Name, &givenOrgID).Return(givenExistingRole, nil)

	// When
	role, err := useCase.Execute(context.Background(), givenRequest)

	// Then
	assert.Error(t, err)
	assert.Nil(t, role)
	assert.Contains(t, err.Error(), "role with this name already exists")
	mockRoleRepo.AssertExpectations(t)
}

func TestCreateRoleUseCase_Execute_WithInvalidParentRoleID_ReturnsBadRequest(t *testing.T) {
	// Given
	givenInvalidParentID := "invalid-uuid"
	givenRequest := &domain.CreateRoleRequest{
		Name:         "editor",
		Description:  "Editor role",
		ParentRoleID: &givenInvalidParentID,
	}

	mockRoleRepo := new(providers.MockRoleRepository)
	mockPermRepo := new(providers.MockPermissionRepository)
	mockLogger := new(providers.MockLogger)

	useCase := NewCreateRoleUseCase(mockRoleRepo, mockPermRepo, mockLogger)

	mockRoleRepo.On("GetByName", mock.Anything, givenRequest.Name, (*uuid.UUID)(nil)).Return((*domain.Role)(nil), assert.AnError)

	// When
	role, err := useCase.Execute(context.Background(), givenRequest)

	// Then
	assert.Error(t, err)
	assert.Nil(t, role)
	assert.Contains(t, err.Error(), "invalid parent role ID format")
	mockRoleRepo.AssertExpectations(t)
}

func TestCreateRoleUseCase_Execute_WithNonExistentParentRole_ReturnsNotFound(t *testing.T) {
	// Given
	givenParentRoleID := uuid.New()
	givenParentRoleIDString := givenParentRoleID.String()
	givenRequest := &domain.CreateRoleRequest{
		Name:         "editor",
		Description:  "Editor role",
		ParentRoleID: &givenParentRoleIDString,
	}

	mockRoleRepo := new(providers.MockRoleRepository)
	mockPermRepo := new(providers.MockPermissionRepository)
	mockLogger := new(providers.MockLogger)

	useCase := NewCreateRoleUseCase(mockRoleRepo, mockPermRepo, mockLogger)

	mockRoleRepo.On("GetByName", mock.Anything, givenRequest.Name, (*uuid.UUID)(nil)).Return((*domain.Role)(nil), assert.AnError)
	mockRoleRepo.On("GetByID", mock.Anything, givenParentRoleID).Return((*domain.Role)(nil), assert.AnError)

	// When
	role, err := useCase.Execute(context.Background(), givenRequest)

	// Then
	assert.Error(t, err)
	assert.Nil(t, role)
	assert.Contains(t, err.Error(), "parent role not found")
	mockRoleRepo.AssertExpectations(t)
}

func TestCreateRoleUseCase_Execute_WithSystemParentRoleInOrgRole_ReturnsBadRequest(t *testing.T) {
	// Given
	givenOrgID := uuid.New()
	givenOrgIDString := givenOrgID.String()
	givenParentRoleID := uuid.New()
	givenParentRoleIDString := givenParentRoleID.String()

	givenRequest := &domain.CreateRoleRequest{
		Name:           "org-admin",
		Description:    "Organization admin",
		OrganizationID: &givenOrgIDString,
		ParentRoleID:   &givenParentRoleIDString,
	}

	givenParentRole := &domain.Role{
		ID:       givenParentRoleID,
		Name:     "system-admin",
		IsSystem: true,
	}

	mockRoleRepo := new(providers.MockRoleRepository)
	mockPermRepo := new(providers.MockPermissionRepository)
	mockLogger := new(providers.MockLogger)

	useCase := NewCreateRoleUseCase(mockRoleRepo, mockPermRepo, mockLogger)

	mockRoleRepo.On("GetByName", mock.Anything, givenRequest.Name, &givenOrgID).Return((*domain.Role)(nil), assert.AnError)
	mockRoleRepo.On("GetByID", mock.Anything, givenParentRoleID).Return(givenParentRole, nil)

	// When
	role, err := useCase.Execute(context.Background(), givenRequest)

	// Then
	assert.Error(t, err)
	assert.Nil(t, role)
	assert.Contains(t, err.Error(), "cannot inherit from system role in organization-specific role")
	mockRoleRepo.AssertExpectations(t)
}

func TestCreateRoleUseCase_Execute_WhenCreateFails_ReturnsInternalServerError(t *testing.T) {
	// Given
	givenRequest := &domain.CreateRoleRequest{
		Name:        "editor",
		Description: "Editor role",
	}

	mockRoleRepo := new(providers.MockRoleRepository)
	mockPermRepo := new(providers.MockPermissionRepository)
	mockLogger := new(providers.MockLogger)

	useCase := NewCreateRoleUseCase(mockRoleRepo, mockPermRepo, mockLogger)

	mockRoleRepo.On("GetByName", mock.Anything, givenRequest.Name, (*uuid.UUID)(nil)).Return((*domain.Role)(nil), assert.AnError)
	mockRoleRepo.On("Create", mock.Anything, mock.Anything).Return(assert.AnError)
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	role, err := useCase.Execute(context.Background(), givenRequest)

	// Then
	assert.Error(t, err)
	assert.Nil(t, role)
	assert.Contains(t, err.Error(), "failed to create role")
	mockRoleRepo.AssertExpectations(t)
}

func TestCreateRoleUseCase_Execute_WithPermissions_AssignsPermissionsSuccessfully(t *testing.T) {
	// Given
	givenPermID1 := uuid.New()
	givenPermID2 := uuid.New()

	givenRequest := &domain.CreateRoleRequest{
		Name:        "editor",
		Description: "Editor role",
		PermissionIDs: []string{
			givenPermID1.String(),
			givenPermID2.String(),
		},
	}

	mockRoleRepo := new(providers.MockRoleRepository)
	mockPermRepo := new(providers.MockPermissionRepository)
	mockLogger := new(providers.MockLogger)

	useCase := NewCreateRoleUseCase(mockRoleRepo, mockPermRepo, mockLogger)

	mockRoleRepo.On("GetByName", mock.Anything, givenRequest.Name, (*uuid.UUID)(nil)).Return((*domain.Role)(nil), assert.AnError)
	mockRoleRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
	mockPermRepo.On("AssignPermissionsToRole", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.MatchedBy(func(perms []uuid.UUID) bool {
		return len(perms) == 2 && perms[0] == givenPermID1 && perms[1] == givenPermID2
	})).Return(nil)
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	role, err := useCase.Execute(context.Background(), givenRequest)

	// Then
	assert.NoError(t, err)
	assert.NotNil(t, role)
	mockRoleRepo.AssertExpectations(t)
	mockPermRepo.AssertExpectations(t)
}

func TestCreateRoleUseCase_Execute_WithInvalidPermissionID_ReturnsBadRequest(t *testing.T) {
	// Given
	givenRequest := &domain.CreateRoleRequest{
		Name:        "editor",
		Description: "Editor role",
		PermissionIDs: []string{
			"invalid-uuid",
		},
	}

	mockRoleRepo := new(providers.MockRoleRepository)
	mockPermRepo := new(providers.MockPermissionRepository)
	mockLogger := new(providers.MockLogger)

	useCase := NewCreateRoleUseCase(mockRoleRepo, mockPermRepo, mockLogger)

	mockRoleRepo.On("GetByName", mock.Anything, givenRequest.Name, (*uuid.UUID)(nil)).Return((*domain.Role)(nil), assert.AnError)
	mockRoleRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	role, err := useCase.Execute(context.Background(), givenRequest)

	// Then
	assert.Error(t, err)
	assert.Nil(t, role)
	assert.Contains(t, err.Error(), "invalid permission ID format")
}

func TestCreateRoleUseCase_Execute_WhenAssignPermissionsFails_ReturnsInternalServerError(t *testing.T) {
	// Given
	givenPermID := uuid.New()

	givenRequest := &domain.CreateRoleRequest{
		Name:        "editor",
		Description: "Editor role",
		PermissionIDs: []string{
			givenPermID.String(),
		},
	}

	mockRoleRepo := new(providers.MockRoleRepository)
	mockPermRepo := new(providers.MockPermissionRepository)
	mockLogger := new(providers.MockLogger)

	useCase := NewCreateRoleUseCase(mockRoleRepo, mockPermRepo, mockLogger)

	mockRoleRepo.On("GetByName", mock.Anything, givenRequest.Name, (*uuid.UUID)(nil)).Return((*domain.Role)(nil), assert.AnError)
	mockRoleRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
	mockPermRepo.On("AssignPermissionsToRole", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.Anything).Return(assert.AnError)
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	role, err := useCase.Execute(context.Background(), givenRequest)

	// Then
	assert.Error(t, err)
	assert.Nil(t, role)
	assert.Contains(t, err.Error(), "failed to assign permissions to role")
	mockRoleRepo.AssertExpectations(t)
	mockPermRepo.AssertExpectations(t)
}
