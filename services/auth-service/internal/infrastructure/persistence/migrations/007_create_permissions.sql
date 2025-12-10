-- Migration: Create permissions table for RBAC
-- Task: T002
-- Description: Permissions define granular access rights in format "service:resource:action"

CREATE TABLE IF NOT EXISTS permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    service VARCHAR(50) NOT NULL,
    resource VARCHAR(100) NOT NULL,
    action VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_permissions_service ON permissions(service);
CREATE INDEX idx_permissions_code ON permissions(code);
CREATE INDEX idx_permissions_service_resource ON permissions(service, resource);
CREATE INDEX idx_permissions_service_resource_action ON permissions(service, resource, action);

-- Check constraint to ensure code format matches "service:resource:action"
ALTER TABLE permissions ADD CONSTRAINT check_permission_code_format
    CHECK (code ~ '^[a-z_]+:[a-z_*]+:[a-z_*]+$');

-- Comments for documentation
COMMENT ON TABLE permissions IS 'Granular permissions for RBAC in format "service:resource:action"';
COMMENT ON COLUMN permissions.code IS 'Unique permission code (e.g., "catalog:products:read")';
COMMENT ON COLUMN permissions.service IS 'Service name (e.g., "catalog", "ddmrp", "auth")';
COMMENT ON COLUMN permissions.resource IS 'Resource type (e.g., "products", "buffers", "users")';
COMMENT ON COLUMN permissions.action IS 'Action (e.g., "read", "write", "delete")';