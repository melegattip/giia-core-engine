# Feature Specification: Remove Dead Code from Incomplete Migration

**Created**: 2025-12-10
**Priority**: 🟠 HIGH
**Effort**: 1 day

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Delete Old Domain Layer (Priority: P1)

As a developer, I need the old `internal/domain/` directory removed so that there's no confusion about which User entity implementation is canonical.

**Why this priority**: Two parallel User entities (old uint-based vs new UUID-based) create confusion and risk of changes being made to wrong implementation. Critical for code clarity.

**Independent Test**: Can be tested by verifying `internal/domain/` directory no longer exists and no imports reference it.

**Acceptance Scenarios**:

1. **Scenario**: Old domain directory deleted
   - **Given** Old architecture exists in `internal/domain/`
   - **When** Cleanup is performed
   - **Then** Directory `internal/domain/` no longer exists

2. **Scenario**: No imports reference old domain
   - **Given** All codebase files
   - **When** Searching for imports of `internal/domain`
   - **Then** Zero matches found (`grep -r "internal/domain"` returns empty)

3. **Scenario**: New domain is only implementation
   - **Given** Codebase after cleanup
   - **When** Looking for User entity
   - **Then** Only `internal/core/domain/user.go` exists (UUID-based)

---

### User Story 2 - Delete Old Repository Layer (Priority: P1)

As a developer, I need the old `internal/repository/` directory removed so that only GORM-based repositories are used consistently.

**Why this priority**: Two parallel repository implementations (raw SQL vs GORM) create maintenance burden and risk of inconsistent data access patterns.

**Independent Test**: Verify `internal/repository/` directory deleted and only `internal/infrastructure/repositories/` exists.

**Acceptance Scenarios**:

1. **Scenario**: Old repository directory deleted
   - **Given** Old raw SQL repositories in `internal/repository/`
   - **When** Cleanup is performed
   - **Then** Directory `internal/repository/` no longer exists

2. **Scenario**: No imports reference old repository
   - **Given** All codebase files
   - **When** Searching for imports of `internal/repository`
   - **Then** Zero matches found

3. **Scenario**: Only GORM repositories remain
   - **Given** Codebase after cleanup
   - **When** Looking for UserRepository
   - **Then** Only `internal/infrastructure/repositories/user_repository.go` exists (GORM-based)

---

### User Story 3 - Delete Old HTTP Handlers (Priority: P1)

As a developer, I need the old `internal/handlers/` directory removed so that only new Clean Architecture entrypoints are used.

**Why this priority**: Two parallel handler implementations create confusion about which HTTP routes are active. Important for API consistency.

**Independent Test**: Verify `internal/handlers/` directory deleted and only `internal/infrastructure/entrypoints/http/handlers/` exists.

**Acceptance Scenarios**:

1. **Scenario**: Old handlers directory deleted
   - **Given** Old HTTP handlers in `internal/handlers/`
   - **When** Cleanup is performed
   - **Then** Directory `internal/handlers/` no longer exists

2. **Scenario**: No routes reference old handlers
   - **Given** Main application setup
   - **When** Checking route registration
   - **Then** Only new entrypoint handlers are registered

3. **Scenario**: HTTP API still works after cleanup
   - **Given** Application running after cleanup
   - **When** Making HTTP requests to API endpoints
   - **Then** All endpoints respond correctly using new handlers

---

### User Story 4 - Delete Monolithic UserService (Priority: P2)

As a developer, I need the old monolithic `internal/usecases/user_service.go` file removed so that only focused use case files are maintained.

**Why this priority**: 356-line monolithic service contradicts Clean Architecture principle of single responsibility. Can be deleted after verifying all functionality migrated to specific use cases.

**Independent Test**: Verify `internal/usecases/user_service.go` deleted and all user operations work via new use cases.

**Acceptance Scenarios**:

1. **Scenario**: Monolithic UserService file deleted
   - **Given** Old UserService with all user operations
   - **When** Cleanup is performed
   - **Then** File `internal/usecases/user_service.go` no longer exists

2. **Scenario**: All operations migrated to specific use cases
   - **Given** User operations (login, register, update profile, etc.)
   - **When** Checking implementation location
   - **Then** Each operation has dedicated use case file in `internal/core/usecases/auth/`

3. **Scenario**: No functionality lost
   - **Given** Application after UserService deletion
   - **When** Testing all user-related features
   - **Then** All features work via new use case implementations

---

### Edge Cases

- What happens if deleted code is still imported somewhere (build should fail - verify with `go build`)?
- How to verify no functionality is lost (run full test suite after deletion)?
- What happens to git history (code remains in history, can be recovered if needed)?
- How to handle documentation references to old structure (update README and docs)?
- What if some code in old files is not in new files (audit first, migrate missing code)?
- How to communicate changes to team (create migration guide, announce in team chat)?

## Requirements *(mandatory)*

### Functional Requirements

#### File Deletion
- **FR-001**: Directory `internal/domain/` MUST be completely deleted
- **FR-002**: Directory `internal/repository/` MUST be completely deleted
- **FR-003**: Directory `internal/handlers/` MUST be completely deleted
- **FR-004**: File `internal/usecases/user_service.go` MUST be deleted
- **FR-005**: Empty parent directories MUST be deleted after cleanup

#### Import Verification
- **FR-006**: Zero imports of `internal/domain` MUST exist (verified by `grep -r "internal/domain"`)
- **FR-007**: Zero imports of `internal/repository` MUST exist (verified by `grep -r "internal/repository"`)
- **FR-008**: Zero imports of `internal/handlers` MUST exist
- **FR-009**: Zero imports of `internal/usecases/user_service` MUST exist
- **FR-010**: Build MUST succeed after deletion (`go build ./...` passes)

#### Migration Verification
- **FR-011**: All User entity usages MUST use `internal/core/domain.User` (UUID-based)
- **FR-012**: All repository usages MUST use `internal/infrastructure/repositories.*`
- **FR-013**: All HTTP handlers MUST use `internal/infrastructure/entrypoints/http/handlers.*`
- **FR-014**: All use case operations MUST use focused use cases in `internal/core/usecases/*`
- **FR-015**: Test suite MUST pass after deletion (no broken tests)

#### Documentation
- **FR-016**: README MUST be updated to remove references to old structure
- **FR-017**: Architecture diagrams MUST reflect only new structure
- **FR-018**: CHANGELOG MUST document migration completion
- **FR-019**: Migration guide MUST be created for team reference
- **FR-020**: Code comments referring to old structure MUST be updated

### Key Entities

- **Old Domain Layer**: `internal/domain/` - uint-based entities (to be deleted)
- **New Domain Layer**: `internal/core/domain/` - UUID-based entities (canonical)
- **Old Repository Layer**: `internal/repository/` - raw SQL (to be deleted)
- **New Repository Layer**: `internal/infrastructure/repositories/` - GORM (canonical)
- **Old Handlers**: `internal/handlers/` (to be deleted)
- **New Entrypoints**: `internal/infrastructure/entrypoints/http/handlers/` (canonical)
- **Monolithic Service**: `internal/usecases/user_service.go` (to be deleted)
- **Focused Use Cases**: `internal/core/usecases/auth/*` (canonical)

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Directories `internal/domain/`, `internal/repository/`, `internal/handlers/` do not exist
- **SC-002**: File `internal/usecases/user_service.go` does not exist
- **SC-003**: Total LOC reduced by approximately 650 lines (92+55+150+356)
- **SC-004**: Zero imports reference deleted code (verified by `grep`)
- **SC-005**: All tests pass after deletion (`go test ./... -count=1`)
- **SC-006**: Application builds successfully (`go build ./...`)
- **SC-007**: Application runs successfully and all endpoints respond
- **SC-008**: Code review confirms no functionality lost
- **SC-009**: README and documentation updated
- **SC-010**: Team notified and migration completion documented in CHANGELOG
