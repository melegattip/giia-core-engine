package entities

import (
	"time"

	"github.com/google/uuid"
)

// PatternType represents the type of detected pattern
type PatternType string

const (
	PatternTypeRecurringStockout PatternType = "recurring_stockout"
	PatternTypeSupplierDelay     PatternType = "supplier_delay"
	PatternTypeDemandSpike       PatternType = "demand_spike"
	PatternTypeSeasonalTrend     PatternType = "seasonal_trend"
	PatternTypeBufferPenetration PatternType = "buffer_penetration"
	PatternTypeLeadTimeVariation PatternType = "lead_time_variation"
	PatternTypeCostAnomaly       PatternType = "cost_anomaly"
)

// PatternSeverity represents how severe the detected pattern is
type PatternSeverity string

const (
	PatternSeverityCritical PatternSeverity = "critical"
	PatternSeverityHigh     PatternSeverity = "high"
	PatternSeverityMedium   PatternSeverity = "medium"
	PatternSeverityLow      PatternSeverity = "low"
)

// PatternStatus represents the lifecycle status of a pattern
type PatternStatus string

const (
	PatternStatusNew        PatternStatus = "new"
	PatternStatusConfirmed  PatternStatus = "confirmed"
	PatternStatusInProgress PatternStatus = "in_progress"
	PatternStatusResolved   PatternStatus = "resolved"
	PatternStatusDismissed  PatternStatus = "dismissed"
)

// Pattern represents a detected cross-event pattern
type Pattern struct {
	ID                uuid.UUID
	OrganizationID    uuid.UUID
	Type              PatternType
	Severity          PatternSeverity
	Status            PatternStatus
	Title             string
	Description       string
	RootCauseAnalysis string
	RelatedEvents     []PatternEvent
	AffectedEntities  []AffectedEntity
	Statistics        PatternStatistics
	Recommendations   []PatternRecommendation
	FirstDetectedAt   time.Time
	LastUpdatedAt     time.Time
	ResolvedAt        *time.Time
	NotificationID    *uuid.UUID
	Metadata          map[string]interface{}
}

// PatternEvent represents an event that contributed to a pattern
type PatternEvent struct {
	EventID   string
	EventType string
	Timestamp time.Time
	Data      map[string]interface{}
}

// AffectedEntity represents an entity affected by the pattern
type AffectedEntity struct {
	EntityType string // "product", "supplier", "warehouse", etc.
	EntityID   uuid.UUID
	EntityName string
	Impact     string
}

// PatternStatistics contains numerical data about the pattern
type PatternStatistics struct {
	Occurrences        int
	TimeWindowDays     int
	AverageInterval    time.Duration
	TotalImpactValue   float64
	AffectedOrderCount int
	TrendDirection     string // "increasing", "decreasing", "stable"
	ConfidenceScore    float64
}

// PatternRecommendation represents a suggested action for the pattern
type PatternRecommendation struct {
	Action          string
	Priority        int
	ExpectedOutcome string
	Effort          string
	EstimatedROI    float64
}

// PatternDetectionConfig contains configuration for pattern detection
type PatternDetectionConfig struct {
	RecurringStockoutThreshold int           // Number of occurrences
	RecurringStockoutWindow    time.Duration // Time window for detection
	SupplierDelayThreshold     int           // Number of late deliveries
	SupplierDelayWindow        time.Duration
	DemandSpikePercentage      float64 // Percentage above average
	DemandSpikeLookbackDays    int
	MinConfidenceScore         float64
	EnableRecurringStockout    bool
	EnableSupplierDelay        bool
	EnableDemandSpike          bool
	EnableSeasonalTrend        bool
	EnableBufferPenetration    bool
	EnableLeadTimeVariation    bool
}

// DefaultPatternDetectionConfig returns the default configuration
func DefaultPatternDetectionConfig() PatternDetectionConfig {
	return PatternDetectionConfig{
		RecurringStockoutThreshold: 3,
		RecurringStockoutWindow:    7 * 24 * time.Hour, // 7 days
		SupplierDelayThreshold:     3,
		SupplierDelayWindow:        14 * 24 * time.Hour, // 14 days
		DemandSpikePercentage:      50.0,                // 50% above average
		DemandSpikeLookbackDays:    30,
		MinConfidenceScore:         0.7,
		EnableRecurringStockout:    true,
		EnableSupplierDelay:        true,
		EnableDemandSpike:          true,
		EnableSeasonalTrend:        true,
		EnableBufferPenetration:    true,
		EnableLeadTimeVariation:    true,
	}
}

// NewPattern creates a new pattern
func NewPattern(
	organizationID uuid.UUID,
	patternType PatternType,
	severity PatternSeverity,
	title string,
	description string,
) *Pattern {
	now := time.Now()
	return &Pattern{
		ID:               uuid.New(),
		OrganizationID:   organizationID,
		Type:             patternType,
		Severity:         severity,
		Status:           PatternStatusNew,
		Title:            title,
		Description:      description,
		RelatedEvents:    []PatternEvent{},
		AffectedEntities: []AffectedEntity{},
		Recommendations:  []PatternRecommendation{},
		FirstDetectedAt:  now,
		LastUpdatedAt:    now,
		Metadata:         make(map[string]interface{}),
	}
}

// AddRelatedEvent adds an event to the pattern
func (p *Pattern) AddRelatedEvent(eventID, eventType string, timestamp time.Time, data map[string]interface{}) {
	p.RelatedEvents = append(p.RelatedEvents, PatternEvent{
		EventID:   eventID,
		EventType: eventType,
		Timestamp: timestamp,
		Data:      data,
	})
	p.LastUpdatedAt = time.Now()
	p.Statistics.Occurrences = len(p.RelatedEvents)
}

// AddAffectedEntity adds an affected entity to the pattern
func (p *Pattern) AddAffectedEntity(entityType string, entityID uuid.UUID, entityName, impact string) {
	p.AffectedEntities = append(p.AffectedEntities, AffectedEntity{
		EntityType: entityType,
		EntityID:   entityID,
		EntityName: entityName,
		Impact:     impact,
	})
}

// AddRecommendation adds a recommendation to the pattern
func (p *Pattern) AddRecommendation(action string, priority int, expectedOutcome, effort string, estimatedROI float64) {
	p.Recommendations = append(p.Recommendations, PatternRecommendation{
		Action:          action,
		Priority:        priority,
		ExpectedOutcome: expectedOutcome,
		Effort:          effort,
		EstimatedROI:    estimatedROI,
	})
}

// Confirm marks the pattern as confirmed
func (p *Pattern) Confirm() {
	p.Status = PatternStatusConfirmed
	p.LastUpdatedAt = time.Now()
}

// MarkInProgress marks the pattern as being addressed
func (p *Pattern) MarkInProgress() {
	p.Status = PatternStatusInProgress
	p.LastUpdatedAt = time.Now()
}

// Resolve marks the pattern as resolved
func (p *Pattern) Resolve() {
	now := time.Now()
	p.Status = PatternStatusResolved
	p.ResolvedAt = &now
	p.LastUpdatedAt = now
}

// Dismiss marks the pattern as dismissed
func (p *Pattern) Dismiss() {
	p.Status = PatternStatusDismissed
	p.LastUpdatedAt = time.Now()
}

// SetRootCauseAnalysis sets the root cause analysis
func (p *Pattern) SetRootCauseAnalysis(analysis string) {
	p.RootCauseAnalysis = analysis
	p.LastUpdatedAt = time.Now()
}

// CalculateTimeWindow calculates the time window between first and last event
func (p *Pattern) CalculateTimeWindow() time.Duration {
	if len(p.RelatedEvents) < 2 {
		return 0
	}

	first := p.RelatedEvents[0].Timestamp
	last := p.RelatedEvents[len(p.RelatedEvents)-1].Timestamp

	for _, e := range p.RelatedEvents {
		if e.Timestamp.Before(first) {
			first = e.Timestamp
		}
		if e.Timestamp.After(last) {
			last = e.Timestamp
		}
	}

	return last.Sub(first)
}

// CalculateAverageInterval calculates the average time between events
func (p *Pattern) CalculateAverageInterval() time.Duration {
	if len(p.RelatedEvents) < 2 {
		return 0
	}

	totalWindow := p.CalculateTimeWindow()
	intervals := len(p.RelatedEvents) - 1
	if intervals <= 0 {
		return 0
	}

	return totalWindow / time.Duration(intervals)
}

// UpdateStatistics updates the pattern statistics
func (p *Pattern) UpdateStatistics() {
	p.Statistics.Occurrences = len(p.RelatedEvents)
	p.Statistics.AverageInterval = p.CalculateAverageInterval()
	p.Statistics.TimeWindowDays = int(p.CalculateTimeWindow().Hours() / 24)
}
