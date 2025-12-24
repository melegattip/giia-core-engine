# GIIA Development Guide

**Version**: 1.0  
**Last Updated**: 2025-12-23  
**Target**: All developers working on GIIA Core Engine

---

## 📖 Table of Contents

1. [Overview](#overview)
2. [Project Structure](#project-structure)
3. [Clean Architecture](#clean-architecture)
4. [Go Coding Standards](#go-coding-standards)
5. [Error Handling](#error-handling)
6. [Logging](#logging)
7. [Testing](#testing)
8. [Git Workflow](#git-workflow)
9. [Code Quality](#code-quality)

---

## Overview

This guide consolidates all development standards for the GIIA Core Engine project. All code contributions must follow these guidelines.

### Key Principles

- **Clean Architecture**: Clear separation between domain, use cases, and infrastructure
- **Typed Errors**: Use `pkg/errors` instead of `fmt.Errorf`
- **Testability**: All business logic must be unit testable
- **Consistency**: Follow Go conventions and project standards

---

## Project Structure

### Monorepo Layout

```
giia-core-engine/
├── services/                    # Microservices
│   ├── auth-service/           # Authentication & RBAC
│   ├── catalog-service/        # Products & Suppliers
│   ├── ddmrp-engine-service/   # Buffer calculations
│   ├── execution-service/      # Orders & Inventory
│   ├── analytics-service/      # KPIs & Reports
│   └── ai-intelligence-hub/    # AI Assistant
│
├── pkg/                         # Shared Packages
│   ├── config/                 # Viper configuration
│   ├── logger/                 # Zerolog logging
│   ├── database/               # GORM connection
│   ├── errors/                 # Typed errors
│   └── events/                 # NATS client
│
├── api/proto/                   # gRPC Definitions
├── k8s/                         # Kubernetes manifests
├── scripts/                     # Utility scripts
└── docs/                        # Documentation
```

### Service Structure

Each service follows Clean Architecture:

```
service-name/
├── cmd/api/main.go              # Entry point
├── internal/
│   ├── core/                    # 🧠 DOMAIN LAYER
│   │   ├── domain/             # Entities, value objects
│   │   ├── usecases/           # Business logic
│   │   └── providers/          # Interface contracts
│   │
│   └── infrastructure/          # 🔌 INFRASTRUCTURE LAYER
│       ├── adapters/           # External integrations
│       ├── repositories/       # Data access (GORM)
│       └── entrypoints/        # HTTP/gRPC handlers
│
├── api/proto/                   # Service-specific protos
├── migrations/                  # Database migrations
└── go.mod                       # Service module
```

---

## Clean Architecture

### Layer Responsibilities

| Layer | Contains | Dependencies |
|-------|----------|--------------|
| **Domain** | Entities, value objects | None |
| **Use Cases** | Business logic | Domain only |
| **Providers** | Interface contracts | Domain only |
| **Infrastructure** | Implementations | All layers |

### Dependency Rule

**Inner layers NEVER depend on outer layers.**

```go
// ✅ CORRECT - Use case depends on interface (provider)
type LoginUseCase struct {
    userRepo providers.UserRepository  // Interface
    logger   providers.Logger
}

// ❌ INCORRECT - Use case depends on concrete implementation
type LoginUseCase struct {
    userRepo *repositories.PostgresUserRepository  // Concrete!
}
```

### Entity Example

```go
// internal/core/domain/user.go
package domain

import (
    "time"
    "github.com/google/uuid"
)

type UserStatus string

const (
    UserStatusActive   UserStatus = "active"
    UserStatusInactive UserStatus = "inactive"
)

type User struct {
    ID             uuid.UUID
    OrganizationID uuid.UUID
    Email          string
    PasswordHash   string
    FirstName      string
    LastName       string
    Status         UserStatus
    CreatedAt      time.Time
    UpdatedAt      time.Time
}

// Domain methods
func (u *User) IsActive() bool {
    return u.Status == UserStatusActive
}
```

### Use Case Example

```go
// internal/core/usecases/auth/login.go
package auth

type LoginUseCase struct {
    userRepo        providers.UserRepository
    passwordService providers.PasswordService
    tokenManager    providers.TokenManager
    logger          providers.Logger
}

func (uc *LoginUseCase) Execute(ctx context.Context, email, password string) (*LoginResponse, error) {
    // 1. Validate input
    if email == "" {
        return nil, pkgErrors.NewBadRequest("email is required")
    }

    // 2. Get user
    user, err := uc.userRepo.GetByEmail(ctx, email)
    if err != nil {
        return nil, err
    }

    // 3. Verify password
    if !uc.passwordService.Verify(password, user.PasswordHash) {
        return nil, pkgErrors.NewUnauthorized("invalid credentials")
    }

    // 4. Generate tokens
    token, err := uc.tokenManager.Generate(user)
    if err != nil {
        return nil, err
    }

    return &LoginResponse{
        AccessToken: token,
        User:        user,
    }, nil
}
```

---

## Go Coding Standards

### Naming Conventions

| Type | Convention | Example |
|------|------------|---------|
| Variables | camelCase | `userCount`, `isValid` |
| Constants | camelCase | `maxRetries`, `defaultTimeout` |
| Exported types | PascalCase | `UserRepository`, `LoginUseCase` |
| Packages | snake_case | `time_manager`, `data_processing` |
| Acronyms | UPPERCASE | `userID`, `httpClient`, `apiURL` |

### Function Structure

```go
func ProcessData(ctx context.Context, input *Input) (*Output, error) {
    // 1. Validate input
    if input == nil {
        return nil, pkgErrors.NewBadRequest("input cannot be nil")
    }

    // 2. Business logic
    result, err := doProcessing(ctx, input)
    if err != nil {
        return nil, err
    }

    // 3. Return result
    return &Output{Data: result}, nil
}
```

### Context Usage

- Always pass `context.Context` as first parameter
- Propagate context through all layers
- Use context for cancellation and timeouts

```go
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*Entity, error) {
    var entity Entity
    err := r.db.WithContext(ctx).First(&entity, id).Error
    return &entity, err
}
```

---

## Error Handling

### Golden Rules

1. **Never use `fmt.Errorf`** - Use typed errors from `pkg/errors`
2. **Never ignore errors** - Always handle or explicitly discard
3. **Preserve error types** - Don't wrap typed errors
4. **Validate early** - Check inputs at function start

### Error Constructors

```go
import pkgErrors "github.com/giia/giia-core-engine/pkg/errors"

// Validation (400)
pkgErrors.NewBadRequest("email is required")

// Authentication (401)
pkgErrors.NewUnauthorized("invalid credentials")

// Authorization (403)
pkgErrors.NewForbidden("insufficient permissions")

// Not Found (404)
pkgErrors.NewNotFound("user not found")

// Conflict (409)
pkgErrors.NewConflict("user already exists")

// Server Error (500)
pkgErrors.NewInternalServerError("database error")
```

### GORM Error Mapping

```go
import "gorm.io/gorm"

func (r *UserRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
    var user domain.User
    err := r.db.WithContext(ctx).First(&user, id).Error

    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, pkgErrors.NewNotFound("user not found")
        }
        return nil, pkgErrors.NewInternalServerError("database query failed")
    }

    return &user, nil
}
```

---

## Logging

### Use Structured Logging

```go
import "github.com/giia/giia-core-engine/pkg/logger"

// Info level
logger.Info(ctx, "Processing request", logs.Tags{
    "user_id": userID,
    "action":  "login",
})

// Error level
logger.Error(ctx, err, "Failed to process request", logs.Tags{
    "user_id":   userID,
    "operation": "database_query",
})
```

### Logging Guidelines

- Include relevant context (user_id, organization_id, operation)
- Don't log sensitive data (passwords, tokens)
- Use appropriate log levels (debug, info, warn, error)
- Log at entry and exit of significant operations

---

## Testing

### Requirements

- **Minimum 85% coverage** for new code
- **Test all error paths** and edge cases
- **Use centralized mocks** from `providers/mocks.go`

### Test Structure (Given-When-Then)

```go
func TestLoginUseCase_Execute_Success(t *testing.T) {
    // Given
    mockUserRepo := new(providers.MockUserRepository)
    mockPasswordService := new(providers.MockPasswordService)
    logger := pkgLogger.New("test", "error")
    
    useCase := NewLoginUseCase(mockUserRepo, mockPasswordService, logger)
    
    givenEmail := "test@example.com"
    givenPassword := "password123"
    givenUser := &domain.User{ID: uuid.New(), Email: givenEmail}
    
    mockUserRepo.On("GetByEmail", mock.Anything, givenEmail).Return(givenUser, nil)
    mockPasswordService.On("Verify", givenPassword, mock.Anything).Return(true)
    
    // When
    result, err := useCase.Execute(context.Background(), givenEmail, givenPassword)
    
    // Then
    assert.NoError(t, err)
    assert.NotNil(t, result)
    assert.Equal(t, givenEmail, result.User.Email)
    mockUserRepo.AssertExpectations(t)
}
```

### Run Tests

```bash
# All tests
go test ./...

# With coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Specific package
go test -v ./internal/core/usecases/auth/...
```

---

## Git Workflow

### Branch Naming

```
feature/GIIA-123-add-login-endpoint
bugfix/GIIA-456-fix-token-validation
hotfix/GIIA-789-critical-security-fix
```

### Commit Messages

Follow [Conventional Commits](https://conventionalcommits.org):

```
feat(auth): add password reset functionality
fix(catalog): resolve duplicate SKU validation
docs(api): update endpoint documentation
test(rbac): add permission check tests
refactor(core): extract validation logic
```

### Pull Request Process

1. Create branch from `develop`
2. Make changes following standards
3. Run `make lint` and `make test`
4. Create PR targeting `develop`
5. Get 2 approvals
6. Squash and merge

---

## Code Quality

### Pre-Commit Checklist

- [ ] Code follows Clean Architecture
- [ ] No `fmt.Errorf` usage
- [ ] All errors handled
- [ ] Unit tests added (≥85% coverage)
- [ ] `make lint` passes
- [ ] `make test` passes
- [ ] Documentation updated

### Linting

```bash
# Run linters
make lint

# Auto-fix issues
make lint-fix

# Format code
make fmt
```

### Pre-commit Hooks

```bash
# Install
pip install pre-commit
pre-commit install

# Run manually
pre-commit run --all-files
```

---

## Related Documentation

- [Coding Standards Detail](./CODING_STANDARDS.md)
- [Error Handling Guide](./ERROR_HANDLING.md)
- [Testing Standards](./TESTING_STANDARDS.md)
- [Linting Guide](./LINTING_GUIDE.md)
- [Git Workflow](./GIT_WORKFLOW.md)

---

**Happy Coding! 🚀**
