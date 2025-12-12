# Implementation Plan: Comprehensive Test Suite

**Date**: 2025-12-10
**Spec**: [spec.md](./spec.md)

## Summary

Implement comprehensive unit tests for auth-service to achieve minimum 85% test coverage across all critical packages. Currently at ~6% coverage with only 3 test files for 77 Go files. This includes fixing failing RBAC tests, adding tests for all use cases, repositories, and infrastructure adapters, following Given-When-Then pattern with table-driven tests and proper mocking.

## Technical Context

**Language/Version**: Go 1.23.4
**Primary Dependencies**: testify/mock, testify/assert, google/uuid, GORM
**Storage**: PostgreSQL 16 (mocked in unit tests via GORM mocks)
**Testing**: go test, testify framework, table-driven tests
**Target Platform**: Linux/Windows server
**Project Type**: Microservice backend
**Performance Goals**: Full test suite <30 seconds execution time
**Constraints**: Must maintain Given-When-Then pattern, use TimeManager mocks, no `time.Sleep()` in tests
**Scale/Scope**: 77 Go files requiring ~75 test files, targeting 85%+ coverage

## Project Structure

### Documentation (this feature)

```text
specs/refactor-01-comprehensive-test-suite/
├── plan.md              # This file
└── spec.md              # Feature specification
```

### Source Code (auth-service)

```text
services/auth-service/
├── internal/
│   ├── core/
│   │   ├── usecases/
│   │   │   ├── auth/
│   │   │   │   ├── login.go
│   │   │   │   ├── login_test.go              # NEW
│   │   │   │   ├── register.go
│   │   │   │   ├── register_test.go           # NEW
│   │   │   │   ├── refresh_token.go
│   │   │   │   ├── refresh_token_test.go      # NEW
│   │   │   │   ├── validate_token.go
│   │   │   │   └── validate_token_test.go     # EXISTS - needs improvement
│   │   │   ├── rbac/
│   │   │   │   ├── check_permission.go
│   │   │   │   ├── check_permission_test.go   # EXISTS - FAILING, needs fix
│   │   │   │   ├── get_user_permissions.go
│   │   │   │   ├── get_user_permissions_test.go # NEW
│   │   │   │   ├── batch_check_permissions.go
│   │   │   │   └── batch_check_permissions_test.go # NEW
│   │   │   └── role/
│   │   │       ├── create_role.go
│   │   │       ├── create_role_test.go        # NEW
│   │   │       ├── assign_permission.go
│   │   │       └── assign_permission_test.go  # NEW
│   │   └── providers/
│   │       └── mocks.go                        # UPDATE - add missing mocks
│   ├── infrastructure/
│   │   ├── repositories/
│   │   │   ├── user_repository.go
│   │   │   ├── user_repository_test.go        # NEW
│   │   │   ├── token_repository.go
│   │   │   ├── token_repository_test.go       # NEW
│   │   │   ├── role_repository.go
│   │   │   ├── role_repository_test.go        # NEW
│   │   │   ├── permission_repository.go
│   │   │   └── permission_repository_test.go  # NEW
│   │   └── adapters/
│   │       ├── jwt/
│   │       │   ├── jwt_manager.go
│   │       │   └── jwt_manager_test.go        # NEW
│   │       ├── cache/
│   │       │   ├── permission_cache.go
│   │       │   └── permission_cache_test.go   # NEW
│   │       └── rate_limiter/
│   │           ├── rate_limiter.go
│   │           └── rate_limiter_test.go       # NEW
├── pkg/
│   ├── time/
│   │   ├── time_manager.go                    # NEW (prerequisite)
│   │   └── time_manager_mock.go               # NEW (prerequisite)
│   ├── logger/
│   │   ├── logger.go
│   │   └── logger_test.go                     # NEW
│   └── database/
│       ├── connection.go
│       └── connection_test.go                 # NEW
└── Makefile                                    # UPDATE - add test-coverage target
```

**Structure Decision**: Tests colocated with source files following Go convention. Each Go file gets corresponding `*_test.go` file in same package for white-box testing.

---

## Phase 1: Prerequisites (Blocking Foundation)

**Purpose**: Fix critical blockers before comprehensive test implementation

**⚠️ CRITICAL**: These must be complete before main test implementation

- [ ] T001 Fix failing RBAC tests (MockResolveInheritanceUseCase, MockPermissionCache.InvalidateUsersWithRole)
- [ ] T002 Implement TimeManager interface in pkg/time/ (RealTimeManager, MockTimeManager)
- [ ] T003 Update providers/mocks.go with all missing mock implementations
- [ ] T004 Add testify/mock and testify/assert to go.mod dependencies
- [ ] T005 Create test helper functions (setupTestContext, setupTestUser, setupTestRoles)

**Checkpoint**: Foundation ready - all mocks compile, TimeManager available, failing tests fixed

---

## Phase 2: Use Case Tests - Auth (Priority: P1)

**Goal**: Achieve 85%+ coverage for authentication use cases

**Independent Test**: `go test ./internal/core/usecases/auth/... -coverprofile=coverage.out`

### Tests for Auth Use Cases

- [ ] T006 [P] [US1] Create login_test.go with table-driven tests (success, invalid email, invalid password, db error, user not found)
- [ ] T007 [P] [US1] Create register_test.go with table-driven tests (success, duplicate email, weak password, validation errors, db error)
- [ ] T008 [P] [US1] Create refresh_token_test.go with tests (success, expired token, invalid token, revoked token, db error)
- [ ] T009 [P] [US1] Improve validate_token_test.go (add edge cases: malformed token, wrong signature, missing claims)

### Implementation for Auth Use Cases

- [ ] T010 [US1] Update LoginUseCase to accept TimeManager in constructor
- [ ] T011 [US1] Update RegisterUseCase to accept TimeManager in constructor
- [ ] T012 [US1] Update RefreshTokenUseCase to accept TimeManager in constructor
- [ ] T013 [US1] Verify all tests follow Given-When-Then pattern
- [ ] T014 [US1] Verify all test variables use `given` and `expected` prefixes
- [ ] T015 [US1] Verify coverage: `go tool cover -html=coverage.out` shows >85%

**Checkpoint**: Auth use cases fully tested, coverage >85%, all tests pass

---

## Phase 3: Use Case Tests - RBAC (Priority: P1)

**Goal**: Achieve 85%+ coverage for RBAC use cases

**Independent Test**: `go test ./internal/core/usecases/rbac/... -coverprofile=coverage.out`

### Tests for RBAC Use Cases

- [ ] T016 [P] [US1] Create get_user_permissions_test.go (success, user not found, role hierarchy, cached permissions, cache miss)
- [ ] T017 [P] [US1] Create batch_check_permissions_test.go (all granted, partial, all denied, empty list, invalid permissions)
- [ ] T018 [P] [US1] Improve check_permission_test.go (add wildcard scenarios, role inheritance, cache invalidation)

### Implementation for RBAC Use Cases

- [ ] T019 [US1] Verify MockPermissionCache has all methods (InvalidateUsersWithRole added in Phase 1)
- [ ] T020 [US1] Verify check_permission tests pass with interface types (not concrete mocks)
- [ ] T021 [US1] Add edge case tests (multiple roles, conflicting permissions, malformed permission strings)
- [ ] T022 [US1] Verify coverage: `go tool cover -html=coverage.out` shows >85%

**Checkpoint**: RBAC use cases fully tested, permission checking verified, coverage >85%

---

## Phase 4: Use Case Tests - Role Management (Priority: P2)

**Goal**: Achieve 85%+ coverage for role management use cases

**Independent Test**: `go test ./internal/core/usecases/role/... -coverprofile=coverage.out`

### Tests for Role Use Cases

- [ ] T023 [US1] Create create_role_test.go (success, duplicate name, validation errors, db error, permission assignment)
- [ ] T024 [US1] Create assign_permission_test.go (success, role not found, permission not found, duplicate assignment)
- [ ] T025 [US1] Create update_role_test.go if exists (success, not found, validation errors)
- [ ] T026 [US1] Create delete_role_test.go if exists (success, role in use, not found, cascade delete)

### Implementation for Role Use Cases

- [ ] T027 [US1] Ensure all role use cases have minimum 5 test scenarios each
- [ ] T028 [US1] Test role hierarchy edge cases (circular dependencies, max depth)
- [ ] T029 [US1] Verify coverage: `go tool cover -html=coverage.out` shows >85%

**Checkpoint**: Role management fully tested, coverage >85%

---

## Phase 5: Repository Tests (Priority: P1)

**Goal**: Achieve 85%+ coverage for all repository implementations

**Independent Test**: `go test ./internal/infrastructure/repositories/... -coverprofile=coverage.out`

### Tests for Repository Layer

- [ ] T030 [P] [US2] Create user_repository_test.go (Create success, duplicate email, FindByID success, not found, Update, Delete, tenant scoping)
- [ ] T031 [P] [US2] Create token_repository_test.go (Create, FindByToken, revoke, cleanup expired, tenant scoping)
- [ ] T032 [P] [US2] Create role_repository_test.go (Create, FindByID, list, assign permissions, resolve hierarchy, tenant scoping)
- [ ] T033 [P] [US2] Create permission_repository_test.go (Create, FindByCode, list, filter by service)

### Implementation for Repository Tests

- [ ] T034 [US2] Use GORM mocks (sqlmock or in-memory SQLite) for database interactions
- [ ] T035 [US2] Test all CRUD operations for each repository
- [ ] T036 [US2] Test constraint violations (unique, foreign key) return typed errors
- [ ] T037 [US2] Test tenant scoping prevents cross-organization data access
- [ ] T038 [US2] Test error mapping (gorm.ErrRecordNotFound → NewResourceNotFound, etc.)
- [ ] T039 [US2] Verify coverage: `go tool cover -html=coverage.out` shows >85%

**Checkpoint**: All repositories tested, tenant isolation verified, coverage >85%

---

## Phase 6: Infrastructure Adapter Tests (Priority: P2)

**Goal**: Achieve 85%+ coverage for infrastructure adapters

**Independent Test**: `go test ./internal/infrastructure/adapters/... -coverprofile=coverage.out`

### Tests for Adapter Layer

- [ ] T040 [US3] Create jwt_manager_test.go (GenerateToken, ValidateToken, expired token, invalid signature, role extraction)
- [ ] T041 [US3] Create permission_cache_test.go (Get hit, Get miss, Set, Invalidate, InvalidateUsersWithRole, TTL expiry)
- [ ] T042 [US3] Create rate_limiter_test.go (allow under limit, deny over limit, window reset, per-user limits)

### Implementation for Adapter Tests

- [ ] T043 [US3] Mock Redis for cache tests (use miniredis or mock)
- [ ] T044 [US3] Use TimeManager mocks for JWT expiry tests
- [ ] T045 [US3] Test error handling (cache unavailable, JWT parse error)
- [ ] T046 [US3] Verify coverage: `go tool cover -html=coverage.out` shows >85%

**Checkpoint**: All adapters tested, external service interactions mocked, coverage >85%

---

## Phase 7: Shared Package Tests (Priority: P3)

**Goal**: Achieve >75% coverage for pkg/* shared libraries

**Independent Test**: `go test ./pkg/... -coverprofile=coverage.out`

### Tests for Shared Packages

- [ ] T047 [US3] Create logger_test.go (Debug, Info, Warn, Error, Fatal, structured tags)
- [ ] T048 [US3] Create database/connection_test.go (successful connection, retry logic, connection pool)
- [ ] T049 [US3] Create time_manager_test.go (RealTimeManager.Now returns UTC, MockTimeManager returns fixed time)

### Implementation for Shared Package Tests

- [ ] T050 [US3] Test logger output format (JSON, timestamp, level, message, tags)
- [ ] T051 [US3] Test database retry logic with mock failures
- [ ] T052 [US3] Test TimeManager format functions (ISO8601, StringToUTC)
- [ ] T053 [US3] Verify coverage: `go tool cover -html=coverage.out` shows >75%

**Checkpoint**: Shared packages tested, coverage >75%

---

## Phase 8: Handler Tests (Priority: P3)

**Goal**: Achieve >75% coverage for HTTP/gRPC handlers

**Independent Test**: `go test ./internal/infrastructure/entrypoints/... -coverprofile=coverage.out`

### Tests for Handler Layer

- [ ] T054 [US4] Create user_handler_test.go improvements (add missing test scenarios)
- [ ] T055 [US4] Create auth_handler_test.go (login endpoint, register endpoint, refresh endpoint, request validation)
- [ ] T056 [US4] Create grpc_server_test.go (gRPC method handlers, error code mapping)

### Implementation for Handler Tests

- [ ] T057 [US4] Use httptest.ResponseRecorder for HTTP tests
- [ ] T058 [US4] Test request validation (binding errors return 400)
- [ ] T059 [US4] Test error response format (JSON structure, status codes)
- [ ] T060 [US4] Test middleware integration (JWT extraction, context population)
- [ ] T061 [US4] Verify coverage: `go tool cover -html=coverage.out` shows >75%

**Checkpoint**: All handlers tested, HTTP/gRPC contracts verified, coverage >75%

---

## Phase 9: Test Quality & Documentation (Cross-Cutting)

**Purpose**: Ensure test quality and proper documentation

- [ ] T062 Run all tests with race detector: `go test ./... -race -count=1`
- [ ] T063 Fix any race conditions detected
- [ ] T064 Verify no `time.Now()` calls in test files: `grep -r "time.Now()" **/*_test.go`
- [ ] T065 Verify no `time.Sleep()` calls in test files (except concurrency tests)
- [ ] T066 Generate overall coverage report: `go test ./... -coverprofile=coverage.out -covermode=atomic`
- [ ] T067 Verify overall coverage >85% for critical packages
- [ ] T068 Add coverage badge to README (if using CI/CD with coverage reporting)
- [ ] T069 Document testing strategy in auth-service/README.md
- [ ] T070 Create test execution guide (how to run tests, interpret coverage, debug failures)
- [ ] T071 Update CLAUDE.md with test examples if needed
- [ ] T072 Code review: verify all tests follow Given-When-Then pattern
- [ ] T073 Code review: verify all mocks use specific parameters (not mock.Anything)
- [ ] T074 Configure CI/CD to fail if coverage drops below 80%
- [ ] T075 Add Makefile targets: `test`, `test-coverage`, `test-race`, `test-verbose`

**Checkpoint**: All tests documented, quality gates in place, CI/CD configured

---

## Dependencies & Execution Order

### Phase Dependencies

- **Prerequisites (Phase 1)**: No dependencies - MUST complete first (blocks everything)
- **Auth Use Cases (Phase 2)**: Depends on Prerequisites
- **RBAC Use Cases (Phase 3)**: Depends on Prerequisites
- **Role Use Cases (Phase 4)**: Depends on Prerequisites, can run parallel with Phase 2-3
- **Repository Tests (Phase 5)**: Depends on Prerequisites, can run parallel with Phase 2-4
- **Adapter Tests (Phase 6)**: Depends on Prerequisites, can run parallel with Phase 2-5
- **Shared Package Tests (Phase 7)**: Depends on TimeManager (Phase 1), can run parallel
- **Handler Tests (Phase 8)**: Depends on use case tests (Phase 2-4), can run after those complete
- **Test Quality (Phase 9)**: Depends on all test phases complete

### Parallel Execution Opportunities

After Phase 1 (Prerequisites), the following can run in parallel:
- Phase 2 (Auth use case tests)
- Phase 3 (RBAC use case tests)
- Phase 4 (Role use case tests)
- Phase 5 (Repository tests)
- Phase 6 (Adapter tests)
- Phase 7 (Shared package tests)

Phase 8 should wait for Phase 2-4 completion.
Phase 9 should wait for all previous phases.

### Within Each Phase

- Create test files before running coverage
- Fix compilation errors before expecting tests to pass
- Run tests frequently (after each test file creation)
- Verify coverage incrementally (don't wait until end)

## Notes

- Tests follow strict Given-When-Then pattern (mandatory from CLAUDE.md)
- Use table-driven tests for functions with 3+ scenarios
- All mocks use specific parameters (avoid `mock.Anything` unless truly necessary)
- TimeManager mocks required for all time-dependent tests
- Target: 85% coverage for use cases/repositories, 75% for handlers/shared packages
- Commit after each phase checkpoint
- Run `go test -race` regularly to catch concurrency issues
- Verify tests pass in CI/CD pipeline before marking phase complete
