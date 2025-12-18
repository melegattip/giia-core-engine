# GIIA Platform API Documentation

## Overview

This document provides comprehensive API documentation for the GIIA (Global Integrated Inventory Analytics) platform Phase 1 services:

- **Auth Service**: Authentication, authorization, and user management
- **Catalog Service**: Product catalog management with DDMRP buffer profiles

**Base URLs:**
- Auth Service: `http://auth.giia.local` (Production: `https://auth.giia.com`)
- Catalog Service: `http://catalog.giia.local` (Production: `https://catalog.giia.com`)

**Version:** v1

---

## Table of Contents

1. [Authentication](#authentication)
2. [Common Patterns](#common-patterns)
3. [Auth Service API](#auth-service-api)
4. [Catalog Service API](#catalog-service-api)
5. [Error Codes](#error-codes)
6. [Rate Limiting](#rate-limiting)

---

## Authentication

### Overview

The GIIA platform uses **JWT (JSON Web Token)** based authentication. All protected endpoints require a valid Bearer token in the Authorization header.

### Authentication Flow

```
1. Register → Create user account
2. Login → Receive access_token and refresh_token
3. API Request → Include access_token in Authorization header
4. Token Expires → Use refresh_token to get new access_token
```

### Token Types

#### Access Token
- **Lifetime**: 15 minutes
- **Purpose**: Authorize API requests
- **Storage**: Memory (not localStorage)

#### Refresh Token
- **Lifetime**: 7 days
- **Purpose**: Obtain new access tokens
- **Storage**: HTTP-only secure cookie

### Using Tokens

Include the access token in the Authorization header:

```http
Authorization: Bearer <access_token>
```

**Example:**
```bash
curl -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..." \
     https://catalog.giia.local/api/v1/products
```

---

## Common Patterns

### Request Headers

All requests should include:

```http
Content-Type: application/json
Accept: application/json
Authorization: Bearer <access_token>  # For protected endpoints
```

### Pagination

List endpoints support pagination with query parameters:

```http
GET /api/v1/products?page=1&page_size=20
```

**Parameters:**
- `page`: Page number (default: 1, min: 1)
- `page_size`: Items per page (default: 20, min: 1, max: 100)

**Response:**
```json
{
  "products": [...],
  "page": 1,
  "page_size": 20,
  "total": 150,
  "total_pages": 8
}
```

### Filtering

Filter results using query parameters:

```http
GET /api/v1/products?category=electronics&status=active
```

### Searching

Search endpoints accept a query parameter:

```http
GET /api/v1/products/search?query=laptop&page=1&page_size=20
```

### Multi-Tenancy

All requests must include `organization_id`:

**Query Parameter:**
```http
GET /api/v1/products?organization_id=550e8400-e29b-41d4-a716-446655440000
```

**Request Body:**
```json
{
  "organization_id": "550e8400-e29b-41d4-a716-446655440000",
  ...
}
```

---

## Auth Service API

### Base URL

```
http://auth.giia.local/api/v1
```

---

### Register User

Create a new user account.

**Endpoint:** `POST /auth/register`

**Authentication:** Not required

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

**Validation Rules:**
- `email`: Valid email format, unique
- `password`: Min 8 characters, must include uppercase, lowercase, and number
- `first_name`: Required, max 100 characters
- `last_name`: Required, max 100 characters
- `phone`: Optional, E.164 format
- `organization_id`: Required, valid UUID

**Success Response:** `201 Created`
```json
{
  "message": "User registered successfully. Please check your email to activate your account."
}
```

**Error Responses:**

`400 Bad Request` - Validation error
```json
{
  "error_code": "BAD_REQUEST",
  "message": "Validation failed",
  "details": "Invalid email format"
}
```

`409 Conflict` - User already exists
```json
{
  "error_code": "CONFLICT",
  "message": "User with this email already exists"
}
```

**Example:**
```bash
curl -X POST http://auth.giia.local/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePassword123!",
    "first_name": "John",
    "last_name": "Doe",
    "organization_id": "550e8400-e29b-41d4-a716-446655440000"
  }'
```

---

### Login

Authenticate a user and receive JWT tokens.

**Endpoint:** `POST /auth/login`

**Authentication:** Not required

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "SecurePassword123!"
}
```

**Success Response:** `200 OK`
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
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

**Note:** The `refresh_token` is returned as an HTTP-only cookie.

**Error Responses:**

`401 Unauthorized` - Invalid credentials
```json
{
  "error_code": "UNAUTHORIZED",
  "message": "Invalid email or password"
}
```

`403 Forbidden` - Account not activated
```json
{
  "error_code": "FORBIDDEN",
  "message": "Account is not activated. Please check your email."
}
```

**Example:**
```bash
curl -X POST http://auth.giia.local/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePassword123!"
  }'
```

---

### Refresh Token

Obtain a new access token using a refresh token.

**Endpoint:** `POST /auth/refresh`

**Authentication:** Refresh token (HTTP-only cookie)

**Request Body:** None (refresh token sent via cookie)

**Success Response:** `200 OK`
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 900
}
```

**Error Responses:**

`401 Unauthorized` - Invalid or expired refresh token
```json
{
  "error_code": "UNAUTHORIZED",
  "message": "Invalid or expired refresh token"
}
```

**Example:**
```bash
curl -X POST http://auth.giia.local/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  --cookie "refresh_token=..."
```

---

### Logout

Invalidate the current session and refresh token.

**Endpoint:** `POST /auth/logout`

**Authentication:** Required (Bearer token)

**Request Body:** None

**Success Response:** `200 OK`
```json
{
  "message": "Logged out successfully"
}
```

**Example:**
```bash
curl -X POST http://auth.giia.local/api/v1/auth/logout \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

---

### Activate Account

Activate a user account using the activation token sent via email.

**Endpoint:** `POST /auth/activate`

**Authentication:** Not required

**Request Body:**
```json
{
  "token": "activation_token_from_email"
}
```

**Success Response:** `200 OK`
```json
{
  "message": "Account activated successfully. You can now log in."
}
```

**Error Responses:**

`400 Bad Request` - Invalid or expired token
```json
{
  "error_code": "BAD_REQUEST",
  "message": "Invalid or expired activation token"
}
```

---

## Catalog Service API

### Base URL

```
http://catalog.giia.local/api/v1
```

**Note:** All Catalog Service endpoints require authentication (Bearer token).

---

### Create Product

Create a new product in the catalog.

**Endpoint:** `POST /products`

**Authentication:** Required

**Request Body:**
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

**Field Descriptions:**
- `organization_id`: Required, UUID of the organization
- `sku`: Required, unique within organization, max 100 characters
- `name`: Required, product name, max 255 characters
- `description`: Optional, product description
- `category`: Optional, product category, max 100 characters
- `unit_of_measure`: Required, unit of measure (e.g., UNIT, KG, LITER)
- `buffer_profile_id`: Optional, UUID of associated DDMRP buffer profile

**Success Response:** `201 Created`
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

**Error Responses:**

`400 Bad Request` - Validation error
```json
{
  "error_code": "BAD_REQUEST",
  "message": "Validation failed",
  "details": "SKU is required"
}
```

`409 Conflict` - SKU already exists
```json
{
  "error_code": "CONFLICT",
  "message": "Product with SKU 'PROD-001' already exists in this organization"
}
```

**Example:**
```bash
curl -X POST http://catalog.giia.local/api/v1/products \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "organization_id": "550e8400-e29b-41d4-a716-446655440000",
    "sku": "PROD-001",
    "name": "Laptop Computer",
    "unit_of_measure": "UNIT"
  }'
```

---

### Get Product

Retrieve a single product by ID.

**Endpoint:** `GET /products/{product_id}`

**Authentication:** Required

**Path Parameters:**
- `product_id`: UUID of the product

**Query Parameters:**
- `organization_id`: Required, UUID of the organization

**Success Response:** `200 OK`
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

**Error Responses:**

`404 Not Found` - Product not found
```json
{
  "error_code": "NOT_FOUND",
  "message": "Product not found"
}
```

**Example:**
```bash
curl http://catalog.giia.local/api/v1/products/760e8400-e29b-41d4-a716-446655440002?organization_id=550e8400-e29b-41d4-a716-446655440000 \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

---

### Update Product

Update an existing product.

**Endpoint:** `PUT /products/{product_id}`

**Authentication:** Required

**Path Parameters:**
- `product_id`: UUID of the product

**Request Body:**
```json
{
  "id": "760e8400-e29b-41d4-a716-446655440002",
  "organization_id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Updated Laptop Computer",
  "description": "Updated description",
  "category": "Updated Electronics",
  "unit_of_measure": "UNIT",
  "status": "active",
  "buffer_profile_id": "650e8400-e29b-41d4-a716-446655440001"
}
```

**Note:** SKU cannot be updated. All other fields are optional.

**Success Response:** `200 OK`
```json
{
  "product": {
    "id": "760e8400-e29b-41d4-a716-446655440002",
    "organization_id": "550e8400-e29b-41d4-a716-446655440000",
    "sku": "PROD-001",
    "name": "Updated Laptop Computer",
    "description": "Updated description",
    "category": "Updated Electronics",
    "unit_of_measure": "UNIT",
    "status": "active",
    "buffer_profile_id": "650e8400-e29b-41d4-a716-446655440001",
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T11:45:00Z"
  }
}
```

**Error Responses:**

`404 Not Found` - Product not found
```json
{
  "error_code": "NOT_FOUND",
  "message": "Product not found"
}
```

**Example:**
```bash
curl -X PUT http://catalog.giia.local/api/v1/products/760e8400-e29b-41d4-a716-446655440002 \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "id": "760e8400-e29b-41d4-a716-446655440002",
    "organization_id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "Updated Laptop Computer"
  }'
```

---

### List Products

Retrieve a paginated list of products.

**Endpoint:** `GET /products`

**Authentication:** Required

**Query Parameters:**
- `organization_id`: Required, UUID of the organization
- `page`: Optional, page number (default: 1)
- `page_size`: Optional, items per page (default: 20, max: 100)
- `status`: Optional, filter by status (active, inactive, discontinued)
- `category`: Optional, filter by category

**Success Response:** `200 OK`
```json
{
  "products": [
    {
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
  ],
  "total": 150,
  "page": 1,
  "page_size": 20
}
```

**Example:**
```bash
curl "http://catalog.giia.local/api/v1/products?organization_id=550e8400-e29b-41d4-a716-446655440000&page=1&page_size=20&status=active" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

---

### Search Products

Full-text search across product fields.

**Endpoint:** `GET /products/search`

**Authentication:** Required

**Query Parameters:**
- `organization_id`: Required, UUID of the organization
- `query`: Required, search query string
- `page`: Optional, page number (default: 1)
- `page_size`: Optional, items per page (default: 20, max: 100)

**Search Fields:**
- SKU
- Name
- Description
- Category

**Success Response:** `200 OK`
```json
{
  "products": [
    {
      "id": "760e8400-e29b-41d4-a716-446655440002",
      "organization_id": "550e8400-e29b-41d4-a716-446655440000",
      "sku": "PROD-001",
      "name": "Laptop Computer",
      "description": "High-performance business laptop",
      "category": "Electronics",
      "unit_of_measure": "UNIT",
      "status": "active",
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T10:30:00Z"
    }
  ],
  "total": 5
}
```

**Example:**
```bash
curl "http://catalog.giia.local/api/v1/products/search?organization_id=550e8400-e29b-41d4-a716-446655440000&query=laptop&page=1" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

---

### Delete Product

Soft delete a product (marks as inactive).

**Endpoint:** `DELETE /products/{product_id}`

**Authentication:** Required

**Path Parameters:**
- `product_id`: UUID of the product

**Query Parameters:**
- `organization_id`: Required, UUID of the organization

**Success Response:** `200 OK`
```json
{
  "success": true
}
```

**Error Responses:**

`404 Not Found` - Product not found
```json
{
  "error_code": "NOT_FOUND",
  "message": "Product not found"
}
```

**Example:**
```bash
curl -X DELETE "http://catalog.giia.local/api/v1/products/760e8400-e29b-41d4-a716-446655440002?organization_id=550e8400-e29b-41d4-a716-446655440000" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

---

## Error Codes

### Standard Error Response Format

```json
{
  "error_code": "ERROR_CODE",
  "message": "Human-readable error message",
  "details": "Additional context or validation errors"
}
```

### Common Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `BAD_REQUEST` | 400 | Invalid request parameters or body |
| `UNAUTHORIZED` | 401 | Missing or invalid authentication |
| `FORBIDDEN` | 403 | Insufficient permissions |
| `NOT_FOUND` | 404 | Resource not found |
| `CONFLICT` | 409 | Resource conflict (e.g., duplicate SKU) |
| `UNPROCESSABLE_ENTITY` | 422 | Validation error |
| `TOO_MANY_REQUESTS` | 429 | Rate limit exceeded |
| `INTERNAL_SERVER_ERROR` | 500 | Internal server error |
| `SERVICE_UNAVAILABLE` | 503 | Service temporarily unavailable |

---

## Rate Limiting

### Limits

- **Authentication endpoints**: 10 requests per minute per IP
- **API endpoints**: 100 requests per minute per user

### Rate Limit Headers

All responses include rate limit information:

```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1642248000
```

### Rate Limit Exceeded Response

```http
HTTP/1.1 429 Too Many Requests
```

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

- API version is included in the URL path: `/api/v1/...`
- Breaking changes will increment the major version: `/api/v2/...`
- Minor updates and bug fixes maintain backward compatibility

---

## Support

For questions or issues:
- Check service logs: `make k8s-logs SERVICE=<service-name>`
- Review integration tests: `tests/integration/`
- Open an issue in the project repository

---

## Changelog

### Version 1.0.0 (2024-01-15)

- Initial API release
- Auth Service: User registration, login, token management
- Catalog Service: Product CRUD operations
- JWT-based authentication
- Multi-tenancy support
- Rate limiting

---

## Additional Resources

- [Integration Tests](../../tests/integration/README.md)
- [Kubernetes Deployment Guide](./README_KUBERNETES.md)
- [Project Documentation](./PROJECT_STATUS.md)
