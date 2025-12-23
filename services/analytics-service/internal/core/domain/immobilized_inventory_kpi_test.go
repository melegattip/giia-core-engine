package domain_test

import (
	"testing"
	"time"

	"github.com/giia/giia-core-engine/services/analytics-service/internal/core/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewImmobilizedInventoryKPI_WithValidData_CreatesKPI(t *testing.T) {
	givenOrgID := uuid.New()
	givenDate := time.Now()
	givenThresholdYears := 2
	givenImmobilizedCount := 25
	givenImmobilizedValue := 15000.0
	givenTotalStockValue := 100000.0

	kpi, err := domain.NewImmobilizedInventoryKPI(
		givenOrgID,
		givenDate,
		givenThresholdYears,
		givenImmobilizedCount,
		givenImmobilizedValue,
		givenTotalStockValue,
	)

	assert.NoError(t, err)
	assert.NotNil(t, kpi)
	assert.Equal(t, givenOrgID, kpi.OrganizationID)
	assert.Equal(t, givenThresholdYears, kpi.ThresholdYears)
	assert.Equal(t, givenImmobilizedCount, kpi.ImmobilizedCount)
	assert.Equal(t, givenImmobilizedValue, kpi.ImmobilizedValue)
	assert.Equal(t, givenTotalStockValue, kpi.TotalStockValue)
	assert.InDelta(t, 15.0, kpi.ImmobilizedPercentage, 0.1)
}

func TestNewImmobilizedInventoryKPI_WithNilOrganizationID_ReturnsError(t *testing.T) {
	kpi, err := domain.NewImmobilizedInventoryKPI(
		uuid.Nil,
		time.Now(),
		2,
		10,
		5000.0,
		100000.0,
	)

	assert.Error(t, err)
	assert.Nil(t, kpi)
	assert.Contains(t, err.Error(), "organization_id is required")
}

func TestNewImmobilizedInventoryKPI_WithZeroSnapshotDate_ReturnsError(t *testing.T) {
	kpi, err := domain.NewImmobilizedInventoryKPI(
		uuid.New(),
		time.Time{},
		2,
		10,
		5000.0,
		100000.0,
	)

	assert.Error(t, err)
	assert.Nil(t, kpi)
	assert.Contains(t, err.Error(), "snapshot_date is required")
}

func TestNewImmobilizedInventoryKPI_WithZeroThresholdYears_ReturnsError(t *testing.T) {
	kpi, err := domain.NewImmobilizedInventoryKPI(
		uuid.New(),
		time.Now(),
		0,
		10,
		5000.0,
		100000.0,
	)

	assert.Error(t, err)
	assert.Nil(t, kpi)
	assert.Contains(t, err.Error(), "threshold_years must be positive")
}

func TestNewImmobilizedInventoryKPI_WithNegativeImmobilizedCount_ReturnsError(t *testing.T) {
	kpi, err := domain.NewImmobilizedInventoryKPI(
		uuid.New(),
		time.Now(),
		2,
		-10,
		5000.0,
		100000.0,
	)

	assert.Error(t, err)
	assert.Nil(t, kpi)
	assert.Contains(t, err.Error(), "immobilized_count cannot be negative")
}

func TestNewImmobilizedInventoryKPI_WithNegativeImmobilizedValue_ReturnsError(t *testing.T) {
	kpi, err := domain.NewImmobilizedInventoryKPI(
		uuid.New(),
		time.Now(),
		2,
		10,
		-5000.0,
		100000.0,
	)

	assert.Error(t, err)
	assert.Nil(t, kpi)
	assert.Contains(t, err.Error(), "immobilized_value cannot be negative")
}

func TestNewImmobilizedInventoryKPI_WithNegativeTotalStockValue_ReturnsError(t *testing.T) {
	kpi, err := domain.NewImmobilizedInventoryKPI(
		uuid.New(),
		time.Now(),
		2,
		10,
		5000.0,
		-100000.0,
	)

	assert.Error(t, err)
	assert.Nil(t, kpi)
	assert.Contains(t, err.Error(), "total_stock_value cannot be negative")
}

func TestNewImmobilizedInventoryKPI_WithZeroTotalStockValue_CalculatesZeroPercentage(t *testing.T) {
	kpi, err := domain.NewImmobilizedInventoryKPI(
		uuid.New(),
		time.Now(),
		2,
		10,
		5000.0,
		0,
	)

	assert.NoError(t, err)
	assert.NotNil(t, kpi)
	assert.Equal(t, 0.0, kpi.ImmobilizedPercentage)
}

func TestCalculateYearsInStock_WithValidDates_CalculatesCorrectly(t *testing.T) {
	purchaseDate := time.Now().AddDate(-2, -6, 0)
	currentDate := time.Now()

	years := domain.CalculateYearsInStock(purchaseDate, currentDate)

	assert.InDelta(t, 2.5, years, 0.1)
}

func TestIsImmobilized_WithOldProduct_ReturnsTrue(t *testing.T) {
	purchaseDate := time.Now().AddDate(-3, 0, 0)
	currentDate := time.Now()
	thresholdYears := 2

	isImmobilized := domain.IsImmobilized(purchaseDate, currentDate, thresholdYears)

	assert.True(t, isImmobilized)
}

func TestIsImmobilized_WithRecentProduct_ReturnsFalse(t *testing.T) {
	purchaseDate := time.Now().AddDate(-1, 0, 0)
	currentDate := time.Now()
	thresholdYears := 2

	isImmobilized := domain.IsImmobilized(purchaseDate, currentDate, thresholdYears)

	assert.False(t, isImmobilized)
}
