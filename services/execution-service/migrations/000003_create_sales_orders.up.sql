-- Sales Orders table
CREATE TABLE IF NOT EXISTS sales_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL,
    so_number VARCHAR(50) NOT NULL,
    customer_id UUID NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    order_date TIMESTAMP NOT NULL,
    due_date TIMESTAMP NOT NULL,
    ship_date TIMESTAMP,
    delivery_note_issued BOOLEAN DEFAULT FALSE,
    delivery_note_number VARCHAR(50),
    delivery_note_date TIMESTAMP,
    total_amount DECIMAL(15,2) DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_so_org_number UNIQUE (organization_id, so_number),
    CONSTRAINT chk_so_status CHECK (status IN (
        'pending', 'confirmed', 'picking', 'packed', 'shipped', 'delivered', 'cancelled'
    ))
);

CREATE INDEX idx_so_org ON sales_orders(organization_id);
CREATE INDEX idx_so_number ON sales_orders(so_number);
CREATE INDEX idx_so_customer ON sales_orders(customer_id);
CREATE INDEX idx_so_status ON sales_orders(status);
CREATE INDEX idx_so_due_date ON sales_orders(due_date);
CREATE INDEX idx_so_qualified_demand ON sales_orders(organization_id, status, delivery_note_issued) 
    WHERE status = 'confirmed' AND delivery_note_issued = FALSE;
CREATE INDEX idx_so_created ON sales_orders(created_at DESC);

COMMENT ON TABLE sales_orders IS 'Sales orders for customer demand';
COMMENT ON COLUMN sales_orders.delivery_note_issued IS 'Flag indicating if delivery note has been issued';
