-- Migration: Create role_permissions table for RBAC
-- Task: T004
-- Description: Junction table linking roles to their permissions

CREATE TABLE IF NOT EXISTS role_permissions (
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    granted_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (role_id, permission_id)
);

-- Indexes for performance
CREATE INDEX idx_role_permissions_role_id ON role_permissions(role_id);
CREATE INDEX idx_role_permissions_permission_id ON role_permissions(permission_id);
CREATE INDEX idx_role_permissions_granted_at ON role_permissions(granted_at);

-- Comments for documentation
COMMENT ON TABLE role_permissions IS 'Links roles to their assigned permissions for RBAC';
COMMENT ON COLUMN role_permissions.role_id IS 'Role that has the permission';
COMMENT ON COLUMN role_permissions.permission_id IS 'Permission granted to the role';
COMMENT ON COLUMN role_permissions.granted_at IS 'Timestamp when permission was granted';