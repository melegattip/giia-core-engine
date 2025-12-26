package sendgrid

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/melegattip/giia-core-engine/pkg/logger"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/core/providers"
)

const (
	sendGridAPIURL = "https://api.sendgrid.com/v3/mail/send"
	defaultTimeout = 30 * time.Second
)

// SendGridClient handles email delivery via SendGrid API
type SendGridClient struct {
	apiKey     string
	fromEmail  string
	fromName   string
	httpClient HTTPClient
	logger     logger.Logger
}

// HTTPClient interface for testing
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Config holds SendGrid client configuration
type Config struct {
	APIKey    string
	FromEmail string
	FromName  string
	Timeout   time.Duration
}

// NewSendGridClient creates a new SendGrid client
func NewSendGridClient(config *Config, log logger.Logger) providers.EmailDeliveryProvider {
	timeout := defaultTimeout
	if config.Timeout > 0 {
		timeout = config.Timeout
	}

	return &SendGridClient{
		apiKey:    config.APIKey,
		fromEmail: config.FromEmail,
		fromName:  config.FromName,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		logger: log,
	}
}

// NewSendGridClientWithHTTPClient creates a client with custom HTTP client (for testing)
func NewSendGridClientWithHTTPClient(config *Config, httpClient HTTPClient, log logger.Logger) *SendGridClient {
	return &SendGridClient{
		apiKey:     config.APIKey,
		fromEmail:  config.FromEmail,
		fromName:   config.FromName,
		httpClient: httpClient,
		logger:     log,
	}
}

// SendGridRequest represents a SendGrid API request
type SendGridRequest struct {
	Personalizations []Personalization `json:"personalizations"`
	From             EmailAddress      `json:"from"`
	Subject          string            `json:"subject"`
	Content          []Content         `json:"content"`
}

// Personalization represents email personalization
type Personalization struct {
	To      []EmailAddress `json:"to"`
	Subject string         `json:"subject,omitempty"`
}

// EmailAddress represents an email address
type EmailAddress struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

// Content represents email content
type Content struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

// SendEmail sends an email via SendGrid
func (c *SendGridClient) SendEmail(ctx context.Context, to []string, subject string, htmlBody string, textBody string) (string, error) {
	c.logger.Info(ctx, "Sending email via SendGrid", logger.Tags{
		"recipients": fmt.Sprintf("%v", to),
		"subject":    subject,
	})

	if len(to) == 0 {
		return "", fmt.Errorf("no recipients specified")
	}

	if c.apiKey == "" {
		return "", fmt.Errorf("SendGrid API key not configured")
	}

	// Build request
	request := c.buildRequest(to, subject, htmlBody, textBody)

	// Marshal to JSON
	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", sendGridAPIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "GIIA-Intelligence-Hub/1.0")

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("SendGrid API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Extract message ID from response header
	messageID := resp.Header.Get("X-Message-Id")
	if messageID == "" {
		messageID = fmt.Sprintf("sg_%d", time.Now().UnixNano())
	}

	c.logger.Info(ctx, "Email sent successfully via SendGrid", logger.Tags{
		"message_id": messageID,
		"recipients": fmt.Sprintf("%v", to),
	})

	return messageID, nil
}

// SendNotification sends a notification email with rich formatting
func (c *SendGridClient) SendNotification(ctx context.Context, notif *domain.AINotification, user *domain.UserNotificationPreferences) error {
	recipients := []string{user.EmailAddress}
	subject := c.buildSubject(notif)
	htmlBody := c.buildHTMLBody(notif)
	textBody := c.buildTextBody(notif)

	_, err := c.SendEmail(ctx, recipients, subject, htmlBody, textBody)
	return err
}

// SendDigest sends a daily digest email
func (c *SendGridClient) SendDigest(ctx context.Context, digest *domain.Digest, user *domain.UserNotificationPreferences) error {
	recipients := []string{user.EmailAddress}
	subject := fmt.Sprintf("📊 GIIA Daily Digest - %s", digest.GeneratedAt.Format("Jan 02, 2006"))

	// Ensure summary is generated
	if digest.SummaryText == "" {
		digest.GenerateSummary()
	}

	htmlBody := c.buildDigestHTMLBody(digest)
	textBody := digest.SummaryText

	_, err := c.SendEmail(ctx, recipients, subject, htmlBody, textBody)
	return err
}

func (c *SendGridClient) buildRequest(to []string, subject string, htmlBody string, textBody string) *SendGridRequest {
	toAddresses := make([]EmailAddress, len(to))
	for i, email := range to {
		toAddresses[i] = EmailAddress{Email: email}
	}

	content := make([]Content, 0, 2)
	if textBody != "" {
		content = append(content, Content{Type: "text/plain", Value: textBody})
	}
	if htmlBody != "" {
		content = append(content, Content{Type: "text/html", Value: htmlBody})
	}

	return &SendGridRequest{
		Personalizations: []Personalization{
			{
				To: toAddresses,
			},
		},
		From: EmailAddress{
			Email: c.fromEmail,
			Name:  c.fromName,
		},
		Subject: subject,
		Content: content,
	}
}

func (c *SendGridClient) buildSubject(notif *domain.AINotification) string {
	switch notif.Priority {
	case domain.NotificationPriorityCritical:
		return "🚨 CRITICAL: " + notif.Title
	case domain.NotificationPriorityHigh:
		return "⚠️ HIGH: " + notif.Title
	case domain.NotificationPriorityMedium:
		return "📌 " + notif.Title
	default:
		return "ℹ️ " + notif.Title
	}
}

func (c *SendGridClient) buildHTMLBody(notif *domain.AINotification) string {
	priorityColor := c.getPriorityColor(notif.Priority)
	priorityLabel := string(notif.Priority)

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body { font-family: 'Segoe UI', Arial, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; background: #f5f5f5; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: white; padding: 30px 20px; border-radius: 12px 12px 0 0; }
        .header h1 { margin: 0 0 10px 0; font-size: 24px; }
        .priority-badge { display: inline-block; padding: 6px 16px; border-radius: 20px; font-size: 12px; font-weight: bold; background: %s; color: white; }
        .content { background: white; padding: 30px; border: 1px solid #e0e0e0; border-top: none; }
        .section { margin-bottom: 25px; }
        .section h2 { color: #667eea; font-size: 18px; margin: 0 0 10px 0; border-bottom: 2px solid #667eea; padding-bottom: 5px; }
        .recommendation { background: #f8f9fa; padding: 15px; margin: 10px 0; border-left: 4px solid #667eea; border-radius: 4px; }
        .recommendation h3 { margin: 0 0 8px 0; color: #333; font-size: 16px; }
        .meta { color: #666; font-size: 14px; }
        .impact-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 10px; }
        .impact-item { background: #fff3cd; padding: 10px; border-radius: 4px; }
        .impact-label { font-size: 12px; color: #856404; }
        .impact-value { font-size: 18px; font-weight: bold; color: #333; }
        .footer { text-align: center; padding: 20px; color: #666; font-size: 12px; background: #f8f9fa; border-radius: 0 0 12px 12px; }
        .button { display: inline-block; padding: 12px 24px; background: #667eea; color: white; text-decoration: none; border-radius: 6px; font-weight: bold; margin: 5px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>%s</h1>
            <span class="priority-badge">%s Priority</span>
        </div>
        <div class="content">
            <div class="section">
                <h2>📋 Summary</h2>
                <p>%s</p>
            </div>
`, priorityColor, notif.Title, priorityLabel, notif.Summary)

	// Add analysis section
	if notif.FullAnalysis != "" {
		html += fmt.Sprintf(`
            <div class="section">
                <h2>🔍 Analysis</h2>
                <p>%s</p>
            </div>
`, notif.FullAnalysis)
	}

	// Add impact section
	if notif.Impact.RiskLevel != "" {
		html += fmt.Sprintf(`
            <div class="section">
                <h2>📊 Impact Assessment</h2>
                <div class="impact-grid">
                    <div class="impact-item">
                        <div class="impact-label">Risk Level</div>
                        <div class="impact-value">%s</div>
                    </div>
                    <div class="impact-item">
                        <div class="impact-label">Revenue Impact</div>
                        <div class="impact-value">$%.2f</div>
                    </div>
                    <div class="impact-item">
                        <div class="impact-label">Affected Orders</div>
                        <div class="impact-value">%d</div>
                    </div>
                    <div class="impact-item">
                        <div class="impact-label">Affected Products</div>
                        <div class="impact-value">%d</div>
                    </div>
                </div>
            </div>
`, notif.Impact.RiskLevel, notif.Impact.RevenueImpact, notif.Impact.AffectedOrders, notif.Impact.AffectedProducts)
	}

	// Add recommendations
	if len(notif.Recommendations) > 0 {
		html += `
            <div class="section">
                <h2>💡 Recommended Actions</h2>
`
		for _, rec := range notif.Recommendations {
			html += fmt.Sprintf(`
                <div class="recommendation">
                    <h3>%d. %s</h3>
                    <p><strong>Why:</strong> %s</p>
                    <p><strong>Expected Outcome:</strong> %s</p>
                    <p class="meta">Effort: %s | Impact: %s</p>
                </div>
`, rec.PriorityOrder, rec.Action, rec.Reasoning, rec.ExpectedOutcome, rec.Effort, rec.Impact)
		}
		html += `            </div>
`
	}

	// Footer
	html += fmt.Sprintf(`
        </div>
        <div class="footer">
            <p>GIIA AI Intelligence Hub</p>
            <p>Sent at %s</p>
            <a href="#" class="button">View in Dashboard</a>
            <a href="#" class="button">Mark as Read</a>
        </div>
    </div>
</body>
</html>
`, notif.CreatedAt.Format("Jan 02, 2006 15:04 MST"))

	return html
}

func (c *SendGridClient) buildTextBody(notif *domain.AINotification) string {
	text := fmt.Sprintf(`
%s
Priority: %s

SUMMARY:
%s

`, notif.Title, notif.Priority, notif.Summary)

	if notif.FullAnalysis != "" {
		text += fmt.Sprintf(`ANALYSIS:
%s

`, notif.FullAnalysis)
	}

	if notif.Impact.RiskLevel != "" {
		text += fmt.Sprintf(`IMPACT ASSESSMENT:
- Risk Level: %s
- Revenue Impact: $%.2f
- Cost Impact: $%.2f
- Affected Orders: %d
- Affected Products: %d

`, notif.Impact.RiskLevel, notif.Impact.RevenueImpact, notif.Impact.CostImpact, notif.Impact.AffectedOrders, notif.Impact.AffectedProducts)
	}

	if len(notif.Recommendations) > 0 {
		text += "RECOMMENDED ACTIONS:\n"
		for _, rec := range notif.Recommendations {
			text += fmt.Sprintf(`
%d. %s
   Why: %s
   Expected Outcome: %s
   Effort: %s | Impact: %s

`, rec.PriorityOrder, rec.Action, rec.Reasoning, rec.ExpectedOutcome, rec.Effort, rec.Impact)
		}
	}

	text += fmt.Sprintf("\nSent at: %s\n", notif.CreatedAt.Format("Jan 02, 2006 15:04 MST"))

	return text
}

func (c *SendGridClient) buildDigestHTMLBody(digest *domain.Digest) string {
	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: 'Segoe UI', Arial, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; background: #f5f5f5; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: white; padding: 30px 20px; border-radius: 12px 12px 0 0; text-align: center; }
        .header h1 { margin: 0; font-size: 28px; }
        .header p { margin: 10px 0 0 0; opacity: 0.9; }
        .content { background: white; padding: 30px; border: 1px solid #e0e0e0; border-top: none; }
        .stats-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 15px; margin-bottom: 25px; }
        .stat-card { background: #f8f9fa; padding: 20px; border-radius: 8px; text-align: center; }
        .stat-value { font-size: 32px; font-weight: bold; color: #667eea; }
        .stat-label { font-size: 12px; color: #666; text-transform: uppercase; }
        .priority-breakdown { margin: 25px 0; }
        .priority-row { display: flex; align-items: center; padding: 10px; border-radius: 4px; margin: 5px 0; }
        .priority-label { flex: 1; font-weight: bold; }
        .priority-count { padding: 4px 12px; border-radius: 20px; color: white; font-weight: bold; }
        .critical { background: #dc3545; }
        .high { background: #fd7e14; }
        .medium { background: #ffc107; color: #333; }
        .low { background: #28a745; }
        .top-items { margin-top: 25px; }
        .top-item { background: #f8f9fa; padding: 15px; margin: 10px 0; border-radius: 8px; border-left: 4px solid #667eea; }
        .footer { text-align: center; padding: 20px; color: #666; font-size: 12px; background: #f8f9fa; border-radius: 0 0 12px 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>📊 Daily Digest</h1>
            <p>%s to %s</p>
        </div>
        <div class="content">
            <div class="stats-grid">
                <div class="stat-card">
                    <div class="stat-value">%d</div>
                    <div class="stat-label">Total</div>
                </div>
                <div class="stat-card">
                    <div class="stat-value">%d</div>
                    <div class="stat-label">Unacted</div>
                </div>
                <div class="stat-card">
                    <div class="stat-value">%d</div>
                    <div class="stat-label">Critical</div>
                </div>
            </div>
`, digest.PeriodStart.Format("Jan 02"), digest.PeriodEnd.Format("Jan 02, 2006"),
		digest.TotalCount, digest.UnactedCount, digest.CountByPriority[domain.NotificationPriorityCritical])

	// Priority breakdown
	html += `            <div class="priority-breakdown">
                <h3>📊 By Priority</h3>
`
	priorityLevels := []struct {
		priority domain.NotificationPriority
		class    string
		emoji    string
	}{
		{domain.NotificationPriorityCritical, "critical", "🚨"},
		{domain.NotificationPriorityHigh, "high", "⚠️"},
		{domain.NotificationPriorityMedium, "medium", "📌"},
		{domain.NotificationPriorityLow, "low", "ℹ️"},
	}

	for _, p := range priorityLevels {
		count := digest.CountByPriority[p.priority]
		if count > 0 {
			html += fmt.Sprintf(`                <div class="priority-row">
                    <span class="priority-label">%s %s</span>
                    <span class="priority-count %s">%d</span>
                </div>
`, p.emoji, p.priority, p.class, count)
		}
	}
	html += `            </div>
`

	// Top items
	if len(digest.TopItems) > 0 {
		html += `            <div class="top-items">
                <h3>🔝 Top Items</h3>
`
		for i, item := range digest.TopItems {
			if i >= 5 {
				break
			}
			html += fmt.Sprintf(`                <div class="top-item">
                    <strong>%s</strong>
                    <p>%s</p>
                </div>
`, item.Title, item.Summary)
		}
		html += `            </div>
`
	}

	html += fmt.Sprintf(`        </div>
        <div class="footer">
            <p>GIIA AI Intelligence Hub</p>
            <p>Generated at %s</p>
        </div>
    </div>
</body>
</html>
`, digest.GeneratedAt.Format("Jan 02, 2006 15:04 MST"))

	return html
}

func (c *SendGridClient) getPriorityColor(priority domain.NotificationPriority) string {
	switch priority {
	case domain.NotificationPriorityCritical:
		return "#dc3545"
	case domain.NotificationPriorityHigh:
		return "#fd7e14"
	case domain.NotificationPriorityMedium:
		return "#ffc107"
	case domain.NotificationPriorityLow:
		return "#28a745"
	default:
		return "#667eea"
	}
}
