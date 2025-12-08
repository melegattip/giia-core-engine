# Integración: Users Service → Auth/IAM Service GIIA

## Resumen Ejecutivo

Tu `users-service` es **excelente** y lo convertiremos en el **6to servicio** de la arquitectura balanceada: **Auth/IAM Service**.

**Cambios necesarios**: Agregar multi-tenancy + RBAC + gRPC + Events
**Tiempo estimado**: 2-3 semanas
**Complejidad**: Media (mayoría son adiciones, no refactorización)

---

## 1. Arquitectura Actual vs. Target

### 1.1 Estado Actual (users-service)

```
users-service/
├── Domain: User, Preferences, NotificationSettings, TwoFA
├── Auth: JWT (access, refresh, email verification, password reset)
├── Security: bcrypt, account locking, 2FA/TOTP
├── API: REST (Gin framework)
├── Database: PostgreSQL
└── Missing: Multi-tenancy, RBAC, gRPC, Events
```

### 1.2 Target (auth-service para GIIA)

```
auth-service/
├── Domain: User, Tenant, Role, Permission + (existing entities)
├── Auth: (keep existing) + multi-tenant context
├── Security: (keep existing) + RBAC
├── APIs: REST (Gin) + gRPC + WebSocket (optional)
├── Database: PostgreSQL (extended schema)
└── Events: NATS (user.created, user.logged_in, tenant.created)
```

---

## 2. Cambios al Schema de Base de Datos

### 2.1 Agregar Multi-Tenancy

```sql
-- =====================================================
-- NUEVO: TABLA DE TENANTS (Organizaciones)
-- =====================================================

CREATE TABLE IF NOT EXISTS tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL, -- company-xyz
    subscription_plan VARCHAR(50) DEFAULT 'FREE', -- FREE, STARTER, PRO, ENTERPRISE
    subscription_status VARCHAR(20) DEFAULT 'ACTIVE', -- ACTIVE, SUSPENDED, CANCELLED
    max_users INT DEFAULT 5,
    settings JSONB, -- Custom settings per tenant
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_tenants_slug ON tenants(slug);
CREATE INDEX idx_tenants_status ON tenants(subscription_status);

-- =====================================================
-- MODIFICAR: TABLA USERS - Agregar tenant_id
-- =====================================================

ALTER TABLE users ADD COLUMN tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE;

-- Migrar datos existentes: crear tenant "default" para users actuales
INSERT INTO tenants (id, name, slug) VALUES ('00000000-0000-0000-0000-000000000000', 'Default Tenant', 'default');
UPDATE users SET tenant_id = '00000000-0000-0000-0000-000000000000' WHERE tenant_id IS NULL;

-- Hacer tenant_id requerido
ALTER TABLE users ALTER COLUMN tenant_id SET NOT NULL;

-- Cambiar unique constraint de email a (tenant_id, email)
ALTER TABLE users DROP CONSTRAINT users_email_key;
CREATE UNIQUE INDEX idx_users_tenant_email ON users(tenant_id, email);

CREATE INDEX idx_users_tenant ON users(tenant_id);

-- =====================================================
-- NUEVO: ROLES Y PERMISSIONS (RBAC)
-- =====================================================

CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE, -- NULL = global role
    name VARCHAR(50) NOT NULL,
    display_name VARCHAR(100),
    description TEXT,
    is_system BOOLEAN DEFAULT false, -- true para roles predefinidos (admin, planner, viewer)
    permissions JSONB NOT NULL, -- ["buffer:read", "buffer:write", "order:approve"]
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, name)
);

CREATE INDEX idx_roles_tenant ON roles(tenant_id);
CREATE INDEX idx_roles_system ON roles(is_system);

-- Tabla pivot: user_roles (many-to-many)
CREATE TABLE IF NOT EXISTS user_roles (
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
    assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    assigned_by INTEGER REFERENCES users(id), -- quien asignó el rol
    PRIMARY KEY (user_id, role_id)
);

CREATE INDEX idx_user_roles_user ON user_roles(user_id);
CREATE INDEX idx_user_roles_role ON user_roles(role_id);

-- =====================================================
-- NUEVO: AUDIT LOG (quien hizo qué, cuándo)
-- =====================================================

CREATE TABLE IF NOT EXISTS audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID REFERENCES tenants(id),
    user_id INTEGER REFERENCES users(id),
    action VARCHAR(50) NOT NULL, -- USER_CREATED, USER_UPDATED, LOGIN_SUCCESS, LOGIN_FAILED, etc.
    resource_type VARCHAR(50), -- USER, TENANT, ROLE
    resource_id VARCHAR(255),
    ip_address INET,
    user_agent TEXT,
    details JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_audit_log_tenant ON audit_log(tenant_id, created_at DESC);
CREATE INDEX idx_audit_log_user ON audit_log(user_id, created_at DESC);
CREATE INDEX idx_audit_log_action ON audit_log(action);

-- =====================================================
-- SEED: ROLES PREDEFINIDOS
-- =====================================================

INSERT INTO roles (id, tenant_id, name, display_name, description, is_system, permissions) VALUES
(
    gen_random_uuid(),
    NULL, -- Global role
    'admin',
    'Administrator',
    'Full system access',
    true,
    '["*:*"]'::jsonb
),
(
    gen_random_uuid(),
    NULL,
    'planner',
    'Demand Planner',
    'Manage buffers, CPD, replenishment suggestions',
    true,
    '["buffer:read", "buffer:write", "buffer:calculate", "cpd:read", "cpd:calculate", "replenishment:read", "replenishment:approve", "product:read", "order:read"]'::jsonb
),
(
    gen_random_uuid(),
    NULL,
    'viewer',
    'Viewer',
    'Read-only access to dashboards',
    true,
    '["buffer:read", "product:read", "order:read", "analytics:read"]'::jsonb
);

-- =====================================================
-- TRIGGER: Auto-actualizar updated_at en tenants y roles
-- =====================================================

CREATE TRIGGER update_tenants_updated_at
    BEFORE UPDATE ON tenants
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_roles_updated_at
    BEFORE UPDATE ON roles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
```

---

## 3. Cambios al Domain Model

### 3.1 Agregar Nuevas Entidades

Crear estos archivos en `internal/domain/`:

**tenant.go**:
```go
package domain

import "time"

type Tenant struct {
    ID                 string    `json:"id"`
    Name               string    `json:"name"`
    Slug               string    `json:"slug"`
    SubscriptionPlan   string    `json:"subscription_plan"`
    SubscriptionStatus string    `json:"subscription_status"`
    MaxUsers           int       `json:"max_users"`
    Settings           string    `json:"settings"` // JSON string
    CreatedAt          time.Time `json:"created_at"`
    UpdatedAt          time.Time `json:"updated_at"`
}

type CreateTenantRequest struct {
    Name  string `json:"name" binding:"required"`
    Slug  string `json:"slug" binding:"required"`
    Plan  string `json:"plan"`
}
```

**role.go**:
```go
package domain

import "time"

type Role struct {
    ID          string    `json:"id"`
    TenantID    *string   `json:"tenant_id"` // NULL para roles globales
    Name        string    `json:"name"`
    DisplayName string    `json:"display_name"`
    Description string    `json:"description"`
    IsSystem    bool      `json:"is_system"`
    Permissions []string  `json:"permissions"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

type Permission string

const (
    // Buffer permissions
    PermissionBufferRead      Permission = "buffer:read"
    PermissionBufferWrite     Permission = "buffer:write"
    PermissionBufferCalculate Permission = "buffer:calculate"

    // CPD permissions
    PermissionCPDRead      Permission = "cpd:read"
    PermissionCPDCalculate Permission = "cpd:calculate"

    // Replenishment permissions
    PermissionReplenishmentRead    Permission = "replenishment:read"
    PermissionReplenishmentApprove Permission = "replenishment:approve"

    // Product permissions
    PermissionProductRead  Permission = "product:read"
    PermissionProductWrite Permission = "product:write"

    // Order permissions
    PermissionOrderRead   Permission = "order:read"
    PermissionOrderCreate Permission = "order:create"

    // Analytics permissions
    PermissionAnalyticsRead Permission = "analytics:read"

    // Wildcard (admin)
    PermissionAll Permission = "*:*"
)

type AssignRoleRequest struct {
    UserID uint   `json:"user_id" binding:"required"`
    RoleID string `json:"role_id" binding:"required"`
}
```

### 3.2 Modificar User Entity

**internal/domain/user.go** (agregar campos):

```go
type User struct {
    ID       uint   `json:"id"`
    TenantID string `json:"tenant_id"` // NUEVO
    Email    string `json:"email"`
    Password string `json:"-"`
    // ... resto de campos existentes

    // NUEVO: Roles asignados (cargados via join)
    Roles []Role `json:"roles,omitempty"`
}

// NUEVO: UpdateUserRequest DTO
type UpdateUserRequest struct {
    FirstName string `json:"first_name"`
    LastName  string `json:"last_name"`
    Phone     string `json:"phone"`
    Avatar    string `json:"avatar,omitempty"`
    IsActive  *bool  `json:"is_active"` // Permitir activar/desactivar
}
```

---

## 4. Agregar gRPC Support

### 4.1 Protobuf Definitions

Crear `api/proto/auth/v1/auth.proto`:

```protobuf
syntax = "proto3";

package giia.auth.v1;

option go_package = "auth-service/pkg/authpb";

import "google/protobuf/timestamp.proto";

service AuthService {
    // Authentication
    rpc Login(LoginRequest) returns (LoginResponse);
    rpc RefreshToken(RefreshTokenRequest) returns (TokenResponse);
    rpc ValidateToken(ValidateTokenRequest) returns (ValidateTokenResponse);

    // User Management
    rpc CreateUser(CreateUserRequest) returns (UserResponse);
    rpc GetUser(GetUserRequest) returns (UserResponse);
    rpc UpdateUser(UpdateUserRequest) returns (UserResponse);
    rpc DeleteUser(DeleteUserRequest) returns (DeleteResponse);
    rpc ListUsers(ListUsersRequest) returns (stream UserResponse);

    // Tenant Management
    rpc CreateTenant(CreateTenantRequest) returns (TenantResponse);
    rpc GetTenant(GetTenantRequest) returns (TenantResponse);

    // Authorization
    rpc CheckPermission(CheckPermissionRequest) returns (PermissionResponse);
    rpc AssignRole(AssignRoleRequest) returns (RoleResponse);
    rpc GetUserRoles(GetUserRolesRequest) returns (RolesResponse);
}

message LoginRequest {
    string email = 1;
    string password = 2;
    string twofa_code = 3;
    string tenant_slug = 4; // NEW: Identify tenant
}

message LoginResponse {
    string access_token = 1;
    string refresh_token = 2;
    google.protobuf.Timestamp expires_at = 3;
    UserResponse user = 4;
}

message ValidateTokenRequest {
    string token = 1;
}

message ValidateTokenResponse {
    bool valid = 1;
    uint32 user_id = 2;
    string tenant_id = 3;
    string email = 4;
    repeated string roles = 5;
    repeated string permissions = 6;
}

message CheckPermissionRequest {
    uint32 user_id = 1;
    string permission = 2; // e.g., "buffer:write"
}

message PermissionResponse {
    bool allowed = 1;
}

message UserResponse {
    uint32 id = 1;
    string tenant_id = 2;
    string email = 3;
    string first_name = 4;
    string last_name = 5;
    bool is_active = 6;
    repeated RoleResponse roles = 7;
}

message RoleResponse {
    string id = 1;
    string name = 2;
    string display_name = 3;
    repeated string permissions = 4;
}

message TenantResponse {
    string id = 1;
    string name = 2;
    string slug = 3;
    string subscription_plan = 4;
}

// ... más mensajes (CreateUserRequest, etc.)
```

### 4.2 gRPC Server Implementation

Crear `internal/adapter/grpc/auth_server.go`:

```go
package grpc

import (
    "context"
    "auth-service/internal/usecases"
    "auth-service/pkg/authpb"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

type AuthServer struct {
    authpb.UnimplementedAuthServiceServer
    userService *usecases.UserService
    authService *usecases.AuthService
}

func NewAuthServer(userService *usecases.UserService, authService *usecases.AuthService) *AuthServer {
    return &AuthServer{
        userService: userService,
        authService: authService,
    }
}

func (s *AuthServer) ValidateToken(ctx context.Context, req *authpb.ValidateTokenRequest) (*authpb.ValidateTokenResponse, error) {
    claims, err := s.authService.ValidateAccessToken(req.Token)
    if err != nil {
        return nil, status.Error(codes.Unauthenticated, "invalid token")
    }

    // Get user with roles
    user, err := s.userService.GetByID(ctx, claims.UserID)
    if err != nil {
        return nil, status.Error(codes.NotFound, "user not found")
    }

    // Extract permissions from roles
    permissions := make([]string, 0)
    roleNames := make([]string, 0)
    for _, role := range user.Roles {
        roleNames = append(roleNames, role.Name)
        permissions = append(permissions, role.Permissions...)
    }

    return &authpb.ValidateTokenResponse{
        Valid:       true,
        UserId:      uint32(user.ID),
        TenantId:    user.TenantID,
        Email:       user.Email,
        Roles:       roleNames,
        Permissions: permissions,
    }, nil
}

func (s *AuthServer) CheckPermission(ctx context.Context, req *authpb.CheckPermissionRequest) (*authpb.PermissionResponse, error) {
    allowed, err := s.authService.HasPermission(ctx, req.UserId, req.Permission)
    if err != nil {
        return nil, status.Error(codes.Internal, "failed to check permission")
    }

    return &authpb.PermissionResponse{
        Allowed: allowed,
    }, nil
}

// ... implementar otros métodos
```

---

## 5. Agregar Event Publishing (NATS)

### 5.1 Event Definitions

Crear `internal/domain/events.go`:

```go
package domain

import "time"

type DomainEvent interface {
    EventType() string
    AggregateID() string
    OccurredAt() time.Time
}

// User Events
type UserCreatedEvent struct {
    UserID    uint      `json:"user_id"`
    TenantID  string    `json:"tenant_id"`
    Email     string    `json:"email"`
    Timestamp time.Time `json:"timestamp"`
}

func (e UserCreatedEvent) EventType() string     { return "auth.user.created" }
func (e UserCreatedEvent) AggregateID() string   { return string(e.UserID) }
func (e UserCreatedEvent) OccurredAt() time.Time { return e.Timestamp }

type UserLoggedInEvent struct {
    UserID    uint      `json:"user_id"`
    TenantID  string    `json:"tenant_id"`
    IPAddress string    `json:"ip_address"`
    Timestamp time.Time `json:"timestamp"`
}

func (e UserLoggedInEvent) EventType() string     { return "auth.user.logged_in" }
func (e UserLoggedInEvent) AggregateID() string   { return string(e.UserID) }
func (e UserLoggedInEvent) OccurredAt() time.Time { return e.Timestamp }

type TenantCreatedEvent struct {
    TenantID  string    `json:"tenant_id"`
    Name      string    `json:"name"`
    Slug      string    `json:"slug"`
    Timestamp time.Time `json:"timestamp"`
}

func (e TenantCreatedEvent) EventType() string     { return "auth.tenant.created" }
func (e TenantCreatedEvent) AggregateID() string   { return e.TenantID }
func (e TenantCreatedEvent) OccurredAt() time.Time { return e.Timestamp }
```

### 5.2 NATS Publisher

Crear `internal/infrastructure/events/nats_publisher.go`:

```go
package events

import (
    "context"
    "encoding/json"
    "fmt"
    "github.com/nats-io/nats.go"
    "auth-service/internal/domain"
)

type EventPublisher interface {
    Publish(ctx context.Context, event domain.DomainEvent) error
}

type natsPublisher struct {
    conn *nats.Conn
}

func NewNATSPublisher(natsURL string) (EventPublisher, error) {
    conn, err := nats.Connect(natsURL)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to NATS: %w", err)
    }

    return &natsPublisher{conn: conn}, nil
}

func (p *natsPublisher) Publish(ctx context.Context, event domain.DomainEvent) error {
    payload, err := json.Marshal(event)
    if err != nil {
        return fmt.Errorf("failed to marshal event: %w", err)
    }

    // Publish to NATS topic
    topic := event.EventType()
    if err := p.conn.Publish(topic, payload); err != nil {
        return fmt.Errorf("failed to publish event to %s: %w", topic, err)
    }

    return nil
}

func (p *natsPublisher) Close() {
    p.conn.Close()
}
```

### 5.3 Integrate in Use Cases

Modificar `internal/usecases/user_service.go`:

```go
type UserService struct {
    repo           repository.UserRepository
    passwordSvc    *auth.PasswordService
    eventPublisher events.EventPublisher // NUEVO
}

func (s *UserService) Register(ctx context.Context, req *domain.RegisterRequest) (*domain.User, error) {
    // ... lógica existente de creación de usuario

    user, err := s.repo.Create(ctx, &domain.User{
        TenantID:  tenantID, // NUEVO: asignar tenant
        Email:     req.Email,
        Password:  hashedPassword,
        FirstName: req.FirstName,
        LastName:  req.LastName,
        // ...
    })

    if err != nil {
        return nil, err
    }

    // NUEVO: Publish event
    event := domain.UserCreatedEvent{
        UserID:    user.ID,
        TenantID:  user.TenantID,
        Email:     user.Email,
        Timestamp: time.Now(),
    }

    if err := s.eventPublisher.Publish(ctx, event); err != nil {
        // Log error but don't fail the operation
        log.Printf("failed to publish user created event: %v", err)
    }

    return user, nil
}
```

---

## 6. API Gateway Integration

### 6.1 Middleware de Autenticación

El API Gateway valida JWT y extrae `tenant_id`:

```go
// En tu API Gateway (Kong, Traefik, o custom Go)
func AuthMiddleware(authServiceClient authpb.AuthServiceClient) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Extract JWT from Authorization header
            authHeader := r.Header.Get("Authorization")
            if authHeader == "" {
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                return
            }

            token := strings.TrimPrefix(authHeader, "Bearer ")

            // Validate token via gRPC to Auth Service
            ctx := r.Context()
            resp, err := authServiceClient.ValidateToken(ctx, &authpb.ValidateTokenRequest{
                Token: token,
            })

            if err != nil || !resp.Valid {
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                return
            }

            // Inject user_id, tenant_id, permissions into context
            ctx = context.WithValue(ctx, "user_id", resp.UserId)
            ctx = context.WithValue(ctx, "tenant_id", resp.TenantId)
            ctx = context.WithValue(ctx, "permissions", resp.Permissions)

            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

### 6.2 Todos los servicios reciben contexto

Los otros servicios (DDMRP Engine, Catalog, etc.) reciben `user_id` y `tenant_id` desde el gateway:

```go
// En DDMRP Engine Service (ejemplo)
func (h *BufferHandler) GetBuffers(w http.ResponseWriter, r *http.Request) {
    tenantID := r.Context().Value("tenant_id").(string)
    userID := r.Context().Value("user_id").(uint32)

    // Filtrar por tenant
    buffers, err := h.bufferService.GetAllBuffers(r.Context(), tenantID)
    // ...
}
```

---

## 7. Plan de Implementación

### Fase 1: Schema & Domain (1 semana)
- [ ] Ejecutar SQL migrations (tenants, roles, multi-tenancy)
- [ ] Agregar domain entities (Tenant, Role)
- [ ] Modificar User entity (tenant_id, roles)
- [ ] Seed roles predefinidos

### Fase 2: RBAC Logic (1 semana)
- [ ] Implementar TenantRepository
- [ ] Implementar RoleRepository
- [ ] Agregar métodos de autorización en UserService
- [ ] Unit tests para RBAC

### Fase 3: gRPC (3-4 días)
- [ ] Escribir protobuf definitions
- [ ] Generar código Go (`protoc`)
- [ ] Implementar gRPC server
- [ ] Integration tests

### Fase 4: Events (2-3 días)
- [ ] Agregar NATS client
- [ ] Implementar EventPublisher
- [ ] Publish events en use cases
- [ ] Test event publishing

### Fase 5: Integration Testing (2-3 días)
- [ ] E2E tests multi-tenant
- [ ] E2E tests RBAC
- [ ] E2E tests gRPC
- [ ] Load testing

---

## 8. Resumen de Cambios

| Componente | Cambios | Complejidad |
|------------|---------|-------------|
| **Database Schema** | Agregar tenants, roles, user_roles, audit_log | Media |
| **Domain** | Agregar Tenant, Role entities | Baja |
| **Repository** | Agregar TenantRepo, RoleRepo | Media |
| **Use Cases** | Agregar tenant_id en operaciones, RBAC logic | Media |
| **API** | Agregar gRPC server | Alta |
| **Events** | Agregar NATS publisher | Baja |
| **Testing** | Nuevos tests para multi-tenancy, RBAC | Media |

**Total**: 2-3 semanas con 1 ingeniero full-time

---

## 9. Siguiente Paso

**¿Quieres que implemente esto?**

Opciones:
1. ✅ **Genero el código completo** (migrations SQL, domain entities, gRPC server, NATS publisher)
2. ✅ **Te guío paso a paso** (commit por commit)
3. ✅ **Solo dame las dudas** y resuelvo específicamente

**¿Qué prefieres?** 🚀
