# Auth Service - Authentication API Documentation

This document describes the REST API endpoints for user authentication, registration, and account management.

## Table of Contents
- [Authentication Endpoints](#authentication-endpoints)
- [User Management Endpoints](#user-management-endpoints)
- [Error Responses](#error-responses)
- [Security](#security)

---

## Base URL

```
http://localhost:8080/api/v1
```

---

## Authentication Endpoints

### 1. User Registration

Register a new user account with email verification.

**Endpoint:** `POST /auth/register`

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "SecureP@ss123",
  "first_name": "John",
  "last_name": "Doe",
  "phone": "+1234567890",
  "organization_id": "uuid-of-organization"
}
```

**Validation Rules:**
- `email`: Valid email format, required
- `password`: Minimum 8 characters, must contain uppercase, lowercase, number, and special character
- `first_name`: Required
- `last_name`: Required
- `organization_id`: Valid UUID, required
- `phone`: Optional

**Success Response (201 Created):**
```json
{
  "message": "User registered successfully. Please check your email for activation instructions."
}
```

**Notes:**
- User account is created with status `inactive`
- Activation email is sent with a 24-hour expiration token
- Email must be unique within the organization

---

### 2. Verify Email / Activate Account

Activate a user account using the verification token sent via email.

**Endpoint:** `POST /auth/verify`

**Request Body:**
```json
{
  "token": "activation-token-from-email"
}
```

**Query Parameter Alternative:**
```
GET /auth/verify?token=activation-token-from-email
```

**Success Response (200 OK):**
```json
{
  "message": "Account activated successfully. You can now log in."
}
```

**Notes:**
- Token expires after 24 hours
- Sets user status to `active`
- Sets `verified_at` timestamp
- Token can only be used once

---

### 3. Login

Authenticate user and receive access tokens.

**Endpoint:** `POST /auth/login`

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "SecureP@ss123"
}
```

**Success Response (200 OK):**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 86400,
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "status": "active",
    "organization_id": "org-uuid",
    "verified_at": "2025-01-15T10:30:00Z",
    "created_at": "2025-01-10T08:00:00Z"
  }
}
```

**Notes:**
- Refresh token is set as HTTP-only cookie
- Access token expires in 24 hours (configurable)
- Refresh token expires in 7 days (configurable)
- Only active users can log in

---

### 4. Refresh Token

Obtain a new access token using the refresh token.

**Endpoint:** `POST /auth/refresh`

**Request:**
- Refresh token can be provided via:
  1. HTTP-only cookie (automatic)
  2. Request body (if cookie not available)

**Request Body (Optional):**
```json
{
  "refresh_token": "refresh-token-string"
}
```

**Success Response (200 OK):**
```json
{
  "access_token": "new-access-token"
}
```

---

### 5. Logout

Revoke user's access and refresh tokens.

**Endpoint:** `POST /auth/logout`

**Headers:**
```
Authorization: Bearer <access-token>
```

**Success Response (200 OK):**
```json
{
  "message": "Logged out successfully"
}
```

**Notes:**
- Blacklists the access token
- Revokes all refresh tokens for the user
- Clears refresh token cookie

---

### 6. Request Password Reset

Initiate password reset flow by requesting a reset link via email.

**Endpoint:** `POST /auth/reset-password`

**Request Body:**
```json
{
  "email": "user@example.com"
}
```

**Headers:**
```
X-Organization-ID: <organization-uuid>
```

**Success Response (200 OK):**
```json
{
  "message": "If the email exists, a password reset link has been sent."
}
```

**Notes:**
- Returns success even if email doesn't exist (security best practice)
- Reset token expires in 1 hour
- Email contains secure reset link
- Previous reset tokens are invalidated

---

### 7. Confirm Password Reset

Complete password reset by submitting new password with reset token.

**Endpoint:** `POST /auth/confirm-reset`

**Request Body:**
```json
{
  "token": "reset-token-from-email",
  "new_password": "NewSecureP@ss456"
}
```

**Password Validation:**
- Minimum 8 characters
- At least one uppercase letter
- At least one lowercase letter
- At least one number
- At least one special character

**Success Response (200 OK):**
```json
{
  "message": "Password has been reset successfully. You can now log in with your new password."
}
```

**Notes:**
- Token must be valid and not expired (1 hour window)
- Token can only be used once
- Old password is replaced with new hashed password
- All existing sessions are invalidated

---

## User Management Endpoints

These endpoints require authentication and proper permissions.

### 8. Activate User (Admin)

Manually activate a user account (admin operation).

**Endpoint:** `PUT /users/:id/activate`

**Headers:**
```
Authorization: Bearer <admin-access-token>
```

**Required Permission:** `users:activate`

**URL Parameters:**
- `id`: User UUID to activate

**Success Response (200 OK):**
```json
{
  "message": "User activated successfully"
}
```

**Notes:**
- Sets user status to `active`
- Sets `verified_at` timestamp
- Publishes `user.activated` event
- Admin cannot be the same as target user (for deactivation)

---

### 9. Deactivate User (Admin)

Manually deactivate a user account (admin operation).

**Endpoint:** `PUT /users/:id/deactivate`

**Headers:**
```
Authorization: Bearer <admin-access-token>
```

**Required Permission:** `users:deactivate`

**URL Parameters:**
- `id`: User UUID to deactivate

**Success Response (200 OK):**
```json
{
  "message": "User deactivated successfully"
}
```

**Notes:**
- Sets user status to `inactive`
- Publishes `user.deactivated` event
- Admin cannot deactivate their own account
- Deactivated users cannot log in

---

## Error Responses

All endpoints follow a consistent error response format:

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error message"
  }
}
```

### Common HTTP Status Codes

| Status Code | Description |
|-------------|-------------|
| 200 | Success |
| 201 | Resource created successfully |
| 400 | Bad request - invalid input |
| 401 | Unauthorized - authentication required or invalid credentials |
| 403 | Forbidden - insufficient permissions |
| 404 | Resource not found |
| 500 | Internal server error |

### Error Examples

**400 Bad Request:**
```json
{
  "error": {
    "code": "BAD_REQUEST",
    "message": "password must be at least 8 characters long"
  }
}
```

**401 Unauthorized:**
```json
{
  "error": {
    "code": "UNAUTHORIZED",
    "message": "invalid or expired token"
  }
}
```

**403 Forbidden:**
```json
{
  "error": {
    "code": "FORBIDDEN",
    "message": "insufficient permissions to activate users"
  }
}
```

---

## Security

### Authentication

Most endpoints require a JWT access token in the Authorization header:

```
Authorization: Bearer <access-token>
```

### Password Requirements

- Minimum 8 characters
- At least one uppercase letter (A-Z)
- At least one lowercase letter (a-z)
- At least one number (0-9)
- At least one special character (!@#$%^&*(),.?":{}|<>)

### Token Security

- **Access Tokens**: Short-lived (24 hours), used for API authentication
- **Refresh Tokens**: Long-lived (7 days), stored as HTTP-only cookies
- **Reset Tokens**: Single-use, expire in 1 hour
- **Activation Tokens**: Single-use, expire in 24 hours

### HTTPS

**Production:** All endpoints must be accessed via HTTPS to protect sensitive data.

### Rate Limiting

API endpoints are rate-limited to prevent abuse:
- Registration: 5 requests per minute per IP
- Login: 5 requests per minute per email
- Password Reset: 3 requests per minute per IP

---

## Multi-Tenancy

The auth service supports multi-tenancy via organizations:

1. Users belong to organizations
2. Email uniqueness is scoped to organization
3. Admin permissions are organization-scoped
4. Organization ID is embedded in JWT tokens

**Organization Context:**
- Extracted from JWT token for authenticated requests
- Provided via `X-Organization-ID` header for public endpoints (like password reset)

---

## Email Templates

### Activation Email

Subject: "Activate Your Account"

Contains:
- Welcome message
- Activation link (expires in 24 hours)
- Fallback URL for manual copy-paste

### Password Reset Email

Subject: "Reset Your Password"

Contains:
- Reset instructions
- Reset link (expires in 1 hour)
- Security warning
- Fallback URL for manual copy-paste

### Welcome Email

Subject: "Welcome to GIIA"

Sent after successful account activation.

---

## Testing

### Example cURL Commands

**Register:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "SecureP@ss123",
    "first_name": "Test",
    "last_name": "User",
    "organization_id": "org-uuid"
  }'
```

**Login:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "SecureP@ss123"
  }'
```

**Verify Email:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/verify \
  -H "Content-Type: application/json" \
  -d '{
    "token": "activation-token"
  }'
```

**Request Password Reset:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/reset-password \
  -H "Content-Type: application/json" \
  -H "X-Organization-ID: org-uuid" \
  -d '{
    "email": "test@example.com"
  }'
```

**Confirm Password Reset:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/confirm-reset \
  -H "Content-Type: application/json" \
  -d '{
    "token": "reset-token",
    "new_password": "NewSecureP@ss456"
  }'
```

---

## Changelog

### Version 1.0 (2025-01-18)
- Added user registration with email verification
- Added password reset flow
- Added admin user activation/deactivation
- Added comprehensive error handling
- Added rate limiting support
- Added multi-tenancy support

---

For more information, see:
- [RBAC Design](RBAC_DESIGN.md)
- [gRPC API Documentation](README_GRPC.md)
- [Testing Strategy](TESTING_STRATEGY.md)
