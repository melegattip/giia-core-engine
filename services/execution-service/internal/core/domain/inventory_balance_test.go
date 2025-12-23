package domain_test

import (
	"testing"

	"github.com/giia/giia-core-engine/services/execution-service/internal/core/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewInventoryBalance_WithValidData_ReturnsInventoryBalance(t *testing.T) {
	givenOrgID := uuid.New()
	givenProductID := uuid.New()
	givenLocationID := uuid.New()

	balance, err := domain.NewInventoryBalance(
		givenOrgID,
		givenProductID,
		givenLocationID,
	)

	assert.NoError(t, err)
	assert.NotNil(t, balance)
	assert.Equal(t, givenOrgID, balance.OrganizationID)
	assert.Equal(t, givenProductID, balance.ProductID)
	assert.Equal(t, givenLocationID, balance.LocationID)
	assert.Equal(t, float64(0), balance.OnHand)
	assert.Equal(t, float64(0), balance.Reserved)
	assert.Equal(t, float64(0), balance.Available)
}

func TestNewInventoryBalance_WithNilOrganizationID_ReturnsError(t *testing.T) {
	balance, err := domain.NewInventoryBalance(
		uuid.Nil,
		uuid.New(),
		uuid.New(),
	)

	assert.Error(t, err)
	assert.Nil(t, balance)
	assert.Contains(t, err.Error(), "organization_id is required")
}

func TestNewInventoryBalance_WithNilProductID_ReturnsError(t *testing.T) {
	balance, err := domain.NewInventoryBalance(
		uuid.New(),
		uuid.Nil,
		uuid.New(),
	)

	assert.Error(t, err)
	assert.Nil(t, balance)
	assert.Contains(t, err.Error(), "product_id is required")
}

func TestNewInventoryBalance_WithNilLocationID_ReturnsError(t *testing.T) {
	balance, err := domain.NewInventoryBalance(
		uuid.New(),
		uuid.New(),
		uuid.Nil,
	)

	assert.Error(t, err)
	assert.Nil(t, balance)
	assert.Contains(t, err.Error(), "location_id is required")
}

func TestInventoryBalance_UpdateOnHand_WithPositiveQuantity_IncreasesOnHand(t *testing.T) {
	givenBalance, _ := domain.NewInventoryBalance(
		uuid.New(),
		uuid.New(),
		uuid.New(),
	)
	givenQuantity := 100.0

	givenBalance.UpdateOnHand(givenQuantity)

	assert.Equal(t, givenQuantity, givenBalance.OnHand)
	assert.Equal(t, givenQuantity, givenBalance.Available)
}

func TestInventoryBalance_UpdateOnHand_WithNegativeQuantity_DecreasesOnHand(t *testing.T) {
	givenBalance, _ := domain.NewInventoryBalance(
		uuid.New(),
		uuid.New(),
		uuid.New(),
	)
	givenBalance.OnHand = 100.0
	givenQuantity := -50.0

	givenBalance.UpdateOnHand(givenQuantity)

	assert.Equal(t, float64(50), givenBalance.OnHand)
	assert.Equal(t, float64(50), givenBalance.Available)
}

func TestInventoryBalance_UpdateReserved_WithValidQuantity_UpdatesReserved(t *testing.T) {
	givenBalance, _ := domain.NewInventoryBalance(
		uuid.New(),
		uuid.New(),
		uuid.New(),
	)
	givenBalance.OnHand = 100.0
	givenReservedQuantity := 30.0

	err := givenBalance.UpdateReserved(givenReservedQuantity)

	assert.NoError(t, err)
	assert.Equal(t, givenReservedQuantity, givenBalance.Reserved)
	assert.Equal(t, float64(70), givenBalance.Available)
}

func TestInventoryBalance_UpdateReserved_WhenExceedingOnHand_ReturnsError(t *testing.T) {
	givenBalance, _ := domain.NewInventoryBalance(
		uuid.New(),
		uuid.New(),
		uuid.New(),
	)
	givenBalance.OnHand = 100.0
	givenReservedQuantity := 150.0

	err := givenBalance.UpdateReserved(givenReservedQuantity)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot reserve more than on-hand quantity")
}

func TestInventoryBalance_UpdateReserved_WhenResultingInNegative_ReturnsError(t *testing.T) {
	givenBalance, _ := domain.NewInventoryBalance(
		uuid.New(),
		uuid.New(),
		uuid.New(),
	)
	givenBalance.OnHand = 100.0
	givenBalance.Reserved = 50.0
	givenReservedQuantity := -100.0

	err := givenBalance.UpdateReserved(givenReservedQuantity)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "reserved quantity cannot be negative")
}

func TestInventoryBalance_CalculateAvailable_WithNoReserved_AvailableEqualsOnHand(t *testing.T) {
	givenBalance, _ := domain.NewInventoryBalance(
		uuid.New(),
		uuid.New(),
		uuid.New(),
	)
	givenBalance.OnHand = 100.0
	givenBalance.Reserved = 0

	givenBalance.CalculateAvailable()

	assert.Equal(t, float64(100), givenBalance.Available)
}

func TestInventoryBalance_CalculateAvailable_WithReserved_CalculatesCorrectAvailable(t *testing.T) {
	givenBalance, _ := domain.NewInventoryBalance(
		uuid.New(),
		uuid.New(),
		uuid.New(),
	)
	givenBalance.OnHand = 100.0
	givenBalance.Reserved = 30.0

	givenBalance.CalculateAvailable()

	assert.Equal(t, float64(70), givenBalance.Available)
}

func TestInventoryBalance_CalculateAvailable_WhenReservedExceedsOnHand_SetsAvailableToZero(t *testing.T) {
	givenBalance, _ := domain.NewInventoryBalance(
		uuid.New(),
		uuid.New(),
		uuid.New(),
	)
	givenBalance.OnHand = 50.0
	givenBalance.Reserved = 70.0

	givenBalance.CalculateAvailable()

	assert.Equal(t, float64(0), givenBalance.Available)
}