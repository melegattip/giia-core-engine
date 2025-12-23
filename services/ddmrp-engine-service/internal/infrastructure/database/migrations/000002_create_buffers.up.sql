CREATE TABLE IF NOT EXISTS buffers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL,
    organization_id UUID NOT NULL,
    buffer_profile_id UUID NOT NULL,
    cpd DECIMAL(15,2) NOT NULL,
    ltd INTEGER NOT NULL,
    red_base DECIMAL(15,2) NOT NULL,
    red_safe DECIMAL(15,2) NOT NULL,
    red_zone DECIMAL(15,2) NOT NULL,
    yellow_zone DECIMAL(15,2) NOT NULL,
    green_zone DECIMAL(15,2) NOT NULL,
    top_of_red DECIMAL(15,2) NOT NULL,
    top_of_yellow DECIMAL(15,2) NOT NULL,
    top_of_green DECIMAL(15,2) NOT NULL,
    on_hand DECIMAL(15,2) NOT NULL DEFAULT 0,
    on_order DECIMAL(15,2) NOT NULL DEFAULT 0,
    qualified_demand DECIMAL(15,2) NOT NULL DEFAULT 0,
    net_flow_position DECIMAL(15,2) NOT NULL DEFAULT 0,
    buffer_penetration DECIMAL(5,2) NOT NULL DEFAULT 0,
    zone VARCHAR(20) NOT NULL DEFAULT 'green',
    alert_level VARCHAR(20) NOT NULL DEFAULT 'normal',
    last_recalculated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_buffer_product UNIQUE (product_id, organization_id),
    CONSTRAINT chk_zone CHECK (zone IN ('green', 'yellow', 'red', 'below_red')),
    CONSTRAINT chk_alert_level CHECK (alert_level IN ('normal', 'monitor', 'replenish', 'critical')),
    CONSTRAINT chk_cpd CHECK (cpd >= 0),
    CONSTRAINT chk_ltd CHECK (ltd > 0),
    CONSTRAINT chk_zones CHECK (red_zone >= 0 AND yellow_zone >= 0 AND green_zone >= 0)
);

CREATE INDEX idx_buffers_product ON buffers(product_id, organization_id);
CREATE INDEX idx_buffers_org ON buffers(organization_id);
CREATE INDEX idx_buffers_zone ON buffers(zone);
CREATE INDEX idx_buffers_alert ON buffers(alert_level);
CREATE INDEX idx_buffers_last_recalc ON buffers(last_recalculated_at DESC);

COMMENT ON TABLE buffers IS 'DDMRP buffer zones and status for products';
COMMENT ON COLUMN buffers.cpd IS 'Current/Adjusted CPD (Consumo Promedio Diario) - ceiling value';
COMMENT ON COLUMN buffers.ltd IS 'Lead Time Decoupled (days)';
COMMENT ON COLUMN buffers.red_base IS 'Red Base = DLT × CPD × %LT';
COMMENT ON COLUMN buffers.red_safe IS 'Red Safe = Red Base × %CV';
COMMENT ON COLUMN buffers.red_zone IS 'Total Red Zone = Red Base + Red Safe';
COMMENT ON COLUMN buffers.yellow_zone IS 'Yellow Zone = CPD × DLT';
COMMENT ON COLUMN buffers.green_zone IS 'Green Zone = MAX(MOQ, FO × CPD, DLT × CPD × %LT)';
COMMENT ON COLUMN buffers.net_flow_position IS 'NFP = On-Hand + On-Order - Qualified Demand';
COMMENT ON COLUMN buffers.buffer_penetration IS 'Percentage (0-100) of NFP within total buffer';
