CREATE TABLE IF NOT EXISTS buffer_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    buffer_id UUID NOT NULL REFERENCES buffers(id) ON DELETE CASCADE,
    product_id UUID NOT NULL,
    organization_id UUID NOT NULL,
    snapshot_date DATE NOT NULL,
    cpd DECIMAL(15,2) NOT NULL,
    dlt INTEGER NOT NULL,
    red_zone DECIMAL(15,2) NOT NULL,
    red_base DECIMAL(15,2) NOT NULL,
    red_safe DECIMAL(15,2) NOT NULL,
    yellow_zone DECIMAL(15,2) NOT NULL,
    green_zone DECIMAL(15,2) NOT NULL,
    lead_time_factor DECIMAL(5,2) NOT NULL,
    variability_factor DECIMAL(5,2) NOT NULL,
    moq INTEGER,
    order_frequency INTEGER,
    has_adjustments BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_buffer_history_date UNIQUE (buffer_id, snapshot_date)
);

CREATE INDEX idx_buffer_history_buffer ON buffer_history(buffer_id);
CREATE INDEX idx_buffer_history_product ON buffer_history(product_id, organization_id);
CREATE INDEX idx_buffer_history_date ON buffer_history(snapshot_date DESC);
CREATE INDEX idx_buffer_history_org ON buffer_history(organization_id);

COMMENT ON TABLE buffer_history IS 'Daily snapshots of buffer calculations for trend analysis and auditing';
COMMENT ON COLUMN buffer_history.snapshot_date IS 'Date of the snapshot (daily recalculation)';
COMMENT ON COLUMN buffer_history.has_adjustments IS 'Whether FAD or buffer adjustments were applied in this snapshot';
COMMENT ON COLUMN buffer_history.lead_time_factor IS '%LT from buffer profile used in calculation';
COMMENT ON COLUMN buffer_history.variability_factor IS '%CV from buffer profile used in calculation';
COMMENT ON COLUMN buffer_history.moq IS 'MOQ (Minimum Order Quantity) considered for Green Zone';
COMMENT ON COLUMN buffer_history.order_frequency IS 'FO (Order Frequency) used for Green Zone calculation';
