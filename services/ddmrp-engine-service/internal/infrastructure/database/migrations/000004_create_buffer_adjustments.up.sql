CREATE TABLE IF NOT EXISTS buffer_adjustments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    buffer_id UUID NOT NULL REFERENCES buffers(id) ON DELETE CASCADE,
    product_id UUID NOT NULL,
    organization_id UUID NOT NULL,
    adjustment_type VARCHAR(30) NOT NULL,
    target_zone VARCHAR(20) NOT NULL,
    factor DECIMAL(5,2) NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    reason TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL,
    CONSTRAINT chk_buffer_adj_type CHECK (adjustment_type IN ('zone_factor', 'planned_event', 'spike_management', 'seasonal_prepare')),
    CONSTRAINT chk_buffer_target_zone CHECK (target_zone IN ('red', 'yellow', 'green', 'all')),
    CONSTRAINT chk_buffer_factor CHECK (factor > 0),
    CONSTRAINT chk_buffer_dates CHECK (end_date >= start_date)
);

CREATE INDEX idx_buffer_adj_buffer ON buffer_adjustments(buffer_id);
CREATE INDEX idx_buffer_adj_product ON buffer_adjustments(product_id, organization_id);
CREATE INDEX idx_buffer_adj_dates ON buffer_adjustments(start_date, end_date);
CREATE INDEX idx_buffer_adj_active ON buffer_adjustments(buffer_id, start_date, end_date);

COMMENT ON TABLE buffer_adjustments IS 'Manual buffer zone adjustments for planned events';
COMMENT ON COLUMN buffer_adjustments.target_zone IS 'Which zone to adjust: red, yellow, green, or all';
COMMENT ON COLUMN buffer_adjustments.factor IS 'Multiplier for zone size (e.g., 1.2 = 20% increase)';
COMMENT ON COLUMN buffer_adjustments.adjustment_type IS 'Type: zone_factor, planned_event, spike_management, seasonal_prepare';
