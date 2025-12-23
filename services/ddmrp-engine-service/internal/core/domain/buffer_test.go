package domain

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCalculateBufferZones_Success(t *testing.T) {
	givenCPD := 100.0
	givenLTD := 30
	givenLeadTimeFactor := 0.5
	givenVariabilityFactor := 0.5
	givenMOQ := 500
	givenOrderFrequency := 7

	redBase, redSafe, redZone, yellowZone, greenZone := CalculateBufferZones(
		givenCPD,
		givenLTD,
		givenLeadTimeFactor,
		givenVariabilityFactor,
		givenMOQ,
		givenOrderFrequency,
	)

	expectedRedBase := 30.0 * 100.0 * 0.5
	expectedRedSafe := expectedRedBase * 0.5
	expectedRedZone := expectedRedBase + expectedRedSafe
	expectedYellowZone := 100.0 * 30.0
	expectedGreenZone := 1500.0

	assert.Equal(t, expectedRedBase, redBase)
	assert.Equal(t, expectedRedSafe, redSafe)
	assert.Equal(t, expectedRedZone, redZone)
	assert.Equal(t, expectedYellowZone, yellowZone)
	assert.Equal(t, expectedGreenZone, greenZone)
}

func TestCalculateBufferZones_GreenZoneIsMax(t *testing.T) {
	givenCPD := 50.0
	givenLTD := 20
	givenLeadTimeFactor := 0.6
	givenVariabilityFactor := 0.4
	givenMOQ := 100
	givenOrderFrequency := 10

	_, _, _, _, greenZone := CalculateBufferZones(
		givenCPD,
		givenLTD,
		givenLeadTimeFactor,
		givenVariabilityFactor,
		givenMOQ,
		givenOrderFrequency,
	)

	option3 := float64(givenLTD) * givenCPD * givenLeadTimeFactor
	expectedMax := option3

	assert.Equal(t, expectedMax, greenZone)
}

func TestApplyAdjustedCPD_SingleFAD(t *testing.T) {
	givenBaseCPD := 100.0
	givenFADs := []DemandAdjustment{
		{Factor: 1.5},
	}

	adjustedCPD := ApplyAdjustedCPD(givenBaseCPD, givenFADs)

	expectedCPD := 150.0

	assert.Equal(t, expectedCPD, adjustedCPD)
}

func TestApplyAdjustedCPD_MultipleFADs(t *testing.T) {
	givenBaseCPD := 100.0
	givenFADs := []DemandAdjustment{
		{Factor: 1.5},
		{Factor: 1.2},
	}

	adjustedCPD := ApplyAdjustedCPD(givenBaseCPD, givenFADs)

	expectedCPD := 180.0

	assert.Equal(t, expectedCPD, adjustedCPD)
}

func TestApplyAdjustedCPD_NoFADs(t *testing.T) {
	givenBaseCPD := 100.5
	givenFADs := []DemandAdjustment{}

	adjustedCPD := ApplyAdjustedCPD(givenBaseCPD, givenFADs)

	expectedCPD := 101.0

	assert.Equal(t, expectedCPD, adjustedCPD)
}

func TestBuffer_CalculateNFP(t *testing.T) {
	givenBuffer := &Buffer{
		OnHand:          150.0,
		OnOrder:         200.0,
		QualifiedDemand: 100.0,
	}

	givenBuffer.CalculateNFP()

	expectedNFP := 250.0

	assert.Equal(t, expectedNFP, givenBuffer.NetFlowPosition)
}

func TestBuffer_DetermineZone_Green(t *testing.T) {
	givenBuffer := &Buffer{
		RedZone:         1000.0,
		YellowZone:      2000.0,
		GreenZone:       1000.0,
		OnHand:          3500.0,
		OnOrder:         0.0,
		QualifiedDemand: 0.0,
	}

	givenBuffer.DetermineZone()

	assert.Equal(t, ZoneGreen, givenBuffer.Zone)
	assert.Equal(t, AlertNormal, givenBuffer.AlertLevel)
	assert.Equal(t, 1000.0, givenBuffer.TopOfRed)
	assert.Equal(t, 3000.0, givenBuffer.TopOfYellow)
	assert.Equal(t, 4000.0, givenBuffer.TopOfGreen)
}

func TestBuffer_DetermineZone_Yellow(t *testing.T) {
	givenBuffer := &Buffer{
		RedZone:         1000.0,
		YellowZone:      2000.0,
		GreenZone:       1000.0,
		OnHand:          2500.0,
		OnOrder:         0.0,
		QualifiedDemand: 0.0,
	}

	givenBuffer.DetermineZone()

	assert.Equal(t, ZoneYellow, givenBuffer.Zone)
	assert.Equal(t, AlertMonitor, givenBuffer.AlertLevel)
}

func TestBuffer_DetermineZone_Red(t *testing.T) {
	givenBuffer := &Buffer{
		RedZone:         1000.0,
		YellowZone:      2000.0,
		GreenZone:       1000.0,
		OnHand:          500.0,
		OnOrder:         0.0,
		QualifiedDemand: 0.0,
	}

	givenBuffer.DetermineZone()

	assert.Equal(t, ZoneRed, givenBuffer.Zone)
	assert.Equal(t, AlertReplenish, givenBuffer.AlertLevel)
}

func TestBuffer_DetermineZone_BelowRed(t *testing.T) {
	givenBuffer := &Buffer{
		RedZone:         1000.0,
		YellowZone:      2000.0,
		GreenZone:       1000.0,
		OnHand:          0.0,
		OnOrder:         0.0,
		QualifiedDemand: 500.0,
	}

	givenBuffer.DetermineZone()

	assert.Equal(t, ZoneBelowRed, givenBuffer.Zone)
	assert.Equal(t, AlertCritical, givenBuffer.AlertLevel)
}

func TestBuffer_Validate_Success(t *testing.T) {
	givenBuffer := &Buffer{
		ProductID:       newUUID(),
		OrganizationID:  newUUID(),
		BufferProfileID: newUUID(),
		CPD:             100.0,
		LTD:             30,
		RedZone:         500.0,
		YellowZone:      1000.0,
		GreenZone:       500.0,
	}

	err := givenBuffer.Validate()

	assert.NoError(t, err)
}

func TestBuffer_Validate_MissingProductID(t *testing.T) {
	givenBuffer := &Buffer{
		OrganizationID:  newUUID(),
		BufferProfileID: newUUID(),
		CPD:             100.0,
		LTD:             30,
	}

	err := givenBuffer.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "product_id is required")
}

func newUUID() uuid.UUID {
	return uuid.New()
}
