package product

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

func TestGetProductUseCase_Execute_WithValidID_ReturnsProduct(t *testing.T) {
	givenProductID := uuid.New()
	givenProduct := &domain.Product{
		ID:            givenProductID,
		SKU:           "WIDGET-001",
		Name:          "Premium Widget",
		UnitOfMeasure: "EA",
		Status:        domain.ProductStatusActive,
	}

	mockRepo := new(MockProductRepository)
	mockLogger := logger.New("test", "debug")

	mockRepo.On("GetByID", mock.Anything, givenProductID).Return(givenProduct, nil)

	useCase := NewGetProductUseCase(mockRepo, mockLogger)

	product, err := useCase.Execute(context.Background(), givenProductID, false)

	assert.NoError(t, err)
	assert.NotNil(t, product)
	assert.Equal(t, givenProductID, product.ID)
	assert.Equal(t, "WIDGET-001", product.SKU)
	mockRepo.AssertExpectations(t)
}

func TestGetProductUseCase_Execute_WithNilID_ReturnsError(t *testing.T) {
	mockRepo := new(MockProductRepository)
	mockLogger := logger.New("test", "debug")

	useCase := NewGetProductUseCase(mockRepo, mockLogger)

	product, err := useCase.Execute(context.Background(), uuid.Nil, false)

	assert.Error(t, err)
	assert.Nil(t, product)
	assert.True(t, errors.IsBadRequest(err))
	mockRepo.AssertNotCalled(t, "GetByID")
}

func TestGetProductUseCase_Execute_WithNonExistentProduct_ReturnsError(t *testing.T) {
	givenProductID := uuid.New()
	givenNotFoundError := errors.NewNotFound("product not found")

	mockRepo := new(MockProductRepository)
	mockLogger := logger.New("test", "debug")

	mockRepo.On("GetByID", mock.Anything, givenProductID).Return(nil, givenNotFoundError)

	useCase := NewGetProductUseCase(mockRepo, mockLogger)

	product, err := useCase.Execute(context.Background(), givenProductID, false)

	assert.Error(t, err)
	assert.Nil(t, product)
	assert.True(t, errors.IsNotFound(err))
	mockRepo.AssertExpectations(t)
}

func TestGetProductUseCase_Execute_WithIncludeSuppliers_CallsCorrectMethod(t *testing.T) {
	givenProductID := uuid.New()
	givenProduct := &domain.Product{
		ID:            givenProductID,
		SKU:           "WIDGET-001",
		Name:          "Premium Widget",
		UnitOfMeasure: "EA",
		Status:        domain.ProductStatusActive,
	}

	mockRepo := new(MockProductRepository)
	mockLogger := logger.New("test", "debug")

	mockRepo.On("GetByIDWithSuppliers", mock.Anything, givenProductID).Return(givenProduct, nil)

	useCase := NewGetProductUseCase(mockRepo, mockLogger)

	product, err := useCase.Execute(context.Background(), givenProductID, true)

	assert.NoError(t, err)
	assert.NotNil(t, product)
	mockRepo.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "GetByID")
}
