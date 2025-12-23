CREATE TABLE IF NOT EXISTS demand_adjustments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL,
    organization_id UUID NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    adjustment_type VARCHAR(30) NOT NULL,
    factor DECIMAL(5,2) NOT NULL,
    reason TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL,
    CONSTRAINT chk_adjustment_type CHECK (adjustment_type IN ('fad', 'seasonal', 'new_product', 'discontinue', 'promotion')),
    CONSTRAINT chk_factor CHECK (factor >= 0),
    CONSTRAINT chk_dates CHECK (end_date >= start_date)
);

CREATE INDEX idx_demand_adj_product ON demand_adjustments(product_id, organization_id);
CREATE INDEX idx_demand_adj_dates ON demand_adjustments(start_date, end_date);
CREATE INDEX idx_demand_adj_org ON demand_adjustments(organization_id);
CREATE INDEX idx_demand_adj_active ON demand_adjustments(product_id, organization_id, start_date, end_date);

COMMENT ON TABLE demand_adjustments IS 'FAD (Factor de Ajuste de Demanda) - Demand adjustment factors';
COMMENT ON COLUMN demand_adjustments.factor IS 'Multiplier for CPD (e.g., 1.5 = 50% increase, 0.0 = discontinue)';
COMMENT ON COLUMN demand_adjustments.adjustment_type IS 'Type of adjustment: fad, seasonal, new_product, discontinue, promotion';
COMMENT ON COLUMN demand_adjustments.reason IS 'Business reason for the adjustment';
