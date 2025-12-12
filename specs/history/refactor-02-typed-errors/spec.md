# Feature Specification: Enforce Typed Errors Throughout Codebase

**Created**: 2025-12-10
**Priority**: 🔴 CRITICAL
**Effort**: 2-3 days

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Repository Layer Typed Errors (Priority: P1)

As a developer, I need all repository methods to return typed errors from `pkg/errors` so that HTTP handlers can map errors to correct status codes without guessing.

**Why this priority**: Repository layer has 95% `fmt.Errorf` usage (highest violation rate). Typed errors enable proper HTTP status code mapping and structured error responses. Critical for API consistency.

**Independent Test**: Can be tested by making repository calls that trigger errors and verifying the error type matches expected `pkg/errors` constructor (e.g., `errors.NewInternalServerError`).

**Acceptance Scenarios**:

1. **Scenario**: Database constraint violation returns BadRequest
   - **Given** UserRepository.Create called with duplicate email
   - **When** GORM returns `ErrDuplicatedKey`
   - **Then** repository returns `pkgErrors.NewConflict("user with this email already exists")`

2. **Scenario**: Record not found returns NotFound
   - **Given** UserRepository.FindByEmail called with non-existent email
   - **When** GORM returns `ErrRecordNotFound`
   - **Then** repository returns `pkgErrors.NewResourceNotFound("user not found")`

3. **Scenario**: Database connection error returns InternalServerError
   - **Given** Any repository operation fails due to database error
   - **When** GORM returns connection or query error
   - **Then** repository returns `pkgErrors.NewInternalServerError("database operation failed")`

---

### User Story 2 - Use Case Layer Typed Errors (Priority: P1)

As a developer, I need all use case methods to return typed errors so that business logic errors are distinguishable from infrastructure errors.

**Why this priority**: Use cases currently mix `fmt.Errorf` and typed errors (60/40 split). Inconsistent error handling makes debugging difficult and error responses unpredictable.

**Independent Test**: Run use case tests and verify all error returns use typed constructors from `pkg/errors`.

**Acceptance Scenarios**:

1. **Scenario**: Validation errors return BadRequest
   - **Given** LoginUseCase.Execute called with empty email
   - **When** validation fails
   - **Then** use case returns `pkgErrors.NewBadRequest("email is required")`

2. **Scenario**: Authentication failure returns Unauthorized
   - **Given** LoginUseCase with invalid password
   - **When** password verification fails
   - **Then** use case returns `pkgErrors.NewUnauthorizedRequest("invalid credentials")`

3. **Scenario**: Authorization failure returns Forbidden
   - **Given** User lacks required permission for operation
   - **When** permission check fails
   - **Then** use case returns `pkgErrors.NewForbidden("insufficient permissions")`

---

### User Story 3 - Infrastructure Adapter Typed Errors (Priority: P2)

As a developer, I need all infrastructure adapters (JWT, cache, email) to return typed errors so that adapter failures are handled consistently.

**Why this priority**: Adapters have 90% `fmt.Errorf` usage. Important for consistency but lower priority than core business logic layers.

**Independent Test**: Test each adapter's error paths and verify typed error returns.

**Acceptance Scenarios**:

1. **Scenario**: JWT validation failure
   - **Given** JWTManager.ValidateToken with expired token
   - **When** token validation fails
   - **Then** adapter returns `pkgErrors.NewUnauthorizedRequest("token expired")`

2. **Scenario**: Cache connection failure
   - **Given** RedisPermissionCache with Redis unavailable
   - **When** cache operation fails
   - **Then** adapter returns `pkgErrors.NewInternalServerError("cache unavailable")`

3. **Scenario**: Email service failure
   - **Given** EmailService with SMTP error
   - **When** email send fails
   - **Then** adapter returns `pkgErrors.NewInternalServerError("failed to send email")`

---

### User Story 4 - Error Response Consistency (Priority: P2)

As an API consumer, I need consistent error response format across all endpoints so that error handling in client applications is predictable.

**Why this priority**: Improves developer experience for frontend/mobile teams. Enables standardized error handling in client SDKs.

**Independent Test**: Make HTTP requests that trigger various errors and verify all return consistent JSON structure with error code, message, and HTTP status.

**Acceptance Scenarios**:

1. **Scenario**: Validation error response format
   - **Given** HTTP POST request with invalid payload
   - **When** handler returns error response
   - **Then** response has structure: `{"error_code": "BAD_REQUEST", "message": "email is required", "http_status": 400}`

2. **Scenario**: Authorization error response format
   - **Given** HTTP request without required permission
   - **When** authorization fails
   - **Then** response has structure: `{"error_code": "FORBIDDEN", "message": "insufficient permissions", "http_status": 403}`

3. **Scenario**: Internal error response format
   - **Given** HTTP request causes database error
   - **When** internal error occurs
   - **Then** response has structure: `{"error_code": "INTERNAL_SERVER_ERROR", "message": "internal server error", "http_status": 500}` (no sensitive details exposed)

---

### Edge Cases

- What happens when error wrapping loses type information (ensure `errors.As` works)?
- How to preserve error context while using typed errors (use typed error with context fields)?
- What happens when multiple errors occur (return first critical error)?
- How to handle errors from external libraries (wrap in typed errors at boundary)?
- How to differentiate between client errors (4xx) and server errors (5xx)?
- How to log errors without exposing sensitive information in responses?
- How to handle error code conflicts (use centralized error code registry)?

## Requirements *(mandatory)*

### Functional Requirements

#### Error Construction
- **FR-001**: All repository methods MUST return typed errors using constructors from `pkg/errors`
- **FR-002**: All use case methods MUST return typed errors using constructors from `pkg/errors`
- **FR-003**: All infrastructure adapter methods MUST return typed errors using constructors from `pkg/errors`
- **FR-004**: Zero occurrences of `fmt.Errorf` in `internal/core/*` (business logic layer)
- **FR-005**: Zero occurrences of `fmt.Errorf` in `internal/infrastructure/repositories/*`

#### Error Types
- **FR-006**: System MUST use `NewBadRequest()` for validation errors (400)
- **FR-007**: System MUST use `NewUnauthorizedRequest()` for authentication failures (401)
- **FR-008**: System MUST use `NewForbidden()` for authorization failures (403)
- **FR-009**: System MUST use `NewResourceNotFound()` for missing resources (404)
- **FR-010**: System MUST use `NewConflict()` for duplicate resource errors (409)
- **FR-011**: System MUST use `NewTooManyRequests()` for rate limit violations (429)
- **FR-012**: System MUST use `NewInternalServerError()` for infrastructure failures (500)

#### Error Handling
- **FR-013**: Repositories MUST map GORM errors to appropriate typed errors:
  - `gorm.ErrRecordNotFound` → `NewResourceNotFound()`
  - `gorm.ErrDuplicatedKey` → `NewConflict()`
  - Other errors → `NewInternalServerError()`
- **FR-014**: Use cases MUST preserve error types from repository layer (don't wrap typed errors)
- **FR-015**: HTTP handlers MUST extract HTTP status code from typed errors using `err.HTTPStatus`
- **FR-016**: gRPC handlers MUST map typed errors to gRPC status codes:
  - `BadRequest` → `codes.InvalidArgument`
  - `UnauthorizedRequest` → `codes.Unauthenticated`
  - `Forbidden` → `codes.PermissionDenied`
  - `ResourceNotFound` → `codes.NotFound`
  - `InternalServerError` → `codes.Internal`

#### Error Responses
- **FR-017**: All HTTP error responses MUST include: error_code, message, http_status
- **FR-018**: Internal errors MUST NOT expose sensitive information (stack traces, database details)
- **FR-019**: Error messages MUST be actionable for API consumers
- **FR-020**: Error responses MUST be logged with full context (request ID, user ID, error details)

#### Error Testing
- **FR-021**: All error paths MUST have corresponding unit tests
- **FR-022**: Tests MUST verify error type using `errors.As()` or type assertion
- **FR-023**: Tests MUST verify error message contains expected information
- **FR-024**: Tests MUST verify HTTP status code matches error type
- **FR-025**: Integration tests MUST verify error response JSON structure

### Key Entities

- **CustomError**: Base error type with ErrorCode, Message, HTTPStatus fields
- **Error Constructor**: Function that creates typed error (e.g., `NewBadRequest`)
- **Error Code**: Enum of error types (BAD_REQUEST, UNAUTHORIZED_REQUEST, FORBIDDEN, etc.)
- **Error Response**: JSON structure returned to API clients
- **Error Context**: Additional fields for logging (user_id, request_id, operation)

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Zero `fmt.Errorf` occurrences in `internal/core/*` (verified by `grep -r "fmt.Errorf" internal/core`)
- **SC-002**: Zero `fmt.Errorf` occurrences in `internal/infrastructure/repositories/*`
- **SC-003**: 100% of repository error paths return typed errors (verified by code review)
- **SC-004**: 100% of use case error paths return typed errors (verified by code review)
- **SC-005**: All HTTP error responses follow consistent JSON structure (verified by integration tests)
- **SC-006**: golangci-lint custom rule blocks new `fmt.Errorf` usage in prohibited packages
- **SC-007**: Error handling tests achieve 100% coverage of error paths
- **SC-008**: API documentation includes error code reference table
- **SC-009**: Error response time under 5ms for typed error construction and serialization
- **SC-010**: Developer documentation updated with error handling guidelines and examples
