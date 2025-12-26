package domain

import (
	"time"

	"github.com/google/uuid"
)

// ChannelConfig represents configuration for a notification channel
type ChannelConfig struct {
	ID             uuid.UUID
	OrganizationID uuid.UUID
	Channel        DeliveryChannel
	Enabled        bool
	Priority       int // Order of preference (1 = highest)

	// Rate limiting
	MaxPerHour    int
	MaxPerDay     int
	CurrentHourly int
	CurrentDaily  int
	LastResetHour time.Time
	LastResetDay  time.Time

	// Quiet hours (optional)
	QuietHoursEnabled bool
	QuietHoursStart   string // Format: "HH:MM"
	QuietHoursEnd     string // Format: "HH:MM"
	Timezone          string

	// Channel-specific config stored as key-value
	Config map[string]string

	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewChannelConfig creates a new channel configuration
func NewChannelConfig(
	organizationID uuid.UUID,
	channel DeliveryChannel,
) *ChannelConfig {
	now := time.Now()
	return &ChannelConfig{
		ID:             uuid.New(),
		OrganizationID: organizationID,
		Channel:        channel,
		Enabled:        true,
		Priority:       getPriorityForChannel(channel),
		MaxPerHour:     100,
		MaxPerDay:      1000,
		CurrentHourly:  0,
		CurrentDaily:   0,
		LastResetHour:  now,
		LastResetDay:   now,
		Timezone:       "UTC",
		Config:         make(map[string]string),
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

// getPriorityForChannel returns default priority for a channel
func getPriorityForChannel(channel DeliveryChannel) int {
	switch channel {
	case DeliveryChannelSMS:
		return 1 // Highest priority for SMS (critical alerts)
	case DeliveryChannelSlack:
		return 2
	case DeliveryChannelEmail:
		return 3
	case DeliveryChannelWebhook:
		return 4
	case DeliveryChannelInApp:
		return 5
	default:
		return 10
	}
}

// CanSend checks if the channel can send based on rate limits
func (c *ChannelConfig) CanSend() bool {
	if !c.Enabled {
		return false
	}

	now := time.Now()

	// Reset hourly counter if needed
	if now.Sub(c.LastResetHour) >= time.Hour {
		c.CurrentHourly = 0
		c.LastResetHour = now
	}

	// Reset daily counter if needed
	if now.Sub(c.LastResetDay) >= 24*time.Hour {
		c.CurrentDaily = 0
		c.LastResetDay = now
	}

	return c.CurrentHourly < c.MaxPerHour && c.CurrentDaily < c.MaxPerDay
}

// IncrementUsage increments the rate limit counters
func (c *ChannelConfig) IncrementUsage() {
	c.CurrentHourly++
	c.CurrentDaily++
	c.UpdatedAt = time.Now()
}

// IsInQuietHours checks if current time is within quiet hours
func (c *ChannelConfig) IsInQuietHours(currentTime time.Time) bool {
	if !c.QuietHoursEnabled {
		return false
	}

	// Load timezone
	loc, err := time.LoadLocation(c.Timezone)
	if err != nil {
		loc = time.UTC
	}

	localTime := currentTime.In(loc)

	// Parse quiet hours
	startHour, startMin := parseTimeString(c.QuietHoursStart)
	endHour, endMin := parseTimeString(c.QuietHoursEnd)

	currentMinutes := localTime.Hour()*60 + localTime.Minute()
	startMinutes := startHour*60 + startMin
	endMinutes := endHour*60 + endMin

	// Handle overnight quiet hours (e.g., 22:00 to 06:00)
	if startMinutes > endMinutes {
		return currentMinutes >= startMinutes || currentMinutes < endMinutes
	}

	return currentMinutes >= startMinutes && currentMinutes < endMinutes
}

// parseTimeString parses a time string in format "HH:MM"
func parseTimeString(timeStr string) (hour, minute int) {
	if len(timeStr) != 5 {
		return 0, 0
	}

	hour = int(timeStr[0]-'0')*10 + int(timeStr[1]-'0')
	minute = int(timeStr[3]-'0')*10 + int(timeStr[4]-'0')

	return hour, minute
}

// ChannelConfigSet represents a set of channel configurations for an organization
type ChannelConfigSet struct {
	Configs        map[DeliveryChannel]*ChannelConfig
	OrganizationID uuid.UUID
}

// NewChannelConfigSet creates a new channel config set
func NewChannelConfigSet(organizationID uuid.UUID) *ChannelConfigSet {
	return &ChannelConfigSet{
		Configs:        make(map[DeliveryChannel]*ChannelConfig),
		OrganizationID: organizationID,
	}
}

// AddConfig adds a channel configuration
func (s *ChannelConfigSet) AddConfig(config *ChannelConfig) {
	s.Configs[config.Channel] = config
}

// GetConfig returns configuration for a channel
func (s *ChannelConfigSet) GetConfig(channel DeliveryChannel) *ChannelConfig {
	return s.Configs[channel]
}

// GetEnabledChannelsByPriority returns enabled channels sorted by priority
func (s *ChannelConfigSet) GetEnabledChannelsByPriority() []DeliveryChannel {
	channels := make([]DeliveryChannel, 0)

	// Sort by priority (lower number = higher priority)
	for priority := 1; priority <= 10; priority++ {
		for channel, config := range s.Configs {
			if config.Enabled && config.Priority == priority {
				channels = append(channels, channel)
			}
		}
	}

	return channels
}

// GetAvailableChannels returns channels that are enabled and within rate limits
func (s *ChannelConfigSet) GetAvailableChannels(currentTime time.Time) []DeliveryChannel {
	channels := make([]DeliveryChannel, 0)

	for channel, config := range s.Configs {
		if config.CanSend() && !config.IsInQuietHours(currentTime) {
			channels = append(channels, channel)
		}
	}

	return channels
}

// Digest represents a daily digest of notifications
type Digest struct {
	ID              uuid.UUID
	OrganizationID  uuid.UUID
	UserID          uuid.UUID
	GeneratedAt     time.Time
	PeriodStart     time.Time
	PeriodEnd       time.Time
	TotalCount      int
	CountByPriority map[NotificationPriority]int
	CountByType     map[NotificationType]int
	UnactedCount    int
	TopItems        []*AINotification
	SummaryText     string
	HTMLContent     string
	DeliveryStatus  DeliveryStatus
	DeliveredAt     *time.Time
	DeliveryChannel DeliveryChannel
}

// NewDigest creates a new digest
func NewDigest(
	organizationID uuid.UUID,
	userID uuid.UUID,
	periodStart time.Time,
	periodEnd time.Time,
) *Digest {
	return &Digest{
		ID:              uuid.New(),
		OrganizationID:  organizationID,
		UserID:          userID,
		GeneratedAt:     time.Now(),
		PeriodStart:     periodStart,
		PeriodEnd:       periodEnd,
		CountByPriority: make(map[NotificationPriority]int),
		CountByType:     make(map[NotificationType]int),
		TopItems:        make([]*AINotification, 0),
		DeliveryStatus:  DeliveryStatusPending,
	}
}

// AddNotification adds a notification to the digest
func (d *Digest) AddNotification(notification *AINotification) {
	d.TotalCount++
	d.CountByPriority[notification.Priority]++
	d.CountByType[notification.Type]++

	if notification.Status == NotificationStatusUnread {
		d.UnactedCount++
	}

	// Keep top items by priority (max 10)
	if len(d.TopItems) < 10 {
		d.TopItems = append(d.TopItems, notification)
	} else if isHigherPriority(notification.Priority, d.TopItems[len(d.TopItems)-1].Priority) {
		d.TopItems[len(d.TopItems)-1] = notification
	}
}

// isHigherPriority returns true if p1 is higher priority than p2
func isHigherPriority(p1, p2 NotificationPriority) bool {
	priorities := map[NotificationPriority]int{
		NotificationPriorityCritical: 1,
		NotificationPriorityHigh:     2,
		NotificationPriorityMedium:   3,
		NotificationPriorityLow:      4,
	}

	return priorities[p1] < priorities[p2]
}

// GenerateSummary generates a text summary of the digest
func (d *Digest) GenerateSummary() {
	summary := "📊 Daily Notification Digest\n\n"
	summary += "📅 Period: " + d.PeriodStart.Format("Jan 02") + " to " + d.PeriodEnd.Format("Jan 02, 2006") + "\n\n"

	summary += "📈 Summary:\n"
	summary += "  • Total notifications: " + intToString(d.TotalCount) + "\n"
	summary += "  • Unacted: " + intToString(d.UnactedCount) + "\n\n"

	summary += "📊 By Priority:\n"
	if count, ok := d.CountByPriority[NotificationPriorityCritical]; ok && count > 0 {
		summary += "  🚨 Critical: " + intToString(count) + "\n"
	}
	if count, ok := d.CountByPriority[NotificationPriorityHigh]; ok && count > 0 {
		summary += "  ⚠️ High: " + intToString(count) + "\n"
	}
	if count, ok := d.CountByPriority[NotificationPriorityMedium]; ok && count > 0 {
		summary += "  📌 Medium: " + intToString(count) + "\n"
	}
	if count, ok := d.CountByPriority[NotificationPriorityLow]; ok && count > 0 {
		summary += "  ℹ️ Low: " + intToString(count) + "\n"
	}

	if len(d.TopItems) > 0 {
		summary += "\n🔝 Top Items:\n"
		for i, item := range d.TopItems {
			if i >= 5 {
				break
			}
			summary += "  " + intToString(i+1) + ". " + item.Title + "\n"
		}
	}

	d.SummaryText = summary
}

// intToString converts int to string without importing strconv
func intToString(n int) string {
	if n == 0 {
		return "0"
	}

	negative := n < 0
	if negative {
		n = -n
	}

	digits := make([]byte, 0, 10)
	for n > 0 {
		digits = append(digits, byte('0'+n%10))
		n /= 10
	}

	// Reverse
	for i, j := 0, len(digits)-1; i < j; i, j = i+1, j-1 {
		digits[i], digits[j] = digits[j], digits[i]
	}

	if negative {
		return "-" + string(digits)
	}
	return string(digits)
}
