package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/melegattip/giia-core-engine/pkg/logger"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/core/domain"
)

// MockHTTPClient is a mock HTTP client for testing
type MockHTTPClient struct {
	mock.Mock
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*http.Response), args.Error(1)
}

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

func TestNewSlackClient(t *testing.T) {
	mockLogger := new(MockLogger)
	config := &Config{
		WebhookURL: "https://hooks.slack.com/services/xxx/yyy/zzz",
		Timeout:    10 * time.Second,
	}

	client := NewSlackClient(config, mockLogger)

	assert.NotNil(t, client)
}

func TestSlackClient_SendWebhook_Success(t *testing.T) {
	mockHTTP := new(MockHTTPClient)
	mockLogger := new(MockLogger)

	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()

	config := &Config{
		WebhookURL: "https://hooks.slack.com/services/xxx/yyy/zzz",
	}

	client := NewSlackClientWithHTTPClient(config, mockHTTP, mockLogger)

	// Create mock response
	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString("ok")),
		Header:     make(http.Header),
	}

	mockHTTP.On("Do", mock.AnythingOfType("*http.Request")).Return(resp, nil)

	payload := map[string]interface{}{
		"text": "Test message",
	}

	err := client.SendWebhook(context.Background(), config.WebhookURL, payload)

	require.NoError(t, err)
	mockHTTP.AssertExpectations(t)
}

func TestSlackClient_SendWebhook_APIError(t *testing.T) {
	mockHTTP := new(MockHTTPClient)
	mockLogger := new(MockLogger)

	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()

	config := &Config{
		WebhookURL: "https://hooks.slack.com/services/xxx/yyy/zzz",
	}

	client := NewSlackClientWithHTTPClient(config, mockHTTP, mockLogger)

	// Create error response
	resp := &http.Response{
		StatusCode: 400,
		Body:       io.NopCloser(bytes.NewBufferString("invalid_payload")),
		Header:     make(http.Header),
	}

	mockHTTP.On("Do", mock.AnythingOfType("*http.Request")).Return(resp, nil)

	payload := map[string]interface{}{
		"text": "Test message",
	}

	err := client.SendWebhook(context.Background(), config.WebhookURL, payload)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "Slack API error")
}

func TestSlackClient_PostNotification_Success(t *testing.T) {
	mockHTTP := new(MockHTTPClient)
	mockLogger := new(MockLogger)

	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()

	config := &Config{
		WebhookURL: "https://hooks.slack.com/services/xxx/yyy/zzz",
	}

	client := NewSlackClientWithHTTPClient(config, mockHTTP, mockLogger)

	notification := &domain.AINotification{
		ID:             uuid.New(),
		OrganizationID: uuid.New(),
		UserID:         uuid.New(),
		Type:           domain.NotificationTypeAlert,
		Priority:       domain.NotificationPriorityCritical,
		Title:          "Critical Stock Alert",
		Summary:        "Stock level below threshold for Product XYZ",
		Status:         domain.NotificationStatusUnread,
		Impact: domain.ImpactAssessment{
			RiskLevel:        "high",
			RevenueImpact:    50000.00,
			AffectedOrders:   15,
			AffectedProducts: 3,
		},
		Recommendations: []domain.Recommendation{
			{
				Action:          "Reorder immediately",
				ExpectedOutcome: "Prevent stockout",
				PriorityOrder:   1,
			},
		},
		CreatedAt: time.Now(),
	}

	// Create mock response
	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString("ok")),
		Header:     make(http.Header),
	}

	mockHTTP.On("Do", mock.AnythingOfType("*http.Request")).Return(resp, nil)

	err := client.PostNotification(context.Background(), notification, "")

	require.NoError(t, err)
	mockHTTP.AssertExpectations(t)
}

func TestSlackClient_PostNotification_NoWebhookURL(t *testing.T) {
	mockHTTP := new(MockHTTPClient)
	mockLogger := new(MockLogger)

	config := &Config{
		WebhookURL: "",
	}

	client := NewSlackClientWithHTTPClient(config, mockHTTP, mockLogger)

	notification := &domain.AINotification{
		ID:       uuid.New(),
		Title:    "Test Alert",
		Priority: domain.NotificationPriorityMedium,
	}

	err := client.PostNotification(context.Background(), notification, "")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no webhook URL")
}

func TestSlackClient_PostNotification_WithCustomChannel(t *testing.T) {
	mockHTTP := new(MockHTTPClient)
	mockLogger := new(MockLogger)

	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()

	config := &Config{
		WebhookURL: "https://hooks.slack.com/default",
	}

	client := NewSlackClientWithHTTPClient(config, mockHTTP, mockLogger)

	notification := &domain.AINotification{
		ID:        uuid.New(),
		Title:     "Test Alert",
		Summary:   "Test summary",
		Priority:  domain.NotificationPriorityMedium,
		Status:    domain.NotificationStatusUnread,
		CreatedAt: time.Now(),
	}

	customWebhook := "https://hooks.slack.com/custom-channel"

	var capturedRequest *http.Request
	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString("ok")),
	}

	mockHTTP.On("Do", mock.AnythingOfType("*http.Request")).Run(func(args mock.Arguments) {
		capturedRequest = args.Get(0).(*http.Request)
	}).Return(resp, nil)

	err := client.PostNotification(context.Background(), notification, customWebhook)

	require.NoError(t, err)
	assert.Equal(t, customWebhook, capturedRequest.URL.String())
}

func TestSlackClient_PostDigest_Success(t *testing.T) {
	mockHTTP := new(MockHTTPClient)
	mockLogger := new(MockLogger)

	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()

	config := &Config{
		WebhookURL: "https://hooks.slack.com/services/xxx/yyy/zzz",
	}

	client := NewSlackClientWithHTTPClient(config, mockHTTP, mockLogger)

	digest := &domain.Digest{
		ID:             uuid.New(),
		OrganizationID: uuid.New(),
		UserID:         uuid.New(),
		GeneratedAt:    time.Now(),
		PeriodStart:    time.Now().Add(-24 * time.Hour),
		PeriodEnd:      time.Now(),
		TotalCount:     25,
		CountByPriority: map[domain.NotificationPriority]int{
			domain.NotificationPriorityCritical: 2,
			domain.NotificationPriorityHigh:     5,
			domain.NotificationPriorityMedium:   10,
			domain.NotificationPriorityLow:      8,
		},
		UnactedCount: 5,
		TopItems: []*domain.AINotification{
			{
				ID:       uuid.New(),
				Title:    "Critical Alert",
				Priority: domain.NotificationPriorityCritical,
			},
		},
	}

	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString("ok")),
	}

	mockHTTP.On("Do", mock.AnythingOfType("*http.Request")).Return(resp, nil)

	err := client.PostDigest(context.Background(), digest, "")

	require.NoError(t, err)
	mockHTTP.AssertExpectations(t)
}

func TestSlackClient_BuildBlockKitMessage(t *testing.T) {
	mockLogger := new(MockLogger)

	config := &Config{
		WebhookURL: "https://hooks.slack.com/services/xxx/yyy/zzz",
	}

	client := NewSlackClientWithHTTPClient(config, nil, mockLogger)

	notification := &domain.AINotification{
		ID:             uuid.New(),
		OrganizationID: uuid.New(),
		UserID:         uuid.New(),
		Type:           domain.NotificationTypeAlert,
		Priority:       domain.NotificationPriorityCritical,
		Title:          "Critical Alert",
		Summary:        "Test summary",
		FullAnalysis:   "Detailed analysis",
		Status:         domain.NotificationStatusUnread,
		Impact: domain.ImpactAssessment{
			RiskLevel:        "high",
			RevenueImpact:    10000.00,
			AffectedOrders:   5,
			AffectedProducts: 2,
		},
		Recommendations: []domain.Recommendation{
			{
				Action:          "Take action 1",
				ExpectedOutcome: "Good outcome",
				PriorityOrder:   1,
			},
			{
				Action:          "Take action 2",
				ExpectedOutcome: "Another outcome",
				PriorityOrder:   2,
			},
		},
		CreatedAt: time.Now(),
	}

	message := client.buildBlockKitMessage(notification)

	assert.NotNil(t, message)
	assert.NotEmpty(t, message.Text)
	assert.NotEmpty(t, message.Blocks)
	assert.NotEmpty(t, message.Attachments)

	// Verify attachments have correct color for critical priority
	assert.Equal(t, "#dc3545", message.Attachments[0].Color)

	// Verify blocks include header, sections, and actions
	hasHeader := false
	hasActions := false
	for _, block := range message.Blocks {
		if block.Type == "header" {
			hasHeader = true
		}
		if block.Type == "actions" {
			hasActions = true
		}
	}
	assert.True(t, hasHeader, "Should have header block")
	assert.True(t, hasActions, "Should have actions block")
}

func TestSlackClient_GetPriorityEmoji(t *testing.T) {
	mockLogger := new(MockLogger)

	config := &Config{}
	client := NewSlackClientWithHTTPClient(config, nil, mockLogger)

	assert.Equal(t, "🚨", client.getPriorityEmoji(domain.NotificationPriorityCritical))
	assert.Equal(t, "⚠️", client.getPriorityEmoji(domain.NotificationPriorityHigh))
	assert.Equal(t, "📌", client.getPriorityEmoji(domain.NotificationPriorityMedium))
	assert.Equal(t, "ℹ️", client.getPriorityEmoji(domain.NotificationPriorityLow))
}

func TestSlackClient_GetPriorityColor(t *testing.T) {
	mockLogger := new(MockLogger)

	config := &Config{}
	client := NewSlackClientWithHTTPClient(config, nil, mockLogger)

	assert.Equal(t, "#dc3545", client.getPriorityColor(domain.NotificationPriorityCritical))
	assert.Equal(t, "#fd7e14", client.getPriorityColor(domain.NotificationPriorityHigh))
	assert.Equal(t, "#ffc107", client.getPriorityColor(domain.NotificationPriorityMedium))
	assert.Equal(t, "#28a745", client.getPriorityColor(domain.NotificationPriorityLow))
}

func TestSlackClient_DeliveryLatency(t *testing.T) {
	mockHTTP := new(MockHTTPClient)
	mockLogger := new(MockLogger)

	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()

	config := &Config{
		WebhookURL: "https://hooks.slack.com/services/xxx/yyy/zzz",
	}

	client := NewSlackClientWithHTTPClient(config, mockHTTP, mockLogger)

	notification := &domain.AINotification{
		ID:        uuid.New(),
		Title:     "Test Alert",
		Summary:   "Test summary",
		Priority:  domain.NotificationPriorityHigh,
		Status:    domain.NotificationStatusUnread,
		CreatedAt: time.Now(),
	}

	// Simulate fast response
	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString("ok")),
	}

	mockHTTP.On("Do", mock.AnythingOfType("*http.Request")).Return(resp, nil)

	startTime := time.Now()
	err := client.PostNotification(context.Background(), notification, "")
	latency := time.Since(startTime)

	require.NoError(t, err)
	// Verify the call completes quickly (< 2s target)
	assert.Less(t, latency, 2*time.Second, "Delivery should complete in < 2s")
}

func TestSlackClient_MessageFormat(t *testing.T) {
	mockHTTP := new(MockHTTPClient)
	mockLogger := new(MockLogger)

	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()

	config := &Config{
		WebhookURL: "https://hooks.slack.com/services/xxx/yyy/zzz",
	}

	client := NewSlackClientWithHTTPClient(config, mockHTTP, mockLogger)

	notification := &domain.AINotification{
		ID:        uuid.New(),
		Title:     "Zone Status Alert",
		Summary:   "Product ABC is in RED zone",
		Priority:  domain.NotificationPriorityCritical,
		Status:    domain.NotificationStatusUnread,
		CreatedAt: time.Now(),
		Recommendations: []domain.Recommendation{
			{
				Action:          "Review inventory levels",
				ExpectedOutcome: "Prevent stockout",
				PriorityOrder:   1,
			},
		},
	}

	var capturedBody []byte
	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString("ok")),
	}

	mockHTTP.On("Do", mock.AnythingOfType("*http.Request")).Run(func(args mock.Arguments) {
		req := args.Get(0).(*http.Request)
		capturedBody, _ = io.ReadAll(req.Body)
	}).Return(resp, nil)

	err := client.PostNotification(context.Background(), notification, "")
	require.NoError(t, err)

	// Verify the message format
	var message SlackMessage
	err = json.Unmarshal(capturedBody, &message)
	require.NoError(t, err)

	// Should include product and zone status in text
	assert.Contains(t, message.Text, "Zone Status Alert")

	// Should include recommendations
	foundRecommendation := false
	for _, block := range message.Blocks {
		if block.Text != nil && block.Text.Type == "mrkdwn" {
			if bytes.Contains([]byte(block.Text.Text), []byte("Review inventory")) {
				foundRecommendation = true
				break
			}
		}
	}
	assert.True(t, foundRecommendation, "Should include recommendation in blocks")
}

func TestSanitizeURL(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"https://hooks.slack.com/services/T0123/B0123/xyzabc123", "hooks.slack.com/***"},
		{"https://example.com/short", "https://example.com/short"},
		{"https://other-webhook.com/verylongpathherethatexceedsfiftycharacters", "https://other-webhook.com/very..."},
	}

	for _, tc := range tests {
		result := sanitizeURL(tc.input)
		assert.Equal(t, tc.expected, result)
	}
}
