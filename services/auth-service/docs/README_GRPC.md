# Auth Service - gRPC API Documentation

## Overview

The Auth Service provides a gRPC API for authentication and authorization operations. This document describes how to use the gRPC service for inter-service communication within the GIIA microservices architecture.

## gRPC Server Configuration

### Ports
- **HTTP Server**: `:8081` (configurable via `PORT` env var)
- **gRPC Server**: `:9091` (configurable via `GRPC_PORT` env var)

### Features
- JWT token validation
- Permission checking (single and batch)
- User information retrieval
- Health checks
- Prometheus metrics
- Request logging with request IDs
- Panic recovery
- Multi-tenancy support

## Service Definition

### Proto Files Location
```
services/auth-service/api/proto/auth/v1/
├── messages.proto    # Shared message types
└── auth.proto        # Service and RPC definitions
```

### Generated Code Location
```
services/auth-service/api/proto/gen/go/auth/v1/
├── messages.pb.go
├── auth.pb.go
└── auth_grpc.pb.go
```

## Available RPCs

### 1. ValidateToken

Validates a JWT token and returns user information.

**Request:**
```protobuf
message ValidateTokenRequest {
  string token = 1;
}
```

**Response:**
```protobuf
message ValidateTokenResponse {
  bool valid = 1;
  string reason = 2;
  UserInfo user = 3;
  int64 expires_at = 4;
}
```

**Usage Example:**
```go
resp, err := client.ValidateToken(ctx, &authv1.ValidateTokenRequest{
    Token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
})

if err != nil {
    return fmt.Errorf("failed to validate token: %w", err)
}

if !resp.Valid {
    return fmt.Errorf("token invalid: %s", resp.Reason)
}

// Use user information
userID := resp.User.UserId
email := resp.User.Email
```

### 2. CheckPermission

Checks if a user has a specific permission.

**Request:**
```protobuf
message CheckPermissionRequest {
  string user_id = 1;
  string organization_id = 2;
  string permission = 3;  // Format: "service:resource:action"
}
```

**Response:**
```protobuf
message CheckPermissionResponse {
  bool allowed = 1;
  string reason = 2;
}
```

**Usage Example:**
```go
resp, err := client.CheckPermission(ctx, &authv1.CheckPermissionRequest{
    UserId:     "550e8400-e29b-41d4-a716-446655440000",
    Permission: "catalog:products:read",
})

if err != nil {
    return fmt.Errorf("failed to check permission: %w", err)
}

if !resp.Allowed {
    return fmt.Errorf("permission denied: %s", resp.Reason)
}
```

### 3. BatchCheckPermissions

Checks multiple permissions in a single call for better performance.

**Request:**
```protobuf
message BatchCheckPermissionsRequest {
  string user_id = 1;
  string organization_id = 2;
  repeated string permissions = 3;
}
```

**Response:**
```protobuf
message BatchCheckPermissionsResponse {
  repeated bool results = 1;  // Results in same order as request permissions
}
```

**Usage Example:**
```go
resp, err := client.BatchCheckPermissions(ctx, &authv1.BatchCheckPermissionsRequest{
    UserId: "550e8400-e29b-41d4-a716-446655440000",
    Permissions: []string{
        "catalog:products:read",
        "catalog:products:write",
        "inventory:items:read",
    },
})

if err != nil {
    return fmt.Errorf("failed to check permissions: %w", err)
}

// Check individual results
canRead := resp.Results[0]
canWrite := resp.Results[1]
canReadInventory := resp.Results[2]
```

### 4. GetUser

Retrieves user information by user ID.

**Request:**
```protobuf
message GetUserRequest {
  string user_id = 1;
  string organization_id = 2;  // Optional: validates user belongs to org
}
```

**Response:**
```protobuf
message GetUserResponse {
  UserInfo user = 1;
}

message UserInfo {
  string user_id = 1;
  string organization_id = 2;
  string email = 3;
  repeated string roles = 4;
  string name = 5;
  string status = 6;
  string first_name = 7;
  string last_name = 8;
}
```

**Usage Example:**
```go
resp, err := client.GetUser(ctx, &authv1.GetUserRequest{
    UserId:         "550e8400-e29b-41d4-a716-446655440000",
    OrganizationId: "660e8400-e29b-41d4-a716-446655440001",
})

if err != nil {
    return fmt.Errorf("failed to get user: %w", err)
}

user := resp.User
fmt.Printf("User: %s (%s)\n", user.Name, user.Email)
```

## Client Library Usage

### Single Client

```go
package main

import (
    "context"
    "log"

    "github.com/giia/giia-core-engine/services/auth-service/internal/infrastructure/grpc/client"
)

func main() {
    // Create client
    authClient, err := client.NewAuthClient(&client.ClientConfig{
        Address: "localhost:9091",
        Timeout: 10 * time.Second,
    })
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }
    defer authClient.Close()

    // Validate token
    resp, err := authClient.ValidateToken(context.Background(), "token", "request-id")
    if err != nil {
        log.Fatalf("Failed to validate token: %v", err)
    }

    if resp.Valid {
        log.Printf("Token valid for user: %s", resp.User.Email)
    }
}
```

### Connection Pool (Recommended for High Traffic)

```go
package main

import (
    "context"
    "log"

    "github.com/giia/giia-core-engine/services/auth-service/internal/infrastructure/grpc/client"
)

func main() {
    // Create connection pool
    pool, err := client.NewConnectionPool("localhost:9091", 10)
    if err != nil {
        log.Fatalf("Failed to create connection pool: %v", err)
    }
    defer pool.Close()

    // Get client from pool
    authClient := pool.GetClient()

    // Use client
    resp, err := authClient.CheckPermission(
        context.Background(),
        "user-id",
        "catalog:products:read",
        "request-id",
    )
    if err != nil {
        log.Fatalf("Failed to check permission: %v", err)
    }

    if resp.Allowed {
        log.Println("Permission granted")
    }
}
```

### Using as HTTP Middleware

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/giia/giia-core-engine/services/auth-service/internal/infrastructure/grpc/client"
)

func main() {
    // Create auth client
    authClient, _ := client.NewAuthClient(&client.ClientConfig{
        Address: "localhost:9091",
    })
    defer authClient.Close()

    // Create middleware
    authMiddleware := client.NewAuthMiddleware(authClient)

    // Setup routes
    r := gin.Default()

    // Public routes
    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })

    // Protected routes
    protected := r.Group("/api/v1")
    protected.Use(authMiddleware.RequireAuth())
    {
        protected.GET("/products", listProducts)

        // Routes with specific permissions
        protected.POST("/products",
            authMiddleware.RequirePermission("catalog:products:write"),
            createProduct,
        )
    }

    r.Run(":8080")
}
```

## Health Checks

The gRPC server implements the standard gRPC health check protocol.

### Using grpc-health-probe

```bash
# Install grpc-health-probe
go install github.com/grpc-ecosystem/grpc-health-probe@latest

# Check health
grpc-health-probe -addr=localhost:9091
```

### Kubernetes Liveness/Readiness Probes

```yaml
livenessProbe:
  exec:
    command: ["/bin/grpc-health-probe", "-addr=:9091"]
  initialDelaySeconds: 10
  periodSeconds: 10

readinessProbe:
  exec:
    command: ["/bin/grpc-health-probe", "-addr=:9091"]
  initialDelaySeconds: 5
  periodSeconds: 5
```

## Metrics

The gRPC server exposes Prometheus metrics for monitoring.

### Available Metrics

- `grpc_requests_total{method, code}` - Total number of gRPC requests
- `grpc_request_duration_seconds{method}` - Request duration histogram

### Accessing Metrics

```bash
# Assuming metrics are exposed on HTTP server
curl http://localhost:8081/metrics
```

### Example Prometheus Queries

```promql
# Request rate by method
rate(grpc_requests_total[5m])

# P95 latency by method
histogram_quantile(0.95, rate(grpc_request_duration_seconds_bucket[5m]))

# Error rate
rate(grpc_requests_total{code!="OK"}[5m])
```

## Request IDs and Tracing

The gRPC server supports request IDs for distributed tracing.

### Client Side (Sending Request ID)

```go
import "google.golang.org/grpc/metadata"

func callWithRequestID(client authv1.AuthServiceClient, requestID string) {
    // Add request ID to metadata
    md := metadata.Pairs("x-request-id", requestID)
    ctx := metadata.NewOutgoingContext(context.Background(), md)

    // Make request
    resp, err := client.ValidateToken(ctx, &authv1.ValidateTokenRequest{
        Token: "token",
    })
}
```

### Server Side (Extracting Request ID)

Request IDs are automatically extracted and included in logs by the logging interceptor.

## Error Handling

### gRPC Status Codes

| gRPC Code | When Used |
|-----------|-----------|
| `OK` | Request succeeded |
| `INVALID_ARGUMENT` | Invalid input (empty required fields, bad format) |
| `UNAUTHENTICATED` | Invalid or expired token |
| `PERMISSION_DENIED` | User lacks required permission or belongs to different org |
| `NOT_FOUND` | User or resource not found |
| `INTERNAL` | Internal server error |

### Error Handling Example

```go
import (
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

resp, err := client.ValidateToken(ctx, req)
if err != nil {
    st, ok := status.FromError(err)
    if ok {
        switch st.Code() {
        case codes.InvalidArgument:
            log.Println("Bad request:", st.Message())
        case codes.Unauthenticated:
            log.Println("Authentication failed:", st.Message())
        case codes.Internal:
            log.Println("Server error:", st.Message())
        default:
            log.Println("Unexpected error:", st.Message())
        }
    }
    return err
}
```

## Performance Considerations

### Caching

Permission checks are cached in Redis for 5 minutes to improve performance:
- Cache hit rate target: >95%
- Average latency: <5ms (cache hit)
- Average latency: <50ms (cache miss)

### Batch Operations

Use `BatchCheckPermissions` instead of multiple `CheckPermission` calls:

**Bad:**
```go
for _, perm := range permissions {
    resp, _ := client.CheckPermission(ctx, userID, perm, requestID)
    results = append(results, resp.Allowed)
}
```

**Good:**
```go
resp, _ := client.BatchCheckPermissions(ctx, userID, permissions, requestID)
results = resp.Results
```

### Connection Pooling

For high-traffic services, use connection pooling:

```go
// Create pool once at startup
pool, _ := client.NewConnectionPool("localhost:9091", 10)
defer pool.Close()

// Reuse clients from pool
authClient := pool.GetClient()
```

## Security

### TLS/mTLS

For production deployments, enable TLS:

```go
import "google.golang.org/grpc/credentials"

creds, err := credentials.NewClientTLSFromFile("ca.pem", "")
conn, err := grpc.Dial(
    "auth-service:9091",
    grpc.WithTransportCredentials(creds),
)
```

### Token Security

- Tokens are never logged
- Failed validation attempts are logged for security monitoring
- Tokens should be transmitted over encrypted connections only

## Troubleshooting

### Connection Refused

```
Error: connection refused
```

**Solutions:**
1. Verify gRPC server is running: `lsof -i :9091`
2. Check firewall rules
3. Verify address in client configuration

### Context Deadline Exceeded

```
Error: context deadline exceeded
```

**Solutions:**
1. Increase client timeout
2. Check network latency
3. Check server load and scaling

### Permission Denied

```
Error: permission denied: user belongs to different organization
```

**Solutions:**
1. Verify user belongs to the organization
2. Check organization_id parameter matches user's org
3. Verify user has active status

### Redis Connection Failed

```
Warning: Failed to connect to Redis
```

**Impact:**
- Permission checks will not be cached
- Performance degraded (direct database queries)
- No loss of functionality

**Solutions:**
1. Verify Redis is running
2. Check Redis connection string in config
3. Verify network connectivity to Redis

## Development Tools

### grpcurl

Test gRPC endpoints using grpcurl:

```bash
# List services
grpcurl -plaintext localhost:9091 list

# List methods
grpcurl -plaintext localhost:9091 list auth.v1.AuthService

# Call method
grpcurl -plaintext -d '{"token": "eyJ..."}' \
    localhost:9091 auth.v1.AuthService/ValidateToken
```

### BloomRPC

Use BloomRPC GUI client for testing:
1. Import proto files from `api/proto/auth/v1/`
2. Connect to `localhost:9091`
3. Test RPCs interactively

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `GRPC_PORT` | gRPC server port | `9091` |
| `REDIS_HOST` | Redis host for caching | `localhost` |
| `REDIS_PORT` | Redis port | `6379` |
| `REDIS_DB` | Redis database number | `1` |
| `JWT_SECRET` | JWT signing secret | - |

## Migration Guide

### From HTTP to gRPC

**Before (HTTP):**
```go
resp, err := http.Post(
    "http://auth-service:8081/api/v1/auth/validate",
    "application/json",
    bytes.NewBuffer(jsonData),
)
```

**After (gRPC):**
```go
resp, err := client.ValidateToken(
    ctx,
    "token",
    "request-id",
)
```

**Benefits:**
- 10x faster (binary protocol)
- Type-safe (protocol buffers)
- Built-in connection pooling
- Automatic retries and load balancing

## References

- [gRPC Documentation](https://grpc.io/docs/)
- [Protocol Buffers](https://developers.google.com/protocol-buffers)
- [gRPC Health Check Protocol](https://github.com/grpc/grpc/blob/master/doc/health-checking.md)
- [Prometheus Metrics](https://prometheus.io/docs/concepts/metric_types/)
