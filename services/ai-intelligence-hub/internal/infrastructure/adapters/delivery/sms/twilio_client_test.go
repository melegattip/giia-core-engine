package sms

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

func TestNewTwilioClient(t *testing.T) {
	mockLogger := new(MockLogger)
	config := &Config{
		AccountSID: "AC123456",
		AuthToken:  "auth-token-123",
		FromNumber: "+15551234567",
		Timeout:    10 * time.Second,
	}

	client := NewTwilioClient(config, mockLogger)

	assert.NotNil(t, client)
}

func TestTwilioClient_SendSMS_Success(t *testing.T) {
	mockHTTP := new(MockHTTPClient)
	mockLogger := new(MockLogger)

	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()

	config := &Config{
		AccountSID: "AC123456",
		AuthToken:  "auth-token-123",
		FromNumber: "+15551234567",
	}

	client := NewTwilioClientWithHTTPClient(config, mockHTTP, mockLogger)

	// Create mock response
	twilioResp := TwilioResponse{
		SID:    "SM123456789",
		Status: "queued",
	}
	respBody, _ := json.Marshal(twilioResp)

	resp := &http.Response{
		StatusCode: 201,
		Body:       io.NopCloser(bytes.NewBuffer(respBody)),
		Header:     make(http.Header),
	}

	mockHTTP.On("Do", mock.AnythingOfType("*http.Request")).Return(resp, nil)

	messageID, err := client.SendSMS(context.Background(), "+15559876543", "Test message")

	require.NoError(t, err)
	assert.Equal(t, "SM123456789", messageID)
	mockHTTP.AssertExpectations(t)
}

func TestTwilioClient_SendSMS_MissingConfig(t *testing.T) {
	mockHTTP := new(MockHTTPClient)
	mockLogger := new(MockLogger)

	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()

	tests := []struct {
		name   string
		config *Config
		errMsg string
	}{
		{
			name:   "Missing AccountSID",
			config: &Config{AuthToken: "token", FromNumber: "+15551234567"},
			errMsg: "Account SID not configured",
		},
		{
			name:   "Missing AuthToken",
			config: &Config{AccountSID: "AC123", FromNumber: "+15551234567"},
			errMsg: "Auth Token not configured",
		},
		{
			name:   "Missing FromNumber",
			config: &Config{AccountSID: "AC123", AuthToken: "token"},
			errMsg: "From Number not configured",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client := NewTwilioClientWithHTTPClient(tc.config, mockHTTP, mockLogger)
			_, err := client.SendSMS(context.Background(), "+15559876543", "Test")
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.errMsg)
		})
	}
}

func TestTwilioClient_SendSMS_InvalidPhone(t *testing.T) {
	mockHTTP := new(MockHTTPClient)
	mockLogger := new(MockLogger)

	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()

	config := &Config{
		AccountSID: "AC123456",
		AuthToken:  "auth-token-123",
		FromNumber: "+15551234567",
	}

	client := NewTwilioClientWithHTTPClient(config, mockHTTP, mockLogger)

	_, err := client.SendSMS(context.Background(), "invalid", "Test")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid phone number")
}

func TestTwilioClient_SendSMS_EmptyMessage(t *testing.T) {
	mockHTTP := new(MockHTTPClient)
	mockLogger := new(MockLogger)

	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()

	config := &Config{
		AccountSID: "AC123456",
		AuthToken:  "auth-token-123",
		FromNumber: "+15551234567",
	}

	client := NewTwilioClientWithHTTPClient(config, mockHTTP, mockLogger)

	_, err := client.SendSMS(context.Background(), "+15559876543", "")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "message is required")
}

func TestTwilioClient_SendSMS_APIError(t *testing.T) {
	mockHTTP := new(MockHTTPClient)
	mockLogger := new(MockLogger)

	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()

	config := &Config{
		AccountSID: "AC123456",
		AuthToken:  "auth-token-123",
		FromNumber: "+15551234567",
	}

	client := NewTwilioClientWithHTTPClient(config, mockHTTP, mockLogger)

	// Create error response
	twilioResp := TwilioResponse{
		ErrorCode:    21211,
		ErrorMessage: "Invalid 'To' phone number",
	}
	respBody, _ := json.Marshal(twilioResp)

	resp := &http.Response{
		StatusCode: 400,
		Body:       io.NopCloser(bytes.NewBuffer(respBody)),
		Header:     make(http.Header),
	}

	mockHTTP.On("Do", mock.AnythingOfType("*http.Request")).Return(resp, nil)

	_, err := client.SendSMS(context.Background(), "+15559876543", "Test")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "Twilio API error")
}

func TestTwilioClient_SendNotification_CriticalOnly(t *testing.T) {
	mockHTTP := new(MockHTTPClient)
	mockLogger := new(MockLogger)

	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()

	config := &Config{
		AccountSID: "AC123456",
		AuthToken:  "auth-token-123",
		FromNumber: "+15551234567",
	}

	client := NewTwilioClientWithHTTPClient(config, mockHTTP, mockLogger)

	// Non-critical notification should be skipped
	notification := &domain.AINotification{
		ID:       uuid.New(),
		Title:    "Test Alert",
		Priority: domain.NotificationPriorityMedium,
	}

	err := client.SendNotification(context.Background(), notification, "+15559876543")

	require.NoError(t, err)
	// HTTP client should not have been called
	mockHTTP.AssertNotCalled(t, "Do", mock.Anything)
}

func TestTwilioClient_SendNotification_Critical_Success(t *testing.T) {
	mockHTTP := new(MockHTTPClient)
	mockLogger := new(MockLogger)

	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()

	config := &Config{
		AccountSID: "AC123456",
		AuthToken:  "auth-token-123",
		FromNumber: "+15551234567",
	}

	client := NewTwilioClientWithHTTPClient(config, mockHTTP, mockLogger)

	notification := &domain.AINotification{
		ID:       uuid.New(),
		Title:    "Critical Stock Alert",
		Summary:  "Immediate action required",
		Priority: domain.NotificationPriorityCritical,
		Recommendations: []domain.Recommendation{
			{Action: "Reorder now", PriorityOrder: 1},
		},
	}

	twilioResp := TwilioResponse{
		SID:    "SM123456789",
		Status: "queued",
	}
	respBody, _ := json.Marshal(twilioResp)

	resp := &http.Response{
		StatusCode: 201,
		Body:       io.NopCloser(bytes.NewBuffer(respBody)),
	}

	mockHTTP.On("Do", mock.AnythingOfType("*http.Request")).Return(resp, nil)

	err := client.SendNotification(context.Background(), notification, "+15559876543")

	require.NoError(t, err)
	mockHTTP.AssertExpectations(t)
}

func TestTwilioClient_SendDigest_Success(t *testing.T) {
	mockHTTP := new(MockHTTPClient)
	mockLogger := new(MockLogger)

	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()

	config := &Config{
		AccountSID: "AC123456",
		AuthToken:  "auth-token-123",
		FromNumber: "+15551234567",
	}

	client := NewTwilioClientWithHTTPClient(config, mockHTTP, mockLogger)

	digest := &domain.Digest{
		ID:         uuid.New(),
		TotalCount: 25,
		CountByPriority: map[domain.NotificationPriority]int{
			domain.NotificationPriorityCritical: 3,
			domain.NotificationPriorityHigh:     7,
		},
		UnactedCount: 5,
	}

	twilioResp := TwilioResponse{
		SID:    "SM987654321",
		Status: "queued",
	}
	respBody, _ := json.Marshal(twilioResp)

	resp := &http.Response{
		StatusCode: 201,
		Body:       io.NopCloser(bytes.NewBuffer(respBody)),
	}

	mockHTTP.On("Do", mock.AnythingOfType("*http.Request")).Return(resp, nil)

	err := client.SendDigest(context.Background(), digest, "+15559876543")

	require.NoError(t, err)
	mockHTTP.AssertExpectations(t)
}

func TestTwilioClient_BuildNotificationSMS(t *testing.T) {
	mockLogger := new(MockLogger)

	config := &Config{
		AccountSID: "AC123456",
		AuthToken:  "auth-token-123",
		FromNumber: "+15551234567",
	}

	client := NewTwilioClientWithHTTPClient(config, nil, mockLogger)

	notification := &domain.AINotification{
		ID:       uuid.New(),
		Title:    "Critical Stock Alert",
		Summary:  "Stock level for Product XYZ is below minimum threshold",
		Priority: domain.NotificationPriorityCritical,
		Recommendations: []domain.Recommendation{
			{Action: "Reorder immediately", PriorityOrder: 1},
		},
	}

	message := client.buildNotificationSMS(notification)

	assert.LessOrEqual(t, len(message), maxSMSLength, "Message should not exceed 160 chars")
	assert.Contains(t, message, "🚨")
	assert.Contains(t, message, "GIIA ALERT")
	assert.Contains(t, message, "Critical Stock Alert")
}

func TestTwilioClient_BuildDigestSMS(t *testing.T) {
	mockLogger := new(MockLogger)

	config := &Config{
		AccountSID: "AC123456",
		AuthToken:  "auth-token-123",
		FromNumber: "+15551234567",
	}

	client := NewTwilioClientWithHTTPClient(config, nil, mockLogger)

	digest := &domain.Digest{
		TotalCount: 25,
		CountByPriority: map[domain.NotificationPriority]int{
			domain.NotificationPriorityCritical: 3,
		},
		UnactedCount: 5,
	}

	message := client.buildDigestSMS(digest)

	assert.LessOrEqual(t, len(message), maxSMSLength, "Message should not exceed 160 chars")
	assert.Contains(t, message, "📊")
	assert.Contains(t, message, "25 notifications")
	assert.Contains(t, message, "3 critical")
	assert.Contains(t, message, "5 unacted")
}

func TestTruncateMessage(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"Short message", 160, "Short message"},
		{"This is a very long message that exceeds the limit", 30, "This is a very long messag..."},
		{"Exactly160charsExactly160charsExactly160charsExactly160charsExactly160charsExactly160charsExactly160charsExactly160charsExactly160charsExactly160chars12345678", 160, "Exactly160charsExactly160charsExactly160charsExactly160charsExactly160charsExactly160charsExactly160charsExactly160charsExactly160charsExactly160chars12345678"},
	}

	for _, tc := range tests {
		result := truncateMessage(tc.input, tc.maxLen)
		assert.LessOrEqual(t, len(result), tc.maxLen)
	}
}

func TestIsValidPhoneNumber(t *testing.T) {
	tests := []struct {
		phone    string
		expected bool
	}{
		{"+15551234567", true},
		{"15551234567", true},
		{"+1-555-123-4567", true},
		{"+1 555 123 4567", true},
		{"invalid", false},
		{"123", false},
		{"+abc123456789", false},
	}

	for _, tc := range tests {
		result := isValidPhoneNumber(tc.phone)
		assert.Equal(t, tc.expected, result, "Failed for phone: %s", tc.phone)
	}
}

func TestSanitizePhoneNumber(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"+15551234567", "+1****67"},
		{"12345", "12****45"},
		{"123", "****"},
	}

	for _, tc := range tests {
		result := sanitizePhoneNumber(tc.input)
		assert.Equal(t, tc.expected, result)
	}
}

func TestTwilioClient_DeliveryLatency(t *testing.T) {
	mockHTTP := new(MockHTTPClient)
	mockLogger := new(MockLogger)

	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()

	config := &Config{
		AccountSID: "AC123456",
		AuthToken:  "auth-token-123",
		FromNumber: "+15551234567",
	}

	client := NewTwilioClientWithHTTPClient(config, mockHTTP, mockLogger)

	twilioResp := TwilioResponse{
		SID:    "SM123456789",
		Status: "queued",
	}
	respBody, _ := json.Marshal(twilioResp)

	resp := &http.Response{
		StatusCode: 201,
		Body:       io.NopCloser(bytes.NewBuffer(respBody)),
	}

	mockHTTP.On("Do", mock.AnythingOfType("*http.Request")).Return(resp, nil)

	startTime := time.Now()
	_, err := client.SendSMS(context.Background(), "+15559876543", "Test message")
	latency := time.Since(startTime)

	require.NoError(t, err)
	// Verify the call completes quickly (< 30s target for critical alerts)
	assert.Less(t, latency, 30*time.Second, "SMS delivery should complete in < 30s")
}

func TestTwilioClient_GetDeliveryStatus(t *testing.T) {
	mockHTTP := new(MockHTTPClient)
	mockLogger := new(MockLogger)

	config := &Config{
		AccountSID: "AC123456",
		AuthToken:  "auth-token-123",
		FromNumber: "+15551234567",
	}

	client := NewTwilioClientWithHTTPClient(config, mockHTTP, mockLogger)

	twilioResp := TwilioResponse{
		SID:    "SM123456789",
		Status: "delivered",
	}
	respBody, _ := json.Marshal(twilioResp)

	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBuffer(respBody)),
	}

	mockHTTP.On("Do", mock.AnythingOfType("*http.Request")).Return(resp, nil)

	status, err := client.GetDeliveryStatus(context.Background(), "SM123456789")

	require.NoError(t, err)
	assert.Equal(t, "delivered", status)
}

func TestTwilioClient_RequestFormat(t *testing.T) {
	mockHTTP := new(MockHTTPClient)
	mockLogger := new(MockLogger)

	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()

	config := &Config{
		AccountSID: "AC123456",
		AuthToken:  "auth-token-123",
		FromNumber: "+15551234567",
	}

	client := NewTwilioClientWithHTTPClient(config, mockHTTP, mockLogger)

	var capturedRequest *http.Request

	twilioResp := TwilioResponse{SID: "SM123", Status: "queued"}
	respBody, _ := json.Marshal(twilioResp)

	resp := &http.Response{
		StatusCode: 201,
		Body:       io.NopCloser(bytes.NewBuffer(respBody)),
	}

	mockHTTP.On("Do", mock.AnythingOfType("*http.Request")).Run(func(args mock.Arguments) {
		capturedRequest = args.Get(0).(*http.Request)
	}).Return(resp, nil)

	_, err := client.SendSMS(context.Background(), "+15559876543", "Test message")

	require.NoError(t, err)
	require.NotNil(t, capturedRequest)

	// Verify request format
	assert.Equal(t, "POST", capturedRequest.Method)
	assert.Contains(t, capturedRequest.Header.Get("Authorization"), "Basic")
	assert.Equal(t, "application/x-www-form-urlencoded", capturedRequest.Header.Get("Content-Type"))
	assert.Contains(t, capturedRequest.URL.String(), "api.twilio.com")
	assert.Contains(t, capturedRequest.URL.String(), "AC123456")

	// Read and verify body
	body, _ := io.ReadAll(capturedRequest.Body)
	assert.Contains(t, string(body), "To=%2B15559876543")
	assert.Contains(t, string(body), "From=%2B15551234567")
	assert.Contains(t, string(body), "Body=Test+message")
}
