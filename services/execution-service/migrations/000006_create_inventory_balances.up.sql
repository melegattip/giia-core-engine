-- Inventory Balances table
CREATE TABLE IF NOT EXISTS inventory_balances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL,
    product_id UUID NOT NULL,
    location_id UUID NOT NULL,
    on_hand DECIMAL(15,4) DEFAULT 0,
    reserved DECIMAL(15,4) DEFAULT 0,
    available DECIMAL(15,4) DEFAULT 0,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_inv_bal_org_product_location UNIQUE (organization_id, product_id, location_id)
);

CREATE INDEX idx_inv_bal_org ON inventory_balances(organization_id);
CREATE INDEX idx_inv_bal_product ON inventory_balances(product_id);
CREATE INDEX idx_inv_bal_location ON inventory_balances(location_id);
CREATE INDEX idx_inv_bal_low_stock ON inventory_balances(organization_id, available) 
    WHERE available <= 0;

COMMENT ON TABLE inventory_balances IS 'Current inventory levels by product and location';
COMMENT ON COLUMN inventory_balances.on_hand IS 'Total physical quantity on hand';
COMMENT ON COLUMN inventory_balances.reserved IS 'Quantity reserved for orders';
COMMENT ON COLUMN inventory_balances.available IS 'Available quantity (on_hand - reserved)';
