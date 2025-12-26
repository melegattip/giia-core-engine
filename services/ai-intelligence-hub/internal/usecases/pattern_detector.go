package usecases

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/melegattip/giia-core-engine/pkg/events"
	"github.com/melegattip/giia-core-engine/pkg/logger"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/core/providers"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/domain/entities"
)

// PatternDetector detects cross-event patterns for proactive alerting
type PatternDetector struct {
	eventStore        EventStore
	patternStore      PatternStore
	notificationStore providers.NotificationRepository
	alerter           AlertService
	aiAnalyzer        providers.AIAnalyzer
	config            entities.PatternDetectionConfig
	logger            logger.Logger
}

// EventStore interface for accessing historical events
type EventStore interface {
	GetEventsByType(ctx context.Context, orgID uuid.UUID, eventType string, since time.Time) ([]*events.Event, error)
	GetEventsByEntity(ctx context.Context, orgID uuid.UUID, entityType, entityID string, since time.Time) ([]*events.Event, error)
	GetRecentEvents(ctx context.Context, orgID uuid.UUID, limit int) ([]*events.Event, error)
}

// PatternStore interface for persisting detected patterns
type PatternStore interface {
	Save(ctx context.Context, pattern *entities.Pattern) error
	FindByOrganization(ctx context.Context, orgID uuid.UUID, status entities.PatternStatus) ([]*entities.Pattern, error)
	FindByEntity(ctx context.Context, entityType string, entityID uuid.UUID) ([]*entities.Pattern, error)
	Update(ctx context.Context, pattern *entities.Pattern) error
}

// AlertService interface for sending pattern alerts
type AlertService interface {
	SendPatternAlert(ctx context.Context, pattern *entities.Pattern, notification *domain.AINotification) error
}

// PatternDetectorConfig holds configuration for pattern detector
type PatternDetectorConfig struct {
	EventStore        EventStore
	PatternStore      PatternStore
	NotificationStore providers.NotificationRepository
	Alerter           AlertService
	AIAnalyzer        providers.AIAnalyzer
	Config            entities.PatternDetectionConfig
}

// NewPatternDetector creates a new pattern detector
func NewPatternDetector(cfg PatternDetectorConfig, log logger.Logger) *PatternDetector {
	if cfg.Config.RecurringStockoutThreshold == 0 {
		cfg.Config = entities.DefaultPatternDetectionConfig()
	}

	return &PatternDetector{
		eventStore:        cfg.EventStore,
		patternStore:      cfg.PatternStore,
		notificationStore: cfg.NotificationStore,
		alerter:           cfg.Alerter,
		aiAnalyzer:        cfg.AIAnalyzer,
		config:            cfg.Config,
		logger:            log,
	}
}

// DetectPatterns runs all pattern detection algorithms
func (pd *PatternDetector) DetectPatterns(ctx context.Context, orgID uuid.UUID) ([]*entities.Pattern, error) {
	pd.logger.Info(ctx, "Starting pattern detection", logger.Tags{
		"organization_id": orgID.String(),
	})

	var allPatterns []*entities.Pattern

	// Detect recurring stockouts
	if pd.config.EnableRecurringStockout && pd.eventStore != nil {
		patterns, err := pd.detectRecurringStockouts(ctx, orgID)
		if err != nil {
			pd.logger.Error(ctx, err, "Failed to detect recurring stockouts", nil)
		} else {
			allPatterns = append(allPatterns, patterns...)
		}
	}

	// Detect supplier delays
	if pd.config.EnableSupplierDelay && pd.eventStore != nil {
		patterns, err := pd.detectSupplierDelays(ctx, orgID)
		if err != nil {
			pd.logger.Error(ctx, err, "Failed to detect supplier delays", nil)
		} else {
			allPatterns = append(allPatterns, patterns...)
		}
	}

	// Detect demand spikes
	if pd.config.EnableDemandSpike && pd.eventStore != nil {
		patterns, err := pd.detectDemandSpikes(ctx, orgID)
		if err != nil {
			pd.logger.Error(ctx, err, "Failed to detect demand spikes", nil)
		} else {
			allPatterns = append(allPatterns, patterns...)
		}
	}

	// Detect buffer penetration patterns
	if pd.config.EnableBufferPenetration && pd.eventStore != nil {
		patterns, err := pd.detectBufferPenetrationPatterns(ctx, orgID)
		if err != nil {
			pd.logger.Error(ctx, err, "Failed to detect buffer penetration patterns", nil)
		} else {
			allPatterns = append(allPatterns, patterns...)
		}
	}

	pd.logger.Info(ctx, "Pattern detection completed", logger.Tags{
		"organization_id": orgID.String(),
		"patterns_found":  len(allPatterns),
	})

	// Persist and alert for each pattern
	for _, pattern := range allPatterns {
		if pd.patternStore != nil {
			if err := pd.patternStore.Save(ctx, pattern); err != nil {
				pd.logger.Error(ctx, err, "Failed to save pattern", logger.Tags{
					"pattern_id": pattern.ID.String(),
				})
			}
		}

		// Generate AI root cause analysis
		if pd.aiAnalyzer != nil {
			if err := pd.enrichPatternWithAI(ctx, pattern); err != nil {
				pd.logger.Warn(ctx, "Failed to enrich pattern with AI", logger.Tags{
					"pattern_id": pattern.ID.String(),
					"error":      err.Error(),
				})
			}
		}

		// Send alert
		if pd.alerter != nil {
			notification := pd.createPatternNotification(pattern)
			if err := pd.alerter.SendPatternAlert(ctx, pattern, notification); err != nil {
				pd.logger.Error(ctx, err, "Failed to send pattern alert", logger.Tags{
					"pattern_id": pattern.ID.String(),
				})
			}
		}
	}

	return allPatterns, nil
}

// detectRecurringStockouts finds products with multiple stockout events
func (pd *PatternDetector) detectRecurringStockouts(ctx context.Context, orgID uuid.UUID) ([]*entities.Pattern, error) {
	since := time.Now().Add(-pd.config.RecurringStockoutWindow)

	// Get stockout events
	stockoutEvents, err := pd.eventStore.GetEventsByType(ctx, orgID, "buffer.critical", since)
	if err != nil {
		return nil, fmt.Errorf("failed to get stockout events: %w", err)
	}

	// Group by product
	productEvents := make(map[string][]*events.Event)
	for _, event := range stockoutEvents {
		productID := extractProductID(event)
		if productID != "" {
			productEvents[productID] = append(productEvents[productID], event)
		}
	}

	var patterns []*entities.Pattern

	// Find recurring stockouts
	for productID, productEvts := range productEvents {
		if len(productEvts) >= pd.config.RecurringStockoutThreshold {
			pattern := entities.NewPattern(
				orgID,
				entities.PatternTypeRecurringStockout,
				pd.calculateSeverity(len(productEvts), pd.config.RecurringStockoutThreshold),
				fmt.Sprintf("Recurring stockout risk for product %s", productID),
				fmt.Sprintf("Product %s has experienced %d critical buffer events in the past %d days",
					productID, len(productEvts), int(pd.config.RecurringStockoutWindow.Hours()/24)),
			)

			// Add related events
			for _, evt := range productEvts {
				pattern.AddRelatedEvent(evt.ID, evt.Type, evt.Timestamp, evt.Data)
			}

			// Add affected entity
			pattern.AddAffectedEntity("product", parseUUID(productID), productID, "recurring_stockout")

			// Update statistics
			pattern.UpdateStatistics()
			pattern.Statistics.ConfidenceScore = pd.calculateConfidence(len(productEvts), pd.config.RecurringStockoutThreshold)

			patterns = append(patterns, pattern)
		}
	}

	return patterns, nil
}

// detectSupplierDelays finds suppliers with multiple late deliveries
func (pd *PatternDetector) detectSupplierDelays(ctx context.Context, orgID uuid.UUID) ([]*entities.Pattern, error) {
	since := time.Now().Add(-pd.config.SupplierDelayWindow)

	// Get delivery delay events
	delayEvents, err := pd.eventStore.GetEventsByType(ctx, orgID, "delivery.delayed", since)
	if err != nil {
		return nil, fmt.Errorf("failed to get delay events: %w", err)
	}

	// Group by supplier
	supplierEvents := make(map[string][]*events.Event)
	for _, event := range delayEvents {
		supplierID := extractSupplierID(event)
		if supplierID != "" {
			supplierEvents[supplierID] = append(supplierEvents[supplierID], event)
		}
	}

	var patterns []*entities.Pattern

	// Find recurring delays
	for supplierID, supplierEvts := range supplierEvents {
		if len(supplierEvts) >= pd.config.SupplierDelayThreshold {
			supplierName := extractSupplierName(supplierEvts[0])

			pattern := entities.NewPattern(
				orgID,
				entities.PatternTypeSupplierDelay,
				pd.calculateSeverity(len(supplierEvts), pd.config.SupplierDelayThreshold),
				fmt.Sprintf("Recurring delays from supplier %s", supplierName),
				fmt.Sprintf("Supplier %s has %d late deliveries in the past %d days",
					supplierName, len(supplierEvts), int(pd.config.SupplierDelayWindow.Hours()/24)),
			)

			// Add related events
			for _, evt := range supplierEvts {
				pattern.AddRelatedEvent(evt.ID, evt.Type, evt.Timestamp, evt.Data)
			}

			// Add affected entity
			pattern.AddAffectedEntity("supplier", parseUUID(supplierID), supplierName, "recurring_delays")

			// Calculate total delay impact
			totalDelayDays := 0
			for _, evt := range supplierEvts {
				if delay, ok := evt.Data["delay_days"].(float64); ok {
					totalDelayDays += int(delay)
				}
			}

			pattern.Statistics.TotalImpactValue = float64(totalDelayDays)
			pattern.UpdateStatistics()
			pattern.Statistics.ConfidenceScore = pd.calculateConfidence(len(supplierEvts), pd.config.SupplierDelayThreshold)

			patterns = append(patterns, pattern)
		}
	}

	return patterns, nil
}

// detectDemandSpikes finds unusual demand increases
func (pd *PatternDetector) detectDemandSpikes(ctx context.Context, orgID uuid.UUID) ([]*entities.Pattern, error) {
	since := time.Now().Add(-time.Duration(pd.config.DemandSpikeLookbackDays) * 24 * time.Hour)

	// Get demand events
	demandEvents, err := pd.eventStore.GetEventsByType(ctx, orgID, "demand.updated", since)
	if err != nil {
		return nil, fmt.Errorf("failed to get demand events: %w", err)
	}

	// Group by product
	productDemand := make(map[string][]float64)
	for _, event := range demandEvents {
		productID := extractProductID(event)
		if demand, ok := event.Data["quantity"].(float64); ok && productID != "" {
			productDemand[productID] = append(productDemand[productID], demand)
		}
	}

	var patterns []*entities.Pattern

	// Analyze demand patterns
	for productID, demands := range productDemand {
		if len(demands) < 3 {
			continue
		}

		// Calculate average and check for spikes
		var sum float64
		for _, d := range demands[:len(demands)-1] { // Exclude most recent
			sum += d
		}
		avg := sum / float64(len(demands)-1)
		recentDemand := demands[len(demands)-1]

		percentIncrease := ((recentDemand - avg) / avg) * 100

		if percentIncrease >= pd.config.DemandSpikePercentage {
			pattern := entities.NewPattern(
				orgID,
				entities.PatternTypeDemandSpike,
				pd.calculateDemandSpikeSeverity(percentIncrease),
				fmt.Sprintf("Demand spike detected for product %s", productID),
				fmt.Sprintf("Product %s has %.1f%% demand increase compared to average",
					productID, percentIncrease),
			)

			pattern.AddAffectedEntity("product", parseUUID(productID), productID, "demand_spike")
			pattern.Statistics.TotalImpactValue = recentDemand - avg
			pattern.Statistics.ConfidenceScore = 0.8

			patterns = append(patterns, pattern)
		}
	}

	return patterns, nil
}

// detectBufferPenetrationPatterns finds products with frequent buffer zone crossings
func (pd *PatternDetector) detectBufferPenetrationPatterns(ctx context.Context, orgID uuid.UUID) ([]*entities.Pattern, error) {
	since := time.Now().Add(-7 * 24 * time.Hour) // Last 7 days

	// Get buffer status events
	bufferEvents, err := pd.eventStore.GetEventsByType(ctx, orgID, "buffer.status_changed", since)
	if err != nil {
		return nil, fmt.Errorf("failed to get buffer events: %w", err)
	}

	// Track zone transitions per product
	type transition struct {
		fromZone string
		toZone   string
		time     time.Time
	}
	productTransitions := make(map[string][]transition)

	for _, event := range bufferEvents {
		productID := extractProductID(event)
		if productID == "" {
			continue
		}

		fromZone, _ := event.Data["from_zone"].(string)
		toZone, _ := event.Data["to_zone"].(string)

		if fromZone != "" && toZone != "" {
			productTransitions[productID] = append(productTransitions[productID], transition{
				fromZone: fromZone,
				toZone:   toZone,
				time:     event.Timestamp,
			})
		}
	}

	var patterns []*entities.Pattern

	// Analyze transitions
	for productID, transitions := range productTransitions {
		// Count concerning transitions (yellow->red or red zone entries)
		criticalCount := 0
		for _, t := range transitions {
			if t.toZone == "red" || (t.fromZone == "yellow" && t.toZone == "red") {
				criticalCount++
			}
		}

		if criticalCount >= 3 {
			pattern := entities.NewPattern(
				orgID,
				entities.PatternTypeBufferPenetration,
				entities.PatternSeverityHigh,
				fmt.Sprintf("Frequent buffer zone violations for product %s", productID),
				fmt.Sprintf("Product %s has entered critical buffer zones %d times in 7 days",
					productID, criticalCount),
			)

			pattern.AddAffectedEntity("product", parseUUID(productID), productID, "buffer_instability")
			pattern.Statistics.Occurrences = criticalCount
			pattern.Statistics.ConfidenceScore = 0.85

			patterns = append(patterns, pattern)
		}
	}

	return patterns, nil
}

// enrichPatternWithAI uses AI to generate root cause analysis
func (pd *PatternDetector) enrichPatternWithAI(ctx context.Context, pattern *entities.Pattern) error {
	// Build event summaries
	var eventSummaries []string
	for _, evt := range pattern.RelatedEvents {
		eventSummaries = append(eventSummaries, fmt.Sprintf("- %s at %s: %v",
			evt.EventType, evt.Timestamp.Format("2006-01-02 15:04"), evt.Data))
	}

	// Build entity list
	var entityList []string
	for _, entity := range pattern.AffectedEntities {
		entityList = append(entityList, fmt.Sprintf("%s: %s (%s)",
			entity.EntityType, entity.EntityName, entity.Impact))
	}

	// Create analysis request
	request := &providers.AIAnalysisRequest{
		Event: &events.Event{
			ID:        pattern.ID.String(),
			Type:      string(pattern.Type),
			Source:    "pattern_detector",
			Timestamp: pattern.FirstDetectedAt,
			Data: map[string]interface{}{
				"occurrences":      pattern.Statistics.Occurrences,
				"time_window_days": pattern.Statistics.TimeWindowDays,
				"severity":         string(pattern.Severity),
			},
		},
		Context: map[string]interface{}{
			"pattern_type":      string(pattern.Type),
			"affected_entities": entityList,
			"event_count":       len(pattern.RelatedEvents),
		},
		Prompt: fmt.Sprintf(`Analyze this recurring pattern and identify root causes:
Pattern Type: %s
Occurrences: %d in %d days
Affected Entities: %s
Events: %s`,
			pattern.Type,
			pattern.Statistics.Occurrences,
			pattern.Statistics.TimeWindowDays,
			strings.Join(entityList, ", "),
			strings.Join(eventSummaries, "\n")),
	}

	response, err := pd.aiAnalyzer.Analyze(ctx, request)
	if err != nil {
		return err
	}

	// Update pattern with AI insights
	pattern.SetRootCauseAnalysis(response.FullAnalysis)

	for i, rec := range response.Recommendations {
		var estimatedROI float64
		if response.ImpactAssessment.RevenueImpact > 0 {
			estimatedROI = response.ImpactAssessment.RevenueImpact / float64(i+1)
		}
		pattern.AddRecommendation(rec.Action, i+1, rec.ExpectedOutcome, rec.Effort, estimatedROI)
	}

	pattern.Statistics.ConfidenceScore = response.Confidence

	return nil
}

// createPatternNotification creates a notification for a detected pattern
func (pd *PatternDetector) createPatternNotification(pattern *entities.Pattern) *domain.AINotification {
	priority := domain.NotificationPriorityMedium
	switch pattern.Severity {
	case entities.PatternSeverityCritical:
		priority = domain.NotificationPriorityCritical
	case entities.PatternSeverityHigh:
		priority = domain.NotificationPriorityHigh
	case entities.PatternSeverityLow:
		priority = domain.NotificationPriorityLow
	}

	notification := domain.NewNotification(
		pattern.OrganizationID,
		uuid.Nil, // Will be set by alerter based on subscription
		domain.NotificationTypeInsight,
		priority,
		pattern.Title,
		pattern.Description,
	)

	notification.FullAnalysis = pattern.RootCauseAnalysis
	notification.Reasoning = fmt.Sprintf("Pattern detected with %d occurrences and %.0f%% confidence",
		pattern.Statistics.Occurrences, pattern.Statistics.ConfidenceScore*100)

	// Convert recommendations
	for _, rec := range pattern.Recommendations {
		notification.Recommendations = append(notification.Recommendations, domain.Recommendation{
			Action:          rec.Action,
			ExpectedOutcome: rec.ExpectedOutcome,
			Effort:          rec.Effort,
			PriorityOrder:   rec.Priority,
		})
	}

	// Sort recommendations by priority
	sort.Slice(notification.Recommendations, func(i, j int) bool {
		return notification.Recommendations[i].PriorityOrder < notification.Recommendations[j].PriorityOrder
	})

	return notification
}

// calculateSeverity determines pattern severity based on occurrence count
func (pd *PatternDetector) calculateSeverity(occurrences, threshold int) entities.PatternSeverity {
	ratio := float64(occurrences) / float64(threshold)
	if ratio >= 3 {
		return entities.PatternSeverityCritical
	} else if ratio >= 2 {
		return entities.PatternSeverityHigh
	} else if ratio >= 1.5 {
		return entities.PatternSeverityMedium
	}
	return entities.PatternSeverityLow
}

// calculateDemandSpikeSeverity determines severity based on percentage increase
func (pd *PatternDetector) calculateDemandSpikeSeverity(percentIncrease float64) entities.PatternSeverity {
	if percentIncrease >= 200 {
		return entities.PatternSeverityCritical
	} else if percentIncrease >= 100 {
		return entities.PatternSeverityHigh
	} else if percentIncrease >= 75 {
		return entities.PatternSeverityMedium
	}
	return entities.PatternSeverityLow
}

// calculateConfidence determines confidence based on data quality
func (pd *PatternDetector) calculateConfidence(occurrences, threshold int) float64 {
	base := 0.7
	bonus := float64(occurrences-threshold) * 0.05
	confidence := base + bonus
	if confidence > 0.95 {
		return 0.95
	}
	return confidence
}

// Helper functions to extract data from events
func extractProductID(event *events.Event) string {
	if id, ok := event.Data["product_id"].(string); ok {
		return id
	}
	if id, ok := event.Data["buffer_id"].(string); ok {
		return id
	}
	return ""
}

func extractSupplierID(event *events.Event) string {
	if id, ok := event.Data["supplier_id"].(string); ok {
		return id
	}
	return ""
}

func extractSupplierName(event *events.Event) string {
	if name, ok := event.Data["supplier_name"].(string); ok {
		return name
	}
	return extractSupplierID(event)
}

func parseUUID(s string) uuid.UUID {
	id, _ := uuid.Parse(s)
	return id
}

// Compile-time check that ensures any errors type checking are caught
var _ = errors.New
