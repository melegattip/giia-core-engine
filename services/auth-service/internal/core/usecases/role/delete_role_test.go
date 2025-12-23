package role

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/melegattip/giia-core-engine/services/auth-service/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/auth-service/internal/core/providers"
)

func TestDeleteRoleUseCase_Execute_WithValidRole_DeletesRoleSuccessfully(t *testing.T) {
	// Given
	givenRoleID := uuid.New()

	givenRole := &domain.Role{
		ID:       givenRoleID,
		Name:     "editor",
		IsSystem: false,
	}

	mockRoleRepo := new(providers.MockRoleRepository)
	mockCache := new(providers.MockPermissionCache)
	mockLogger := new(providers.MockLogger)

	useCase := NewDeleteRoleUseCase(mockRoleRepo, mockCache, mockLogger)

	mockRoleRepo.On("GetByID", mock.Anything, givenRoleID).Return(givenRole, nil)
	mockRoleRepo.On("GetUsersWithRole", mock.Anything, givenRoleID).Return([]uuid.UUID{}, nil)
	mockRoleRepo.On("Delete", mock.Anything, givenRoleID).Return(nil)
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	err := useCase.Execute(context.Background(), givenRoleID)

	// Then
	assert.NoError(t, err)
	mockRoleRepo.AssertExpectations(t)
}

func TestDeleteRoleUseCase_Execute_WithNilRoleID_ReturnsBadRequest(t *testing.T) {
	// Given
	mockRoleRepo := new(providers.MockRoleRepository)
	mockCache := new(providers.MockPermissionCache)
	mockLogger := new(providers.MockLogger)

	useCase := NewDeleteRoleUseCase(mockRoleRepo, mockCache, mockLogger)

	// When
	err := useCase.Execute(context.Background(), uuid.Nil)

	// Then
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "role ID cannot be empty")
}

func TestDeleteRoleUseCase_Execute_WithNonExistentRole_ReturnsNotFound(t *testing.T) {
	// Given
	givenRoleID := uuid.New()

	mockRoleRepo := new(providers.MockRoleRepository)
	mockCache := new(providers.MockPermissionCache)
	mockLogger := new(providers.MockLogger)

	useCase := NewDeleteRoleUseCase(mockRoleRepo, mockCache, mockLogger)

	mockRoleRepo.On("GetByID", mock.Anything, givenRoleID).Return((*domain.Role)(nil), assert.AnError)
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	err := useCase.Execute(context.Background(), givenRoleID)

	// Then
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "role not found")
	mockRoleRepo.AssertExpectations(t)
}

func TestDeleteRoleUseCase_Execute_WithSystemRole_ReturnsBadRequest(t *testing.T) {
	// Given
	givenRoleID := uuid.New()

	givenRole := &domain.Role{
		ID:       givenRoleID,
		Name:     "admin",
		IsSystem: true,
	}

	mockRoleRepo := new(providers.MockRoleRepository)
	mockCache := new(providers.MockPermissionCache)
	mockLogger := new(providers.MockLogger)

	useCase := NewDeleteRoleUseCase(mockRoleRepo, mockCache, mockLogger)

	mockRoleRepo.On("GetByID", mock.Anything, givenRoleID).Return(givenRole, nil)

	// When
	err := useCase.Execute(context.Background(), givenRoleID)

	// Then
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot delete system roles")
	mockRoleRepo.AssertExpectations(t)
}

func TestDeleteRoleUseCase_Execute_WhenGetUsersWithRoleFails_ReturnsInternalServerError(t *testing.T) {
	// Given
	givenRoleID := uuid.New()

	givenRole := &domain.Role{
		ID:       givenRoleID,
		Name:     "editor",
		IsSystem: false,
	}

	mockRoleRepo := new(providers.MockRoleRepository)
	mockCache := new(providers.MockPermissionCache)
	mockLogger := new(providers.MockLogger)

	useCase := NewDeleteRoleUseCase(mockRoleRepo, mockCache, mockLogger)

	mockRoleRepo.On("GetByID", mock.Anything, givenRoleID).Return(givenRole, nil)
	mockRoleRepo.On("GetUsersWithRole", mock.Anything, givenRoleID).Return(([]uuid.UUID)(nil), assert.AnError)
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	err := useCase.Execute(context.Background(), givenRoleID)

	// Then
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to verify role usage")
	mockRoleRepo.AssertExpectations(t)
}

func TestDeleteRoleUseCase_Execute_WhenDeleteFails_ReturnsInternalServerError(t *testing.T) {
	// Given
	givenRoleID := uuid.New()

	givenRole := &domain.Role{
		ID:       givenRoleID,
		Name:     "editor",
		IsSystem: false,
	}

	mockRoleRepo := new(providers.MockRoleRepository)
	mockCache := new(providers.MockPermissionCache)
	mockLogger := new(providers.MockLogger)

	useCase := NewDeleteRoleUseCase(mockRoleRepo, mockCache, mockLogger)

	mockRoleRepo.On("GetByID", mock.Anything, givenRoleID).Return(givenRole, nil)
	mockRoleRepo.On("GetUsersWithRole", mock.Anything, givenRoleID).Return([]uuid.UUID{}, nil)
	mockRoleRepo.On("Delete", mock.Anything, givenRoleID).Return(assert.AnError)
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	err := useCase.Execute(context.Background(), givenRoleID)

	// Then
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete role")
	mockRoleRepo.AssertExpectations(t)
}

func TestDeleteRoleUseCase_Execute_WithAffectedUsers_InvalidatesCache(t *testing.T) {
	// Given
	givenRoleID := uuid.New()
	givenUserID1 := uuid.New()
	givenUserID2 := uuid.New()

	givenRole := &domain.Role{
		ID:       givenRoleID,
		Name:     "editor",
		IsSystem: false,
	}

	mockRoleRepo := new(providers.MockRoleRepository)
	mockCache := new(providers.MockPermissionCache)
	mockLogger := new(providers.MockLogger)

	useCase := NewDeleteRoleUseCase(mockRoleRepo, mockCache, mockLogger)

	mockRoleRepo.On("GetByID", mock.Anything, givenRoleID).Return(givenRole, nil)
	mockRoleRepo.On("GetUsersWithRole", mock.Anything, givenRoleID).Return([]uuid.UUID{givenUserID1, givenUserID2}, nil)
	mockRoleRepo.On("Delete", mock.Anything, givenRoleID).Return(nil)
	mockCache.On("InvalidateUsersWithRole", mock.Anything, mock.MatchedBy(func(userIDs []string) bool {
		return len(userIDs) == 2
	})).Return(nil)
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	err := useCase.Execute(context.Background(), givenRoleID)

	// Then
	assert.NoError(t, err)
	mockRoleRepo.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestDeleteRoleUseCase_Execute_WhenCacheInvalidationFails_StillReturnsSuccess(t *testing.T) {
	// Given
	givenRoleID := uuid.New()
	givenUserID := uuid.New()

	givenRole := &domain.Role{
		ID:       givenRoleID,
		Name:     "editor",
		IsSystem: false,
	}

	mockRoleRepo := new(providers.MockRoleRepository)
	mockCache := new(providers.MockPermissionCache)
	mockLogger := new(providers.MockLogger)

	useCase := NewDeleteRoleUseCase(mockRoleRepo, mockCache, mockLogger)

	mockRoleRepo.On("GetByID", mock.Anything, givenRoleID).Return(givenRole, nil)
	mockRoleRepo.On("GetUsersWithRole", mock.Anything, givenRoleID).Return([]uuid.UUID{givenUserID}, nil)
	mockRoleRepo.On("Delete", mock.Anything, givenRoleID).Return(nil)
	mockCache.On("InvalidateUsersWithRole", mock.Anything, mock.Anything).Return(assert.AnError)
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

	// When
	err := useCase.Execute(context.Background(), givenRoleID)

	// Then
	assert.NoError(t, err)
	mockRoleRepo.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}
