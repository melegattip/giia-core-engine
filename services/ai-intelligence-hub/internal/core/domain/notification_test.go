package domain_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/core/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewNotification(t *testing.T) {
	orgID := uuid.New()
	userID := uuid.New()

	notif := domain.NewNotification(
		orgID,
		userID,
		domain.NotificationTypeAlert,
		domain.NotificationPriorityCritical,
		"Test Title",
		"Test Summary",
	)

	assert.NotNil(t, notif)
	assert.NotEqual(t, uuid.Nil, notif.ID)
	assert.Equal(t, orgID, notif.OrganizationID)
	assert.Equal(t, userID, notif.UserID)
	assert.Equal(t, domain.NotificationTypeAlert, notif.Type)
	assert.Equal(t, domain.NotificationPriorityCritical, notif.Priority)
	assert.Equal(t, "Test Title", notif.Title)
	assert.Equal(t, "Test Summary", notif.Summary)
	assert.Equal(t, domain.NotificationStatusUnread, notif.Status)
	assert.NotNil(t, notif.SourceEvents)
	assert.NotNil(t, notif.RelatedEntities)
	assert.NotNil(t, notif.Recommendations)
	assert.Empty(t, notif.SourceEvents)
	assert.Empty(t, notif.RelatedEntities)
	assert.Empty(t, notif.Recommendations)
}

func TestAINotification_MarkAsRead(t *testing.T) {
	notif := domain.NewNotification(
		uuid.New(),
		uuid.New(),
		domain.NotificationTypeInfo,
		domain.NotificationPriorityLow,
		"Test",
		"Test Summary",
	)

	// Verify initial state
	assert.Equal(t, domain.NotificationStatusUnread, notif.Status)
	assert.Nil(t, notif.ReadAt)

	// Mark as read
	beforeMark := time.Now()
	notif.MarkAsRead()
	afterMark := time.Now()

	// Verify changes
	assert.Equal(t, domain.NotificationStatusRead, notif.Status)
	require.NotNil(t, notif.ReadAt)
	assert.True(t, notif.ReadAt.After(beforeMark) || notif.ReadAt.Equal(beforeMark))
	assert.True(t, notif.ReadAt.Before(afterMark) || notif.ReadAt.Equal(afterMark))
	assert.Nil(t, notif.ActedAt)
	assert.Nil(t, notif.DismissedAt)
}

func TestAINotification_MarkAsActedUpon(t *testing.T) {
	notif := domain.NewNotification(
		uuid.New(),
		uuid.New(),
		domain.NotificationTypeWarning,
		domain.NotificationPriorityHigh,
		"Test",
		"Test Summary",
	)

	// Verify initial state
	assert.Equal(t, domain.NotificationStatusUnread, notif.Status)
	assert.Nil(t, notif.ActedAt)

	// Mark as acted upon
	beforeMark := time.Now()
	notif.MarkAsActedUpon()
	afterMark := time.Now()

	// Verify changes
	assert.Equal(t, domain.NotificationStatusActedUpon, notif.Status)
	require.NotNil(t, notif.ActedAt)
	assert.True(t, notif.ActedAt.After(beforeMark) || notif.ActedAt.Equal(beforeMark))
	assert.True(t, notif.ActedAt.Before(afterMark) || notif.ActedAt.Equal(afterMark))
	assert.Nil(t, notif.ReadAt)
	assert.Nil(t, notif.DismissedAt)
}

func TestAINotification_Dismiss(t *testing.T) {
	notif := domain.NewNotification(
		uuid.New(),
		uuid.New(),
		domain.NotificationTypeSuggestion,
		domain.NotificationPriorityMedium,
		"Test",
		"Test Summary",
	)

	// Verify initial state
	assert.Equal(t, domain.NotificationStatusUnread, notif.Status)
	assert.Nil(t, notif.DismissedAt)

	// Dismiss
	beforeMark := time.Now()
	notif.Dismiss()
	afterMark := time.Now()

	// Verify changes
	assert.Equal(t, domain.NotificationStatusDismissed, notif.Status)
	require.NotNil(t, notif.DismissedAt)
	assert.True(t, notif.DismissedAt.After(beforeMark) || notif.DismissedAt.Equal(beforeMark))
	assert.True(t, notif.DismissedAt.Before(afterMark) || notif.DismissedAt.Equal(afterMark))
	assert.Nil(t, notif.ReadAt)
	assert.Nil(t, notif.ActedAt)
}

func TestNotificationTypes(t *testing.T) {
	tests := []struct {
		name        string
		notifType   domain.NotificationType
		expectedStr string
	}{
		{"Alert", domain.NotificationTypeAlert, "alert"},
		{"Warning", domain.NotificationTypeWarning, "warning"},
		{"Info", domain.NotificationTypeInfo, "info"},
		{"Suggestion", domain.NotificationTypeSuggestion, "suggestion"},
		{"Insight", domain.NotificationTypeInsight, "insight"},
		{"Digest", domain.NotificationTypeDigest, "digest"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedStr, string(tt.notifType))
		})
	}
}

func TestNotificationPriorities(t *testing.T) {
	tests := []struct {
		name        string
		priority    domain.NotificationPriority
		expectedStr string
	}{
		{"Critical", domain.NotificationPriorityCritical, "critical"},
		{"High", domain.NotificationPriorityHigh, "high"},
		{"Medium", domain.NotificationPriorityMedium, "medium"},
		{"Low", domain.NotificationPriorityLow, "low"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedStr, string(tt.priority))
		})
	}
}

func TestNotificationStatuses(t *testing.T) {
	tests := []struct {
		name        string
		status      domain.NotificationStatus
		expectedStr string
	}{
		{"Unread", domain.NotificationStatusUnread, "unread"},
		{"Read", domain.NotificationStatusRead, "read"},
		{"ActedUpon", domain.NotificationStatusActedUpon, "acted_upon"},
		{"Dismissed", domain.NotificationStatusDismissed, "dismissed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedStr, string(tt.status))
		})
	}
}

func TestAINotification_CompleteWorkflow(t *testing.T) {
	// Create notification
	notif := domain.NewNotification(
		uuid.New(),
		uuid.New(),
		domain.NotificationTypeAlert,
		domain.NotificationPriorityCritical,
		"Critical Buffer Alert",
		"Product ABC below minimum buffer",
	)

	// Add full analysis and reasoning
	notif.FullAnalysis = "Detailed AI analysis of the buffer situation..."
	notif.Reasoning = "DDMRP methodology indicates immediate action required"

	// Add impact assessment
	duration := 24 * time.Hour
	notif.Impact = domain.ImpactAssessment{
		RiskLevel:        "critical",
		RevenueImpact:    15000.00,
		CostImpact:       200.00,
		TimeToImpact:     &duration,
		AffectedOrders:   5,
		AffectedProducts: 1,
	}

	// Add recommendations
	notif.Recommendations = []domain.Recommendation{
		{
			Action:          "Place emergency replenishment order",
			Reasoning:       "Current stock insufficient for lead time",
			ExpectedOutcome: "Stockout prevented, buffer restored",
			Effort:          "medium",
			Impact:          "high",
			PriorityOrder:   1,
		},
		{
			Action:          "Review buffer parameters",
			Reasoning:       "Frequent red zone penetration indicates undersized buffer",
			ExpectedOutcome: "Improved buffer sizing, reduced future stockout risk",
			Effort:          "low",
			Impact:          "medium",
			PriorityOrder:   2,
		},
	}

	// Add source events and related entities
	notif.SourceEvents = []string{"evt-123", "evt-456"}
	notif.RelatedEntities = map[string][]string{
		"product_ids":  {"prod-abc-123"},
		"supplier_ids": {"sup-xyz-789"},
	}

	// Verify complete structure
	assert.NotEqual(t, uuid.Nil, notif.ID)
	assert.Equal(t, "Critical Buffer Alert", notif.Title)
	assert.Equal(t, domain.NotificationStatusUnread, notif.Status)
	assert.Equal(t, "critical", notif.Impact.RiskLevel)
	assert.Equal(t, 15000.00, notif.Impact.RevenueImpact)
	assert.Len(t, notif.Recommendations, 2)
	assert.Equal(t, "Place emergency replenishment order", notif.Recommendations[0].Action)
	assert.Len(t, notif.SourceEvents, 2)
	assert.Len(t, notif.RelatedEntities["product_ids"], 1)

	// Test status transitions
	notif.MarkAsRead()
	assert.Equal(t, domain.NotificationStatusRead, notif.Status)
	assert.NotNil(t, notif.ReadAt)

	notif.MarkAsActedUpon()
	assert.Equal(t, domain.NotificationStatusActedUpon, notif.Status)
	assert.NotNil(t, notif.ActedAt)
}

func TestImpactAssessment(t *testing.T) {
	duration := 48 * time.Hour
	impact := domain.ImpactAssessment{
		RiskLevel:        "high",
		RevenueImpact:    25000.00,
		CostImpact:       500.00,
		TimeToImpact:     &duration,
		AffectedOrders:   10,
		AffectedProducts: 3,
	}

	assert.Equal(t, "high", impact.RiskLevel)
	assert.Equal(t, 25000.00, impact.RevenueImpact)
	assert.Equal(t, 500.00, impact.CostImpact)
	assert.Equal(t, 48*time.Hour, *impact.TimeToImpact)
	assert.Equal(t, 10, impact.AffectedOrders)
	assert.Equal(t, 3, impact.AffectedProducts)
}

func TestRecommendation(t *testing.T) {
	rec := domain.Recommendation{
		Action:          "Increase buffer green zone",
		Reasoning:       "Historical demand variability higher than expected",
		ExpectedOutcome: "Reduced stockout risk by 40%",
		Effort:          "low",
		Impact:          "high",
		ActionURL:       "/buffers/prod-123/edit",
		PriorityOrder:   1,
	}

	assert.Equal(t, "Increase buffer green zone", rec.Action)
	assert.Equal(t, "Historical demand variability higher than expected", rec.Reasoning)
	assert.Equal(t, "Reduced stockout risk by 40%", rec.ExpectedOutcome)
	assert.Equal(t, "low", rec.Effort)
	assert.Equal(t, "high", rec.Impact)
	assert.Equal(t, "/buffers/prod-123/edit", rec.ActionURL)
	assert.Equal(t, 1, rec.PriorityOrder)
}
