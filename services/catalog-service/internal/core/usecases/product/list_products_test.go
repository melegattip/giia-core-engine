package product

import (
	"context"
	"testing"

	"github.com/giia/giia-core-engine/pkg/logger"
	"github.com/giia/giia-core-engine/services/catalog-service/internal/core/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestListProductsUseCase_Execute_WithValidRequest_ReturnsProducts(t *testing.T) {
	givenOrgID := uuid.New()
	givenProducts := []*domain.Product{
		{
			ID:             uuid.New(),
			SKU:            "WIDGET-001",
			Name:           "Widget 1",
			UnitOfMeasure:  "EA",
			Status:         domain.ProductStatusActive,
			OrganizationID: givenOrgID,
		},
		{
			ID:             uuid.New(),
			SKU:            "WIDGET-002",
			Name:           "Widget 2",
			UnitOfMeasure:  "EA",
			Status:         domain.ProductStatusActive,
			OrganizationID: givenOrgID,
		},
	}

	mockRepo := new(MockProductRepository)
	mockLogger := logger.New("test", "debug")

	mockRepo.On("List", mock.Anything, mock.Anything, 1, 20).Return(givenProducts, int64(2), nil)

	useCase := NewListProductsUseCase(mockRepo, mockLogger)

	ctx := context.WithValue(context.Background(), "organization_id", givenOrgID)
	request := &ListProductsRequest{
		Page:     1,
		PageSize: 20,
	}

	response, err := useCase.Execute(ctx, request)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, 2, len(response.Products))
	assert.Equal(t, int64(2), response.TotalCount)
	assert.Equal(t, 1, response.Page)
	mockRepo.AssertExpectations(t)
}

func TestListProductsUseCase_Execute_WithEmptyList_ReturnsEmptyResponse(t *testing.T) {
	givenOrgID := uuid.New()
	givenProducts := []*domain.Product{}

	mockRepo := new(MockProductRepository)
	mockLogger := logger.New("test", "debug")

	mockRepo.On("List", mock.Anything, mock.Anything, 1, 20).Return(givenProducts, int64(0), nil)

	useCase := NewListProductsUseCase(mockRepo, mockLogger)

	ctx := context.WithValue(context.Background(), "organization_id", givenOrgID)
	request := &ListProductsRequest{
		Page:     1,
		PageSize: 20,
	}

	response, err := useCase.Execute(ctx, request)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, 0, len(response.Products))
	assert.Equal(t, int64(0), response.TotalCount)
	mockRepo.AssertExpectations(t)
}

func TestListProductsUseCase_Execute_WithDefaultPagination_UsesDefaults(t *testing.T) {
	givenOrgID := uuid.New()
	givenProducts := []*domain.Product{}

	mockRepo := new(MockProductRepository)
	mockLogger := logger.New("test", "debug")

	mockRepo.On("List", mock.Anything, mock.Anything, 1, 20).Return(givenProducts, int64(0), nil)

	useCase := NewListProductsUseCase(mockRepo, mockLogger)

	ctx := context.WithValue(context.Background(), "organization_id", givenOrgID)
	request := &ListProductsRequest{
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

func TestListProductsUseCase_Execute_WithExcessivePageSize_LimitsTo100(t *testing.T) {
	givenOrgID := uuid.New()
	givenProducts := []*domain.Product{}

	mockRepo := new(MockProductRepository)
	mockLogger := logger.New("test", "debug")

	mockRepo.On("List", mock.Anything, mock.Anything, 1, 100).Return(givenProducts, int64(0), nil)

	useCase := NewListProductsUseCase(mockRepo, mockLogger)

	ctx := context.WithValue(context.Background(), "organization_id", givenOrgID)
	request := &ListProductsRequest{
		Page:     1,
		PageSize: 500,
	}

	response, err := useCase.Execute(ctx, request)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, 100, response.PageSize)
	mockRepo.AssertExpectations(t)
}

func TestListProductsUseCase_Execute_WithStatusFilter_FiltersCorrectly(t *testing.T) {
	givenOrgID := uuid.New()
	givenProducts := []*domain.Product{
		{
			ID:             uuid.New(),
			SKU:            "WIDGET-001",
			Name:           "Widget 1",
			UnitOfMeasure:  "EA",
			Status:         domain.ProductStatusActive,
			OrganizationID: givenOrgID,
		},
	}

	mockRepo := new(MockProductRepository)
	mockLogger := logger.New("test", "debug")

	mockRepo.On("List", mock.Anything, mock.MatchedBy(func(filters map[string]interface{}) bool {
		return filters["status"] == "active"
	}), 1, 20).Return(givenProducts, int64(1), nil)

	useCase := NewListProductsUseCase(mockRepo, mockLogger)

	ctx := context.WithValue(context.Background(), "organization_id", givenOrgID)
	request := &ListProductsRequest{
		Page:     1,
		PageSize: 20,
		Status:   "active",
	}

	response, err := useCase.Execute(ctx, request)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, 1, len(response.Products))
	mockRepo.AssertExpectations(t)
}
