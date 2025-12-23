package domain_test

import (
	"testing"
	"time"

	"github.com/melegattip/giia-core-engine/services/analytics-service/internal/core/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewInventoryRotationKPI_WithValidData_CreatesKPI(t *testing.T) {
	givenOrgID := uuid.New()
	givenDate := time.Now()
	givenSales30Days := 150000.0
	givenAvgStock := 50000.0
	givenTopProducts := []domain.RotatingProduct{
		{ProductID: uuid.New(), SKU: "P001", Name: "Product 1", Sales30Days: 10000, AvgStockValue: 2000, RotationRatio: 5.0},
	}
	givenSlowProducts := []domain.RotatingProduct{
		{ProductID: uuid.New(), SKU: "P002", Name: "Product 2", Sales30Days: 100, AvgStockValue: 1000, RotationRatio: 0.1},
	}

	kpi, err := domain.NewInventoryRotationKPI(
		givenOrgID,
		givenDate,
		givenSales30Days,
		givenAvgStock,
		givenTopProducts,
		givenSlowProducts,
	)

	assert.NoError(t, err)
	assert.NotNil(t, kpi)
	assert.Equal(t, givenOrgID, kpi.OrganizationID)
	assert.Equal(t, givenSales30Days, kpi.SalesLast30Days)
	assert.Equal(t, givenAvgStock, kpi.AvgMonthlyStock)
	assert.InDelta(t, 3.0, kpi.RotationRatio, 0.1)
	assert.Len(t, kpi.TopRotatingProducts, 1)
	assert.Len(t, kpi.SlowRotatingProducts, 1)
}

func TestNewInventoryRotationKPI_WithNilOrganizationID_ReturnsError(t *testing.T) {
	kpi, err := domain.NewInventoryRotationKPI(
		uuid.Nil,
		time.Now(),
		10000.0,
		5000.0,
		nil,
		nil,
	)

	assert.Error(t, err)
	assert.Nil(t, kpi)
	assert.Contains(t, err.Error(), "organization_id is required")
}

func TestNewInventoryRotationKPI_WithZeroSnapshotDate_ReturnsError(t *testing.T) {
	kpi, err := domain.NewInventoryRotationKPI(
		uuid.New(),
		time.Time{},
		10000.0,
		5000.0,
		nil,
		nil,
	)

	assert.Error(t, err)
	assert.Nil(t, kpi)
	assert.Contains(t, err.Error(), "snapshot_date is required")
}

func TestNewInventoryRotationKPI_WithNegativeSales_ReturnsError(t *testing.T) {
	kpi, err := domain.NewInventoryRotationKPI(
		uuid.New(),
		time.Now(),
		-10000.0,
		5000.0,
		nil,
		nil,
	)

	assert.Error(t, err)
	assert.Nil(t, kpi)
	assert.Contains(t, err.Error(), "sales_last_30_days cannot be negative")
}

func TestNewInventoryRotationKPI_WithNegativeAvgStock_ReturnsError(t *testing.T) {
	kpi, err := domain.NewInventoryRotationKPI(
		uuid.New(),
		time.Now(),
		10000.0,
		-5000.0,
		nil,
		nil,
	)

	assert.Error(t, err)
	assert.Nil(t, kpi)
	assert.Contains(t, err.Error(), "avg_monthly_stock cannot be negative")
}

func TestNewInventoryRotationKPI_WithNilProducts_InitializesEmptySlices(t *testing.T) {
	kpi, err := domain.NewInventoryRotationKPI(
		uuid.New(),
		time.Now(),
		10000.0,
		5000.0,
		nil,
		nil,
	)

	assert.NoError(t, err)
	assert.NotNil(t, kpi)
	assert.NotNil(t, kpi.TopRotatingProducts)
	assert.NotNil(t, kpi.SlowRotatingProducts)
	assert.Len(t, kpi.TopRotatingProducts, 0)
	assert.Len(t, kpi.SlowRotatingProducts, 0)
}

func TestNewInventoryRotationKPI_WithZeroAvgStock_CalculatesZeroRatio(t *testing.T) {
	kpi, err := domain.NewInventoryRotationKPI(
		uuid.New(),
		time.Now(),
		10000.0,
		0,
		nil,
		nil,
	)

	assert.NoError(t, err)
	assert.NotNil(t, kpi)
	assert.Equal(t, 0.0, kpi.RotationRatio)
}

func TestCalculateProductRotation_WithValidValues_CalculatesCorrectly(t *testing.T) {
	sales := 5000.0
	avgStock := 1000.0

	rotation := domain.CalculateProductRotation(sales, avgStock)

	assert.Equal(t, 5.0, rotation)
}

func TestCalculateProductRotation_WithZeroAvgStock_ReturnsZero(t *testing.T) {
	sales := 5000.0
	avgStock := 0.0

	rotation := domain.CalculateProductRotation(sales, avgStock)

	assert.Equal(t, 0.0, rotation)
}

func TestNewRotatingProduct_CalculatesRotationRatio(t *testing.T) {
	product := domain.NewRotatingProduct(
		uuid.New(),
		"SKU-001",
		"Test Product",
		5000.0,
		1000.0,
	)

	assert.Equal(t, 5.0, product.RotationRatio)
}
