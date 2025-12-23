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

func TestDeleteProductUseCase_Execute_WithValidID_DeletesProduct(t *testing.T) {
	givenProductID := uuid.New()
	givenProduct := &domain.Product{
		ID:            givenProductID,
		SKU:           "WIDGET-001",
		Name:          "Premium Widget",
		UnitOfMeasure: "EA",
		Status:        domain.ProductStatusActive,
	}

	mockRepo := new(MockProductRepository)
	mockPublisher := new(MockEventPublisher)
	mockLogger := logger.New("test", "debug")

	mockRepo.On("GetByID", mock.Anything, givenProductID).Return(givenProduct, nil)
	mockRepo.On("Delete", mock.Anything, givenProductID).Return(nil)
	mockPublisher.On("PublishProductDeleted", mock.Anything, mock.MatchedBy(func(p *domain.Product) bool {
		return p.ID == givenProductID
	})).Return(nil)

	useCase := NewDeleteProductUseCase(mockRepo, mockPublisher, mockLogger)

	err := useCase.Execute(context.Background(), givenProductID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockPublisher.AssertExpectations(t)
}

func TestDeleteProductUseCase_Execute_WithNilID_ReturnsError(t *testing.T) {
	mockRepo := new(MockProductRepository)
	mockPublisher := new(MockEventPublisher)
	mockLogger := logger.New("test", "debug")

	useCase := NewDeleteProductUseCase(mockRepo, mockPublisher, mockLogger)

	err := useCase.Execute(context.Background(), uuid.Nil)

	assert.Error(t, err)
	assert.True(t, errors.IsBadRequest(err))
	mockRepo.AssertNotCalled(t, "GetByID")
	mockRepo.AssertNotCalled(t, "Delete")
	mockPublisher.AssertNotCalled(t, "PublishProductDeleted")
}

func TestDeleteProductUseCase_Execute_WithNonExistentProduct_ReturnsError(t *testing.T) {
	givenProductID := uuid.New()
	givenNotFoundError := errors.NewNotFound("product not found")

	mockRepo := new(MockProductRepository)
	mockPublisher := new(MockEventPublisher)
	mockLogger := logger.New("test", "debug")

	mockRepo.On("GetByID", mock.Anything, givenProductID).Return(nil, givenNotFoundError)

	useCase := NewDeleteProductUseCase(mockRepo, mockPublisher, mockLogger)

	err := useCase.Execute(context.Background(), givenProductID)

	assert.Error(t, err)
	assert.True(t, errors.IsNotFound(err))
	mockRepo.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "Delete")
	mockPublisher.AssertNotCalled(t, "PublishProductDeleted")
}

func TestDeleteProductUseCase_Execute_WithRepositoryError_ReturnsError(t *testing.T) {
	givenProductID := uuid.New()
	givenProduct := &domain.Product{
		ID:            givenProductID,
		SKU:           "WIDGET-001",
		Name:          "Premium Widget",
		UnitOfMeasure: "EA",
		Status:        domain.ProductStatusActive,
	}
	givenDatabaseError := errors.NewInternalServerError("database error")

	mockRepo := new(MockProductRepository)
	mockPublisher := new(MockEventPublisher)
	mockLogger := logger.New("test", "debug")

	mockRepo.On("GetByID", mock.Anything, givenProductID).Return(givenProduct, nil)
	mockRepo.On("Delete", mock.Anything, givenProductID).Return(givenDatabaseError)

	useCase := NewDeleteProductUseCase(mockRepo, mockPublisher, mockLogger)

	err := useCase.Execute(context.Background(), givenProductID)

	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
	mockPublisher.AssertNotCalled(t, "PublishProductDeleted")
}

func TestDeleteProductUseCase_Execute_WithEventPublishError_StillSucceeds(t *testing.T) {
	givenProductID := uuid.New()
	givenProduct := &domain.Product{
		ID:            givenProductID,
		SKU:           "WIDGET-001",
		Name:          "Premium Widget",
		UnitOfMeasure: "EA",
		Status:        domain.ProductStatusActive,
	}
	givenEventError := errors.NewInternalServerError("event publish failed")

	mockRepo := new(MockProductRepository)
	mockPublisher := new(MockEventPublisher)
	mockLogger := logger.New("test", "debug")

	mockRepo.On("GetByID", mock.Anything, givenProductID).Return(givenProduct, nil)
	mockRepo.On("Delete", mock.Anything, givenProductID).Return(nil)
	mockPublisher.On("PublishProductDeleted", mock.Anything, mock.Anything).Return(givenEventError)

	useCase := NewDeleteProductUseCase(mockRepo, mockPublisher, mockLogger)

	err := useCase.Execute(context.Background(), givenProductID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockPublisher.AssertExpectations(t)
}
