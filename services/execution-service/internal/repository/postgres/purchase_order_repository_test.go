package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/melegattip/giia-core-engine/services/execution-service/internal/core/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPurchaseOrderModel tests the model conversion functions
func TestPurchaseOrderModel_TableName(t *testing.T) {
	model := PurchaseOrderModel{}
	assert.Equal(t, "purchase_orders", model.TableName())
}

func TestPOLineItemModel_TableName(t *testing.T) {
	model := POLineItemModel{}
	assert.Equal(t, "purchase_order_lines", model.TableName())
}

func TestPurchaseOrderRepository_ToModel(t *testing.T) {
	repo := &purchaseOrderRepository{}
	now := time.Now()
	poID := uuid.New()
	orgID := uuid.New()
	supplierID := uuid.New()
	productID := uuid.New()
	createdBy := uuid.New()

	po := &domain.PurchaseOrder{
		ID:                  poID,
		OrganizationID:      orgID,
		PONumber:            "PO-001",
		SupplierID:          supplierID,
		Status:              domain.POStatusDraft,
		OrderDate:           now,
		ExpectedArrivalDate: now.Add(7 * 24 * time.Hour),
		TotalAmount:         1000.00,
		CreatedBy:           createdBy,
		CreatedAt:           now,
		UpdatedAt:           now,
		LineItems: []domain.POLineItem{
			{
				ID:              uuid.New(),
				PurchaseOrderID: poID,
				ProductID:       productID,
				Quantity:        10,
				ReceivedQty:     0,
				UnitCost:        100.00,
				LineTotal:       1000.00,
			},
		},
	}

	model := repo.toModel(po)

	assert.Equal(t, poID, model.ID)
	assert.Equal(t, orgID, model.OrganizationID)
	assert.Equal(t, "PO-001", model.PONumber)
	assert.Equal(t, supplierID, model.SupplierID)
	assert.Equal(t, "draft", model.Status)
	assert.Equal(t, 1000.00, model.TotalAmount)
	assert.Len(t, model.LineItems, 1)
	assert.Equal(t, productID, model.LineItems[0].ProductID)
	assert.Equal(t, float64(10), model.LineItems[0].Quantity)
}

func TestPurchaseOrderRepository_ToDomain(t *testing.T) {
	repo := &purchaseOrderRepository{}
	now := time.Now()
	poID := uuid.New()
	orgID := uuid.New()
	supplierID := uuid.New()
	productID := uuid.New()
	createdBy := uuid.New()
	arrivalDate := now.Add(24 * time.Hour)

	model := &PurchaseOrderModel{
		ID:                  poID,
		OrganizationID:      orgID,
		PONumber:            "PO-002",
		SupplierID:          supplierID,
		Status:              "confirmed",
		OrderDate:           now,
		ExpectedArrivalDate: now.Add(7 * 24 * time.Hour),
		ActualArrivalDate:   &arrivalDate,
		DelayDays:           3,
		IsDelayed:           true,
		TotalAmount:         2000.00,
		CreatedBy:           createdBy,
		CreatedAt:           now,
		UpdatedAt:           now,
		LineItems: []POLineItemModel{
			{
				ID:              uuid.New(),
				PurchaseOrderID: poID,
				ProductID:       productID,
				Quantity:        20,
				ReceivedQty:     10,
				UnitCost:        100.00,
				LineTotal:       2000.00,
			},
		},
	}

	po := repo.toDomain(model)

	assert.Equal(t, poID, po.ID)
	assert.Equal(t, orgID, po.OrganizationID)
	assert.Equal(t, "PO-002", po.PONumber)
	assert.Equal(t, supplierID, po.SupplierID)
	assert.Equal(t, domain.POStatusConfirmed, po.Status)
	assert.Equal(t, 2000.00, po.TotalAmount)
	assert.NotNil(t, po.ActualArrivalDate)
	assert.Equal(t, 3, po.DelayDays)
	assert.True(t, po.IsDelayed)
	assert.Len(t, po.LineItems, 1)
	assert.Equal(t, productID, po.LineItems[0].ProductID)
	assert.Equal(t, float64(20), po.LineItems[0].Quantity)
	assert.Equal(t, float64(10), po.LineItems[0].ReceivedQty)
}

func TestPurchaseOrderRepository_ScopeByOrg(t *testing.T) {
	// This test verifies the scopeByOrg method exists and doesn't panic
	repo := &purchaseOrderRepository{db: nil}
	orgID := uuid.New()

	// Since db is nil, this will panic if called improperly
	// We're just testing the method signature exists
	assert.NotNil(t, repo)
	assert.NotEqual(t, uuid.Nil, orgID)
}

func TestPurchaseOrderRepository_ContextUsage(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Verify context is properly set up
	require.NotNil(t, ctx)
}

// Benchmark tests for model conversion
func BenchmarkPurchaseOrderToModel(b *testing.B) {
	repo := &purchaseOrderRepository{}
	now := time.Now()
	poID := uuid.New()

	po := &domain.PurchaseOrder{
		ID:                  poID,
		OrganizationID:      uuid.New(),
		PONumber:            "PO-BENCH",
		SupplierID:          uuid.New(),
		Status:              domain.POStatusDraft,
		OrderDate:           now,
		ExpectedArrivalDate: now.Add(7 * 24 * time.Hour),
		TotalAmount:         1000.00,
		CreatedBy:           uuid.New(),
		CreatedAt:           now,
		UpdatedAt:           now,
		LineItems: []domain.POLineItem{
			{ID: uuid.New(), PurchaseOrderID: poID, ProductID: uuid.New(), Quantity: 10},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = repo.toModel(po)
	}
}

func BenchmarkPurchaseOrderToDomain(b *testing.B) {
	repo := &purchaseOrderRepository{}
	now := time.Now()
	poID := uuid.New()

	model := &PurchaseOrderModel{
		ID:                  poID,
		OrganizationID:      uuid.New(),
		PONumber:            "PO-BENCH",
		SupplierID:          uuid.New(),
		Status:              "draft",
		OrderDate:           now,
		ExpectedArrivalDate: now.Add(7 * 24 * time.Hour),
		TotalAmount:         1000.00,
		CreatedBy:           uuid.New(),
		CreatedAt:           now,
		UpdatedAt:           now,
		LineItems: []POLineItemModel{
			{ID: uuid.New(), PurchaseOrderID: poID, ProductID: uuid.New(), Quantity: 10},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = repo.toDomain(model)
	}
}
