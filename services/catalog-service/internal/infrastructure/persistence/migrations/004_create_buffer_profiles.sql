-- Migration: Create buffer_profiles table
-- Description: Buffer profile templates for DDMRP buffer configuration

SET search_path TO catalog;

CREATE TABLE IF NOT EXISTS buffer_profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    lead_time_factor DECIMAL(5, 2) NOT NULL DEFAULT 1.0,
    variability_factor DECIMAL(5, 2) NOT NULL DEFAULT 1.0,
    target_service_level INT NOT NULL DEFAULT 95,
    organization_id UUID NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(name, organization_id),
    CONSTRAINT chk_lead_time_factor CHECK (lead_time_factor > 0),
    CONSTRAINT chk_variability_factor CHECK (variability_factor > 0),
    CONSTRAINT chk_target_service_level CHECK (target_service_level >= 0 AND target_service_level <= 100)
);

CREATE INDEX idx_buffer_profiles_organization_id ON buffer_profiles(organization_id);
CREATE INDEX idx_buffer_profiles_name ON buffer_profiles(name);

CREATE TRIGGER update_buffer_profiles_updated_at BEFORE UPDATE ON buffer_profiles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

ALTER TABLE products
    ADD CONSTRAINT fk_products_buffer_profile
    FOREIGN KEY (buffer_profile_id) REFERENCES buffer_profiles(id) ON DELETE SET NULL;

COMMENT ON TABLE buffer_profiles IS 'Buffer profile templates for DDMRP buffer sizing';
COMMENT ON COLUMN buffer_profiles.name IS 'Profile name (e.g., "Fast Movers", "Slow Movers")';
COMMENT ON COLUMN buffer_profiles.lead_time_factor IS 'Multiplier for lead time in buffer calculations';
COMMENT ON COLUMN buffer_profiles.variability_factor IS 'Multiplier for demand variability in buffer calculations';
COMMENT ON COLUMN buffer_profiles.target_service_level IS 'Target service level percentage (0-100)';
