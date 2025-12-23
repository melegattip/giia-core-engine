package domain_test

import (
	"testing"
	"time"

	"github.com/melegattip/giia-core-engine/services/analytics-service/internal/core/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewDaysInInventoryKPI_WithValidData_CreatesKPI(t *testing.T) {
	givenOrgID := uuid.New()
	givenDate := time.Now()
	givenTotalValuedDays := 50000.0
	givenAvgValuedDays := 250.0
	givenTotalProducts := 200

	kpi, err := domain.NewDaysInInventoryKPI(
		givenOrgID,
		givenDate,
		givenTotalValuedDays,
		givenAvgValuedDays,
		givenTotalProducts,
	)

	assert.NoError(t, err)
	assert.NotNil(t, kpi)
	assert.Equal(t, givenOrgID, kpi.OrganizationID)
	assert.Equal(t, givenTotalValuedDays, kpi.TotalValuedDays)
	assert.Equal(t, givenAvgValuedDays, kpi.AverageValuedDays)
	assert.Equal(t, givenTotalProducts, kpi.TotalProducts)
}

func TestNewDaysInInventoryKPI_WithNilOrganizationID_ReturnsError(t *testing.T) {
	kpi, err := domain.NewDaysInInventoryKPI(
		uuid.Nil,
		time.Now(),
		1000.0,
		10.0,
		100,
	)

	assert.Error(t, err)
	assert.Nil(t, kpi)
	assert.Contains(t, err.Error(), "organization_id is required")
}

func TestNewDaysInInventoryKPI_WithZeroSnapshotDate_ReturnsError(t *testing.T) {
	kpi, err := domain.NewDaysInInventoryKPI(
		uuid.New(),
		time.Time{},
		1000.0,
		10.0,
		100,
	)

	assert.Error(t, err)
	assert.Nil(t, kpi)
	assert.Contains(t, err.Error(), "snapshot_date is required")
}

func TestNewDaysInInventoryKPI_WithNegativeTotalValuedDays_ReturnsError(t *testing.T) {
	kpi, err := domain.NewDaysInInventoryKPI(
		uuid.New(),
		time.Now(),
		-1000.0,
		10.0,
		100,
	)

	assert.Error(t, err)
	assert.Nil(t, kpi)
	assert.Contains(t, err.Error(), "total_valued_days cannot be negative")
}

func TestNewDaysInInventoryKPI_WithNegativeAverageValuedDays_ReturnsError(t *testing.T) {
	kpi, err := domain.NewDaysInInventoryKPI(
		uuid.New(),
		time.Now(),
		1000.0,
		-10.0,
		100,
	)

	assert.Error(t, err)
	assert.Nil(t, kpi)
	assert.Contains(t, err.Error(), "average_valued_days cannot be negative")
}

func TestNewDaysInInventoryKPI_WithNegativeTotalProducts_ReturnsError(t *testing.T) {
	kpi, err := domain.NewDaysInInventoryKPI(
		uuid.New(),
		time.Now(),
		1000.0,
		10.0,
		-100,
	)

	assert.Error(t, err)
	assert.Nil(t, kpi)
	assert.Contains(t, err.Error(), "total_products cannot be negative")
}

func TestCalculateValuedDays_WithValidProduct_CalculatesCorrectly(t *testing.T) {
	purchaseDate := time.Now().AddDate(0, 0, -30)
	currentDate := time.Now()

	product := domain.ProductInventoryAge{
		Quantity:     100,
		UnitCost:     50,
		PurchaseDate: purchaseDate,
	}

	valuedDays := domain.CalculateValuedDays(product, currentDate)

	expectedTotalValue := 100.0 * 50.0
	expectedDays := int(currentDate.Sub(purchaseDate).Hours() / 24)
	expectedValuedDays := float64(expectedDays) * expectedTotalValue

	assert.InDelta(t, expectedValuedDays, valuedDays, 1.0)
}

func TestCalculateValuedDays_WithFuturePurchaseDate_ReturnsZero(t *testing.T) {
	purchaseDate := time.Now().AddDate(0, 0, 10)
	currentDate := time.Now()

	product := domain.ProductInventoryAge{
		Quantity:     100,
		UnitCost:     50,
		PurchaseDate: purchaseDate,
	}

	valuedDays := domain.CalculateValuedDays(product, currentDate)

	assert.Equal(t, 0.0, valuedDays)
}
