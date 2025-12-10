-- Migration: Create roles table for RBAC
-- Task: T001
-- Description: Roles define user permissions within organizations

CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
    parent_role_id UUID REFERENCES roles(id) ON DELETE SET NULL,
    is_system BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_role_per_org UNIQUE(name, organization_id)
);

-- Indexes for performance
CREATE INDEX idx_roles_organization_id ON roles(organization_id);
CREATE INDEX idx_roles_parent_role_id ON roles(parent_role_id);
CREATE INDEX idx_roles_is_system ON roles(is_system);
CREATE INDEX idx_roles_name ON roles(name);

-- Trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_roles_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_roles_timestamp
    BEFORE UPDATE ON roles
    FOR EACH ROW
    EXECUTE FUNCTION update_roles_updated_at();

-- Comments for documentation
COMMENT ON TABLE roles IS 'Defines roles for role-based access control';
COMMENT ON COLUMN roles.organization_id IS 'NULL for system-wide roles, UUID for org-specific roles';
COMMENT ON COLUMN roles.parent_role_id IS 'For role hierarchy - child roles inherit parent permissions';
COMMENT ON COLUMN roles.is_system IS 'TRUE for predefined system roles (Admin, Manager, Analyst, Viewer)';