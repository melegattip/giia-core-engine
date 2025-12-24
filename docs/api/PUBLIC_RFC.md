# GIIA Platform - Public API Reference (RFC)

**Version**: 1.0  
**Last Updated**: 2025-12-23  
**Status**: Draft  
**Base URL (Dev)**: `http://localhost:808x`  
**Base URL (Prod)**: `https://api.giia.io`

---

## 📖 Table of Contents

1. [Overview](#overview)
2. [Authentication](#authentication)
3. [Common Patterns](#common-patterns)
4. [Auth Service API](#auth-service-api)
5. [Catalog Service API](#catalog-service-api)
6. [Error Codes](#error-codes)
7. [Rate Limiting](#rate-limiting)
8. [Versioning](#versioning)

---

## Overview

The GIIA Platform exposes RESTful APIs for external consumption and gRPC for internal service-to-service communication. This document covers the public REST API endpoints.

### API Conventions

- **Content-Type**: `application/json`
- **Authentication**: Bearer JWT tokens
- **Multi-tenancy**: `organization_id` required for tenant-scoped resources
- **Pagination**: Query params `page` (default: 1) and `page_size` (default: 20, max: 100)
- **Date format**: ISO 8601 (`2025-12-23T12:00:00Z`)

---

## Authentication

### Token Flow

```
1. Register → Create user account (optional email verification)
2. Login → Receive access_token and refresh_token
3. API Request → Include access_token in Authorization header
4. Token Expires → Use refresh_token to get new access_token
5. Logout → Invalidate tokens
```

### Token Types

| Token | Lifetime | Storage | Purpose |
|-------|----------|---------|---------|
| Access Token | 15 minutes | Memory | Authorize API requests |
| Refresh Token | 7 days | HTTP-only cookie | Obtain new access tokens |

### Using Tokens

```http
Authorization: Bearer <access_token>
```

**Example:**
```bash
curl -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..." \
     https://api.giia.io/api/v1/products
```

---

## Common Patterns

### Request Headers

```http
Content-Type: application/json
Accept: application/json
Authorization: Bearer <access_token>
```

### Pagination

**Request:**
```http
GET /api/v1/products?page=1&page_size=20
```

**Response:**
```json
{
  "data": [...],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total": 150,
    "total_pages": 8
  }
}
```

### Filtering & Sorting

```http
GET /api/v1/products?category=electronics&status=active&sort=name&order=asc
```

### Multi-Tenancy

All tenant-scoped resources require `organization_id`:

```http
GET /api/v1/products?organization_id=550e8400-e29b-41d4-a716-446655440000
```

---

## Auth Service API

**Base URL**: `/api/v1/auth`

### POST /auth/register

Create a new user account.

**Authentication**: Not required

**Request:**
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
- `email`: Valid email format, unique per organization
- `password`: Min 8 chars, uppercase, lowercase, digit, special char
- `first_name`, `last_name`: Required, max 100 chars
- `organization_id`: Required, valid UUID

**Response (201 Created):**
```json
{
  "message": "User registered successfully. Check email for activation."
}
```

**Rate Limit**: 3 attempts per 60 minutes per IP

---

### POST /auth/login

Authenticate and receive JWT tokens.

**Authentication**: Not required

**Request:**
```json
{
  "email": "user@example.com",
  "password": "SecurePassword123!"
}
```

**Response (200 OK):**
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

**Note**: `refresh_token` is set as HTTP-only cookie.

**Rate Limit**: 5 attempts per 15 minutes per IP

---

### POST /auth/refresh

Obtain new access token using refresh token.

**Authentication**: Refresh token (HTTP-only cookie)

**Response (200 OK):**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "expires_in": 900
}
```

---

### POST /auth/logout

Invalidate current session.

**Authentication**: Required (Bearer token)

**Response (200 OK):**
```json
{
  "message": "Logged out successfully"
}
```

---

### POST /auth/activate

Activate user account with email token.

**Authentication**: Not required

**Request:**
```json
{
  "token": "activation-token-from-email"
}
```

**Response (200 OK):**
```json
{
  "message": "Account activated successfully."
}
```

---

## Catalog Service API

**Base URL**: `/api/v1`

### Products

#### POST /products

Create a new product.

**Authentication**: Required

**Request:**
```json
{
  "organization_id": "550e8400-e29b-41d4-a716-446655440000",
  "sku": "PROD-001",
  "name": "Laptop Computer",
  "description": "High-performance business laptop",
  "category": "Electronics",
  "unit_of_measure": "UNIT",
  "buffer_profile_id": "650e8400-e29b-41d4-a716-446655440001"
}
```

**Response (201 Created):**
```json
{
  "product": {
    "id": "760e8400-e29b-41d4-a716-446655440002",
    "organization_id": "550e8400-e29b-41d4-a716-446655440000",
    "sku": "PROD-001",
    "name": "Laptop Computer",
    "description": "High-performance business laptop",
    "category": "Electronics",
    "unit_of_measure": "UNIT",
    "status": "active",
    "buffer_profile_id": "650e8400-e29b-41d4-a716-446655440001",
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  }
}
```

---

#### GET /products/{id}

Get product by ID.

**Authentication**: Required

**Query Parameters:**
- `organization_id` (required): UUID of the organization

**Response (200 OK):**
```json
{
  "product": {
    "id": "760e8400-e29b-41d4-a716-446655440002",
    "organization_id": "550e8400-e29b-41d4-a716-446655440000",
    "sku": "PROD-001",
    "name": "Laptop Computer",
    "status": "active",
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  }
}
```

---

#### GET /products

List products with pagination and filtering.

**Authentication**: Required

**Query Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `organization_id` | UUID | Yes | Organization filter |
| `page` | int | No | Page number (default: 1) |
| `page_size` | int | No | Items per page (default: 20, max: 100) |
| `status` | string | No | Filter: active, inactive, discontinued |
| `category` | string | No | Filter by category |

**Response (200 OK):**
```json
{
  "products": [...],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total": 150,
    "total_pages": 8
  }
}
```

---

#### PUT /products/{id}

Update an existing product.

**Authentication**: Required

**Request:**
```json
{
  "id": "760e8400-e29b-41d4-a716-446655440002",
  "organization_id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Updated Laptop Computer",
  "description": "Updated description"
}
```

**Note**: SKU cannot be updated.

**Response (200 OK):**
```json
{
  "product": { ... }
}
```

---

#### DELETE /products/{id}

Soft delete a product (marks as inactive).

**Authentication**: Required

**Query Parameters:**
- `organization_id` (required): UUID of the organization

**Response (200 OK):**
```json
{
  "success": true
}
```

---

#### GET /products/search

Full-text search across product fields.

**Authentication**: Required

**Query Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `organization_id` | UUID | Yes | Organization filter |
| `query` | string | Yes | Search query |
| `page` | int | No | Page number |
| `page_size` | int | No | Items per page |

**Searches**: SKU, Name, Description, Category

**Response (200 OK):**
```json
{
  "products": [...],
  "total": 5
}
```

---

## Error Codes

### Error Response Format

```json
{
  "error_code": "ERROR_CODE",
  "message": "Human-readable error message",
  "details": "Additional context"
}
```

### Standard Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `BAD_REQUEST` | 400 | Invalid request parameters |
| `UNAUTHORIZED` | 401 | Missing or invalid authentication |
| `FORBIDDEN` | 403 | Insufficient permissions |
| `NOT_FOUND` | 404 | Resource not found |
| `CONFLICT` | 409 | Resource conflict (e.g., duplicate) |
| `UNPROCESSABLE_ENTITY` | 422 | Validation error |
| `TOO_MANY_REQUESTS` | 429 | Rate limit exceeded |
| `INTERNAL_SERVER_ERROR` | 500 | Server error |
| `SERVICE_UNAVAILABLE` | 503 | Service temporarily unavailable |

---

## Rate Limiting

### Limits

| Endpoint Type | Limit |
|---------------|-------|
| Authentication | 10 requests/minute/IP |
| API endpoints | 100 requests/minute/user |

### Headers

```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1642248000
```

### Exceeded Response (429)

```json
{
  "error_code": "TOO_MANY_REQUESTS",
  "message": "Rate limit exceeded. Please try again later.",
  "retry_after": 60
}
```

---

## Versioning

### Current Version

**v1** - Current stable version

### Version Strategy

- API versions in URL path: `/api/v1/...`
- Breaking changes = new major version
- Deprecated versions supported for 12 months

### Deprecation Notice

Deprecated endpoints include header:
```http
X-API-Deprecation: true
X-API-Sunset-Date: 2026-01-01
```

---

## gRPC Services

For internal service-to-service communication, see [gRPC Contracts](./GRPC_CONTRACTS.md).

### Available Services

| Service | Port | Proto File |
|---------|------|------------|
| Auth Service | 9081 | `api/proto/auth/v1/auth.proto` |
| Catalog Service | 9082 | `api/proto/catalog/v1/*.proto` |
| DDMRP Engine | 9083 | `api/proto/ddmrp/v1/*.proto` |
| Execution Service | 9084 | `api/proto/execution/v1/*.proto` |
| Analytics Service | 9085 | `api/proto/analytics/v1/*.proto` |
| AI Service | 9086 | `api/proto/ai/v1/*.proto` |

---

## SDKs & Examples

### cURL Examples

**Login:**
```bash
curl -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "SecurePass123!"}'
```

**List Products:**
```bash
curl "http://localhost:8082/api/v1/products?organization_id=UUID&page=1" \
  -H "Authorization: Bearer $TOKEN"
```

---

## Changelog

| Version | Date | Changes |
|---------|------|---------|
| 1.0 | 2025-12-23 | Initial public API specification |

---

**API maintained by the GIIA Team** 📡
