package postgres

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/melegattip/giia-core-engine/services/execution-service/internal/core/domain"
	"github.com/stretchr/testify/assert"
)

func TestSalesOrderModel_TableName(t *testing.T) {
	model := SalesOrderModel{}
	assert.Equal(t, "sales_orders", model.TableName())
}

func TestSOLineItemModel_TableName(t *testing.T) {
	model := SOLineItemModel{}
	assert.Equal(t, "sales_order_lines", model.TableName())
}

func TestSalesOrderRepository_ToModel(t *testing.T) {
	repo := &salesOrderRepository{}
	now := time.Now()
	soID := uuid.New()
	orgID := uuid.New()
	customerID := uuid.New()
	productID := uuid.New()

	so := &domain.SalesOrder{
		ID:                 soID,
		OrganizationID:     orgID,
		SONumber:           "SO-001",
		CustomerID:         customerID,
		Status:             domain.SOStatusPending,
		OrderDate:          now,
		DueDate:            now.Add(3 * 24 * time.Hour),
		TotalAmount:        500.00,
		DeliveryNoteIssued: false,
		CreatedAt:          now,
		UpdatedAt:          now,
		LineItems: []domain.SOLineItem{
			{
				ID:           uuid.New(),
				SalesOrderID: soID,
				ProductID:    productID,
				Quantity:     5,
				UnitPrice:    100.00,
				LineTotal:    500.00,
			},
		},
	}

	model := repo.toModel(so)

	assert.Equal(t, soID, model.ID)
	assert.Equal(t, orgID, model.OrganizationID)
	assert.Equal(t, "SO-001", model.SONumber)
	assert.Equal(t, customerID, model.CustomerID)
	assert.Equal(t, "pending", model.Status)
	assert.Equal(t, 500.00, model.TotalAmount)
	assert.False(t, model.DeliveryNoteIssued)
	assert.Len(t, model.LineItems, 1)
	assert.Equal(t, productID, model.LineItems[0].ProductID)
}

func TestSalesOrderRepository_ToDomain(t *testing.T) {
	repo := &salesOrderRepository{}
	now := time.Now()
	soID := uuid.New()
	orgID := uuid.New()
	customerID := uuid.New()
	productID := uuid.New()
	shipDate := now.Add(24 * time.Hour)
	noteDate := now.Add(12 * time.Hour)

	model := &SalesOrderModel{
		ID:                 soID,
		OrganizationID:     orgID,
		SONumber:           "SO-002",
		CustomerID:         customerID,
		Status:             "confirmed",
		OrderDate:          now,
		DueDate:            now.Add(3 * 24 * time.Hour),
		ShipDate:           &shipDate,
		DeliveryNoteIssued: true,
		DeliveryNoteNumber: "DN-001",
		DeliveryNoteDate:   &noteDate,
		TotalAmount:        750.00,
		CreatedAt:          now,
		UpdatedAt:          now,
		LineItems: []SOLineItemModel{
			{
				ID:           uuid.New(),
				SalesOrderID: soID,
				ProductID:    productID,
				Quantity:     10,
				UnitPrice:    75.00,
				LineTotal:    750.00,
			},
		},
	}

	so := repo.toDomain(model)

	assert.Equal(t, soID, so.ID)
	assert.Equal(t, orgID, so.OrganizationID)
	assert.Equal(t, "SO-002", so.SONumber)
	assert.Equal(t, customerID, so.CustomerID)
	assert.Equal(t, domain.SOStatusConfirmed, so.Status)
	assert.Equal(t, 750.00, so.TotalAmount)
	assert.True(t, so.DeliveryNoteIssued)
	assert.Equal(t, "DN-001", so.DeliveryNoteNumber)
	assert.NotNil(t, so.ShipDate)
	assert.NotNil(t, so.DeliveryNoteDate)
	assert.Len(t, so.LineItems, 1)
	assert.Equal(t, productID, so.LineItems[0].ProductID)
}

func TestSalesOrderRepository_ToModelWithNilDates(t *testing.T) {
	repo := &salesOrderRepository{}
	now := time.Now()

	so := &domain.SalesOrder{
		ID:             uuid.New(),
		OrganizationID: uuid.New(),
		SONumber:       "SO-003",
		CustomerID:     uuid.New(),
		Status:         domain.SOStatusPending,
		OrderDate:      now,
		DueDate:        now.Add(3 * 24 * time.Hour),
		ShipDate:       nil,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	model := repo.toModel(so)

	assert.Nil(t, model.ShipDate)
	assert.Nil(t, model.DeliveryNoteDate)
}

func TestSalesOrderRepository_ToDomainWithNilDates(t *testing.T) {
	repo := &salesOrderRepository{}
	now := time.Now()

	model := &SalesOrderModel{
		ID:             uuid.New(),
		OrganizationID: uuid.New(),
		SONumber:       "SO-004",
		CustomerID:     uuid.New(),
		Status:         "pending",
		OrderDate:      now,
		DueDate:        now.Add(3 * 24 * time.Hour),
		ShipDate:       nil,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	so := repo.toDomain(model)

	assert.Nil(t, so.ShipDate)
	assert.Nil(t, so.DeliveryNoteDate)
}

func BenchmarkSalesOrderToModel(b *testing.B) {
	repo := &salesOrderRepository{}
	now := time.Now()
	soID := uuid.New()

	so := &domain.SalesOrder{
		ID:             soID,
		OrganizationID: uuid.New(),
		SONumber:       "SO-BENCH",
		CustomerID:     uuid.New(),
		Status:         domain.SOStatusPending,
		OrderDate:      now,
		DueDate:        now.Add(3 * 24 * time.Hour),
		CreatedAt:      now,
		UpdatedAt:      now,
		LineItems: []domain.SOLineItem{
			{ID: uuid.New(), SalesOrderID: soID, ProductID: uuid.New(), Quantity: 5},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = repo.toModel(so)
	}
}
