package domain_test

import (
	"testing"
	"time"

	"github.com/giia/giia-core-engine/services/execution-service/internal/core/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewPurchaseOrder_WithValidData_ReturnsPurchaseOrder(t *testing.T) {
	givenOrgID := uuid.New()
	givenSupplierID := uuid.New()
	givenCreatedBy := uuid.New()
	givenPONumber := "PO-001"
	givenOrderDate := time.Now()
	givenExpectedArrivalDate := time.Now().AddDate(0, 0, 7)
	givenLineItems := []domain.POLineItem{
		{
			ID:        uuid.New(),
			ProductID: uuid.New(),
			Quantity:  10,
			UnitCost:  100,
			LineTotal: 1000,
		},
	}

	po, err := domain.NewPurchaseOrder(
		givenOrgID,
		givenSupplierID,
		givenCreatedBy,
		givenPONumber,
		givenOrderDate,
		givenExpectedArrivalDate,
		givenLineItems,
	)

	assert.NoError(t, err)
	assert.NotNil(t, po)
	assert.Equal(t, givenOrgID, po.OrganizationID)
	assert.Equal(t, givenSupplierID, po.SupplierID)
	assert.Equal(t, givenCreatedBy, po.CreatedBy)
	assert.Equal(t, givenPONumber, po.PONumber)
	assert.Equal(t, domain.POStatusDraft, po.Status)
	assert.Equal(t, float64(1000), po.TotalAmount)
	assert.Equal(t, false, po.IsDelayed)
	assert.Equal(t, 0, po.DelayDays)
	assert.Len(t, po.LineItems, 1)
}

func TestNewPurchaseOrder_WithNilOrganizationID_ReturnsError(t *testing.T) {
	givenSupplierID := uuid.New()
	givenCreatedBy := uuid.New()
	givenPONumber := "PO-001"
	givenOrderDate := time.Now()
	givenExpectedArrivalDate := time.Now().AddDate(0, 0, 7)
	givenLineItems := []domain.POLineItem{
		{ProductID: uuid.New(), Quantity: 10, UnitCost: 100, LineTotal: 1000},
	}

	po, err := domain.NewPurchaseOrder(
		uuid.Nil,
		givenSupplierID,
		givenCreatedBy,
		givenPONumber,
		givenOrderDate,
		givenExpectedArrivalDate,
		givenLineItems,
	)

	assert.Error(t, err)
	assert.Nil(t, po)
	assert.Contains(t, err.Error(), "organization_id is required")
}

func TestNewPurchaseOrder_WithNilSupplierID_ReturnsError(t *testing.T) {
	givenOrgID := uuid.New()
	givenCreatedBy := uuid.New()
	givenPONumber := "PO-001"
	givenOrderDate := time.Now()
	givenExpectedArrivalDate := time.Now().AddDate(0, 0, 7)
	givenLineItems := []domain.POLineItem{
		{ProductID: uuid.New(), Quantity: 10, UnitCost: 100, LineTotal: 1000},
	}

	po, err := domain.NewPurchaseOrder(
		givenOrgID,
		uuid.Nil,
		givenCreatedBy,
		givenPONumber,
		givenOrderDate,
		givenExpectedArrivalDate,
		givenLineItems,
	)

	assert.Error(t, err)
	assert.Nil(t, po)
	assert.Contains(t, err.Error(), "supplier_id is required")
}

func TestNewPurchaseOrder_WithEmptyPONumber_ReturnsError(t *testing.T) {
	givenOrgID := uuid.New()
	givenSupplierID := uuid.New()
	givenCreatedBy := uuid.New()
	givenOrderDate := time.Now()
	givenExpectedArrivalDate := time.Now().AddDate(0, 0, 7)
	givenLineItems := []domain.POLineItem{
		{ProductID: uuid.New(), Quantity: 10, UnitCost: 100, LineTotal: 1000},
	}

	po, err := domain.NewPurchaseOrder(
		givenOrgID,
		givenSupplierID,
		givenCreatedBy,
		"",
		givenOrderDate,
		givenExpectedArrivalDate,
		givenLineItems,
	)

	assert.Error(t, err)
	assert.Nil(t, po)
	assert.Contains(t, err.Error(), "po_number is required")
}

func TestNewPurchaseOrder_WithEmptyLineItems_ReturnsError(t *testing.T) {
	givenOrgID := uuid.New()
	givenSupplierID := uuid.New()
	givenCreatedBy := uuid.New()
	givenPONumber := "PO-001"
	givenOrderDate := time.Now()
	givenExpectedArrivalDate := time.Now().AddDate(0, 0, 7)

	po, err := domain.NewPurchaseOrder(
		givenOrgID,
		givenSupplierID,
		givenCreatedBy,
		givenPONumber,
		givenOrderDate,
		givenExpectedArrivalDate,
		[]domain.POLineItem{},
	)

	assert.Error(t, err)
	assert.Nil(t, po)
	assert.Contains(t, err.Error(), "at least one line item is required")
}

func TestNewPurchaseOrder_WithNilCreatedBy_ReturnsError(t *testing.T) {
	givenOrgID := uuid.New()
	givenSupplierID := uuid.New()
	givenPONumber := "PO-001"
	givenOrderDate := time.Now()
	givenExpectedArrivalDate := time.Now().AddDate(0, 0, 7)
	givenLineItems := []domain.POLineItem{
		{ProductID: uuid.New(), Quantity: 10, UnitCost: 100, LineTotal: 1000},
	}

	po, err := domain.NewPurchaseOrder(
		givenOrgID,
		givenSupplierID,
		uuid.Nil,
		givenPONumber,
		givenOrderDate,
		givenExpectedArrivalDate,
		givenLineItems,
	)

	assert.Error(t, err)
	assert.Nil(t, po)
	assert.Contains(t, err.Error(), "created_by is required")
}

func TestNewPurchaseOrder_WithInvalidLineItem_ReturnsError(t *testing.T) {
	givenOrgID := uuid.New()
	givenSupplierID := uuid.New()
	givenCreatedBy := uuid.New()
	givenPONumber := "PO-001"
	givenOrderDate := time.Now()
	givenExpectedArrivalDate := time.Now().AddDate(0, 0, 7)
	givenLineItems := []domain.POLineItem{
		{ProductID: uuid.Nil, Quantity: 10, UnitCost: 100, LineTotal: 1000},
	}

	po, err := domain.NewPurchaseOrder(
		givenOrgID,
		givenSupplierID,
		givenCreatedBy,
		givenPONumber,
		givenOrderDate,
		givenExpectedArrivalDate,
		givenLineItems,
	)

	assert.Error(t, err)
	assert.Nil(t, po)
	assert.Contains(t, err.Error(), "product_id is required")
}

func TestNewPurchaseOrder_WithZeroQuantity_ReturnsError(t *testing.T) {
	givenOrgID := uuid.New()
	givenSupplierID := uuid.New()
	givenCreatedBy := uuid.New()
	givenPONumber := "PO-001"
	givenOrderDate := time.Now()
	givenExpectedArrivalDate := time.Now().AddDate(0, 0, 7)
	givenLineItems := []domain.POLineItem{
		{ProductID: uuid.New(), Quantity: 0, UnitCost: 100, LineTotal: 0},
	}

	po, err := domain.NewPurchaseOrder(
		givenOrgID,
		givenSupplierID,
		givenCreatedBy,
		givenPONumber,
		givenOrderDate,
		givenExpectedArrivalDate,
		givenLineItems,
	)

	assert.Error(t, err)
	assert.Nil(t, po)
	assert.Contains(t, err.Error(), "quantity must be greater than zero")
}

func TestPurchaseOrder_CheckDelay_WhenActualArrivalIsLater_SetsIsDelayedTrue(t *testing.T) {
	givenExpectedDate := time.Now().AddDate(0, 0, -5)
	givenActualDate := time.Now()
	po := &domain.PurchaseOrder{
		ExpectedArrivalDate: givenExpectedDate,
		ActualArrivalDate:   &givenActualDate,
		Status:              domain.POStatusReceived,
	}

	po.CheckDelay()

	assert.True(t, po.IsDelayed)
	assert.True(t, po.DelayDays > 0)
}

func TestPurchaseOrder_CheckDelay_WhenActualArrivalIsEarlier_SetsIsDelayedFalse(t *testing.T) {
	givenExpectedDate := time.Now()
	givenActualDate := time.Now().AddDate(0, 0, -2)
	po := &domain.PurchaseOrder{
		ExpectedArrivalDate: givenExpectedDate,
		ActualArrivalDate:   &givenActualDate,
		Status:              domain.POStatusReceived,
	}

	po.CheckDelay()

	assert.False(t, po.IsDelayed)
	assert.True(t, po.DelayDays < 0)
}

func TestPurchaseOrder_CheckDelay_WhenNoActualDateAndPastDue_SetsIsDelayedTrue(t *testing.T) {
	givenExpectedDate := time.Now().AddDate(0, 0, -3)
	po := &domain.PurchaseOrder{
		ExpectedArrivalDate: givenExpectedDate,
		ActualArrivalDate:   nil,
		Status:              domain.POStatusConfirmed,
	}

	po.CheckDelay()

	assert.True(t, po.IsDelayed)
	assert.True(t, po.DelayDays > 0)
}

func TestPurchaseOrder_CheckDelay_WhenNoActualDateAndNotPastDue_SetsIsDelayedFalse(t *testing.T) {
	givenExpectedDate := time.Now().AddDate(0, 0, 5)
	po := &domain.PurchaseOrder{
		ExpectedArrivalDate: givenExpectedDate,
		ActualArrivalDate:   nil,
		Status:              domain.POStatusConfirmed,
	}

	po.CheckDelay()

	assert.False(t, po.IsDelayed)
}

func TestPurchaseOrder_Confirm_WithDraftStatus_UpdatesStatusToConfirmed(t *testing.T) {
	po := &domain.PurchaseOrder{
		Status: domain.POStatusDraft,
	}

	err := po.Confirm()

	assert.NoError(t, err)
	assert.Equal(t, domain.POStatusConfirmed, po.Status)
}

func TestPurchaseOrder_Confirm_WithReceivedStatus_ReturnsError(t *testing.T) {
	po := &domain.PurchaseOrder{
		Status: domain.POStatusReceived,
	}

	err := po.Confirm()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "can only confirm draft or pending")
}

func TestPurchaseOrder_Cancel_WithDraftStatus_UpdatesStatusToCancelled(t *testing.T) {
	po := &domain.PurchaseOrder{
		Status: domain.POStatusDraft,
	}

	err := po.Cancel()

	assert.NoError(t, err)
	assert.Equal(t, domain.POStatusCancelled, po.Status)
}

func TestPurchaseOrder_Cancel_WithReceivedStatus_ReturnsError(t *testing.T) {
	po := &domain.PurchaseOrder{
		Status: domain.POStatusReceived,
	}

	err := po.Cancel()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot cancel received")
}

func TestPurchaseOrder_UpdateReceiptStatus_WithAllItemsReceived_UpdatesStatusToReceived(t *testing.T) {
	po := &domain.PurchaseOrder{
		Status: domain.POStatusConfirmed,
		LineItems: []domain.POLineItem{
			{Quantity: 10, ReceivedQty: 10},
			{Quantity: 20, ReceivedQty: 20},
		},
	}

	po.UpdateReceiptStatus()

	assert.Equal(t, domain.POStatusReceived, po.Status)
	assert.NotNil(t, po.ActualArrivalDate)
}

func TestPurchaseOrder_UpdateReceiptStatus_WithPartialReceipts_UpdatesStatusToPartial(t *testing.T) {
	po := &domain.PurchaseOrder{
		Status: domain.POStatusConfirmed,
		LineItems: []domain.POLineItem{
			{Quantity: 10, ReceivedQty: 5},
			{Quantity: 20, ReceivedQty: 0},
		},
	}

	po.UpdateReceiptStatus()

	assert.Equal(t, domain.POStatusPartial, po.Status)
}

func TestPOStatus_IsValid_WithValidStatus_ReturnsTrue(t *testing.T) {
	givenStatuses := []domain.POStatus{
		domain.POStatusDraft,
		domain.POStatusPending,
		domain.POStatusConfirmed,
		domain.POStatusPartial,
		domain.POStatusReceived,
		domain.POStatusClosed,
		domain.POStatusCancelled,
	}

	for _, status := range givenStatuses {
		assert.True(t, status.IsValid())
	}
}

func TestPOStatus_IsValid_WithInvalidStatus_ReturnsFalse(t *testing.T) {
	givenStatus := domain.POStatus("invalid_status")

	assert.False(t, givenStatus.IsValid())
}