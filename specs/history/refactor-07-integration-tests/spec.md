# Feature Specification: Add Integration Tests for Database Layer

**Created**: 2025-12-10
**Priority**: 🟡 MEDIUM
**Effort**: 3-4 days

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Repository Integration Tests (Priority: P1)

As a developer, I need integration tests for all repository implementations using real PostgreSQL so that SQL queries, GORM behavior, and database constraints are verified.

**Why this priority**: Repository logic (tenant scoping, query construction, transactions) cannot be properly tested with mocks. Integration tests catch SQL errors and constraint violations that unit tests miss.

**Independent Test**: Run `go test -tags=integration ./internal/infrastructure/repositories/... -v` and verify all repository CRUD operations tested against real database.

**Acceptance Scenarios**:

1. **Scenario**: UserRepository CRUD with real database
   - **Given** PostgreSQL testcontainer running
   - **When** Creating, reading, updating, deleting users
   - **Then** All operations succeed and data persists correctly

2. **Scenario**: Tenant isolation verification
   - **Given** Users from two different organizations
   - **When** Querying with org A context
   - **Then** Only org A users returned (org B data not accessible)

3. **Scenario**: Constraint violation handling
   - **Given** User with email already in database
   - **When** Creating user with duplicate email
   - **Then** Repository returns proper typed error (Conflict)

---

### User Story 2 - Transaction Rollback Testing (Priority: P2)

As a developer, I need integration tests for transaction scenarios so that rollback behavior is verified under error conditions.

**Why this priority**: Transaction handling is critical for data consistency. Important to test but lower priority than basic CRUD operations.

**Independent Test**: Run transaction tests and verify failed operations don't persist data.

**Acceptance Scenarios**:

1. **Scenario**: Transaction rollback on error
   - **Given** Transaction with multiple operations
   - **When** Second operation fails
   - **Then** First operation is rolled back (no partial state)

2. **Scenario**: Transaction commit on success
   - **Given** Transaction with multiple operations
   - **When** All operations succeed
   - **Then** All changes persisted atomically

---

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Integration tests MUST use testcontainers for PostgreSQL
- **FR-002**: Integration tests MUST use `//go:build integration` build tag
- **FR-003**: UserRepository MUST have integration tests for CRUD operations
- **FR-004**: TokenRepository MUST have integration tests for token lifecycle
- **FR-005**: RoleRepository MUST have integration tests for role hierarchy
- **FR-006**: All integration tests MUST verify tenant isolation
- **FR-007**: Makefile MUST include `test-integration` target
- **FR-008**: Integration tests MUST run in CI/CD pipeline
- **FR-009**: Integration tests MUST clean up resources (defer container cleanup)
- **FR-010**: README MUST document how to run integration tests

### Key Entities

- **Testcontainer**: Docker container for PostgreSQL running during tests
- **Integration Test**: Test using real database instead of mocks
- **Transaction Test**: Test verifying ACID properties
- **Tenant Isolation Test**: Test ensuring org A cannot access org B data

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: All repositories have integration tests (`*_integration_test.go` files exist)
- **SC-002**: Integration tests pass consistently: `go test -tags=integration ./...`
- **SC-003**: Tenant isolation verified by attempting cross-tenant access
- **SC-004**: Constraint violations (unique, foreign key) tested and verified
- **SC-005**: Transaction rollback tests prevent partial state persistence
- **SC-006**: Integration tests complete in under 60 seconds
- **SC-007**: CI/CD pipeline includes integration test stage
- **SC-008**: Coverage report includes integration test coverage
