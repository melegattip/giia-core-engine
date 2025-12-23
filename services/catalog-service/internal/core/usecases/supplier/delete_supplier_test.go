package supplier

import (
	"context"
	"testing"

	"github.com/giia/giia-core-engine/pkg/errors"
	"github.com/giia/giia-core-engine/pkg/logger"
	"github.com/giia/giia-core-engine/services/catalog-service/internal/core/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDeleteSupplierUseCase_Execute_WithValidID_DeletesSupplier(t *testing.T) {
	givenSupplierID := uuid.New()
	givenOrgID := uuid.New()
	givenSupplier := &domain.Supplier{
		ID:             givenSupplierID,
		Code:           "SUP-001",
		Name:           "Test Supplier",
		Status:         domain.SupplierStatusActive,
		OrganizationID: givenOrgID,
	}

	mockRepo := new(MockSupplierRepository)
	mockPublisher := new(MockSupplierEventPublisher)
	mockLogger := logger.New("test", "debug")

	mockRepo.On("GetByID", mock.Anything, givenSupplierID).Return(givenSupplier, nil)
	mockRepo.On("Delete", mock.Anything, givenSupplierID).Return(nil)
	mockPublisher.On("PublishSupplierDeleted", mock.Anything, mock.MatchedBy(func(s *domain.Supplier) bool {
		return s.ID == givenSupplierID
	})).Return(nil)

	useCase := NewDeleteSupplierUseCase(mockRepo, mockPublisher, mockLogger)

	ctx := context.WithValue(context.Background(), "organization_id", givenOrgID)
	err := useCase.Execute(ctx, givenSupplierID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockPublisher.AssertExpectations(t)
}

func TestDeleteSupplierUseCase_Execute_WithNilID_ReturnsError(t *testing.T) {
	mockRepo := new(MockSupplierRepository)
	mockPublisher := new(MockSupplierEventPublisher)
	mockLogger := logger.New("test", "debug")

	useCase := NewDeleteSupplierUseCase(mockRepo, mockPublisher, mockLogger)

	ctx := context.WithValue(context.Background(), "organization_id", uuid.New())
	err := useCase.Execute(ctx, uuid.Nil)

	assert.Error(t, err)
	assert.True(t, errors.IsBadRequest(err))
	mockRepo.AssertNotCalled(t, "GetByID")
	mockRepo.AssertNotCalled(t, "Delete")
	mockPublisher.AssertNotCalled(t, "PublishSupplierDeleted")
}

func TestDeleteSupplierUseCase_Execute_WithMissingOrgID_ReturnsError(t *testing.T) {
	mockRepo := new(MockSupplierRepository)
	mockPublisher := new(MockSupplierEventPublisher)
	mockLogger := logger.New("test", "debug")

	useCase := NewDeleteSupplierUseCase(mockRepo, mockPublisher, mockLogger)

	err := useCase.Execute(context.Background(), uuid.New())

	assert.Error(t, err)
	assert.True(t, errors.IsBadRequest(err))
	mockRepo.AssertNotCalled(t, "GetByID")
	mockRepo.AssertNotCalled(t, "Delete")
}

func TestDeleteSupplierUseCase_Execute_WithNonExistentSupplier_ReturnsError(t *testing.T) {
	givenSupplierID := uuid.New()
	givenOrgID := uuid.New()
	givenNotFoundError := errors.NewNotFound("supplier not found")

	mockRepo := new(MockSupplierRepository)
	mockPublisher := new(MockSupplierEventPublisher)
	mockLogger := logger.New("test", "debug")

	mockRepo.On("GetByID", mock.Anything, givenSupplierID).Return(nil, givenNotFoundError)

	useCase := NewDeleteSupplierUseCase(mockRepo, mockPublisher, mockLogger)

	ctx := context.WithValue(context.Background(), "organization_id", givenOrgID)
	err := useCase.Execute(ctx, givenSupplierID)

	assert.Error(t, err)
	assert.True(t, errors.IsNotFound(err))
	mockRepo.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "Delete")
	mockPublisher.AssertNotCalled(t, "PublishSupplierDeleted")
}

func TestDeleteSupplierUseCase_Execute_WithRepositoryError_ReturnsError(t *testing.T) {
	givenSupplierID := uuid.New()
	givenOrgID := uuid.New()
	givenSupplier := &domain.Supplier{
		ID:             givenSupplierID,
		Code:           "SUP-001",
		Name:           "Test Supplier",
		Status:         domain.SupplierStatusActive,
		OrganizationID: givenOrgID,
	}
	givenDatabaseError := errors.NewInternalServerError("database error")

	mockRepo := new(MockSupplierRepository)
	mockPublisher := new(MockSupplierEventPublisher)
	mockLogger := logger.New("test", "debug")

	mockRepo.On("GetByID", mock.Anything, givenSupplierID).Return(givenSupplier, nil)
	mockRepo.On("Delete", mock.Anything, givenSupplierID).Return(givenDatabaseError)

	useCase := NewDeleteSupplierUseCase(mockRepo, mockPublisher, mockLogger)

	ctx := context.WithValue(context.Background(), "organization_id", givenOrgID)
	err := useCase.Execute(ctx, givenSupplierID)

	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
	mockPublisher.AssertNotCalled(t, "PublishSupplierDeleted")
}

func TestDeleteSupplierUseCase_Execute_WithEventPublishError_StillSucceeds(t *testing.T) {
	givenSupplierID := uuid.New()
	givenOrgID := uuid.New()
	givenSupplier := &domain.Supplier{
		ID:             givenSupplierID,
		Code:           "SUP-001",
		Name:           "Test Supplier",
		Status:         domain.SupplierStatusActive,
		OrganizationID: givenOrgID,
	}
	givenEventError := errors.NewInternalServerError("event publish failed")

	mockRepo := new(MockSupplierRepository)
	mockPublisher := new(MockSupplierEventPublisher)
	mockLogger := logger.New("test", "debug")

	mockRepo.On("GetByID", mock.Anything, givenSupplierID).Return(givenSupplier, nil)
	mockRepo.On("Delete", mock.Anything, givenSupplierID).Return(nil)
	mockPublisher.On("PublishSupplierDeleted", mock.Anything, mock.Anything).Return(givenEventError)

	useCase := NewDeleteSupplierUseCase(mockRepo, mockPublisher, mockLogger)

	ctx := context.WithValue(context.Background(), "organization_id", givenOrgID)
	err := useCase.Execute(ctx, givenSupplierID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockPublisher.AssertExpectations(t)
}
