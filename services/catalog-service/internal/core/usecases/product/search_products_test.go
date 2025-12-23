package product

import (
	"context"
	"testing"

	"github.com/melegattip/giia-core-engine/pkg/errors"
	"github.com/melegattip/giia-core-engine/pkg/logger"
	"github.com/melegattip/giia-core-engine/services/catalog-service/internal/core/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSearchProductsUseCase_Execute_WithValidQuery_ReturnsProducts(t *testing.T) {
	givenOrgID := uuid.New()
	givenQuery := "widget"
	givenProducts := []*domain.Product{
		{
			ID:             uuid.New(),
			SKU:            "WIDGET-001",
			Name:           "Premium Widget",
			UnitOfMeasure:  "EA",
			Status:         domain.ProductStatusActive,
			OrganizationID: givenOrgID,
		},
	}

	mockRepo := new(MockProductRepository)
	mockLogger := logger.New("test", "debug")

	mockRepo.On("Search", mock.Anything, givenQuery, mock.Anything, 1, 20).Return(givenProducts, int64(1), nil)

	useCase := NewSearchProductsUseCase(mockRepo, mockLogger)

	ctx := context.WithValue(context.Background(), "organization_id", givenOrgID)
	request := &SearchProductsRequest{
		Query:    givenQuery,
		Page:     1,
		PageSize: 20,
	}

	response, err := useCase.Execute(ctx, request)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, 1, len(response.Products))
	assert.Equal(t, int64(1), response.TotalCount)
	mockRepo.AssertExpectations(t)
}

func TestSearchProductsUseCase_Execute_WithEmptyQuery_ReturnsEmptyResults(t *testing.T) {
	givenOrgID := uuid.New()
	givenQuery := ""
	givenProducts := []*domain.Product{}

	mockRepo := new(MockProductRepository)
	mockLogger := logger.New("test", "debug")

	mockRepo.On("Search", mock.Anything, givenQuery, mock.Anything, 1, 20).Return(givenProducts, int64(0), nil)

	useCase := NewSearchProductsUseCase(mockRepo, mockLogger)

	ctx := context.WithValue(context.Background(), "organization_id", givenOrgID)
	request := &SearchProductsRequest{
		Query:    givenQuery,
		Page:     1,
		PageSize: 20,
	}

	response, err := useCase.Execute(ctx, request)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, 0, len(response.Products))
	mockRepo.AssertExpectations(t)
}

func TestSearchProductsUseCase_Execute_WithDefaultPagination_UsesDefaults(t *testing.T) {
	givenOrgID := uuid.New()
	givenQuery := "widget"
	givenProducts := []*domain.Product{}

	mockRepo := new(MockProductRepository)
	mockLogger := logger.New("test", "debug")

	mockRepo.On("Search", mock.Anything, givenQuery, mock.Anything, 1, 20).Return(givenProducts, int64(0), nil)

	useCase := NewSearchProductsUseCase(mockRepo, mockLogger)

	ctx := context.WithValue(context.Background(), "organization_id", givenOrgID)
	request := &SearchProductsRequest{
		Query:    givenQuery,
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

func TestSearchProductsUseCase_Execute_WithExcessivePageSize_LimitsTo100(t *testing.T) {
	givenOrgID := uuid.New()
	givenQuery := "widget"
	givenProducts := []*domain.Product{}

	mockRepo := new(MockProductRepository)
	mockLogger := logger.New("test", "debug")

	mockRepo.On("Search", mock.Anything, givenQuery, mock.Anything, 1, 100).Return(givenProducts, int64(0), nil)

	useCase := NewSearchProductsUseCase(mockRepo, mockLogger)

	ctx := context.WithValue(context.Background(), "organization_id", givenOrgID)
	request := &SearchProductsRequest{
		Query:    givenQuery,
		Page:     1,
		PageSize: 500,
	}

	response, err := useCase.Execute(ctx, request)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, 100, response.PageSize)
	mockRepo.AssertExpectations(t)
}

func TestSearchProductsUseCase_Execute_WithRepositoryError_ReturnsError(t *testing.T) {
	givenOrgID := uuid.New()
	givenQuery := "widget"
	givenError := errors.NewInternalServerError("database error")

	mockRepo := new(MockProductRepository)
	mockLogger := logger.New("test", "debug")

	mockRepo.On("Search", mock.Anything, givenQuery, mock.Anything, 1, 20).Return(nil, int64(0), givenError)

	useCase := NewSearchProductsUseCase(mockRepo, mockLogger)

	ctx := context.WithValue(context.Background(), "organization_id", givenOrgID)
	request := &SearchProductsRequest{
		Query:    givenQuery,
		Page:     1,
		PageSize: 20,
	}

	response, err := useCase.Execute(ctx, request)

	assert.Error(t, err)
	assert.Nil(t, response)
	mockRepo.AssertExpectations(t)
}
