package domain

import (
	"time"

	"github.com/google/uuid"
)

type UserNotificationPreferences struct {
	ID     uuid.UUID
	UserID uuid.UUID
	OrganizationID uuid.UUID

	EnableInApp bool
	EnableEmail bool
	EnableSMS   bool
	EnableSlack bool
	SlackWebhookURL string

	InAppMinPriority  NotificationPriority
	EmailMinPriority  NotificationPriority
	SMSMinPriority    NotificationPriority

	DigestTime      string
	QuietHoursStart *time.Time
	QuietHoursEnd   *time.Time
	Timezone        string

	MaxAlertsPerHour int
	MaxEmailsPerDay  int

	DetailLevel       string
	IncludeCharts     bool
	IncludeHistorical bool

	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewUserPreferences(userID uuid.UUID, organizationID uuid.UUID) *UserNotificationPreferences {
	return &UserNotificationPreferences{
		ID:                uuid.New(),
		UserID:            userID,
		OrganizationID:    organizationID,
		EnableInApp:       true,
		EnableEmail:       true,
		EnableSMS:         false,
		EnableSlack:       false,
		InAppMinPriority:  NotificationPriorityLow,
		EmailMinPriority:  NotificationPriorityMedium,
		SMSMinPriority:    NotificationPriorityCritical,
		DigestTime:        "06:00",
		Timezone:          "UTC",
		MaxAlertsPerHour:  10,
		MaxEmailsPerDay:   50,
		DetailLevel:       "detailed",
		IncludeCharts:     true,
		IncludeHistorical: true,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
}
