package domain

import (
	"math"
	"time"

	"github.com/google/uuid"
)

type Buffer struct {
	ID                 uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ProductID          uuid.UUID  `json:"product_id" gorm:"type:uuid;not null;uniqueIndex:uq_buffer_product;index:idx_buffers_product"`
	OrganizationID     uuid.UUID  `json:"organization_id" gorm:"type:uuid;not null;uniqueIndex:uq_buffer_product;index:idx_buffers_product,idx_buffers_org"`
	BufferProfileID    uuid.UUID  `json:"buffer_profile_id" gorm:"type:uuid;not null"`
	CPD                float64    `json:"cpd" gorm:"type:decimal(15,2);not null"`
	LTD                int        `json:"ltd" gorm:"not null"`
	RedBase            float64    `json:"red_base" gorm:"type:decimal(15,2);not null"`
	RedSafe            float64    `json:"red_safe" gorm:"type:decimal(15,2);not null"`
	RedZone            float64    `json:"red_zone" gorm:"type:decimal(15,2);not null"`
	YellowZone         float64    `json:"yellow_zone" gorm:"type:decimal(15,2);not null"`
	GreenZone          float64    `json:"green_zone" gorm:"type:decimal(15,2);not null"`
	TopOfRed           float64    `json:"top_of_red" gorm:"type:decimal(15,2);not null"`
	TopOfYellow        float64    `json:"top_of_yellow" gorm:"type:decimal(15,2);not null"`
	TopOfGreen         float64    `json:"top_of_green" gorm:"type:decimal(15,2);not null"`
	OnHand             float64    `json:"on_hand" gorm:"type:decimal(15,2);not null;default:0"`
	OnOrder            float64    `json:"on_order" gorm:"type:decimal(15,2);not null;default:0"`
	QualifiedDemand    float64    `json:"qualified_demand" gorm:"type:decimal(15,2);not null;default:0"`
	NetFlowPosition    float64    `json:"net_flow_position" gorm:"type:decimal(15,2);not null;default:0"`
	BufferPenetration  float64    `json:"buffer_penetration" gorm:"type:decimal(5,2);not null;default:0"`
	Zone               ZoneType   `json:"zone" gorm:"type:varchar(20);not null;default:'green';index:idx_buffers_zone"`
	AlertLevel         AlertLevel `json:"alert_level" gorm:"type:varchar(20);not null;default:'normal';index:idx_buffers_alert"`
	LastRecalculatedAt time.Time  `json:"last_recalculated_at" gorm:"not null;default:CURRENT_TIMESTAMP;index:idx_buffers_last_recalc"`
	CreatedAt          time.Time  `json:"created_at" gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt          time.Time  `json:"updated_at" gorm:"not null;default:CURRENT_TIMESTAMP"`
}

type ZoneType string

const (
	ZoneGreen    ZoneType = "green"
	ZoneYellow   ZoneType = "yellow"
	ZoneRed      ZoneType = "red"
	ZoneBelowRed ZoneType = "below_red"
	ZoneAll      ZoneType = "all"
)

type AlertLevel string

const (
	AlertNormal    AlertLevel = "normal"
	AlertMonitor   AlertLevel = "monitor"
	AlertReplenish AlertLevel = "replenish"
	AlertCritical  AlertLevel = "critical"
)

func (Buffer) TableName() string {
	return "buffers"
}

func CalculateBufferZones(
	cpd float64,
	ltd int,
	leadTimeFactor float64,
	variabilityFactor float64,
	moq int,
	orderFrequency int,
) (redBase, redSafe, redZone, yellowZone, greenZone float64) {
	redBase = float64(ltd) * cpd * leadTimeFactor
	redSafe = redBase * variabilityFactor
	redZone = redBase + redSafe
	yellowZone = cpd * float64(ltd)

	option1 := float64(moq)
	option2 := float64(orderFrequency) * cpd
	option3 := float64(ltd) * cpd * leadTimeFactor

	greenZone = math.Max(option1, math.Max(option2, option3))

	return
}

func (b *Buffer) CalculateNFP() {
	b.NetFlowPosition = b.OnHand + b.OnOrder - b.QualifiedDemand
}

func (b *Buffer) DetermineZone() {
	b.CalculateNFP()

	b.TopOfRed = b.RedZone
	b.TopOfYellow = b.TopOfRed + b.YellowZone
	b.TopOfGreen = b.TopOfYellow + b.GreenZone

	switch {
	case b.NetFlowPosition >= b.TopOfYellow:
		b.Zone = ZoneGreen
		b.AlertLevel = AlertNormal
	case b.NetFlowPosition >= b.TopOfRed:
		b.Zone = ZoneYellow
		b.AlertLevel = AlertMonitor
	case b.NetFlowPosition > 0:
		b.Zone = ZoneRed
		b.AlertLevel = AlertReplenish
	default:
		b.Zone = ZoneBelowRed
		b.AlertLevel = AlertCritical
	}

	if b.TopOfGreen > 0 {
		b.BufferPenetration = (b.NetFlowPosition / b.TopOfGreen) * 100
	}
}

func ApplyAdjustedCPD(baseCPD float64, activeFADs []DemandAdjustment) float64 {
	adjustedCPD := baseCPD

	for _, fad := range activeFADs {
		adjustedCPD *= fad.Factor
	}

	return math.Ceil(adjustedCPD)
}

func (b *Buffer) Validate() error {
	if b.ProductID == uuid.Nil {
		return NewValidationError("product_id is required")
	}
	if b.OrganizationID == uuid.Nil {
		return NewValidationError("organization_id is required")
	}
	if b.BufferProfileID == uuid.Nil {
		return NewValidationError("buffer_profile_id is required")
	}
	if b.CPD < 0 {
		return NewValidationError("cpd must be non-negative")
	}
	if b.LTD <= 0 {
		return NewValidationError("ltd must be greater than 0")
	}
	if b.RedZone < 0 || b.YellowZone < 0 || b.GreenZone < 0 {
		return NewValidationError("buffer zones must be non-negative")
	}
	return nil
}
