-- Migration: Create suppliers table
-- Description: Suppliers provide products with associated lead times and reliability metrics

SET search_path TO catalog;

CREATE TABLE IF NOT EXISTS suppliers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(100) NOT NULL,
    name VARCHAR(255) NOT NULL,
    lead_time_days INT NOT NULL DEFAULT 0,
    reliability_rating INT DEFAULT 80,
    contact_info JSONB,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    organization_id UUID NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(code, organization_id),
    CONSTRAINT chk_reliability_rating CHECK (reliability_rating >= 0 AND reliability_rating <= 100),
    CONSTRAINT chk_lead_time_days CHECK (lead_time_days >= 0)
);

CREATE INDEX idx_suppliers_organization_id ON suppliers(organization_id);
CREATE INDEX idx_suppliers_code ON suppliers(code);
CREATE INDEX idx_suppliers_status ON suppliers(status);
CREATE INDEX idx_suppliers_org_code ON suppliers(organization_id, code);

CREATE TRIGGER update_suppliers_updated_at BEFORE UPDATE ON suppliers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE suppliers IS 'Supplier master data for product sourcing';
COMMENT ON COLUMN suppliers.code IS 'Supplier code - unique per organization';
COMMENT ON COLUMN suppliers.lead_time_days IS 'Default lead time in days for this supplier';
COMMENT ON COLUMN suppliers.reliability_rating IS 'Reliability rating from 0-100';
COMMENT ON COLUMN suppliers.contact_info IS 'JSON object with contact details (email, phone, address)';
