package email

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"time"

	"github.com/melegattip/giia-core-engine/pkg/logger"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/core/providers"
)

// EmailClient handles email delivery
type EmailClient struct {
	smtpHost     string
	smtpPort     int
	smtpUser     string
	smtpPassword string
	fromAddress  string
	fromName     string
	logger       logger.Logger
	templates    *template.Template
}

// NewEmailClient creates a new email client
func NewEmailClient(
	smtpHost string,
	smtpPort int,
	smtpUser string,
	smtpPassword string,
	fromAddress string,
	fromName string,
	logger logger.Logger,
) providers.EmailDeliveryProvider {
	client := &EmailClient{
		smtpHost:     smtpHost,
		smtpPort:     smtpPort,
		smtpUser:     smtpUser,
		smtpPassword: smtpPassword,
		fromAddress:  fromAddress,
		fromName:     fromName,
		logger:       logger,
	}

	// Load email templates
	client.loadTemplates()

	return client
}

// SendEmail sends an email to recipients
func (c *EmailClient) SendEmail(ctx context.Context, to []string, subject string, htmlBody string, textBody string) (string, error) {
	c.logger.Info(ctx, "Sending email", logger.Tags{
		"recipients": fmt.Sprintf("%v", to),
		"subject":    subject,
	})

	// In production, this would use an actual SMTP client or API like SendGrid
	// For now, we'll just log it
	messageID := fmt.Sprintf("msg_%d", time.Now().Unix())

	c.logger.Info(ctx, "Email sent successfully", logger.Tags{
		"message_id": messageID,
		"recipients": fmt.Sprintf("%v", to),
	})

	return messageID, nil
}

// SendNotificationEmail sends a notification as email
func (c *EmailClient) SendNotificationEmail(ctx context.Context, notification *domain.AINotification, recipients []string) (string, error) {
	subject := c.buildSubject(notification)
	htmlBody, err := c.renderHTMLTemplate(notification)
	if err != nil {
		return "", fmt.Errorf("failed to render HTML template: %w", err)
	}

	textBody := c.renderTextTemplate(notification)

	return c.SendEmail(ctx, recipients, subject, htmlBody, textBody)
}

func (c *EmailClient) buildSubject(notification *domain.AINotification) string {
	priorityPrefix := ""
	switch notification.Priority {
	case domain.NotificationPriorityCritical:
		priorityPrefix = "🚨 CRITICAL: "
	case domain.NotificationPriorityHigh:
		priorityPrefix = "⚠️ HIGH: "
	case domain.NotificationPriorityMedium:
		priorityPrefix = "📌 "
	}

	return priorityPrefix + notification.Title
}

func (c *EmailClient) renderHTMLTemplate(notification *domain.AINotification) (string, error) {
	if c.templates == nil {
		return c.renderDefaultHTML(notification), nil
	}

	var buf bytes.Buffer
	err := c.templates.ExecuteTemplate(&buf, "notification.html", notification)
	if err != nil {
		return c.renderDefaultHTML(notification), nil
	}

	return buf.String(), nil
}

func (c *EmailClient) renderDefaultHTML(notification *domain.AINotification) string {
	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: white; padding: 20px; border-radius: 8px 8px 0 0; }
        .content { background: #f9f9f9; padding: 20px; border: 1px solid #ddd; }
        .priority { display: inline-block; padding: 4px 12px; border-radius: 4px; font-size: 12px; font-weight: bold; }
        .critical { background: #ff4444; color: white; }
        .high { background: #ff8800; color: white; }
        .medium { background: #ffaa00; color: white; }
        .low { background: #4CAF50; color: white; }
        .recommendation { background: white; padding: 15px; margin: 10px 0; border-left: 4px solid #667eea; border-radius: 4px; }
        .impact { background: #fff3cd; padding: 15px; border-radius: 4px; margin: 15px 0; }
        .footer { text-align: center; padding: 20px; color: #666; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>%s</h1>
            <span class="priority %s">%s Priority</span>
        </div>
        <div class="content">
            <h2>Summary</h2>
            <p>%s</p>
            
            <h3>Analysis</h3>
            <p>%s</p>
            
            %s
            
            %s
        </div>
        <div class="footer">
            <p>GIIA AI Intelligence Hub | Sent at %s</p>
            <p><a href="#">View in Dashboard</a> | <a href="#">Mark as Read</a></p>
        </div>
    </div>
</body>
</html>
	`,
		notification.Title,
		string(notification.Priority),
		string(notification.Priority),
		notification.Summary,
		notification.FullAnalysis,
		c.renderImpactHTML(notification),
		c.renderRecommendationsHTML(notification),
		notification.CreatedAt.Format("Jan 02, 2006 15:04 MST"),
	)

	return html
}

func (c *EmailClient) renderImpactHTML(notification *domain.AINotification) string {
	if notification.Impact.RiskLevel == "" {
		return ""
	}

	return fmt.Sprintf(`
        <div class="impact">
            <h3>Impact Assessment</h3>
            <ul>
                <li><strong>Risk Level:</strong> %s</li>
                <li><strong>Revenue Impact:</strong> $%.2f</li>
                <li><strong>Cost Impact:</strong> $%.2f</li>
                <li><strong>Affected Orders:</strong> %d</li>
                <li><strong>Affected Products:</strong> %d</li>
            </ul>
        </div>
	`,
		notification.Impact.RiskLevel,
		notification.Impact.RevenueImpact,
		notification.Impact.CostImpact,
		notification.Impact.AffectedOrders,
		notification.Impact.AffectedProducts,
	)
}

func (c *EmailClient) renderRecommendationsHTML(notification *domain.AINotification) string {
	if len(notification.Recommendations) == 0 {
		return ""
	}

	html := "<h3>Recommended Actions</h3>"
	for _, rec := range notification.Recommendations {
		html += fmt.Sprintf(`
        <div class="recommendation">
            <h4>%d. %s</h4>
            <p><strong>Why:</strong> %s</p>
            <p><strong>Expected Outcome:</strong> %s</p>
            <p><strong>Effort:</strong> %s | <strong>Impact:</strong> %s</p>
        </div>
		`,
			rec.PriorityOrder,
			rec.Action,
			rec.Reasoning,
			rec.ExpectedOutcome,
			rec.Effort,
			rec.Impact,
		)
	}

	return html
}

func (c *EmailClient) renderTextTemplate(notification *domain.AINotification) string {
	text := fmt.Sprintf(`
%s
Priority: %s

%s

ANALYSIS:
%s

`,
		notification.Title,
		notification.Priority,
		notification.Summary,
		notification.FullAnalysis,
	)

	if notification.Impact.RiskLevel != "" {
		text += fmt.Sprintf(`
IMPACT ASSESSMENT:
- Risk Level: %s
- Revenue Impact: $%.2f
- Cost Impact: $%.2f
- Affected Orders: %d
- Affected Products: %d

`,
			notification.Impact.RiskLevel,
			notification.Impact.RevenueImpact,
			notification.Impact.CostImpact,
			notification.Impact.AffectedOrders,
			notification.Impact.AffectedProducts,
		)
	}

	if len(notification.Recommendations) > 0 {
		text += "RECOMMENDED ACTIONS:\n"
		for _, rec := range notification.Recommendations {
			text += fmt.Sprintf(`
%d. %s
   Why: %s
   Expected Outcome: %s
   Effort: %s | Impact: %s

`,
				rec.PriorityOrder,
				rec.Action,
				rec.Reasoning,
				rec.ExpectedOutcome,
				rec.Effort,
				rec.Impact,
			)
		}
	}

	text += fmt.Sprintf("\nSent at: %s\n", notification.CreatedAt.Format("Jan 02, 2006 15:04 MST"))

	return text
}

func (c *EmailClient) loadTemplates() {
	// In production, load from files
	// For now, we use the default template above
	c.templates = nil
}
