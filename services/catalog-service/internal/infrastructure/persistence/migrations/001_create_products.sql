-- Migration: Create products table
-- Description: Products are the core entities in the catalog representing SKUs

CREATE SCHEMA IF NOT EXISTS catalog;

SET search_path TO catalog;

CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sku VARCHAR(100) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(100),
    unit_of_measure VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    organization_id UUID NOT NULL,
    buffer_profile_id UUID,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(sku, organization_id)
);

CREATE INDEX idx_products_organization_id ON products(organization_id);
CREATE INDEX idx_products_sku ON products(sku);
CREATE INDEX idx_products_status ON products(status);
CREATE INDEX idx_products_category ON products(category);
CREATE INDEX idx_products_org_sku ON products(organization_id, sku);
CREATE INDEX idx_products_org_status ON products(organization_id, status);

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_products_updated_at BEFORE UPDATE ON products
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE products IS 'Product master data for the catalog';
COMMENT ON COLUMN products.sku IS 'Stock Keeping Unit - unique per organization';
COMMENT ON COLUMN products.unit_of_measure IS 'Unit of measurement (e.g., EA, KG, L)';
COMMENT ON COLUMN products.status IS 'Product status: active, inactive, discontinued';
COMMENT ON COLUMN products.buffer_profile_id IS 'Reference to buffer profile template';
