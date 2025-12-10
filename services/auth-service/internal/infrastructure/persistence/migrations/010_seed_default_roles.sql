-- Migration: Seed default system roles for RBAC
-- Task: T018
-- Description: Create predefined system roles with hierarchy

-- Insert system roles with hierarchy (Viewer → Analyst → Manager → Admin)
INSERT INTO roles (id, name, description, is_system, parent_role_id, organization_id) VALUES
    ('00000000-0000-0000-0000-000000000010', 'Viewer', 'Read-only access to all services', true, NULL, NULL),
    ('00000000-0000-0000-0000-000000000020', 'Analyst', 'Read all services plus write analytics', true, '00000000-0000-0000-0000-000000000010', NULL),
    ('00000000-0000-0000-0000-000000000030', 'Manager', 'Manage catalog, DDMRP, execution services', true, '00000000-0000-0000-0000-000000000020', NULL),
    ('00000000-0000-0000-0000-000000000040', 'Admin', 'Full system access with all permissions', true, '00000000-0000-0000-0000-000000000030', NULL)
ON CONFLICT (id) DO NOTHING;

-- Insert wildcard admin permission (matches everything)
INSERT INTO permissions (id, code, description, service, resource, action) VALUES
    ('00000000-0000-0000-0000-000000000001', '*:*:*', 'Wildcard permission - grants all access', '*', '*', '*')
ON CONFLICT (code) DO NOTHING;

-- Assign wildcard permission to Admin role
INSERT INTO role_permissions (role_id, permission_id) VALUES
    ('00000000-0000-0000-0000-000000000040', '00000000-0000-0000-0000-000000000001')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Comments
COMMENT ON COLUMN roles.parent_role_id IS 'Role hierarchy: Admin > Manager > Analyst > Viewer';