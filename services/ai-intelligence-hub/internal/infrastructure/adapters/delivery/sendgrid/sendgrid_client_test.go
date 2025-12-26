package sendgrid

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
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

func TestNewSendGridClient(t *testing.T) {
	mockLogger := new(MockLogger)
	config := &Config{
		APIKey:    "test-api-key",
		FromEmail: "test@example.com",
		FromName:  "Test Sender",
		Timeout:   10 * time.Second,
	}

	client := NewSendGridClient(config, mockLogger)

	assert.NotNil(t, client)
}

func TestSendGridClient_SendEmail_Success(t *testing.T) {
	mockHTTP := new(MockHTTPClient)
	mockLogger := new(MockLogger)

	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()

	config := &Config{
		APIKey:    "test-api-key",
		FromEmail: "test@example.com",
		FromName:  "Test Sender",
	}

	client := NewSendGridClientWithHTTPClient(config, mockHTTP, mockLogger)

	// Create mock response
	resp := &http.Response{
		StatusCode: 202,
		Body:       io.NopCloser(bytes.NewBufferString("")),
		Header:     make(http.Header),
	}
	resp.Header.Set("X-Message-Id", "test-message-id")

	mockHTTP.On("Do", mock.AnythingOfType("*http.Request")).Return(resp, nil)

	messageID, err := client.SendEmail(
		context.Background(),
		[]string{"recipient@example.com"},
		"Test Subject",
		"<p>HTML Body</p>",
		"Text Body",
	)

	require.NoError(t, err)
	assert.Equal(t, "test-message-id", messageID)
	mockHTTP.AssertExpectations(t)
}

func TestSendGridClient_SendEmail_NoRecipients(t *testing.T) {
	mockHTTP := new(MockHTTPClient)
	mockLogger := new(MockLogger)

	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()

	config := &Config{
		APIKey:    "test-api-key",
		FromEmail: "test@example.com",
		FromName:  "Test Sender",
	}

	client := NewSendGridClientWithHTTPClient(config, mockHTTP, mockLogger)

	_, err := client.SendEmail(
		context.Background(),
		[]string{},
		"Test Subject",
		"<p>HTML Body</p>",
		"Text Body",
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no recipients")
}

func TestSendGridClient_SendEmail_NoAPIKey(t *testing.T) {
	mockHTTP := new(MockHTTPClient)
	mockLogger := new(MockLogger)

	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()

	config := &Config{
		APIKey:    "",
		FromEmail: "test@example.com",
		FromName:  "Test Sender",
	}

	client := NewSendGridClientWithHTTPClient(config, mockHTTP, mockLogger)

	_, err := client.SendEmail(
		context.Background(),
		[]string{"recipient@example.com"},
		"Test Subject",
		"<p>HTML Body</p>",
		"Text Body",
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "API key not configured")
}

func TestSendGridClient_SendEmail_APIError(t *testing.T) {
	mockHTTP := new(MockHTTPClient)
	mockLogger := new(MockLogger)

	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()

	config := &Config{
		APIKey:    "test-api-key",
		FromEmail: "test@example.com",
		FromName:  "Test Sender",
	}

	client := NewSendGridClientWithHTTPClient(config, mockHTTP, mockLogger)

	// Create error response
	resp := &http.Response{
		StatusCode: 400,
		Body:       io.NopCloser(bytes.NewBufferString(`{"errors":[{"message":"Invalid email"}]}`)),
		Header:     make(http.Header),
	}

	mockHTTP.On("Do", mock.AnythingOfType("*http.Request")).Return(resp, nil)

	_, err := client.SendEmail(
		context.Background(),
		[]string{"invalid-email"},
		"Test Subject",
		"<p>HTML Body</p>",
		"Text Body",
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "SendGrid API error")
}

func TestSendGridClient_SendNotification_Success(t *testing.T) {
	mockHTTP := new(MockHTTPClient)
	mockLogger := new(MockLogger)

	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()

	config := &Config{
		APIKey:    "test-api-key",
		FromEmail: "test@example.com",
		FromName:  "Test Sender",
	}

	client := NewSendGridClientWithHTTPClient(config, mockHTTP, mockLogger)

	notification := &domain.AINotification{
		ID:             uuid.New(),
		OrganizationID: uuid.New(),
		UserID:         uuid.New(),
		Type:           domain.NotificationTypeAlert,
		Priority:       domain.NotificationPriorityCritical,
		Title:          "Critical Stock Alert",
		Summary:        "Stock level below threshold",
		FullAnalysis:   "Detailed analysis here",
		Impact: domain.ImpactAssessment{
			RiskLevel:        "high",
			RevenueImpact:    50000.00,
			AffectedOrders:   15,
			AffectedProducts: 3,
		},
		Recommendations: []domain.Recommendation{
			{
				Action:          "Reorder immediately",
				Reasoning:       "Current stock will run out in 2 days",
				ExpectedOutcome: "Prevent stockout",
				Effort:          "Low",
				Impact:          "High",
				PriorityOrder:   1,
			},
		},
		CreatedAt: time.Now(),
	}

	user := &domain.UserNotificationPreferences{
		EmailAddress: "user@example.com",
	}

	// Create mock response
	resp := &http.Response{
		StatusCode: 202,
		Body:       io.NopCloser(bytes.NewBufferString("")),
		Header:     make(http.Header),
	}
	resp.Header.Set("X-Message-Id", "test-message-id")

	mockHTTP.On("Do", mock.AnythingOfType("*http.Request")).Return(resp, nil)

	err := client.SendNotification(context.Background(), notification, user)

	require.NoError(t, err)
	mockHTTP.AssertExpectations(t)
}

func TestSendGridClient_SendDigest_Success(t *testing.T) {
	mockHTTP := new(MockHTTPClient)
	mockLogger := new(MockLogger)

	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()

	config := &Config{
		APIKey:    "test-api-key",
		FromEmail: "test@example.com",
		FromName:  "Test Sender",
	}

	client := NewSendGridClientWithHTTPClient(config, mockHTTP, mockLogger)

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
		CountByType: map[domain.NotificationType]int{
			domain.NotificationTypeAlert:   7,
			domain.NotificationTypeWarning: 8,
			domain.NotificationTypeInfo:    10,
		},
		UnactedCount: 5,
		TopItems: []*domain.AINotification{
			{
				ID:       uuid.New(),
				Title:    "Critical Alert 1",
				Summary:  "Summary 1",
				Priority: domain.NotificationPriorityCritical,
			},
			{
				ID:       uuid.New(),
				Title:    "High Priority Alert",
				Summary:  "Summary 2",
				Priority: domain.NotificationPriorityHigh,
			},
		},
	}

	user := &domain.UserNotificationPreferences{
		EmailAddress: "user@example.com",
	}

	// Create mock response
	resp := &http.Response{
		StatusCode: 202,
		Body:       io.NopCloser(bytes.NewBufferString("")),
		Header:     make(http.Header),
	}
	resp.Header.Set("X-Message-Id", "digest-message-id")

	mockHTTP.On("Do", mock.AnythingOfType("*http.Request")).Return(resp, nil)

	err := client.SendDigest(context.Background(), digest, user)

	require.NoError(t, err)
	mockHTTP.AssertExpectations(t)
}

func TestSendGridClient_BuildRequest(t *testing.T) {
	mockLogger := new(MockLogger)

	config := &Config{
		APIKey:    "test-api-key",
		FromEmail: "sender@example.com",
		FromName:  "Test Sender",
	}

	client := NewSendGridClientWithHTTPClient(config, nil, mockLogger)

	request := client.buildRequest(
		[]string{"recipient1@example.com", "recipient2@example.com"},
		"Test Subject",
		"<p>HTML</p>",
		"Plain text",
	)

	assert.Equal(t, "sender@example.com", request.From.Email)
	assert.Equal(t, "Test Sender", request.From.Name)
	assert.Equal(t, "Test Subject", request.Subject)
	assert.Len(t, request.Personalizations, 1)
	assert.Len(t, request.Personalizations[0].To, 2)
	assert.Len(t, request.Content, 2)
	assert.Equal(t, "text/plain", request.Content[0].Type)
	assert.Equal(t, "text/html", request.Content[1].Type)
}

func TestSendGridClient_BuildSubject(t *testing.T) {
	mockLogger := new(MockLogger)

	config := &Config{
		APIKey:    "test-api-key",
		FromEmail: "sender@example.com",
		FromName:  "Test Sender",
	}

	client := NewSendGridClientWithHTTPClient(config, nil, mockLogger)

	testCases := []struct {
		priority domain.NotificationPriority
		title    string
		expected string
	}{
		{domain.NotificationPriorityCritical, "Alert", "🚨 CRITICAL: Alert"},
		{domain.NotificationPriorityHigh, "Warning", "⚠️ HIGH: Warning"},
		{domain.NotificationPriorityMedium, "Info", "📌 Info"},
		{domain.NotificationPriorityLow, "Note", "ℹ️ Note"},
	}

	for _, tc := range testCases {
		notif := &domain.AINotification{
			Priority: tc.priority,
			Title:    tc.title,
		}
		subject := client.buildSubject(notif)
		assert.Equal(t, tc.expected, subject, "Failed for priority: %s", tc.priority)
	}
}

func TestSendGridClient_BuildHTMLBody(t *testing.T) {
	mockLogger := new(MockLogger)

	config := &Config{
		APIKey:    "test-api-key",
		FromEmail: "sender@example.com",
		FromName:  "Test Sender",
	}

	client := NewSendGridClientWithHTTPClient(config, nil, mockLogger)

	notification := &domain.AINotification{
		ID:             uuid.New(),
		OrganizationID: uuid.New(),
		UserID:         uuid.New(),
		Type:           domain.NotificationTypeAlert,
		Priority:       domain.NotificationPriorityCritical,
		Title:          "Test Alert",
		Summary:        "Test summary",
		FullAnalysis:   "Full analysis text",
		Impact: domain.ImpactAssessment{
			RiskLevel:        "high",
			RevenueImpact:    10000.00,
			AffectedOrders:   5,
			AffectedProducts: 2,
		},
		Recommendations: []domain.Recommendation{
			{
				Action:          "Take action",
				Reasoning:       "Because",
				ExpectedOutcome: "Good result",
				Effort:          "Low",
				Impact:          "High",
				PriorityOrder:   1,
			},
		},
		CreatedAt: time.Now(),
	}

	html := client.buildHTMLBody(notification)

	assert.Contains(t, html, "Test Alert")
	assert.Contains(t, html, "Test summary")
	assert.Contains(t, html, "Full analysis text")
	assert.Contains(t, html, "high")
	assert.Contains(t, html, "$10000.00")
	assert.Contains(t, html, "Take action")
}

func TestSendGridClient_GetPriorityColor(t *testing.T) {
	mockLogger := new(MockLogger)

	config := &Config{
		APIKey:    "test-api-key",
		FromEmail: "sender@example.com",
		FromName:  "Test Sender",
	}

	client := NewSendGridClientWithHTTPClient(config, nil, mockLogger)

	assert.Equal(t, "#dc3545", client.getPriorityColor(domain.NotificationPriorityCritical))
	assert.Equal(t, "#fd7e14", client.getPriorityColor(domain.NotificationPriorityHigh))
	assert.Equal(t, "#ffc107", client.getPriorityColor(domain.NotificationPriorityMedium))
	assert.Equal(t, "#28a745", client.getPriorityColor(domain.NotificationPriorityLow))
}

func TestSendGridClient_RequestFormat(t *testing.T) {
	mockHTTP := new(MockHTTPClient)
	mockLogger := new(MockLogger)

	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()

	config := &Config{
		APIKey:    "test-api-key",
		FromEmail: "test@example.com",
		FromName:  "Test Sender",
	}

	client := NewSendGridClientWithHTTPClient(config, mockHTTP, mockLogger)

	var capturedRequest *http.Request

	resp := &http.Response{
		StatusCode: 202,
		Body:       io.NopCloser(bytes.NewBufferString("")),
		Header:     make(http.Header),
	}
	resp.Header.Set("X-Message-Id", "test-id")

	mockHTTP.On("Do", mock.AnythingOfType("*http.Request")).Run(func(args mock.Arguments) {
		capturedRequest = args.Get(0).(*http.Request)
	}).Return(resp, nil)

	_, err := client.SendEmail(
		context.Background(),
		[]string{"recipient@example.com"},
		"Test Subject",
		"<p>HTML</p>",
		"Text",
	)

	require.NoError(t, err)
	require.NotNil(t, capturedRequest)

	// Verify request format
	assert.Equal(t, "POST", capturedRequest.Method)
	assert.Equal(t, "Bearer test-api-key", capturedRequest.Header.Get("Authorization"))
	assert.Equal(t, "application/json", capturedRequest.Header.Get("Content-Type"))
	assert.True(t, strings.Contains(capturedRequest.URL.String(), "sendgrid.com"))

	// Verify request body
	body, _ := io.ReadAll(capturedRequest.Body)
	var sgRequest SendGridRequest
	err = json.Unmarshal(body, &sgRequest)
	require.NoError(t, err)

	assert.Equal(t, "Test Subject", sgRequest.Subject)
	assert.Equal(t, "test@example.com", sgRequest.From.Email)
}
