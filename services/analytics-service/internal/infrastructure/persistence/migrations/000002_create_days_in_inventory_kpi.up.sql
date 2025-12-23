-- Days in Inventory KPI table for tracking valued days in inventory
CREATE TABLE IF NOT EXISTS days_in_inventory_kpi (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL,
    snapshot_date DATE NOT NULL,
    total_valued_days DECIMAL(20,2) NOT NULL,
    average_valued_days DECIMAL(10,2) NOT NULL,
    total_products INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_days_in_inventory_org_date UNIQUE (organization_id, snapshot_date)
);

CREATE INDEX idx_days_in_inventory_org ON days_in_inventory_kpi(organization_id);
CREATE INDEX idx_days_in_inventory_date ON days_in_inventory_kpi(snapshot_date DESC);
CREATE INDEX idx_days_in_inventory_org_date ON days_in_inventory_kpi(organization_id, snapshot_date DESC);

COMMENT ON TABLE days_in_inventory_kpi IS 'Daily snapshots of valued days in inventory KPI';
COMMENT ON COLUMN days_in_inventory_kpi.total_valued_days IS 'Sum of (DaysInStock × UnitCost × Quantity) for all products';
COMMENT ON COLUMN days_in_inventory_kpi.average_valued_days IS 'Average valued days per product';
