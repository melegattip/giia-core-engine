package postgres

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestInventoryBalanceModel_TableName(t *testing.T) {
	model := InventoryBalanceModel{}
	assert.Equal(t, "inventory_balances", model.TableName())
}

func TestInventoryBalanceRepository_ToDomain(t *testing.T) {
	repo := &inventoryBalanceRepository{}
	now := time.Now()
	balID := uuid.New()
	orgID := uuid.New()
	productID := uuid.New()
	locationID := uuid.New()

	model := &InventoryBalanceModel{
		ID:             balID,
		OrganizationID: orgID,
		ProductID:      productID,
		LocationID:     locationID,
		OnHand:         100,
		Reserved:       30,
		Available:      70,
		UpdatedAt:      now,
	}

	balance := repo.toDomain(model)

	assert.Equal(t, balID, balance.ID)
	assert.Equal(t, orgID, balance.OrganizationID)
	assert.Equal(t, productID, balance.ProductID)
	assert.Equal(t, locationID, balance.LocationID)
	assert.Equal(t, float64(100), balance.OnHand)
	assert.Equal(t, float64(30), balance.Reserved)
	assert.Equal(t, float64(70), balance.Available)
}

func TestInventoryBalanceRepository_ToDomainZeroValues(t *testing.T) {
	repo := &inventoryBalanceRepository{}
	now := time.Now()

	model := &InventoryBalanceModel{
		ID:             uuid.New(),
		OrganizationID: uuid.New(),
		ProductID:      uuid.New(),
		LocationID:     uuid.New(),
		OnHand:         0,
		Reserved:       0,
		Available:      0,
		UpdatedAt:      now,
	}

	balance := repo.toDomain(model)

	assert.Equal(t, float64(0), balance.OnHand)
	assert.Equal(t, float64(0), balance.Reserved)
	assert.Equal(t, float64(0), balance.Available)
}

func TestInventoryBalanceRepository_ToDomainNegativeOnHand(t *testing.T) {
	repo := &inventoryBalanceRepository{}
	now := time.Now()

	// Edge case: negative on-hand (can happen with adjustments)
	model := &InventoryBalanceModel{
		ID:             uuid.New(),
		OrganizationID: uuid.New(),
		ProductID:      uuid.New(),
		LocationID:     uuid.New(),
		OnHand:         -10,
		Reserved:       0,
		Available:      0,
		UpdatedAt:      now,
	}

	balance := repo.toDomain(model)

	assert.Equal(t, float64(-10), balance.OnHand)
	assert.Equal(t, float64(0), balance.Available)
}

func TestInventoryBalanceRepository_ToDomainOverReserved(t *testing.T) {
	repo := &inventoryBalanceRepository{}
	now := time.Now()

	// Edge case: reserved exceeds on-hand
	model := &InventoryBalanceModel{
		ID:             uuid.New(),
		OrganizationID: uuid.New(),
		ProductID:      uuid.New(),
		LocationID:     uuid.New(),
		OnHand:         50,
		Reserved:       80,
		Available:      0, // GREATEST(on_hand - reserved, 0)
		UpdatedAt:      now,
	}

	balance := repo.toDomain(model)

	assert.Equal(t, float64(50), balance.OnHand)
	assert.Equal(t, float64(80), balance.Reserved)
	assert.Equal(t, float64(0), balance.Available)
}

func BenchmarkInventoryBalanceToDomain(b *testing.B) {
	repo := &inventoryBalanceRepository{}
	now := time.Now()

	model := &InventoryBalanceModel{
		ID:             uuid.New(),
		OrganizationID: uuid.New(),
		ProductID:      uuid.New(),
		LocationID:     uuid.New(),
		OnHand:         100,
		Reserved:       30,
		Available:      70,
		UpdatedAt:      now,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = repo.toDomain(model)
	}
}
