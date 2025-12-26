package usecases

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/melegattip/giia-core-engine/pkg/logger"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/core/providers"
)

// GenerateDigestUseCase handles the generation and delivery of daily digests
type GenerateDigestUseCase struct {
	notifRepo       providers.NotificationRepository
	prefsRepo       providers.PreferencesRepository
	deliveryUseCase *DeliverNotificationUseCase
	logger          logger.Logger
}

// GenerateDigestInput represents the input for digest generation
type GenerateDigestInput struct {
	UserID         uuid.UUID
	OrganizationID uuid.UUID
	PeriodStart    time.Time
	PeriodEnd      time.Time
	DeliverNow     bool
}

// GenerateDigestOutput represents the result of digest generation
type GenerateDigestOutput struct {
	Digest        *domain.Digest
	Delivered     bool
	DeliveryError error
}

// ScheduledDigestConfig holds configuration for scheduled digest generation
type ScheduledDigestConfig struct {
	DigestTime    string // Format: "HH:MM"
	Timezone      string
	MaxItemsInTop int
}

// NewGenerateDigestUseCase creates a new generate digest use case
func NewGenerateDigestUseCase(
	notifRepo providers.NotificationRepository,
	prefsRepo providers.PreferencesRepository,
	deliveryUseCase *DeliverNotificationUseCase,
	log logger.Logger,
) *GenerateDigestUseCase {
	return &GenerateDigestUseCase{
		notifRepo:       notifRepo,
		prefsRepo:       prefsRepo,
		deliveryUseCase: deliveryUseCase,
		logger:          log,
	}
}

// Execute generates a digest for the specified user and period
func (uc *GenerateDigestUseCase) Execute(ctx context.Context, input *GenerateDigestInput) (*GenerateDigestOutput, error) {
	uc.logger.Info(ctx, "Generating digest", logger.Tags{
		"user_id":      input.UserID.String(),
		"org_id":       input.OrganizationID.String(),
		"period_start": input.PeriodStart.Format(time.RFC3339),
		"period_end":   input.PeriodEnd.Format(time.RFC3339),
	})

	startTime := time.Now()

	// Create the digest
	digest := domain.NewDigest(
		input.OrganizationID,
		input.UserID,
		input.PeriodStart,
		input.PeriodEnd,
	)

	// Fetch notifications for the period
	notifications, err := uc.fetchNotificationsForPeriod(ctx, input.UserID, input.OrganizationID, input.PeriodStart, input.PeriodEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch notifications: %w", err)
	}

	uc.logger.Info(ctx, "Fetched notifications for digest", logger.Tags{
		"count": fmt.Sprintf("%d", len(notifications)),
	})

	// Process notifications
	for _, notif := range notifications {
		digest.AddNotification(notif)
	}

	// Sort top items by priority
	uc.sortTopItemsByPriority(digest)

	// Generate summary text and HTML
	digest.GenerateSummary()
	digest.HTMLContent = uc.generateHTMLContent(digest)

	generationTime := time.Since(startTime)

	uc.logger.Info(ctx, "Digest generated", logger.Tags{
		"total_count":     fmt.Sprintf("%d", digest.TotalCount),
		"unacted_count":   fmt.Sprintf("%d", digest.UnactedCount),
		"generation_time": fmt.Sprintf("%dms", generationTime.Milliseconds()),
	})

	output := &GenerateDigestOutput{
		Digest: digest,
	}

	// Deliver if requested
	if input.DeliverNow {
		err := uc.deliverDigest(ctx, digest, input.UserID, input.OrganizationID)
		if err != nil {
			output.DeliveryError = err
			uc.logger.Error(ctx, err, "Failed to deliver digest", logger.Tags{
				"user_id": input.UserID.String(),
			})
		} else {
			output.Delivered = true
			digest.DeliveryStatus = domain.DeliveryStatusDelivered
			now := time.Now()
			digest.DeliveredAt = &now
		}
	}

	return output, nil
}

// GenerateAndDeliverScheduledDigests generates and delivers digests for all users at their configured time
func (uc *GenerateDigestUseCase) GenerateAndDeliverScheduledDigests(ctx context.Context, organizationID uuid.UUID, currentTime time.Time) (int, int, error) {
	uc.logger.Info(ctx, "Starting scheduled digest generation", logger.Tags{
		"org_id":       organizationID.String(),
		"current_time": currentTime.Format(time.RFC3339),
	})

	startTime := time.Now()

	// Get all users who should receive digests at this time
	users, err := uc.getUsersForDigestTime(ctx, organizationID, currentTime)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get users for digest: %w", err)
	}

	successCount := 0
	failedCount := 0

	for _, userID := range users {
		// Calculate period (last 24 hours)
		periodEnd := currentTime
		periodStart := periodEnd.Add(-24 * time.Hour)

		input := &GenerateDigestInput{
			UserID:         userID,
			OrganizationID: organizationID,
			PeriodStart:    periodStart,
			PeriodEnd:      periodEnd,
			DeliverNow:     true,
		}

		output, err := uc.Execute(ctx, input)
		if err != nil {
			uc.logger.Error(ctx, err, "Failed to generate digest for user", logger.Tags{
				"user_id": userID.String(),
			})
			failedCount++
			continue
		}

		if output.Delivered {
			successCount++
		} else if output.DeliveryError != nil {
			failedCount++
		}
	}

	totalTime := time.Since(startTime)

	uc.logger.Info(ctx, "Scheduled digest generation complete", logger.Tags{
		"success_count": fmt.Sprintf("%d", successCount),
		"failed_count":  fmt.Sprintf("%d", failedCount),
		"total_time":    fmt.Sprintf("%dms", totalTime.Milliseconds()),
	})

	return successCount, failedCount, nil
}

// ShouldGenerateDigestNow checks if it's time to generate a digest based on config
func (uc *GenerateDigestUseCase) ShouldGenerateDigestNow(config *ScheduledDigestConfig, currentTime time.Time) bool {
	loc, err := time.LoadLocation(config.Timezone)
	if err != nil {
		loc = time.UTC
	}

	localTime := currentTime.In(loc)

	// Parse configured time
	digestHour, digestMin := parseDigestTime(config.DigestTime)

	// Check if current time matches (within 5 minute window)
	currentHour := localTime.Hour()
	currentMin := localTime.Minute()

	if currentHour != digestHour {
		return false
	}

	// Allow 5 minute window for generation
	return currentMin >= digestMin && currentMin < digestMin+5
}

func (uc *GenerateDigestUseCase) fetchNotificationsForPeriod(
	ctx context.Context,
	userID uuid.UUID,
	organizationID uuid.UUID,
	periodStart time.Time,
	periodEnd time.Time,
) ([]*domain.AINotification, error) {
	// Fetch all notifications for the user
	filters := &providers.NotificationFilters{
		Limit: 1000, // Max notifications to include
	}

	notifications, err := uc.notifRepo.List(ctx, userID, organizationID, filters)
	if err != nil {
		return nil, err
	}

	// Filter by period
	filtered := make([]*domain.AINotification, 0)
	for _, notif := range notifications {
		if notif.CreatedAt.After(periodStart) && notif.CreatedAt.Before(periodEnd) {
			filtered = append(filtered, notif)
		}
	}

	return filtered, nil
}

func (uc *GenerateDigestUseCase) sortTopItemsByPriority(digest *domain.Digest) {
	if len(digest.TopItems) == 0 {
		return
	}

	priorityOrder := map[domain.NotificationPriority]int{
		domain.NotificationPriorityCritical: 1,
		domain.NotificationPriorityHigh:     2,
		domain.NotificationPriorityMedium:   3,
		domain.NotificationPriorityLow:      4,
	}

	sort.Slice(digest.TopItems, func(i, j int) bool {
		pi := priorityOrder[digest.TopItems[i].Priority]
		pj := priorityOrder[digest.TopItems[j].Priority]
		if pi != pj {
			return pi < pj
		}
		// If same priority, sort by creation time (newest first)
		return digest.TopItems[i].CreatedAt.After(digest.TopItems[j].CreatedAt)
	})

	// Limit to top 10
	if len(digest.TopItems) > 10 {
		digest.TopItems = digest.TopItems[:10]
	}
}

func (uc *GenerateDigestUseCase) deliverDigest(ctx context.Context, digest *domain.Digest, userID uuid.UUID, organizationID uuid.UUID) error {
	// Get user preferences
	prefs, err := uc.prefsRepo.GetByUserID(ctx, userID, organizationID)
	if err != nil {
		return fmt.Errorf("failed to get user preferences: %w", err)
	}

	// Create a digest notification
	digestNotification := &domain.AINotification{
		ID:             uuid.New(),
		OrganizationID: organizationID,
		UserID:         userID,
		Type:           domain.NotificationTypeDigest,
		Priority:       domain.NotificationPriorityLow,
		Title:          fmt.Sprintf("Daily Digest - %s", digest.GeneratedAt.Format("Jan 02, 2006")),
		Summary:        digest.SummaryText,
		FullAnalysis:   digest.HTMLContent,
		Status:         domain.NotificationStatusUnread,
		CreatedAt:      time.Now(),
	}

	// Build channels and recipients
	channels := make([]domain.DeliveryChannel, 0)
	recipients := make(map[domain.DeliveryChannel]string)

	if prefs.EnableEmail && prefs.EmailAddress != "" {
		channels = append(channels, domain.DeliveryChannelEmail)
		recipients[domain.DeliveryChannelEmail] = prefs.EmailAddress
	}

	if prefs.EnableSlack && prefs.SlackWebhookURL != "" {
		channels = append(channels, domain.DeliveryChannelSlack)
		recipients[domain.DeliveryChannelSlack] = prefs.SlackWebhookURL
	}

	if len(channels) == 0 {
		return nil // No channels configured
	}

	// Use delivery use case
	input := &DeliverNotificationInput{
		Notification: digestNotification,
		Channels:     channels,
		Recipients:   recipients,
		ForceDeliver: true, // Digests bypass quiet hours
	}

	_, err = uc.deliveryUseCase.Execute(ctx, input)
	return err
}

func (uc *GenerateDigestUseCase) getUsersForDigestTime(ctx context.Context, organizationID uuid.UUID, currentTime time.Time) ([]uuid.UUID, error) {
	// In production, this would query the database for users whose digest time matches
	// For now, return empty slice - this would be implemented with a preferences query
	return []uuid.UUID{}, nil
}

func (uc *GenerateDigestUseCase) generateHTMLContent(digest *domain.Digest) string {
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
        .content { background: white; padding: 30px; border: 1px solid #e0e0e0; border-top: none; }
        .stats { display: flex; justify-content: space-around; text-align: center; margin-bottom: 25px; flex-wrap: wrap; }
        .stat { padding: 15px; min-width: 80px; }
        .stat-value { font-size: 32px; font-weight: bold; color: #667eea; }
        .stat-label { font-size: 12px; color: #666; text-transform: uppercase; }
        .priority-bar { background: #f8f9fa; padding: 10px; margin: 5px 0; border-radius: 4px; display: flex; justify-content: space-between; }
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
            <div class="stats">
                <div class="stat">
                    <div class="stat-value">%d</div>
                    <div class="stat-label">Total</div>
                </div>
                <div class="stat">
                    <div class="stat-value">%d</div>
                    <div class="stat-label">Unacted</div>
                </div>
                <div class="stat">
                    <div class="stat-value">%d</div>
                    <div class="stat-label">Critical</div>
                </div>
            </div>
`, digest.PeriodStart.Format("Jan 02"), digest.PeriodEnd.Format("Jan 02, 2006"),
		digest.TotalCount, digest.UnactedCount, digest.CountByPriority[domain.NotificationPriorityCritical])

	// Priority breakdown
	html += `            <h3>📊 By Priority</h3>
`
	priorityEmojis := map[domain.NotificationPriority]string{
		domain.NotificationPriorityCritical: "🚨",
		domain.NotificationPriorityHigh:     "⚠️",
		domain.NotificationPriorityMedium:   "📌",
		domain.NotificationPriorityLow:      "ℹ️",
	}

	for _, p := range []domain.NotificationPriority{
		domain.NotificationPriorityCritical,
		domain.NotificationPriorityHigh,
		domain.NotificationPriorityMedium,
		domain.NotificationPriorityLow,
	} {
		if count := digest.CountByPriority[p]; count > 0 {
			html += fmt.Sprintf(`            <div class="priority-bar">
                <span>%s %s</span>
                <span><strong>%d</strong></span>
            </div>
`, priorityEmojis[p], p, count)
		}
	}

	// Top items
	if len(digest.TopItems) > 0 {
		html += `            <h3>🔝 Top Items</h3>
`
		for i, item := range digest.TopItems {
			if i >= 5 {
				break
			}
			html += fmt.Sprintf(`            <div class="top-item">
                <strong>%s</strong>
                <p>%s</p>
            </div>
`, item.Title, item.Summary)
		}
	}

	html += fmt.Sprintf(`        </div>
        <div class="footer">
            <p>GIIA AI Intelligence Hub</p>
            <p>Generated at %s</p>
            <a href="https://giia.app/notifications">View All Notifications</a>
        </div>
    </div>
</body>
</html>
`, digest.GeneratedAt.Format("Jan 02, 2006 15:04 MST"))

	return html
}

// parseDigestTime parses a time string in format "HH:MM"
func parseDigestTime(timeStr string) (hour, minute int) {
	if len(timeStr) < 5 {
		return 6, 0 // Default: 6:00 AM
	}

	hour = int(timeStr[0]-'0')*10 + int(timeStr[1]-'0')
	minute = int(timeStr[3]-'0')*10 + int(timeStr[4]-'0')

	return hour, minute
}
