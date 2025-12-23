package domain_test

import (
	"testing"
	"time"

	"github.com/giia/giia-core-engine/services/analytics-service/internal/core/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewBufferAnalytics_WithValidData_CreatesAnalytics(t *testing.T) {
	givenProductID := uuid.New()
	givenOrgID := uuid.New()
	givenDate := time.Now()

	analytics, err := domain.NewBufferAnalytics(
		givenProductID,
		givenOrgID,
		givenDate,
		10.0,  // CPD
		50.0,  // RedZone
		30.0,  // RedBase
		20.0,  // RedSafe
		100.0, // YellowZone
		150.0, // GreenZone
		15,    // LTD
		0.5,   // LeadTimeFactor
		0.5,   // VariabilityFactor
		100,   // MOQ
		30,    // OrderFrequency
		false, // HasAdjustments
	)

	assert.NoError(t, err)
	assert.NotNil(t, analytics)
	assert.Equal(t, givenProductID, analytics.ProductID)
	assert.Equal(t, givenOrgID, analytics.OrganizationID)
	assert.InDelta(t, 15.0, analytics.OptimalOrderFreq, 0.1) // 150 / 10
	assert.InDelta(t, 5.0, analytics.SafetyDays, 0.1)        // 50 / 10
	assert.InDelta(t, 0.67, analytics.AvgOpenOrders, 0.1)    // 100 / 150
}

func TestNewBufferAnalytics_WithNilProductID_ReturnsError(t *testing.T) {
	analytics, err := domain.NewBufferAnalytics(
		uuid.Nil,
		uuid.New(),
		time.Now(),
		10.0, 50.0, 30.0, 20.0, 100.0, 150.0,
		15, 0.5, 0.5, 100, 30, false,
	)

	assert.Error(t, err)
	assert.Nil(t, analytics)
	assert.Contains(t, err.Error(), "product_id is required")
}

func TestNewBufferAnalytics_WithNilOrganizationID_ReturnsError(t *testing.T) {
	analytics, err := domain.NewBufferAnalytics(
		uuid.New(),
		uuid.Nil,
		time.Now(),
		10.0, 50.0, 30.0, 20.0, 100.0, 150.0,
		15, 0.5, 0.5, 100, 30, false,
	)

	assert.Error(t, err)
	assert.Nil(t, analytics)
	assert.Contains(t, err.Error(), "organization_id is required")
}

func TestNewBufferAnalytics_WithZeroDate_ReturnsError(t *testing.T) {
	analytics, err := domain.NewBufferAnalytics(
		uuid.New(),
		uuid.New(),
		time.Time{},
		10.0, 50.0, 30.0, 20.0, 100.0, 150.0,
		15, 0.5, 0.5, 100, 30, false,
	)

	assert.Error(t, err)
	assert.Nil(t, analytics)
	assert.Contains(t, err.Error(), "date is required")
}

func TestNewBufferAnalytics_WithNegativeCPD_ReturnsError(t *testing.T) {
	analytics, err := domain.NewBufferAnalytics(
		uuid.New(),
		uuid.New(),
		time.Now(),
		-10.0, 50.0, 30.0, 20.0, 100.0, 150.0,
		15, 0.5, 0.5, 100, 30, false,
	)

	assert.Error(t, err)
	assert.Nil(t, analytics)
	assert.Contains(t, err.Error(), "cpd cannot be negative")
}

func TestNewBufferAnalytics_WithNegativeLTD_ReturnsError(t *testing.T) {
	analytics, err := domain.NewBufferAnalytics(
		uuid.New(),
		uuid.New(),
		time.Now(),
		10.0, 50.0, 30.0, 20.0, 100.0, 150.0,
		-15, 0.5, 0.5, 100, 30, false,
	)

	assert.Error(t, err)
	assert.Nil(t, analytics)
	assert.Contains(t, err.Error(), "ltd cannot be negative")
}

func TestBufferAnalytics_CalculateOptimalOrderFrequency_WithValidData(t *testing.T) {
	analytics := &domain.BufferAnalytics{
		CPD:       10.0,
		GreenZone: 150.0,
	}

	freq := analytics.CalculateOptimalOrderFrequency()

	assert.Equal(t, 15.0, freq)
}

func TestBufferAnalytics_CalculateOptimalOrderFrequency_WithZeroCPD_ReturnsZero(t *testing.T) {
	analytics := &domain.BufferAnalytics{
		CPD:       0.0,
		GreenZone: 150.0,
	}

	freq := analytics.CalculateOptimalOrderFrequency()

	assert.Equal(t, 0.0, freq)
}

func TestBufferAnalytics_CalculateSafetyDays_WithValidData(t *testing.T) {
	analytics := &domain.BufferAnalytics{
		CPD:     10.0,
		RedZone: 50.0,
	}

	safetyDays := analytics.CalculateSafetyDays()

	assert.Equal(t, 5.0, safetyDays)
}

func TestBufferAnalytics_CalculateAvgOpenOrders_WithValidData(t *testing.T) {
	analytics := &domain.BufferAnalytics{
		YellowZone: 100.0,
		GreenZone:  150.0,
	}

	avgOrders := analytics.CalculateAvgOpenOrders()

	assert.InDelta(t, 0.67, avgOrders, 0.01)
}

func TestBufferAnalytics_CalculateAvgOpenOrders_WithZeroGreenZone_ReturnsZero(t *testing.T) {
	analytics := &domain.BufferAnalytics{
		YellowZone: 100.0,
		GreenZone:  0.0,
	}

	avgOrders := analytics.CalculateAvgOpenOrders()

	assert.Equal(t, 0.0, avgOrders)
}
