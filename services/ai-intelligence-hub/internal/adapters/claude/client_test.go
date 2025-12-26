package claude

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/melegattip/giia-core-engine/pkg/events"
	"github.com/melegattip/giia-core-engine/pkg/logger"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/core/providers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockLogger implements logger.Logger for testing
type MockLogger struct{}

func (m *MockLogger) Debug(ctx context.Context, msg string, tags logger.Tags)            {}
func (m *MockLogger) Info(ctx context.Context, msg string, tags logger.Tags)             {}
func (m *MockLogger) Warn(ctx context.Context, msg string, tags logger.Tags)             {}
func (m *MockLogger) Error(ctx context.Context, err error, msg string, tags logger.Tags) {}
func (m *MockLogger) Fatal(ctx context.Context, err error, msg string, tags logger.Tags) {}

func TestNewClaudeClient(t *testing.T) {
	lg := &MockLogger{}

	tests := []struct {
		name          string
		config        ClaudeClientConfig
		expectedModel string
	}{
		{
			name:          "default model",
			config:        ClaudeClientConfig{APIKey: "test-key"},
			expectedModel: DefaultModel,
		},
		{
			name: "custom model",
			config: ClaudeClientConfig{
				APIKey: "test-key",
				Model:  HaikuModel,
			},
			expectedModel: HaikuModel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClaudeClient(tt.config, lg)
			require.NotNil(t, client)

			cc, ok := client.(*ClaudeClient)
			require.True(t, ok)
			assert.Equal(t, tt.expectedModel, cc.model)
		})
	}
}

func TestClaudeClient_Analyze_NilRequest(t *testing.T) {
	lg := &MockLogger{}
	client := NewClaudeClient(ClaudeClientConfig{APIKey: "test-key"}, lg)

	_, err := client.Analyze(context.Background(), nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be nil")
}

func TestClaudeClient_Analyze_NoAPIKey(t *testing.T) {
	lg := &MockLogger{}
	client := NewClaudeClient(ClaudeClientConfig{
		APIKey:          "",
		FallbackEnabled: true,
	}, lg)

	request := &providers.AIAnalysisRequest{
		Event: &events.Event{
			ID:        "test-event",
			Type:      "buffer.critical",
			Source:    "test",
			Timestamp: time.Now(),
			Data:      map[string]interface{}{"product_id": "prod-123"},
		},
		Context: map[string]interface{}{"test": "context"},
		Prompt:  "Analyze this event",
	}

	// Should trigger fallback
	response, err := client.Analyze(context.Background(), request)
	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.Contains(t, response.Summary, "fallback")
}

func TestClaudeClient_Analyze_MockServer(t *testing.T) {
	lg := &MockLogger{}

	// Create mock server
	mockResp := MessageResponse{
		ID:   "msg_123",
		Type: "message",
		Role: "assistant",
		Content: []ContentBlock{
			{
				Type: "text",
				Text: `{
					"summary": "Critical buffer status detected",
					"full_analysis": "The buffer has fallen below minimum threshold",
					"reasoning": "Based on DDMRP principles",
					"recommendations": [
						{
							"action": "Place emergency order",
							"reasoning": "Prevent stockout",
							"expected_outcome": "Buffer restored",
							"effort": "medium",
							"impact": "high"
						}
					],
					"impact_assessment": {
						"risk_level": "critical",
						"revenue_impact": 10000.00,
						"cost_impact": 200.00,
						"time_to_impact_hours": 24,
						"affected_orders": 5,
						"affected_products": 1
					},
					"confidence": 0.92
				}`,
			},
		},
		Model:      "claude-sonnet-4-20250514",
		StopReason: "end_turn",
		Usage: Usage{
			InputTokens:  100,
			OutputTokens: 200,
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "test-api-key", r.Header.Get("x-api-key"))
		assert.Equal(t, AnthropicVersion, r.Header.Get("anthropic-version"))

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResp)
	}))
	defer server.Close()

	client := NewClaudeClient(ClaudeClientConfig{
		APIKey:  "test-api-key",
		BaseURL: server.URL,
	}, lg)

	request := &providers.AIAnalysisRequest{
		Event: &events.Event{
			ID:        "test-event",
			Type:      "buffer.critical",
			Source:    "test",
			Timestamp: time.Now(),
			Data:      map[string]interface{}{"product_id": "prod-123"},
		},
		Context: map[string]interface{}{"test": "context"},
		Prompt:  "Analyze this event",
	}

	response, err := client.Analyze(context.Background(), request)
	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "Critical buffer status detected", response.Summary)
	assert.Equal(t, "critical", response.ImpactAssessment.RiskLevel)
	assert.Equal(t, 0.92, response.Confidence)
	assert.Len(t, response.Recommendations, 1)
	assert.Equal(t, "Place emergency order", response.Recommendations[0].Action)
}

func TestClaudeClient_Analyze_RateLimitRetry(t *testing.T) {
	lg := &MockLogger{}

	callCount := 0
	mockResp := MessageResponse{
		ID:   "msg_123",
		Type: "message",
		Role: "assistant",
		Content: []ContentBlock{
			{
				Type: "text",
				Text: `{"summary": "Success after retry", "full_analysis": "Test", "reasoning": "Test", "recommendations": [], "impact_assessment": {"risk_level": "low"}, "confidence": 0.8}`,
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount < 3 {
			// Return rate limit error
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"type": "error", "error": {"type": "rate_limit_error", "message": "Rate limited"}}`))
			return
		}
		// Success on third try
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResp)
	}))
	defer server.Close()

	client := NewClaudeClient(ClaudeClientConfig{
		APIKey:  "test-api-key",
		BaseURL: server.URL,
		Timeout: 5 * time.Second,
	}, lg)

	request := &providers.AIAnalysisRequest{
		Event: &events.Event{
			ID:        "test-event",
			Type:      "test",
			Source:    "test",
			Timestamp: time.Now(),
			Data:      map[string]interface{}{},
		},
	}

	response, err := client.Analyze(context.Background(), request)
	require.NoError(t, err)
	assert.Equal(t, "Success after retry", response.Summary)
	assert.Equal(t, 3, callCount)
}

func TestClaudeClient_Analyze_ServerError_FallbackEnabled(t *testing.T) {
	lg := &MockLogger{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"type": "error", "error": {"type": "server_error", "message": "Internal error"}}`))
	}))
	defer server.Close()

	client := NewClaudeClient(ClaudeClientConfig{
		APIKey:          "test-api-key",
		BaseURL:         server.URL,
		FallbackEnabled: true,
		Timeout:         2 * time.Second,
	}, lg)

	request := &providers.AIAnalysisRequest{
		Event: &events.Event{
			ID:        "test-event",
			Type:      "buffer.warning",
			Source:    "test",
			Timestamp: time.Now(),
			Data:      map[string]interface{}{},
		},
	}

	response, err := client.Analyze(context.Background(), request)
	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.Contains(t, response.Summary, "fallback")
}

func TestClaudeClient_FallbackAnalysis(t *testing.T) {
	lg := &MockLogger{}

	cc := &ClaudeClient{
		fallbackEnabled: true,
		logger:          lg,
	}

	tests := []struct {
		name          string
		eventType     string
		expectedRisk  string
		minConfidence float64
	}{
		{
			name:          "critical event",
			eventType:     "buffer.critical_stockout",
			expectedRisk:  "high",
			minConfidence: 0.7,
		},
		{
			name:          "warning event",
			eventType:     "buffer.warning",
			expectedRisk:  "medium",
			minConfidence: 0.65,
		},
		{
			name:          "info event",
			eventType:     "system.info",
			expectedRisk:  "low",
			minConfidence: 0.6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &providers.AIAnalysisRequest{
				Event: &events.Event{
					ID:        "test",
					Type:      tt.eventType,
					Source:    "test",
					Timestamp: time.Now(),
					Data:      map[string]interface{}{},
				},
			}

			response, err := cc.fallbackAnalysis(request)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedRisk, response.ImpactAssessment.RiskLevel)
			assert.GreaterOrEqual(t, response.Confidence, tt.minConfidence)
			assert.Len(t, response.Recommendations, 2)
		})
	}
}

func TestClaudeClient_ParseResponse_Valid(t *testing.T) {
	lg := &MockLogger{}
	cc := &ClaudeClient{logger: lg}

	responseText := `Some preamble text
{
	"summary": "Test summary",
	"full_analysis": "Test analysis",
	"reasoning": "Test reasoning",
	"recommendations": [
		{
			"action": "Do something",
			"reasoning": "Because",
			"expected_outcome": "Good result",
			"effort": "low",
			"impact": "high"
		}
	],
	"impact_assessment": {
		"risk_level": "medium",
		"revenue_impact": 5000.0,
		"cost_impact": 100.0,
		"time_to_impact_hours": 48,
		"affected_orders": 3,
		"affected_products": 2
	},
	"confidence": 0.85
}
Some postamble text`

	response, err := cc.parseResponse(responseText)
	require.NoError(t, err)
	assert.Equal(t, "Test summary", response.Summary)
	assert.Equal(t, "Test analysis", response.FullAnalysis)
	assert.Equal(t, "medium", response.ImpactAssessment.RiskLevel)
	assert.Equal(t, 5000.0, response.ImpactAssessment.RevenueImpact)
	assert.Equal(t, 48, response.ImpactAssessment.TimeToImpactHours)
	assert.Equal(t, 0.85, response.Confidence)
	require.Len(t, response.Recommendations, 1)
	assert.Equal(t, "Do something", response.Recommendations[0].Action)
}

func TestClaudeClient_ParseResponse_InvalidJSON(t *testing.T) {
	lg := &MockLogger{}
	cc := &ClaudeClient{logger: lg}

	_, err := cc.parseResponse("No JSON here")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid JSON")
}

func TestClaudeClient_ParseResponse_MissingSummary(t *testing.T) {
	lg := &MockLogger{}
	cc := &ClaudeClient{logger: lg}

	responseText := `{"full_analysis": "test", "reasoning": "test"}`

	_, err := cc.parseResponse(responseText)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing summary")
}

func TestAPIError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *APIError
		expected string
	}{
		{
			name: "with type",
			err: &APIError{
				StatusCode: 429,
				Type:       "rate_limit_error",
				Message:    "Too many requests",
			},
			expected: "Claude API error (429) rate_limit_error: Too many requests",
		},
		{
			name: "without type",
			err: &APIError{
				StatusCode: 500,
				Message:    "Internal server error",
			},
			expected: "Claude API error (500): Internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "rate limit error",
			err:      &APIError{StatusCode: 429},
			expected: true,
		},
		{
			name:     "server error 500",
			err:      &APIError{StatusCode: 500},
			expected: true,
		},
		{
			name:     "server error 503",
			err:      &APIError{StatusCode: 503},
			expected: true,
		},
		{
			name:     "bad request",
			err:      &APIError{StatusCode: 400},
			expected: false,
		},
		{
			name:     "unauthorized",
			err:      &APIError{StatusCode: 401},
			expected: false,
		},
		{
			name:     "network error",
			err:      assert.AnError,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, isRetryableError(tt.err))
		})
	}
}
