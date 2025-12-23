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

func TestUpdateProductUseCase_Execute_WithValidData_ReturnsUpdatedProduct(t *testing.T) {
	givenProductID := uuid.New()
	givenOrgID := uuid.New()
	givenExistingProduct := &domain.Product{
		ID:             givenProductID,
		SKU:            "WIDGET-001",
		Name:           "Old Name",
		Description:    "Old Description",
		UnitOfMeasure:  "EA",
		Status:         domain.ProductStatusActive,
		OrganizationID: givenOrgID,
	}

	mockRepo := new(MockProductRepository)
	mockPublisher := new(MockEventPublisher)
	mockLogger := logger.New("test", "debug")

	mockRepo.On("GetByID", mock.Anything, givenProductID).Return(givenExistingProduct, nil)
	mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(p *domain.Product) bool {
		return p.ID == givenProductID && p.Name == "New Name"
	})).Return(nil)
	mockPublisher.On("PublishProductUpdated", mock.Anything, mock.Anything).Return(nil)

	useCase := NewUpdateProductUseCase(mockRepo, mockPublisher, mockLogger)

	ctx := context.WithValue(context.Background(), "organization_id", givenOrgID)
	request := &UpdateProductRequest{
		ID:   givenProductID,
		Name: "New Name",
	}

	product, err := useCase.Execute(ctx, request)

	assert.NoError(t, err)
	assert.NotNil(t, product)
	assert.Equal(t, "New Name", product.Name)
	assert.Equal(t, "WIDGET-001", product.SKU)
	mockRepo.AssertExpectations(t)
	mockPublisher.AssertExpectations(t)
}

func TestUpdateProductUseCase_Execute_WithNilRequest_ReturnsError(t *testing.T) {
	mockRepo := new(MockProductRepository)
	mockPublisher := new(MockEventPublisher)
	mockLogger := logger.New("test", "debug")

	useCase := NewUpdateProductUseCase(mockRepo, mockPublisher, mockLogger)

	product, err := useCase.Execute(context.Background(), nil)

	assert.Error(t, err)
	assert.Nil(t, product)
	assert.True(t, errors.IsBadRequest(err))
	mockRepo.AssertNotCalled(t, "GetByID")
}

func TestUpdateProductUseCase_Execute_WithNilProductID_ReturnsError(t *testing.T) {
	mockRepo := new(MockProductRepository)
	mockPublisher := new(MockEventPublisher)
	mockLogger := logger.New("test", "debug")

	useCase := NewUpdateProductUseCase(mockRepo, mockPublisher, mockLogger)

	request := &UpdateProductRequest{
		ID:   uuid.Nil,
		Name: "New Name",
	}

	product, err := useCase.Execute(context.Background(), request)

	assert.Error(t, err)
	assert.Nil(t, product)
	assert.True(t, errors.IsBadRequest(err))
	mockRepo.AssertNotCalled(t, "GetByID")
}

func TestUpdateProductUseCase_Execute_WithNonExistentProduct_ReturnsError(t *testing.T) {
	givenProductID := uuid.New()
	givenNotFoundError := errors.NewNotFound("product not found")

	mockRepo := new(MockProductRepository)
	mockPublisher := new(MockEventPublisher)
	mockLogger := logger.New("test", "debug")

	mockRepo.On("GetByID", mock.Anything, givenProductID).Return(nil, givenNotFoundError)

	useCase := NewUpdateProductUseCase(mockRepo, mockPublisher, mockLogger)

	request := &UpdateProductRequest{
		ID:   givenProductID,
		Name: "New Name",
	}

	product, err := useCase.Execute(context.Background(), request)

	assert.Error(t, err)
	assert.Nil(t, product)
	assert.True(t, errors.IsNotFound(err))
	mockRepo.AssertExpectations(t)
	mockPublisher.AssertNotCalled(t, "PublishProductUpdated")
}

func TestUpdateProductUseCase_Execute_WithInvalidData_ReturnsValidationError(t *testing.T) {
	givenProductID := uuid.New()
	givenOrgID := uuid.New()
	givenExistingProduct := &domain.Product{
		ID:             givenProductID,
		SKU:            "WIDGET-001",
		Name:           "Old Name",
		UnitOfMeasure:  "EA",
		Status:         domain.ProductStatusActive,
		OrganizationID: givenOrgID,
	}

	mockRepo := new(MockProductRepository)
	mockPublisher := new(MockEventPublisher)
	mockLogger := logger.New("test", "debug")

	mockRepo.On("GetByID", mock.Anything, givenProductID).Return(givenExistingProduct, nil)

	useCase := NewUpdateProductUseCase(mockRepo, mockPublisher, mockLogger)

	longName := "This is a very long product name that exceeds the maximum allowed length of 255 characters for product names in the system. This validation test ensures that the system properly rejects product updates when the name field contains more than 255 characters which would violate the database schema constraints and business rules."
	request := &UpdateProductRequest{
		ID:   givenProductID,
		Name: longName,
	}

	product, err := useCase.Execute(context.Background(), request)

	assert.Error(t, err)
	assert.Nil(t, product)
	assert.True(t, errors.IsBadRequest(err))
	mockRepo.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "Update")
	mockPublisher.AssertNotCalled(t, "PublishProductUpdated")
}
