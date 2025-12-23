package sales_order_test

import (
	"context"
	"testing"

	"github.com/giia/giia-core-engine/services/execution-service/internal/core/domain"
	"github.com/giia/giia-core-engine/services/execution-service/internal/core/providers"
	"github.com/giia/giia-core-engine/services/execution-service/internal/core/usecases/sales_order"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockInventoryTxnRepository struct {
	mock.Mock
}

func (m *MockInventoryTxnRepository) Create(ctx context.Context, txn *domain.InventoryTransaction) error {
	args := m.Called(ctx, txn)
	return args.Error(0)
}

func (m *MockInventoryTxnRepository) GetByID(ctx context.Context, id, organizationID uuid.UUID) (*domain.InventoryTransaction, error) {
	args := m.Called(ctx, id, organizationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.InventoryTransaction), args.Error(1)
}

func (m *MockInventoryTxnRepository) List(ctx context.Context, organizationID, productID uuid.UUID, filters map[string]interface{}, page, pageSize int) ([]*domain.InventoryTransaction, int64, error) {
	args := m.Called(ctx, organizationID, productID, filters, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*domain.InventoryTransaction), args.Get(1).(int64), args.Error(2)
}

func (m *MockInventoryTxnRepository) GetByReferenceID(ctx context.Context, referenceType string, referenceID, organizationID uuid.UUID) ([]*domain.InventoryTransaction, error) {
	args := m.Called(ctx, referenceType, referenceID, organizationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.InventoryTransaction), args.Error(1)
}

type MockInventoryBalanceRepository struct {
	mock.Mock
}

func (m *MockInventoryBalanceRepository) GetOrCreate(ctx context.Context, organizationID, productID, locationID uuid.UUID) (*domain.InventoryBalance, error) {
	args := m.Called(ctx, organizationID, productID, locationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.InventoryBalance), args.Error(1)
}

func (m *MockInventoryBalanceRepository) UpdateOnHand(ctx context.Context, organizationID, productID, locationID uuid.UUID, quantity float64) error {
	args := m.Called(ctx, organizationID, productID, locationID, quantity)
	return args.Error(0)
}

func (m *MockInventoryBalanceRepository) UpdateReserved(ctx context.Context, organizationID, productID, locationID uuid.UUID, quantity float64) error {
	args := m.Called(ctx, organizationID, productID, locationID, quantity)
	return args.Error(0)
}

func (m *MockInventoryBalanceRepository) GetByProduct(ctx context.Context, organizationID, productID uuid.UUID) ([]*domain.InventoryBalance, error) {
	args := m.Called(ctx, organizationID, productID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.InventoryBalance), args.Error(1)
}

func (m *MockInventoryBalanceRepository) GetByLocation(ctx context.Context, organizationID, locationID uuid.UUID) ([]*domain.InventoryBalance, error) {
	args := m.Called(ctx, organizationID, locationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.InventoryBalance), args.Error(1)
}

type MockDDMRPClient struct {
	mock.Mock
}

func (m *MockDDMRPClient) GetBufferStatus(ctx context.Context, organizationID, productID uuid.UUID) (*providers.BufferStatus, error) {
	args := m.Called(ctx, organizationID, productID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*providers.BufferStatus), args.Error(1)
}

func (m *MockDDMRPClient) UpdateNetFlowPosition(ctx context.Context, organizationID, productID uuid.UUID) error {
	args := m.Called(ctx, organizationID, productID)
	return args.Error(0)
}

func (m *MockDDMRPClient) GetProductsInRedZone(ctx context.Context, organizationID uuid.UUID) ([]*providers.BufferStatus, error) {
	args := m.Called(ctx, organizationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*providers.BufferStatus), args.Error(1)
}

func TestIssueDeliveryNoteUseCase_Execute_WithValidData_IssuesDeliveryNote(t *testing.T) {
	givenOrgID := uuid.New()
	givenSOID := uuid.New()
	givenProductID := uuid.New()
	givenLocationID := uuid.New()
	givenIssuedBy := uuid.New()
	givenDeliveryNoteNumber := "DN-001"

	givenSO := &domain.SalesOrder{
		ID:                 givenSOID,
		OrganizationID:     givenOrgID,
		SONumber:           "SO-001",
		Status:             domain.SOStatusConfirmed,
		DeliveryNoteIssued: false,
		LineItems: []domain.SOLineItem{
			{
				ID:        uuid.New(),
				ProductID: givenProductID,
				Quantity:  10,
				UnitPrice: 150,
				LineTotal: 1500,
			},
		},
	}

	mockSORepo := new(MockSORepository)
	mockInventoryTxnRepo := new(MockInventoryTxnRepository)
	mockInventoryBalanceRepo := new(MockInventoryBalanceRepository)
	mockDDMRPClient := new(MockDDMRPClient)
	mockPublisher := new(MockEventPublisher)

	mockSORepo.On("GetByID", mock.Anything, givenSOID, givenOrgID).Return(givenSO, nil)
	mockInventoryTxnRepo.On("Create", mock.Anything, mock.MatchedBy(func(txn *domain.InventoryTransaction) bool {
		return txn.ProductID == givenProductID && txn.Type == domain.TransactionIssue
	})).Return(nil)
	mockInventoryBalanceRepo.On("UpdateOnHand", mock.Anything, givenOrgID, givenProductID, givenLocationID, -10.0).Return(nil)
	mockDDMRPClient.On("UpdateNetFlowPosition", mock.Anything, givenOrgID, givenProductID).Return(nil)
	mockPublisher.On("PublishInventoryUpdated", mock.Anything, mock.Anything).Return(nil)
	mockSORepo.On("Update", mock.Anything, mock.MatchedBy(func(so *domain.SalesOrder) bool {
		return so.DeliveryNoteIssued && so.DeliveryNoteNumber == givenDeliveryNoteNumber
	})).Return(nil)
	mockPublisher.On("PublishDeliveryNoteIssued", mock.Anything, mock.Anything).Return(nil)

	useCase := sales_order.NewIssueDeliveryNoteUseCase(
		mockSORepo,
		mockInventoryTxnRepo,
		mockInventoryBalanceRepo,
		mockDDMRPClient,
		mockPublisher,
	)

	input := &sales_order.IssueDeliveryNoteInput{
		SOID:               givenSOID,
		OrganizationID:     givenOrgID,
		LocationID:         givenLocationID,
		DeliveryNoteNumber: givenDeliveryNoteNumber,
		IssuedBy:           givenIssuedBy,
	}

	so, err := useCase.Execute(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, so)
	assert.True(t, so.DeliveryNoteIssued)
	assert.Equal(t, givenDeliveryNoteNumber, so.DeliveryNoteNumber)
	mockSORepo.AssertExpectations(t)
	mockInventoryTxnRepo.AssertExpectations(t)
	mockInventoryBalanceRepo.AssertExpectations(t)
	mockDDMRPClient.AssertExpectations(t)
	mockPublisher.AssertExpectations(t)
}

func TestIssueDeliveryNoteUseCase_Execute_WithNilInput_ReturnsError(t *testing.T) {
	mockSORepo := new(MockSORepository)
	mockInventoryTxnRepo := new(MockInventoryTxnRepository)
	mockInventoryBalanceRepo := new(MockInventoryBalanceRepository)
	mockDDMRPClient := new(MockDDMRPClient)
	mockPublisher := new(MockEventPublisher)

	useCase := sales_order.NewIssueDeliveryNoteUseCase(
		mockSORepo,
		mockInventoryTxnRepo,
		mockInventoryBalanceRepo,
		mockDDMRPClient,
		mockPublisher,
	)

	so, err := useCase.Execute(context.Background(), nil)

	assert.Error(t, err)
	assert.Nil(t, so)
	assert.Contains(t, err.Error(), "input cannot be nil")
}

func TestIssueDeliveryNoteUseCase_Execute_WithNilSOID_ReturnsError(t *testing.T) {
	mockSORepo := new(MockSORepository)
	mockInventoryTxnRepo := new(MockInventoryTxnRepository)
	mockInventoryBalanceRepo := new(MockInventoryBalanceRepository)
	mockDDMRPClient := new(MockDDMRPClient)
	mockPublisher := new(MockEventPublisher)

	useCase := sales_order.NewIssueDeliveryNoteUseCase(
		mockSORepo,
		mockInventoryTxnRepo,
		mockInventoryBalanceRepo,
		mockDDMRPClient,
		mockPublisher,
	)

	input := &sales_order.IssueDeliveryNoteInput{
		SOID:               uuid.Nil,
		OrganizationID:     uuid.New(),
		LocationID:         uuid.New(),
		DeliveryNoteNumber: "DN-001",
		IssuedBy:           uuid.New(),
	}

	so, err := useCase.Execute(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, so)
	assert.Contains(t, err.Error(), "so_id is required")
}

func TestIssueDeliveryNoteUseCase_Execute_WithEmptyDeliveryNoteNumber_ReturnsError(t *testing.T) {
	mockSORepo := new(MockSORepository)
	mockInventoryTxnRepo := new(MockInventoryTxnRepository)
	mockInventoryBalanceRepo := new(MockInventoryBalanceRepository)
	mockDDMRPClient := new(MockDDMRPClient)
	mockPublisher := new(MockEventPublisher)

	useCase := sales_order.NewIssueDeliveryNoteUseCase(
		mockSORepo,
		mockInventoryTxnRepo,
		mockInventoryBalanceRepo,
		mockDDMRPClient,
		mockPublisher,
	)

	input := &sales_order.IssueDeliveryNoteInput{
		SOID:               uuid.New(),
		OrganizationID:     uuid.New(),
		LocationID:         uuid.New(),
		DeliveryNoteNumber: "",
		IssuedBy:           uuid.New(),
	}

	so, err := useCase.Execute(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, so)
	assert.Contains(t, err.Error(), "delivery_note_number is required")
}

func TestIssueDeliveryNoteUseCase_Execute_WhenAlreadyIssued_ReturnsError(t *testing.T) {
	givenOrgID := uuid.New()
	givenSOID := uuid.New()
	givenSO := &domain.SalesOrder{
		ID:                 givenSOID,
		OrganizationID:     givenOrgID,
		Status:             domain.SOStatusConfirmed,
		DeliveryNoteIssued: true,
		DeliveryNoteNumber: "DN-EXISTING",
	}

	mockSORepo := new(MockSORepository)
	mockInventoryTxnRepo := new(MockInventoryTxnRepository)
	mockInventoryBalanceRepo := new(MockInventoryBalanceRepository)
	mockDDMRPClient := new(MockDDMRPClient)
	mockPublisher := new(MockEventPublisher)

	mockSORepo.On("GetByID", mock.Anything, givenSOID, givenOrgID).Return(givenSO, nil)

	useCase := sales_order.NewIssueDeliveryNoteUseCase(
		mockSORepo,
		mockInventoryTxnRepo,
		mockInventoryBalanceRepo,
		mockDDMRPClient,
		mockPublisher,
	)

	input := &sales_order.IssueDeliveryNoteInput{
		SOID:               givenSOID,
		OrganizationID:     givenOrgID,
		LocationID:         uuid.New(),
		DeliveryNoteNumber: "DN-NEW",
		IssuedBy:           uuid.New(),
	}

	so, err := useCase.Execute(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, so)
	assert.Contains(t, err.Error(), "delivery note already issued")
	mockSORepo.AssertExpectations(t)
}

func TestIssueDeliveryNoteUseCase_Execute_WhenInventoryUpdateFails_ReturnsError(t *testing.T) {
	givenOrgID := uuid.New()
	givenSOID := uuid.New()
	givenProductID := uuid.New()
	givenLocationID := uuid.New()
	givenSO := &domain.SalesOrder{
		ID:                 givenSOID,
		OrganizationID:     givenOrgID,
		Status:             domain.SOStatusConfirmed,
		DeliveryNoteIssued: false,
		LineItems: []domain.SOLineItem{
			{ID: uuid.New(), ProductID: givenProductID, Quantity: 10},
		},
	}

	mockSORepo := new(MockSORepository)
	mockInventoryTxnRepo := new(MockInventoryTxnRepository)
	mockInventoryBalanceRepo := new(MockInventoryBalanceRepository)
	mockDDMRPClient := new(MockDDMRPClient)
	mockPublisher := new(MockEventPublisher)

	mockSORepo.On("GetByID", mock.Anything, givenSOID, givenOrgID).Return(givenSO, nil)
	mockInventoryTxnRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
	mockInventoryBalanceRepo.On("UpdateOnHand", mock.Anything, givenOrgID, givenProductID, givenLocationID, -10.0).Return(assert.AnError)

	useCase := sales_order.NewIssueDeliveryNoteUseCase(
		mockSORepo,
		mockInventoryTxnRepo,
		mockInventoryBalanceRepo,
		mockDDMRPClient,
		mockPublisher,
	)

	input := &sales_order.IssueDeliveryNoteInput{
		SOID:               givenSOID,
		OrganizationID:     givenOrgID,
		LocationID:         givenLocationID,
		DeliveryNoteNumber: "DN-001",
		IssuedBy:           uuid.New(),
	}

	so, err := useCase.Execute(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, so)
	mockSORepo.AssertExpectations(t)
	mockInventoryTxnRepo.AssertExpectations(t)
	mockInventoryBalanceRepo.AssertExpectations(t)
}

func TestIssueDeliveryNoteUseCase_Execute_WhenTransactionCreationFails_ReturnsError(t *testing.T) {
	givenOrgID := uuid.New()
	givenSOID := uuid.New()
	givenProductID := uuid.New()
	givenLocationID := uuid.New()
	givenSO := &domain.SalesOrder{
		ID:                 givenSOID,
		OrganizationID:     givenOrgID,
		Status:             domain.SOStatusConfirmed,
		DeliveryNoteIssued: false,
		LineItems: []domain.SOLineItem{
			{ID: uuid.New(), ProductID: givenProductID, Quantity: 10},
		},
	}

	mockSORepo := new(MockSORepository)
	mockInventoryTxnRepo := new(MockInventoryTxnRepository)
	mockInventoryBalanceRepo := new(MockInventoryBalanceRepository)
	mockDDMRPClient := new(MockDDMRPClient)
	mockPublisher := new(MockEventPublisher)

	mockSORepo.On("GetByID", mock.Anything, givenSOID, givenOrgID).Return(givenSO, nil)
	mockInventoryTxnRepo.On("Create", mock.Anything, mock.Anything).Return(assert.AnError)

	useCase := sales_order.NewIssueDeliveryNoteUseCase(
		mockSORepo,
		mockInventoryTxnRepo,
		mockInventoryBalanceRepo,
		mockDDMRPClient,
		mockPublisher,
	)

	input := &sales_order.IssueDeliveryNoteInput{
		SOID:               givenSOID,
		OrganizationID:     givenOrgID,
		LocationID:         givenLocationID,
		DeliveryNoteNumber: "DN-001",
		IssuedBy:           uuid.New(),
	}

	so, err := useCase.Execute(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, so)
	mockSORepo.AssertExpectations(t)
	mockInventoryTxnRepo.AssertExpectations(t)
}

func TestIssueDeliveryNoteUseCase_Execute_WhenSOUpdateFails_ReturnsError(t *testing.T) {
	givenOrgID := uuid.New()
	givenSOID := uuid.New()
	givenProductID := uuid.New()
	givenLocationID := uuid.New()
	givenSO := &domain.SalesOrder{
		ID:                 givenSOID,
		OrganizationID:     givenOrgID,
		Status:             domain.SOStatusConfirmed,
		DeliveryNoteIssued: false,
		LineItems: []domain.SOLineItem{
			{ID: uuid.New(), ProductID: givenProductID, Quantity: 10},
		},
	}

	mockSORepo := new(MockSORepository)
	mockInventoryTxnRepo := new(MockInventoryTxnRepository)
	mockInventoryBalanceRepo := new(MockInventoryBalanceRepository)
	mockDDMRPClient := new(MockDDMRPClient)
	mockPublisher := new(MockEventPublisher)

	mockSORepo.On("GetByID", mock.Anything, givenSOID, givenOrgID).Return(givenSO, nil)
	mockInventoryTxnRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
	mockInventoryBalanceRepo.On("UpdateOnHand", mock.Anything, givenOrgID, givenProductID, givenLocationID, -10.0).Return(nil)
	mockDDMRPClient.On("UpdateNetFlowPosition", mock.Anything, givenOrgID, givenProductID).Return(nil)
	mockPublisher.On("PublishInventoryUpdated", mock.Anything, mock.Anything).Return(nil)
	mockSORepo.On("Update", mock.Anything, mock.Anything).Return(assert.AnError)

	useCase := sales_order.NewIssueDeliveryNoteUseCase(
		mockSORepo,
		mockInventoryTxnRepo,
		mockInventoryBalanceRepo,
		mockDDMRPClient,
		mockPublisher,
	)

	input := &sales_order.IssueDeliveryNoteInput{
		SOID:               givenSOID,
		OrganizationID:     givenOrgID,
		LocationID:         givenLocationID,
		DeliveryNoteNumber: "DN-001",
		IssuedBy:           uuid.New(),
	}

	so, err := useCase.Execute(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, so)
	mockSORepo.AssertExpectations(t)
	mockInventoryTxnRepo.AssertExpectations(t)
	mockInventoryBalanceRepo.AssertExpectations(t)
}
