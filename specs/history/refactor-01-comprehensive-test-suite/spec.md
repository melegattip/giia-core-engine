# Feature Specification: Comprehensive Test Suite

**Created**: 2025-12-10
**Priority**: 🔴 CRITICAL
**Effort**: 5-8 days

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Core Use Case Testing (Priority: P1)

As a developer, I need comprehensive unit tests for all use cases so that I can refactor code with confidence and catch regressions before production.

**Why this priority**: Without tests, the codebase is fragile and unmaintainable. Any refactoring risks breaking production. Currently at ~6% coverage vs 85% target - this is a blocker for production deployment.

**Independent Test**: Can be fully tested by running `go test ./internal/core/usecases/... -coverprofile=coverage.out` and verifying coverage is above 85% for all packages.

**Acceptance Scenarios**:

1. **Scenario**: Login use case has comprehensive test coverage
   - **Given** LoginUseCase with all dependencies mocked
   - **When** tests run with coverage analysis
   - **Then** coverage report shows >85% coverage with success, validation error, and system error scenarios tested

2. **Scenario**: Registration use case handles all edge cases
   - **Given** RegisterUseCase tests
   - **When** running test suite
   - **Then** tests cover: duplicate email, invalid email format, weak password, database errors, and successful registration

3. **Scenario**: RBAC permission checks are thoroughly tested
   - **Given** CheckPermissionUseCase tests
   - **When** executing test scenarios
   - **Then** tests verify: wildcard permissions, role inheritance, cached permissions, permission denial, and multi-role scenarios

---

### User Story 2 - Repository Layer Testing (Priority: P1)

As a developer, I need unit tests for all repository implementations so that database queries, tenant scoping, and error handling are verified.

**Why this priority**: Repository layer is critical infrastructure - bugs here cause data corruption, tenant isolation breaches, or application crashes. Current coverage: 0%.

**Independent Test**: Run `go test ./internal/infrastructure/repositories/... -coverprofile=coverage.out` and verify >85% coverage with all CRUD operations tested.

**Acceptance Scenarios**:

1. **Scenario**: UserRepository CRUD operations tested
   - **Given** UserRepository with mocked GORM database
   - **When** tests execute Create, Read, Update, Delete operations
   - **Then** all operations verify: success cases, not found errors, constraint violations, and tenant scoping

2. **Scenario**: TokenRepository lifecycle tested
   - **Given** TokenRepository tests
   - **When** running token lifecycle scenarios
   - **Then** tests cover: token creation, validation, expiry, revocation, and cleanup of expired tokens

3. **Scenario**: RoleRepository hierarchy tested
   - **Given** RoleRepository with role hierarchy data
   - **When** testing role operations
   - **Then** tests verify: role creation, permission assignment, parent role resolution, and circular dependency prevention

---

### User Story 3 - Infrastructure Adapters Testing (Priority: P2)

As a developer, I need tests for JWT manager, caching, and rate limiting adapters so that external integrations work reliably.

**Why this priority**: These adapters are critical for security and performance but currently have 0% coverage. Can be implemented after use case/repository tests.

**Independent Test**: Run `go test ./internal/infrastructure/adapters/... -coverprofile=coverage.out` and verify >85% coverage.

**Acceptance Scenarios**:

1. **Scenario**: JWT Manager token generation and validation
   - **Given** JWTManager with test configuration
   - **When** generating and validating tokens
   - **Then** tests verify: token generation, signature validation, expiry handling, role extraction, and invalid token rejection

2. **Scenario**: Permission cache operations
   - **Given** RedisPermissionCache with mock Redis
   - **When** executing cache operations
   - **Then** tests verify: cache hit, cache miss, TTL expiry, invalidation, and bulk invalidation

3. **Scenario**: Rate limiter enforcement
   - **Given** RateLimiter adapter
   - **When** simulating request patterns
   - **Then** tests verify: rate limit enforcement, reset after window, per-user limits, and whitelisted IPs

---

### User Story 4 - HTTP/gRPC Handler Testing (Priority: P3)

As a developer, I need tests for HTTP and gRPC handlers so that request validation, error responses, and middleware integration work correctly.

**Why this priority**: Important for API contract validation but lower priority than business logic. Current coverage: 1.9%.

**Independent Test**: Run `go test ./internal/infrastructure/entrypoints/... -coverprofile=coverage.out` and verify >75% coverage (lower threshold acceptable for handlers).

**Acceptance Scenarios**:

1. **Scenario**: HTTP handler request validation
   - **Given** AuthHandler with mocked use cases
   - **When** sending invalid requests
   - **Then** handler returns 400 Bad Request with clear error messages

2. **Scenario**: gRPC handler error mapping
   - **Given** gRPC server implementation
   - **When** use case returns typed errors
   - **Then** gRPC handler maps errors to correct gRPC status codes (InvalidArgument, Unauthenticated, PermissionDenied, Internal)

3. **Scenario**: Middleware integration
   - **Given** HTTP handlers with authentication middleware
   - **When** making requests
   - **Then** tests verify: JWT extraction, context population, tenant isolation, and unauthorized access rejection

---

### Edge Cases

- What happens when mock expectations are not met (test framework should fail with clear message)?
- How to handle tests that depend on time (use TimeManager mocks)?
- What happens when database constraints change (tests should catch breaking changes)?
- How to test race conditions (use `-race` flag and concurrent test scenarios)?
- How to handle flaky tests (ensure deterministic mocks, no real time dependencies)?
- How to test graceful degradation (simulate service unavailability)?
- How to verify test isolation (tests should not affect each other)?

## Requirements *(mandatory)*

### Functional Requirements

#### Test Coverage
- **FR-001**: System MUST achieve minimum 85% test coverage for `internal/core/usecases/*` packages
- **FR-002**: System MUST achieve minimum 85% test coverage for `internal/infrastructure/repositories/*` packages
- **FR-003**: System MUST achieve minimum 85% test coverage for `internal/infrastructure/adapters/*` packages
- **FR-004**: System MUST achieve minimum 75% test coverage for `internal/infrastructure/entrypoints/*` packages
- **FR-005**: All tests MUST pass with `go test ./... -count=1 -race` (race detection enabled)

#### Test Structure
- **FR-006**: All tests MUST follow Given-When-Then pattern as defined in CLAUDE.md
- **FR-007**: Test names MUST follow format: `TestFunctionName_Scenario_ExpectedBehavior`
- **FR-008**: Table-driven tests MUST be used for functions with multiple scenarios (minimum 3 scenarios)
- **FR-009**: Test variables MUST use `given` prefix for inputs and `expected` prefix for outputs
- **FR-010**: All mocks MUST use specific parameters instead of `mock.Anything` where values are known

#### Mock Usage
- **FR-011**: All repository interfaces MUST have mock implementations in `providers/mocks.go`
- **FR-012**: TimeManager MUST be mocked in all tests (no direct `time.Now()` or `time.Sleep()`)
- **FR-013**: Mock expectations MUST be validated with `AssertExpectations(t)` at end of each test
- **FR-014**: Mock methods MUST match interface signatures exactly
- **FR-015**: Mock setup MUST be specific to test scenario (avoid global mock configuration)

#### Test Categories
- **FR-016**: Unit tests MUST test single function/method in isolation with all dependencies mocked
- **FR-017**: Each use case MUST have minimum 5 test scenarios: success, validation errors (2), system errors, edge cases
- **FR-018**: Each repository method MUST have minimum 4 test scenarios: success, not found, constraint violation, database error
- **FR-019**: Error handling paths MUST be explicitly tested (not just happy path)
- **FR-020**: All public functions MUST have corresponding unit tests

#### Test Quality
- **FR-021**: Tests MUST be deterministic (same input always produces same output)
- **FR-022**: Tests MUST be isolated (no shared state between tests)
- **FR-023**: Tests MUST clean up resources (defer mock cleanup, context cancellation)
- **FR-024**: Tests MUST use meaningful assertions with clear failure messages
- **FR-025**: Tests MUST not depend on external services (use mocks for all I/O)

### Key Entities

- **Test Suite**: Collection of tests for a specific package/component
- **Test Scenario**: Individual test case with Given-When-Then structure
- **Mock**: Test double implementing interface for dependency injection
- **Coverage Report**: Analysis showing percentage of code executed by tests
- **Test Fixture**: Reusable test data setup (e.g., `setupTestData()` functions)

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Overall test coverage reaches 85% or higher for critical packages
- **SC-002**: All 77 Go files in auth-service have corresponding test files (77 test files created)
- **SC-003**: Zero failing test suites (currently 1 failing: rbac tests)
- **SC-004**: All tests pass with race detector enabled (`go test -race`)
- **SC-005**: Test execution time under 30 seconds for full suite (with parallel execution)
- **SC-006**: Code coverage report generated successfully: `go test ./... -coverprofile=coverage.out -covermode=atomic`
- **SC-007**: No `time.Now()` or `time.Sleep()` calls in test files (use TimeManager mocks)
- **SC-008**: 100% of use cases have table-driven tests with minimum 5 scenarios each
- **SC-009**: CI/CD pipeline fails if coverage drops below 80% threshold
- **SC-010**: Test documentation added to auth-service README explaining how to run tests and interpret coverage
