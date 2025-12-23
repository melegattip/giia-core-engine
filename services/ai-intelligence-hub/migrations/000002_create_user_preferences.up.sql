-- User Notification Preferences
CREATE TABLE IF NOT EXISTS user_notification_preferences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    organization_id UUID NOT NULL,

    enable_in_app BOOLEAN NOT NULL DEFAULT true,
    enable_email BOOLEAN NOT NULL DEFAULT true,
    enable_sms BOOLEAN NOT NULL DEFAULT false,
    enable_slack BOOLEAN NOT NULL DEFAULT false,
    slack_webhook_url TEXT,

    in_app_min_priority VARCHAR(20) NOT NULL DEFAULT 'low',
    email_min_priority VARCHAR(20) NOT NULL DEFAULT 'medium',
    sms_min_priority VARCHAR(20) NOT NULL DEFAULT 'critical',

    digest_time TIME NOT NULL DEFAULT '06:00:00',
    quiet_hours_start TIME,
    quiet_hours_end TIME,
    timezone VARCHAR(50) NOT NULL DEFAULT 'UTC',

    max_alerts_per_hour INTEGER NOT NULL DEFAULT 10,
    max_emails_per_day INTEGER NOT NULL DEFAULT 50,

    detail_level VARCHAR(20) NOT NULL DEFAULT 'detailed',
    include_charts BOOLEAN NOT NULL DEFAULT true,
    include_historical BOOLEAN NOT NULL DEFAULT true,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_user_preferences UNIQUE (user_id, organization_id),
    CONSTRAINT chk_detail_level CHECK (detail_level IN ('brief', 'detailed', 'comprehensive'))
);

CREATE INDEX idx_user_prefs_user ON user_notification_preferences(user_id, organization_id);
CREATE INDEX idx_user_prefs_org ON user_notification_preferences(organization_id);
