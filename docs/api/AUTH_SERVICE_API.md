# GIIA Auth Service API Reference

**Version**: 1.0  
**Last Updated**: 2025-12-23  
**Base URL**: `http://localhost:8081/api/v1`

---

## 📖 Overview

The Auth Service provides authentication, authorization, and multi-tenancy management for the GIIA platform.

### Key Features

- ✅ JWT-based authentication (access + refresh tokens)
- ✅ Multi-tenant organization isolation
- ✅ Role-Based Access Control (RBAC)
- ✅ User registration with email verification
- ✅ Password reset functionality
- ✅ Rate limiting for security endpoints

---

## 🔐 Authentication Endpoints

### POST /auth/register

Create a new user account.

**Authentication**: Not required  
**Rate Limit**: 3/60min/IP

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "SecurePassword123!",
  "first_name": "John",
  "last_name": "Doe",
  "phone": "+1234567890",
  "organization_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Validation:**
| Field | Rules |
|-------|-------|
| `email` | Required, valid email format, unique per org |
| `password` | Min 8 chars, uppercase, lowercase, digit, special char |
| `first_name` | Required, max 100 chars |
| `last_name` | Required, max 100 chars |
| `phone` | Optional, E.164 format |
| `organization_id` | Required, valid UUID |

**Success Response (201 Created):**
```json
{
  "message": "User registered successfully. Please check your email for activation instructions."
}
```

**Error Responses:**

| Status | Error Code | Description |
|--------|------------|-------------|
| 400 | BAD_REQUEST | Validation failed |
| 409 | CONFLICT | User with email already exists |

---

### POST /auth/login

Authenticate user and receive JWT tokens.

**Authentication**: Not required  
**Rate Limit**: 5/15min/IP

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "SecurePassword123!"
}
```

**Success Response (200 OK):**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "expires_in": 900,
  "user": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "email": "user@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "organization_id": "550e8400-e29b-41d4-a716-446655440000",
    "status": "active",
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

**Note**: `refresh_token` is returned as HTTP-only cookie with 7-day expiry.

**Error Responses:**

| Status | Error Code | Description |
|--------|------------|-------------|
| 401 | UNAUTHORIZED | Invalid email or password |
| 403 | FORBIDDEN | Account not activated |

---

### POST /auth/refresh

Obtain new access token using refresh token.

**Authentication**: Refresh token (HTTP-only cookie)

**Success Response (200 OK):**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "expires_in": 900
}
```

**Error Responses:**

| Status | Error Code | Description |
|--------|------------|-------------|
| 401 | UNAUTHORIZED | Invalid or expired refresh token |

---

### POST /auth/logout

Invalidate current session and refresh token.

**Authentication**: Required (Bearer token)

**Success Response (200 OK):**
```json
{
  "message": "Logged out successfully"
}
```

---

### POST /auth/activate

Activate user account using email token.

**Authentication**: Not required

**Request Body:**
```json
{
  "token": "activation-token-from-email"
}
```

**Alternative:** `GET /auth/activate?token=activation-token`

**Success Response (200 OK):**
```json
{
  "message": "Account activated successfully. You can now log in."
}
```

**Error Responses:**

| Status | Error Code | Description |
|--------|------------|-------------|
| 400 | BAD_REQUEST | Invalid or expired activation token |

---

### POST /auth/forgot-password

Request password reset email.

**Authentication**: Not required  
**Rate Limit**: 3/60min/IP

**Request Body:**
```json
{
  "email": "user@example.com",
  "organization_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Success Response (200 OK):**
```json
{
  "message": "If an account exists, a password reset email has been sent."
}
```

**Note**: Always returns success to prevent user enumeration.

---

### POST /auth/reset-password

Reset password using token from email.

**Authentication**: Not required

**Request Body:**
```json
{
  "token": "reset-token-from-email",
  "password": "NewSecurePassword123!"
}
```

**Success Response (200 OK):**
```json
{
  "message": "Password reset successfully. You can now log in."
}
```

---

## 👤 User Endpoints

### GET /users/me

Get current authenticated user.

**Authentication**: Required

**Success Response (200 OK):**
```json
{
  "user": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "email": "user@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "phone": "+1234567890",
    "organization_id": "550e8400-e29b-41d4-a716-446655440000",
    "status": "active",
    "roles": ["user", "analyst"],
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

---

### PUT /users/me

Update current user profile.

**Authentication**: Required

**Request Body:**
```json
{
  "first_name": "John",
  "last_name": "Smith",
  "phone": "+1987654321"
}
```

**Success Response (200 OK):**
```json
{
  "user": { ... }
}
```

---

## 🔑 gRPC Service

**Port**: 9081  
**Proto File**: `api/proto/auth/v1/auth.proto`

### Service Definition

```protobuf
service AuthService {
  rpc ValidateToken(ValidateTokenRequest) returns (ValidateTokenResponse);
  rpc CheckPermission(CheckPermissionRequest) returns (CheckPermissionResponse);
  rpc BatchCheckPermissions(BatchCheckPermissionsRequest) returns (BatchCheckPermissionsResponse);
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
}
```

### RPC Methods

#### ValidateToken

Validate JWT access token.

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
  User user = 2;
  string error_message = 3;
}
```

---

#### CheckPermission

Check if user has specific permission.

**Request:**
```protobuf
message CheckPermissionRequest {
  string user_id = 1;
  string organization_id = 2;
  string resource = 3;  // e.g., "products"
  string action = 4;    // e.g., "create", "read", "update", "delete"
}
```

**Response:**
```protobuf
message CheckPermissionResponse {
  bool allowed = 1;
  string reason = 2;
}
```

---

#### BatchCheckPermissions

Check multiple permissions at once.

**Request:**
```protobuf
message BatchCheckPermissionsRequest {
  string user_id = 1;
  string organization_id = 2;
  repeated PermissionCheck permissions = 3;
}
```

**Response:**
```protobuf
message BatchCheckPermissionsResponse {
  repeated PermissionResult results = 1;
}
```

---

## 🔒 JWT Token Structure

### Access Token Claims

```json
{
  "sub": "user-uuid",
  "email": "user@example.com",
  "organization_id": "org-uuid",
  "roles": ["user", "analyst"],
  "permissions": ["products:read", "products:create"],
  "iat": 1642248000,
  "exp": 1642248900
}
```

### Token Lifetimes

| Token | Lifetime | Storage |
|-------|----------|---------|
| Access Token | 15 min | Client memory |
| Refresh Token | 7 days | HTTP-only cookie |

---

## 📊 Health Check

### GET /health

**Response (200 OK):**
```json
{
  "status": "healthy",
  "database": "connected",
  "redis": "connected",
  "version": "1.0.0"
}
```

---

## 🛡️ Security

### Password Requirements

- Minimum 8 characters
- At least 1 uppercase letter
- At least 1 lowercase letter
- At least 1 digit
- At least 1 special character (!@#$%^&*()_+-=[]{}|;:,.<>?)

### Token Security

- Passwords hashed with bcrypt (cost 12)
- Refresh tokens stored as SHA-256 hashes
- Revoked access tokens blacklisted in Redis
- HTTPS required in production

### Rate Limiting

| Endpoint | Limit |
|----------|-------|
| /auth/login | 5/15min/IP |
| /auth/register | 3/60min/IP |
| /auth/forgot-password | 3/60min/IP |
| Other endpoints | 100/min/user |

---

## 📚 Related Documentation

- [Public API RFC](./PUBLIC_RFC.md)
- [gRPC Contracts](./GRPC_CONTRACTS.md)
- [Auth Service README](/services/auth-service/README.md)

---

**Auth Service maintained by the GIIA Team** 🔐
