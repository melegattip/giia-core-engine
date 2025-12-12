# Implementation Plan: Fix Failing RBAC Test Suite

**Date**: 2025-12-10
**Spec**: [spec.md](./spec.md)

## Summary

Fix compilation errors in internal/core/usecases/rbac/check_permission_test.go: create missing MockResolveInheritanceUseCase, add missing InvalidateUsersWithRole method to MockPermissionCache, fix type mismatches in test setup. Quick win to unblock CI/CD.

## Technical Context

**Language/Version**: Go 1.23.4
**Primary Dependencies**: testify/mock, google/uuid
**Testing**: Fix 9 existing test cases to compile and pass
**Project Type**: Microservice backend
**Effort**: 0.5 day (quick fix)
**Constraints**: Must not break existing passing tests

## Project Structure

### Files to Modify

```text
services/auth-service/
├── internal/
│   └── core/
│       ├── providers/
│       │   └── mocks.go                                # MODIFY - add missing mocks
│       └── usecases/
│           └── rbac/
│               └── check_permission_test.go            # MODIFY - fix type mismatches
```

---

## Phase 1: Create Missing Mock - MockResolveInheritanceUseCase

**Goal**: Add missing mock to providers/mocks.go

- [ ] T001 Open internal/core/providers/mocks.go
- [ ] T002 Identify ResolveInheritanceUseCase interface signature
- [ ] T003 Add MockResolveInheritanceUseCase struct with mock.Mock embedding
- [ ] T004 Implement Execute method: `Execute(ctx context.Context, roleID uuid.UUID) ([]*domain.Role, error)`
- [ ] T005 Verify mock implements interface correctly
- [ ] T006 Build providers package: `go build ./internal/core/providers`

**Checkpoint**: MockResolveInheritanceUseCase exists and compiles

---

## Phase 2: Fix MockPermissionCache - Add Missing Method

**Goal**: Add InvalidateUsersWithRole method to MockPermissionCache

- [ ] T007 Open internal/core/providers/mocks.go
- [ ] T008 Locate MockPermissionCache struct
- [ ] T009 Add method: `InvalidateUsersWithRole(ctx context.Context, roleID uuid.UUID) error`
- [ ] T010 Implement using m.Called(ctx, roleID) pattern
- [ ] T011 Verify MockPermissionCache now fully implements PermissionCache interface
- [ ] T012 Build providers package: `go build ./internal/core/providers`

**Checkpoint**: MockPermissionCache has all required methods

---

## Phase 3: Fix Test Type Mismatches

**Goal**: Fix check_permission_test.go to use interface types

- [ ] T013 Open internal/core/usecases/rbac/check_permission_test.go
- [ ] T014 Add missing import: `domain "github.com/giia/giia-core-engine/services/auth-service/internal/core/domain"`
- [ ] T015 Fix Line 50: change `mockGetUserPerms := new(MockGetUserPermissionsUseCase)` to use interface type
- [ ] T016 Fix Line 73: same type fix for mockGetUserPerms
- [ ] T017 Fix all 9 test cases: ensure NewCheckPermissionUseCase receives interface type
- [ ] T018 Verify compilation: `go build ./internal/core/usecases/rbac`

**Checkpoint**: Test file compiles without errors

---

## Phase 4: Verify Tests Pass

**Goal**: Run tests and verify all 9 scenarios pass

- [ ] T019 Run RBAC tests: `go test ./internal/core/usecases/rbac -v -count=1`
- [ ] T020 Verify all 9 test cases pass
- [ ] T021 Fix any remaining assertion failures
- [ ] T022 Run with race detector: `go test ./internal/core/usecases/rbac -race -count=1`
- [ ] T023 Verify coverage: `go test ./internal/core/usecases/rbac -coverprofile=coverage.out`
- [ ] T024 Document coverage percentage (should be >70% for RBAC package)

**Checkpoint**: All RBAC tests pass, no compilation errors

---

## Phase 5: Finalize & Document

**Purpose**: Ensure fix is complete and documented

- [ ] T025 Run full test suite: `go test ./... -count=1`
- [ ] T026 Verify no regressions in other test packages
- [ ] T027 Update CHANGELOG: note RBAC test fix
- [ ] T028 Git commit: "fix: resolve RBAC test compilation errors"
- [ ] T029 Verify CI/CD pipeline: ensure tests run in automation
- [ ] T030 Code review: verify mock implementations follow testify/mock patterns

**Checkpoint**: Fix complete, committed, CI/CD passing

---

## Dependencies & Execution Order

### Sequential Execution Required

Phases must run in order:
1. Phase 1 (Create MockResolveInheritanceUseCase) - must complete first
2. Phase 2 (Fix MockPermissionCache) - can run parallel with Phase 1
3. Phase 3 (Fix type mismatches) - depends on Phase 1-2 completion
4. Phase 4 (Verify tests) - depends on Phase 3
5. Phase 5 (Finalize) - depends on Phase 4

## Notes

- This is a quick win - should take <4 hours
- No business logic changes, only test infrastructure fixes
- Verify mock method signatures match interface exactly
- Use testify/mock.Called() pattern for all mocks
- Commit immediately after tests pass
- This unblocks other test development work
- Priority: Complete this before comprehensive test suite work (refactor-01)
