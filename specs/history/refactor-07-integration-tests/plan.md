# Implementation Plan: Add Integration Tests for Database Layer

**Date**: 2025-12-10
**Spec**: [spec.md](./spec.md)

## Summary

Create integration tests for all repository implementations using testcontainers with real PostgreSQL. Verify SQL queries, GORM behavior, database constraints, and tenant isolation. Currently 0% integration test coverage.

## Technical Context

**Language/Version**: Go 1.23.4
**Primary Dependencies**: testcontainers-go, testcontainers/modules/postgres, GORM, PostgreSQL 16
**Testing**: Integration tests with real database, build tag `integration`
**Target Platform**: Linux/Windows development, Linux CI/CD
**Performance Goals**: Integration test suite <60 seconds
**Constraints**: Must use testcontainers (no external DB), clean up resources
**Scale/Scope**: 4 repositories × ~6 test scenarios each = ~24 integration tests

## Project Structure

### Source Code

```text
services/auth-service/
├── internal/
│   └── infrastructure/
│       └── repositories/
│           ├── user_repository.go
│           ├── user_repository_integration_test.go        # NEW
│           ├── token_repository.go
│           ├── token_repository_integration_test.go       # NEW
│           ├── role_repository.go
│           ├── role_repository_integration_test.go        # NEW
│           ├── permission_repository.go
│           └── permission_repository_integration_test.go  # NEW
├── Makefile                                                # UPDATE - add test-integration target
├── go.mod                                                  # UPDATE - add testcontainers deps
└── .github/
    └── workflows/
        └── integration-tests.yml                           # NEW - CI/CD workflow
```

---

## Phase 1: Setup Integration Test Infrastructure

**Purpose**: Configure testcontainers and build tags

- [ ] T001 Add testcontainers dependencies to go.mod: `go get github.com/testcontainers/testcontainers-go`
- [ ] T002 Add postgres module: `go get github.com/testcontainers/testcontainers-go/modules/postgres`
- [ ] T003 Create test helper: `internal/infrastructure/repositories/test_helpers.go`
- [ ] T004 Implement setupTestDB() helper: starts testcontainer, returns *gorm.DB
- [ ] T005 Implement teardownTestDB() helper: cleanup container and connections
- [ ] T006 Implement seedTestData() helper: inserts test organizations, users, roles
- [ ] T007 Test basic setup: verify testcontainer starts and connects

**Checkpoint**: Testcontainer infrastructure working, can start/stop PostgreSQL

---

## Phase 2: UserRepository Integration Tests (Priority: P1)

**Goal**: Comprehensive integration tests for UserRepository

- [ ] T008 [P] [US1] Create user_repository_integration_test.go with `//go:build integration` tag
- [ ] T009 [P] [US1] Test Create success: user persists with all fields
- [ ] T010 [P] [US1] Test Create duplicate email: GORM returns ErrDuplicatedKey, repository returns Conflict error
- [ ] T011 [P] [US1] Test FindByID success: retrieves user by UUID
- [ ] T012 [P] [US1] Test FindByID not found: returns NewResourceNotFound error
- [ ] T013 [P] [US1] Test FindByEmail success: retrieves user by email
- [ ] T014 [P] [US1] Test Update: changes persist to database
- [ ] T015 [P] [US1] Test Delete: user removed from database
- [ ] T016 [P] [US1] Test tenant scoping: org A cannot retrieve org B users
- [ ] T017 [P] [US1] Test List with pagination: returns correct page of users
- [ ] T018 [P] [US1] Run tests: `go test -tags=integration ./internal/infrastructure/repositories -run TestUserRepository -v`

**Checkpoint**: UserRepository fully tested with real database, tenant isolation verified

---

## Phase 3: TokenRepository Integration Tests (Priority: P1)

**Goal**: Integration tests for token lifecycle

- [ ] T019 [P] [US1] Create token_repository_integration_test.go with `//go:build integration` tag
- [ ] T020 [P] [US1] Test Create: token persists with hashed value, expiry, user relationship
- [ ] T021 [P] [US1] Test FindByToken: retrieves token by hashed value
- [ ] T022 [P] [US1] Test Revoke: token marked as revoked
- [ ] T023 [P] [US1] Test CleanupExpired: removes tokens past expiry date
- [ ] T024 [P] [US1] Test FindByUserID: returns all tokens for user
- [ ] T025 [P] [US1] Test tenant scoping: org A tokens isolated from org B
- [ ] T026 [P] [US1] Test cascade delete: user deletion removes associated tokens
- [ ] T027 [P] [US1] Run tests: `go test -tags=integration ./internal/infrastructure/repositories -run TestTokenRepository -v`

**Checkpoint**: Token lifecycle fully tested, cascade behavior verified

---

## Phase 4: RoleRepository Integration Tests (Priority: P1)

**Goal**: Integration tests for role hierarchy and permissions

- [ ] T028 [P] [US1] Create role_repository_integration_test.go with `//go:build integration` tag
- [ ] T029 [P] [US1] Test Create: role persists with name, description, organization
- [ ] T030 [P] [US1] Test FindByID: retrieves role by UUID
- [ ] T031 [P] [US1] Test AssignPermissions: many-to-many relationship persists
- [ ] T032 [P] [US1] Test ResolveHierarchy: returns parent roles recursively
- [ ] T033 [P] [US1] Test GetPermissions: returns all permissions including inherited
- [ ] T034 [P] [US1] Test tenant scoping: org A roles isolated from org B
- [ ] T035 [P] [US1] Test Delete: role removed, permission associations cleaned up
- [ ] T036 [P] [US1] Test circular dependency prevention: cannot create role loop
- [ ] T037 [P] [US1] Run tests: `go test -tags=integration ./internal/infrastructure/repositories -run TestRoleRepository -v`

**Checkpoint**: Role hierarchy and permissions fully tested

---

## Phase 5: PermissionRepository Integration Tests (Priority: P2)

**Goal**: Integration tests for permission management

- [ ] T038 [US1] Create permission_repository_integration_test.go with `//go:build integration` tag
- [ ] T039 [US1] Test Create: permission persists with code, service, resource, action
- [ ] T040 [US1] Test FindByCode: retrieves permission by unique code
- [ ] T041 [US1] Test List: returns all permissions, optionally filtered by service
- [ ] T042 [US1] Test wildcard permissions: catalog:*:*, *:products:read, etc.
- [ ] T043 [US1] Test duplicate code: constraint violation handled
- [ ] T044 [US1] Run tests: `go test -tags=integration ./internal/infrastructure/repositories -run TestPermissionRepository -v`

**Checkpoint**: Permission management fully tested

---

## Phase 6: Transaction Integration Tests (Priority: P2)

**Goal**: Test ACID properties and rollback behavior

- [ ] T045 [US2] Create transaction_test.go with `//go:build integration` tag
- [ ] T046 [US2] Test transaction commit: multiple operations persist atomically
- [ ] T047 [US2] Test transaction rollback: failed operation reverts all changes
- [ ] T048 [US2] Test nested transactions: GORM SavePoint behavior
- [ ] T049 [US2] Test concurrent access: no race conditions with proper locking
- [ ] T050 [US2] Run tests: `go test -tags=integration ./internal/infrastructure/repositories -run TestTransaction -v`

**Checkpoint**: Transaction behavior verified

---

## Phase 7: Makefile & CI/CD Integration

**Purpose**: Automate integration test execution

- [ ] T051 Add Makefile target: `test-integration` runs `go test -tags=integration ./... -v`
- [ ] T052 Add Makefile target: `test-integration-coverage` generates coverage report
- [ ] T053 Create .github/workflows/integration-tests.yml
- [ ] T054 Configure workflow: runs on PR, uses Docker for testcontainers
- [ ] T055 Add workflow step: setup Go, Docker
- [ ] T056 Add workflow step: run integration tests
- [ ] T057 Add workflow step: upload coverage report
- [ ] T058 Test CI/CD: create PR and verify integration tests run
- [ ] T059 Configure coverage threshold: fail if integration coverage <70%

**Checkpoint**: Integration tests run automatically in CI/CD

---

## Phase 8: Documentation & Best Practices

**Purpose**: Document integration test approach

- [ ] T060 Create docs/testing/integration-tests.md
- [ ] T061 Document testcontainers setup and usage
- [ ] T062 Document how to run integration tests locally
- [ ] T063 Document how to debug failed integration tests
- [ ] T064 Update auth-service/README.md: add integration test section
- [ ] T065 Document tenant isolation testing approach
- [ ] T066 Document transaction testing patterns
- [ ] T067 Create troubleshooting guide: common testcontainer issues
- [ ] T068 Update CLAUDE.md: add integration test examples

**Checkpoint**: Integration testing fully documented

---

## Dependencies & Execution Order

### Phase Dependencies

- **Infrastructure (Phase 1)**: No dependencies - must complete first
- **UserRepository (Phase 2)**: Depends on Phase 1
- **TokenRepository (Phase 3)**: Depends on Phase 1, can run parallel with Phase 2
- **RoleRepository (Phase 4)**: Depends on Phase 1, can run parallel with Phase 2-3
- **PermissionRepository (Phase 5)**: Depends on Phase 1, can run parallel with Phase 2-4
- **Transactions (Phase 6)**: Depends on Phase 2-5 (needs repositories implemented)
- **Makefile/CI (Phase 7)**: Depends on Phase 2-6
- **Documentation (Phase 8)**: Depends on all previous phases

### Parallel Execution Opportunities

After Phase 1, Phases 2-5 can run in parallel (different repository tests).

## Notes

- Integration tests use `//go:build integration` tag to separate from unit tests
- Testcontainers handles Docker automatically - no manual setup needed
- Each test should start with fresh database (use BeforeEach pattern)
- Clean up testcontainers after each test to avoid resource leaks
- Integration tests are slower than unit tests - use wisely
- Focus on scenarios that cannot be tested with mocks (SQL correctness, constraints, transactions)
- Tenant isolation is critical - must be tested in integration tests
- CI/CD must have Docker access for testcontainers
- Estimated execution time: 3-4 days for comprehensive integration test suite
