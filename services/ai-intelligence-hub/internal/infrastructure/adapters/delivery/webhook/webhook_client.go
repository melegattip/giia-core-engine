package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/melegattip/giia-core-engine/pkg/logger"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/core/providers"
)

// WebhookClient handles webhook delivery
type WebhookClient struct {
	httpClient *http.Client
	logger     logger.Logger
}

// NewWebhookClient creates a new webhook client
func NewWebhookClient(logger logger.Logger) providers.WebhookDeliveryProvider {
	return &WebhookClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// SendWebhook sends a webhook to the specified URL
func (c *WebhookClient) SendWebhook(ctx context.Context, webhookURL string, payload interface{}) error {
	c.logger.Info(ctx, "Sending webhook", logger.Tags{
		"url": webhookURL,
	})

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "GIIA-Intelligence-Hub/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	c.logger.Info(ctx, "Webhook sent successfully", logger.Tags{
		"url":    webhookURL,
		"status": fmt.Sprintf("%d", resp.StatusCode),
	})

	return nil
}

// SendNotificationWebhook sends a notification as webhook
func (c *WebhookClient) SendNotificationWebhook(ctx context.Context, notification *domain.AINotification, webhookURL string) error {
	payload := c.buildWebhookPayload(notification)
	return c.SendWebhook(ctx, webhookURL, payload)
}

// SendSlackNotification sends a notification to Slack
func (c *WebhookClient) SendSlackNotification(ctx context.Context, notification *domain.AINotification, webhookURL string) error {
	payload := c.buildSlackPayload(notification)
	return c.SendWebhook(ctx, webhookURL, payload)
}

func (c *WebhookClient) buildWebhookPayload(notification *domain.AINotification) map[string]interface{} {
	payload := map[string]interface{}{
		"id":              notification.ID.String(),
		"organization_id": notification.OrganizationID.String(),
		"user_id":         notification.UserID.String(),
		"type":            notification.Type,
		"priority":        notification.Priority,
		"title":           notification.Title,
		"summary":         notification.Summary,
		"full_analysis":   notification.FullAnalysis,
		"status":          notification.Status,
		"created_at":      notification.CreatedAt,
	}

	if notification.Impact.RiskLevel != "" {
		payload["impact"] = map[string]interface{}{
			"risk_level":        notification.Impact.RiskLevel,
			"revenue_impact":    notification.Impact.RevenueImpact,
			"cost_impact":       notification.Impact.CostImpact,
			"affected_orders":   notification.Impact.AffectedOrders,
			"affected_products": notification.Impact.AffectedProducts,
		}
	}

	if len(notification.Recommendations) > 0 {
		recs := make([]map[string]interface{}, len(notification.Recommendations))
		for i, rec := range notification.Recommendations {
			recs[i] = map[string]interface{}{
				"action":           rec.Action,
				"reasoning":        rec.Reasoning,
				"expected_outcome": rec.ExpectedOutcome,
				"effort":           rec.Effort,
				"impact":           rec.Impact,
				"priority_order":   rec.PriorityOrder,
			}
		}
		payload["recommendations"] = recs
	}

	return payload
}

func (c *WebhookClient) buildSlackPayload(notification *domain.AINotification) map[string]interface{} {
	// Build Slack-formatted message
	color := c.getSlackColor(notification.Priority)
	emoji := c.getSlackEmoji(notification.Priority)

	fields := []map[string]interface{}{
		{
			"title": "Priority",
			"value": string(notification.Priority),
			"short": true,
		},
		{
			"title": "Type",
			"value": string(notification.Type),
			"short": true,
		},
	}

	if notification.Impact.RiskLevel != "" {
		fields = append(fields,
			map[string]interface{}{
				"title": "Risk Level",
				"value": notification.Impact.RiskLevel,
				"short": true,
			},
			map[string]interface{}{
				"title": "Revenue Impact",
				"value": fmt.Sprintf("$%.2f", notification.Impact.RevenueImpact),
				"short": true,
			},
		)
	}

	attachment := map[string]interface{}{
		"color":     color,
		"title":     emoji + " " + notification.Title,
		"text":      notification.Summary,
		"fields":    fields,
		"footer":    "GIIA AI Intelligence Hub",
		"ts":        notification.CreatedAt.Unix(),
		"mrkdwn_in": []string{"text", "fields"},
	}

	// Add recommendations as fields
	if len(notification.Recommendations) > 0 {
		recText := "*Recommended Actions:*\n"
		for _, rec := range notification.Recommendations {
			recText += fmt.Sprintf("%d. *%s*\n   _%s_\n", rec.PriorityOrder, rec.Action, rec.ExpectedOutcome)
		}
		fields = append(fields, map[string]interface{}{
			"title": "Recommendations",
			"value": recText,
			"short": false,
		})
		attachment["fields"] = fields
	}

	return map[string]interface{}{
		"text":        fmt.Sprintf("New %s notification", notification.Priority),
		"attachments": []interface{}{attachment},
	}
}

func (c *WebhookClient) getSlackColor(priority domain.NotificationPriority) string {
	switch priority {
	case domain.NotificationPriorityCritical:
		return "#ff0000" // Red
	case domain.NotificationPriorityHigh:
		return "#ff8800" // Orange
	case domain.NotificationPriorityMedium:
		return "#ffaa00" // Yellow
	case domain.NotificationPriorityLow:
		return "#4CAF50" // Green
	default:
		return "#667eea" // Default purple
	}
}

func (c *WebhookClient) getSlackEmoji(priority domain.NotificationPriority) string {
	switch priority {
	case domain.NotificationPriorityCritical:
		return "🚨"
	case domain.NotificationPriorityHigh:
		return "⚠️"
	case domain.NotificationPriorityMedium:
		return "📌"
	case domain.NotificationPriorityLow:
		return "ℹ️"
	default:
		return "📬"
	}
}
