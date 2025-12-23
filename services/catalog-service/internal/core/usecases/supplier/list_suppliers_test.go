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

func TestListSuppliersUseCase_Execute_WithValidRequest_ReturnsSuppliers(t *testing.T) {
	givenOrgID := uuid.New()
	givenSuppliers := []*domain.Supplier{
		{
			ID:                uuid.New(),
			Code:              "SUP-001",
			Name:              "Supplier 1",
			LeadTimeDays:      7,
			ReliabilityRating: 95,
			Status:            domain.SupplierStatusActive,
			OrganizationID:    givenOrgID,
		},
		{
			ID:                uuid.New(),
			Code:              "SUP-002",
			Name:              "Supplier 2",
			LeadTimeDays:      10,
			ReliabilityRating: 90,
			Status:            domain.SupplierStatusActive,
			OrganizationID:    givenOrgID,
		},
	}

	mockRepo := new(MockSupplierRepository)
	mockLogger := logger.New("test", "debug")

	mockRepo.On("List", mock.Anything, mock.Anything, 1, 20).Return(givenSuppliers, int64(2), nil)

	useCase := NewListSuppliersUseCase(mockRepo, mockLogger)

	ctx := context.WithValue(context.Background(), "organization_id", givenOrgID)
	request := &ListSuppliersRequest{
		Page:     1,
		PageSize: 20,
	}

	response, err := useCase.Execute(ctx, request)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, 2, len(response.Suppliers))
	assert.Equal(t, int64(2), response.Total)
	assert.Equal(t, 1, response.Page)
	mockRepo.AssertExpectations(t)
}

func TestListSuppliersUseCase_Execute_WithNilRequest_ReturnsError(t *testing.T) {
	mockRepo := new(MockSupplierRepository)
	mockLogger := logger.New("test", "debug")

	useCase := NewListSuppliersUseCase(mockRepo, mockLogger)

	ctx := context.WithValue(context.Background(), "organization_id", uuid.New())
	response, err := useCase.Execute(ctx, nil)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.True(t, errors.IsBadRequest(err))
	mockRepo.AssertNotCalled(t, "List")
}

func TestListSuppliersUseCase_Execute_WithMissingOrgID_ReturnsError(t *testing.T) {
	mockRepo := new(MockSupplierRepository)
	mockLogger := logger.New("test", "debug")

	useCase := NewListSuppliersUseCase(mockRepo, mockLogger)

	request := &ListSuppliersRequest{
		Page:     1,
		PageSize: 20,
	}

	response, err := useCase.Execute(context.Background(), request)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.True(t, errors.IsBadRequest(err))
	mockRepo.AssertNotCalled(t, "List")
}

func TestListSuppliersUseCase_Execute_WithEmptyList_ReturnsEmptyResponse(t *testing.T) {
	givenOrgID := uuid.New()
	givenSuppliers := []*domain.Supplier{}

	mockRepo := new(MockSupplierRepository)
	mockLogger := logger.New("test", "debug")

	mockRepo.On("List", mock.Anything, mock.Anything, 1, 20).Return(givenSuppliers, int64(0), nil)

	useCase := NewListSuppliersUseCase(mockRepo, mockLogger)

	ctx := context.WithValue(context.Background(), "organization_id", givenOrgID)
	request := &ListSuppliersRequest{
		Page:     1,
		PageSize: 20,
	}

	response, err := useCase.Execute(ctx, request)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, 0, len(response.Suppliers))
	assert.Equal(t, int64(0), response.Total)
	mockRepo.AssertExpectations(t)
}

func TestListSuppliersUseCase_Execute_WithDefaultPagination_UsesDefaults(t *testing.T) {
	givenOrgID := uuid.New()
	givenSuppliers := []*domain.Supplier{}

	mockRepo := new(MockSupplierRepository)
	mockLogger := logger.New("test", "debug")

	mockRepo.On("List", mock.Anything, mock.Anything, 1, 20).Return(givenSuppliers, int64(0), nil)

	useCase := NewListSuppliersUseCase(mockRepo, mockLogger)

	ctx := context.WithValue(context.Background(), "organization_id", givenOrgID)
	request := &ListSuppliersRequest{
		Page:     0,
		PageSize: 0,
	}

	response, err := useCase.Execute(ctx, request)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, 1, response.Page)
	assert.Equal(t, 20, response.PageSize)
	mockRepo.AssertExpectations(t)
}

func TestListSuppliersUseCase_Execute_WithExcessivePageSize_LimitsTo100(t *testing.T) {
	givenOrgID := uuid.New()
	givenSuppliers := []*domain.Supplier{}

	mockRepo := new(MockSupplierRepository)
	mockLogger := logger.New("test", "debug")

	mockRepo.On("List", mock.Anything, mock.Anything, 1, 100).Return(givenSuppliers, int64(0), nil)

	useCase := NewListSuppliersUseCase(mockRepo, mockLogger)

	ctx := context.WithValue(context.Background(), "organization_id", givenOrgID)
	request := &ListSuppliersRequest{
		Page:     1,
		PageSize: 500,
	}

	response, err := useCase.Execute(ctx, request)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, 100, response.PageSize)
	mockRepo.AssertExpectations(t)
}

func TestListSuppliersUseCase_Execute_WithStatusFilter_FiltersCorrectly(t *testing.T) {
	givenOrgID := uuid.New()
	givenSuppliers := []*domain.Supplier{
		{
			ID:                uuid.New(),
			Code:              "SUP-001",
			Name:              "Active Supplier",
			LeadTimeDays:      7,
			ReliabilityRating: 95,
			Status:            domain.SupplierStatusActive,
			OrganizationID:    givenOrgID,
		},
	}

	mockRepo := new(MockSupplierRepository)
	mockLogger := logger.New("test", "debug")

	mockRepo.On("List", mock.Anything, mock.MatchedBy(func(filters map[string]interface{}) bool {
		return filters["status"] == "active"
	}), 1, 20).Return(givenSuppliers, int64(1), nil)

	useCase := NewListSuppliersUseCase(mockRepo, mockLogger)

	ctx := context.WithValue(context.Background(), "organization_id", givenOrgID)
	request := &ListSuppliersRequest{
		Page:     1,
		PageSize: 20,
		Status:   "active",
	}

	response, err := useCase.Execute(ctx, request)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, 1, len(response.Suppliers))
	mockRepo.AssertExpectations(t)
}
