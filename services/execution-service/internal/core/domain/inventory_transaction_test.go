package domain_test

import (
	"testing"

	"github.com/giia/giia-core-engine/services/execution-service/internal/core/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewInventoryTransaction_WithValidData_ReturnsInventoryTransaction(t *testing.T) {
	givenOrgID := uuid.New()
	givenProductID := uuid.New()
	givenLocationID := uuid.New()
	givenCreatedBy := uuid.New()
	givenType := domain.TransactionReceipt
	givenQuantity := 100.0
	givenUnitCost := 50.0
	givenReferenceType := "purchase_order"
	givenReferenceID := uuid.New()
	givenReason := "PO receipt"

	txn, err := domain.NewInventoryTransaction(
		givenOrgID,
		givenProductID,
		givenLocationID,
		givenCreatedBy,
		givenType,
		givenQuantity,
		givenUnitCost,
		givenReferenceType,
		givenReferenceID,
		givenReason,
	)

	assert.NoError(t, err)
	assert.NotNil(t, txn)
	assert.Equal(t, givenOrgID, txn.OrganizationID)
	assert.Equal(t, givenProductID, txn.ProductID)
	assert.Equal(t, givenLocationID, txn.LocationID)
	assert.Equal(t, givenCreatedBy, txn.CreatedBy)
	assert.Equal(t, givenType, txn.Type)
	assert.Equal(t, givenQuantity, txn.Quantity)
	assert.Equal(t, givenUnitCost, txn.UnitCost)
	assert.Equal(t, givenReferenceType, txn.ReferenceType)
	assert.Equal(t, givenReferenceID, txn.ReferenceID)
	assert.Equal(t, givenReason, txn.Reason)
}

func TestNewInventoryTransaction_WithNilOrganizationID_ReturnsError(t *testing.T) {
	txn, err := domain.NewInventoryTransaction(
		uuid.Nil,
		uuid.New(),
		uuid.New(),
		uuid.New(),
		domain.TransactionReceipt,
		100.0,
		50.0,
		"purchase_order",
		uuid.New(),
		"test",
	)

	assert.Error(t, err)
	assert.Nil(t, txn)
	assert.Contains(t, err.Error(), "organization_id is required")
}

func TestNewInventoryTransaction_WithNilProductID_ReturnsError(t *testing.T) {
	txn, err := domain.NewInventoryTransaction(
		uuid.New(),
		uuid.Nil,
		uuid.New(),
		uuid.New(),
		domain.TransactionReceipt,
		100.0,
		50.0,
		"purchase_order",
		uuid.New(),
		"test",
	)

	assert.Error(t, err)
	assert.Nil(t, txn)
	assert.Contains(t, err.Error(), "product_id is required")
}

func TestNewInventoryTransaction_WithNilLocationID_ReturnsError(t *testing.T) {
	txn, err := domain.NewInventoryTransaction(
		uuid.New(),
		uuid.New(),
		uuid.Nil,
		uuid.New(),
		domain.TransactionReceipt,
		100.0,
		50.0,
		"purchase_order",
		uuid.New(),
		"test",
	)

	assert.Error(t, err)
	assert.Nil(t, txn)
	assert.Contains(t, err.Error(), "location_id is required")
}

func TestNewInventoryTransaction_WithNilCreatedBy_ReturnsError(t *testing.T) {
	txn, err := domain.NewInventoryTransaction(
		uuid.New(),
		uuid.New(),
		uuid.New(),
		uuid.Nil,
		domain.TransactionReceipt,
		100.0,
		50.0,
		"purchase_order",
		uuid.New(),
		"test",
	)

	assert.Error(t, err)
	assert.Nil(t, txn)
	assert.Contains(t, err.Error(), "created_by is required")
}

func TestNewInventoryTransaction_WithInvalidType_ReturnsError(t *testing.T) {
	txn, err := domain.NewInventoryTransaction(
		uuid.New(),
		uuid.New(),
		uuid.New(),
		uuid.New(),
		domain.TransactionType("invalid"),
		100.0,
		50.0,
		"purchase_order",
		uuid.New(),
		"test",
	)

	assert.Error(t, err)
	assert.Nil(t, txn)
	assert.Contains(t, err.Error(), "invalid transaction type")
}

func TestNewInventoryTransaction_WithZeroQuantity_ReturnsError(t *testing.T) {
	txn, err := domain.NewInventoryTransaction(
		uuid.New(),
		uuid.New(),
		uuid.New(),
		uuid.New(),
		domain.TransactionReceipt,
		0,
		50.0,
		"purchase_order",
		uuid.New(),
		"test",
	)

	assert.Error(t, err)
	assert.Nil(t, txn)
	assert.Contains(t, err.Error(), "quantity cannot be zero")
}

func TestNewInventoryTransaction_WithNegativeUnitCost_ReturnsError(t *testing.T) {
	txn, err := domain.NewInventoryTransaction(
		uuid.New(),
		uuid.New(),
		uuid.New(),
		uuid.New(),
		domain.TransactionReceipt,
		100.0,
		-50.0,
		"purchase_order",
		uuid.New(),
		"test",
	)

	assert.Error(t, err)
	assert.Nil(t, txn)
	assert.Contains(t, err.Error(), "unit_cost cannot be negative")
}

func TestTransactionType_IsValid_WithValidTypes_ReturnsTrue(t *testing.T) {
	givenTypes := []domain.TransactionType{
		domain.TransactionReceipt,
		domain.TransactionIssue,
		domain.TransactionTransfer,
		domain.TransactionAdjustment,
	}

	for _, txnType := range givenTypes {
		assert.True(t, txnType.IsValid())
	}
}

func TestTransactionType_IsValid_WithInvalidType_ReturnsFalse(t *testing.T) {
	givenType := domain.TransactionType("invalid_type")

	assert.False(t, givenType.IsValid())
}