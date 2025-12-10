# RBAC Design Documentation

## Overview

This document describes the Role-Based Access Control (RBAC) system implemented in the GIIA Auth Service. The RBAC system provides fine-grained access control through roles, permissions, and role hierarchy.

## Architecture

### Components

```
┌─────────────────────────────────────────────────────────────┐
│                      HTTP Layer                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐  │
│  │ Role Handler │  │ Permission   │  │ Permission       │  │
│  │              │  │ Handler      │  │ Middleware       │  │
│  └──────────────┘  └──────────────┘  └──────────────────┘  │
└─────────────────────────────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────┐
│                    Use Cases Layer                           │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐  │
│  │ Role         │  │ RBAC         │  │ Permission       │  │
│  │ Management   │  │ Permission   │  │ Seeding          │  │
│  │              │  │ Checking     │  │                  │  │
│  └──────────────┘  └──────────────┘  └──────────────────┘  │
└─────────────────────────────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────┐
│                 Infrastructure Layer                         │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐  │
│  │ Role Repo    │  │ Permission   │  │ Redis Cache      │  │
│  │ (PostgreSQL) │  │ Repo         │  │ (5min TTL)       │  │
│  └──────────────┘  └──────────────┘  └──────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

### Database Schema

#### Roles Table
```sql
CREATE TABLE roles (
    id UUID PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    organization_id UUID REFERENCES organizations(id),  -- NULL for system roles
    parent_role_id UUID REFERENCES roles(id),           -- Role hierarchy
    is_system BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    UNIQUE(name, organization_id)
);
```

#### Permissions Table
```sql
CREATE TABLE permissions (
    id UUID PRIMARY KEY,
    code VARCHAR(255) NOT NULL UNIQUE,  -- "service:resource:action"
    description TEXT,
    service VARCHAR(50) NOT NULL,
    resource VARCHAR(100) NOT NULL,
    action VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL
);
```

#### User Roles (Junction)
```sql
CREATE TABLE user_roles (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id),
    role_id UUID NOT NULL REFERENCES roles(id),
    assigned_at TIMESTAMP NOT NULL,
    assigned_by UUID REFERENCES users(id),  -- Audit trail
    UNIQUE(user_id, role_id)
);
```

#### Role Permissions (Junction)
```sql
CREATE TABLE role_permissions (
    role_id UUID NOT NULL REFERENCES roles(id),
    permission_id UUID NOT NULL REFERENCES permissions(id),
    granted_at TIMESTAMP NOT NULL,
    PRIMARY KEY (role_id, permission_id)
);
```

## Permission System

### Permission Format

Permissions follow the format: `service:resource:action`

Examples:
- `catalog:products:read` - Read products in catalog service
- `ddmrp:buffers:write` - Create/update DDMRP buffers
- `auth:roles:delete` - Delete roles
- `*:*:*` - Wildcard (all permissions)

### Wildcard Permissions

The system supports wildcard permissions for flexible access control:

- `*:*:*` - All permissions (Admin role)
- `catalog:*:*` - All catalog service permissions
- `catalog:products:*` - All product actions
- `*:*:read` - All read permissions across all services

### Permission Checking Algorithm

```go
func CheckPermission(userPermissions []string, requiredPermission string) bool {
    // 1. Check for wildcard admin permission
    if contains(userPermissions, "*:*:*") {
        return true
    }

    // 2. Check for exact match
    if contains(userPermissions, requiredPermission) {
        return true
    }

    // 3. Check for wildcard patterns
    for _, userPerm := range userPermissions {
        if matchesWildcard(userPerm, requiredPermission) {
            return true
        }
    }

    return false
}
```

## Role Hierarchy

### System Roles

The RBAC system defines 4 predefined system roles with inheritance:

```
┌─────────────┐
│   Admin     │  ← Full access (*:*:*)
└──────┬──────┘
       │ inherits
┌──────▼──────┐
│   Manager   │  ← Catalog, DDMRP, Execution (read + write)
└──────┬──────┘
       │ inherits
┌──────▼──────┐
│   Analyst   │  ← All services (read) + Analytics (write)
└──────┬──────┘
       │ inherits
┌──────▼──────┐
│   Viewer    │  ← All services (read only)
└─────────────┘
```

### Role Inheritance Resolution

When a user has a role, they inherit all permissions from:
1. The role itself
2. The parent role (if `parent_role_id` is set)
3. The grandparent role (recursively)

**Example:**
- User has "Manager" role
- Manager inherits from "Analyst"
- Analyst inherits from "Viewer"
- User receives permissions from: Manager + Analyst + Viewer

### Circular Reference Prevention

The system prevents circular role hierarchies:
```go
// ❌ Invalid: Role A → Role B → Role A
// ✅ Valid: Viewer → Analyst → Manager → Admin
```

## Caching Strategy

### Redis Cache

User permissions are cached in Redis with:
- **Key Format**: `user:{user_id}:permissions`
- **TTL**: 5 minutes
- **Data Format**: JSON array of permission codes

### Cache Invalidation

Cache is invalidated when:
1. User is assigned a new role
2. User role is removed
3. Role permissions are updated (invalidates all users with that role)
4. Role is deleted (invalidates all users with that role)

### Cache Hit/Miss Flow

```
┌────────────────────┐
│ Permission Check   │
└─────────┬──────────┘
          │
          ▼
    ┌─────────────┐        Cache Hit
    │ Check Redis ├────────────────────┐
    │    Cache    │                    │
    └─────┬───────┘                    │
          │ Cache Miss                 │
          ▼                            ▼
    ┌─────────────┐              ┌─────────┐
    │ Get User    │              │ Return  │
    │ Roles from  │              │ Result  │
    │  Database   │              └─────────┘
    └─────┬───────┘
          │
          ▼
    ┌─────────────┐
    │ Resolve     │
    │ Hierarchy   │
    └─────┬───────┘
          │
          ▼
    ┌─────────────┐
    │ Cache Result│
    │ (5 min TTL) │
    └─────┬───────┘
          │
          ▼
    ┌─────────────┐
    │ Return      │
    │ Result      │
    └─────────────┘
```

## Multi-Tenancy

### System Roles vs Organization Roles

- **System Roles** (`organization_id = NULL`):
  - Predefined: Admin, Manager, Analyst, Viewer
  - Cannot be modified or deleted
  - Available to all organizations

- **Organization Roles** (`organization_id = <UUID>`):
  - Created by organization admins
  - Scoped to specific organization
  - Can inherit from system roles

### Tenant Isolation

Roles are filtered by organization:
```sql
SELECT * FROM roles
WHERE organization_id = ? OR is_system = true
ORDER BY name;
```

## Permission Middleware

### Usage

Protect HTTP endpoints with permission checks:

```go
// Require specific permission
router.POST("/api/v1/roles",
    permissionMiddleware.RequirePermission("auth:roles:write"),
    roleHandler.CreateRole,
)

// Require any of multiple permissions
router.GET("/api/v1/data",
    permissionMiddleware.RequireAnyPermission(
        "catalog:products:read",
        "ddmrp:buffers:read",
    ),
    dataHandler.GetData,
)
```

### Middleware Flow

```
1. Extract user_id from JWT (via TenantMiddleware)
2. Get user permissions (from cache or DB)
3. Check if required permission is granted
4. If yes → Continue to handler
5. If no → Return 403 Forbidden
```

## API Endpoints

### Role Management

#### Create Role
```http
POST /api/v1/roles
Authorization: Bearer <token>
Permission: auth:roles:write

{
  "name": "Custom Manager",
  "description": "Custom role for managers",
  "organization_id": "org-uuid",
  "parent_role_id": "analyst-uuid",
  "permission_ids": ["perm1-uuid", "perm2-uuid"]
}
```

#### Update Role
```http
PUT /api/v1/roles/{roleId}
Authorization: Bearer <token>
Permission: auth:roles:write

{
  "name": "Updated Name",
  "description": "Updated description",
  "permission_ids": ["perm1-uuid", "perm3-uuid"]
}
```

#### Delete Role
```http
DELETE /api/v1/roles/{roleId}
Authorization: Bearer <token>
Permission: auth:roles:delete
```

#### Assign Role to User
```http
POST /api/v1/roles/assign
Authorization: Bearer <token>
Permission: auth:roles:write

{
  "user_id": "user-uuid",
  "role_id": "role-uuid"
}
```

### Permission Checking

#### Check Single Permission
```http
POST /api/v1/permissions/check
Authorization: Bearer <token>

{
  "user_id": "user-uuid",
  "permission": "catalog:products:write"
}

Response:
{
  "allowed": true
}
```

#### Batch Check Permissions
```http
POST /api/v1/permissions/batch-check
Authorization: Bearer <token>

{
  "user_id": "user-uuid",
  "permissions": [
    "catalog:products:read",
    "catalog:products:write",
    "ddmrp:buffers:read"
  ]
}

Response:
{
  "results": {
    "catalog:products:read": true,
    "catalog:products:write": false,
    "ddmrp:buffers:read": true
  }
}
```

## Performance Characteristics

### Permission Check Latency

- **Cache Hit**: < 10ms (p95)
- **Cache Miss**: < 50ms (p95)
- **Cache Hit Rate**: > 95% (normal operation)

### Optimization Techniques

1. **Redis Caching**: 5-minute TTL reduces database queries
2. **Batch Permission Loading**: Single query loads all user roles
3. **Recursive Hierarchy Resolution**: Deduplicates permissions
4. **Wildcard Short-Circuit**: Stops checking if `*:*:*` found

## Security Considerations

### Audit Logging

All permission checks are logged with:
- `user_id`
- `permission` requested
- `allowed` (true/false)
- `timestamp`
- `endpoint` accessed

### Permission Validation

- User must be authenticated (valid JWT)
- User must have assigned role
- Role must have required permission (or inherit it)
- System roles cannot be deleted or modified

### Edge Cases Handled

1. **Role deleted while users have it**: User-role assignments cascade delete
2. **Circular hierarchy**: Prevented during role creation
3. **Cache unavailable**: Falls back to database query
4. **Permission system unavailable**: Fails closed (deny access)

## Migration & Seeding

### Setup Process

1. Run migrations:
   ```bash
   006_create_roles.sql
   007_create_permissions.sql
   008_create_user_roles.sql
   009_create_role_permissions.sql
   010_seed_default_roles.sql
   ```

2. Seed permissions:
   ```bash
   go run services/auth-service/scripts/seed_permissions.go
   ```

3. Verify setup:
   ```sql
   SELECT COUNT(*) FROM permissions;  -- Should show ~40 permissions
   SELECT COUNT(*) FROM roles WHERE is_system = true;  -- Should show 4 roles
   ```

## Testing

### Unit Tests

- Role hierarchy resolution
- Permission matching (exact + wildcard)
- Cache invalidation logic
- Circular reference detection

### Integration Tests

- Full permission check flow
- Role assignment and verification
- Cache hit/miss scenarios
- Multi-role permission aggregation

## Future Enhancements

1. **Resource-Level Permissions** (P3):
   - `ddmrp:buffers:write:buffer-123`
   - Scope permissions to specific resource instances

2. **Permission Groups**:
   - Bundle related permissions
   - Simplify role creation

3. **Audit Dashboard**:
   - View permission check history
   - Analyze access patterns
   - Compliance reporting

4. **Dynamic Permission Loading**:
   - Hot-reload permissions without restart
   - Service discovery for new permissions

---

**Last Updated**: 2025-12-09
**Version**: 1.0
**Status**: Production Ready (P1 Features)
