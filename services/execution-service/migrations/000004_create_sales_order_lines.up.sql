-- Sales Order Line Items table
CREATE TABLE IF NOT EXISTS sales_order_lines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sales_order_id UUID NOT NULL REFERENCES sales_orders(id) ON DELETE CASCADE,
    product_id UUID NOT NULL,
    quantity DECIMAL(15,4) NOT NULL,
    unit_price DECIMAL(15,4) DEFAULT 0,
    line_total DECIMAL(15,2) DEFAULT 0,
    CONSTRAINT chk_soli_quantity CHECK (quantity > 0)
);

CREATE INDEX idx_soli_so ON sales_order_lines(sales_order_id);
CREATE INDEX idx_soli_product ON sales_order_lines(product_id);

COMMENT ON TABLE sales_order_lines IS 'Line items for sales orders';
