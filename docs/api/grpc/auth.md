# Auth Service gRPC API

**Package:** `auth.v1`  
**Port:** 9081  
**Proto File:** `services/auth-service/api/proto/auth/v1/auth.proto`

---

## Service Definition

```protobuf
service AuthService {
  rpc ValidateToken(ValidateTokenRequest) returns (ValidateTokenResponse);
  rpc CheckPermission(CheckPermissionRequest) returns (CheckPermissionResponse);
  rpc BatchCheckPermissions(BatchCheckPermissionsRequest) returns (BatchCheckPermissionsResponse);
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
}
```

---

## Methods

### ValidateToken

Validates a JWT access token and returns user information.

**Request:**
```protobuf
message ValidateTokenRequest {
  string token = 1;  // JWT access token
}
```

**Response:**
```protobuf
message ValidateTokenResponse {
  bool valid = 1;           // Whether token is valid
  string reason = 2;        // Reason if invalid
  UserInfo user = 3;        // User info if valid
  int64 expires_at = 4;     // Token expiration timestamp
}
```

**Example (Go):**
```go
resp, err := client.ValidateToken(ctx, &authv1.ValidateTokenRequest{
    Token: "eyJhbGciOiJIUzI1NiIs...",
})
if err != nil {
    log.Fatal(err)
}
if !resp.Valid {
    log.Printf("Token invalid: %s", resp.Reason)
}
```

---

### CheckPermission

Checks if a user has a specific permission.

**Request:**
```protobuf
message CheckPermissionRequest {
  string user_id = 1;         // User UUID
  string organization_id = 2; // Organization UUID
  string permission = 3;      // Permission in format "resource:action"
}
```

**Response:**
```protobuf
message CheckPermissionResponse {
  bool allowed = 1;    // Whether permission is granted
  string reason = 2;   // Explanation if denied
}
```

**Example (Go):**
```go
resp, err := client.CheckPermission(ctx, &authv1.CheckPermissionRequest{
    UserId:         userID,
    OrganizationId: orgID,
    Permission:     "products:create",
})
if resp.Allowed {
    // Proceed with operation
}
```

---

### BatchCheckPermissions

Checks multiple permissions in a single call for efficiency.

**Request:**
```protobuf
message BatchCheckPermissionsRequest {
  string user_id = 1;
  string organization_id = 2;
  repeated string permissions = 3;  // List of permissions to check
}
```

**Response:**
```protobuf
message BatchCheckPermissionsResponse {
  repeated bool results = 1;  // Results in same order as request
}
```

**Example (Go):**
```go
resp, err := client.BatchCheckPermissions(ctx, &authv1.BatchCheckPermissionsRequest{
    UserId:         userID,
    OrganizationId: orgID,
    Permissions:    []string{"products:read", "products:create", "orders:create"},
})
// resp.Results = [true, true, false]
```

---

### GetUser

Retrieves user information by user ID.

**Request:**
```protobuf
message GetUserRequest {
  string user_id = 1;
  string organization_id = 2;
}
```

**Response:**
```protobuf
message GetUserResponse {
  UserInfo user = 1;
}
```

---

## Message Types

### UserInfo

```protobuf
message UserInfo {
  string id = 1;
  string email = 2;
  string first_name = 3;
  string last_name = 4;
  string organization_id = 5;
  string status = 6;            // pending, active, suspended
  repeated string roles = 7;
  repeated string permissions = 8;
}
```

---

## Error Codes

| gRPC Code | Description |
|-----------|-------------|
| `UNAUTHENTICATED` (16) | Invalid or expired token |
| `PERMISSION_DENIED` (7) | User lacks required permission |
| `NOT_FOUND` (5) | User not found |
| `INVALID_ARGUMENT` (3) | Invalid request parameters |
| `INTERNAL` (13) | Server error |

---

## Connection Example

```go
import (
    authv1 "github.com/giia/giia-core-engine/services/auth-service/api/proto/gen/go/auth/v1"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

func NewAuthClient() (authv1.AuthServiceClient, error) {
    conn, err := grpc.Dial(
        "auth-service:9081",
        grpc.WithTransportCredentials(insecure.NewCredentials()),
    )
    if err != nil {
        return nil, err
    }
    return authv1.NewAuthServiceClient(conn), nil
}
```

---

## Usage in Middleware

The Auth Service is typically called from other services' middleware:

```go
func AuthMiddleware(authClient authv1.AuthServiceClient) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            token := extractToken(r)
            
            resp, err := authClient.ValidateToken(r.Context(), &authv1.ValidateTokenRequest{
                Token: token,
            })
            
            if err != nil || !resp.Valid {
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                return
            }
            
            ctx := context.WithValue(r.Context(), "user", resp.User)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

---

**Related Documentation:**
- [Auth Service REST API](/docs/api/AUTH_SERVICE_API.md)
- [Public API RFC](/docs/api/PUBLIC_RFC.md)
