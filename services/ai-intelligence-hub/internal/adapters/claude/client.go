package claude

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/melegattip/giia-core-engine/pkg/errors"
	"github.com/melegattip/giia-core-engine/pkg/logger"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/core/providers"
)

const (
	DefaultBaseURL   = "https://api.anthropic.com/v1/messages"
	DefaultModel     = "claude-sonnet-4-20250514"
	HaikuModel       = "claude-3-5-haiku-20241022"
	DefaultMaxTokens = 4096
	DefaultTimeout   = 30 * time.Second
	MaxRetries       = 3
	InitialBackoffMs = 500
	AnthropicVersion = "2023-06-01"
)

// ClaudeClient implements the real Anthropic Claude API integration
type ClaudeClient struct {
	apiKey          string
	model           string
	baseURL         string
	httpClient      *http.Client
	logger          logger.Logger
	maxTokens       int
	promptBuilder   *PromptBuilder
	fallbackEnabled bool
}

// ClaudeClientConfig holds configuration for the Claude client
type ClaudeClientConfig struct {
	APIKey          string
	Model           string
	BaseURL         string
	Timeout         time.Duration
	MaxTokens       int
	FallbackEnabled bool
}

// MessageRequest represents a request to the Claude API
type MessageRequest struct {
	Model       string    `json:"model"`
	MaxTokens   int       `json:"max_tokens"`
	System      string    `json:"system,omitempty"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
}

// Message represents a message in the conversation
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// MessageResponse represents a response from the Claude API
type MessageResponse struct {
	ID           string         `json:"id"`
	Type         string         `json:"type"`
	Role         string         `json:"role"`
	Content      []ContentBlock `json:"content"`
	Model        string         `json:"model"`
	StopReason   string         `json:"stop_reason"`
	StopSequence string         `json:"stop_sequence,omitempty"`
	Usage        Usage          `json:"usage"`
}

// ContentBlock represents a block of content in the response
type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Usage represents token usage information
type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// ErrorResponse represents an error from the Claude API
type ErrorResponse struct {
	Type  string `json:"type"`
	Error struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

// NewClaudeClient creates a new Claude API client
func NewClaudeClient(config ClaudeClientConfig, log logger.Logger) providers.AIAnalyzer {
	if config.Model == "" {
		config.Model = DefaultModel
	}
	if config.BaseURL == "" {
		config.BaseURL = DefaultBaseURL
	}
	if config.Timeout == 0 {
		config.Timeout = DefaultTimeout
	}
	if config.MaxTokens == 0 {
		config.MaxTokens = DefaultMaxTokens
	}

	return &ClaudeClient{
		apiKey:    config.APIKey,
		model:     config.Model,
		baseURL:   config.BaseURL,
		maxTokens: config.MaxTokens,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		logger:          log,
		promptBuilder:   NewPromptBuilder(),
		fallbackEnabled: config.FallbackEnabled,
	}
}

// Analyze performs AI analysis using the Claude API
func (c *ClaudeClient) Analyze(ctx context.Context, request *providers.AIAnalysisRequest) (*providers.AIAnalysisResponse, error) {
	if request == nil {
		return nil, errors.NewBadRequest("analysis request cannot be nil")
	}

	// Build the prompt using the prompt builder
	systemPrompt, userPrompt := c.promptBuilder.Build(request)

	c.logger.Debug(ctx, "Sending request to Claude API", logger.Tags{
		"model":         c.model,
		"event_type":    request.Event.Type,
		"system_length": len(systemPrompt),
		"prompt_length": len(userPrompt),
	})

	// Call the API with retry logic
	responseText, err := c.callWithRetry(ctx, systemPrompt, userPrompt)
	if err != nil {
		// If API call fails and fallback is enabled, use rule-based analysis
		if c.fallbackEnabled {
			c.logger.Warn(ctx, "Claude API failed, using fallback analysis", logger.Tags{
				"error": err.Error(),
			})
			return c.fallbackAnalysis(request)
		}
		return nil, err
	}

	// Parse the response
	response, err := c.parseResponse(responseText)
	if err != nil {
		if c.fallbackEnabled {
			c.logger.Warn(ctx, "Failed to parse Claude response, using fallback", logger.Tags{
				"error": err.Error(),
			})
			return c.fallbackAnalysis(request)
		}
		return nil, err
	}

	c.logger.Info(ctx, "AI analysis completed", logger.Tags{
		"event_type": request.Event.Type,
		"confidence": response.Confidence,
		"risk_level": response.ImpactAssessment.RiskLevel,
	})

	return response, nil
}

// callWithRetry calls the Claude API with exponential backoff retry
func (c *ClaudeClient) callWithRetry(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	var lastErr error
	backoff := time.Duration(InitialBackoffMs) * time.Millisecond

	for attempt := 0; attempt < MaxRetries; attempt++ {
		if attempt > 0 {
			c.logger.Debug(ctx, "Retrying Claude API call", logger.Tags{
				"attempt": attempt + 1,
				"backoff": backoff.String(),
			})

			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(backoff):
			}
			backoff *= 2 // Exponential backoff
		}

		response, err := c.makeRequest(ctx, systemPrompt, userPrompt)
		if err == nil {
			return response, nil
		}

		lastErr = err

		// Check if error is retryable
		if !isRetryableError(err) {
			return "", err
		}
	}

	return "", fmt.Errorf("max retries exceeded: %w", lastErr)
}

// makeRequest makes a single request to the Claude API
func (c *ClaudeClient) makeRequest(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	// Check if API key is configured
	if c.apiKey == "" {
		return "", errors.NewBadRequest("Claude API key not configured")
	}

	// Build the request body
	reqBody := MessageRequest{
		Model:     c.model,
		MaxTokens: c.maxTokens,
		System:    systemPrompt,
		Messages: []Message{
			{Role: "user", Content: userPrompt},
		},
		Temperature: 0.3, // Lower temperature for more consistent analysis
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", AnthropicVersion)

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Handle error responses
	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err == nil {
			return "", &APIError{
				StatusCode: resp.StatusCode,
				Type:       errResp.Error.Type,
				Message:    errResp.Error.Message,
			}
		}
		return "", &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(body),
		}
	}

	// Parse successful response
	var msgResp MessageResponse
	if err := json.Unmarshal(body, &msgResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Extract text content
	var textContent strings.Builder
	for _, block := range msgResp.Content {
		if block.Type == "text" {
			textContent.WriteString(block.Text)
		}
	}

	return textContent.String(), nil
}

// parseResponse parses the Claude response into a structured AIAnalysisResponse
func (c *ClaudeClient) parseResponse(responseText string) (*providers.AIAnalysisResponse, error) {
	responseText = strings.TrimSpace(responseText)

	// Find JSON in the response
	start := strings.Index(responseText, "{")
	end := strings.LastIndex(responseText, "}")
	if start == -1 || end == -1 {
		return nil, errors.NewInternalServerError("invalid JSON response from AI")
	}

	jsonText := responseText[start : end+1]

	var parsed struct {
		Summary         string `json:"summary"`
		FullAnalysis    string `json:"full_analysis"`
		Reasoning       string `json:"reasoning"`
		Recommendations []struct {
			Action          string `json:"action"`
			Reasoning       string `json:"reasoning"`
			ExpectedOutcome string `json:"expected_outcome"`
			Effort          string `json:"effort"`
			Impact          string `json:"impact"`
		} `json:"recommendations"`
		ImpactAssessment struct {
			RiskLevel         string  `json:"risk_level"`
			RevenueImpact     float64 `json:"revenue_impact"`
			CostImpact        float64 `json:"cost_impact"`
			TimeToImpactHours int     `json:"time_to_impact_hours"`
			AffectedOrders    int     `json:"affected_orders"`
			AffectedProducts  int     `json:"affected_products"`
		} `json:"impact_assessment"`
		Confidence float64 `json:"confidence"`
	}

	if err := json.Unmarshal([]byte(jsonText), &parsed); err != nil {
		return nil, errors.NewInternalServerError(fmt.Sprintf("failed to parse AI response: %v", err))
	}

	if parsed.Summary == "" {
		return nil, errors.NewInternalServerError("AI response missing summary")
	}

	response := &providers.AIAnalysisResponse{
		Summary:      parsed.Summary,
		FullAnalysis: parsed.FullAnalysis,
		Reasoning:    parsed.Reasoning,
		Confidence:   parsed.Confidence,
		ImpactAssessment: providers.AIImpactAssessment{
			RiskLevel:         parsed.ImpactAssessment.RiskLevel,
			RevenueImpact:     parsed.ImpactAssessment.RevenueImpact,
			CostImpact:        parsed.ImpactAssessment.CostImpact,
			TimeToImpactHours: parsed.ImpactAssessment.TimeToImpactHours,
			AffectedOrders:    parsed.ImpactAssessment.AffectedOrders,
			AffectedProducts:  parsed.ImpactAssessment.AffectedProducts,
		},
	}

	for _, rec := range parsed.Recommendations {
		response.Recommendations = append(response.Recommendations, providers.AIRecommendation{
			Action:          rec.Action,
			Reasoning:       rec.Reasoning,
			ExpectedOutcome: rec.ExpectedOutcome,
			Effort:          rec.Effort,
			Impact:          rec.Impact,
		})
	}

	return response, nil
}

// fallbackAnalysis provides rule-based analysis when Claude API is unavailable
func (c *ClaudeClient) fallbackAnalysis(request *providers.AIAnalysisRequest) (*providers.AIAnalysisResponse, error) {
	// Determine risk level based on event type
	riskLevel := "medium"
	confidence := 0.6

	eventType := request.Event.Type
	if strings.Contains(eventType, "critical") || strings.Contains(eventType, "stockout") {
		riskLevel = "high"
		confidence = 0.75
	} else if strings.Contains(eventType, "warning") {
		riskLevel = "medium"
		confidence = 0.7
	} else if strings.Contains(eventType, "info") {
		riskLevel = "low"
		confidence = 0.65
	}

	return &providers.AIAnalysisResponse{
		Summary:      fmt.Sprintf("Automated analysis for %s event. AI service temporarily unavailable - using rule-based fallback.", eventType),
		FullAnalysis: "This is an automated fallback analysis generated when the AI service is unavailable. Please review the event details and take appropriate action based on your organization's standard operating procedures.",
		Reasoning:    "Fallback analysis based on event type classification. Full AI analysis will be available when the service is restored.",
		Recommendations: []providers.AIRecommendation{
			{
				Action:          "Review event details manually",
				Reasoning:       "AI analysis unavailable, manual review recommended",
				ExpectedOutcome: "Informed decision based on human expertise",
				Effort:          "medium",
				Impact:          "medium",
			},
			{
				Action:          "Monitor for similar events",
				Reasoning:       "Pattern detection may require multiple data points",
				ExpectedOutcome: "Better understanding of trend",
				Effort:          "low",
				Impact:          "medium",
			},
		},
		ImpactAssessment: providers.AIImpactAssessment{
			RiskLevel:         riskLevel,
			RevenueImpact:     0, // Cannot estimate without AI
			CostImpact:        0,
			TimeToImpactHours: 24, // Default assumption
			AffectedOrders:    1,
			AffectedProducts:  1,
		},
		Confidence: confidence,
	}, nil
}

// APIError represents an error from the Claude API
type APIError struct {
	StatusCode int
	Type       string
	Message    string
}

func (e *APIError) Error() string {
	if e.Type != "" {
		return fmt.Sprintf("Claude API error (%d) %s: %s", e.StatusCode, e.Type, e.Message)
	}
	return fmt.Sprintf("Claude API error (%d): %s", e.StatusCode, e.Message)
}

// isRetryableError determines if an error should be retried
func isRetryableError(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		// Retry on rate limit (429) and server errors (5xx)
		return apiErr.StatusCode == 429 || apiErr.StatusCode >= 500
	}
	// Retry on network errors
	return true
}
