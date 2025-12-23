package domain_test

import (
	"testing"
	"time"

	"github.com/giia/giia-core-engine/services/execution-service/internal/core/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewSalesOrder_WithValidData_ReturnsSalesOrder(t *testing.T) {
	givenOrgID := uuid.New()
	givenCustomerID := uuid.New()
	givenSONumber := "SO-001"
	givenOrderDate := time.Now()
	givenDueDate := time.Now().AddDate(0, 0, 7)
	givenLineItems := []domain.SOLineItem{
		{
			ID:        uuid.New(),
			ProductID: uuid.New(),
			Quantity:  10,
			UnitPrice: 150,
			LineTotal: 1500,
		},
	}

	so, err := domain.NewSalesOrder(
		givenOrgID,
		givenCustomerID,
		givenSONumber,
		givenOrderDate,
		givenDueDate,
		givenLineItems,
	)

	assert.NoError(t, err)
	assert.NotNil(t, so)
	assert.Equal(t, givenOrgID, so.OrganizationID)
	assert.Equal(t, givenCustomerID, so.CustomerID)
	assert.Equal(t, givenSONumber, so.SONumber)
	assert.Equal(t, domain.SOStatusPending, so.Status)
	assert.Equal(t, float64(1500), so.TotalAmount)
	assert.Equal(t, false, so.DeliveryNoteIssued)
	assert.Len(t, so.LineItems, 1)
}

func TestNewSalesOrder_WithNilOrganizationID_ReturnsError(t *testing.T) {
	givenCustomerID := uuid.New()
	givenSONumber := "SO-001"
	givenOrderDate := time.Now()
	givenDueDate := time.Now().AddDate(0, 0, 7)
	givenLineItems := []domain.SOLineItem{
		{ProductID: uuid.New(), Quantity: 10, UnitPrice: 150, LineTotal: 1500},
	}

	so, err := domain.NewSalesOrder(
		uuid.Nil,
		givenCustomerID,
		givenSONumber,
		givenOrderDate,
		givenDueDate,
		givenLineItems,
	)

	assert.Error(t, err)
	assert.Nil(t, so)
	assert.Contains(t, err.Error(), "organization_id is required")
}

func TestNewSalesOrder_WithNilCustomerID_ReturnsError(t *testing.T) {
	givenOrgID := uuid.New()
	givenSONumber := "SO-001"
	givenOrderDate := time.Now()
	givenDueDate := time.Now().AddDate(0, 0, 7)
	givenLineItems := []domain.SOLineItem{
		{ProductID: uuid.New(), Quantity: 10, UnitPrice: 150, LineTotal: 1500},
	}

	so, err := domain.NewSalesOrder(
		givenOrgID,
		uuid.Nil,
		givenSONumber,
		givenOrderDate,
		givenDueDate,
		givenLineItems,
	)

	assert.Error(t, err)
	assert.Nil(t, so)
	assert.Contains(t, err.Error(), "customer_id is required")
}

func TestNewSalesOrder_WithEmptySONumber_ReturnsError(t *testing.T) {
	givenOrgID := uuid.New()
	givenCustomerID := uuid.New()
	givenOrderDate := time.Now()
	givenDueDate := time.Now().AddDate(0, 0, 7)
	givenLineItems := []domain.SOLineItem{
		{ProductID: uuid.New(), Quantity: 10, UnitPrice: 150, LineTotal: 1500},
	}

	so, err := domain.NewSalesOrder(
		givenOrgID,
		givenCustomerID,
		"",
		givenOrderDate,
		givenDueDate,
		givenLineItems,
	)

	assert.Error(t, err)
	assert.Nil(t, so)
	assert.Contains(t, err.Error(), "so_number is required")
}

func TestNewSalesOrder_WithEmptyLineItems_ReturnsError(t *testing.T) {
	givenOrgID := uuid.New()
	givenCustomerID := uuid.New()
	givenSONumber := "SO-001"
	givenOrderDate := time.Now()
	givenDueDate := time.Now().AddDate(0, 0, 7)

	so, err := domain.NewSalesOrder(
		givenOrgID,
		givenCustomerID,
		givenSONumber,
		givenOrderDate,
		givenDueDate,
		[]domain.SOLineItem{},
	)

	assert.Error(t, err)
	assert.Nil(t, so)
	assert.Contains(t, err.Error(), "at least one line item is required")
}

func TestSalesOrder_IsQualifiedDemand_WhenConfirmedAndNoDeliveryNote_ReturnsTrue(t *testing.T) {
	so := &domain.SalesOrder{
		Status:             domain.SOStatusConfirmed,
		DeliveryNoteIssued: false,
	}

	result := so.IsQualifiedDemand()

	assert.True(t, result)
}

func TestSalesOrder_IsQualifiedDemand_WhenDeliveryNoteIssued_ReturnsFalse(t *testing.T) {
	so := &domain.SalesOrder{
		Status:             domain.SOStatusConfirmed,
		DeliveryNoteIssued: true,
	}

	result := so.IsQualifiedDemand()

	assert.False(t, result)
}

func TestSalesOrder_IsQualifiedDemand_WhenNotConfirmed_ReturnsFalse(t *testing.T) {
	so := &domain.SalesOrder{
		Status:             domain.SOStatusPending,
		DeliveryNoteIssued: false,
	}

	result := so.IsQualifiedDemand()

	assert.False(t, result)
}

func TestSalesOrder_IssueDeliveryNote_WithValidData_IssuesDeliveryNote(t *testing.T) {
	givenNoteNumber := "DN-001"
	so := &domain.SalesOrder{
		Status:             domain.SOStatusConfirmed,
		DeliveryNoteIssued: false,
	}

	err := so.IssueDeliveryNote(givenNoteNumber)

	assert.NoError(t, err)
	assert.True(t, so.DeliveryNoteIssued)
	assert.Equal(t, givenNoteNumber, so.DeliveryNoteNumber)
	assert.NotNil(t, so.DeliveryNoteDate)
}

func TestSalesOrder_IssueDeliveryNote_WhenAlreadyIssued_ReturnsError(t *testing.T) {
	so := &domain.SalesOrder{
		Status:             domain.SOStatusConfirmed,
		DeliveryNoteIssued: true,
		DeliveryNoteNumber: "DN-001",
	}

	err := so.IssueDeliveryNote("DN-002")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "delivery note already issued")
}

func TestSalesOrder_IssueDeliveryNote_WithEmptyNoteNumber_ReturnsError(t *testing.T) {
	so := &domain.SalesOrder{
		Status:             domain.SOStatusConfirmed,
		DeliveryNoteIssued: false,
	}

	err := so.IssueDeliveryNote("")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "delivery note number is required")
}

func TestSalesOrder_IssueDeliveryNote_WithInvalidStatus_ReturnsError(t *testing.T) {
	so := &domain.SalesOrder{
		Status:             domain.SOStatusPending,
		DeliveryNoteIssued: false,
	}

	err := so.IssueDeliveryNote("DN-001")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "can only issue delivery note for confirmed")
}

func TestSalesOrder_Confirm_WithPendingStatus_UpdatesStatusToConfirmed(t *testing.T) {
	so := &domain.SalesOrder{
		Status: domain.SOStatusPending,
	}

	err := so.Confirm()

	assert.NoError(t, err)
	assert.Equal(t, domain.SOStatusConfirmed, so.Status)
}

func TestSalesOrder_Confirm_WithConfirmedStatus_ReturnsError(t *testing.T) {
	so := &domain.SalesOrder{
		Status: domain.SOStatusConfirmed,
	}

	err := so.Confirm()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "can only confirm pending")
}

func TestSalesOrder_Cancel_WithPendingStatus_UpdatesStatusToCancelled(t *testing.T) {
	so := &domain.SalesOrder{
		Status: domain.SOStatusPending,
	}

	err := so.Cancel()

	assert.NoError(t, err)
	assert.Equal(t, domain.SOStatusCancelled, so.Status)
}

func TestSalesOrder_Cancel_WithShippedStatus_ReturnsError(t *testing.T) {
	so := &domain.SalesOrder{
		Status: domain.SOStatusShipped,
	}

	err := so.Cancel()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot cancel shipped")
}

func TestSalesOrder_MarkAsShipped_WithPackedStatus_UpdatesStatusToShipped(t *testing.T) {
	so := &domain.SalesOrder{
		Status: domain.SOStatusPacked,
	}

	err := so.MarkAsShipped()

	assert.NoError(t, err)
	assert.Equal(t, domain.SOStatusShipped, so.Status)
	assert.NotNil(t, so.ShipDate)
}

func TestSalesOrder_MarkAsShipped_WithPendingStatus_ReturnsError(t *testing.T) {
	so := &domain.SalesOrder{
		Status: domain.SOStatusPending,
	}

	err := so.MarkAsShipped()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "can only ship packed orders")
}

func TestSOStatus_IsValid_WithValidStatus_ReturnsTrue(t *testing.T) {
	givenStatuses := []domain.SOStatus{
		domain.SOStatusPending,
		domain.SOStatusConfirmed,
		domain.SOStatusPicking,
		domain.SOStatusPacked,
		domain.SOStatusShipped,
		domain.SOStatusDelivered,
		domain.SOStatusCancelled,
	}

	for _, status := range givenStatuses {
		assert.True(t, status.IsValid())
	}
}

func TestSOStatus_IsValid_WithInvalidStatus_ReturnsFalse(t *testing.T) {
	givenStatus := domain.SOStatus("invalid_status")

	assert.False(t, givenStatus.IsValid())
}