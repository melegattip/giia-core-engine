# Implementation Plan: TimeManager for All Date Operations

**Date**: 2025-12-10
**Spec**: [spec.md](./spec.md)

## Summary

Implement TimeManager interface with production and mock implementations, then inject into all components that use time operations. Replace all 24 direct `time.Now()` calls with TimeManager interface to enable testable, deterministic time operations. Violates CLAUDE.md mandate: "MANDATORY: Always use TimeManager for date operations."

## Technical Context

**Language/Version**: Go 1.23.4
**Primary Dependencies**: testify/mock, time (standard library)
**Testing**: MockTimeManager for deterministic tests
**Project Type**: Microservice backend
**Performance Goals**: TimeManager calls <1μs overhead vs direct time.Now()
**Constraints**: Must maintain UTC timezone, support ISO8601 formatting, backward compatible
**Scale/Scope**: 24 time.Now() occurrences across 14 files, all tests need MockTimeManager

## Project Structure

### Source Code

```text
services/auth-service/
├── pkg/
│   └── time/
│       ├── time_manager.go                # NEW - interface + RealTimeManager
│       ├── time_manager_mock.go           # NEW - MockTimeManager
│       └── time_manager_test.go           # NEW - tests for both implementations
├── internal/
│   ├── core/
│   │   └── usecases/
│   │       ├── auth/
│   │       │   ├── login.go               # MODIFY - inject TimeManager
│   │       │   ├── register.go            # MODIFY - inject TimeManager
│   │       │   ├── refresh_token.go       # MODIFY - inject TimeManager
│   │       │   └── validate_token.go      # MODIFY - inject TimeManager
│   │       ├── rbac/
│   │       │   └── get_user_permissions.go # MODIFY - inject TimeManager
│   │       └── role/
│   │           └── *.go                   # MODIFY - if using time operations
│   ├── infrastructure/
│   │   ├── adapters/
│   │   │   ├── jwt/jwt_manager.go         # MODIFY - inject TimeManager (2 time.Now)
│   │   │   └── cache/permission_cache.go  # MODIFY - inject TimeManager
│   │   └── repositories/
│   │       ├── token_repository.go        # MODIFY - inject TimeManager (3 time.Now)
│   │       └── user_repository.go         # MODIFY - if using timestamps
│   └── usecases/
│       └── user_service.go                # MODIFY - inject TimeManager (5 time.Now)
├── cmd/
│   └── api/
│       └── main.go                        # MODIFY - inject RealTimeManager into DI container
└── CLAUDE.md                              # UPDATE - add TimeManager usage examples
```

---

## Phase 1: TimeManager Interface & Implementations

**Purpose**: Create foundational TimeManager interface and implementations

- [ ] T001 Create pkg/time/time_manager.go with TimeManager interface
- [ ] T002 Implement RealTimeManager struct with Now(), FormatToISO8601(), StringToUTC(), FormatToOffset() methods
- [ ] T003 Implement MockTimeManager struct with testify/mock embedding
- [ ] T004 Create time_manager_test.go: test RealTimeManager returns UTC, MockTimeManager returns configured time
- [ ] T005 Verify all TimeManager methods return UTC timezone
- [ ] T006 Document TimeManager interface in pkg/time/README.md

**Checkpoint**: TimeManager interface defined, both implementations tested and working

---

## Phase 2: Inject TimeManager into Use Cases - Auth

**Goal**: Update auth use cases to use TimeManager instead of time.Now()

- [ ] T007 [P] [US2] Update LoginUseCase constructor: add TimeManager parameter
- [ ] T008 [P] [US2] Update LoginUseCase.Execute: replace `time.Now()` with `uc.timeManager.Now()`
- [ ] T009 [P] [US2] Update RegisterUseCase constructor: add TimeManager parameter
- [ ] T010 [P] [US2] Update RegisterUseCase.Execute: replace time operations with TimeManager
- [ ] T011 [P] [US2] Update RefreshTokenUseCase constructor: add TimeManager parameter
- [ ] T012 [P] [US2] Update RefreshTokenUseCase: replace time operations with TimeManager
- [ ] T013 [P] [US2] Update ValidateTokenUseCase constructor: add TimeManager parameter
- [ ] T014 [P] [US2] Update ValidateTokenUseCase: replace time operations with TimeManager
- [ ] T015 [P] [US2] Verify zero time.Now() calls in internal/core/usecases/auth/*

**Checkpoint**: Auth use cases use TimeManager, no direct time.Now() calls

---

## Phase 3: Inject TimeManager into Infrastructure Adapters

**Goal**: Update JWT manager and other adapters to use TimeManager

- [ ] T016 [P] [US3] Update JWTManager constructor: add TimeManager parameter
- [ ] T017 [P] [US3] Update JWTManager.GenerateToken: use timeManager.Now() for expiry calculation
- [ ] T018 [P] [US3] Update JWTManager.ValidateToken: use timeManager.Now() for expiry check
- [ ] T019 [P] [US3] Update TokenRepository constructor: add TimeManager parameter
- [ ] T020 [P] [US3] Update TokenRepository cleanup/expiry methods: use TimeManager
- [ ] T021 [P] [US3] Update PermissionCache if using TTL: use TimeManager
- [ ] T022 [P] [US3] Verify zero time.Now() calls in internal/infrastructure/adapters/*

**Checkpoint**: All adapters use TimeManager, token expiry logic testable

---

## Phase 4: Inject TimeManager into Repositories

**Goal**: Update repositories to use TimeManager for timestamps

- [ ] T023 [US3] Update TokenRepository: replace time.Now() with timeManager.Now()
- [ ] T024 [US3] Update UserRepository if using created_at/updated_at: use TimeManager
- [ ] T025 [US3] Update RoleRepository if using timestamps: use TimeManager
- [ ] T026 [US3] Verify zero time.Now() calls in internal/infrastructure/repositories/*

**Checkpoint**: All repositories use TimeManager for timestamp operations

---

## Phase 5: Update Legacy Code (Old Architecture)

**Goal**: Update old architecture files that still use time.Now()

- [ ] T027 [US3] Update internal/usecases/user_service.go: inject and use TimeManager (5 occurrences)
- [ ] T028 [US3] Update internal/infrastructure/auth/jwt_service.go: use TimeManager (3 occurrences)
- [ ] T029 [US3] Update internal/infrastructure/auth/twofa_service.go: use TimeManager (2 occurrences)
- [ ] T030 [US3] Verify zero time.Now() calls in internal/usecases/* and internal/infrastructure/auth/*

**Checkpoint**: Legacy code migrated (will be deleted in refactor-04)

---

## Phase 6: Update Main Application (Dependency Injection)

**Goal**: Wire RealTimeManager into application DI container

- [ ] T031 Instantiate RealTimeManager in main.go: `realTimeManager := pkgTime.NewRealTimeManager()`
- [ ] T032 Pass timeManager to all use case constructors
- [ ] T033 Pass timeManager to all adapter constructors
- [ ] T034 Pass timeManager to all repository constructors
- [ ] T035 Verify application builds successfully
- [ ] T036 Run application and verify no time-related errors

**Checkpoint**: Application uses RealTimeManager in production

---

## Phase 7: Update All Tests with MockTimeManager

**Goal**: Replace time.Now() and time.Sleep() in all tests

- [ ] T037 [P] [US4] Update login_test.go: inject MockTimeManager, configure fixed time
- [ ] T038 [P] [US4] Update register_test.go: inject MockTimeManager
- [ ] T039 [P] [US4] Update validate_token_test.go: replace time.Sleep with mock time progression
- [ ] T040 [P] [US4] Update refresh_token_test.go: use MockTimeManager for expiry tests
- [ ] T041 [P] [US4] Update check_permission_test.go: inject MockTimeManager if needed
- [ ] T042 [P] [US4] Update jwt_manager_test.go: use MockTimeManager for token expiry tests
- [ ] T043 [P] [US4] Update token_repository_test.go: use MockTimeManager for cleanup tests
- [ ] T044 [P] [US4] Update all other test files: inject MockTimeManager
- [ ] T045 [P] [US4] Verify zero time.Now() calls in *_test.go files
- [ ] T046 [P] [US4] Verify zero time.Sleep() calls in *_test.go files (except concurrency tests)

**Checkpoint**: All tests use MockTimeManager, deterministic test execution

---

## Phase 8: Documentation & Enforcement

**Purpose**: Document TimeManager usage and prevent future violations

- [ ] T047 Update CLAUDE.md: add TimeManager usage section with examples
- [ ] T048 Create pkg/time/README.md: document interface methods, usage patterns
- [ ] T049 Add TimeManager examples to docs/error-handling-guide.md
- [ ] T050 Update .golangci.yml: add custom rule to ban time.Now() in internal/core/* and internal/infrastructure/*
- [ ] T051 Run golangci-lint: verify time.Now() blocked in prohibited packages
- [ ] T052 Verify all tests pass after TimeManager migration
- [ ] T053 Measure test execution time improvement (should be faster without time.Sleep)
- [ ] T054 Code review: verify all time operations go through TimeManager
- [ ] T055 Update CI/CD: enforce golangci-lint rule

**Checkpoint**: TimeManager documented, violations prevented by linter

---

## Dependencies & Execution Order

### Phase Dependencies

- **TimeManager Implementation (Phase 1)**: No dependencies - must complete first
- **Use Cases (Phase 2)**: Depends on Phase 1
- **Adapters (Phase 3)**: Depends on Phase 1, can run parallel with Phase 2
- **Repositories (Phase 4)**: Depends on Phase 1, can run parallel with Phase 2-3
- **Legacy Code (Phase 5)**: Depends on Phase 1, can run parallel with Phase 2-4
- **Main Application (Phase 6)**: Depends on Phase 2-5 completion
- **Tests (Phase 7)**: Depends on Phase 6 (application must work first)
- **Documentation (Phase 8)**: Depends on all previous phases

### Execution Strategy

1. Complete Phase 1 (TimeManager implementation)
2. Run Phase 2-5 in parallel (use cases, adapters, repositories, legacy)
3. Complete Phase 6 (wire into main)
4. Complete Phase 7 (update all tests)
5. Complete Phase 8 (documentation and enforcement)

## Notes

- All times must be in UTC (no timezone conversions in business logic)
- MockTimeManager should support both fixed time and time progression
- Test execution time should improve by >50% after removing time.Sleep()
- TimeManager overhead should be negligible (<1μs per call)
- Use golangci-lint to prevent future time.Now() usage
- Commit after each phase completion
- Verify tests are deterministic (same input → same output every run)
- Document TimeManager mock usage in test examples
