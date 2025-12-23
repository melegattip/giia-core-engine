-- Inventory Rotation KPI table for tracking rotation performance
CREATE TABLE IF NOT EXISTS inventory_rotation_kpi (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL,
    snapshot_date DATE NOT NULL,
    sales_last_30_days DECIMAL(15,2) NOT NULL,
    avg_monthly_stock DECIMAL(15,2) NOT NULL,
    rotation_ratio DECIMAL(10,2) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_rotation_org_date UNIQUE (organization_id, snapshot_date)
);

CREATE INDEX idx_rotation_org ON inventory_rotation_kpi(organization_id);
CREATE INDEX idx_rotation_date ON inventory_rotation_kpi(snapshot_date DESC);
CREATE INDEX idx_rotation_org_date ON inventory_rotation_kpi(organization_id, snapshot_date DESC);

-- Top/Slow Rotating Products detail table
CREATE TABLE IF NOT EXISTS rotating_products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    kpi_id UUID NOT NULL REFERENCES inventory_rotation_kpi(id) ON DELETE CASCADE,
    product_id UUID NOT NULL,
    sku VARCHAR(100) NOT NULL,
    name VARCHAR(200) NOT NULL,
    sales_30_days DECIMAL(15,2) NOT NULL,
    avg_stock_value DECIMAL(15,2) NOT NULL,
    rotation_ratio DECIMAL(10,2) NOT NULL,
    category VARCHAR(20) NOT NULL CHECK (category IN ('top', 'slow')),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_rotating_products_kpi ON rotating_products(kpi_id);
CREATE INDEX idx_rotating_products_product ON rotating_products(product_id);
CREATE INDEX idx_rotating_products_category ON rotating_products(kpi_id, category);

COMMENT ON TABLE inventory_rotation_kpi IS 'Daily snapshots of inventory rotation KPI';
COMMENT ON COLUMN inventory_rotation_kpi.rotation_ratio IS '(Sales Last 30 Days) / (Average Monthly Stock)';
COMMENT ON TABLE rotating_products IS 'Top and slow rotating products detail';
