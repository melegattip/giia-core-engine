-- Buffer Analytics table for DDMRP buffer tracking and trend analysis
CREATE TABLE IF NOT EXISTS buffer_analytics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL,
    organization_id UUID NOT NULL,
    snapshot_date DATE NOT NULL,
    cpd DECIMAL(15,2) NOT NULL,
    red_zone DECIMAL(15,2) NOT NULL,
    red_base DECIMAL(15,2) NOT NULL,
    red_safe DECIMAL(15,2) NOT NULL,
    yellow_zone DECIMAL(15,2) NOT NULL,
    green_zone DECIMAL(15,2) NOT NULL,
    ltd INTEGER NOT NULL,
    lead_time_factor DECIMAL(5,2) NOT NULL,
    variability_factor DECIMAL(5,2) NOT NULL,
    moq INTEGER,
    order_frequency INTEGER,
    optimal_order_freq DECIMAL(10,2),
    safety_days DECIMAL(10,2),
    avg_open_orders DECIMAL(10,2),
    has_adjustments BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_buffer_analytics_product_date UNIQUE (product_id, organization_id, snapshot_date)
);

CREATE INDEX idx_buffer_analytics_product ON buffer_analytics(product_id, organization_id);
CREATE INDEX idx_buffer_analytics_org ON buffer_analytics(organization_id);
CREATE INDEX idx_buffer_analytics_date ON buffer_analytics(snapshot_date DESC);
CREATE INDEX idx_buffer_analytics_org_date ON buffer_analytics(organization_id, snapshot_date DESC);

COMMENT ON TABLE buffer_analytics IS 'Daily buffer configuration snapshots synchronized from DDMRP Engine';
COMMENT ON COLUMN buffer_analytics.cpd IS 'Calculated Planning Demand (CPD)';
COMMENT ON COLUMN buffer_analytics.ltd IS 'Lead Time Days (LTD)';
COMMENT ON COLUMN buffer_analytics.optimal_order_freq IS 'Green Zone / CPD';
COMMENT ON COLUMN buffer_analytics.safety_days IS 'Red Zone / CPD';
COMMENT ON COLUMN buffer_analytics.avg_open_orders IS 'Yellow Zone / Green Zone';
COMMENT ON COLUMN buffer_analytics.has_adjustments IS 'Whether FAD or buffer adjustments were applied';
