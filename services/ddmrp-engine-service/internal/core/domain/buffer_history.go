package domain

import (
	"time"

	"github.com/google/uuid"
)

type BufferHistory struct {
	ID                 uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	BufferID           uuid.UUID `json:"buffer_id" gorm:"type:uuid;not null;uniqueIndex:uq_buffer_history_date;index:idx_buffer_history_buffer"`
	ProductID          uuid.UUID `json:"product_id" gorm:"type:uuid;not null;index:idx_buffer_history_product"`
	OrganizationID     uuid.UUID `json:"organization_id" gorm:"type:uuid;not null;index:idx_buffer_history_product,idx_buffer_history_org"`
	SnapshotDate       time.Time `json:"snapshot_date" gorm:"type:date;not null;uniqueIndex:uq_buffer_history_date;index:idx_buffer_history_date"`
	CPD                float64   `json:"cpd" gorm:"type:decimal(15,2);not null"`
	DLT                int       `json:"dlt" gorm:"not null"`
	RedZone            float64   `json:"red_zone" gorm:"type:decimal(15,2);not null"`
	RedBase            float64   `json:"red_base" gorm:"type:decimal(15,2);not null"`
	RedSafe            float64   `json:"red_safe" gorm:"type:decimal(15,2);not null"`
	YellowZone         float64   `json:"yellow_zone" gorm:"type:decimal(15,2);not null"`
	GreenZone          float64   `json:"green_zone" gorm:"type:decimal(15,2);not null"`
	LeadTimeFactor     float64   `json:"lead_time_factor" gorm:"type:decimal(5,2);not null"`
	VariabilityFactor  float64   `json:"variability_factor" gorm:"type:decimal(5,2);not null"`
	MOQ                *int      `json:"moq" gorm:"null"`
	OrderFrequency     *int      `json:"order_frequency" gorm:"null"`
	HasAdjustments     bool      `json:"has_adjustments" gorm:"not null;default:false"`
	CreatedAt          time.Time `json:"created_at" gorm:"not null;default:CURRENT_TIMESTAMP"`
}

func (BufferHistory) TableName() string {
	return "buffer_history"
}

func NewBufferHistory(
	buffer *Buffer,
	leadTimeFactor, variabilityFactor float64,
	moq, orderFrequency *int,
	hasAdjustments bool,
) *BufferHistory {
	now := time.Now()
	snapshotDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	return &BufferHistory{
		ID:                uuid.New(),
		BufferID:          buffer.ID,
		ProductID:         buffer.ProductID,
		OrganizationID:    buffer.OrganizationID,
		SnapshotDate:      snapshotDate,
		CPD:               buffer.CPD,
		DLT:               buffer.LTD,
		RedZone:           buffer.RedZone,
		RedBase:           buffer.RedBase,
		RedSafe:           buffer.RedSafe,
		YellowZone:        buffer.YellowZone,
		GreenZone:         buffer.GreenZone,
		LeadTimeFactor:    leadTimeFactor,
		VariabilityFactor: variabilityFactor,
		MOQ:               moq,
		OrderFrequency:    orderFrequency,
		HasAdjustments:    hasAdjustments,
		CreatedAt:         time.Now(),
	}
}
