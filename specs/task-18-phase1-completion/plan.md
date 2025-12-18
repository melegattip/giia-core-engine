# Task 18: Phase 1 Completion - Implementation Plan

**Task ID**: task-18-phase1-completion
**Phase**: 2A - Complete to 100%
**Priority**: P1 (High)
**Estimated Duration**: 1 week (5 days)
**Dependencies**: Tasks 11, 12, 13 (in progress/complete)

---

## 1. Technical Context

### Current State
- **Phase 1 Completion**: 93%
- **Auth Service**: 95% complete (missing: seed data, audit logging)
- **Catalog Service**: 85% complete (missing: gRPC proto files)
- **Shared Packages**: 85% complete (missing: DLQ consumer, event schema validation)
- **Infrastructure**: 95% complete (missing: health checks, graceful shutdown)

### Technology Stack
- **Language**: Go 1.23.4
- **Architecture**: Monorepo Microservices (Clean Architecture)
- **Database**: PostgreSQL 16 (GORM ORM)
- **Event Streaming**: NATS JetStream 2.x
- **Cache**: Redis 7
- **gRPC**: Protocol Buffers v3
- **Container Orchestration**: Kubernetes (Minikube local, K8s production)
- **CI/CD**: GitHub Actions

### Key Design Decisions
- **RBAC Seed Data**: Idempotent scripts with `ON CONFLICT DO NOTHING`
- **Audit Logging**: Synchronous database logging + asynchronous event publishing
- **gRPC Reflection**: Enabled on all services for debugging/tooling
- **DLQ Processing**: Standalone consumer with exponential backoff retry (max 3 attempts)
- **Health Checks**: Separate liveness (always OK) and readiness (checks dependencies)
- **Graceful Shutdown**: 30-second timeout for in-flight requests

---

## 2. Project Structure

### Files to Create

```
giia-core-engine/
├── services/auth-service/
│   ├── scripts/
│   │   ├── seed-roles.sh                                [NEW] Seed RBAC data
│   │   └── README.md                                    [NEW] Script documentation
│   ├── internal/
│   │   ├── core/
│   │   │   ├── domain/
│   │   │   │   └── audit_log.go                        [NEW] Audit log entity
│   │   │   ├── providers/
│   │   │   │   └── audit_logger.go                     [NEW] Audit logger interface
│   │   │   └── usecases/
│   │   │       └── auth/
│   │   │           └── check_permission.go             [MODIFY] Add audit logging
│   │   └── infrastructure/
│   │       ├── adapters/
│   │       │   └── audit/
│   │       │       └── audit_logger.go                 [NEW] Audit logger implementation
│   │       ├── entrypoints/
│   │       │   └── http/
│   │       │       └── health_handlers.go              [NEW] Health check endpoints
│   │       └── grpc/
│   │           └── server/
│   │               └── server.go                       [MODIFY] Enable reflection
│   ├── test/
│   │   └── performance/
│   │       └── permission_check_bench_test.go          [NEW] Performance benchmarks
│   └── cmd/
│       └── server/
│           └── main.go                                 [MODIFY] Graceful shutdown
│
├── services/catalog-service/
│   ├── api/
│   │   └── proto/
│   │       └── catalog/
│   │           └── v1/
│   │               ├── catalog.proto                   [NEW] Catalog gRPC API
│   │               ├── catalog.pb.go                   [GENERATED]
│   │               └── catalog_grpc.pb.go              [GENERATED]
│   ├── internal/
│   │   ├── core/
│   │   │   └── usecases/
│   │   │       ├── supplier/
│   │   │       │   ├── create_supplier.go              [MODIFY] Publish supplier.created
│   │   │       │   └── update_supplier.go              [MODIFY] Publish supplier.updated
│   │   │       └── buffer_profile/
│   │   │           ├── create_buffer_profile.go        [MODIFY] Publish events
│   │   │           └── update_buffer_profile.go        [MODIFY] Publish events
│   │   └── infrastructure/
│   │       ├── entrypoints/
│   │       │   ├── http/
│   │       │   │   └── health_handlers.go              [NEW] Health check endpoints
│   │       │   └── grpc/
│   │       │       └── catalog_handlers.go             [NEW] gRPC handlers
│   │       └── grpc/
│   │           └── server/
│   │               └── server.go                       [NEW] gRPC server with reflection
│   └── cmd/
│       └── server/
│           └── main.go                                 [MODIFY] Start gRPC + graceful shutdown
│
├── pkg/
│   ├── events/
│   │   ├── schema_validator.go                         [NEW] Event schema validation
│   │   └── dlq_consumer.go                             [NEW] DLQ consumer
│   └── health/
│       ├── checker.go                                  [NEW] Health check utilities
│       └── readiness.go                                [NEW] Readiness check helper
│
├── examples/
│   └── grpc-clients/
│       ├── auth-client/
│       │   ├── main.go                                 [NEW] Auth client example
│       │   └── README.md                               [NEW] Usage instructions
│       └── catalog-client/
│           ├── main.go                                 [NEW] Catalog client example
│           └── README.md                               [NEW] Usage instructions
│
├── docs/
│   ├── architecture/
│   │   ├── sequence-diagrams.md                        [NEW] User flows
│   │   └── component-diagrams.md                       [UPDATE] Add catalog gRPC
│   ├── api/
│   │   ├── openapi/
│   │   │   ├── auth-service.yaml                       [NEW] OpenAPI spec
│   │   │   └── catalog-service.yaml                    [NEW] OpenAPI spec
│   │   └── proto/
│   │       └── README.md                               [NEW] Proto documentation
│   └── runbooks/
│       ├── README.md                                   [NEW] Runbooks overview
│       ├── deployment.md                               [NEW] Deployment runbook
│       ├── rollback.md                                 [NEW] Rollback runbook
│       ├── scaling.md                                  [NEW] Scaling runbook
│       └── troubleshooting.md                          [NEW] Troubleshooting guide
│
├── deploy/
│   ├── environments/
│   │   ├── .env.dev                                    [NEW] Development config
│   │   ├── .env.staging                                [NEW] Staging config
│   │   └── .env.prod                                   [NEW] Production config
│   └── k8s/
│       ├── auth-service/
│       │   ├── deployment.yaml                         [MODIFY] Add health probes
│       │   └── service.yaml                            [MODIFY] Add gRPC port
│       └── catalog-service/
│           ├── deployment.yaml                         [MODIFY] Add health probes + gRPC
│           ├── service.yaml                            [MODIFY] Add gRPC port
│           └── grpc-service.yaml                       [NEW] gRPC-specific service
│
└── scripts/
    └── generate-proto-catalog.sh                       [NEW] Generate catalog proto

```

---

## 3. Implementation Phases

### Phase 1: RBAC Seed Data and Audit Logging (Day 1)

#### T001: Create Seed Data Script
**Estimated Time**: 2 hours

**File**: `services/auth-service/scripts/seed-roles.sh`

```bash
#!/bin/bash
set -e

# Load environment
source .env.dev

echo "Starting RBAC seed data..."

# Seed roles, permissions, and role-permission associations
psql "$DATABASE_URL" << 'EOF'
BEGIN;

-- Insert standard roles
INSERT INTO roles (id, organization_id, name, description, level, created_at, updated_at) VALUES
  (gen_random_uuid(), NULL, 'Admin', 'Full system administrator with all permissions', 1, NOW(), NOW()),
  (gen_random_uuid(), NULL, 'Manager', 'Operations manager with write access to core entities', 2, NOW(), NOW()),
  (gen_random_uuid(), NULL, 'Analyst', 'Read-only analyst for data analysis', 3, NOW(), NOW()),
  (gen_random_uuid(), NULL, 'Viewer', 'Read-only viewer for basic viewing', 4, NOW(), NOW())
ON CONFLICT (name) DO NOTHING;

-- Insert standard permissions
INSERT INTO permissions (id, name, description, resource, action, created_at, updated_at) VALUES
  -- User permissions
  (gen_random_uuid(), 'users:read', 'Read user data', 'users', 'read', NOW(), NOW()),
  (gen_random_uuid(), 'users:write', 'Create and update users', 'users', 'write', NOW(), NOW()),
  (gen_random_uuid(), 'users:delete', 'Delete users', 'users', 'delete', NOW(), NOW()),
  (gen_random_uuid(), 'users:activate', 'Activate/deactivate users', 'users', 'activate', NOW(), NOW()),

  -- Catalog permissions
  (gen_random_uuid(), 'catalog:read', 'Read catalog data', 'catalog', 'read', NOW(), NOW()),
  (gen_random_uuid(), 'catalog:write', 'Create and update catalog items', 'catalog', 'write', NOW(), NOW()),
  (gen_random_uuid(), 'catalog:delete', 'Delete catalog items', 'catalog', 'delete', NOW(), NOW()),

  -- Buffer permissions
  (gen_random_uuid(), 'buffers:read', 'Read buffer data', 'buffers', 'read', NOW(), NOW()),
  (gen_random_uuid(), 'buffers:write', 'Create and update buffers', 'buffers', 'write', NOW(), NOW()),
  (gen_random_uuid(), 'buffers:delete', 'Delete buffers', 'buffers', 'delete', NOW(), NOW()),

  -- Order permissions
  (gen_random_uuid(), 'orders:read', 'Read orders', 'orders', 'read', NOW(), NOW()),
  (gen_random_uuid(), 'orders:write', 'Create and update orders', 'orders', 'write', NOW(), NOW()),
  (gen_random_uuid(), 'orders:delete', 'Delete orders', 'orders', 'delete', NOW(), NOW()),
  (gen_random_uuid(), 'orders:approve', 'Approve orders', 'orders', 'approve', NOW(), NOW()),

  -- Report permissions
  (gen_random_uuid(), 'reports:read', 'View reports', 'reports', 'read', NOW(), NOW()),
  (gen_random_uuid(), 'reports:write', 'Create custom reports', 'reports', 'write', NOW(), NOW()),

  -- Settings permissions
  (gen_random_uuid(), 'settings:read', 'View system settings', 'settings', 'read', NOW(), NOW()),
  (gen_random_uuid(), 'settings:write', 'Modify system settings', 'settings', 'write', NOW(), NOW())
ON CONFLICT (name) DO NOTHING;

-- Assign permissions to Admin role (all permissions)
INSERT INTO role_permissions (role_id, permission_id, created_at)
SELECT
  (SELECT id FROM roles WHERE name = 'Admin'),
  p.id,
  NOW()
FROM permissions p
ON CONFLICT DO NOTHING;

-- Assign permissions to Manager role
INSERT INTO role_permissions (role_id, permission_id, created_at)
SELECT
  (SELECT id FROM roles WHERE name = 'Manager'),
  p.id,
  NOW()
FROM permissions p
WHERE p.name IN (
  'users:read', 'users:write',
  'catalog:read', 'catalog:write',
  'buffers:read', 'buffers:write',
  'orders:read', 'orders:write', 'orders:approve',
  'reports:read'
)
ON CONFLICT DO NOTHING;

-- Assign permissions to Analyst role
INSERT INTO role_permissions (role_id, permission_id, created_at)
SELECT
  (SELECT id FROM roles WHERE name = 'Analyst'),
  p.id,
  NOW()
FROM permissions p
WHERE p.name IN (
  'catalog:read',
  'buffers:read',
  'orders:read',
  'reports:read'
)
ON CONFLICT DO NOTHING;

-- Assign permissions to Viewer role
INSERT INTO role_permissions (role_id, permission_id, created_at)
SELECT
  (SELECT id FROM roles WHERE name = 'Viewer'),
  p.id,
  NOW()
FROM permissions p
WHERE p.name IN (
  'catalog:read',
  'buffers:read',
  'orders:read'
)
ON CONFLICT DO NOTHING;

COMMIT;

-- Display summary
SELECT
  r.name AS role,
  r.level,
  COUNT(rp.permission_id) AS permission_count
FROM roles r
LEFT JOIN role_permissions rp ON r.id = rp.role_id
WHERE r.organization_id IS NULL
GROUP BY r.name, r.level
ORDER BY r.level;

EOF

echo "RBAC seed data completed successfully!"
```

**Testing**:
```bash
# Run seed script
bash services/auth-service/scripts/seed-roles.sh

# Verify in database
psql $DATABASE_URL -c "SELECT name, level FROM roles WHERE organization_id IS NULL ORDER BY level;"
```

#### T002: Implement Audit Logging

**File**: `services/auth-service/internal/core/domain/audit_log.go`

```go
package domain

import (
	"time"

	"github.com/google/uuid"
)

type AuditLog struct {
	ID             uuid.UUID
	OrganizationID uuid.UUID
	UserID         uuid.UUID
	Action         AuditAction
	Resource       string
	ResourceID     *uuid.UUID
	Result         AuditResult
	IPAddress      string
	UserAgent      string
	RequestID      string
	Metadata       map[string]interface{}
	CreatedAt      time.Time
}

type AuditAction string

const (
	AuditActionPermissionCheck AuditAction = "permission.check"
	AuditActionRoleAssign      AuditAction = "role.assign"
	AuditActionRoleRevoke      AuditAction = "role.revoke"
	AuditActionUserCreate      AuditAction = "user.create"
	AuditActionUserUpdate      AuditAction = "user.update"
	AuditActionUserDelete      AuditAction = "user.delete"
	AuditActionLogin           AuditAction = "user.login"
	AuditActionLogout          AuditAction = "user.logout"
)

type AuditResult string

const (
	AuditResultAllowed AuditResult = "allowed"
	AuditResultDenied  AuditResult = "denied"
	AuditResultSuccess AuditResult = "success"
	AuditResultFailure AuditResult = "failure"
)
```

**File**: `services/auth-service/internal/core/providers/audit_logger.go`

```go
package providers

import (
	"context"

	"github.com/giia/giia-core-engine/services/auth-service/internal/core/domain"
	"github.com/google/uuid"
)

type AuditLogger interface {
	LogPermissionCheck(ctx context.Context, userID uuid.UUID, permission string, allowed bool, metadata map[string]interface{}) error
	LogRoleAssignment(ctx context.Context, userID uuid.UUID, roleID uuid.UUID, action domain.AuditAction) error
	LogUserAction(ctx context.Context, action domain.AuditAction, userID uuid.UUID, result domain.AuditResult, metadata map[string]interface{}) error
}
```

**File**: `services/auth-service/internal/infrastructure/adapters/audit/audit_logger.go`

```go
package audit

import (
	"context"
	"fmt"

	"github.com/giia/giia-core-engine/pkg/logger"
	"github.com/giia/giia-core-engine/services/auth-service/internal/core/domain"
	"github.com/giia/giia-core-engine/services/auth-service/internal/core/providers"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type auditLogger struct {
	db     *gorm.DB
	logger logger.Logger
}

func NewAuditLogger(db *gorm.DB, logger logger.Logger) providers.AuditLogger {
	return &auditLogger{
		db:     db,
		logger: logger,
	}
}

func (a *auditLogger) LogPermissionCheck(ctx context.Context, userID uuid.UUID, permission string, allowed bool, metadata map[string]interface{}) error {
	result := domain.AuditResultDenied
	if allowed {
		result = domain.AuditResultAllowed
	}

	// Extract request metadata from context
	requestID := getRequestID(ctx)
	ipAddress := getIPAddress(ctx)
	userAgent := getUserAgent(ctx)
	orgID := getOrganizationID(ctx)

	auditLog := domain.AuditLog{
		ID:             uuid.New(),
		OrganizationID: orgID,
		UserID:         userID,
		Action:         domain.AuditActionPermissionCheck,
		Resource:       permission,
		Result:         result,
		IPAddress:      ipAddress,
		UserAgent:      userAgent,
		RequestID:      requestID,
		Metadata:       metadata,
	}

	// Save to database
	if err := a.db.WithContext(ctx).Create(&auditLog).Error; err != nil {
		a.logger.Error(ctx, err, "Failed to save audit log", logger.Tags{
			"user_id":    userID.String(),
			"permission": permission,
			"allowed":    allowed,
		})
		return fmt.Errorf("failed to save audit log: %w", err)
	}

	// Also log to structured logger
	a.logger.Info(ctx, "Permission check", logger.Tags{
		"user_id":    userID.String(),
		"permission": permission,
		"allowed":    allowed,
		"request_id": requestID,
	})

	return nil
}

func (a *auditLogger) LogRoleAssignment(ctx context.Context, userID uuid.UUID, roleID uuid.UUID, action domain.AuditAction) error {
	requestID := getRequestID(ctx)
	ipAddress := getIPAddress(ctx)
	orgID := getOrganizationID(ctx)

	auditLog := domain.AuditLog{
		ID:             uuid.New(),
		OrganizationID: orgID,
		UserID:         userID,
		Action:         action,
		Resource:       "role",
		ResourceID:     &roleID,
		Result:         domain.AuditResultSuccess,
		IPAddress:      ipAddress,
		RequestID:      requestID,
	}

	if err := a.db.WithContext(ctx).Create(&auditLog).Error; err != nil {
		a.logger.Error(ctx, err, "Failed to save role assignment audit log")
		return fmt.Errorf("failed to save audit log: %w", err)
	}

	return nil
}

func (a *auditLogger) LogUserAction(ctx context.Context, action domain.AuditAction, userID uuid.UUID, result domain.AuditResult, metadata map[string]interface{}) error {
	requestID := getRequestID(ctx)
	ipAddress := getIPAddress(ctx)
	userAgent := getUserAgent(ctx)
	orgID := getOrganizationID(ctx)

	auditLog := domain.AuditLog{
		ID:             uuid.New(),
		OrganizationID: orgID,
		UserID:         userID,
		Action:         action,
		Result:         result,
		IPAddress:      ipAddress,
		UserAgent:      userAgent,
		RequestID:      requestID,
		Metadata:       metadata,
	}

	if err := a.db.WithContext(ctx).Create(&auditLog).Error; err != nil {
		a.logger.Error(ctx, err, "Failed to save user action audit log")
		return fmt.Errorf("failed to save audit log: %w", err)
	}

	return nil
}

// Helper functions to extract context values
func getRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value("request_id").(string); ok {
		return requestID
	}
	return ""
}

func getIPAddress(ctx context.Context) string {
	if ip, ok := ctx.Value("ip_address").(string); ok {
		return ip
	}
	return ""
}

func getUserAgent(ctx context.Context) string {
	if ua, ok := ctx.Value("user_agent").(string); ok {
		return ua
	}
	return ""
}

func getOrganizationID(ctx context.Context) uuid.UUID {
	if orgID, ok := ctx.Value("organization_id").(uuid.UUID); ok {
		return orgID
	}
	return uuid.Nil
}
```

**Database Migration**:
```sql
-- services/auth-service/migrations/YYYYMMDDHHMMSS_create_audit_logs.up.sql
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY,
    organization_id UUID NOT NULL REFERENCES organizations(id),
    user_id UUID NOT NULL REFERENCES users(id),
    action VARCHAR(100) NOT NULL,
    resource VARCHAR(200) NOT NULL,
    resource_id UUID,
    result VARCHAR(50) NOT NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    request_id VARCHAR(100),
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_org_user ON audit_logs(organization_id, user_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at DESC);
CREATE INDEX idx_audit_logs_request_id ON audit_logs(request_id) WHERE request_id IS NOT NULL;
```

#### T003: Performance Testing

**File**: `services/auth-service/test/performance/permission_check_bench_test.go`

```go
package performance

import (
	"context"
	"testing"

	"github.com/giia/giia-core-engine/services/auth-service/internal/core/usecases/auth"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func BenchmarkCheckPermission_Cached(b *testing.B) {
	// Setup
	ctx := context.Background()
	useCase := setupCheckPermissionUseCase() // Sets up with Redis cache
	userID := uuid.New()
	permission := "catalog:read"

	// Warm up cache
	_, _ = useCase.Execute(ctx, userID, permission)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		allowed, err := useCase.Execute(ctx, userID, permission)
		assert.NoError(b, err)
		assert.True(b, allowed)
	}
}

func BenchmarkCheckPermission_Uncached(b *testing.B) {
	// Setup without cache
	ctx := context.Background()
	useCase := setupCheckPermissionUseCaseNoCache()
	userID := uuid.New()
	permission := "catalog:read"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		allowed, err := useCase.Execute(ctx, userID, permission)
		assert.NoError(b, err)
		assert.True(b, allowed)
	}
}

// Target: <10ms p95 (cached), <50ms p95 (uncached)
// Run: go test -bench=BenchmarkCheckPermission -benchtime=10s -benchmem
```

---

### Phase 2: gRPC Reflection and Client Examples (Day 2)

#### T004: Enable gRPC Reflection

**File**: `services/auth-service/internal/infrastructure/grpc/server/server.go`

```go
package server

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pkgLogger "github.com/giia/giia-core-engine/pkg/logger"
	authpb "github.com/giia/giia-core-engine/services/auth-service/api/proto/auth/v1"
	"github.com/giia/giia-core-engine/services/auth-service/internal/infrastructure/grpc/handlers"
)

func NewGRPCServer(authHandler *handlers.AuthServiceHandler, logger pkgLogger.Logger) *grpc.Server {
	// Create gRPC server with interceptors
	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			loggingInterceptor(logger),
			errorInterceptor(logger),
		),
	)

	// Register services
	authpb.RegisterAuthServiceServer(server, authHandler)

	// Enable reflection for debugging and tooling
	reflection.Register(server)

	logger.Info(context.Background(), "gRPC server created with reflection enabled")

	return server
}
```

**Testing with grpcurl**:
```bash
# List services
grpcurl -plaintext localhost:9091 list

# List methods for AuthService
grpcurl -plaintext localhost:9091 list auth.v1.AuthService

# Describe a method
grpcurl -plaintext localhost:9091 describe auth.v1.AuthService.ValidateToken

# Call a method
grpcurl -plaintext -d '{"token": "eyJhbGci..."}' localhost:9091 auth.v1.AuthService.ValidateToken
```

#### T005: Create Client Examples

**File**: `examples/grpc-clients/auth-client/main.go`

```go
package main

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	authpb "github.com/giia/giia-core-engine/services/auth-service/api/proto/auth/v1"
)

func main() {
	// Connect to Auth service
	conn, err := grpc.Dial(
		"localhost:9091",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithTimeout(5*time.Second),
	)
	if err != nil {
		log.Fatalf("Failed to connect to Auth service: %v", err)
	}
	defer conn.Close()

	log.Println("Connected to Auth service at localhost:9091")

	client := authpb.NewAuthServiceClient(conn)

	// Example 1: Validate Token
	log.Println("\n=== Example 1: Validate Token ===")
	validateTokenExample(client)

	// Example 2: Check Permission
	log.Println("\n=== Example 2: Check Permission ===")
	checkPermissionExample(client)

	// Example 3: Batch Check Permissions
	log.Println("\n=== Example 3: Batch Check Permissions ===")
	batchCheckPermissionsExample(client)

	// Example 4: Get User
	log.Println("\n=== Example 4: Get User ===")
	getUserExample(client)
}

func validateTokenExample(client authpb.AuthServiceClient) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.ValidateToken(ctx, &authpb.ValidateTokenRequest{
		Token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...", // Replace with real token
	})
	if err != nil {
		log.Printf("ValidateToken failed: %v", err)
		return
	}

	log.Printf("Token valid: %v", resp.Valid)
	if resp.Valid {
		log.Printf("User ID: %s", resp.UserId)
		log.Printf("Organization ID: %s", resp.OrganizationId)
	} else {
		log.Printf("Error: %s", resp.Error)
	}
}

func checkPermissionExample(client authpb.AuthServiceClient) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.CheckPermission(ctx, &authpb.CheckPermissionRequest{
		UserId:     "550e8400-e29b-41d4-a716-446655440000", // Replace with real user ID
		Permission: "catalog:write",
	})
	if err != nil {
		log.Printf("CheckPermission failed: %v", err)
		return
	}

	log.Printf("Permission 'catalog:write' allowed: %v", resp.Allowed)
	if !resp.Allowed {
		log.Printf("Reason: %s", resp.Reason)
	}
}

func batchCheckPermissionsExample(client authpb.AuthServiceClient) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.BatchCheckPermissions(ctx, &authpb.BatchCheckPermissionsRequest{
		UserId:      "550e8400-e29b-41d4-a716-446655440000",
		Permissions: []string{"catalog:read", "catalog:write", "buffers:read", "orders:approve"},
	})
	if err != nil {
		log.Printf("BatchCheckPermissions failed: %v", err)
		return
	}

	log.Println("Batch permission check results:")
	for permission, allowed := range resp.Results {
		log.Printf("  %s: %v", permission, allowed)
	}
}

func getUserExample(client authpb.AuthServiceClient) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.GetUser(ctx, &authpb.GetUserRequest{
		UserId: "550e8400-e29b-41d4-a716-446655440000",
	})
	if err != nil {
		log.Printf("GetUser failed: %v", err)
		return
	}

	log.Printf("User: %s (%s)", resp.User.Name, resp.User.Email)
	log.Printf("Status: %s", resp.User.Status)
	log.Printf("Roles: %v", resp.User.Roles)
}
```

**File**: `examples/grpc-clients/auth-client/README.md`

```markdown
# Auth Service gRPC Client Example

This example demonstrates how to call the Auth service gRPC API from Go code.

## Prerequisites

- Auth service running at `localhost:9091`
- Valid JWT token and user ID for testing

## Running the Example

```bash
# Install dependencies
go mod download

# Run the example
go run main.go
```

## Available Examples

1. **ValidateToken**: Validate a JWT token
2. **CheckPermission**: Check if a user has a specific permission
3. **BatchCheckPermissions**: Check multiple permissions at once
4. **GetUser**: Retrieve user information

## Integration in Your Service

To integrate Auth service calls in your own service:

```go
import (
    "google.golang.org/grpc"
    authpb "github.com/giia/giia-core-engine/services/auth-service/api/proto/auth/v1"
)

// Create client
conn, _ := grpc.Dial("auth-service:9091", grpc.WithInsecure())
client := authpb.NewAuthServiceClient(conn)

// Check permission
resp, _ := client.CheckPermission(ctx, &authpb.CheckPermissionRequest{
    UserId:     userID,
    Permission: "catalog:write",
})

if !resp.Allowed {
    return errors.New("permission denied")
}
```

## Error Handling

All RPC calls return standard gRPC errors. Use `status.Code()` to check error types:

```go
import "google.golang.org/grpc/status"

resp, err := client.ValidateToken(ctx, req)
if err != nil {
    st := status.Convert(err)
    switch st.Code() {
    case codes.Unauthenticated:
        // Handle authentication error
    case codes.PermissionDenied:
        // Handle permission error
    default:
        // Handle other errors
    }
}
```
```

#### T006: Catalog Service Proto Files

**File**: `services/catalog-service/api/proto/catalog/v1/catalog.proto`

```protobuf
syntax = "proto3";

package catalog.v1;

option go_package = "github.com/giia/giia-core-engine/services/catalog-service/api/proto/catalog/v1;catalogpb";

import "google/protobuf/timestamp.proto";

service CatalogService {
  // Product operations
  rpc GetProduct(GetProductRequest) returns (GetProductResponse);
  rpc ListProducts(ListProductsRequest) returns (ListProductsResponse);
  rpc SearchProducts(SearchProductsRequest) returns (SearchProductsResponse);

  // Supplier operations
  rpc GetSupplier(GetSupplierRequest) returns (GetSupplierResponse);
  rpc ListSuppliers(ListSuppliersRequest) returns (ListSuppliersResponse);

  // Buffer Profile operations
  rpc GetBufferProfile(GetBufferProfileRequest) returns (GetBufferProfileResponse);
  rpc ListBufferProfiles(ListBufferProfilesRequest) returns (ListBufferProfilesResponse);
}

// ===== Product Messages =====

message Product {
  string id = 1;
  string organization_id = 2;
  string sku = 3;
  string name = 4;
  string description = 5;
  string category = 6;
  string unit_of_measure = 7;
  double standard_cost = 8;
  string status = 9; // "active", "inactive", "discontinued"
  google.protobuf.Timestamp created_at = 10;
  google.protobuf.Timestamp updated_at = 11;
}

message GetProductRequest {
  string id = 1;
  string organization_id = 2;
}

message GetProductResponse {
  Product product = 1;
}

message ListProductsRequest {
  string organization_id = 1;
  string status = 2; // Filter by status (optional)
  int32 page = 3;
  int32 page_size = 4;
}

message ListProductsResponse {
  repeated Product products = 1;
  int32 total = 2;
  int32 page = 3;
  int32 page_size = 4;
}

message SearchProductsRequest {
  string organization_id = 1;
  string query = 2; // Search by SKU or name
  string category = 3; // Filter by category (optional)
}

message SearchProductsResponse {
  repeated Product products = 1;
}

// ===== Supplier Messages =====

message Supplier {
  string id = 1;
  string organization_id = 2;
  string code = 3;
  string name = 4;
  string contact_name = 5;
  string contact_email = 6;
  string contact_phone = 7;
  int32 default_lead_time = 8; // Days
  string status = 9; // "active", "inactive"
  google.protobuf.Timestamp created_at = 10;
  google.protobuf.Timestamp updated_at = 11;
}

message GetSupplierRequest {
  string id = 1;
  string organization_id = 2;
}

message GetSupplierResponse {
  Supplier supplier = 1;
}

message ListSuppliersRequest {
  string organization_id = 1;
  string status = 2; // Filter by status (optional)
}

message ListSuppliersResponse {
  repeated Supplier suppliers = 1;
}

// ===== Buffer Profile Messages =====

message BufferProfile {
  string id = 1;
  string organization_id = 2;
  string name = 3;
  string adu_method = 4; // "average", "exponential", "weighted"
  double lead_time_factor = 5;
  double variability_factor = 6;
  string status = 7; // "active", "inactive"
  google.protobuf.Timestamp created_at = 8;
  google.protobuf.Timestamp updated_at = 9;
}

message GetBufferProfileRequest {
  string id = 1;
  string organization_id = 2;
}

message GetBufferProfileResponse {
  BufferProfile buffer_profile = 1;
}

message ListBufferProfilesRequest {
  string organization_id = 1;
}

message ListBufferProfilesResponse {
  repeated BufferProfile buffer_profiles = 1;
}
```

**Generate Proto Files**:

**File**: `scripts/generate-proto-catalog.sh`

```bash
#!/bin/bash
set -e

PROTO_DIR="services/catalog-service/api/proto"
OUTPUT_DIR="services/catalog-service/api/proto"

echo "Generating Catalog service proto files..."

protoc \
  --go_out=$OUTPUT_DIR \
  --go_opt=paths=source_relative \
  --go-grpc_out=$OUTPUT_DIR \
  --go-grpc_opt=paths=source_relative \
  -I=$PROTO_DIR \
  $PROTO_DIR/catalog/v1/catalog.proto

echo "Proto generation complete!"
echo "Generated files:"
echo "  - $OUTPUT_DIR/catalog/v1/catalog.pb.go"
echo "  - $OUTPUT_DIR/catalog/v1/catalog_grpc.pb.go"
```

```bash
# Run generation
bash scripts/generate-proto-catalog.sh
```

---

### Phase 3: NATS Event System Polish (Day 3)

#### T007: Publish Additional Auth Events

**File**: `services/auth-service/internal/core/usecases/auth/register_user.go` (modify existing)

```go
// After successful registration
if err := uc.eventPublisher.Publish(ctx, &pkgEvents.Event{
	Type:    "user.registered",
	Subject: fmt.Sprintf("auth.users.%s", user.ID.String()),
	Data: map[string]interface{}{
		"user_id":         user.ID.String(),
		"email":           user.Email,
		"name":            user.Name,
		"organization_id": user.OrganizationID.String(),
		"status":          string(user.Status),
		"registered_at":   uc.timeManager.Now().Format(time.RFC3339),
	},
}); err != nil {
	uc.logger.Error(ctx, err, "Failed to publish user.registered event")
}
```

**File**: `services/auth-service/internal/core/usecases/auth/verify_email.go` (modify existing)

```go
// After email verification
if err := uc.eventPublisher.Publish(ctx, &pkgEvents.Event{
	Type:    "user.verified",
	Subject: fmt.Sprintf("auth.users.%s", user.ID.String()),
	Data: map[string]interface{}{
		"user_id":     user.ID.String(),
		"email":       user.Email,
		"verified_at": uc.timeManager.Now().Format(time.RFC3339),
	},
}); err != nil {
	uc.logger.Error(ctx, err, "Failed to publish user.verified event")
}
```

**File**: `services/catalog-service/internal/core/usecases/supplier/create_supplier.go` (modify existing)

```go
// After supplier creation
if err := uc.eventPublisher.Publish(ctx, &pkgEvents.Event{
	Type:    "supplier.created",
	Subject: fmt.Sprintf("catalog.suppliers.%s", supplier.ID.String()),
	Data: map[string]interface{}{
		"supplier_id":     supplier.ID.String(),
		"code":            supplier.Code,
		"name":            supplier.Name,
		"organization_id": supplier.OrganizationID.String(),
		"created_at":      uc.timeManager.Now().Format(time.RFC3339),
	},
}); err != nil {
	uc.logger.Error(ctx, err, "Failed to publish supplier.created event")
}
```

#### T008: Implement DLQ Consumer

**File**: `pkg/events/dlq_consumer.go`

```go
package events

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/giia/giia-core-engine/pkg/logger"
)

type DLQConsumer struct {
	subscriber Subscriber
	logger     logger.Logger
}

func NewDLQConsumer(subscriber Subscriber, logger logger.Logger) *DLQConsumer {
	return &DLQConsumer{
		subscriber: subscriber,
		logger:     logger,
	}
}

func (c *DLQConsumer) Start(ctx context.Context) error {
	c.logger.Info(ctx, "Starting DLQ consumer...")

	return c.subscriber.Subscribe(ctx, "DLQ_EVENTS", "dlq.>", func(event *Event) error {
		// Log failed event
		c.logger.Error(ctx, nil, "Event failed and sent to DLQ", logger.Tags{
			"event_type":    event.Type,
			"event_subject": event.Subject,
			"event_id":      event.ID,
		})

		// Check retry count from metadata
		retryCount := getRetryCount(event)

		if retryCount < 3 {
			// Retry with exponential backoff
			backoffDuration := time.Duration(math.Pow(2, float64(retryCount))) * time.Second
			c.logger.Info(ctx, fmt.Sprintf("Retrying event after %v", backoffDuration), logger.Tags{
				"event_type":  event.Type,
				"retry_count": retryCount,
			})

			time.Sleep(backoffDuration)

			// Increment retry count and retry
			event.Data["retry_count"] = retryCount + 1
			return c.retryEvent(ctx, event)
		}

		// Max retries exceeded - alert
		c.alertCriticalFailure(event)

		return nil
	})
}

func (c *DLQConsumer) retryEvent(ctx context.Context, event *Event) error {
	// Re-publish to original subject
	originalSubject := event.Subject
	if dlqPrefix := "dlq."; len(originalSubject) > len(dlqPrefix) {
		originalSubject = originalSubject[len(dlqPrefix):]
	}

	// This would require a publisher instance
	c.logger.Info(ctx, "Retrying event on original subject", logger.Tags{
		"original_subject": originalSubject,
		"event_type":       event.Type,
	})

	// TODO: Implement retry logic with publisher
	return nil
}

func (c *DLQConsumer) alertCriticalFailure(event *Event) {
	// Alert via logging (could integrate with monitoring/alerting system)
	c.logger.Error(context.Background(), nil, "CRITICAL: Event failed after max retries", logger.Tags{
		"event_type":    event.Type,
		"event_subject": event.Subject,
		"event_id":      event.ID,
		"retry_count":   getRetryCount(event),
	})

	// TODO: Integrate with alerting system (PagerDuty, Slack, etc.)
}

func getRetryCount(event *Event) int {
	if retryCount, ok := event.Data["retry_count"].(int); ok {
		return retryCount
	}
	if retryCount, ok := event.Data["retry_count"].(float64); ok {
		return int(retryCount)
	}
	return 0
}
```

#### T009: Event Schema Validation

**File**: `pkg/events/schema_validator.go`

```go
package events

import (
	"fmt"
)

type EventSchema struct {
	Type           string
	RequiredFields []string
	OptionalFields []string
}

var schemas = map[string]EventSchema{
	"user.registered": {
		Type:           "user.registered",
		RequiredFields: []string{"user_id", "email", "organization_id", "status"},
		OptionalFields: []string{"name", "registered_at"},
	},
	"user.verified": {
		Type:           "user.verified",
		RequiredFields: []string{"user_id", "email", "verified_at"},
	},
	"user.password_changed": {
		Type:           "user.password_changed",
		RequiredFields: []string{"user_id", "changed_at"},
	},
	"product.created": {
		Type:           "product.created",
		RequiredFields: []string{"product_id", "sku", "name", "organization_id"},
	},
	"supplier.created": {
		Type:           "supplier.created",
		RequiredFields: []string{"supplier_id", "code", "name", "organization_id"},
	},
	"buffer_profile.created": {
		Type:           "buffer_profile.created",
		RequiredFields: []string{"buffer_profile_id", "name", "organization_id"},
	},
}

func ValidateEvent(event *Event) error {
	schema, exists := schemas[event.Type]
	if !exists {
		return fmt.Errorf("unknown event type: %s", event.Type)
	}

	// Check required fields
	for _, field := range schema.RequiredFields {
		if _, ok := event.Data[field]; !ok {
			return fmt.Errorf("missing required field '%s' for event type '%s'", field, event.Type)
		}
	}

	return nil
}

func RegisterSchema(eventType string, requiredFields []string, optionalFields []string) {
	schemas[eventType] = EventSchema{
		Type:           eventType,
		RequiredFields: requiredFields,
		OptionalFields: optionalFields,
	}
}
```

---

### Phase 4: Production Readiness (Day 4)

#### T010: Health Check Endpoints

**File**: `pkg/health/checker.go`

```go
package health

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/nats-io/nats.go"
	"gorm.io/gorm"
)

type HealthChecker struct {
	db    *gorm.DB
	nats  *nats.Conn
	redis *redis.Client
}

type HealthStatus struct {
	Status string            `json:"status"`
	Checks map[string]string `json:"checks"`
}

func NewHealthChecker(db *gorm.DB, natsConn *nats.Conn, redisClient *redis.Client) *HealthChecker {
	return &HealthChecker{
		db:    db,
		nats:  natsConn,
		redis: redisClient,
	}
}

func (h *HealthChecker) CheckReadiness(ctx context.Context) HealthStatus {
	checks := make(map[string]string)

	// Check database
	checks["database"] = h.checkDatabase(ctx)

	// Check NATS
	checks["nats"] = h.checkNATS()

	// Check Redis
	checks["redis"] = h.checkRedis(ctx)

	// Determine overall status
	status := "healthy"
	for _, checkStatus := range checks {
		if checkStatus != "healthy" {
			status = "unhealthy"
			break
		}
	}

	return HealthStatus{
		Status: status,
		Checks: checks,
	}
}

func (h *HealthChecker) checkDatabase(ctx context.Context) string {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := h.db.WithContext(ctx).Exec("SELECT 1").Error; err != nil {
		return "unhealthy"
	}
	return "healthy"
}

func (h *HealthChecker) checkNATS() string {
	if h.nats == nil || !h.nats.IsConnected() {
		return "unhealthy"
	}
	return "healthy"
}

func (h *HealthChecker) checkRedis(ctx context.Context) string {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := h.redis.Ping(ctx).Err(); err != nil {
		return "unhealthy"
	}
	return "healthy"
}
```

**File**: `services/auth-service/internal/infrastructure/entrypoints/http/health_handlers.go`

```go
package http

import (
	"encoding/json"
	"net/http"

	"github.com/giia/giia-core-engine/pkg/health"
	pkgLogger "github.com/giia/giia-core-engine/pkg/logger"
)

type HealthHandlers struct {
	healthChecker *health.HealthChecker
	logger        pkgLogger.Logger
}

func NewHealthHandlers(healthChecker *health.HealthChecker, logger pkgLogger.Logger) *HealthHandlers {
	return &HealthHandlers{
		healthChecker: healthChecker,
		logger:        logger,
	}
}

// Liveness probe - always returns 200 if service is running
func (h *HealthHandlers) Liveness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "alive",
	})
}

// Readiness probe - checks dependencies
func (h *HealthHandlers) Readiness(w http.ResponseWriter, r *http.Request) {
	healthStatus := h.healthChecker.CheckReadiness(r.Context())

	statusCode := http.StatusOK
	if healthStatus.Status != "healthy" {
		statusCode = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(healthStatus)
}
```

#### T011: Graceful Shutdown

**File**: `services/auth-service/cmd/server/main.go` (modify existing)

```go
func main() {
	// ... existing setup code ...

	// Create HTTP server
	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.HTTPPort),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Create gRPC server
	grpcServer := setupGRPCServer(useCases, logger)
	grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.GRPCPort))
	if err != nil {
		logger.Error(context.Background(), err, "Failed to create gRPC listener")
		os.Exit(1)
	}

	// Setup shutdown signal handling
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start HTTP server
	go func() {
		logger.Info(context.Background(), fmt.Sprintf("Starting HTTP server on port %s", cfg.HTTPPort))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error(context.Background(), err, "HTTP server error")
		}
	}()

	// Start gRPC server
	go func() {
		logger.Info(context.Background(), fmt.Sprintf("Starting gRPC server on port %s", cfg.GRPCPort))
		if err := grpcServer.Serve(grpcListener); err != nil {
			logger.Error(context.Background(), err, "gRPC server error")
		}
	}()

	// Wait for shutdown signal
	<-quit
	logger.Info(context.Background(), "Received shutdown signal, initiating graceful shutdown...")

	// Create shutdown context with 30-second timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Shutdown HTTP server
	logger.Info(shutdownCtx, "Shutting down HTTP server...")
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error(shutdownCtx, err, "HTTP server forced to shutdown")
	} else {
		logger.Info(shutdownCtx, "HTTP server shutdown complete")
	}

	// Shutdown gRPC server
	logger.Info(shutdownCtx, "Shutting down gRPC server...")
	stopped := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(stopped)
	}()

	// Wait for graceful stop or timeout
	select {
	case <-stopped:
		logger.Info(shutdownCtx, "gRPC server shutdown complete")
	case <-shutdownCtx.Done():
		logger.Error(shutdownCtx, nil, "gRPC server shutdown timeout, forcing stop")
		grpcServer.Stop()
	}

	// Close database connection
	logger.Info(context.Background(), "Closing database connection...")
	sqlDB, _ := db.DB()
	if err := sqlDB.Close(); err != nil {
		logger.Error(context.Background(), err, "Failed to close database")
	}

	// Close NATS connection
	logger.Info(context.Background(), "Closing NATS connection...")
	natsConn.Close()

	// Close Redis connection
	logger.Info(context.Background(), "Closing Redis connection...")
	if err := redisClient.Close(); err != nil {
		logger.Error(context.Background(), err, "Failed to close Redis")
	}

	logger.Info(context.Background(), "Server exited successfully")
}
```

#### T012: Environment Configs

**File**: `deploy/environments/.env.dev`

```bash
# Development Environment Configuration

# Service
SERVICE_NAME=auth-service
ENVIRONMENT=development
LOG_LEVEL=debug

# HTTP Server
HTTP_PORT=8083
HTTP_HOST=0.0.0.0

# gRPC Server
GRPC_PORT=9091
GRPC_HOST=0.0.0.0

# Database
DATABASE_URL=postgresql://giia_user:giia_password@localhost:5432/giia_auth_dev?sslmode=disable
DATABASE_MAX_OPEN_CONNS=25
DATABASE_MAX_IDLE_CONNS=5
DATABASE_CONN_MAX_LIFETIME=300s

# Redis
REDIS_URL=redis://localhost:6379/0
REDIS_PASSWORD=
REDIS_DB=0
REDIS_POOL_SIZE=10

# NATS
NATS_URL=nats://localhost:4222
NATS_CLUSTER_ID=giia-cluster
NATS_CLIENT_ID=auth-service-dev

# JWT
JWT_SECRET=dev-secret-key-change-in-production
JWT_EXPIRATION=24h
JWT_REFRESH_EXPIRATION=168h

# CORS
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:8080
CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,PATCH,OPTIONS
CORS_ALLOWED_HEADERS=Content-Type,Authorization

# Rate Limiting
RATE_LIMIT_ENABLED=false
RATE_LIMIT_REQUESTS_PER_MINUTE=1000

# Email (Development - Console)
EMAIL_PROVIDER=console
EMAIL_FROM=noreply@giia-dev.local

# Feature Flags
ENABLE_AUDIT_LOGGING=true
ENABLE_METRICS=true
ENABLE_GRPC_REFLECTION=true
```

**File**: `deploy/environments/.env.staging`

```bash
# Staging Environment Configuration

# Service
SERVICE_NAME=auth-service
ENVIRONMENT=staging
LOG_LEVEL=info

# HTTP Server
HTTP_PORT=8083
HTTP_HOST=0.0.0.0

# gRPC Server
GRPC_PORT=9091
GRPC_HOST=0.0.0.0

# Database (Use Kubernetes secrets in production)
DATABASE_URL=${DATABASE_URL}
DATABASE_MAX_OPEN_CONNS=50
DATABASE_MAX_IDLE_CONNS=10
DATABASE_CONN_MAX_LIFETIME=600s

# Redis
REDIS_URL=${REDIS_URL}
REDIS_PASSWORD=${REDIS_PASSWORD}
REDIS_DB=0
REDIS_POOL_SIZE=20

# NATS
NATS_URL=${NATS_URL}
NATS_CLUSTER_ID=giia-cluster-staging
NATS_CLIENT_ID=auth-service-staging

# JWT
JWT_SECRET=${JWT_SECRET}
JWT_EXPIRATION=24h
JWT_REFRESH_EXPIRATION=168h

# CORS
CORS_ALLOWED_ORIGINS=https://staging.giia.com
CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,PATCH,OPTIONS
CORS_ALLOWED_HEADERS=Content-Type,Authorization

# Rate Limiting
RATE_LIMIT_ENABLED=true
RATE_LIMIT_REQUESTS_PER_MINUTE=100

# Email
EMAIL_PROVIDER=smtp
EMAIL_SMTP_HOST=${EMAIL_SMTP_HOST}
EMAIL_SMTP_PORT=587
EMAIL_SMTP_USERNAME=${EMAIL_SMTP_USERNAME}
EMAIL_SMTP_PASSWORD=${EMAIL_SMTP_PASSWORD}
EMAIL_FROM=noreply@giia-staging.com

# Feature Flags
ENABLE_AUDIT_LOGGING=true
ENABLE_METRICS=true
ENABLE_GRPC_REFLECTION=true
```

**File**: `deploy/environments/.env.prod`

```bash
# Production Environment Configuration

# Service
SERVICE_NAME=auth-service
ENVIRONMENT=production
LOG_LEVEL=warn

# HTTP Server
HTTP_PORT=8083
HTTP_HOST=0.0.0.0

# gRPC Server
GRPC_PORT=9091
GRPC_HOST=0.0.0.0

# Database (MUST use Kubernetes secrets)
DATABASE_URL=${DATABASE_URL}
DATABASE_MAX_OPEN_CONNS=100
DATABASE_MAX_IDLE_CONNS=25
DATABASE_CONN_MAX_LIFETIME=1800s

# Redis
REDIS_URL=${REDIS_URL}
REDIS_PASSWORD=${REDIS_PASSWORD}
REDIS_DB=0
REDIS_POOL_SIZE=50

# NATS
NATS_URL=${NATS_URL}
NATS_CLUSTER_ID=giia-cluster-production
NATS_CLIENT_ID=auth-service-prod

# JWT
JWT_SECRET=${JWT_SECRET}
JWT_EXPIRATION=12h
JWT_REFRESH_EXPIRATION=72h

# CORS
CORS_ALLOWED_ORIGINS=https://app.giia.com,https://www.giia.com
CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,PATCH,OPTIONS
CORS_ALLOWED_HEADERS=Content-Type,Authorization

# Rate Limiting
RATE_LIMIT_ENABLED=true
RATE_LIMIT_REQUESTS_PER_MINUTE=60

# Email
EMAIL_PROVIDER=smtp
EMAIL_SMTP_HOST=${EMAIL_SMTP_HOST}
EMAIL_SMTP_PORT=587
EMAIL_SMTP_USERNAME=${EMAIL_SMTP_USERNAME}
EMAIL_SMTP_PASSWORD=${EMAIL_SMTP_PASSWORD}
EMAIL_FROM=noreply@giia.com

# Feature Flags
ENABLE_AUDIT_LOGGING=true
ENABLE_METRICS=true
ENABLE_GRPC_REFLECTION=false
```

---

### Phase 5: Documentation (Day 5)

#### T013: Architecture Diagrams

**File**: `docs/architecture/sequence-diagrams.md`

```markdown
# Sequence Diagrams

## 1. User Registration Flow

```mermaid
sequenceDiagram
    participant Frontend
    participant AuthService
    participant Database
    participant EmailService
    participant NATS

    Frontend->>AuthService: POST /api/auth/register
    AuthService->>AuthService: Validate input
    AuthService->>Database: Check email exists
    Database-->>AuthService: Not found
    AuthService->>Database: Create user (status=pending)
    Database-->>AuthService: User created
    AuthService->>AuthService: Generate verification token
    AuthService->>Database: Save token
    AuthService->>EmailService: Send verification email
    AuthService->>NATS: Publish user.registered event
    AuthService-->>Frontend: 201 Created
    Frontend-->>User: Check email for verification link
```

## 2. Token Validation Flow (gRPC)

```mermaid
sequenceDiagram
    participant CatalogService
    participant AuthService
    participant Redis
    participant Database

    CatalogService->>AuthService: gRPC ValidateToken(token)
    AuthService->>AuthService: Parse JWT
    alt Token in cache
        AuthService->>Redis: Get cached validation
        Redis-->>AuthService: Valid/Invalid
    else Token not cached
        AuthService->>Database: Verify user exists + active
        Database-->>AuthService: User data
        AuthService->>Redis: Cache validation (5min TTL)
    end
    AuthService-->>CatalogService: ValidationResponse
```

## 3. Permission Check Flow

```mermaid
sequenceDiagram
    participant Frontend
    participant CatalogService
    participant AuthService
    participant Redis
    participant Database
    participant AuditLog

    Frontend->>CatalogService: POST /api/products (with JWT)
    CatalogService->>AuthService: gRPC CheckPermission(userID, "catalog:write")
    AuthService->>Redis: Get user permissions (cache)
    alt Cache hit
        Redis-->>AuthService: Cached permissions
    else Cache miss
        AuthService->>Database: Get user roles + permissions
        Database-->>AuthService: Permission list
        AuthService->>Redis: Cache permissions (5min TTL)
    end
    AuthService->>AuditLog: Log permission check (async)
    AuthService-->>CatalogService: CheckPermissionResponse(allowed=true)
    CatalogService->>Database: Create product
    CatalogService->>NATS: Publish product.created event
    CatalogService-->>Frontend: 201 Created
```

## 4. Buffer Calculation Flow

```mermaid
sequenceDiagram
    participant Frontend
    participant DDMRPService
    participant CatalogService
    participant Database
    participant NATS

    Frontend->>DDMRPService: POST /api/buffers/calculate
    DDMRPService->>CatalogService: gRPC GetProduct(productID)
    CatalogService-->>DDMRPService: Product data
    DDMRPService->>CatalogService: gRPC GetBufferProfile(profileID)
    CatalogService-->>DDMRPService: Buffer profile
    DDMRPService->>DDMRPService: Calculate ADU
    DDMRPService->>DDMRPService: Calculate buffer zones
    DDMRPService->>Database: Save buffer calculation
    DDMRPService->>NATS: Publish buffer.calculated event
    DDMRPService-->>Frontend: 200 OK (buffer zones)
```
```

#### T014: API Documentation

**File**: `docs/api/openapi/auth-service.yaml` (OpenAPI 3.0 spec - abbreviated)

```yaml
openapi: 3.0.3
info:
  title: Auth Service API
  description: Authentication and authorization service for GIIA
  version: 1.0.0

servers:
  - url: http://localhost:8083/api
    description: Development
  - url: https://staging-api.giia.com/api
    description: Staging
  - url: https://api.giia.com/api
    description: Production

paths:
  /auth/register:
    post:
      summary: Register new user
      tags: [Authentication]
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required: [email, password, name]
              properties:
                email:
                  type: string
                  format: email
                password:
                  type: string
                  minLength: 8
                name:
                  type: string
      responses:
        '201':
          description: User registered successfully
        '400':
          description: Invalid input
        '409':
          description: Email already exists

  /auth/login:
    post:
      summary: Login user
      tags: [Authentication]
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required: [email, password]
              properties:
                email:
                  type: string
                password:
                  type: string
      responses:
        '200':
          description: Login successful
          content:
            application/json:
              schema:
                type: object
                properties:
                  access_token:
                    type: string
                  refresh_token:
                    type: string
                  expires_in:
                    type: integer

  /health/live:
    get:
      summary: Liveness probe
      tags: [Health]
      responses:
        '200':
          description: Service is alive

  /health/ready:
    get:
      summary: Readiness probe
      tags: [Health]
      responses:
        '200':
          description: Service is ready
        '503':
          description: Service is not ready
```

#### T015: Runbooks

**File**: `docs/runbooks/deployment.md`

```markdown
# Deployment Runbook

## Pre-Deployment Checklist

- [ ] All tests passing (`make test`)
- [ ] Linters passing (`make lint`)
- [ ] Database migrations prepared
- [ ] Configuration reviewed for target environment
- [ ] Secrets updated in Kubernetes
- [ ] Backup taken of production database

## Deployment Steps

### 1. Deploy to Staging

```bash
# Tag release
git tag -a v1.2.0 -m "Release v1.2.0"
git push origin v1.2.0

# Deploy to staging
kubectl config use-context staging
kubectl apply -f k8s/auth-service/
kubectl rollout status deployment/auth-service

# Verify health
kubectl get pods -l app=auth-service
kubectl logs -l app=auth-service --tail=50

# Run smoke tests
./scripts/smoke-tests-staging.sh
```

### 2. Run Database Migrations

```bash
# Connect to staging database
kubectl port-forward svc/postgres 5432:5432

# Run migrations
cd services/auth-service
go run cmd/migrate/main.go up

# Verify migration
psql $DATABASE_URL -c "\dt"
```

### 3. Deploy to Production

```bash
# Switch to production context
kubectl config use-context production

# Deploy with rolling update
kubectl apply -f k8s/auth-service/
kubectl rollout status deployment/auth-service

# Monitor deployment
watch kubectl get pods -l app=auth-service

# Verify health checks
curl https://api.giia.com/health/ready

# Check logs for errors
kubectl logs -l app=auth-service --tail=100 | grep ERROR
```

### 4. Post-Deployment Verification

```bash
# Run smoke tests
./scripts/smoke-tests-production.sh

# Check key metrics
# - Response times <200ms p95
# - Error rate <0.1%
# - All pods healthy

# Verify gRPC endpoints
grpcurl api.giia.com:9091 list

# Monitor for 30 minutes
# - Check Grafana dashboards
# - Monitor error logs in Kibana
# - Verify no increase in error rate
```

## Rollback Procedure

If issues detected:

```bash
# Rollback deployment
kubectl rollout undo deployment/auth-service

# Verify rollback
kubectl rollout status deployment/auth-service

# Check logs
kubectl logs -l app=auth-service --tail=100
```
```

**File**: `docs/runbooks/troubleshooting.md`

```markdown
# Troubleshooting Guide

## Common Issues

### 1. Health Check Failing

**Symptom**: `/health/ready` returns 503

**Diagnosis**:
```bash
# Check pod status
kubectl get pods -l app=auth-service

# Check logs
kubectl logs <pod-name> --tail=100

# Check dependencies
kubectl exec <pod-name> -- psql $DATABASE_URL -c "SELECT 1"
kubectl exec <pod-name> -- redis-cli -u $REDIS_URL ping
```

**Solutions**:
- Database connection: Check `DATABASE_URL` secret
- Redis connection: Check `REDIS_URL` secret
- NATS connection: Verify NATS service is running

---

### 2. Permission Checks Slow

**Symptom**: Permission checks >50ms

**Diagnosis**:
```bash
# Check Redis cache hit rate
redis-cli INFO stats | grep keyspace

# Run performance benchmark
cd services/auth-service
go test -bench=BenchmarkCheckPermission -benchtime=10s
```

**Solutions**:
- Increase Redis connection pool: `REDIS_POOL_SIZE=50`
- Check Redis memory: `redis-cli INFO memory`
- Verify cache TTL is appropriate (5 minutes default)

---

### 3. gRPC Connection Errors

**Symptom**: Catalog service cannot reach Auth service

**Diagnosis**:
```bash
# Test gRPC connection
grpcurl -plaintext auth-service:9091 list

# Check service discovery
kubectl get svc auth-service
nslookup auth-service.default.svc.cluster.local
```

**Solutions**:
- Verify gRPC port in service: `kubectl get svc auth-service -o yaml`
- Check firewall rules
- Verify gRPC reflection enabled

---

### 4. Event Publishing Failures

**Symptom**: Events not appearing in NATS streams

**Diagnosis**:
```bash
# Check NATS connection
nats stream ls
nats stream info AUTH_EVENTS

# Check DLQ for failed events
nats stream info DLQ_EVENTS
```

**Solutions**:
- Verify NATS connection: Check `NATS_URL` config
- Check stream configuration: `nats stream ls`
- Review DLQ consumer logs: `kubectl logs dlq-consumer`

---

## Emergency Contacts

- **On-Call Engineer**: [PagerDuty]
- **Database Admin**: [Contact]
- **DevOps Team**: [Slack Channel]
```

---

## 4. Testing Strategy

### Unit Tests

Each new file/function requires unit tests:

```bash
# Run all tests
go test ./... -v -count=1

# Test coverage
go test ./services/auth-service/... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

**Target**: 80%+ coverage for all new code

### Integration Tests

Test inter-service communication:

```bash
# Test gRPC clients
go test ./examples/grpc-clients/... -v

# Test event publishing/subscribing
go test ./pkg/events/... -v -tags=integration
```

### Performance Tests

```bash
# Run benchmarks
go test ./services/auth-service/test/performance/... -bench=. -benchtime=10s

# Verify targets:
# - Permission checks: <10ms p95 (cached)
# - Permission checks: <50ms p95 (uncached)
# - Health checks: <100ms p95
```

### End-to-End Tests

```bash
# Run smoke tests (staging)
./scripts/smoke-tests-staging.sh

# Includes:
# - User registration flow
# - Token validation
# - Permission checks
# - gRPC service calls
# - Health check endpoints
```

---

## 5. Dependencies and Execution Order

### Dependency Graph

```
T001 (Seed Data) ─────────┐
T002 (Audit Logging) ─────┤
T003 (Performance Tests) ─┴─> Phase 1 Complete
                              │
T004 (gRPC Reflection) ───────┐
T005 (Client Examples) ───────┤
T006 (Catalog Proto) ─────────┴─> Phase 2 Complete
                                  │
T007 (Additional Events) ─────────┐
T008 (DLQ Consumer) ──────────────┤
T009 (Event Validation) ──────────┴─> Phase 3 Complete
                                      │
T010 (Health Checks) ─────────────────┐
T011 (Graceful Shutdown) ─────────────┤
T012 (Environment Configs) ───────────┴─> Phase 4 Complete
                                          │
T013 (Architecture Diagrams) ─────────────┐
T014 (API Documentation) ─────────────────┤
T015 (Runbooks) ──────────────────────────┴─> Phase 5 Complete (100%)
```

### Parallel Execution Opportunities

- **Phase 1**: T001, T002 can run in parallel; T003 depends on T002
- **Phase 2**: T004, T005, T006 can run in parallel
- **Phase 3**: T007, T008, T009 can run in parallel
- **Phase 4**: T010, T011, T012 can run in parallel
- **Phase 5**: T013, T014, T015 can run in parallel

---

## 6. Acceptance Checklist

### RBAC and Audit Logging
- [ ] Seed data script creates 4 roles with correct permissions
- [ ] Seed script is idempotent (can run multiple times)
- [ ] Audit logs saved to database for permission checks
- [ ] Audit logs include request metadata (IP, user agent, request ID)
- [ ] Performance benchmarks show <10ms p95 for cached permission checks

### gRPC Reflection and Examples
- [ ] gRPC reflection enabled on auth-service
- [ ] gRPC reflection enabled on catalog-service
- [ ] grpcurl can list and describe services
- [ ] Auth client example works (token validation, permission check)
- [ ] Catalog client example works (get product, list suppliers)
- [ ] Proto files generated for catalog service

### NATS Event System
- [ ] user.registered event published on registration
- [ ] user.verified event published on email verification
- [ ] supplier.created event published on supplier creation
- [ ] buffer_profile.created event published
- [ ] DLQ consumer processes failed events
- [ ] DLQ consumer retries with exponential backoff
- [ ] DLQ consumer alerts after max retries
- [ ] Event schema validation catches invalid events

### Production Readiness
- [ ] /health/live endpoint returns 200
- [ ] /health/ready checks database, NATS, Redis
- [ ] Graceful shutdown waits for in-flight requests (30s timeout)
- [ ] Configuration validated on startup
- [ ] .env.dev, .env.staging, .env.prod configs created
- [ ] Secrets documented in README
- [ ] Kubernetes deployment manifests updated with health probes

### Documentation
- [ ] Architecture sequence diagrams created (4 flows)
- [ ] OpenAPI specs generated for auth and catalog REST APIs
- [ ] Proto documentation generated
- [ ] Deployment runbook complete
- [ ] Troubleshooting runbook complete with 4+ scenarios
- [ ] Scaling runbook complete

### Final Verification
- [ ] All unit tests pass (`go test ./...`)
- [ ] All integration tests pass
- [ ] Performance benchmarks meet targets
- [ ] All linters pass (`golangci-lint run`)
- [ ] Pre-commit hooks pass
- [ ] Code reviewed and approved
- [ ] **Phase 1 at 100% completion ✅**

---

## 7. Risk Mitigation

| Risk | Impact | Probability | Mitigation Strategy |
|------|--------|-------------|---------------------|
| Seed data conflicts with existing data | Medium | Medium | Use `ON CONFLICT DO NOTHING` in SQL, test in staging first |
| Performance targets not met (<10ms p95) | Medium | Low | Implement caching early, run benchmarks frequently |
| DLQ consumer complexity underestimated | Medium | Medium | Start with simple retry logic, iterate based on production patterns |
| Documentation becomes stale quickly | Low | High | Automate OpenAPI/proto doc generation, make docs part of PR reviews |
| Graceful shutdown not tested thoroughly | High | Low | Add integration tests for shutdown scenarios, test in staging |
| Missing events in production | High | Low | Use event schema validation, monitor DLQ, alert on anomalies |

---

## 8. Definition of Done

- All 15 tasks (T001-T015) completed
- All acceptance checklist items checked
- All tests passing (unit, integration, performance)
- Code reviewed and approved by 2+ engineers
- Documentation complete and reviewed
- Deployed to staging and smoke tested
- Performance validated in staging environment
- **Phase 1 completion verified at 100%**

---

**Document Version**: 1.0
**Last Updated**: 2025-12-16
**Status**: Ready for Implementation
**Estimated Completion**: 5 days (1 week)
**Next Step**: Begin Phase 1 - RBAC Seed Data and Audit Logging
