# Implementation Plan: Enforce Typed Errors Throughout Codebase

**Date**: 2025-12-10
**Spec**: [spec.md](./spec.md)

## Summary

Replace all 169 occurrences of `fmt.Errorf` with typed errors from `pkg/errors` to enable proper HTTP status code mapping and structured error responses. Focus on repository layer (95% violations), use case layer (40% violations), and infrastructure adapters (90% violations). Violates CLAUDE.md guideline: "NO fmt.Errorf: Prefer typed errors over generic wrapping."

## Technical Context

**Language/Version**: Go 1.23.4
**Primary Dependencies**: pkg/errors (custom error types), GORM, net/http
**Storage**: PostgreSQL 16 (error mapping for GORM errors)
**Testing**: Existing test suite + new error type tests
**Target Platform**: Linux/Windows server
**Project Type**: Microservice backend
**Performance Goals**: Error construction <1ms, no performance degradation
**Constraints**: Must preserve error context, map GORM errors correctly, maintain HTTP status codes
**Scale/Scope**: 169 fmt.Errorf occurrences across 14 files

## Project Structure

### Source Code (auth-service)

```text
services/auth-service/
├── internal/
│   ├── infrastructure/
│   │   ├── repositories/
│   │   │   ├── user_repository.go           # MODIFY - 95% fmt.Errorf
│   │   │   ├── token_repository.go          # MODIFY - 95% fmt.Errorf
│   │   │   ├── role_repository.go           # MODIFY - 95% fmt.Errorf
│   │   │   └── permission_repository.go     # MODIFY - 95% fmt.Errorf
│   │   └── adapters/
│   │       ├── jwt/jwt_manager.go           # MODIFY - 90% fmt.Errorf
│   │       ├── cache/permission_cache.go    # MODIFY - 90% fmt.Errorf
│   │       └── email/email_service.go       # MODIFY - 90% fmt.Errorf
│   └── core/
│       └── usecases/
│           ├── auth/*.go                     # MODIFY - 40% fmt.Errorf
│           ├── rbac/*.go                     # MODIFY - 40% fmt.Errorf
│           └── role/*.go                     # MODIFY - 40% fmt.Errorf
├── pkg/
│   ├── errors/
│   │   ├── errors.go                         # VERIFY - typed error system exists
│   │   └── errors_test.go                    # NEW - test error type verification
│   └── database/
│       └── connection.go                     # MODIFY - fmt.Errorf usage
├── docs/
│   └── error-handling-guide.md               # NEW - error handling documentation
└── .golangci.yml                             # NEW - custom rule to ban fmt.Errorf
```

---

## Phase 1: Setup & Verification

**Purpose**: Verify error system and create enforcement rules

- [ ] T001 Verify pkg/errors has all required constructors (NewBadRequest, NewUnauthorizedRequest, NewForbidden, NewResourceNotFound, NewConflict, NewTooManyRequests, NewInternalServerError)
- [ ] T002 Create errors_test.go to verify error type checking with errors.As()
- [ ] T003 Document GORM error mapping strategy (ErrRecordNotFound → NewResourceNotFound, ErrDuplicatedKey → NewConflict, other → NewInternalServerError)
- [ ] T004 Create .golangci.yml with custom rule to ban fmt.Errorf in internal/core/* and internal/infrastructure/repositories/*
- [ ] T005 Document error handling guidelines in docs/error-handling-guide.md

---

## Phase 2: Repository Layer - UserRepository (Priority: P1)

**Goal**: Replace all fmt.Errorf in UserRepository with typed errors

**Independent Test**: Verify `grep "fmt.Errorf" internal/infrastructure/repositories/user_repository.go` returns zero matches

- [ ] T006 [P] [US1] Update Create method: map GORM errors to typed errors
- [ ] T007 [P] [US1] Update FindByID method: ErrRecordNotFound → NewResourceNotFound
- [ ] T008 [P] [US1] Update FindByEmail method: map errors appropriately
- [ ] T009 [P] [US1] Update Update method: map constraint violations
- [ ] T010 [P] [US1] Update Delete method: handle not found and cascade errors
- [ ] T011 [P] [US1] Update List method: handle query errors
- [ ] T012 [P] [US1] Test UserRepository: verify error types returned (use errors.As)
- [ ] T013 [P] [US1] Verify zero fmt.Errorf occurrences in user_repository.go

**Checkpoint**: UserRepository returns only typed errors, tests verify error types

---

## Phase 3: Repository Layer - Other Repositories (Priority: P1)

**Goal**: Replace all fmt.Errorf in remaining repositories

- [ ] T014 [P] [US1] Update TokenRepository methods with typed errors
- [ ] T015 [P] [US1] Update RoleRepository methods with typed errors
- [ ] T016 [P] [US1] Update PermissionRepository methods with typed errors
- [ ] T017 [P] [US1] Add error type tests for all repositories
- [ ] T018 [P] [US1] Verify zero fmt.Errorf in internal/infrastructure/repositories/*

**Checkpoint**: All repositories use typed errors, zero fmt.Errorf violations

---

## Phase 4: Use Case Layer - Auth Use Cases (Priority: P1)

**Goal**: Replace all fmt.Errorf in auth use cases

- [ ] T019 [P] [US2] Update LoginUseCase: use NewBadRequest for validation, NewUnauthorizedRequest for auth failures
- [ ] T020 [P] [US2] Update RegisterUseCase: use NewBadRequest for validation, preserve repository errors
- [ ] T021 [P] [US2] Update RefreshTokenUseCase: use NewUnauthorizedRequest for token errors
- [ ] T022 [P] [US2] Update ValidateTokenUseCase: map JWT errors to typed errors
- [ ] T023 [P] [US2] Add error type tests for auth use cases
- [ ] T024 [P] [US2] Verify zero fmt.Errorf in internal/core/usecases/auth/*

**Checkpoint**: Auth use cases use typed errors appropriately

---

## Phase 5: Use Case Layer - RBAC & Role Use Cases (Priority: P1)

**Goal**: Replace all fmt.Errorf in RBAC and role use cases

- [ ] T025 [P] [US2] Update CheckPermissionUseCase: use NewForbidden for authorization failures
- [ ] T026 [P] [US2] Update GetUserPermissionsUseCase: preserve repository errors
- [ ] T027 [P] [US2] Update role management use cases: use NewBadRequest for validation
- [ ] T028 [P] [US2] Add error type tests for RBAC/role use cases
- [ ] T029 [P] [US2] Verify zero fmt.Errorf in internal/core/usecases/rbac/* and internal/core/usecases/role/*

**Checkpoint**: All use cases use typed errors, zero fmt.Errorf violations in core

---

## Phase 6: Infrastructure Adapters (Priority: P2)

**Goal**: Replace all fmt.Errorf in infrastructure adapters

- [ ] T030 [US3] Update JWTManager: use NewUnauthorizedRequest for token validation failures, NewInternalServerError for signing errors
- [ ] T031 [US3] Update RedisPermissionCache: use NewInternalServerError for Redis errors
- [ ] T032 [US3] Update EmailService: use NewInternalServerError for SMTP errors
- [ ] T033 [US3] Update RateLimiter: use NewTooManyRequests for rate limit violations
- [ ] T034 [US3] Add error type tests for adapters
- [ ] T035 [US3] Verify zero fmt.Errorf in internal/infrastructure/adapters/*

**Checkpoint**: All adapters use typed errors

---

## Phase 7: Shared Packages (Priority: P2)

**Goal**: Replace fmt.Errorf in pkg/*

- [ ] T036 [US3] Update pkg/database/connection.go: use NewInternalServerError for connection errors
- [ ] T037 [US3] Update any other pkg/* files with fmt.Errorf usage
- [ ] T038 [US3] Verify zero fmt.Errorf in pkg/*

**Checkpoint**: Shared packages use typed errors

---

## Phase 8: HTTP/gRPC Error Response Consistency (Priority: P2)

**Goal**: Ensure consistent error responses across all endpoints

- [ ] T039 [US4] Verify HTTP handlers extract HTTPStatus from typed errors
- [ ] T040 [US4] Verify gRPC handlers map typed errors to correct gRPC codes
- [ ] T041 [US4] Test error response JSON structure (error_code, message, http_status)
- [ ] T042 [US4] Verify internal errors don't expose sensitive information
- [ ] T043 [US4] Add integration tests for error responses
- [ ] T044 [US4] Document error response format in API documentation

**Checkpoint**: All error responses follow consistent format

---

## Phase 9: Enforcement & Documentation

**Purpose**: Prevent future violations and document error handling

- [ ] T045 Run golangci-lint with custom rule: verify fmt.Errorf blocked
- [ ] T046 Update CLAUDE.md with error handling examples (before/after)
- [ ] T047 Create error code reference table in docs/
- [ ] T048 Add error handling guidelines to developer onboarding
- [ ] T049 Code review checklist: verify typed errors used
- [ ] T050 Verify all tests pass after error type migration
- [ ] T051 Run full test suite with coverage: ensure no regressions
- [ ] T052 Performance test: verify error construction <1ms
- [ ] T053 Update CI/CD to run golangci-lint with fmt.Errorf ban
- [ ] T054 Team training: present error handling guidelines

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - must complete first
- **Repository Layers (Phase 2-3)**: Depends on Phase 1, can run in parallel
- **Use Case Layers (Phase 4-5)**: Depends on Phase 2-3 (repositories must return typed errors first)
- **Adapters (Phase 6)**: Can run parallel with Phase 4-5
- **Shared Packages (Phase 7)**: Can run parallel with Phase 2-6
- **Error Responses (Phase 8)**: Depends on Phase 4-5 completion
- **Enforcement (Phase 9)**: Depends on all previous phases

### Execution Strategy

1. Complete Phase 1 (Setup)
2. Run Phase 2-3 in parallel (repositories)
3. Run Phase 4-6 in parallel (use cases and adapters)
4. Complete Phase 7 (shared packages)
5. Complete Phase 8 (error responses)
6. Complete Phase 9 (enforcement)

## Notes

- Each file modification should include corresponding test updates
- Verify error type with `errors.As()` in tests
- Preserve error context while using typed errors
- Document error mapping strategy for each layer
- Run tests after each file modification
- Commit after each phase completion
- Use golangci-lint to prevent regressions
- Error messages should be actionable for API consumers
