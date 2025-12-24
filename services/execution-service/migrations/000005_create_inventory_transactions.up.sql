-- Inventory Transactions table
CREATE TABLE IF NOT EXISTS inventory_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL,
    product_id UUID NOT NULL,
    location_id UUID NOT NULL,
    type VARCHAR(20) NOT NULL,
    quantity DECIMAL(15,4) NOT NULL,
    unit_cost DECIMAL(15,4) DEFAULT 0,
    reference_type VARCHAR(50),
    reference_id UUID,
    reason TEXT,
    transaction_date TIMESTAMP NOT NULL,
    created_by UUID NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_inv_txn_type CHECK (type IN (
        'receipt', 'issue', 'transfer', 'adjustment'
    )),
    CONSTRAINT chk_inv_txn_quantity CHECK (quantity != 0)
);

CREATE INDEX idx_inv_txn_org ON inventory_transactions(organization_id);
CREATE INDEX idx_inv_txn_product ON inventory_transactions(product_id);
CREATE INDEX idx_inv_txn_location ON inventory_transactions(location_id);
CREATE INDEX idx_inv_txn_type ON inventory_transactions(type);
CREATE INDEX idx_inv_txn_ref ON inventory_transactions(reference_type, reference_id);
CREATE INDEX idx_inv_txn_date ON inventory_transactions(transaction_date DESC);
CREATE INDEX idx_inv_txn_org_product_date ON inventory_transactions(organization_id, product_id, transaction_date DESC);

COMMENT ON TABLE inventory_transactions IS 'Inventory movement transactions';
COMMENT ON COLUMN inventory_transactions.type IS 'Type of transaction: receipt, issue, transfer, adjustment';
COMMENT ON COLUMN inventory_transactions.reference_type IS 'Type of source document (purchase_order, sales_order, etc)';
