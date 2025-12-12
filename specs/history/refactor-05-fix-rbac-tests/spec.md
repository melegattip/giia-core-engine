# Feature Specification: Fix Failing RBAC Test Suite

**Created**: 2025-12-10
**Priority**: 🟠 HIGH
**Effort**: 0.5 day

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Fix Mock Implementation Errors (Priority: P1)

As a developer, I need all mock implementations to match their interface definitions so that RBAC tests compile and run successfully.

**Why this priority**: Test suite compilation failure blocks CI/CD and creates false security (tests exist but don't run). Low-hanging fruit that must be fixed immediately.

**Independent Test**: Run `go test ./internal/core/usecases/rbac -v` and verify all 9 test cases pass.

**Acceptance Scenarios**:

1. **Scenario**: MockResolveInheritanceUseCase exists
   - **Given** CheckPermissionTest references MockResolveInheritanceUseCase
   - **When** Compiling test suite
   - **Then** Mock exists in `providers/mocks.go` and implements correct interface

2. **Scenario**: MockPermissionCache has all required methods
   - **Given** PermissionCache interface defines InvalidateUsersWithRole method
   - **When** Using MockPermissionCache in tests
   - **Then** Mock implements all interface methods including InvalidateUsersWithRole

3. **Scenario**: Test uses interface types correctly
   - **Given** CheckPermissionUseCase accepts GetUserPermissionsUseCaseInterface
   - **When** Tests create mock instance
   - **Then** Tests use interface type, not concrete type in NewCheckPermissionUseCase call

---

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: MockResolveInheritanceUseCase MUST exist in `providers/mocks.go`
- **FR-002**: MockResolveInheritanceUseCase MUST implement Execute method with signature: `Execute(ctx context.Context, roleID uuid.UUID) ([]*domain.Role, error)`
- **FR-003**: MockPermissionCache MUST implement InvalidateUsersWithRole method
- **FR-004**: All 9 test cases in check_permission_test.go MUST pass
- **FR-005**: Test file MUST import domain package correctly

### Key Entities

- **MockResolveInheritanceUseCase**: Mock for role inheritance resolution
- **MockPermissionCache**: Mock for permission caching with full interface implementation
- **GetUserPermissionsUseCaseInterface**: Interface type used in tests

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: `go test ./internal/core/usecases/rbac -v -count=1` passes with 9/9 tests
- **SC-002**: Zero compilation errors in rbac test package
- **SC-003**: All mock expectations are validated in tests
- **SC-004**: Coverage report generated for rbac package
- **SC-005**: No breaking changes to existing passing tests
