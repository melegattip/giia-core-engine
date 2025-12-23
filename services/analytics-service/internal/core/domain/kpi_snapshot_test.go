package domain_test

import (
	"testing"
	"time"

	"github.com/giia/giia-core-engine/services/analytics-service/internal/core/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewKPISnapshot_WithValidData_CreatesSnapshot(t *testing.T) {
	givenOrgID := uuid.New()
	givenDate := time.Now()

	snapshot, err := domain.NewKPISnapshot(
		givenOrgID,
		givenDate,
		3.5,  // InventoryTurnover
		2.5,  // StockoutRate
		95.0, // ServiceLevel
		10.0, // ExcessInventoryPct
		70.0, // BufferScoreGreen
		20.0, // BufferScoreYellow
		10.0, // BufferScoreRed
		250000.0, // TotalInventoryValue
	)

	assert.NoError(t, err)
	assert.NotNil(t, snapshot)
	assert.Equal(t, givenOrgID, snapshot.OrganizationID)
	assert.Equal(t, 3.5, snapshot.InventoryTurnover)
	assert.Equal(t, 95.0, snapshot.ServiceLevel)
}

func TestNewKPISnapshot_WithNilOrganizationID_ReturnsError(t *testing.T) {
	snapshot, err := domain.NewKPISnapshot(
		uuid.Nil,
		time.Now(),
		3.5, 2.5, 95.0, 10.0, 70.0, 20.0, 10.0, 250000.0,
	)

	assert.Error(t, err)
	assert.Nil(t, snapshot)
	assert.Contains(t, err.Error(), "organization_id is required")
}

func TestNewKPISnapshot_WithZeroSnapshotDate_ReturnsError(t *testing.T) {
	snapshot, err := domain.NewKPISnapshot(
		uuid.New(),
		time.Time{},
		3.5, 2.5, 95.0, 10.0, 70.0, 20.0, 10.0, 250000.0,
	)

	assert.Error(t, err)
	assert.Nil(t, snapshot)
	assert.Contains(t, err.Error(), "snapshot_date is required")
}

func TestNewKPISnapshot_WithNegativeInventoryTurnover_ReturnsError(t *testing.T) {
	snapshot, err := domain.NewKPISnapshot(
		uuid.New(),
		time.Now(),
		-3.5, 2.5, 95.0, 10.0, 70.0, 20.0, 10.0, 250000.0,
	)

	assert.Error(t, err)
	assert.Nil(t, snapshot)
	assert.Contains(t, err.Error(), "inventory_turnover cannot be negative")
}

func TestNewKPISnapshot_WithStockoutRateAbove100_ReturnsError(t *testing.T) {
	snapshot, err := domain.NewKPISnapshot(
		uuid.New(),
		time.Now(),
		3.5, 105.0, 95.0, 10.0, 70.0, 20.0, 10.0, 250000.0,
	)

	assert.Error(t, err)
	assert.Nil(t, snapshot)
	assert.Contains(t, err.Error(), "stockout_rate must be between 0 and 100")
}

func TestNewKPISnapshot_WithServiceLevelAbove100_ReturnsError(t *testing.T) {
	snapshot, err := domain.NewKPISnapshot(
		uuid.New(),
		time.Now(),
		3.5, 2.5, 105.0, 10.0, 70.0, 20.0, 10.0, 250000.0,
	)

	assert.Error(t, err)
	assert.Nil(t, snapshot)
	assert.Contains(t, err.Error(), "service_level must be between 0 and 100")
}

func TestNewKPISnapshot_WithBufferScoreGreenAbove100_ReturnsError(t *testing.T) {
	snapshot, err := domain.NewKPISnapshot(
		uuid.New(),
		time.Now(),
		3.5, 2.5, 95.0, 10.0, 105.0, 20.0, 10.0, 250000.0,
	)

	assert.Error(t, err)
	assert.Nil(t, snapshot)
	assert.Contains(t, err.Error(), "buffer_score_green must be between 0 and 100")
}

func TestValidateBufferScoreSum_WithValid100Percent_NoError(t *testing.T) {
	err := domain.ValidateBufferScoreSum(70.0, 20.0, 10.0)

	assert.NoError(t, err)
}

func TestValidateBufferScoreSum_WithInvalidSum_ReturnsError(t *testing.T) {
	err := domain.ValidateBufferScoreSum(70.0, 20.0, 5.0)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "buffer scores must sum to approximately 100")
}
