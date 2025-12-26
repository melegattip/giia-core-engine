package usecases

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/melegattip/giia-core-engine/pkg/logger"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/core/providers"
)

// DeliverNotificationUseCase handles the delivery of notifications across channels
type DeliverNotificationUseCase struct {
	emailProvider   providers.EmailDeliveryProvider
	webhookProvider providers.WebhookDeliveryProvider
	smsProvider     providers.SMSDeliveryProvider
	notifRepo       providers.NotificationRepository
	prefsRepo       providers.PreferencesRepository
	deliveryQueue   *domain.DeliveryQueue
	channelConfigs  map[uuid.UUID]*domain.ChannelConfigSet
	logger          logger.Logger
	mu              sync.RWMutex
}

// DeliverNotificationInput represents the input for notification delivery
type DeliverNotificationInput struct {
	Notification *domain.AINotification
	Channels     []domain.DeliveryChannel
	Recipients   map[domain.DeliveryChannel]string // Channel -> recipient (email, phone, webhook URL)
	ForceDeliver bool                              // Bypass quiet hours check
}

// DeliverNotificationOutput represents the result of notification delivery
type DeliverNotificationOutput struct {
	SuccessCount   int
	FailedCount    int
	Results        []ChannelDeliveryResult
	QueuedForRetry int
}

// ChannelDeliveryResult represents the result of delivery to a single channel
type ChannelDeliveryResult struct {
	Channel    domain.DeliveryChannel
	Success    bool
	MessageID  string
	Error      error
	Latency    time.Duration
	RetryCount int
}

// NewDeliverNotificationUseCase creates a new deliver notification use case
func NewDeliverNotificationUseCase(
	emailProvider providers.EmailDeliveryProvider,
	webhookProvider providers.WebhookDeliveryProvider,
	smsProvider providers.SMSDeliveryProvider,
	notifRepo providers.NotificationRepository,
	prefsRepo providers.PreferencesRepository,
	log logger.Logger,
) *DeliverNotificationUseCase {
	return &DeliverNotificationUseCase{
		emailProvider:   emailProvider,
		webhookProvider: webhookProvider,
		smsProvider:     smsProvider,
		notifRepo:       notifRepo,
		prefsRepo:       prefsRepo,
		deliveryQueue:   domain.NewDeliveryQueue(),
		channelConfigs:  make(map[uuid.UUID]*domain.ChannelConfigSet),
		logger:          log,
	}
}

// Execute delivers a notification to the specified channels
func (uc *DeliverNotificationUseCase) Execute(ctx context.Context, input *DeliverNotificationInput) (*DeliverNotificationOutput, error) {
	uc.logger.Info(ctx, "Executing deliver notification use case", logger.Tags{
		"notification_id": input.Notification.ID.String(),
		"channels":        fmt.Sprintf("%v", input.Channels),
	})

	output := &DeliverNotificationOutput{
		Results: make([]ChannelDeliveryResult, 0, len(input.Channels)),
	}

	// Get channel configs for the organization
	configSet := uc.getChannelConfigs(input.Notification.OrganizationID)

	for _, channel := range input.Channels {
		recipient := input.Recipients[channel]
		if recipient == "" {
			uc.logger.Warn(ctx, "No recipient for channel", logger.Tags{
				"channel": string(channel),
			})
			continue
		}

		// Check quiet hours (unless forced)
		if !input.ForceDeliver && configSet != nil {
			config := configSet.GetConfig(channel)
			if config != nil && config.IsInQuietHours(time.Now()) {
				uc.logger.Info(ctx, "Skipping delivery during quiet hours", logger.Tags{
					"channel": string(channel),
				})
				// Queue for later delivery
				uc.queueForRetry(input.Notification, channel, recipient, "quiet_hours")
				output.QueuedForRetry++
				continue
			}

			// Check rate limits
			if config != nil && !config.CanSend() {
				uc.logger.Warn(ctx, "Rate limit reached for channel", logger.Tags{
					"channel": string(channel),
				})
				uc.queueForRetry(input.Notification, channel, recipient, "rate_limit")
				output.QueuedForRetry++
				continue
			}
		}

		result := uc.deliverToChannel(ctx, input.Notification, channel, recipient)
		output.Results = append(output.Results, result)

		if result.Success {
			output.SuccessCount++
			// Increment usage counter
			if configSet != nil {
				if config := configSet.GetConfig(channel); config != nil {
					config.IncrementUsage()
				}
			}
		} else {
			output.FailedCount++
			// Queue for retry
			uc.queueForRetry(input.Notification, channel, recipient, result.Error.Error())
			output.QueuedForRetry++
		}
	}

	uc.logger.Info(ctx, "Notification delivery complete", logger.Tags{
		"notification_id":  input.Notification.ID.String(),
		"success_count":    fmt.Sprintf("%d", output.SuccessCount),
		"failed_count":     fmt.Sprintf("%d", output.FailedCount),
		"queued_for_retry": fmt.Sprintf("%d", output.QueuedForRetry),
	})

	return output, nil
}

// DeliverBasedOnPreferences delivers a notification based on user preferences
func (uc *DeliverNotificationUseCase) DeliverBasedOnPreferences(ctx context.Context, notification *domain.AINotification) (*DeliverNotificationOutput, error) {
	uc.logger.Info(ctx, "Delivering based on user preferences", logger.Tags{
		"notification_id": notification.ID.String(),
		"user_id":         notification.UserID.String(),
	})

	// Get user preferences
	prefs, err := uc.prefsRepo.GetByUserID(ctx, notification.UserID, notification.OrganizationID)
	if err != nil {
		uc.logger.Warn(ctx, "Failed to get user preferences, using defaults", logger.Tags{
			"user_id": notification.UserID.String(),
			"error":   err.Error(),
		})
		prefs = domain.NewUserPreferences(notification.UserID, notification.OrganizationID)
	}

	// Build channels and recipients based on preferences
	channels := make([]domain.DeliveryChannel, 0)
	recipients := make(map[domain.DeliveryChannel]string)

	// Check quiet hours
	if uc.isInQuietHours(prefs) && notification.Priority != domain.NotificationPriorityCritical {
		uc.logger.Info(ctx, "In quiet hours, queueing non-critical notification", logger.Tags{
			"notification_id": notification.ID.String(),
		})
		// Queue for later instead of delivering now
		return &DeliverNotificationOutput{
			QueuedForRetry: 1,
		}, nil
	}

	// Always include in-app
	channels = append(channels, domain.DeliveryChannelInApp)

	// Email
	if prefs.EnableEmail && prefs.EmailAddress != "" {
		if uc.shouldDeliverToChannel(notification.Priority, prefs.EmailMinPriority) {
			channels = append(channels, domain.DeliveryChannelEmail)
			recipients[domain.DeliveryChannelEmail] = prefs.EmailAddress
		}
	}

	// Slack
	if prefs.EnableSlack && prefs.SlackWebhookURL != "" {
		channels = append(channels, domain.DeliveryChannelSlack)
		recipients[domain.DeliveryChannelSlack] = prefs.SlackWebhookURL
	}

	// SMS - only for critical notifications
	if prefs.EnableSMS && prefs.PhoneNumber != "" {
		if uc.shouldDeliverToChannel(notification.Priority, prefs.SMSMinPriority) {
			channels = append(channels, domain.DeliveryChannelSMS)
			recipients[domain.DeliveryChannelSMS] = prefs.PhoneNumber
		}
	}

	input := &DeliverNotificationInput{
		Notification: notification,
		Channels:     channels,
		Recipients:   recipients,
	}

	return uc.Execute(ctx, input)
}

func (uc *DeliverNotificationUseCase) deliverToChannel(
	ctx context.Context,
	notification *domain.AINotification,
	channel domain.DeliveryChannel,
	recipient string,
) ChannelDeliveryResult {
	result := ChannelDeliveryResult{
		Channel: channel,
	}

	startTime := time.Now()

	switch channel {
	case domain.DeliveryChannelEmail:
		result = uc.deliverEmail(ctx, notification, recipient)
	case domain.DeliveryChannelSlack:
		result = uc.deliverSlack(ctx, notification, recipient)
	case domain.DeliveryChannelSMS:
		result = uc.deliverSMS(ctx, notification, recipient)
	case domain.DeliveryChannelWebhook:
		result = uc.deliverWebhook(ctx, notification, recipient)
	case domain.DeliveryChannelInApp:
		// In-app is handled by database storage
		result.Success = true
	default:
		result.Error = fmt.Errorf("unsupported channel: %s", channel)
	}

	result.Latency = time.Since(startTime)
	result.Channel = channel

	return result
}

func (uc *DeliverNotificationUseCase) deliverEmail(ctx context.Context, notification *domain.AINotification, recipient string) ChannelDeliveryResult {
	result := ChannelDeliveryResult{Channel: domain.DeliveryChannelEmail}

	if uc.emailProvider == nil {
		result.Error = fmt.Errorf("email provider not configured")
		return result
	}

	messageID, err := uc.emailProvider.SendEmail(
		ctx,
		[]string{recipient},
		uc.buildEmailSubject(notification),
		uc.buildEmailHTMLBody(notification),
		uc.buildEmailTextBody(notification),
	)

	if err != nil {
		result.Error = err
		uc.logger.Error(ctx, err, "Failed to send email", logger.Tags{
			"notification_id": notification.ID.String(),
			"recipient":       recipient,
		})
		return result
	}

	result.Success = true
	result.MessageID = messageID
	return result
}

func (uc *DeliverNotificationUseCase) deliverSlack(ctx context.Context, notification *domain.AINotification, webhookURL string) ChannelDeliveryResult {
	result := ChannelDeliveryResult{Channel: domain.DeliveryChannelSlack}

	if uc.webhookProvider == nil {
		result.Error = fmt.Errorf("webhook provider not configured")
		return result
	}

	payload := uc.buildSlackPayload(notification)
	err := uc.webhookProvider.SendWebhook(ctx, webhookURL, payload)

	if err != nil {
		result.Error = err
		uc.logger.Error(ctx, err, "Failed to send Slack notification", logger.Tags{
			"notification_id": notification.ID.String(),
		})
		return result
	}

	result.Success = true
	return result
}

func (uc *DeliverNotificationUseCase) deliverSMS(ctx context.Context, notification *domain.AINotification, phoneNumber string) ChannelDeliveryResult {
	result := ChannelDeliveryResult{Channel: domain.DeliveryChannelSMS}

	if uc.smsProvider == nil {
		result.Error = fmt.Errorf("SMS provider not configured")
		return result
	}

	// Only send SMS for critical notifications
	if notification.Priority != domain.NotificationPriorityCritical {
		result.Success = true // Skip silently for non-critical
		return result
	}

	message := uc.buildSMSMessage(notification)
	messageID, err := uc.smsProvider.SendSMS(ctx, phoneNumber, message)

	if err != nil {
		result.Error = err
		uc.logger.Error(ctx, err, "Failed to send SMS", logger.Tags{
			"notification_id": notification.ID.String(),
		})
		return result
	}

	result.Success = true
	result.MessageID = messageID
	return result
}

func (uc *DeliverNotificationUseCase) deliverWebhook(ctx context.Context, notification *domain.AINotification, webhookURL string) ChannelDeliveryResult {
	result := ChannelDeliveryResult{Channel: domain.DeliveryChannelWebhook}

	if uc.webhookProvider == nil {
		result.Error = fmt.Errorf("webhook provider not configured")
		return result
	}

	payload := uc.buildWebhookPayload(notification)
	err := uc.webhookProvider.SendWebhook(ctx, webhookURL, payload)

	if err != nil {
		result.Error = err
		uc.logger.Error(ctx, err, "Failed to send webhook", logger.Tags{
			"notification_id": notification.ID.String(),
		})
		return result
	}

	result.Success = true
	return result
}

func (uc *DeliverNotificationUseCase) queueForRetry(notification *domain.AINotification, channel domain.DeliveryChannel, recipient string, reason string) {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	priority := uc.getPriorityValue(notification.Priority)
	item := domain.NewDeliveryQueueItem(
		notification.ID,
		notification.OrganizationID,
		notification.UserID,
		channel,
		recipient,
		priority,
	)
	item.ScheduleRetry(reason)

	uc.deliveryQueue.AddItem(item)
}

// ProcessRetryQueue processes items in the retry queue
func (uc *DeliverNotificationUseCase) ProcessRetryQueue(ctx context.Context) (int, int, error) {
	uc.mu.RLock()
	retryable := uc.deliveryQueue.GetRetryableItems()
	uc.mu.RUnlock()

	successCount := 0
	failedCount := 0

	for _, item := range retryable {
		// Mark as in-flight
		item.MarkAsInFlight()

		// Get the notification
		notification, err := uc.notifRepo.GetByID(ctx, item.NotificationID, item.OrganizationID)
		if err != nil {
			item.MarkAsFailed(fmt.Sprintf("notification not found: %s", err.Error()))
			failedCount++
			continue
		}

		// Retry delivery
		result := uc.deliverToChannel(ctx, notification, item.Channel, item.Recipient)

		if result.Success {
			item.MarkAsDelivered(result.MessageID)
			successCount++
		} else if item.CanRetry() {
			item.ScheduleRetry(result.Error.Error())
		} else {
			item.MarkAsFailed(result.Error.Error())
			failedCount++
		}
	}

	return successCount, failedCount, nil
}

// GetQueueStats returns statistics about the delivery queue
func (uc *DeliverNotificationUseCase) GetQueueStats() (pending, failed, delivered int) {
	uc.mu.RLock()
	defer uc.mu.RUnlock()

	uc.deliveryQueue.UpdateStats()
	return uc.deliveryQueue.PendingCount, uc.deliveryQueue.FailedCount, uc.deliveryQueue.DeliveredCount
}

// SetChannelConfig sets the channel configuration for an organization
func (uc *DeliverNotificationUseCase) SetChannelConfig(orgID uuid.UUID, configSet *domain.ChannelConfigSet) {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	uc.channelConfigs[orgID] = configSet
}

func (uc *DeliverNotificationUseCase) getChannelConfigs(orgID uuid.UUID) *domain.ChannelConfigSet {
	uc.mu.RLock()
	defer uc.mu.RUnlock()
	return uc.channelConfigs[orgID]
}

func (uc *DeliverNotificationUseCase) isInQuietHours(prefs *domain.UserNotificationPreferences) bool {
	if prefs.QuietHoursStart == nil || prefs.QuietHoursEnd == nil {
		return false
	}

	now := time.Now()
	loc, err := time.LoadLocation(prefs.Timezone)
	if err != nil {
		loc = time.UTC
	}

	localNow := now.In(loc)

	startHour := prefs.QuietHoursStart.Hour()
	startMin := prefs.QuietHoursStart.Minute()
	endHour := prefs.QuietHoursEnd.Hour()
	endMin := prefs.QuietHoursEnd.Minute()

	currentMinutes := localNow.Hour()*60 + localNow.Minute()
	startMinutes := startHour*60 + startMin
	endMinutes := endHour*60 + endMin

	// Handle overnight quiet hours
	if startMinutes > endMinutes {
		return currentMinutes >= startMinutes || currentMinutes < endMinutes
	}

	return currentMinutes >= startMinutes && currentMinutes < endMinutes
}

func (uc *DeliverNotificationUseCase) shouldDeliverToChannel(notifPriority, minPriority domain.NotificationPriority) bool {
	priorities := map[domain.NotificationPriority]int{
		domain.NotificationPriorityCritical: 1,
		domain.NotificationPriorityHigh:     2,
		domain.NotificationPriorityMedium:   3,
		domain.NotificationPriorityLow:      4,
	}

	return priorities[notifPriority] <= priorities[minPriority]
}

func (uc *DeliverNotificationUseCase) getPriorityValue(priority domain.NotificationPriority) int {
	switch priority {
	case domain.NotificationPriorityCritical:
		return 1
	case domain.NotificationPriorityHigh:
		return 2
	case domain.NotificationPriorityMedium:
		return 3
	default:
		return 4
	}
}

func (uc *DeliverNotificationUseCase) buildEmailSubject(notification *domain.AINotification) string {
	switch notification.Priority {
	case domain.NotificationPriorityCritical:
		return "🚨 CRITICAL: " + notification.Title
	case domain.NotificationPriorityHigh:
		return "⚠️ HIGH: " + notification.Title
	default:
		return notification.Title
	}
}

func (uc *DeliverNotificationUseCase) buildEmailHTMLBody(notification *domain.AINotification) string {
	return fmt.Sprintf("<h1>%s</h1><p>%s</p>", notification.Title, notification.Summary)
}

func (uc *DeliverNotificationUseCase) buildEmailTextBody(notification *domain.AINotification) string {
	return fmt.Sprintf("%s\n\n%s", notification.Title, notification.Summary)
}

func (uc *DeliverNotificationUseCase) buildSlackPayload(notification *domain.AINotification) map[string]interface{} {
	emoji := "📬"
	switch notification.Priority {
	case domain.NotificationPriorityCritical:
		emoji = "🚨"
	case domain.NotificationPriorityHigh:
		emoji = "⚠️"
	case domain.NotificationPriorityMedium:
		emoji = "📌"
	case domain.NotificationPriorityLow:
		emoji = "ℹ️"
	}

	return map[string]interface{}{
		"text": fmt.Sprintf("%s %s - %s", emoji, notification.Title, notification.Summary),
	}
}

func (uc *DeliverNotificationUseCase) buildWebhookPayload(notification *domain.AINotification) map[string]interface{} {
	return map[string]interface{}{
		"id":         notification.ID.String(),
		"type":       notification.Type,
		"priority":   notification.Priority,
		"title":      notification.Title,
		"summary":    notification.Summary,
		"created_at": notification.CreatedAt,
	}
}

func (uc *DeliverNotificationUseCase) buildSMSMessage(notification *domain.AINotification) string {
	msg := fmt.Sprintf("🚨 GIIA: %s - %s", notification.Title, notification.Summary)
	if len(msg) > 160 {
		return msg[:157] + "..."
	}
	return msg
}
