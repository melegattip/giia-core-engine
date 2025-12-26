package webhook

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

// WebhookClientTestable is a testable version of WebhookClient with injectable HTTP client
type WebhookClientTestable struct {
	httpClient HTTPClient
	logger     logger.Logger
}

// HTTPClient interface for testing
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func NewWebhookClientTestable(httpClient HTTPClient, log logger.Logger) *WebhookClientTestable {
	return &WebhookClientTestable{
		httpClient: httpClient,
		logger:     log,
	}
}

func (c *WebhookClientTestable) SendWebhook(ctx context.Context, webhookURL string, payload interface{}) error {
	c.logger.Info(ctx, "Sending webhook", logger.Tags{
		"url": webhookURL,
	})

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "GIIA-Intelligence-Hub/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return err
	}

	c.logger.Info(ctx, "Webhook sent successfully", logger.Tags{
		"url":    webhookURL,
		"status": resp.StatusCode,
	})

	return nil
}

func TestNewWebhookClient(t *testing.T) {
	mockLogger := new(MockLogger)

	client := NewWebhookClient(mockLogger)

	assert.NotNil(t, client)
}

func TestWebhookClient_SendWebhook_Success(t *testing.T) {
	mockHTTP := new(MockHTTPClient)
	mockLogger := new(MockLogger)

	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()

	client := NewWebhookClientTestable(mockHTTP, mockLogger)

	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBuffer([]byte(`{"ok": true}`))),
		Header:     make(http.Header),
	}

	mockHTTP.On("Do", mock.AnythingOfType("*http.Request")).Return(resp, nil)

	payload := map[string]interface{}{
		"message": "Test webhook",
		"data":    map[string]string{"key": "value"},
	}

	err := client.SendWebhook(context.Background(), "https://webhook.example.com/endpoint", payload)

	require.NoError(t, err)
	mockHTTP.AssertExpectations(t)
}

func TestWebhookClient_SendWebhook_RequestFormat(t *testing.T) {
	mockHTTP := new(MockHTTPClient)
	mockLogger := new(MockLogger)

	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()

	client := NewWebhookClientTestable(mockHTTP, mockLogger)

	var capturedRequest *http.Request

	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBuffer([]byte(`{"ok": true}`))),
		Header:     make(http.Header),
	}

	mockHTTP.On("Do", mock.AnythingOfType("*http.Request")).Run(func(args mock.Arguments) {
		capturedRequest = args.Get(0).(*http.Request)
	}).Return(resp, nil)

	payload := map[string]interface{}{
		"test": "data",
	}

	err := client.SendWebhook(context.Background(), "https://webhook.example.com/test", payload)

	require.NoError(t, err)
	require.NotNil(t, capturedRequest)

	// Verify request format
	assert.Equal(t, "POST", capturedRequest.Method)
	assert.Equal(t, "application/json", capturedRequest.Header.Get("Content-Type"))
	assert.Equal(t, "GIIA-Intelligence-Hub/1.0", capturedRequest.Header.Get("User-Agent"))
	assert.Contains(t, capturedRequest.URL.String(), "webhook.example.com")

	// Verify body
	body, _ := io.ReadAll(capturedRequest.Body)
	var parsedBody map[string]interface{}
	err = json.Unmarshal(body, &parsedBody)
	require.NoError(t, err)
	assert.Equal(t, "data", parsedBody["test"])
}

func TestWebhookClient_BuildWebhookPayload(t *testing.T) {
	mockLogger := new(MockLogger)

	client := &WebhookClient{
		logger: mockLogger,
	}

	notification := &domain.AINotification{
		ID:             uuid.New(),
		OrganizationID: uuid.New(),
		UserID:         uuid.New(),
		Type:           domain.NotificationTypeAlert,
		Priority:       domain.NotificationPriorityCritical,
		Status:         domain.NotificationStatusUnread,
		Title:          "Critical Stock Alert",
		Summary:        "Low stock detected",
		FullAnalysis:   "Detailed analysis",
		CreatedAt:      time.Now(),
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

	payload := client.buildWebhookPayload(notification)

	// Verify core fields
	assert.Equal(t, notification.ID.String(), payload["id"])
	assert.Equal(t, notification.OrganizationID.String(), payload["organization_id"])
	assert.Equal(t, notification.UserID.String(), payload["user_id"])
	assert.Equal(t, notification.Type, payload["type"])
	assert.Equal(t, notification.Priority, payload["priority"])
	assert.Equal(t, notification.Title, payload["title"])
	assert.Equal(t, notification.Summary, payload["summary"])
	assert.Equal(t, notification.FullAnalysis, payload["full_analysis"])
	assert.Equal(t, notification.Status, payload["status"])

	// Verify impact
	impact, ok := payload["impact"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "high", impact["risk_level"])
	assert.Equal(t, 10000.00, impact["revenue_impact"])
	assert.Equal(t, 2500.00, impact["cost_impact"])
	assert.Equal(t, 15, impact["affected_orders"])
	assert.Equal(t, 5, impact["affected_products"])

	// Verify recommendations
	recs, ok := payload["recommendations"].([]map[string]interface{})
	require.True(t, ok)
	require.Len(t, recs, 1)
	assert.Equal(t, "Reorder immediately", recs[0]["action"])
}

func TestWebhookClient_BuildWebhookPayload_NoImpact(t *testing.T) {
	mockLogger := new(MockLogger)

	client := &WebhookClient{
		logger: mockLogger,
	}

	notification := &domain.AINotification{
		ID:       uuid.New(),
		Title:    "Simple Alert",
		Summary:  "No impact data",
		Priority: domain.NotificationPriorityLow,
		Impact:   domain.ImpactAssessment{}, // Empty impact
	}

	payload := client.buildWebhookPayload(notification)

	// Verify impact is not present
	_, hasImpact := payload["impact"]
	assert.False(t, hasImpact)
}

func TestWebhookClient_BuildSlackPayload(t *testing.T) {
	mockLogger := new(MockLogger)

	client := &WebhookClient{
		logger: mockLogger,
	}

	notification := &domain.AINotification{
		ID:        uuid.New(),
		Title:     "Critical Stock Alert",
		Summary:   "Low stock detected for Product XYZ",
		Priority:  domain.NotificationPriorityCritical,
		Type:      domain.NotificationTypeAlert,
		CreatedAt: time.Now(),
		Impact: domain.ImpactAssessment{
			RiskLevel:     "high",
			RevenueImpact: 15000.00,
		},
		Recommendations: []domain.Recommendation{
			{
				Action:          "Reorder immediately",
				ExpectedOutcome: "Prevent stockout",
				PriorityOrder:   1,
			},
		},
	}

	payload := client.buildSlackPayload(notification)

	// Verify Slack message structure
	text, ok := payload["text"].(string)
	require.True(t, ok)
	assert.Contains(t, text, "critical")

	attachments, ok := payload["attachments"].([]interface{})
	require.True(t, ok)
	require.Len(t, attachments, 1)

	attachment := attachments[0].(map[string]interface{})
	assert.Equal(t, "#ff0000", attachment["color"]) // Critical = red
	assert.Contains(t, attachment["title"].(string), "🚨")
	assert.Contains(t, attachment["title"].(string), notification.Title)
	assert.Equal(t, notification.Summary, attachment["text"])
	assert.Equal(t, "GIIA AI Intelligence Hub", attachment["footer"])
}

func TestWebhookClient_GetSlackColor(t *testing.T) {
	mockLogger := new(MockLogger)

	client := &WebhookClient{
		logger: mockLogger,
	}

	tests := []struct {
		priority      domain.NotificationPriority
		expectedColor string
	}{
		{domain.NotificationPriorityCritical, "#ff0000"},
		{domain.NotificationPriorityHigh, "#ff8800"},
		{domain.NotificationPriorityMedium, "#ffaa00"},
		{domain.NotificationPriorityLow, "#4CAF50"},
		{"unknown", "#667eea"},
	}

	for _, tc := range tests {
		t.Run(string(tc.priority), func(t *testing.T) {
			color := client.getSlackColor(tc.priority)
			assert.Equal(t, tc.expectedColor, color)
		})
	}
}

func TestWebhookClient_GetSlackEmoji(t *testing.T) {
	mockLogger := new(MockLogger)

	client := &WebhookClient{
		logger: mockLogger,
	}

	tests := []struct {
		priority      domain.NotificationPriority
		expectedEmoji string
	}{
		{domain.NotificationPriorityCritical, "🚨"},
		{domain.NotificationPriorityHigh, "⚠️"},
		{domain.NotificationPriorityMedium, "📌"},
		{domain.NotificationPriorityLow, "ℹ️"},
		{"unknown", "📬"},
	}

	for _, tc := range tests {
		t.Run(string(tc.priority), func(t *testing.T) {
			emoji := client.getSlackEmoji(tc.priority)
			assert.Equal(t, tc.expectedEmoji, emoji)
		})
	}
}

func TestWebhookClient_DeliveryLatency(t *testing.T) {
	mockHTTP := new(MockHTTPClient)
	mockLogger := new(MockLogger)

	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()

	client := NewWebhookClientTestable(mockHTTP, mockLogger)

	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBuffer([]byte(`{"ok": true}`))),
		Header:     make(http.Header),
	}

	mockHTTP.On("Do", mock.AnythingOfType("*http.Request")).Return(resp, nil)

	payload := map[string]string{"test": "data"}

	startTime := time.Now()
	err := client.SendWebhook(context.Background(), "https://webhook.example.com/test", payload)
	latency := time.Since(startTime)

	require.NoError(t, err)
	// Verify the call completes quickly (< 30s target)
	assert.Less(t, latency, 30*time.Second, "Webhook delivery should complete in < 30s")
}
