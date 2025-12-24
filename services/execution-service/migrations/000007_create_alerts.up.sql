-- Alerts table
CREATE TABLE IF NOT EXISTS alerts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL,
    alert_type VARCHAR(50) NOT NULL,
    severity VARCHAR(20) NOT NULL,
    resource_type VARCHAR(50) NOT NULL,
    resource_id UUID NOT NULL,
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    data JSONB,
    acknowledged_at TIMESTAMP,
    acknowledged_by UUID,
    resolved_at TIMESTAMP,
    resolved_by UUID,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_alert_type CHECK (alert_type IN (
        'po_delayed', 'po_late_warning', 'buffer_red', 'buffer_below_red',
        'buffer_stockout', 'stock_deviation', 'obsolescence_risk',
        'excess_inventory', 'supplier_delay_pattern'
    )),
    CONSTRAINT chk_alert_severity CHECK (severity IN (
        'info', 'low', 'medium', 'high', 'critical'
    ))
);

CREATE INDEX idx_alert_org ON alerts(organization_id);
CREATE INDEX idx_alert_type ON alerts(alert_type);
CREATE INDEX idx_alert_severity ON alerts(severity);
CREATE INDEX idx_alert_resource ON alerts(resource_type, resource_id);
CREATE INDEX idx_alert_created ON alerts(created_at DESC);
CREATE INDEX idx_alert_active ON alerts(organization_id, acknowledged_at, resolved_at)
    WHERE acknowledged_at IS NULL AND resolved_at IS NULL;
CREATE INDEX idx_alert_org_severity_created ON alerts(organization_id, severity, created_at DESC);

COMMENT ON TABLE alerts IS 'System alerts and notifications';
COMMENT ON COLUMN alerts.data IS 'Additional context data in JSON format';
