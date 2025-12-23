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

func TestGetSupplierUseCase_Execute_WithValidID_ReturnsSupplier(t *testing.T) {
	givenSupplierID := uuid.New()
	givenOrgID := uuid.New()
	givenSupplier := &domain.Supplier{
		ID:                givenSupplierID,
		Code:              "SUP-001",
		Name:              "Premium Supplier",
		LeadTimeDays:      7,
		ReliabilityRating: 95,
		Status:            domain.SupplierStatusActive,
		OrganizationID:    givenOrgID,
	}

	mockRepo := new(MockSupplierRepository)
	mockLogger := logger.New("test", "debug")

	mockRepo.On("GetByID", mock.Anything, givenSupplierID).Return(givenSupplier, nil)

	useCase := NewGetSupplierUseCase(mockRepo, mockLogger)

	ctx := context.WithValue(context.Background(), "organization_id", givenOrgID)
	supplier, err := useCase.Execute(ctx, givenSupplierID)

	assert.NoError(t, err)
	assert.NotNil(t, supplier)
	assert.Equal(t, givenSupplierID, supplier.ID)
	assert.Equal(t, "SUP-001", supplier.Code)
	assert.Equal(t, "Premium Supplier", supplier.Name)
	mockRepo.AssertExpectations(t)
}

func TestGetSupplierUseCase_Execute_WithNilID_ReturnsError(t *testing.T) {
	mockRepo := new(MockSupplierRepository)
	mockLogger := logger.New("test", "debug")

	useCase := NewGetSupplierUseCase(mockRepo, mockLogger)

	ctx := context.WithValue(context.Background(), "organization_id", uuid.New())
	supplier, err := useCase.Execute(ctx, uuid.Nil)

	assert.Error(t, err)
	assert.Nil(t, supplier)
	assert.True(t, errors.IsBadRequest(err))
	mockRepo.AssertNotCalled(t, "GetByID")
}

func TestGetSupplierUseCase_Execute_WithMissingOrgID_ReturnsError(t *testing.T) {
	mockRepo := new(MockSupplierRepository)
	mockLogger := logger.New("test", "debug")

	useCase := NewGetSupplierUseCase(mockRepo, mockLogger)

	supplier, err := useCase.Execute(context.Background(), uuid.New())

	assert.Error(t, err)
	assert.Nil(t, supplier)
	assert.True(t, errors.IsBadRequest(err))
	mockRepo.AssertNotCalled(t, "GetByID")
}

func TestGetSupplierUseCase_Execute_WithNonExistentSupplier_ReturnsError(t *testing.T) {
	givenSupplierID := uuid.New()
	givenOrgID := uuid.New()
	givenNotFoundError := errors.NewNotFound("supplier not found")

	mockRepo := new(MockSupplierRepository)
	mockLogger := logger.New("test", "debug")

	mockRepo.On("GetByID", mock.Anything, givenSupplierID).Return(nil, givenNotFoundError)

	useCase := NewGetSupplierUseCase(mockRepo, mockLogger)

	ctx := context.WithValue(context.Background(), "organization_id", givenOrgID)
	supplier, err := useCase.Execute(ctx, givenSupplierID)

	assert.Error(t, err)
	assert.Nil(t, supplier)
	assert.True(t, errors.IsNotFound(err))
	mockRepo.AssertExpectations(t)
}
