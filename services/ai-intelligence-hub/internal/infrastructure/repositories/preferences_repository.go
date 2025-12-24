package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/melegattip/giia-core-engine/pkg/errors"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/core/providers"
)

type preferencesRepository struct {
	db *sql.DB
}

// NewPreferencesRepository creates a new preferences repository.
func NewPreferencesRepository(db *sql.DB) providers.PreferencesRepository {
	return &preferencesRepository{db: db}
}

func (r *preferencesRepository) Create(ctx context.Context, prefs *domain.UserNotificationPreferences) error {
	if prefs == nil {
		return errors.NewBadRequest("preferences cannot be nil")
	}

	query := `
		INSERT INTO user_notification_preferences (
			id, user_id, organization_id,
			enable_in_app, enable_email, enable_sms, enable_slack,
			slack_webhook_url, email_address, phone_number,
			in_app_min_priority, email_min_priority, sms_min_priority,
			digest_time, quiet_hours_start, quiet_hours_end, timezone,
			max_alerts_per_hour, max_emails_per_day,
			detail_level, include_charts, include_historical,
			created_at, updated_at
		) VALUES (
			$1, $2, $3,
			$4, $5, $6, $7,
			$8, $9, $10,
			$11, $12, $13,
			$14, $15, $16, $17,
			$18, $19,
			$20, $21, $22,
			$23, $24
		)
	`

	_, err := r.db.ExecContext(ctx, query,
		prefs.ID, prefs.UserID, prefs.OrganizationID,
		prefs.EnableInApp, prefs.EnableEmail, prefs.EnableSMS, prefs.EnableSlack,
		prefs.SlackWebhookURL, prefs.EmailAddress, prefs.PhoneNumber,
		prefs.InAppMinPriority, prefs.EmailMinPriority, prefs.SMSMinPriority,
		prefs.DigestTime, prefs.QuietHoursStart, prefs.QuietHoursEnd, prefs.Timezone,
		prefs.MaxAlertsPerHour, prefs.MaxEmailsPerDay,
		prefs.DetailLevel, prefs.IncludeCharts, prefs.IncludeHistorical,
		prefs.CreatedAt, prefs.UpdatedAt,
	)
	if err != nil {
		return errors.NewInternalServerError(fmt.Sprintf("failed to insert preferences: %v", err))
	}

	return nil
}

func (r *preferencesRepository) GetByUserID(ctx context.Context, userID uuid.UUID, organizationID uuid.UUID) (*domain.UserNotificationPreferences, error) {
	query := `
		SELECT
			id, user_id, organization_id,
			enable_in_app, enable_email, enable_sms, enable_slack,
			slack_webhook_url, email_address, phone_number,
			in_app_min_priority, email_min_priority, sms_min_priority,
			digest_time, quiet_hours_start, quiet_hours_end, timezone,
			max_alerts_per_hour, max_emails_per_day,
			detail_level, include_charts, include_historical,
			created_at, updated_at
		FROM user_notification_preferences
		WHERE user_id = $1 AND organization_id = $2
	`

	prefs := &domain.UserNotificationPreferences{}

	err := r.db.QueryRowContext(ctx, query, userID, organizationID).Scan(
		&prefs.ID, &prefs.UserID, &prefs.OrganizationID,
		&prefs.EnableInApp, &prefs.EnableEmail, &prefs.EnableSMS, &prefs.EnableSlack,
		&prefs.SlackWebhookURL, &prefs.EmailAddress, &prefs.PhoneNumber,
		&prefs.InAppMinPriority, &prefs.EmailMinPriority, &prefs.SMSMinPriority,
		&prefs.DigestTime, &prefs.QuietHoursStart, &prefs.QuietHoursEnd, &prefs.Timezone,
		&prefs.MaxAlertsPerHour, &prefs.MaxEmailsPerDay,
		&prefs.DetailLevel, &prefs.IncludeCharts, &prefs.IncludeHistorical,
		&prefs.CreatedAt, &prefs.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.NewNotFound("preferences not found")
	}
	if err != nil {
		return nil, errors.NewInternalServerError(fmt.Sprintf("failed to query preferences: %v", err))
	}

	return prefs, nil
}

func (r *preferencesRepository) Update(ctx context.Context, prefs *domain.UserNotificationPreferences) error {
	if prefs == nil {
		return errors.NewBadRequest("preferences cannot be nil")
	}

	prefs.UpdatedAt = time.Now()

	query := `
		UPDATE user_notification_preferences
		SET
			enable_in_app = $1, enable_email = $2, enable_sms = $3, enable_slack = $4,
			slack_webhook_url = $5, email_address = $6, phone_number = $7,
			in_app_min_priority = $8, email_min_priority = $9, sms_min_priority = $10,
			digest_time = $11, quiet_hours_start = $12, quiet_hours_end = $13, timezone = $14,
			max_alerts_per_hour = $15, max_emails_per_day = $16,
			detail_level = $17, include_charts = $18, include_historical = $19,
			updated_at = $20
		WHERE id = $21 AND organization_id = $22
	`

	result, err := r.db.ExecContext(ctx, query,
		prefs.EnableInApp, prefs.EnableEmail, prefs.EnableSMS, prefs.EnableSlack,
		prefs.SlackWebhookURL, prefs.EmailAddress, prefs.PhoneNumber,
		prefs.InAppMinPriority, prefs.EmailMinPriority, prefs.SMSMinPriority,
		prefs.DigestTime, prefs.QuietHoursStart, prefs.QuietHoursEnd, prefs.Timezone,
		prefs.MaxAlertsPerHour, prefs.MaxEmailsPerDay,
		prefs.DetailLevel, prefs.IncludeCharts, prefs.IncludeHistorical,
		prefs.UpdatedAt,
		prefs.ID, prefs.OrganizationID,
	)
	if err != nil {
		return errors.NewInternalServerError(fmt.Sprintf("failed to update preferences: %v", err))
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.NewInternalServerError("failed to get rows affected")
	}

	if rowsAffected == 0 {
		return errors.NewNotFound("preferences not found")
	}

	return nil
}
