package domain

import (
	"time"

	"github.com/google/uuid"
)

type BufferAnalytics struct {
	ID                uuid.UUID
	ProductID         uuid.UUID
	OrganizationID    uuid.UUID
	Date              time.Time
	CPD               float64
	RedZone           float64
	RedBase           float64
	RedSafe           float64
	YellowZone        float64
	GreenZone         float64
	LTD               int
	LeadTimeFactor    float64
	VariabilityFactor float64
	MOQ               int
	OrderFrequency    int
	OptimalOrderFreq  float64
	SafetyDays        float64
	AvgOpenOrders     float64
	HasAdjustments    bool
	CreatedAt         time.Time
}

func NewBufferAnalytics(
	productID uuid.UUID,
	organizationID uuid.UUID,
	date time.Time,
	cpd float64,
	redZone float64,
	redBase float64,
	redSafe float64,
	yellowZone float64,
	greenZone float64,
	ltd int,
	leadTimeFactor float64,
	variabilityFactor float64,
	moq int,
	orderFrequency int,
	hasAdjustments bool,
) (*BufferAnalytics, error) {
	if productID == uuid.Nil {
		return nil, NewValidationError("product_id is required")
	}
	if organizationID == uuid.Nil {
		return nil, NewValidationError("organization_id is required")
	}
	if date.IsZero() {
		return nil, NewValidationError("date is required")
	}
	if cpd < 0 {
		return nil, NewValidationError("cpd cannot be negative")
	}
	if redZone < 0 {
		return nil, NewValidationError("red_zone cannot be negative")
	}
	if yellowZone < 0 {
		return nil, NewValidationError("yellow_zone cannot be negative")
	}
	if greenZone < 0 {
		return nil, NewValidationError("green_zone cannot be negative")
	}
	if ltd < 0 {
		return nil, NewValidationError("ltd cannot be negative")
	}

	optimalOrderFreq := 0.0
	if cpd > 0 {
		optimalOrderFreq = greenZone / cpd
	}

	safetyDays := 0.0
	if cpd > 0 {
		safetyDays = redZone / cpd
	}

	avgOpenOrders := 0.0
	if greenZone > 0 {
		avgOpenOrders = yellowZone / greenZone
	}

	return &BufferAnalytics{
		ID:                uuid.New(),
		ProductID:         productID,
		OrganizationID:    organizationID,
		Date:              date,
		CPD:               cpd,
		RedZone:           redZone,
		RedBase:           redBase,
		RedSafe:           redSafe,
		YellowZone:        yellowZone,
		GreenZone:         greenZone,
		LTD:               ltd,
		LeadTimeFactor:    leadTimeFactor,
		VariabilityFactor: variabilityFactor,
		MOQ:               moq,
		OrderFrequency:    orderFrequency,
		OptimalOrderFreq:  optimalOrderFreq,
		SafetyDays:        safetyDays,
		AvgOpenOrders:     avgOpenOrders,
		HasAdjustments:    hasAdjustments,
		CreatedAt:         time.Now().UTC(),
	}, nil
}

func (ba *BufferAnalytics) CalculateOptimalOrderFrequency() float64 {
	if ba.CPD > 0 {
		return ba.GreenZone / ba.CPD
	}
	return 0.0
}

func (ba *BufferAnalytics) CalculateSafetyDays() float64 {
	if ba.CPD > 0 {
		return ba.RedZone / ba.CPD
	}
	return 0.0
}

func (ba *BufferAnalytics) CalculateAvgOpenOrders() float64 {
	if ba.GreenZone > 0 {
		return ba.YellowZone / ba.GreenZone
	}
	return 0.0
}
