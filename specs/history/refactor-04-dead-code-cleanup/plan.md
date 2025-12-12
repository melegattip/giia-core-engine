# Implementation Plan: Remove Dead Code from Incomplete Migration

**Date**: 2025-12-10
**Spec**: [spec.md](./spec.md)

## Summary

Delete old architecture files (internal/domain/, internal/repository/, internal/handlers/, internal/usecases/user_service.go) that coexist with new Clean Architecture implementation. Reduces codebase by ~650 LOC, eliminates confusion about canonical implementations, and completes architectural migration.

## Technical Context

**Language/Version**: Go 1.23.4
**Testing**: All tests must pass after deletion
**Build**: Application must build successfully
**Project Type**: Microservice backend
**Scale/Scope**: 4 directories, ~650 LOC to delete, verify zero imports reference deleted code

## Project Structure

### Files to Delete

```text
services/auth-service/
├── internal/
│   ├── domain/
│   │   ├── user.go                        # DELETE - 92 lines, uint-based User
│   │   └── [any other domain files]       # DELETE
│   ├── repository/
│   │   ├── user_repository.go             # DELETE - 55 lines, raw SQL
│   │   └── [any other repository files]   # DELETE
│   ├── handlers/
│   │   ├── user_handler.go                # DELETE - ~150 lines
│   │   └── [any other handler files]      # DELETE
│   └── usecases/
│       └── user_service.go                # DELETE - 356 lines, monolithic
```

### Canonical Implementations (Remain)

```text
services/auth-service/
├── internal/
│   ├── core/
│   │   ├── domain/
│   │   │   └── user.go                    # KEEP - UUID-based, canonical
│   │   └── usecases/
│   │       ├── auth/                      # KEEP - focused use cases
│   │       ├── rbac/                      # KEEP
│   │       └── role/                      # KEEP
│   └── infrastructure/
│       ├── repositories/
│       │   └── user_repository.go         # KEEP - GORM-based, canonical
│       └── entrypoints/
│           └── http/handlers/             # KEEP - new handler structure
```

---

## Phase 1: Pre-Deletion Audit

**Purpose**: Verify what's being deleted and ensure no functionality lost

- [ ] T001 Audit internal/domain/user.go: document all fields and methods
- [ ] T002 Compare old User entity with new User entity: verify all fields migrated
- [ ] T003 Audit internal/repository/user_repository.go: document all methods
- [ ] T004 Compare old repository with new repository: verify all methods exist in new implementation
- [ ] T005 Audit internal/handlers/user_handler.go: document all endpoints
- [ ] T006 Compare old handlers with new handlers: verify all endpoints migrated
- [ ] T007 Audit internal/usecases/user_service.go: document all operations
- [ ] T008 Compare old UserService with new use cases: verify all operations covered
- [ ] T009 Search for imports: `grep -r "internal/domain" --include="*.go" .`
- [ ] T010 Search for imports: `grep -r "internal/repository" --include="*.go" .`
- [ ] T011 Search for imports: `grep -r "internal/handlers" --include="*.go" .`
- [ ] T012 Search for imports: `grep -r "internal/usecases/user_service" --include="*.go" .`
- [ ] T013 Document findings: create audit report showing what will be deleted

**Checkpoint**: Audit complete, confirmed no functionality will be lost

---

## Phase 2: Delete Old Domain Layer

**Goal**: Remove internal/domain/ directory

- [ ] T014 [P] [US1] Run full test suite before deletion: `go test ./... -count=1`
- [ ] T015 [P] [US1] Delete file: `rm internal/domain/user.go`
- [ ] T016 [P] [US1] Delete directory: `rm -rf internal/domain/` (if empty or all files confirmed old)
- [ ] T017 [P] [US1] Verify build succeeds: `go build ./...`
- [ ] T018 [P] [US1] Verify tests pass: `go test ./... -count=1`
- [ ] T019 [P] [US1] Verify zero imports: `grep -r "internal/domain" --include="*.go" .` returns nothing

**Checkpoint**: Old domain layer deleted, build succeeds, tests pass

---

## Phase 3: Delete Old Repository Layer

**Goal**: Remove internal/repository/ directory

- [ ] T020 [P] [US2] Delete file: `rm internal/repository/user_repository.go`
- [ ] T021 [P] [US2] Delete directory: `rm -rf internal/repository/`
- [ ] T022 [P] [US2] Verify build succeeds: `go build ./...`
- [ ] T023 [P] [US2] Verify tests pass: `go test ./... -count=1`
- [ ] T024 [P] [US2] Verify zero imports: `grep -r "internal/repository" --include="*.go" .` returns nothing
- [ ] T025 [P] [US2] Verify all repository calls use `internal/infrastructure/repositories/*`

**Checkpoint**: Old repository layer deleted, only GORM repositories remain

---

## Phase 4: Delete Old HTTP Handlers

**Goal**: Remove internal/handlers/ directory

- [ ] T026 [P] [US3] Verify main.go uses new handlers: check route registration
- [ ] T027 [P] [US3] Delete file: `rm internal/handlers/user_handler.go`
- [ ] T028 [P] [US3] Delete directory: `rm -rf internal/handlers/`
- [ ] T029 [P] [US3] Verify build succeeds: `go build ./...`
- [ ] T030 [P] [US3] Verify tests pass: `go test ./... -count=1`
- [ ] T031 [P] [US3] Verify zero imports: `grep -r "internal/handlers" --include="*.go" .` returns nothing
- [ ] T032 [P] [US3] Test HTTP endpoints: verify all endpoints respond correctly

**Checkpoint**: Old handlers deleted, HTTP API works with new handlers

---

## Phase 5: Delete Monolithic UserService

**Goal**: Remove internal/usecases/user_service.go

- [ ] T033 [US4] Verify all UserService methods migrated to specific use cases
- [ ] T034 [US4] Document mapping: UserService.Login → LoginUseCase, etc.
- [ ] T035 [US4] Delete file: `rm internal/usecases/user_service.go`
- [ ] T036 [US4] Delete directory: `rm -rf internal/usecases/` (if empty)
- [ ] T037 [US4] Verify build succeeds: `go build ./...`
- [ ] T038 [US4] Verify tests pass: `go test ./... -count=1`
- [ ] T039 [US4] Verify zero imports: `grep -r "internal/usecases/user_service" --include="*.go" .` returns nothing
- [ ] T040 [US4] Test all user operations: login, register, update profile, etc.

**Checkpoint**: Monolithic UserService deleted, all operations work via new use cases

---

## Phase 6: Documentation & Cleanup

**Purpose**: Update documentation and finalize cleanup

- [ ] T041 Update README.md: remove references to old architecture
- [ ] T042 Update architecture diagrams: show only new Clean Architecture structure
- [ ] T043 Create CHANGELOG entry: document migration completion
- [ ] T044 Create migration guide for team: explain what was deleted and why
- [ ] T045 Search for TODO comments referencing old code: update or remove
- [ ] T046 Search for code comments mentioning old paths: update to new paths
- [ ] T047 Verify go.mod: remove unused dependencies if any
- [ ] T048 Run `go mod tidy` to clean dependencies
- [ ] T049 Final verification: `go build ./... && go test ./... -count=1 -race`
- [ ] T050 Measure LOC reduction: before vs after
- [ ] T051 Git commit: "refactor: remove old architecture (domain, repository, handlers, monolithic service)"
- [ ] T052 Notify team: announce migration completion in Slack/email
- [ ] T053 Update development onboarding docs: remove old architecture references

**Checkpoint**: Documentation updated, team notified, migration complete

---

## Dependencies & Execution Order

### Phase Dependencies

- **Pre-Deletion Audit (Phase 1)**: No dependencies - must complete first (safety check)
- **Delete Domain (Phase 2)**: Depends on Phase 1 audit
- **Delete Repository (Phase 3)**: Depends on Phase 2, must run sequentially
- **Delete Handlers (Phase 4)**: Depends on Phase 3, must run sequentially
- **Delete UserService (Phase 5)**: Depends on Phase 4, must run sequentially
- **Documentation (Phase 6)**: Depends on all deletion phases complete

### Why Sequential Deletion?

Phases 2-5 must run sequentially because:
1. Old repository might depend on old domain
2. Old handlers might depend on old repository
3. Old UserService might depend on old handlers

Deleting in order (domain → repository → handlers → use case) ensures clean deletion without build failures.

## Notes

- Commit after each successful deletion phase
- Run tests after every file deletion
- Keep audit report for historical reference
- Git history preserves deleted code (can be recovered if needed)
- Verify HTTP endpoints work after handler deletion (manual testing)
- Document any edge cases discovered during deletion
- This is a one-way operation - thorough testing required before pushing
- Consider creating feature branch for this refactor
- Estimated time: 4-6 hours including testing and documentation
