# Error Handling Guide

This document provides comprehensive guidelines for error handling in the GIIA Core Engine project.

## Table of Contents

1. [Error Types](#error-types)
2. [Error Constructors](#error-constructors)
3. [GORM Error Mapping](#gorm-error-mapping)
4. [Layer-Specific Guidelines](#layer-specific-guidelines)
5. [Testing Error Handling](#testing-error-handling)
6. [Best Practices](#best-practices)

---

## Error Types

All errors in this project use the typed error system from `pkg/errors`. The base type is `CustomError`:

```go
type CustomError struct {
    ErrorCode  string
    Message    string
    HTTPStatus int
    Cause      error
}
```

### Available Error Codes

| Error Code | HTTP Status | Use Case |
|-----------|-------------|----------|
| `BAD_REQUEST` | 400 | Invalid input, validation failures |
| `UNAUTHORIZED` | 401 | Authentication failures |
| `FORBIDDEN` | 403 | Authorization failures, insufficient permissions |
| `NOT_FOUND` | 404 | Resource not found |
| `CONFLICT` | 409 | Duplicate resource, constraint violations |
| `UNPROCESSABLE_ENTITY` | 422 | Semantic validation errors |
| `TOO_MANY_REQUESTS` | 429 | Rate limit violations |
| `INTERNAL_SERVER_ERROR` | 500 | Infrastructure failures, unexpected errors |
| `SERVICE_UNAVAILABLE` | 503 | Service temporarily unavailable |

---

## Error Constructors

### Creating Typed Errors

**✅ ALWAYS use typed error constructors:**

```go
import pkgErrors "github.com/giia/giia-core-engine/pkg/errors"

// Validation error
if email == "" {
    return pkgErrors.NewBadRequest("email is required")
}

// Authentication error
if !validPassword {
    return pkgErrors.NewUnauthorized("invalid credentials")
}

// Authorization error
if !hasPermission {
    return pkgErrors.NewForbidden("insufficient permissions")
}

// Resource not found
if user == nil {
    return pkgErrors.NewNotFound("user not found")
}

// Duplicate resource
if existingUser != nil {
    return pkgErrors.NewConflict("user with this email already exists")
}

// Infrastructure error
if dbErr != nil {
    return pkgErrors.NewInternalServerError("database operation failed")
}
```

**❌ NEVER use `fmt.Errorf`:**

```go
// BAD - Don't do this
return fmt.Errorf("user not found")
return fmt.Errorf("invalid email: %w", err)
```

---

## GORM Error Mapping

When working with GORM in repositories, map database errors to appropriate typed errors:

### Standard GORM Error Mapping

```go
import (
    "errors"
    "gorm.io/gorm"
    pkgErrors "github.com/giia/giia-core-engine/pkg/errors"
)

func (r *UserRepository) FindByID(ctx context.Context, id int64) (*domain.User, error) {
    var user domain.User
    err := r.db.WithContext(ctx).First(&user, id).Error

    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, pkgErrors.NewNotFound("user not found")
        }
        return nil, pkgErrors.NewInternalServerError("failed to query user")
    }

    return &user, nil
}
```

### GORM Error Mapping Table

| GORM Error | Typed Error | Use Case |
|-----------|-------------|----------|
| `gorm.ErrRecordNotFound` | `NewNotFound()` | Resource doesn't exist |
| `gorm.ErrDuplicatedKey` | `NewConflict()` | Unique constraint violation |
| `gorm.ErrForeignKeyViolated` | `NewBadRequest()` | Invalid foreign key reference |
| `gorm.ErrInvalidData` | `NewBadRequest()` | Invalid data format |
| Other database errors | `NewInternalServerError()` | Infrastructure failures |

### Complete Repository Error Handling Example

```go
func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
    err := r.db.WithContext(ctx).Create(user).Error

    if err != nil {
        // Check for duplicate key (unique constraint)
        if errors.Is(err, gorm.ErrDuplicatedKey) {
            return pkgErrors.NewConflict("user with this email already exists")
        }

        // Check for foreign key violation
        if errors.Is(err, gorm.ErrForeignKeyViolated) {
            return pkgErrors.NewBadRequest("invalid reference to related entity")
        }

        // All other errors are infrastructure failures
        return pkgErrors.NewInternalServerError("failed to create user")
    }

    return nil
}

func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
    result := r.db.WithContext(ctx).Save(user)

    if result.Error != nil {
        if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
            return pkgErrors.NewConflict("user with this email already exists")
        }
        return pkgErrors.NewInternalServerError("failed to update user")
    }

    if result.RowsAffected == 0 {
        return pkgErrors.NewNotFound("user not found")
    }

    return nil
}

func (r *UserRepository) Delete(ctx context.Context, id int64) error {
    result := r.db.WithContext(ctx).Delete(&domain.User{}, id)

    if result.Error != nil {
        return pkgErrors.NewInternalServerError("failed to delete user")
    }

    if result.RowsAffected == 0 {
        return pkgErrors.NewNotFound("user not found")
    }

    return nil
}
```

---

## Layer-Specific Guidelines

### Repository Layer

**Responsibilities:**
- Map GORM errors to typed errors
- Return `NewNotFound()` for missing resources
- Return `NewConflict()` for constraint violations
- Return `NewInternalServerError()` for database failures

**Example:**

```go
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
    var user domain.User
    err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error

    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, pkgErrors.NewNotFound("user not found")
        }
        return nil, pkgErrors.NewInternalServerError("failed to query user")
    }

    return &user, nil
}
```

### Use Case Layer

**Responsibilities:**
- Validate input parameters (return `NewBadRequest()`)
- Handle business logic errors
- Preserve error types from repository layer
- Return `NewUnauthorized()` for auth failures
- Return `NewForbidden()` for permission failures

**Example:**

```go
func (uc *LoginUseCase) Execute(ctx context.Context, email, password string) (*LoginResponse, error) {
    // Validation
    if email == "" {
        return nil, pkgErrors.NewBadRequest("email is required")
    }
    if password == "" {
        return nil, pkgErrors.NewBadRequest("password is required")
    }

    // Repository call - preserve error type
    user, err := uc.userRepo.FindByEmail(ctx, email)
    if err != nil {
        // If it's a NotFound error, convert to Unauthorized for security
        var notFoundErr *pkgErrors.CustomError
        if errors.As(err, &notFoundErr) && notFoundErr.ErrorCode == pkgErrors.CodeNotFound {
            return nil, pkgErrors.NewUnauthorized("invalid credentials")
        }
        return nil, err // Preserve other error types
    }

    // Business logic
    if !uc.passwordService.Verify(password, user.PasswordHash) {
        return nil, pkgErrors.NewUnauthorized("invalid credentials")
    }

    return &LoginResponse{User: user}, nil
}
```

### Infrastructure Adapters

**Responsibilities:**
- Map adapter-specific errors to typed errors
- Return `NewInternalServerError()` for infrastructure failures
- Return `NewUnauthorized()` for JWT validation failures

**Example - JWT Manager:**

```go
func (j *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        return j.secretKey, nil
    })

    if err != nil {
        if errors.Is(err, jwt.ErrTokenExpired) {
            return nil, pkgErrors.NewUnauthorized("token expired")
        }
        if errors.Is(err, jwt.ErrTokenMalformed) {
            return nil, pkgErrors.NewUnauthorized("malformed token")
        }
        return nil, pkgErrors.NewUnauthorized("invalid token")
    }

    claims, ok := token.Claims.(*Claims)
    if !ok || !token.Valid {
        return nil, pkgErrors.NewUnauthorized("invalid token claims")
    }

    return claims, nil
}
```

**Example - Cache Adapter:**

```go
func (c *RedisPermissionCache) Get(ctx context.Context, key string) ([]byte, error) {
    val, err := c.client.Get(ctx, key).Bytes()

    if err != nil {
        if errors.Is(err, redis.Nil) {
            return nil, pkgErrors.NewNotFound("cache key not found")
        }
        return nil, pkgErrors.NewInternalServerError("cache operation failed")
    }

    return val, nil
}
```

---

## Testing Error Handling

### Verifying Error Types

**Use `errors.As()` to verify error types in tests:**

```go
func TestLoginUseCase_InvalidEmail_ReturnsBadRequest(t *testing.T) {
    // Given
    uc := NewLoginUseCase(mockRepo, mockPasswordService)

    // When
    _, err := uc.Execute(context.Background(), "", "password123")

    // Then
    var customErr *pkgErrors.CustomError
    if !errors.As(err, &customErr) {
        t.Fatal("expected CustomError")
    }

    if customErr.ErrorCode != pkgErrors.CodeBadRequest {
        t.Errorf("expected BAD_REQUEST, got %s", customErr.ErrorCode)
    }

    if customErr.HTTPStatus != http.StatusBadRequest {
        t.Errorf("expected status 400, got %d", customErr.HTTPStatus)
    }
}
```

### Testing Repository Error Mapping

```go
func TestUserRepository_FindByID_NotFound_ReturnsNotFoundError(t *testing.T) {
    // Given
    db, mock, _ := sqlmock.New()
    repo := NewUserRepository(db)

    mock.ExpectQuery("SELECT .* FROM users").
        WillReturnError(gorm.ErrRecordNotFound)

    // When
    _, err := repo.FindByID(context.Background(), 123)

    // Then
    var customErr *pkgErrors.CustomError
    if !errors.As(err, &customErr) {
        t.Fatal("expected CustomError")
    }

    if customErr.ErrorCode != pkgErrors.CodeNotFound {
        t.Errorf("expected NOT_FOUND, got %s", customErr.ErrorCode)
    }
}
```

---

## Best Practices

### 1. Never Use `fmt.Errorf`

```go
// ❌ BAD
return fmt.Errorf("user not found")
return fmt.Errorf("failed to create user: %w", err)

// ✅ GOOD
return pkgErrors.NewNotFound("user not found")
return pkgErrors.NewInternalServerError("failed to create user")
```

### 2. Preserve Error Types

```go
// ❌ BAD - Wrapping loses type information
user, err := uc.userRepo.FindByID(ctx, id)
if err != nil {
    return fmt.Errorf("failed to find user: %w", err)
}

// ✅ GOOD - Preserve the typed error
user, err := uc.userRepo.FindByID(ctx, id)
if err != nil {
    return err // Already a typed error from repository
}
```

### 3. Use Specific Error Messages

```go
// ❌ BAD - Generic, unhelpful
return pkgErrors.NewBadRequest("invalid input")

// ✅ GOOD - Specific, actionable
return pkgErrors.NewBadRequest("email must be a valid email address")
return pkgErrors.NewBadRequest("password must be at least 8 characters")
```

### 4. Don't Expose Sensitive Information

```go
// ❌ BAD - Exposes database details
return pkgErrors.NewInternalServerError(fmt.Sprintf("database error: %v", err))

// ✅ GOOD - Generic message, log details separately
logger.Error(ctx, err, "database query failed", logs.Tags{"table": "users"})
return pkgErrors.NewInternalServerError("failed to query user")
```

### 5. Convert Security-Sensitive Errors

```go
// ✅ GOOD - Don't reveal if user exists
user, err := uc.userRepo.FindByEmail(ctx, email)
if err != nil {
    // Convert NotFound to Unauthorized for security
    var customErr *pkgErrors.CustomError
    if errors.As(err, &customErr) && customErr.ErrorCode == pkgErrors.CodeNotFound {
        return nil, pkgErrors.NewUnauthorized("invalid credentials")
    }
    return nil, err
}

if !uc.passwordService.Verify(password, user.PasswordHash) {
    // Same error message - don't reveal which part failed
    return nil, pkgErrors.NewUnauthorized("invalid credentials")
}
```

### 6. Validate Early

```go
func (uc *CreateUserUseCase) Execute(ctx context.Context, req *CreateUserRequest) error {
    // Validate all inputs first
    if req.Email == "" {
        return pkgErrors.NewBadRequest("email is required")
    }
    if req.Password == "" {
        return pkgErrors.NewBadRequest("password is required")
    }
    if len(req.Password) < 8 {
        return pkgErrors.NewBadRequest("password must be at least 8 characters")
    }

    // Then proceed with business logic
    // ...
}
```

### 7. Log Errors with Context

```go
func (uc *UseCase) Execute(ctx context.Context, userID int64) error {
    user, err := uc.repo.FindByID(ctx, userID)
    if err != nil {
        logger.Error(ctx, err, "Failed to find user", logs.Tags{
            "user_id": userID,
            "operation": "find_user",
        })
        return err
    }

    return nil
}
```

---

## Error Response Format

HTTP error responses follow this structure:

```json
{
    "status_code": 400,
    "error_code": "BAD_REQUEST",
    "message": "email is required"
}
```

The `ToHTTPResponse()` function automatically converts typed errors:

```go
func handleError(w http.ResponseWriter, err error) {
    response := pkgErrors.ToHTTPResponse(err)

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(response.StatusCode)
    json.NewEncoder(w).Encode(response)
}
```

---

## Migration Checklist

When migrating code to use typed errors:

- [ ] Replace all `fmt.Errorf` with typed error constructors
- [ ] Map GORM errors to appropriate typed errors
- [ ] Validate inputs and return `NewBadRequest()`
- [ ] Preserve error types when propagating errors
- [ ] Add error type assertions in tests
- [ ] Verify error messages are specific and actionable
- [ ] Ensure sensitive information is not exposed
- [ ] Log errors with appropriate context
- [ ] Update tests to verify error types
- [ ] Run linters to catch any remaining `fmt.Errorf` usage

---

## Quick Reference

| Scenario | Error Constructor |
|---------|------------------|
| Missing required field | `NewBadRequest("field is required")` |
| Invalid format | `NewBadRequest("invalid email format")` |
| Authentication failed | `NewUnauthorized("invalid credentials")` |
| Missing/invalid token | `NewUnauthorized("token expired")` |
| Insufficient permissions | `NewForbidden("insufficient permissions")` |
| Resource not found | `NewNotFound("resource not found")` |
| Duplicate resource | `NewConflict("resource already exists")` |
| Rate limit | `NewTooManyRequests("rate limit exceeded")` |
| Database error | `NewInternalServerError("database operation failed")` |
| External service error | `NewInternalServerError("external service unavailable")` |

---

## Support

For questions or clarifications on error handling, refer to:
- [CLAUDE.md](../CLAUDE.md) - Project development guidelines
- [pkg/errors](../pkg/errors/) - Error package source code
- Error handling examples in existing use cases and repositories
