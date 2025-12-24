package providers

import (
	"context"

	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/core/domain"
)

// NotificationDeliveryChannel represents a delivery channel type
type NotificationDeliveryChannel string

const (
	ChannelEmail   NotificationDeliveryChannel = "email"
	ChannelWebhook NotificationDeliveryChannel = "webhook"
	ChannelSMS     NotificationDeliveryChannel = "sms"
	ChannelInApp   NotificationDeliveryChannel = "in_app"
)

// DeliveryRequest represents a notification delivery request
type DeliveryRequest struct {
	Notification *domain.AINotification
	Channels     []NotificationDeliveryChannel
	Recipients   []string // Email addresses, phone numbers, webhook URLs, etc.
	Metadata     map[string]string
}

// DeliveryResponse represents the result of a delivery attempt
type DeliveryResponse struct {
	Channel    NotificationDeliveryChannel
	Success    bool
	MessageID  string
	Error      error
	Recipient  string
	RetryCount int
}

// NotificationDeliveryService handles delivering notifications to various channels
type NotificationDeliveryService interface {
	// Deliver sends a notification through specified channels
	Deliver(ctx context.Context, request *DeliveryRequest) ([]*DeliveryResponse, error)

	// DeliverToChannel sends a notification through a specific channel
	DeliverToChannel(ctx context.Context, notification *domain.AINotification, channel NotificationDeliveryChannel, recipients []string) (*DeliveryResponse, error)
}

// EmailDeliveryProvider handles email delivery
type EmailDeliveryProvider interface {
	SendEmail(ctx context.Context, to []string, subject string, htmlBody string, textBody string) (messageID string, err error)
}

// WebhookDeliveryProvider handles webhook delivery
type WebhookDeliveryProvider interface {
	SendWebhook(ctx context.Context, webhookURL string, payload interface{}) error
}

// SMSDeliveryProvider handles SMS delivery
type SMSDeliveryProvider interface {
	SendSMS(ctx context.Context, to string, message string) (messageID string, err error)
}
