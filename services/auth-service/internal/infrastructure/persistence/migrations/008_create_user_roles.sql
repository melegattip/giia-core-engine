-- Migration: Create user_roles table for RBAC
-- Task: T003
-- Description: Junction table linking users to their assigned roles

CREATE TABLE IF NOT EXISTS user_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    assigned_at TIMESTAMP NOT NULL DEFAULT NOW(),
    assigned_by UUID REFERENCES users(id) ON DELETE SET NULL,
    CONSTRAINT unique_user_role UNIQUE(user_id, role_id)
);

-- Indexes for performance
CREATE INDEX idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX idx_user_roles_role_id ON user_roles(role_id);
CREATE INDEX idx_user_roles_assigned_by ON user_roles(assigned_by);
CREATE INDEX idx_user_roles_assigned_at ON user_roles(assigned_at);

-- Comments for documentation
COMMENT ON TABLE user_roles IS 'Links users to their assigned roles for RBAC';
COMMENT ON COLUMN user_roles.user_id IS 'User who has the role assigned';
COMMENT ON COLUMN user_roles.role_id IS 'Role assigned to the user';
COMMENT ON COLUMN user_roles.assigned_at IS 'Timestamp when role was assigned';
COMMENT ON COLUMN user_roles.assigned_by IS 'Admin user who assigned the role (for audit trail)';