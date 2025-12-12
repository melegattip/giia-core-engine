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

func TestAssignRoleUseCase_Execute_WithValidRequest_AssignsRoleSuccessfully(t *testing.T) {
	// Given
	givenUserID := uuid.New()
	givenRoleID := uuid.New()
	givenAssignedBy := uuid.New()

	givenUser := &domain.User{
		ID:     givenUserID,
		Email:  "user@example.com",
		Status: domain.UserStatusActive,
	}

	givenRole := &domain.Role{
		ID:   givenRoleID,
		Name: "editor",
	}

	mockRoleRepo := new(providers.MockRoleRepository)
	mockUserRepo := new(providers.MockUserRepository)
	mockCache := new(providers.MockPermissionCache)
	mockLogger := new(providers.MockLogger)

	useCase := NewAssignRoleUseCase(mockRoleRepo, mockUserRepo, mockCache, mockLogger)

	mockUserRepo.On("GetByID", mock.Anything, givenUserID).Return(givenUser, nil)
	mockRoleRepo.On("GetByID", mock.Anything, givenRoleID).Return(givenRole, nil)
	mockRoleRepo.On("AssignRoleToUser", mock.Anything, givenUserID, givenRoleID, givenAssignedBy).Return(nil)
	mockCache.On("InvalidateUserPermissions", mock.Anything, givenUserID.String()).Return(nil)
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	err := useCase.Execute(context.Background(), givenUserID, givenRoleID, givenAssignedBy)

	// Then
	assert.NoError(t, err)
	mockUserRepo.AssertExpectations(t)
	mockRoleRepo.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestAssignRoleUseCase_Execute_WithNilUserID_ReturnsBadRequest(t *testing.T) {
	// Given
	givenRoleID := uuid.New()
	givenAssignedBy := uuid.New()

	mockRoleRepo := new(providers.MockRoleRepository)
	mockUserRepo := new(providers.MockUserRepository)
	mockCache := new(providers.MockPermissionCache)
	mockLogger := new(providers.MockLogger)

	useCase := NewAssignRoleUseCase(mockRoleRepo, mockUserRepo, mockCache, mockLogger)

	// When
	err := useCase.Execute(context.Background(), uuid.Nil, givenRoleID, givenAssignedBy)

	// Then
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user ID cannot be empty")
}

func TestAssignRoleUseCase_Execute_WithNilRoleID_ReturnsBadRequest(t *testing.T) {
	// Given
	givenUserID := uuid.New()
	givenAssignedBy := uuid.New()

	mockRoleRepo := new(providers.MockRoleRepository)
	mockUserRepo := new(providers.MockUserRepository)
	mockCache := new(providers.MockPermissionCache)
	mockLogger := new(providers.MockLogger)

	useCase := NewAssignRoleUseCase(mockRoleRepo, mockUserRepo, mockCache, mockLogger)

	// When
	err := useCase.Execute(context.Background(), givenUserID, uuid.Nil, givenAssignedBy)

	// Then
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "role ID cannot be empty")
}

func TestAssignRoleUseCase_Execute_WithNilAssignedBy_ReturnsBadRequest(t *testing.T) {
	// Given
	givenUserID := uuid.New()
	givenRoleID := uuid.New()

	mockRoleRepo := new(providers.MockRoleRepository)
	mockUserRepo := new(providers.MockUserRepository)
	mockCache := new(providers.MockPermissionCache)
	mockLogger := new(providers.MockLogger)

	useCase := NewAssignRoleUseCase(mockRoleRepo, mockUserRepo, mockCache, mockLogger)

	// When
	err := useCase.Execute(context.Background(), givenUserID, givenRoleID, uuid.Nil)

	// Then
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "assigned by user ID cannot be empty")
}

func TestAssignRoleUseCase_Execute_WithNonExistentUser_ReturnsNotFound(t *testing.T) {
	// Given
	givenUserID := uuid.New()
	givenRoleID := uuid.New()
	givenAssignedBy := uuid.New()

	mockRoleRepo := new(providers.MockRoleRepository)
	mockUserRepo := new(providers.MockUserRepository)
	mockCache := new(providers.MockPermissionCache)
	mockLogger := new(providers.MockLogger)

	useCase := NewAssignRoleUseCase(mockRoleRepo, mockUserRepo, mockCache, mockLogger)

	mockUserRepo.On("GetByID", mock.Anything, givenUserID).Return((*domain.User)(nil), assert.AnError)
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	err := useCase.Execute(context.Background(), givenUserID, givenRoleID, givenAssignedBy)

	// Then
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
	mockUserRepo.AssertExpectations(t)
}

func TestAssignRoleUseCase_Execute_WithNonExistentRole_ReturnsNotFound(t *testing.T) {
	// Given
	givenUserID := uuid.New()
	givenRoleID := uuid.New()
	givenAssignedBy := uuid.New()

	givenUser := &domain.User{
		ID:     givenUserID,
		Email:  "user@example.com",
		Status: domain.UserStatusActive,
	}

	mockRoleRepo := new(providers.MockRoleRepository)
	mockUserRepo := new(providers.MockUserRepository)
	mockCache := new(providers.MockPermissionCache)
	mockLogger := new(providers.MockLogger)

	useCase := NewAssignRoleUseCase(mockRoleRepo, mockUserRepo, mockCache, mockLogger)

	mockUserRepo.On("GetByID", mock.Anything, givenUserID).Return(givenUser, nil)
	mockRoleRepo.On("GetByID", mock.Anything, givenRoleID).Return((*domain.Role)(nil), assert.AnError)
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	err := useCase.Execute(context.Background(), givenUserID, givenRoleID, givenAssignedBy)

	// Then
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "role not found")
	mockUserRepo.AssertExpectations(t)
	mockRoleRepo.AssertExpectations(t)
}

func TestAssignRoleUseCase_Execute_WhenAssignmentFails_ReturnsInternalServerError(t *testing.T) {
	// Given
	givenUserID := uuid.New()
	givenRoleID := uuid.New()
	givenAssignedBy := uuid.New()

	givenUser := &domain.User{
		ID:     givenUserID,
		Email:  "user@example.com",
		Status: domain.UserStatusActive,
	}

	givenRole := &domain.Role{
		ID:   givenRoleID,
		Name: "editor",
	}

	mockRoleRepo := new(providers.MockRoleRepository)
	mockUserRepo := new(providers.MockUserRepository)
	mockCache := new(providers.MockPermissionCache)
	mockLogger := new(providers.MockLogger)

	useCase := NewAssignRoleUseCase(mockRoleRepo, mockUserRepo, mockCache, mockLogger)

	mockUserRepo.On("GetByID", mock.Anything, givenUserID).Return(givenUser, nil)
	mockRoleRepo.On("GetByID", mock.Anything, givenRoleID).Return(givenRole, nil)
	mockRoleRepo.On("AssignRoleToUser", mock.Anything, givenUserID, givenRoleID, givenAssignedBy).Return(assert.AnError)
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	err := useCase.Execute(context.Background(), givenUserID, givenRoleID, givenAssignedBy)

	// Then
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to assign role to user")
	mockUserRepo.AssertExpectations(t)
	mockRoleRepo.AssertExpectations(t)
}

func TestAssignRoleUseCase_Execute_WhenCacheInvalidationFails_StillReturnsSuccess(t *testing.T) {
	// Given
	givenUserID := uuid.New()
	givenRoleID := uuid.New()
	givenAssignedBy := uuid.New()

	givenUser := &domain.User{
		ID:     givenUserID,
		Email:  "user@example.com",
		Status: domain.UserStatusActive,
	}

	givenRole := &domain.Role{
		ID:   givenRoleID,
		Name: "editor",
	}

	mockRoleRepo := new(providers.MockRoleRepository)
	mockUserRepo := new(providers.MockUserRepository)
	mockCache := new(providers.MockPermissionCache)
	mockLogger := new(providers.MockLogger)

	useCase := NewAssignRoleUseCase(mockRoleRepo, mockUserRepo, mockCache, mockLogger)

	mockUserRepo.On("GetByID", mock.Anything, givenUserID).Return(givenUser, nil)
	mockRoleRepo.On("GetByID", mock.Anything, givenRoleID).Return(givenRole, nil)
	mockRoleRepo.On("AssignRoleToUser", mock.Anything, givenUserID, givenRoleID, givenAssignedBy).Return(nil)
	mockCache.On("InvalidateUserPermissions", mock.Anything, givenUserID.String()).Return(assert.AnError)
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	err := useCase.Execute(context.Background(), givenUserID, givenRoleID, givenAssignedBy)

	// Then
	assert.NoError(t, err)
	mockUserRepo.AssertExpectations(t)
	mockRoleRepo.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}
