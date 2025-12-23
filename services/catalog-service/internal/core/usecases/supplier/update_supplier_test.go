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

func TestUpdateSupplierUseCase_Execute_WithValidData_ReturnsUpdatedSupplier(t *testing.T) {
	givenSupplierID := uuid.New()
	givenOrgID := uuid.New()
	givenExistingSupplier := &domain.Supplier{
		ID:                givenSupplierID,
		Code:              "SUP-001",
		Name:              "Old Name",
		LeadTimeDays:      5,
		ReliabilityRating: 80,
		Status:            domain.SupplierStatusActive,
		OrganizationID:    givenOrgID,
	}

	mockRepo := new(MockSupplierRepository)
	mockPublisher := new(MockSupplierEventPublisher)
	mockLogger := logger.New("test", "debug")

	mockRepo.On("GetByID", mock.Anything, givenSupplierID).Return(givenExistingSupplier, nil)
	mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(s *domain.Supplier) bool {
		return s.ID == givenSupplierID && s.Name == "New Name"
	})).Return(nil)
	mockPublisher.On("PublishSupplierUpdated", mock.Anything, mock.Anything).Return(nil)

	useCase := NewUpdateSupplierUseCase(mockRepo, mockPublisher, mockLogger)

	ctx := context.WithValue(context.Background(), "organization_id", givenOrgID)
	request := &UpdateSupplierRequest{
		ID:                givenSupplierID,
		Name:              "New Name",
		LeadTimeDays:      10,
		ReliabilityRating: 95,
	}

	supplier, err := useCase.Execute(ctx, request)

	assert.NoError(t, err)
	assert.NotNil(t, supplier)
	assert.Equal(t, "New Name", supplier.Name)
	assert.Equal(t, 10, supplier.LeadTimeDays)
	assert.Equal(t, 95, supplier.ReliabilityRating)
	mockRepo.AssertExpectations(t)
	mockPublisher.AssertExpectations(t)
}

func TestUpdateSupplierUseCase_Execute_WithNilRequest_ReturnsError(t *testing.T) {
	mockRepo := new(MockSupplierRepository)
	mockPublisher := new(MockSupplierEventPublisher)
	mockLogger := logger.New("test", "debug")

	useCase := NewUpdateSupplierUseCase(mockRepo, mockPublisher, mockLogger)

	ctx := context.WithValue(context.Background(), "organization_id", uuid.New())
	supplier, err := useCase.Execute(ctx, nil)

	assert.Error(t, err)
	assert.Nil(t, supplier)
	assert.True(t, errors.IsBadRequest(err))
	mockRepo.AssertNotCalled(t, "GetByID")
}

func TestUpdateSupplierUseCase_Execute_WithNilSupplierID_ReturnsError(t *testing.T) {
	mockRepo := new(MockSupplierRepository)
	mockPublisher := new(MockSupplierEventPublisher)
	mockLogger := logger.New("test", "debug")

	useCase := NewUpdateSupplierUseCase(mockRepo, mockPublisher, mockLogger)

	ctx := context.WithValue(context.Background(), "organization_id", uuid.New())
	request := &UpdateSupplierRequest{
		ID:                uuid.Nil,
		Name:              "New Name",
		LeadTimeDays:      10,
		ReliabilityRating: 95,
	}

	supplier, err := useCase.Execute(ctx, request)

	assert.Error(t, err)
	assert.Nil(t, supplier)
	assert.True(t, errors.IsBadRequest(err))
	mockRepo.AssertNotCalled(t, "GetByID")
}

func TestUpdateSupplierUseCase_Execute_WithMissingOrgID_ReturnsError(t *testing.T) {
	mockRepo := new(MockSupplierRepository)
	mockPublisher := new(MockSupplierEventPublisher)
	mockLogger := logger.New("test", "debug")

	useCase := NewUpdateSupplierUseCase(mockRepo, mockPublisher, mockLogger)

	request := &UpdateSupplierRequest{
		ID:                uuid.New(),
		Name:              "New Name",
		LeadTimeDays:      10,
		ReliabilityRating: 95,
	}

	supplier, err := useCase.Execute(context.Background(), request)

	assert.Error(t, err)
	assert.Nil(t, supplier)
	assert.True(t, errors.IsBadRequest(err))
	mockRepo.AssertNotCalled(t, "GetByID")
}

func TestUpdateSupplierUseCase_Execute_WithNonExistentSupplier_ReturnsError(t *testing.T) {
	givenSupplierID := uuid.New()
	givenOrgID := uuid.New()
	givenNotFoundError := errors.NewNotFound("supplier not found")

	mockRepo := new(MockSupplierRepository)
	mockPublisher := new(MockSupplierEventPublisher)
	mockLogger := logger.New("test", "debug")

	mockRepo.On("GetByID", mock.Anything, givenSupplierID).Return(nil, givenNotFoundError)

	useCase := NewUpdateSupplierUseCase(mockRepo, mockPublisher, mockLogger)

	ctx := context.WithValue(context.Background(), "organization_id", givenOrgID)
	request := &UpdateSupplierRequest{
		ID:                givenSupplierID,
		Name:              "New Name",
		LeadTimeDays:      10,
		ReliabilityRating: 95,
	}

	supplier, err := useCase.Execute(ctx, request)

	assert.Error(t, err)
	assert.Nil(t, supplier)
	assert.True(t, errors.IsNotFound(err))
	mockRepo.AssertExpectations(t)
	mockPublisher.AssertNotCalled(t, "PublishSupplierUpdated")
}

func TestUpdateSupplierUseCase_Execute_WithInvalidData_ReturnsValidationError(t *testing.T) {
	givenSupplierID := uuid.New()
	givenOrgID := uuid.New()
	givenExistingSupplier := &domain.Supplier{
		ID:                givenSupplierID,
		Code:              "SUP-001",
		Name:              "Old Name",
		LeadTimeDays:      5,
		ReliabilityRating: 80,
		Status:            domain.SupplierStatusActive,
		OrganizationID:    givenOrgID,
	}

	mockRepo := new(MockSupplierRepository)
	mockPublisher := new(MockSupplierEventPublisher)
	mockLogger := logger.New("test", "debug")

	mockRepo.On("GetByID", mock.Anything, givenSupplierID).Return(givenExistingSupplier, nil)

	useCase := NewUpdateSupplierUseCase(mockRepo, mockPublisher, mockLogger)

	ctx := context.WithValue(context.Background(), "organization_id", givenOrgID)
	request := &UpdateSupplierRequest{
		ID:                givenSupplierID,
		Name:              "New Name",
		LeadTimeDays:      10,
		ReliabilityRating: 150,
	}

	supplier, err := useCase.Execute(ctx, request)

	assert.Error(t, err)
	assert.Nil(t, supplier)
	assert.True(t, errors.IsBadRequest(err))
	mockRepo.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "Update")
	mockPublisher.AssertNotCalled(t, "PublishSupplierUpdated")
}
