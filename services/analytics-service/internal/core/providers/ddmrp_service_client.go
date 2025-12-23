package providers

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type BufferHistory struct {
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
	HasAdjustments    bool
}

type BufferZoneDistribution struct {
	OrganizationID uuid.UUID
	Date           time.Time
	GreenCount     int
	YellowCount    int
	RedCount       int
	TotalProducts  int
}

type DDMRPServiceClient interface {
	GetBufferHistory(ctx context.Context, organizationID, productID uuid.UUID, date time.Time) (*BufferHistory, error)
	ListBufferHistory(ctx context.Context, organizationID uuid.UUID, startDate, endDate time.Time) ([]*BufferHistory, error)
	GetBufferZoneDistribution(ctx context.Context, organizationID uuid.UUID, date time.Time) (*BufferZoneDistribution, error)
}
