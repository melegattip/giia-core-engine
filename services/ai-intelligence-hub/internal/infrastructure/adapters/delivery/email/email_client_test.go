package email

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/melegattip/giia-core-engine/pkg/logger"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/core/domain"
)

// MockLogger is a mock logger for testing
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Debug(ctx context.Context, msg string, tags logger.Tags) {
	m.Called(ctx, msg, tags)
}

func (m *MockLogger) Info(ctx context.Context, msg string, tags logger.Tags) {
	m.Called(ctx, msg, tags)
}

func (m *MockLogger) Warn(ctx context.Context, msg string, tags logger.Tags) {
	m.Called(ctx, msg, tags)
}

func (m *MockLogger) Error(ctx context.Context, err error, msg string, tags logger.Tags) {
	m.Called(ctx, err, msg, tags)
}

func (m *MockLogger) Fatal(ctx context.Context, err error, msg string, tags logger.Tags) {
	m.Called(ctx, err, msg, tags)
}

func TestNewEmailClient(t *testing.T) {
	mockLogger := new(MockLogger)

	client := NewEmailClient(
		"smtp.example.com",
		587,
		"user@example.com",
		"password123",
		"noreply@giia.com",
		"GIIA Notifications",
		mockLogger,
	)

	assert.NotNil(t, client)
}

func TestEmailClient_SendEmail_Success(t *testing.T) {
	mockLogger := new(MockLogger)

	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()

	client := &EmailClient{
		smtpHost:     "smtp.example.com",
		smtpPort:     587,
		smtpUser:     "user@example.com",
		smtpPassword: "password123",
		fromAddress:  "noreply@giia.com",
		fromName:     "GIIA Notifications",
		logger:       mockLogger,
	}

	messageID, err := client.SendEmail(
		context.Background(),
		[]string{"recipient@example.com"},
		"Test Subject",
		"<html><body>HTML Body</body></html>",
		"Text Body",
	)

	require.NoError(t, err)
	assert.NotEmpty(t, messageID)
	assert.Contains(t, messageID, "msg_")
}

func TestEmailClient_SendEmail_MultipleRecipients(t *testing.T) {
	mockLogger := new(MockLogger)

	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()

	client := &EmailClient{
		smtpHost:     "smtp.example.com",
		smtpPort:     587,
		smtpUser:     "user@example.com",
		smtpPassword: "password123",
		fromAddress:  "noreply@giia.com",
		fromName:     "GIIA Notifications",
		logger:       mockLogger,
	}

	recipients := []string{
		"user1@example.com",
		"user2@example.com",
		"user3@example.com",
	}

	messageID, err := client.SendEmail(
		context.Background(),
		recipients,
		"Test Subject",
		"<html><body>HTML Body</body></html>",
		"Text Body",
	)

	require.NoError(t, err)
	assert.NotEmpty(t, messageID)
}

func TestEmailClient_SendNotificationEmail_Success(t *testing.T) {
	mockLogger := new(MockLogger)

	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()

	client := &EmailClient{
		smtpHost:     "smtp.example.com",
		smtpPort:     587,
		smtpUser:     "user@example.com",
		smtpPassword: "password123",
		fromAddress:  "noreply@giia.com",
		fromName:     "GIIA Notifications",
		logger:       mockLogger,
	}

	notification := &domain.AINotification{
		ID:           uuid.New(),
		Title:        "Critical Stock Alert",
		Summary:      "Stock level for Product XYZ is critically low",
		FullAnalysis: "Detailed analysis of the stock situation...",
		Priority:     domain.NotificationPriorityCritical,
		Type:         domain.NotificationTypeAlert,
		Status:       domain.NotificationStatusUnread,
		CreatedAt:    time.Now(),
		Impact: domain.ImpactAssessment{
			RiskLevel:        "high",
			RevenueImpact:    15000.00,
			CostImpact:       5000.00,
			AffectedOrders:   25,
			AffectedProducts: 3,
		},
		Recommendations: []domain.Recommendation{
			{
				Action:          "Reorder immediately",
				Reasoning:       "Stock will run out in 2 days",
				ExpectedOutcome: "Prevent stockout",
				Effort:          "low",
				Impact:          "high",
				PriorityOrder:   1,
			},
		},
	}

	messageID, err := client.SendNotificationEmail(
		context.Background(),
		notification,
		[]string{"manager@company.com"},
	)

	require.NoError(t, err)
	assert.NotEmpty(t, messageID)
}

func TestEmailClient_BuildSubject(t *testing.T) {
	mockLogger := new(MockLogger)

	client := &EmailClient{
		logger: mockLogger,
	}

	tests := []struct {
		name     string
		priority domain.NotificationPriority
		title    string
		expected string
	}{
		{
			name:     "Critical priority",
			priority: domain.NotificationPriorityCritical,
			title:    "Stock Alert",
			expected: "🚨 CRITICAL: Stock Alert",
		},
		{
			name:     "High priority",
			priority: domain.NotificationPriorityHigh,
			title:    "Demand Forecast Update",
			expected: "⚠️ HIGH: Demand Forecast Update",
		},
		{
			name:     "Medium priority",
			priority: domain.NotificationPriorityMedium,
			title:    "Inventory Update",
			expected: "📌 Inventory Update",
		},
		{
			name:     "Low priority",
			priority: domain.NotificationPriorityLow,
			title:    "Weekly Summary",
			expected: "Weekly Summary",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			notification := &domain.AINotification{
				Title:    tc.title,
				Priority: tc.priority,
			}

			subject := client.buildSubject(notification)
			assert.Equal(t, tc.expected, subject)
		})
	}
}

func TestEmailClient_RenderDefaultHTML(t *testing.T) {
	mockLogger := new(MockLogger)

	client := &EmailClient{
		logger: mockLogger,
	}

	notification := &domain.AINotification{
		ID:           uuid.New(),
		Title:        "Critical Stock Alert",
		Summary:      "Stock level is critically low",
		FullAnalysis: "Detailed analysis here",
		Priority:     domain.NotificationPriorityCritical,
		CreatedAt:    time.Now(),
		Impact: domain.ImpactAssessment{
			RiskLevel:        "high",
			RevenueImpact:    15000.00,
			CostImpact:       5000.00,
			AffectedOrders:   25,
			AffectedProducts: 3,
		},
		Recommendations: []domain.Recommendation{
			{
				Action:          "Reorder immediately",
				Reasoning:       "Stock will run out in 2 days",
				ExpectedOutcome: "Prevent stockout",
				Effort:          "low",
				Impact:          "high",
				PriorityOrder:   1,
			},
		},
	}

	html := client.renderDefaultHTML(notification)

	// Verify HTML structure
	assert.Contains(t, html, "<!DOCTYPE html>")
	assert.Contains(t, html, notification.Title)
	assert.Contains(t, html, notification.Summary)
	assert.Contains(t, html, notification.FullAnalysis)
	assert.Contains(t, html, "GIIA AI Intelligence Hub")

	// Verify priority styling
	assert.Contains(t, html, "critical")

	// Verify impact section
	assert.Contains(t, html, "Impact Assessment")
	assert.Contains(t, html, "high")
	assert.Contains(t, html, "$15000.00")
	assert.Contains(t, html, "$5000.00")
	assert.Contains(t, html, "25")
	assert.Contains(t, html, "3")

	// Verify recommendations
	assert.Contains(t, html, "Recommended Actions")
	assert.Contains(t, html, "Reorder immediately")
}

func TestEmailClient_RenderHTMLTemplate(t *testing.T) {
	mockLogger := new(MockLogger)

	client := &EmailClient{
		logger:    mockLogger,
		templates: nil, // No custom templates
	}

	notification := &domain.AINotification{
		ID:        uuid.New(),
		Title:     "Test Notification",
		Summary:   "Test summary",
		Priority:  domain.NotificationPriorityMedium,
		CreatedAt: time.Now(),
	}

	html, err := client.renderHTMLTemplate(notification)

	require.NoError(t, err)
	assert.Contains(t, html, "<!DOCTYPE html>")
	assert.Contains(t, html, notification.Title)
}

func TestEmailClient_RenderImpactHTML(t *testing.T) {
	mockLogger := new(MockLogger)

	client := &EmailClient{
		logger: mockLogger,
	}

	t.Run("With impact data", func(t *testing.T) {
		notification := &domain.AINotification{
			Impact: domain.ImpactAssessment{
				RiskLevel:        "high",
				RevenueImpact:    10000.00,
				CostImpact:       2500.00,
				AffectedOrders:   15,
				AffectedProducts: 5,
			},
		}

		html := client.renderImpactHTML(notification)

		assert.Contains(t, html, "Impact Assessment")
		assert.Contains(t, html, "high")
		assert.Contains(t, html, "$10000.00")
		assert.Contains(t, html, "$2500.00")
		assert.Contains(t, html, "15")
		assert.Contains(t, html, "5")
	})

	t.Run("Without impact data", func(t *testing.T) {
		notification := &domain.AINotification{
			Impact: domain.ImpactAssessment{},
		}

		html := client.renderImpactHTML(notification)

		assert.Empty(t, html)
	})
}

func TestEmailClient_RenderRecommendationsHTML(t *testing.T) {
	mockLogger := new(MockLogger)

	client := &EmailClient{
		logger: mockLogger,
	}

	t.Run("With recommendations", func(t *testing.T) {
		notification := &domain.AINotification{
			Recommendations: []domain.Recommendation{
				{
					Action:          "Reorder product A",
					Reasoning:       "Low stock",
					ExpectedOutcome: "Prevent stockout",
					Effort:          "low",
					Impact:          "high",
					PriorityOrder:   1,
				},
				{
					Action:          "Review supplier B",
					Reasoning:       "Delivery delays",
					ExpectedOutcome: "Improve lead times",
					Effort:          "medium",
					Impact:          "medium",
					PriorityOrder:   2,
				},
			},
		}

		html := client.renderRecommendationsHTML(notification)

		assert.Contains(t, html, "Recommended Actions")
		assert.Contains(t, html, "Reorder product A")
		assert.Contains(t, html, "Low stock")
		assert.Contains(t, html, "Prevent stockout")
		assert.Contains(t, html, "Review supplier B")
	})

	t.Run("Without recommendations", func(t *testing.T) {
		notification := &domain.AINotification{
			Recommendations: []domain.Recommendation{},
		}

		html := client.renderRecommendationsHTML(notification)

		assert.Empty(t, html)
	})
}

func TestEmailClient_RenderTextTemplate(t *testing.T) {
	mockLogger := new(MockLogger)

	client := &EmailClient{
		logger: mockLogger,
	}

	notification := &domain.AINotification{
		ID:           uuid.New(),
		Title:        "Stock Alert",
		Summary:      "Low stock detected",
		FullAnalysis: "Detailed analysis of the situation",
		Priority:     domain.NotificationPriorityCritical,
		CreatedAt:    time.Now(),
		Impact: domain.ImpactAssessment{
			RiskLevel:        "high",
			RevenueImpact:    10000.00,
			CostImpact:       2500.00,
			AffectedOrders:   15,
			AffectedProducts: 5,
		},
		Recommendations: []domain.Recommendation{
			{
				Action:          "Reorder immediately",
				Reasoning:       "Stock running low",
				ExpectedOutcome: "Prevent stockout",
				Effort:          "low",
				Impact:          "high",
				PriorityOrder:   1,
			},
		},
	}

	text := client.renderTextTemplate(notification)

	// Verify text content
	assert.Contains(t, text, notification.Title)
	assert.Contains(t, text, notification.Summary)
	assert.Contains(t, text, notification.FullAnalysis)
	assert.Contains(t, text, "ANALYSIS:")
	assert.Contains(t, text, "IMPACT ASSESSMENT:")
	assert.Contains(t, text, "RECOMMENDED ACTIONS:")
	assert.Contains(t, text, "Reorder immediately")
}

func TestEmailClient_LoadTemplates(t *testing.T) {
	mockLogger := new(MockLogger)

	client := &EmailClient{
		logger: mockLogger,
	}

	// Load templates (currently sets to nil for default templates)
	client.loadTemplates()

	// Templates should be nil (using default)
	assert.Nil(t, client.templates)
}
