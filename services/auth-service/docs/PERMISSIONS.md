# GIIA Permissions Reference

This document lists all available permissions in the GIIA system, organized by service.

## Permission Format

All permissions follow the format: `service:resource:action`

- **Service**: The microservice (e.g., `auth`, `catalog`, `ddmrp`)
- **Resource**: The entity type (e.g., `products`, `buffers`, `users`)
- **Action**: The operation (e.g., `read`, `write`, `delete`)

## System Roles & Permissions

### Admin Role
- **Permission**: `*:*:*` (wildcard - all permissions)
- **Description**: Full system access

### Manager Role
Inherits from Analyst + these permissions:

| Permission | Description |
|------------|-------------|
| `catalog:products:write` | Create and update products |
| `catalog:products:delete` | Delete products |
| `catalog:suppliers:write` | Create and update suppliers |
| `catalog:suppliers:delete` | Delete suppliers |
| `catalog:profiles:write` | Create and update product profiles |
| `catalog:profiles:delete` | Delete product profiles |
| `ddmrp:buffers:write` | Create and update DDMRP buffers |
| `ddmrp:buffers:delete` | Delete DDMRP buffers |
| `ddmrp:calculations:execute` | Execute DDMRP calculations |
| `ddmrp:zones:write` | Update buffer zones |
| `execution:orders:write` | Create and update execution orders |
| `execution:orders:delete` | Delete execution orders |
| `execution:schedules:write` | Create and update schedules |

### Analyst Role
Inherits from Viewer + these permissions:

| Permission | Description |
|------------|-------------|
| `analytics:reports:write` | Create and update analytics reports |
| `analytics:dashboards:write` | Create and update dashboards |

### Viewer Role
Base permissions (read-only access):

| Permission | Description |
|------------|-------------|
| `auth:users:read` | View user information |
| `auth:roles:read` | View roles and permissions |
| `auth:permissions:read` | View permissions |
| `catalog:products:read` | View products |
| `catalog:suppliers:read` | View suppliers |
| `catalog:profiles:read` | View product profiles |
| `ddmrp:buffers:read` | View DDMRP buffers |
| `ddmrp:calculations:read` | View DDMRP calculations |
| `ddmrp:zones:read` | View buffer zones |
| `execution:orders:read` | View execution orders |
| `execution:schedules:read` | View schedules |
| `analytics:reports:read` | View analytics reports |
| `analytics:dashboards:read` | View dashboards |
| `ai_agent:queries:read` | View AI agent queries |
| `ai_agent:models:read` | View AI models |

## Service-Specific Permissions

### Auth Service

| Permission | Description | Default Roles |
|------------|-------------|---------------|
| `auth:users:read` | View user information | Viewer, Analyst, Manager, Admin |
| `auth:users:write` | Create and update users | Admin |
| `auth:users:delete` | Delete users | Admin |
| `auth:roles:read` | View roles and permissions | Viewer, Analyst, Manager, Admin |
| `auth:roles:write` | Create and update roles | Admin |
| `auth:roles:delete` | Delete roles | Admin |
| `auth:permissions:read` | View permissions | Viewer, Analyst, Manager, Admin |
| `auth:permissions:write` | Create and update permissions | Admin |

### Catalog Service

| Permission | Description | Default Roles |
|------------|-------------|---------------|
| `catalog:products:read` | View products | Viewer, Analyst, Manager, Admin |
| `catalog:products:write` | Create and update products | Manager, Admin |
| `catalog:products:delete` | Delete products | Manager, Admin |
| `catalog:suppliers:read` | View suppliers | Viewer, Analyst, Manager, Admin |
| `catalog:suppliers:write` | Create and update suppliers | Manager, Admin |
| `catalog:suppliers:delete` | Delete suppliers | Manager, Admin |
| `catalog:profiles:read` | View product profiles | Viewer, Analyst, Manager, Admin |
| `catalog:profiles:write` | Create and update product profiles | Manager, Admin |
| `catalog:profiles:delete` | Delete product profiles | Manager, Admin |

### DDMRP Service

| Permission | Description | Default Roles |
|------------|-------------|---------------|
| `ddmrp:buffers:read` | View DDMRP buffers | Viewer, Analyst, Manager, Admin |
| `ddmrp:buffers:write` | Create and update DDMRP buffers | Manager, Admin |
| `ddmrp:buffers:delete` | Delete DDMRP buffers | Manager, Admin |
| `ddmrp:calculations:read` | View DDMRP calculations | Viewer, Analyst, Manager, Admin |
| `ddmrp:calculations:execute` | Execute DDMRP calculations | Manager, Admin |
| `ddmrp:zones:read` | View buffer zones | Viewer, Analyst, Manager, Admin |
| `ddmrp:zones:write` | Update buffer zones | Manager, Admin |

### Execution Service

| Permission | Description | Default Roles |
|------------|-------------|---------------|
| `execution:orders:read` | View execution orders | Viewer, Analyst, Manager, Admin |
| `execution:orders:write` | Create and update execution orders | Manager, Admin |
| `execution:orders:delete` | Delete execution orders | Manager, Admin |
| `execution:schedules:read` | View schedules | Viewer, Analyst, Manager, Admin |
| `execution:schedules:write` | Create and update schedules | Manager, Admin |

### Analytics Service

| Permission | Description | Default Roles |
|------------|-------------|---------------|
| `analytics:reports:read` | View analytics reports | Viewer, Analyst, Manager, Admin |
| `analytics:reports:write` | Create and update reports | Analyst, Manager, Admin |
| `analytics:dashboards:read` | View dashboards | Viewer, Analyst, Manager, Admin |
| `analytics:dashboards:write` | Create and update dashboards | Analyst, Manager, Admin |

### AI Agent Service

| Permission | Description | Default Roles |
|------------|-------------|---------------|
| `ai_agent:queries:read` | View AI agent queries | Viewer, Analyst, Manager, Admin |
| `ai_agent:queries:write` | Execute AI agent queries | Analyst, Manager, Admin |
| `ai_agent:models:read` | View AI models | Viewer, Analyst, Manager, Admin |
| `ai_agent:models:write` | Update AI models | Admin |

## Wildcard Patterns

### Service-Level Wildcards

| Pattern | Description |
|---------|-------------|
| `catalog:*:*` | All catalog service permissions |
| `ddmrp:*:*` | All DDMRP service permissions |
| `auth:*:*` | All auth service permissions |

### Resource-Level Wildcards

| Pattern | Description |
|---------|-------------|
| `catalog:products:*` | All product operations (read, write, delete) |
| `ddmrp:buffers:*` | All buffer operations |

### Action-Level Wildcards

| Pattern | Description |
|---------|-------------|
| `*:*:read` | Read-only access to all services and resources |
| `*:*:write` | Write access to all services and resources |

## Custom Roles

Organizations can create custom roles by combining any of these permissions. Custom roles can also inherit from system roles.

### Example: Custom "Data Analyst" Role

```json
{
  "name": "Data Analyst",
  "description": "Analyst with limited catalog access",
  "parent_role_id": "viewer-role-id",
  "permissions": [
    "catalog:products:read",
    "analytics:reports:read",
    "analytics:reports:write",
    "analytics:dashboards:write"
  ]
}
```

### Example: Custom "Inventory Manager" Role

```json
{
  "name": "Inventory Manager",
  "description": "Manages catalog and DDMRP buffers",
  "parent_role_id": "analyst-role-id",
  "permissions": [
    "catalog:products:write",
    "catalog:suppliers:write",
    "ddmrp:buffers:read",
    "ddmrp:buffers:write"
  ]
}
```

## Permission Seeding

All permissions are seeded automatically via:

```bash
go run services/auth-service/scripts/seed_permissions.go
```

This script creates all permissions listed above and assigns them to the appropriate system roles.

## Adding New Permissions

To add new permissions:

1. Update `scripts/seed_permissions.go` with new permission definitions
2. Run the seed script to create permissions in database
3. Assign permissions to appropriate system roles
4. Update this documentation
5. Run tests to verify permission checks work correctly

## Permission Checking Examples

### Via HTTP Middleware

```go
router.POST("/api/v1/products",
    permissionMiddleware.RequirePermission("catalog:products:write"),
    productHandler.CreateProduct,
)
```

### Via Permission Service

```go
allowed, err := checkPermissionUseCase.Execute(ctx, userID, "catalog:products:write")
if !allowed {
    return errors.NewForbidden("insufficient permissions")
}
```

### Batch Check

```go
results, err := batchCheckUseCase.Execute(ctx, userID, []string{
    "catalog:products:read",
    "catalog:products:write",
    "ddmrp:buffers:read",
})
// results: map[string]bool
```

---

**Last Updated**: 2025-12-09
**Total Permissions**: 40+
**System Roles**: 4 (Viewer, Analyst, Manager, Admin)
