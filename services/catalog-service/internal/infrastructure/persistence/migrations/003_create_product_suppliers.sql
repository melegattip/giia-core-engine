-- Migration: Create product_suppliers join table
-- Description: Many-to-many relationship between products and suppliers with supplier-specific attributes

SET search_path TO catalog;

CREATE TABLE IF NOT EXISTS product_suppliers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    supplier_id UUID NOT NULL REFERENCES suppliers(id) ON DELETE CASCADE,
    lead_time_days INT NOT NULL,
    unit_cost DECIMAL(12, 2),
    is_primary_supplier BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(product_id, supplier_id),
    CONSTRAINT chk_ps_lead_time_days CHECK (lead_time_days >= 0),
    CONSTRAINT chk_ps_unit_cost CHECK (unit_cost IS NULL OR unit_cost >= 0)
);

CREATE INDEX idx_product_suppliers_product_id ON product_suppliers(product_id);
CREATE INDEX idx_product_suppliers_supplier_id ON product_suppliers(supplier_id);
CREATE INDEX idx_product_suppliers_primary ON product_suppliers(product_id, is_primary_supplier) WHERE is_primary_supplier = true;

CREATE TRIGGER update_product_suppliers_updated_at BEFORE UPDATE ON product_suppliers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE product_suppliers IS 'Product-Supplier relationships with supplier-specific lead times and costs';
COMMENT ON COLUMN product_suppliers.lead_time_days IS 'Product-specific lead time from this supplier';
COMMENT ON COLUMN product_suppliers.unit_cost IS 'Cost per unit from this supplier';
COMMENT ON COLUMN product_suppliers.is_primary_supplier IS 'Whether this is the primary supplier for this product';
