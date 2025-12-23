-- KPI Snapshots table for overall inventory performance metrics
CREATE TABLE IF NOT EXISTS kpi_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL,
    snapshot_date DATE NOT NULL,
    inventory_turnover DECIMAL(10,2),
    stockout_rate DECIMAL(5,2),
    service_level DECIMAL(5,2),
    excess_inventory_pct DECIMAL(5,2),
    buffer_score_green DECIMAL(5,2),
    buffer_score_yellow DECIMAL(5,2),
    buffer_score_red DECIMAL(5,2),
    total_inventory_value DECIMAL(15,2),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_kpi_snapshot_org_date UNIQUE (organization_id, snapshot_date)
);

CREATE INDEX idx_kpi_snapshots_org ON kpi_snapshots(organization_id);
CREATE INDEX idx_kpi_snapshots_date ON kpi_snapshots(snapshot_date DESC);
CREATE INDEX idx_kpi_snapshots_org_date ON kpi_snapshots(organization_id, snapshot_date DESC);

COMMENT ON TABLE kpi_snapshots IS 'Daily snapshots of overall inventory performance KPIs';
COMMENT ON COLUMN kpi_snapshots.inventory_turnover IS 'Inventory turnover ratio (Sales / Avg Stock)';
COMMENT ON COLUMN kpi_snapshots.buffer_score_green IS 'Percentage of products in green buffer zone';
COMMENT ON COLUMN kpi_snapshots.buffer_score_yellow IS 'Percentage of products in yellow buffer zone';
COMMENT ON COLUMN kpi_snapshots.buffer_score_red IS 'Percentage of products in red buffer zone';
