package testhelpers

import (
	"time"

	"github.com/google/uuid"
	"github.com/melegattip/giia-core-engine/pkg/events"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/core/providers"
)

// CreateTestEvent creates a test event with default values
func CreateTestEvent(eventType string, data map[string]interface{}) *events.Event {
	return &events.Event{
		ID:             uuid.NewString(),
		Type:           eventType,
		OrganizationID: uuid.NewString(),
		Timestamp:      time.Now(),
		Data:           data,
	}
}

// CreateBufferBelowMinimumEvent creates a buffer.below_minimum event for testing
func CreateBufferBelowMinimumEvent(productID string, currentStock, minBuffer, dailyConsumption float64) *events.Event {
	return CreateTestEvent("buffer.below_minimum", map[string]interface{}{
		"product_id":        productID,
		"current_stock":     currentStock,
		"min_buffer":        minBuffer,
		"daily_consumption": dailyConsumption,
	})
}

// CreateTestNotification creates a test notification with default values
func CreateTestNotification() *domain.AINotification {
	return domain.NewNotification(
		uuid.New(),
		uuid.New(),
		domain.NotificationTypeAlert,
		domain.NotificationPriorityCritical,
		"Test Notification",
		"Test Summary",
	)
}

// CreateTestAIResponse creates a test AI analysis response
func CreateTestAIResponse() *providers.AIAnalysisResponse {
	return &providers.AIAnalysisResponse{
		Summary:      "Test AI Summary",
		FullAnalysis: "Test Full Analysis",
		Reasoning:    "Test Reasoning",
		Recommendations: []providers.AIRecommendation{
			{
				Action:          "Test Action",
				Reasoning:       "Test Action Reasoning",
				ExpectedOutcome: "Test Outcome",
				Effort:          "medium",
				Impact:          "high",
			},
		},
		ImpactAssessment: providers.AIImpactAssessment{
			RiskLevel:         "high",
			RevenueImpact:     10000.00,
			CostImpact:        500.00,
			TimeToImpactHours: 48,
			AffectedOrders:    5,
			AffectedProducts:  1,
		},
		Confidence: 0.85,
	}
}

// CreateCompleteNotification creates a notification with full details for testing
func CreateCompleteNotification(orgID, userID uuid.UUID) *domain.AINotification {
	notif := domain.NewNotification(
		orgID,
		userID,
		domain.NotificationTypeAlert,
		domain.NotificationPriorityCritical,
		"Complete Test Notification",
		"This is a test notification with all fields populated",
	)

	notif.FullAnalysis = "Detailed analysis of the situation..."
	notif.Reasoning = "Based on DDMRP methodology and current data..."

	duration := 24 * time.Hour
	notif.Impact = domain.ImpactAssessment{
		RiskLevel:        "critical",
		RevenueImpact:    25000.00,
		CostImpact:       1000.00,
		TimeToImpact:     &duration,
		AffectedOrders:   10,
		AffectedProducts: 3,
	}

	notif.Recommendations = []domain.Recommendation{
		{
			Action:          "Emergency replenishment order",
			Reasoning:       "Stock level critical",
			ExpectedOutcome: "Prevent stockout",
			Effort:          "high",
			Impact:          "critical",
			PriorityOrder:   1,
		},
		{
			Action:          "Review buffer parameters",
			Reasoning:       "Frequent red zone penetration",
			ExpectedOutcome: "Improved buffer sizing",
			Effort:          "medium",
			Impact:          "high",
			PriorityOrder:   2,
		},
	}

	notif.SourceEvents = []string{"evt-1", "evt-2"}
	notif.RelatedEntities = map[string][]string{
		"product_ids":  {"prod-123", "prod-456"},
		"supplier_ids": {"sup-789"},
	}

	return notif
}

// CreateNotificationFilters creates test notification filters
func CreateNotificationFilters() *providers.NotificationFilters {
	return &providers.NotificationFilters{
		Types: []domain.NotificationType{
			domain.NotificationTypeAlert,
			domain.NotificationTypeWarning,
		},
		Priorities: []domain.NotificationPriority{
			domain.NotificationPriorityCritical,
			domain.NotificationPriorityHigh,
		},
		Statuses: []domain.NotificationStatus{
			domain.NotificationStatusUnread,
		},
		Limit:  10,
		Offset: 0,
	}
}

// AssertNotificationHasBasicFields checks that a notification has all required basic fields
func AssertNotificationHasBasicFields(notif *domain.AINotification) bool {
	return notif.ID != uuid.Nil &&
		notif.OrganizationID != uuid.Nil &&
		notif.UserID != uuid.Nil &&
		notif.Type != "" &&
		notif.Priority != "" &&
		notif.Title != "" &&
		notif.Status != "" &&
		!notif.CreatedAt.IsZero()
}
