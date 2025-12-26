package sms

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/melegattip/giia-core-engine/pkg/logger"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/core/providers"
)

const (
	twilioAPIBaseURL = "https://api.twilio.com/2010-04-01"
	defaultTimeout   = 30 * time.Second
	maxSMSLength     = 160
)

// TwilioClient handles SMS delivery via Twilio API
type TwilioClient struct {
	accountSID string
	authToken  string
	fromNumber string
	httpClient HTTPClient
	logger     logger.Logger
}

// HTTPClient interface for testing
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Config holds Twilio client configuration
type Config struct {
	AccountSID string
	AuthToken  string
	FromNumber string
	Timeout    time.Duration
}

// TwilioResponse represents a Twilio API response
type TwilioResponse struct {
	SID          string `json:"sid"`
	Status       string `json:"status"`
	ErrorCode    int    `json:"error_code,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
}

// NewTwilioClient creates a new Twilio client
func NewTwilioClient(config *Config, log logger.Logger) providers.SMSDeliveryProvider {
	timeout := defaultTimeout
	if config.Timeout > 0 {
		timeout = config.Timeout
	}

	return &TwilioClient{
		accountSID: config.AccountSID,
		authToken:  config.AuthToken,
		fromNumber: config.FromNumber,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		logger: log,
	}
}

// NewTwilioClientWithHTTPClient creates a client with custom HTTP client (for testing)
func NewTwilioClientWithHTTPClient(config *Config, httpClient HTTPClient, log logger.Logger) *TwilioClient {
	return &TwilioClient{
		accountSID: config.AccountSID,
		authToken:  config.AuthToken,
		fromNumber: config.FromNumber,
		httpClient: httpClient,
		logger:     log,
	}
}

// SendSMS sends an SMS message via Twilio
func (c *TwilioClient) SendSMS(ctx context.Context, to string, message string) (string, error) {
	c.logger.Info(ctx, "Sending SMS via Twilio", logger.Tags{
		"to":          sanitizePhoneNumber(to),
		"message_len": fmt.Sprintf("%d", len(message)),
	})

	if err := c.validate(to, message); err != nil {
		return "", err
	}

	// Truncate message if needed
	truncatedMessage := truncateMessage(message, maxSMSLength)

	// Build request
	apiURL := fmt.Sprintf("%s/Accounts/%s/Messages.json", twilioAPIBaseURL, c.accountSID)

	formData := url.Values{}
	formData.Set("To", to)
	formData.Set("From", c.fromNumber)
	formData.Set("Body", truncatedMessage)

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set authentication
	auth := base64.StdEncoding.EncodeToString([]byte(c.accountSID + ":" + c.authToken))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "GIIA-Intelligence-Hub/1.0")

	// Send request
	startTime := time.Now()
	resp, err := c.httpClient.Do(req)
	latency := time.Since(startTime)

	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var twilioResp TwilioResponse
	if err := json.Unmarshal(body, &twilioResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for errors
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("Twilio API error (status %d): %s - %s",
			resp.StatusCode, twilioResp.ErrorMessage, string(body))
	}

	if twilioResp.ErrorCode != 0 {
		return "", fmt.Errorf("Twilio error %d: %s", twilioResp.ErrorCode, twilioResp.ErrorMessage)
	}

	c.logger.Info(ctx, "SMS sent successfully via Twilio", logger.Tags{
		"message_sid": twilioResp.SID,
		"status":      twilioResp.Status,
		"latency_ms":  fmt.Sprintf("%d", latency.Milliseconds()),
	})

	return twilioResp.SID, nil
}

// SendNotification sends a notification as SMS (for critical alerts only)
func (c *TwilioClient) SendNotification(ctx context.Context, notif *domain.AINotification, phoneNumber string) error {
	// Only send SMS for critical alerts
	if notif.Priority != domain.NotificationPriorityCritical {
		c.logger.Info(ctx, "Skipping SMS for non-critical notification", logger.Tags{
			"notification_id": notif.ID.String(),
			"priority":        string(notif.Priority),
		})
		return nil
	}

	message := c.buildNotificationSMS(notif)
	_, err := c.SendSMS(ctx, phoneNumber, message)
	return err
}

// SendDigest sends a digest summary as SMS
func (c *TwilioClient) SendDigest(ctx context.Context, digest *domain.Digest, phoneNumber string) error {
	message := c.buildDigestSMS(digest)
	_, err := c.SendSMS(ctx, phoneNumber, message)
	return err
}

func (c *TwilioClient) validate(to string, message string) error {
	if c.accountSID == "" {
		return fmt.Errorf("Twilio Account SID not configured")
	}
	if c.authToken == "" {
		return fmt.Errorf("Twilio Auth Token not configured")
	}
	if c.fromNumber == "" {
		return fmt.Errorf("Twilio From Number not configured")
	}
	if to == "" {
		return fmt.Errorf("recipient phone number is required")
	}
	if message == "" {
		return fmt.Errorf("message is required")
	}
	if !isValidPhoneNumber(to) {
		return fmt.Errorf("invalid phone number format")
	}
	return nil
}

func (c *TwilioClient) buildNotificationSMS(notif *domain.AINotification) string {
	emoji := c.getPriorityEmoji(notif.Priority)

	// Build compact SMS message (max 160 chars)
	message := fmt.Sprintf("%s GIIA ALERT: %s", emoji, notif.Title)

	// Add summary if space allows
	remainingChars := maxSMSLength - len(message) - 3 // 3 for " - "
	if remainingChars > 20 && notif.Summary != "" {
		summary := notif.Summary
		if len(summary) > remainingChars {
			summary = summary[:remainingChars-3] + "..."
		}
		message += " - " + summary
	}

	// Add top recommendation if space allows
	if len(notif.Recommendations) > 0 {
		remainingChars = maxSMSLength - len(message) - 10
		if remainingChars > 15 {
			action := notif.Recommendations[0].Action
			if len(action) > remainingChars {
				action = action[:remainingChars-3] + "..."
			}
			message += " Action: " + action
		}
	}

	return truncateMessage(message, maxSMSLength)
}

func (c *TwilioClient) buildDigestSMS(digest *domain.Digest) string {
	criticalCount := digest.CountByPriority[domain.NotificationPriorityCritical]

	message := fmt.Sprintf("📊 GIIA Digest: %d notifications", digest.TotalCount)

	if criticalCount > 0 {
		message += fmt.Sprintf(", %d critical", criticalCount)
	}

	if digest.UnactedCount > 0 {
		message += fmt.Sprintf(", %d unacted", digest.UnactedCount)
	}

	return truncateMessage(message, maxSMSLength)
}

func (c *TwilioClient) getPriorityEmoji(priority domain.NotificationPriority) string {
	switch priority {
	case domain.NotificationPriorityCritical:
		return "🚨"
	case domain.NotificationPriorityHigh:
		return "⚠️"
	default:
		return "📌"
	}
}

// truncateMessage ensures the message doesn't exceed the max length
func truncateMessage(message string, maxLength int) string {
	if len(message) <= maxLength {
		return message
	}
	return message[:maxLength-3] + "..."
}

// isValidPhoneNumber performs basic phone number validation
func isValidPhoneNumber(phone string) bool {
	if len(phone) < 10 {
		return false
	}

	// Allow + at the start
	startIdx := 0
	if phone[0] == '+' {
		startIdx = 1
	}

	// Check that remaining characters are digits
	for i := startIdx; i < len(phone); i++ {
		if phone[i] < '0' || phone[i] > '9' {
			// Allow spaces and dashes
			if phone[i] != ' ' && phone[i] != '-' {
				return false
			}
		}
	}

	return true
}

// sanitizePhoneNumber masks the phone number for logging
func sanitizePhoneNumber(phone string) string {
	if len(phone) <= 4 {
		return "****"
	}
	return phone[:2] + "****" + phone[len(phone)-2:]
}

// GetDeliveryStatus checks the delivery status of an SMS
func (c *TwilioClient) GetDeliveryStatus(ctx context.Context, messageSID string) (string, error) {
	apiURL := fmt.Sprintf("%s/Accounts/%s/Messages/%s.json", twilioAPIBaseURL, c.accountSID, messageSID)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	auth := base64.StdEncoding.EncodeToString([]byte(c.accountSID + ":" + c.authToken))
	req.Header.Set("Authorization", "Basic "+auth)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var twilioResp TwilioResponse
	if err := json.Unmarshal(body, &twilioResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	return twilioResp.Status, nil
}
