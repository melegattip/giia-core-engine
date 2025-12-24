-- Purchase Order Line Items table
CREATE TABLE IF NOT EXISTS purchase_order_lines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    purchase_order_id UUID NOT NULL REFERENCES purchase_orders(id) ON DELETE CASCADE,
    product_id UUID NOT NULL,
    quantity DECIMAL(15,4) NOT NULL,
    received_qty DECIMAL(15,4) DEFAULT 0,
    unit_cost DECIMAL(15,4) DEFAULT 0,
    line_total DECIMAL(15,2) DEFAULT 0,
    CONSTRAINT chk_poli_quantity CHECK (quantity > 0),
    CONSTRAINT chk_poli_received CHECK (received_qty >= 0)
);

CREATE INDEX idx_poli_po ON purchase_order_lines(purchase_order_id);
CREATE INDEX idx_poli_product ON purchase_order_lines(product_id);

COMMENT ON TABLE purchase_order_lines IS 'Line items for purchase orders';
COMMENT ON COLUMN purchase_order_lines.received_qty IS 'Quantity already received against this line';
