package alert_test

import (
	"context"
	"testing"
	"time"

	"github.com/melegattip/giia-core-engine/services/execution-service/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/execution-service/internal/core/usecases/alert"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockPORepository struct {
	mock.Mock
}

func (m *MockPORepository) Create(ctx context.Context, po *domain.PurchaseOrder) error {
	args := m.Called(ctx, po)
	return args.Error(0)
}

func (m *MockPORepository) GetByID(ctx context.Context, id, organizationID uuid.UUID) (*domain.PurchaseOrder, error) {
	args := m.Called(ctx, id, organizationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PurchaseOrder), args.Error(1)
}

func (m *MockPORepository) GetByPONumber(ctx context.Context, poNumber string, organizationID uuid.UUID) (*domain.PurchaseOrder, error) {
	args := m.Called(ctx, poNumber, organizationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PurchaseOrder), args.Error(1)
}

func (m *MockPORepository) Update(ctx context.Context, po *domain.PurchaseOrder) error {
	args := m.Called(ctx, po)
	return args.Error(0)
}

func (m *MockPORepository) Delete(ctx context.Context, id, organizationID uuid.UUID) error {
	args := m.Called(ctx, id, organizationID)
	return args.Error(0)
}

func (m *MockPORepository) List(ctx context.Context, organizationID uuid.UUID, filters map[string]interface{}, page, pageSize int) ([]*domain.PurchaseOrder, int64, error) {
	args := m.Called(ctx, organizationID, filters, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*domain.PurchaseOrder), args.Get(1).(int64), args.Error(2)
}

func (m *MockPORepository) GetDelayedOrders(ctx context.Context, organizationID uuid.UUID) ([]*domain.PurchaseOrder, error) {
	args := m.Called(ctx, organizationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.PurchaseOrder), args.Error(1)
}

type MockAlertRepository struct {
	mock.Mock
}

func (m *MockAlertRepository) Create(ctx context.Context, alert *domain.Alert) error {
	args := m.Called(ctx, alert)
	return args.Error(0)
}

func (m *MockAlertRepository) GetByID(ctx context.Context, id, organizationID uuid.UUID) (*domain.Alert, error) {
	args := m.Called(ctx, id, organizationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Alert), args.Error(1)
}

func (m *MockAlertRepository) Update(ctx context.Context, alert *domain.Alert) error {
	args := m.Called(ctx, alert)
	return args.Error(0)
}

func (m *MockAlertRepository) List(ctx context.Context, organizationID uuid.UUID, filters map[string]interface{}, page, pageSize int) ([]*domain.Alert, int64, error) {
	args := m.Called(ctx, organizationID, filters, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*domain.Alert), args.Get(1).(int64), args.Error(2)
}

func (m *MockAlertRepository) GetActiveAlerts(ctx context.Context, organizationID uuid.UUID) ([]*domain.Alert, error) {
	args := m.Called(ctx, organizationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Alert), args.Error(1)
}

func (m *MockAlertRepository) GetByResourceID(ctx context.Context, resourceType string, resourceID, organizationID uuid.UUID) ([]*domain.Alert, error) {
	args := m.Called(ctx, resourceType, resourceID, organizationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Alert), args.Error(1)
}

type MockEventPublisher struct {
	mock.Mock
}

func (m *MockEventPublisher) PublishPOCreated(ctx context.Context, po *domain.PurchaseOrder) error {
	args := m.Called(ctx, po)
	return args.Error(0)
}

func (m *MockEventPublisher) PublishPOUpdated(ctx context.Context, po *domain.PurchaseOrder) error {
	args := m.Called(ctx, po)
	return args.Error(0)
}

func (m *MockEventPublisher) PublishPOReceived(ctx context.Context, po *domain.PurchaseOrder) error {
	args := m.Called(ctx, po)
	return args.Error(0)
}

func (m *MockEventPublisher) PublishPOCancelled(ctx context.Context, po *domain.PurchaseOrder) error {
	args := m.Called(ctx, po)
	return args.Error(0)
}

func (m *MockEventPublisher) PublishSOCreated(ctx context.Context, so *domain.SalesOrder) error {
	args := m.Called(ctx, so)
	return args.Error(0)
}

func (m *MockEventPublisher) PublishSOUpdated(ctx context.Context, so *domain.SalesOrder) error {
	args := m.Called(ctx, so)
	return args.Error(0)
}

func (m *MockEventPublisher) PublishSOCancelled(ctx context.Context, so *domain.SalesOrder) error {
	args := m.Called(ctx, so)
	return args.Error(0)
}

func (m *MockEventPublisher) PublishDeliveryNoteIssued(ctx context.Context, so *domain.SalesOrder) error {
	args := m.Called(ctx, so)
	return args.Error(0)
}

func (m *MockEventPublisher) PublishInventoryUpdated(ctx context.Context, txn *domain.InventoryTransaction) error {
	args := m.Called(ctx, txn)
	return args.Error(0)
}

func (m *MockEventPublisher) PublishAlertCreated(ctx context.Context, alert *domain.Alert) error {
	args := m.Called(ctx, alert)
	return args.Error(0)
}

func TestGeneratePODelayAlertsUseCase_Execute_WithDelayedPOs_CreatesAlerts(t *testing.T) {
	givenOrgID := uuid.New()
	givenPOID := uuid.New()
	givenDelayedPO := &domain.PurchaseOrder{
		ID:                  givenPOID,
		OrganizationID:      givenOrgID,
		PONumber:            "PO-001",
		Status:              domain.POStatusConfirmed,
		ExpectedArrivalDate: time.Now().AddDate(0, 0, -5),
		IsDelayed:           false,
	}

	mockPORepo := new(MockPORepository)
	mockAlertRepo := new(MockAlertRepository)
	mockPublisher := new(MockEventPublisher)

	mockPORepo.On("GetDelayedOrders", mock.Anything, givenOrgID).Return([]*domain.PurchaseOrder{givenDelayedPO}, nil)
	mockAlertRepo.On("GetByResourceID", mock.Anything, "purchase_order", givenPOID, givenOrgID).Return([]*domain.Alert{}, nil)
	mockAlertRepo.On("Create", mock.Anything, mock.MatchedBy(func(alert *domain.Alert) bool {
		return alert.AlertType == domain.AlertTypePODelayed && alert.ResourceID == givenPOID
	})).Return(nil)
	mockPublisher.On("PublishAlertCreated", mock.Anything, mock.Anything).Return(nil)

	useCase := alert.NewGeneratePODelayAlertsUseCase(mockPORepo, mockAlertRepo, mockPublisher)

	err := useCase.Execute(context.Background(), givenOrgID)

	assert.NoError(t, err)
	mockPORepo.AssertExpectations(t)
	mockAlertRepo.AssertExpectations(t)
	mockPublisher.AssertExpectations(t)
}

func TestGeneratePODelayAlertsUseCase_Execute_WithNoDelayedPOs_CreatesNoAlerts(t *testing.T) {
	givenOrgID := uuid.New()

	mockPORepo := new(MockPORepository)
	mockAlertRepo := new(MockAlertRepository)
	mockPublisher := new(MockEventPublisher)

	mockPORepo.On("GetDelayedOrders", mock.Anything, givenOrgID).Return([]*domain.PurchaseOrder{}, nil)

	useCase := alert.NewGeneratePODelayAlertsUseCase(mockPORepo, mockAlertRepo, mockPublisher)

	err := useCase.Execute(context.Background(), givenOrgID)

	assert.NoError(t, err)
	mockPORepo.AssertExpectations(t)
	mockAlertRepo.AssertNotCalled(t, "Create")
	mockPublisher.AssertNotCalled(t, "PublishAlertCreated")
}

func TestGeneratePODelayAlertsUseCase_Execute_WithExistingActiveAlert_DoesNotCreateDuplicate(t *testing.T) {
	givenOrgID := uuid.New()
	givenPOID := uuid.New()
	givenDelayedPO := &domain.PurchaseOrder{
		ID:                  givenPOID,
		OrganizationID:      givenOrgID,
		PONumber:            "PO-001",
		Status:              domain.POStatusConfirmed,
		ExpectedArrivalDate: time.Now().AddDate(0, 0, -5),
		IsDelayed:           false,
	}
	givenExistingAlert := &domain.Alert{
		ID:             uuid.New(),
		OrganizationID: givenOrgID,
		AlertType:      domain.AlertTypePODelayed,
		ResourceID:     givenPOID,
	}

	mockPORepo := new(MockPORepository)
	mockAlertRepo := new(MockAlertRepository)
	mockPublisher := new(MockEventPublisher)

	mockPORepo.On("GetDelayedOrders", mock.Anything, givenOrgID).Return([]*domain.PurchaseOrder{givenDelayedPO}, nil)
	mockAlertRepo.On("GetByResourceID", mock.Anything, "purchase_order", givenPOID, givenOrgID).Return([]*domain.Alert{givenExistingAlert}, nil)

	useCase := alert.NewGeneratePODelayAlertsUseCase(mockPORepo, mockAlertRepo, mockPublisher)

	err := useCase.Execute(context.Background(), givenOrgID)

	assert.NoError(t, err)
	mockPORepo.AssertExpectations(t)
	mockAlertRepo.AssertExpectations(t)
	mockAlertRepo.AssertNotCalled(t, "Create")
	mockPublisher.AssertNotCalled(t, "PublishAlertCreated")
}

func TestGeneratePODelayAlertsUseCase_Execute_WithNilOrganizationID_ReturnsError(t *testing.T) {
	mockPORepo := new(MockPORepository)
	mockAlertRepo := new(MockAlertRepository)
	mockPublisher := new(MockEventPublisher)

	useCase := alert.NewGeneratePODelayAlertsUseCase(mockPORepo, mockAlertRepo, mockPublisher)

	err := useCase.Execute(context.Background(), uuid.Nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "organization_id is required")
}