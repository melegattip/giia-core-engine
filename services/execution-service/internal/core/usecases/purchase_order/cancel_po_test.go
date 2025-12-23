package purchase_order_test

import (
	"context"
	"testing"

	"github.com/giia/giia-core-engine/services/execution-service/internal/core/domain"
	"github.com/giia/giia-core-engine/services/execution-service/internal/core/usecases/purchase_order"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCancelPOUseCase_Execute_WithDraftStatus_CancelsPurchaseOrder(t *testing.T) {
	givenOrgID := uuid.New()
	givenPOID := uuid.New()
	givenPO := &domain.PurchaseOrder{
		ID:             givenPOID,
		OrganizationID: givenOrgID,
		PONumber:       "PO-001",
		Status:         domain.POStatusDraft,
	}

	mockPORepo := new(MockPORepository)
	mockPublisher := new(MockEventPublisher)

	mockPORepo.On("GetByID", mock.Anything, givenPOID, givenOrgID).Return(givenPO, nil)
	mockPORepo.On("Update", mock.Anything, mock.MatchedBy(func(po *domain.PurchaseOrder) bool {
		return po.Status == domain.POStatusCancelled
	})).Return(nil)
	mockPublisher.On("PublishPOCancelled", mock.Anything, mock.Anything).Return(nil)

	useCase := purchase_order.NewCancelPOUseCase(mockPORepo, mockPublisher)

	input := &purchase_order.CancelPOInput{
		POID:           givenPOID,
		OrganizationID: givenOrgID,
	}

	po, err := useCase.Execute(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, po)
	assert.Equal(t, domain.POStatusCancelled, po.Status)
	mockPORepo.AssertExpectations(t)
	mockPublisher.AssertExpectations(t)
}

func TestCancelPOUseCase_Execute_WithNilInput_ReturnsError(t *testing.T) {
	mockPORepo := new(MockPORepository)
	mockPublisher := new(MockEventPublisher)

	useCase := purchase_order.NewCancelPOUseCase(mockPORepo, mockPublisher)

	po, err := useCase.Execute(context.Background(), nil)

	assert.Error(t, err)
	assert.Nil(t, po)
	assert.Contains(t, err.Error(), "input cannot be nil")
}

func TestCancelPOUseCase_Execute_WithReceivedStatus_ReturnsError(t *testing.T) {
	givenOrgID := uuid.New()
	givenPOID := uuid.New()
	givenPO := &domain.PurchaseOrder{
		ID:             givenPOID,
		OrganizationID: givenOrgID,
		Status:         domain.POStatusReceived,
	}

	mockPORepo := new(MockPORepository)
	mockPublisher := new(MockEventPublisher)

	mockPORepo.On("GetByID", mock.Anything, givenPOID, givenOrgID).Return(givenPO, nil)

	useCase := purchase_order.NewCancelPOUseCase(mockPORepo, mockPublisher)

	input := &purchase_order.CancelPOInput{
		POID:           givenPOID,
		OrganizationID: givenOrgID,
	}

	po, err := useCase.Execute(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, po)
	assert.Contains(t, err.Error(), "cannot cancel received")
	mockPORepo.AssertExpectations(t)
}

func TestCancelPOUseCase_Execute_WhenPONotFound_ReturnsError(t *testing.T) {
	givenOrgID := uuid.New()
	givenPOID := uuid.New()

	mockPORepo := new(MockPORepository)
	mockPublisher := new(MockEventPublisher)

	mockPORepo.On("GetByID", mock.Anything, givenPOID, givenOrgID).Return((*domain.PurchaseOrder)(nil), assert.AnError)

	useCase := purchase_order.NewCancelPOUseCase(mockPORepo, mockPublisher)

	input := &purchase_order.CancelPOInput{
		POID:           givenPOID,
		OrganizationID: givenOrgID,
	}

	po, err := useCase.Execute(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, po)
	mockPORepo.AssertExpectations(t)
}

func TestCancelPOUseCase_Execute_WhenRepositoryUpdateFails_ReturnsError(t *testing.T) {
	givenOrgID := uuid.New()
	givenPOID := uuid.New()
	givenPO := &domain.PurchaseOrder{
		ID:             givenPOID,
		OrganizationID: givenOrgID,
		PONumber:       "PO-003",
		Status:         domain.POStatusDraft,
	}

	mockPORepo := new(MockPORepository)
	mockPublisher := new(MockEventPublisher)

	mockPORepo.On("GetByID", mock.Anything, givenPOID, givenOrgID).Return(givenPO, nil)
	mockPORepo.On("Update", mock.Anything, mock.Anything).Return(assert.AnError)

	useCase := purchase_order.NewCancelPOUseCase(mockPORepo, mockPublisher)

	input := &purchase_order.CancelPOInput{
		POID:           givenPOID,
		OrganizationID: givenOrgID,
	}

	po, err := useCase.Execute(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, po)
	mockPORepo.AssertExpectations(t)
}

func TestCancelPOUseCase_Execute_WithNilPOID_ReturnsError(t *testing.T) {
	mockPORepo := new(MockPORepository)
	mockPublisher := new(MockEventPublisher)

	useCase := purchase_order.NewCancelPOUseCase(mockPORepo, mockPublisher)

	input := &purchase_order.CancelPOInput{
		POID:           uuid.Nil,
		OrganizationID: uuid.New(),
	}

	po, err := useCase.Execute(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, po)
	assert.Contains(t, err.Error(), "po_id is required")
}

func TestCancelPOUseCase_Execute_WithNilOrganizationID_ReturnsError(t *testing.T) {
	mockPORepo := new(MockPORepository)
	mockPublisher := new(MockEventPublisher)

	useCase := purchase_order.NewCancelPOUseCase(mockPORepo, mockPublisher)

	input := &purchase_order.CancelPOInput{
		POID:           uuid.New(),
		OrganizationID: uuid.Nil,
	}

	po, err := useCase.Execute(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, po)
	assert.Contains(t, err.Error(), "organization_id is required")
}