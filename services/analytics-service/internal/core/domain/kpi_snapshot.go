package domain

import (
	"time"

	"github.com/google/uuid"
)

type KPISnapshot struct {
	ID                 uuid.UUID
	OrganizationID     uuid.UUID
	SnapshotDate       time.Time
	InventoryTurnover  float64
	StockoutRate       float64
	ServiceLevel       float64
	ExcessInventoryPct float64
	BufferScoreGreen   float64
	BufferScoreYellow  float64
	BufferScoreRed     float64
	TotalInventoryValue float64
	CreatedAt          time.Time
}

func NewKPISnapshot(
	organizationID uuid.UUID,
	snapshotDate time.Time,
	inventoryTurnover float64,
	stockoutRate float64,
	serviceLevel float64,
	excessInventoryPct float64,
	bufferScoreGreen float64,
	bufferScoreYellow float64,
	bufferScoreRed float64,
	totalInventoryValue float64,
) (*KPISnapshot, error) {
	if organizationID == uuid.Nil {
		return nil, NewValidationError("organization_id is required")
	}
	if snapshotDate.IsZero() {
		return nil, NewValidationError("snapshot_date is required")
	}
	if inventoryTurnover < 0 {
		return nil, NewValidationError("inventory_turnover cannot be negative")
	}
	if stockoutRate < 0 || stockoutRate > 100 {
		return nil, NewValidationError("stockout_rate must be between 0 and 100")
	}
	if serviceLevel < 0 || serviceLevel > 100 {
		return nil, NewValidationError("service_level must be between 0 and 100")
	}
	if excessInventoryPct < 0 {
		return nil, NewValidationError("excess_inventory_pct cannot be negative")
	}
	if bufferScoreGreen < 0 || bufferScoreGreen > 100 {
		return nil, NewValidationError("buffer_score_green must be between 0 and 100")
	}
	if bufferScoreYellow < 0 || bufferScoreYellow > 100 {
		return nil, NewValidationError("buffer_score_yellow must be between 0 and 100")
	}
	if bufferScoreRed < 0 || bufferScoreRed > 100 {
		return nil, NewValidationError("buffer_score_red must be between 0 and 100")
	}
	if totalInventoryValue < 0 {
		return nil, NewValidationError("total_inventory_value cannot be negative")
	}

	return &KPISnapshot{
		ID:                  uuid.New(),
		OrganizationID:      organizationID,
		SnapshotDate:        snapshotDate,
		InventoryTurnover:   inventoryTurnover,
		StockoutRate:        stockoutRate,
		ServiceLevel:        serviceLevel,
		ExcessInventoryPct:  excessInventoryPct,
		BufferScoreGreen:    bufferScoreGreen,
		BufferScoreYellow:   bufferScoreYellow,
		BufferScoreRed:      bufferScoreRed,
		TotalInventoryValue: totalInventoryValue,
		CreatedAt:           time.Now().UTC(),
	}, nil
}

func ValidateBufferScoreSum(green, yellow, red float64) error {
	total := green + yellow + red
	if total < 99.9 || total > 100.1 {
		return NewValidationError("buffer scores must sum to approximately 100")
	}
	return nil
}
