package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/melegattip/giia-core-engine/pkg/logger"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/core/providers"
)

const (
	defaultTimeout = 10 * time.Second
)

// SlackClient handles Slack webhook delivery
type SlackClient struct {
	webhookURL string
	httpClient HTTPClient
	logger     logger.Logger
}

// HTTPClient interface for testing
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Config holds Slack client configuration
type Config struct {
	WebhookURL string
	Timeout    time.Duration
}

// NewSlackClient creates a new Slack client
func NewSlackClient(config *Config, log logger.Logger) providers.WebhookDeliveryProvider {
	timeout := defaultTimeout
	if config.Timeout > 0 {
		timeout = config.Timeout
	}

	return &SlackClient{
		webhookURL: config.WebhookURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		logger: log,
	}
}

// NewSlackClientWithHTTPClient creates a client with custom HTTP client (for testing)
func NewSlackClientWithHTTPClient(config *Config, httpClient HTTPClient, log logger.Logger) *SlackClient {
	return &SlackClient{
		webhookURL: config.WebhookURL,
		httpClient: httpClient,
		logger:     log,
	}
}

// SlackMessage represents a Slack message payload
type SlackMessage struct {
	Text        string       `json:"text,omitempty"`
	Blocks      []Block      `json:"blocks,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

// Block represents a Slack Block Kit block
type Block struct {
	Type     string   `json:"type"`
	Text     *Text    `json:"text,omitempty"`
	Fields   []Text   `json:"fields,omitempty"`
	Elements []Button `json:"elements,omitempty"`
	BlockID  string   `json:"block_id,omitempty"`
}

// Text represents a Slack text object
type Text struct {
	Type  string `json:"type"`
	Text  string `json:"text"`
	Emoji bool   `json:"emoji,omitempty"`
}

// Button represents a Slack button element
type Button struct {
	Type     string `json:"type"`
	Text     Text   `json:"text"`
	ActionID string `json:"action_id,omitempty"`
	URL      string `json:"url,omitempty"`
	Style    string `json:"style,omitempty"`
}

// Attachment represents a Slack attachment (legacy but still useful for colors)
type Attachment struct {
	Color    string `json:"color"`
	Fallback string `json:"fallback"`
}

// SendWebhook sends a generic webhook payload to a Slack webhook
func (c *SlackClient) SendWebhook(ctx context.Context, webhookURL string, payload interface{}) error {
	c.logger.Info(ctx, "Sending Slack webhook", logger.Tags{
		"url": sanitizeURL(webhookURL),
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

	startTime := time.Now()
	resp, err := c.httpClient.Do(req)
	latency := time.Since(startTime)

	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Slack API error (status %d): %s", resp.StatusCode, string(body))
	}

	c.logger.Info(ctx, "Slack webhook sent successfully", logger.Tags{
		"url":        sanitizeURL(webhookURL),
		"status":     fmt.Sprintf("%d", resp.StatusCode),
		"latency_ms": fmt.Sprintf("%d", latency.Milliseconds()),
	})

	return nil
}

// PostNotification sends a notification to a Slack channel using Block Kit
func (c *SlackClient) PostNotification(ctx context.Context, notif *domain.AINotification, channel string) error {
	webhookURL := channel
	if webhookURL == "" {
		webhookURL = c.webhookURL
	}

	if webhookURL == "" {
		return fmt.Errorf("no webhook URL configured")
	}

	message := c.buildBlockKitMessage(notif)
	return c.SendWebhook(ctx, webhookURL, message)
}

// PostDigest sends a digest to Slack
func (c *SlackClient) PostDigest(ctx context.Context, digest *domain.Digest, channel string) error {
	webhookURL := channel
	if webhookURL == "" {
		webhookURL = c.webhookURL
	}

	if webhookURL == "" {
		return fmt.Errorf("no webhook URL configured")
	}

	message := c.buildDigestMessage(digest)
	return c.SendWebhook(ctx, webhookURL, message)
}

func (c *SlackClient) buildBlockKitMessage(notif *domain.AINotification) *SlackMessage {
	emoji := c.getPriorityEmoji(notif.Priority)
	color := c.getPriorityColor(notif.Priority)

	blocks := make([]Block, 0)

	// Header block
	blocks = append(blocks, Block{
		Type: "header",
		Text: &Text{
			Type:  "plain_text",
			Text:  fmt.Sprintf("%s %s", emoji, notif.Title),
			Emoji: true,
		},
	})

	// Context block - priority and type
	priorityText := fmt.Sprintf("*Priority:* %s  |  *Type:* %s  |  *Status:* %s",
		string(notif.Priority),
		string(notif.Type),
		string(notif.Status))

	blocks = append(blocks, Block{
		Type: "section",
		Text: &Text{
			Type: "mrkdwn",
			Text: priorityText,
		},
	})

	// Divider
	blocks = append(blocks, Block{Type: "divider"})

	// Summary section
	blocks = append(blocks, Block{
		Type: "section",
		Text: &Text{
			Type: "mrkdwn",
			Text: fmt.Sprintf("*📋 Summary*\n%s", notif.Summary),
		},
	})

	// Analysis section (if present)
	if notif.FullAnalysis != "" {
		analysis := notif.FullAnalysis
		if len(analysis) > 500 {
			analysis = analysis[:497] + "..."
		}
		blocks = append(blocks, Block{
			Type: "section",
			Text: &Text{
				Type: "mrkdwn",
				Text: fmt.Sprintf("*🔍 Analysis*\n%s", analysis),
			},
		})
	}

	// Impact section (if present)
	if notif.Impact.RiskLevel != "" {
		impactFields := []Text{
			{Type: "mrkdwn", Text: fmt.Sprintf("*Risk Level*\n%s", notif.Impact.RiskLevel)},
			{Type: "mrkdwn", Text: fmt.Sprintf("*Revenue Impact*\n$%.2f", notif.Impact.RevenueImpact)},
			{Type: "mrkdwn", Text: fmt.Sprintf("*Affected Orders*\n%d", notif.Impact.AffectedOrders)},
			{Type: "mrkdwn", Text: fmt.Sprintf("*Affected Products*\n%d", notif.Impact.AffectedProducts)},
		}

		blocks = append(blocks, Block{
			Type:   "section",
			Fields: impactFields,
		})
	}

	// Recommendations (top 3)
	if len(notif.Recommendations) > 0 {
		blocks = append(blocks, Block{Type: "divider"})

		recText := "*💡 Recommended Actions*\n"
		maxRecs := 3
		if len(notif.Recommendations) < maxRecs {
			maxRecs = len(notif.Recommendations)
		}

		for i := 0; i < maxRecs; i++ {
			rec := notif.Recommendations[i]
			recText += fmt.Sprintf("*%d. %s*\n_%s_\n\n", rec.PriorityOrder, rec.Action, rec.ExpectedOutcome)
		}

		blocks = append(blocks, Block{
			Type: "section",
			Text: &Text{
				Type: "mrkdwn",
				Text: recText,
			},
		})
	}

	// Action buttons
	blocks = append(blocks, Block{
		Type: "actions",
		Elements: []Button{
			{
				Type: "button",
				Text: Text{Type: "plain_text", Text: "View in Dashboard", Emoji: true},
				URL:  fmt.Sprintf("https://giia.app/notifications/%s", notif.ID.String()),
			},
			{
				Type:     "button",
				Text:     Text{Type: "plain_text", Text: "Mark as Read", Emoji: true},
				ActionID: fmt.Sprintf("mark_read_%s", notif.ID.String()),
			},
		},
	})

	// Footer with timestamp
	blocks = append(blocks, Block{
		Type: "context",
		Elements: []Button{
			{
				Type: "mrkdwn",
				Text: Text{
					Type: "mrkdwn",
					Text: fmt.Sprintf("🕐 Sent at %s | GIIA AI Intelligence Hub", notif.CreatedAt.Format("Jan 02, 2006 15:04 MST")),
				},
			},
		},
	})

	return &SlackMessage{
		Text:   fmt.Sprintf("%s %s - %s", emoji, notif.Title, notif.Summary),
		Blocks: blocks,
		Attachments: []Attachment{
			{
				Color:    color,
				Fallback: notif.Title,
			},
		},
	}
}

func (c *SlackClient) buildDigestMessage(digest *domain.Digest) *SlackMessage {
	blocks := make([]Block, 0)

	// Header
	blocks = append(blocks, Block{
		Type: "header",
		Text: &Text{
			Type:  "plain_text",
			Text:  "📊 Daily Notification Digest",
			Emoji: true,
		},
	})

	// Period
	blocks = append(blocks, Block{
		Type: "section",
		Text: &Text{
			Type: "mrkdwn",
			Text: fmt.Sprintf("📅 *Period:* %s to %s",
				digest.PeriodStart.Format("Jan 02"),
				digest.PeriodEnd.Format("Jan 02, 2006")),
		},
	})

	blocks = append(blocks, Block{Type: "divider"})

	// Stats fields
	statsFields := []Text{
		{Type: "mrkdwn", Text: fmt.Sprintf("*Total*\n%d", digest.TotalCount)},
		{Type: "mrkdwn", Text: fmt.Sprintf("*Unacted*\n%d", digest.UnactedCount)},
		{Type: "mrkdwn", Text: fmt.Sprintf("*🚨 Critical*\n%d", digest.CountByPriority[domain.NotificationPriorityCritical])},
		{Type: "mrkdwn", Text: fmt.Sprintf("*⚠️ High*\n%d", digest.CountByPriority[domain.NotificationPriorityHigh])},
	}

	blocks = append(blocks, Block{
		Type:   "section",
		Fields: statsFields,
	})

	// Top items
	if len(digest.TopItems) > 0 {
		blocks = append(blocks, Block{Type: "divider"})

		topText := "*🔝 Top Items*\n"
		maxItems := 5
		if len(digest.TopItems) < maxItems {
			maxItems = len(digest.TopItems)
		}

		for i := 0; i < maxItems; i++ {
			item := digest.TopItems[i]
			emoji := c.getPriorityEmoji(item.Priority)
			topText += fmt.Sprintf("%d. %s %s\n", i+1, emoji, item.Title)
		}

		blocks = append(blocks, Block{
			Type: "section",
			Text: &Text{
				Type: "mrkdwn",
				Text: topText,
			},
		})
	}

	// Action button
	blocks = append(blocks, Block{
		Type: "actions",
		Elements: []Button{
			{
				Type: "button",
				Text: Text{Type: "plain_text", Text: "View All Notifications", Emoji: true},
				URL:  "https://giia.app/notifications",
			},
		},
	})

	return &SlackMessage{
		Text:   fmt.Sprintf("📊 Daily Digest: %d notifications, %d unacted", digest.TotalCount, digest.UnactedCount),
		Blocks: blocks,
	}
}

func (c *SlackClient) getPriorityEmoji(priority domain.NotificationPriority) string {
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

func (c *SlackClient) getPriorityColor(priority domain.NotificationPriority) string {
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

// sanitizeURL removes sensitive parts of the URL for logging
func sanitizeURL(url string) string {
	if len(url) > 50 {
		// Keep only the domain part for logging
		if strings.Contains(url, "hooks.slack.com") {
			return "hooks.slack.com/***"
		}
		return url[:30] + "..."
	}
	return url
}
