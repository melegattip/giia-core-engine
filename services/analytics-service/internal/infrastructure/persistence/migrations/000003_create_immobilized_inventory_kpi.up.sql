-- Immobilized Inventory KPI table for tracking old inventory
CREATE TABLE IF NOT EXISTS immobilized_inventory_kpi (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL,
    snapshot_date DATE NOT NULL,
    threshold_years INTEGER NOT NULL,
    immobilized_count INTEGER NOT NULL,
    immobilized_value DECIMAL(15,2) NOT NULL,
    total_stock_value DECIMAL(15,2) NOT NULL,
    immobilized_percentage DECIMAL(5,2) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_immobilized_org_date_threshold UNIQUE (organization_id, snapshot_date, threshold_years)
);

CREATE INDEX idx_immobilized_org ON immobilized_inventory_kpi(organization_id);
CREATE INDEX idx_immobilized_date ON immobilized_inventory_kpi(snapshot_date DESC);
CREATE INDEX idx_immobilized_org_date ON immobilized_inventory_kpi(organization_id, snapshot_date DESC);

COMMENT ON TABLE immobilized_inventory_kpi IS 'Daily snapshots of immobilized inventory KPI';
COMMENT ON COLUMN immobilized_inventory_kpi.threshold_years IS 'Configurable threshold in years (e.g., 1, 2, 3)';
COMMENT ON COLUMN immobilized_inventory_kpi.immobilized_count IS 'Number of products older than threshold';
COMMENT ON COLUMN immobilized_inventory_kpi.immobilized_percentage IS '(ImmobilizedValue / TotalStockValue) × 100';
