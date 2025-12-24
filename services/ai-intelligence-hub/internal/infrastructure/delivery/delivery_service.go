package delivery

import (
	"context"
	"fmt"

	"github.com/melegattip/giia-core-engine/pkg/logger"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/core/providers"
	emailAdapter "github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/infrastructure/adapters/delivery/email"
	webhookAdapter "github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/infrastructure/adapters/delivery/webhook"
)

// DeliveryService orchestrates notification delivery across multiple channels
type DeliveryService struct {
	emailProvider   providers.EmailDeliveryProvider
	webhookProvider providers.WebhookDeliveryProvider
	prefsRepo       providers.PreferencesRepository
	logger          logger.Logger
}

// NewDeliveryService creates a new delivery service
func NewDeliveryService(
	emailProvider providers.EmailDeliveryProvider,
	webhookProvider providers.WebhookDeliveryProvider,
	prefsRepo providers.PreferencesRepository,
	logger logger.Logger,
) providers.NotificationDeliveryService {
	return &DeliveryService{
		emailProvider:   emailProvider,
		webhookProvider: webhookProvider,
		prefsRepo:       prefsRepo,
		logger:          logger,
	}
}

// Deliver sends a notification through all specified channels
func (s *DeliveryService) Deliver(ctx context.Context, request *providers.DeliveryRequest) ([]*providers.DeliveryResponse, error) {
	s.logger.Info(ctx, "Delivering notification", logger.Tags{
		"notification_id": request.Notification.ID.String(),
		"channels":        fmt.Sprintf("%v", request.Channels),
	})

	responses := make([]*providers.DeliveryResponse, 0, len(request.Channels))

	for _, channel := range request.Channels {
		response, err := s.DeliverToChannel(ctx, request.Notification, channel, request.Recipients)
		if err != nil {
			s.logger.Error(ctx, err, "Failed to deliver to channel", logger.Tags{
				"channel":         string(channel),
				"notification_id": request.Notification.ID.String(),
			})
			response = &providers.DeliveryResponse{
				Channel: channel,
				Success: false,
				Error:   err,
			}
		}
		responses = append(responses, response)
	}

	return responses, nil
}

// DeliverToChannel sends a notification through a specific channel
func (s *DeliveryService) DeliverToChannel(
	ctx context.Context,
	notification *domain.AINotification,
	channel providers.NotificationDeliveryChannel,
	recipients []string,
) (*providers.DeliveryResponse, error) {
	response := &providers.DeliveryResponse{
		Channel: channel,
		Success: false,
	}

	switch channel {
	case providers.ChannelEmail:
		return s.deliverEmail(ctx, notification, recipients, response)

	case providers.ChannelWebhook:
		return s.deliverWebhook(ctx, notification, recipients, response)

	case providers.ChannelSMS:
		// SMS delivery would go here
		response.Error = fmt.Errorf("SMS delivery not implemented yet")
		return response, response.Error

	case providers.ChannelInApp:
		// In-app is handled by database storage (already done)
		response.Success = true
		return response, nil

	default:
		response.Error = fmt.Errorf("unknown channel: %s", channel)
		return response, response.Error
	}
}

func (s *DeliveryService) deliverEmail(
	ctx context.Context,
	notification *domain.AINotification,
	recipients []string,
	response *providers.DeliveryResponse,
) (*providers.DeliveryResponse, error) {
	if len(recipients) == 0 {
		response.Error = fmt.Errorf("no email recipients specified")
		return response, response.Error
	}

	// Use the email client's SendNotificationEmail method
	emailClient, ok := s.emailProvider.(*emailAdapter.EmailClient)
	if !ok {
		// Fallback if type assertion fails
		messageID, err := s.emailProvider.SendEmail(ctx, recipients, notification.Title, notification.Summary, notification.Summary)
		if err != nil {
			response.Error = err
			return response, err
		}
		response.MessageID = messageID
		response.Success = true
		response.Recipient = recipients[0]
		return response, nil
	}

	messageID, err := emailClient.SendNotificationEmail(ctx, notification, recipients)
	if err != nil {
		response.Error = err
		return response, err
	}

	response.MessageID = messageID
	response.Success = true
	response.Recipient = recipients[0]

	return response, nil
}

func (s *DeliveryService) deliverWebhook(
	ctx context.Context,
	notification *domain.AINotification,
	webhookURLs []string,
	response *providers.DeliveryResponse,
) (*providers.DeliveryResponse, error) {
	if len(webhookURLs) == 0 {
		response.Error = fmt.Errorf("no webhook URLs specified")
		return response, response.Error
	}

	// Use the webhook client's SendNotificationWebhook method
	webhookClient, ok := s.webhookProvider.(*webhookAdapter.WebhookClient)
	if !ok {
		// Fallback to generic webhook
		payload := map[string]interface{}{
			"id":       notification.ID.String(),
			"title":    notification.Title,
			"summary":  notification.Summary,
			"priority": notification.Priority,
		}
		err := s.webhookProvider.SendWebhook(ctx, webhookURLs[0], payload)
		if err != nil {
			response.Error = err
			return response, err
		}
		response.Success = true
		response.Recipient = webhookURLs[0]
		return response, nil
	}

	// For Slack webhooks (detected by slack.com in URL)
	for _, url := range webhookURLs {
		var err error
		if containsSlack(url) {
			err = webhookClient.SendSlackNotification(ctx, notification, url)
		} else {
			err = webhookClient.SendNotificationWebhook(ctx, notification, url)
		}

		if err != nil {
			response.Error = err
			return response, err
		}
	}

	response.Success = true
	response.Recipient = webhookURLs[0]

	return response, nil
}

// DeliverBasedOnPreferences delivers notification based on user preferences
func (s *DeliveryService) DeliverBasedOnPreferences(ctx context.Context, notification *domain.AINotification) ([]*providers.DeliveryResponse, error) {
	// Get user preferences
	prefs, err := s.prefsRepo.GetByUserID(ctx, notification.UserID, notification.OrganizationID)
	if err != nil {
		// If no preferences, deliver to in-app only
		s.logger.Warn(ctx, "No user preferences found, using defaults", logger.Tags{
			"user_id": notification.UserID.String(),
		})
		return s.deliverWithDefaults(ctx, notification)
	}

	// Build delivery request based on preferences
	request := &providers.DeliveryRequest{
		Notification: notification,
		Channels:     s.selectChannelsFromPreferences(prefs, notification.Priority),
		Recipients:   s.getRecipientsFromPreferences(prefs),
	}

	return s.Deliver(ctx, request)
}

func (s *DeliveryService) selectChannelsFromPreferences(
	prefs *domain.UserNotificationPreferences,
	priority domain.NotificationPriority,
) []providers.NotificationDeliveryChannel {
	channels := []providers.NotificationDeliveryChannel{providers.ChannelInApp} // Always include in-app

	// Check if email is enabled
	if prefs.EnableEmail {
		channels = append(channels, providers.ChannelEmail)
	}

	// Check if Slack is enabled
	if prefs.EnableSlack {
		channels = append(channels, providers.ChannelWebhook)
	}

	// SMS only for critical notifications if enabled
	if prefs.EnableSMS && priority == domain.NotificationPriorityCritical {
		channels = append(channels, providers.ChannelSMS)
	}

	return channels
}

func (s *DeliveryService) getRecipientsFromPreferences(prefs *domain.UserNotificationPreferences) []string {
	recipients := make([]string, 0)

	if prefs.EnableEmail && prefs.EmailAddress != "" {
		recipients = append(recipients, prefs.EmailAddress)
	}

	if prefs.EnableSlack && prefs.SlackWebhookURL != "" {
		recipients = append(recipients, prefs.SlackWebhookURL)
	}

	if prefs.EnableSMS && prefs.PhoneNumber != "" {
		recipients = append(recipients, prefs.PhoneNumber)
	}

	return recipients
}

func (s *DeliveryService) deliverWithDefaults(ctx context.Context, notification *domain.AINotification) ([]*providers.DeliveryResponse, error) {
	// Default: in-app only
	request := &providers.DeliveryRequest{
		Notification: notification,
		Channels:     []providers.NotificationDeliveryChannel{providers.ChannelInApp},
		Recipients:   []string{},
	}

	return s.Deliver(ctx, request)
}

func containsSlack(url string) bool {
	return len(url) > 0 && (url[:8] == "https://" || url[:7] == "http://") &&
		(len(url) > 20 && url[8:20] == "hooks.slack." || len(url) > 15 && url[8:15] == "slack.c")
}
