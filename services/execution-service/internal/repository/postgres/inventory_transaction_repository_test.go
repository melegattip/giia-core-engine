package postgres

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/melegattip/giia-core-engine/services/execution-service/internal/core/domain"
	"github.com/stretchr/testify/assert"
)

func TestInventoryTransactionModel_TableName(t *testing.T) {
	model := InventoryTransactionModel{}
	assert.Equal(t, "inventory_transactions", model.TableName())
}

func TestInventoryTransactionRepository_ToModel(t *testing.T) {
	repo := &inventoryTransactionRepository{}
	now := time.Now()
	txnID := uuid.New()
	orgID := uuid.New()
	productID := uuid.New()
	locationID := uuid.New()
	referenceID := uuid.New()
	createdBy := uuid.New()

	txn := &domain.InventoryTransaction{
		ID:              txnID,
		OrganizationID:  orgID,
		ProductID:       productID,
		LocationID:      locationID,
		Type:            domain.TransactionReceipt,
		Quantity:        100,
		UnitCost:        10.50,
		ReferenceType:   "purchase_order",
		ReferenceID:     referenceID,
		Reason:          "Stock receipt from supplier",
		TransactionDate: now,
		CreatedBy:       createdBy,
		CreatedAt:       now,
	}

	model := repo.toModel(txn)

	assert.Equal(t, txnID, model.ID)
	assert.Equal(t, orgID, model.OrganizationID)
	assert.Equal(t, productID, model.ProductID)
	assert.Equal(t, locationID, model.LocationID)
	assert.Equal(t, "receipt", model.Type)
	assert.Equal(t, float64(100), model.Quantity)
	assert.Equal(t, 10.50, model.UnitCost)
	assert.Equal(t, "purchase_order", model.ReferenceType)
	assert.Equal(t, referenceID, model.ReferenceID)
	assert.Equal(t, "Stock receipt from supplier", model.Reason)
}

func TestInventoryTransactionRepository_ToDomain(t *testing.T) {
	repo := &inventoryTransactionRepository{}
	now := time.Now()
	txnID := uuid.New()
	orgID := uuid.New()
	productID := uuid.New()
	locationID := uuid.New()
	referenceID := uuid.New()
	createdBy := uuid.New()

	model := &InventoryTransactionModel{
		ID:              txnID,
		OrganizationID:  orgID,
		ProductID:       productID,
		LocationID:      locationID,
		Type:            "issue",
		Quantity:        -50,
		UnitCost:        10.50,
		ReferenceType:   "sales_order",
		ReferenceID:     referenceID,
		Reason:          "Stock issued for order",
		TransactionDate: now,
		CreatedBy:       createdBy,
		CreatedAt:       now,
	}

	txn := repo.toDomain(model)

	assert.Equal(t, txnID, txn.ID)
	assert.Equal(t, orgID, txn.OrganizationID)
	assert.Equal(t, productID, txn.ProductID)
	assert.Equal(t, locationID, txn.LocationID)
	assert.Equal(t, domain.TransactionIssue, txn.Type)
	assert.Equal(t, float64(-50), txn.Quantity)
	assert.Equal(t, 10.50, txn.UnitCost)
	assert.Equal(t, "sales_order", txn.ReferenceType)
	assert.Equal(t, referenceID, txn.ReferenceID)
	assert.Equal(t, "Stock issued for order", txn.Reason)
}

func TestInventoryTransactionRepository_AllTransactionTypes(t *testing.T) {
	repo := &inventoryTransactionRepository{}
	now := time.Now()

	testCases := []struct {
		domainType domain.TransactionType
		modelType  string
	}{
		{domain.TransactionReceipt, "receipt"},
		{domain.TransactionIssue, "issue"},
		{domain.TransactionTransfer, "transfer"},
		{domain.TransactionAdjustment, "adjustment"},
	}

	for _, tc := range testCases {
		t.Run(string(tc.domainType), func(t *testing.T) {
			txn := &domain.InventoryTransaction{
				ID:              uuid.New(),
				OrganizationID:  uuid.New(),
				ProductID:       uuid.New(),
				LocationID:      uuid.New(),
				Type:            tc.domainType,
				Quantity:        100,
				TransactionDate: now,
				CreatedBy:       uuid.New(),
				CreatedAt:       now,
			}

			model := repo.toModel(txn)
			assert.Equal(t, tc.modelType, model.Type)

			// Convert back
			domainAgain := repo.toDomain(model)
			assert.Equal(t, tc.domainType, domainAgain.Type)
		})
	}
}

func BenchmarkInventoryTransactionToModel(b *testing.B) {
	repo := &inventoryTransactionRepository{}
	now := time.Now()

	txn := &domain.InventoryTransaction{
		ID:              uuid.New(),
		OrganizationID:  uuid.New(),
		ProductID:       uuid.New(),
		LocationID:      uuid.New(),
		Type:            domain.TransactionReceipt,
		Quantity:        100,
		UnitCost:        10.50,
		ReferenceType:   "purchase_order",
		ReferenceID:     uuid.New(),
		TransactionDate: now,
		CreatedBy:       uuid.New(),
		CreatedAt:       now,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = repo.toModel(txn)
	}
}
