# GIIA Authentication Guide

This guide covers authentication and authorization for the GIIA platform APIs.

---

## Overview

GIIA uses **JWT (JSON Web Tokens)** for authentication with a dual-token strategy:
- **Access Token**: Short-lived (15 min), used for API requests
- **Refresh Token**: Long-lived (7 days), stored as HTTP-only cookie

---

## Authentication Flow

```
┌─────────────┐      1. Login         ┌─────────────┐
│   Client    │─────────────────────►│ Auth Service │
│             │◄─────────────────────│             │
│             │  Access + Refresh     └─────────────┘
│             │        Tokens
│             │
│             │      2. API Request    ┌─────────────┐
│             │─────────────────────►│ Any Service  │
│             │ Authorization: Bearer │             │
│             │◄─────────────────────│             │
│             │       Response        └─────────────┘
│             │
│             │      3. Token Expired  ┌─────────────┐
│             │─────────────────────►│ Auth Service │
│             │   POST /auth/refresh  │             │
│             │◄─────────────────────│             │
│             │   New Access Token    └─────────────┘
└─────────────┘
```

---

## Login

### Request

```bash
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "SecurePassword123!"
}
```

### Response

```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 900,
  "user": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "email": "user@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "organization_id": "550e8400-e29b-41d4-a716-446655440000",
    "status": "active",
    "roles": ["user", "analyst"]
  }
}
```

**Note:** The `refresh_token` is set as an HTTP-only cookie, not in the response body.

### Rate Limiting

- **5 requests per 15 minutes per IP**
- After exceeding: 429 Too Many Requests

---

## Using Access Tokens

### HTTP Header

```bash
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### With Organization Context

Most endpoints require organization context:

```bash
curl http://localhost:8082/api/v1/products \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "X-Organization-ID: $ORG_ID"
```

**Note:** The `X-Organization-ID` header is optional if the user belongs to only one organization; it's extracted from the JWT in that case.

---

## Token Structure

### Access Token Claims

```json
{
  "sub": "123e4567-e89b-12d3-a456-426614174000",
  "email": "user@example.com",
  "organization_id": "550e8400-e29b-41d4-a716-446655440000",
  "roles": ["user", "analyst"],
  "permissions": ["products:read", "products:create", "analytics:read"],
  "iat": 1642248000,
  "exp": 1642248900
}
```

| Claim | Description |
|-------|-------------|
| `sub` | User ID (UUID) |
| `email` | User email |
| `organization_id` | Current organization |
| `roles` | Assigned roles |
| `permissions` | Resolved permissions |
| `iat` | Issued at (Unix timestamp) |
| `exp` | Expires at (Unix timestamp) |

---

## Token Refresh

When the access token expires (after 15 minutes), refresh it:

### Request

```bash
POST /api/v1/auth/refresh
Cookie: refresh_token=...
```

### Response

```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 900
}
```

### Best Practice: Proactive Refresh

Refresh before expiry to avoid failed requests:

```javascript
// Check token expiry
const tokenPayload = JSON.parse(atob(accessToken.split('.')[1]));
const expiresAt = tokenPayload.exp * 1000;
const now = Date.now();
const refreshThreshold = 60 * 1000; // 1 minute before expiry

if (expiresAt - now < refreshThreshold) {
  await refreshToken();
}
```

---

## Logout

### Request

```bash
POST /api/v1/auth/logout
Authorization: Bearer <access_token>
```

### Response

```json
{
  "message": "Logged out successfully"
}
```

This invalidates:
- The current access token (added to blacklist)
- The refresh token (deleted from database)
- The HTTP-only cookie (cleared)

---

## Password Management

### Forgot Password

```bash
POST /api/v1/auth/forgot-password
Content-Type: application/json

{
  "email": "user@example.com",
  "organization_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Note:** Always returns success to prevent user enumeration.

### Reset Password

```bash
POST /api/v1/auth/reset-password
Content-Type: application/json

{
  "token": "reset-token-from-email",
  "password": "NewSecurePassword123!"
}
```

### Password Requirements

- Minimum 8 characters
- At least 1 uppercase letter
- At least 1 lowercase letter
- At least 1 digit
- At least 1 special character (!@#$%^&*()_+-=[]{}|;:,.<>?)

---

## Multi-Tenancy

### Organization Isolation

All data is isolated by organization. Users can belong to multiple organizations.

### Switching Organizations

If a user belongs to multiple organizations, specify the target:

```bash
curl http://localhost:8082/api/v1/products \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "X-Organization-ID: different-org-uuid"
```

The user must have access to the specified organization.

---

## Role-Based Access Control (RBAC)

### Default Roles

| Role | Description |
|------|-------------|
| `super_admin` | Full platform access |
| `org_admin` | Full organization access |
| `manager` | Can manage products, orders, buffers |
| `analyst` | Read-only analytics access |
| `user` | Basic read access |

### Permission Format

```
resource:action
```

Examples:
- `products:create`
- `orders:read`
- `analytics:read`
- `users:delete`

### Checking Permissions

Via Auth Service gRPC:

```go
resp, err := authClient.CheckPermission(ctx, &authv1.CheckPermissionRequest{
    UserId:         userID,
    OrganizationId: orgID,
    Permission:     "products:create",
})

if !resp.Allowed {
    return errors.New("permission denied")
}
```

---

## Security Best Practices

### Client-Side

1. **Never store access tokens in localStorage** - Use memory or secure cookies
2. **Implement token refresh** - Don't wait for 401 errors
3. **Clear tokens on logout** - Don't leave stale tokens

### Token Storage Recommendations

| Platform | Recommendation |
|----------|----------------|
| Browser SPA | In-memory variable |
| React Native | Secure storage |
| Mobile App | Keychain/Keystore |
| Backend | Environment variables |

### Secure Request Pattern

```javascript
class APIClient {
  constructor() {
    this.accessToken = null;
    this.refreshPromise = null;
  }

  async request(url, options = {}) {
    // Ensure valid token
    await this.ensureValidToken();
    
    const response = await fetch(url, {
      ...options,
      headers: {
        ...options.headers,
        'Authorization': `Bearer ${this.accessToken}`,
      },
    });

    // Handle token expiry
    if (response.status === 401) {
      await this.refreshToken();
      return this.request(url, options);
    }

    return response;
  }

  async ensureValidToken() {
    if (!this.accessToken) {
      throw new Error('Not authenticated');
    }
    
    // Check expiry
    const payload = this.parseToken(this.accessToken);
    const expiresIn = payload.exp * 1000 - Date.now();
    
    if (expiresIn < 60000) { // Less than 1 minute
      await this.refreshToken();
    }
  }

  async refreshToken() {
    // Prevent concurrent refresh calls
    if (this.refreshPromise) {
      return this.refreshPromise;
    }
    
    this.refreshPromise = fetch('/api/v1/auth/refresh', {
      method: 'POST',
      credentials: 'include', // Include cookies
    })
    .then(r => r.json())
    .then(data => {
      this.accessToken = data.access_token;
    })
    .finally(() => {
      this.refreshPromise = null;
    });
    
    return this.refreshPromise;
  }
}
```

---

## gRPC Authentication

For service-to-service gRPC calls, include the token in metadata:

```go
import "google.golang.org/grpc/metadata"

func callWithAuth(ctx context.Context, token string) error {
    md := metadata.Pairs("authorization", "Bearer "+token)
    ctx = metadata.NewOutgoingContext(ctx, md)
    
    // Make gRPC call with authenticated context
    resp, err := client.SomeMethod(ctx, request)
    return err
}
```

---

## Troubleshooting

### Common Issues

| Error | Cause | Solution |
|-------|-------|----------|
| 401 Unauthorized | Invalid/expired token | Refresh or re-login |
| 403 Forbidden | Insufficient permissions | Check user roles |
| 429 Too Many Requests | Rate limit exceeded | Wait and retry |
| Token decode error | Malformed token | Check token format |

### Debugging Tokens

```bash
# Decode JWT (base64)
echo "eyJhbGciOi..." | cut -d'.' -f2 | base64 -d | jq

# Check expiry
node -e "console.log(new Date(JSON.parse(atob('eyJhbGciOi...'.split('.')[1])).exp * 1000))"
```

---

## Related Documentation

- [Getting Started](./getting-started.md)
- [API Reference](./index.md)
- [Auth Service OpenAPI](/services/auth-service/docs/openapi.yaml)
- [Auth gRPC Documentation](./grpc/auth.md)
