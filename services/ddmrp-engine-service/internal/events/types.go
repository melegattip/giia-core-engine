// Package events provides event types for the DDMRP Engine Service.
package events

import (
	"time"
)

// Event type constants for DDMRP.
const (
	TypeBufferCreated        = "buffer.created"
	TypeBufferUpdated        = "buffer.updated"
	TypeBufferCalculated     = "buffer.calculated"
	TypeBufferStatusChanged  = "buffer.status_changed"
	TypeBufferAlertTriggered = "buffer.alert_triggered"
	TypeBufferZoneChanged    = "buffer.zone_changed"

	TypeFADCreated = "fad.created"
	TypeFADUpdated = "fad.updated"
	TypeFADDeleted = "fad.deleted"
	TypeFADApplied = "fad.applied"

	TypeADUCalculated = "adu.calculated"
)

// NATS subject constants.
const (
	SubjectBufferCreated        = "ddmrp.buffer.created"
	SubjectBufferUpdated        = "ddmrp.buffer.updated"
	SubjectBufferCalculated     = "ddmrp.buffer.calculated"
	SubjectBufferStatusChanged  = "ddmrp.buffer.status_changed"
	SubjectBufferAlertTriggered = "ddmrp.buffer.alert_triggered"
	SubjectBufferZoneChanged    = "ddmrp.buffer.zone_changed"

	SubjectFADCreated = "ddmrp.fad.created"
	SubjectFADUpdated = "ddmrp.fad.updated"
	SubjectFADDeleted = "ddmrp.fad.deleted"
	SubjectFADApplied = "ddmrp.fad.applied"

	SubjectADUCalculated = "ddmrp.adu.calculated"
)

// Zone type constants.
const (
	ZoneRed    = "red"
	ZoneYellow = "yellow"
	ZoneGreen  = "green"
)

// Alert level constants.
const (
	AlertLevelCritical = "critical"
	AlertLevelWarning  = "warning"
	AlertLevelInfo     = "info"
)

// BufferCreatedEvent represents a new buffer creation event.
type BufferCreatedEvent struct {
	BufferID       string    `json:"buffer_id"`
	OrganizationID string    `json:"organization_id"`
	ProductID      string    `json:"product_id"`
	LocationID     string    `json:"location_id"`
	ProfileType    string    `json:"profile_type,omitempty"`
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
}

// BufferUpdatedEvent represents a buffer update event.
type BufferUpdatedEvent struct {
	BufferID       string                 `json:"buffer_id"`
	OrganizationID string                 `json:"organization_id"`
	ProductID      string                 `json:"product_id"`
	LocationID     string                 `json:"location_id"`
	Changes        map[string]interface{} `json:"changes,omitempty"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

// BufferCalculatedEvent represents a buffer calculation event.
type BufferCalculatedEvent struct {
	BufferID       string `json:"buffer_id"`
	OrganizationID string `json:"organization_id"`
	ProductID      string `json:"product_id"`
	LocationID     string `json:"location_id"`

	// Calculation inputs
	DLT               int     `json:"dlt"`
	ADU               float64 `json:"adu"`
	LeadTimeFactor    float64 `json:"lead_time_factor,omitempty"`
	VariabilityFactor float64 `json:"variability_factor,omitempty"`

	// Core buffer zones
	RedZoneBase   float64 `json:"red_zone_base"`
	RedZoneSafety float64 `json:"red_zone_safety"`
	RedZone       float64 `json:"red_zone"`
	YellowZone    float64 `json:"yellow_zone"`
	GreenZone     float64 `json:"green_zone"`

	// Derived values
	CPD float64 `json:"cpd"` // Cumulative Physically on hand and Demand
	TOG float64 `json:"tog"` // Top of Green
	TOR float64 `json:"tor"` // Top of Red
	TOY float64 `json:"toy"` // Top of Yellow

	// Current position
	OnHandQty       float64 `json:"on_hand_qty"`
	OpenPOQty       float64 `json:"open_po_qty"`
	OpenSOQty       float64 `json:"open_so_qty"`
	NetFlowPosition float64 `json:"net_flow_position"`

	CalculatedAt time.Time `json:"calculated_at"`
}

// BufferStatusChangedEvent represents a buffer status change.
type BufferStatusChangedEvent struct {
	BufferID       string `json:"buffer_id"`
	OrganizationID string `json:"organization_id"`
	ProductID      string `json:"product_id"`
	LocationID     string `json:"location_id"`

	OldZone    string  `json:"old_zone"`
	NewZone    string  `json:"new_zone"`
	OldNFP     float64 `json:"old_nfp,omitempty"`
	NewNFP     float64 `json:"new_nfp"`
	AlertLevel string  `json:"alert_level"`
	TOG        float64 `json:"tog"`
	TOY        float64 `json:"toy"`
	TOR        float64 `json:"tor"`

	ChangedAt time.Time `json:"changed_at"`
}

// BufferAlertTriggeredEvent represents a buffer alert event.
type BufferAlertTriggeredEvent struct {
	BufferID       string `json:"buffer_id"`
	OrganizationID string `json:"organization_id"`
	ProductID      string `json:"product_id"`
	ProductSKU     string `json:"product_sku,omitempty"`
	ProductName    string `json:"product_name,omitempty"`
	LocationID     string `json:"location_id"`
	LocationName   string `json:"location_name,omitempty"`

	AlertType  string  `json:"alert_type"`
	AlertLevel string  `json:"alert_level"`
	Zone       string  `json:"zone"`
	NFP        float64 `json:"nfp"`
	TOG        float64 `json:"tog"`
	TOY        float64 `json:"toy"`
	TOR        float64 `json:"tor"`

	// Replenishment recommendation
	ReplenishmentQty      float64 `json:"replenishment_qty,omitempty"`
	SuggestedOrderQty     float64 `json:"suggested_order_qty,omitempty"`
	RecommendedSupplierID string  `json:"recommended_supplier_id,omitempty"`
	ExpectedLeadTimeDays  int     `json:"expected_lead_time_days,omitempty"`

	Message     string            `json:"message"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	TriggeredAt time.Time         `json:"triggered_at"`
}

// BufferZoneChangedEvent represents a zone transition event.
type BufferZoneChangedEvent struct {
	BufferStatusChangedEvent
	TransitionReason string `json:"transition_reason,omitempty"`
}

// FADCreatedEvent represents a Flow Adjustment Factor creation.
type FADCreatedEvent struct {
	FADID          string `json:"fad_id"`
	OrganizationID string `json:"organization_id"`
	ProductID      string `json:"product_id"`
	LocationID     string `json:"location_id,omitempty"`

	AdjustmentType string    `json:"adjustment_type"`
	Factor         float64   `json:"factor"`
	StartDate      time.Time `json:"start_date"`
	EndDate        time.Time `json:"end_date"`
	Reason         string    `json:"reason,omitempty"`

	CreatedBy string    `json:"created_by,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// FADUpdatedEvent represents a FAD update.
type FADUpdatedEvent struct {
	FADID          string                 `json:"fad_id"`
	OrganizationID string                 `json:"organization_id"`
	ProductID      string                 `json:"product_id"`
	Changes        map[string]interface{} `json:"changes,omitempty"`
	UpdatedBy      string                 `json:"updated_by,omitempty"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

// FADDeletedEvent represents a FAD deletion.
type FADDeletedEvent struct {
	FADID          string    `json:"fad_id"`
	OrganizationID string    `json:"organization_id"`
	ProductID      string    `json:"product_id"`
	DeletedBy      string    `json:"deleted_by,omitempty"`
	DeletedAt      time.Time `json:"deleted_at"`
}

// FADAppliedEvent represents when a FAD was applied to a buffer calculation.
type FADAppliedEvent struct {
	FADID          string    `json:"fad_id"`
	BufferID       string    `json:"buffer_id"`
	OrganizationID string    `json:"organization_id"`
	ProductID      string    `json:"product_id"`
	Factor         float64   `json:"factor"`
	ImpactOnADU    float64   `json:"impact_on_adu,omitempty"`
	AppliedAt      time.Time `json:"applied_at"`
}

// ADUCalculatedEvent represents an Average Daily Usage calculation.
type ADUCalculatedEvent struct {
	BufferID       string `json:"buffer_id"`
	OrganizationID string `json:"organization_id"`
	ProductID      string `json:"product_id"`
	LocationID     string `json:"location_id"`

	ADU               float64 `json:"adu"`
	PreviousADU       float64 `json:"previous_adu,omitempty"`
	CalculationMethod string  `json:"calculation_method,omitempty"`
	PeriodDays        int     `json:"period_days"`
	DataPointCount    int     `json:"data_point_count,omitempty"`

	CalculatedAt time.Time `json:"calculated_at"`
}
