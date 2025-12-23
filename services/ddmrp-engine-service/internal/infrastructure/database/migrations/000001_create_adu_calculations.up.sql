CREATE TABLE IF NOT EXISTS adu_calculations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL,
    organization_id UUID NOT NULL,
    calculation_date DATE NOT NULL,
    adu_value DECIMAL(15,2) NOT NULL,
    method VARCHAR(20) NOT NULL,
    period_days INTEGER NOT NULL DEFAULT 30,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_adu_product_date UNIQUE (product_id, organization_id, calculation_date),
    CONSTRAINT chk_adu_method CHECK (method IN ('average', 'exponential', 'weighted')),
    CONSTRAINT chk_adu_value CHECK (adu_value >= 0),
    CONSTRAINT chk_period_days CHECK (period_days > 0)
);

CREATE INDEX idx_adu_product ON adu_calculations(product_id, organization_id);
CREATE INDEX idx_adu_calc_date ON adu_calculations(calculation_date DESC);
CREATE INDEX idx_adu_org ON adu_calculations(organization_id);

COMMENT ON TABLE adu_calculations IS 'Average Daily Usage calculations for products';
COMMENT ON COLUMN adu_calculations.method IS 'Calculation method: average, exponential, weighted';
COMMENT ON COLUMN adu_calculations.period_days IS 'Number of days used for calculation (30, 60, 90)';
