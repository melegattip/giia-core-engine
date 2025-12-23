-- AI Notifications table
CREATE TABLE IF NOT EXISTS ai_notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL,
    user_id UUID NOT NULL,

    type VARCHAR(20) NOT NULL,
    priority VARCHAR(20) NOT NULL,

    title VARCHAR(255) NOT NULL,
    summary TEXT NOT NULL,
    full_analysis TEXT,
    reasoning TEXT,

    risk_level VARCHAR(20),
    revenue_impact DECIMAL(15,2),
    cost_impact DECIMAL(15,2),
    time_to_impact_seconds INTEGER,
    affected_orders INTEGER,
    affected_products INTEGER,

    source_events JSONB,
    related_entities JSONB,

    status VARCHAR(20) NOT NULL DEFAULT 'unread',

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    read_at TIMESTAMP,
    acted_at TIMESTAMP,
    dismissed_at TIMESTAMP,

    CONSTRAINT chk_notification_type CHECK (type IN (
        'alert', 'warning', 'info', 'suggestion', 'insight', 'digest'
    )),
    CONSTRAINT chk_notification_priority CHECK (priority IN (
        'critical', 'high', 'medium', 'low'
    )),
    CONSTRAINT chk_notification_status CHECK (status IN (
        'unread', 'read', 'acted_upon', 'dismissed'
    ))
);

CREATE INDEX idx_notifications_user ON ai_notifications(user_id, organization_id);
CREATE INDEX idx_notifications_status ON ai_notifications(status, created_at);
CREATE INDEX idx_notifications_priority ON ai_notifications(priority, created_at);
CREATE INDEX idx_notifications_type ON ai_notifications(type);
CREATE INDEX idx_notifications_org_created ON ai_notifications(organization_id, created_at DESC);

-- Recommendations sub-table
CREATE TABLE IF NOT EXISTS ai_recommendations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    notification_id UUID NOT NULL REFERENCES ai_notifications(id) ON DELETE CASCADE,

    action TEXT NOT NULL,
    reasoning TEXT NOT NULL,
    expected_outcome TEXT,
    effort VARCHAR(20),
    impact VARCHAR(20),
    action_url TEXT,

    priority_order INTEGER NOT NULL,

    CONSTRAINT chk_recommendation_effort CHECK (effort IN ('low', 'medium', 'high')),
    CONSTRAINT chk_recommendation_impact CHECK (impact IN ('low', 'medium', 'high'))
);

CREATE INDEX idx_recommendations_notification ON ai_recommendations(notification_id, priority_order);
