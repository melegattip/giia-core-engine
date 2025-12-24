-- Purchase Orders table
CREATE TABLE IF NOT EXISTS purchase_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL,
    po_number VARCHAR(50) NOT NULL,
    supplier_id UUID NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    order_date TIMESTAMP NOT NULL,
    expected_arrival_date TIMESTAMP NOT NULL,
    actual_arrival_date TIMESTAMP,
    delay_days INTEGER DEFAULT 0,
    is_delayed BOOLEAN DEFAULT FALSE,
    total_amount DECIMAL(15,2) DEFAULT 0,
    created_by UUID NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_po_org_number UNIQUE (organization_id, po_number),
    CONSTRAINT chk_po_status CHECK (status IN (
        'draft', 'pending', 'confirmed', 'partial', 'received', 'closed', 'cancelled'
    ))
);

CREATE INDEX idx_po_org ON purchase_orders(organization_id);
CREATE INDEX idx_po_number ON purchase_orders(po_number);
CREATE INDEX idx_po_supplier ON purchase_orders(supplier_id);
CREATE INDEX idx_po_status ON purchase_orders(status);
CREATE INDEX idx_po_delayed ON purchase_orders(is_delayed) WHERE is_delayed = TRUE;
CREATE INDEX idx_po_expected_date ON purchase_orders(expected_arrival_date);
CREATE INDEX idx_po_created ON purchase_orders(created_at DESC);

COMMENT ON TABLE purchase_orders IS 'Purchase orders for inventory replenishment';
COMMENT ON COLUMN purchase_orders.is_delayed IS 'Flag indicating if the PO is past expected arrival date';
COMMENT ON COLUMN purchase_orders.delay_days IS 'Number of days the PO is delayed';
