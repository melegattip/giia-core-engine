# Feature Specification: TimeManager for All Date Operations

**Created**: 2025-12-10
**Priority**: 🔴 CRITICAL
**Effort**: 1-2 days

## User Scenarios & Testing *(mandatory)*

### User Story 1 - TimeManager Interface and Implementation (Priority: P1)

As a developer, I need a TimeManager interface with production and mock implementations so that all date operations are testable and consistent.

**Why this priority**: Foundation for all time-related testing. Without it, tests use real wall-clock time, making them non-deterministic and slow. Violates CLAUDE.md mandate.

**Independent Test**: Can be tested by instantiating RealTimeManager and MockTimeManager and verifying both implement the interface correctly.

**Acceptance Scenarios**:

1. **Scenario**: RealTimeManager returns current UTC time
   - **Given** RealTimeManager instance
   - **When** calling `Now()`
   - **Then** returns current time in UTC timezone

2. **Scenario**: MockTimeManager returns fixed test time
   - **Given** MockTimeManager configured with fixed time
   - **When** calling `Now()`
   - **Then** returns configured test time

3. **Scenario**: TimeManager provides ISO8601 formatting
   - **Given** TimeManager with specific date
   - **When** calling `FormatToISO8601(date)`
   - **Then** returns string in format "2024-01-15T12:30:00Z"

---

### User Story 2 - Inject TimeManager into All Use Cases (Priority: P1)

As a developer, I need TimeManager injected into all use cases so that token expiry, timestamps, and time-based logic are testable.

**Why this priority**: Use cases have 5+ direct `time.Now()` calls. Cannot properly test token expiry, refresh logic, or time-based workflows without mockable time.

**Independent Test**: Verify all use case constructors accept TimeManager and use it instead of direct `time.Now()` calls.

**Acceptance Scenarios**:

1. **Scenario**: LoginUseCase uses TimeManager for token expiry
   - **Given** LoginUseCase with MockTimeManager set to 2024-01-15 12:00:00
   - **When** generating refresh token with 7-day expiry
   - **Then** token expires at 2024-01-22 12:00:00 (deterministic)

2. **Scenario**: ValidateTokenUseCase uses TimeManager for expiry check
   - **Given** ValidateTokenUseCase with MockTimeManager
   - **When** validating token that expires at 2024-01-20
   - **Then** test can verify both valid (before expiry) and invalid (after expiry) scenarios by adjusting mock time

3. **Scenario**: RegisterUseCase uses TimeManager for created_at timestamp
   - **Given** RegisterUseCase with MockTimeManager set to fixed time
   - **When** creating new user
   - **Then** user.CreatedAt equals mock time (enables deterministic assertions)

---

### User Story 3 - Inject TimeManager into Infrastructure Adapters (Priority: P1)

As a developer, I need TimeManager injected into JWT manager and other adapters so that token generation and validation are testable.

**Why this priority**: JWTManager, TokenRepository, and other adapters use direct `time.Now()`. Critical for security testing (token expiry, replay attacks).

**Independent Test**: Verify all adapters use TimeManager for time operations.

**Acceptance Scenarios**:

1. **Scenario**: JWTManager generates tokens with mockable expiry
   - **Given** JWTManager with MockTimeManager
   - **When** generating access token with 15-minute expiry
   - **Then** token exp claim equals mock time + 15 minutes (verifiable in tests)

2. **Scenario**: TokenRepository cleanup uses TimeManager
   - **Given** TokenRepository with expired tokens
   - **When** cleanup job runs with MockTimeManager
   - **Then** test can control "current time" to verify exact cleanup behavior

3. **Scenario**: RateLimiter window uses TimeManager
   - **Given** RateLimiter with 1-hour window
   - **When** checking rate limit with MockTimeManager
   - **Then** test can advance time to verify window reset logic

---

### User Story 4 - Replace time.Sleep in Tests with Mock Time (Priority: P2)

As a developer, I need tests to use mock time progression instead of `time.Sleep()` so that tests run fast and deterministically.

**Why this priority**: Tests currently use `time.Sleep(10 * time.Millisecond)` which slows test execution and adds non-determinism. Important for test performance but lower priority than implementation.

**Independent Test**: Run test suite and verify no `time.Sleep()` calls exist (except in concurrency tests where required).

**Acceptance Scenarios**:

1. **Scenario**: Token expiry test uses mock time progression
   - **Given** Test validating expired token
   - **When** Test advances MockTimeManager by required duration
   - **Then** Test completes instantly (no sleep) and deterministically

2. **Scenario**: Cache TTL test uses mock time
   - **Given** Test verifying cache expiry
   - **When** MockTimeManager advances past TTL
   - **Then** Cache correctly reports expired entry without waiting

3. **Scenario**: Rate limit window test uses mock time
   - **Given** Test checking rate limit reset
   - **When** MockTimeManager advances to next window
   - **Then** Rate limit correctly resets without real time delay

---

### Edge Cases

- What happens when timezone conversions are needed (always work in UTC)?
- How to handle different date formats (provide multiple format functions)?
- What happens when mock time is not configured in tests (fail fast with clear error)?
- How to test time progression in concurrent scenarios (MockTimeManager thread-safe)?
- How to handle clock skew in distributed systems (use UTC consistently)?
- How to mock time in integration tests (use separate MockTimeManager instance)?
- How to handle daylight saving time (work exclusively in UTC)?

## Requirements *(mandatory)*

### Functional Requirements

#### TimeManager Interface
- **FR-001**: System MUST provide `TimeManager` interface in `pkg/time` package
- **FR-002**: Interface MUST include method `Now() time.Time` (returns current time in UTC)
- **FR-003**: Interface MUST include method `FormatToISO8601(date time.Time) string`
- **FR-004**: Interface MUST include method `StringToUTC(dateString string) (time.Time, error)`
- **FR-005**: Interface MUST include method `FormatToOffset(date time.Time) string`

#### Production Implementation
- **FR-006**: System MUST provide `RealTimeManager` struct implementing TimeManager
- **FR-007**: `RealTimeManager.Now()` MUST return `time.Now().UTC()`
- **FR-008**: `RealTimeManager.FormatToISO8601()` MUST return RFC3339 format with Zulu timezone
- **FR-009**: `RealTimeManager.StringToUTC()` MUST parse common date formats and convert to UTC
- **FR-010**: RealTimeManager MUST be the default implementation in production code

#### Mock Implementation
- **FR-011**: System MUST provide `MockTimeManager` implementing TimeManager and `mock.Mock`
- **FR-012**: MockTimeManager MUST allow configuring fixed time via `On("Now").Return(fixedTime)`
- **FR-013**: MockTimeManager MUST support time progression for sequential calls
- **FR-014**: MockTimeManager MUST be thread-safe for concurrent test execution
- **FR-015**: MockTimeManager MUST be used in 100% of test files (no direct `time.Now()` in tests)

#### Dependency Injection
- **FR-016**: All use case constructors MUST accept TimeManager as parameter
- **FR-017**: All infrastructure adapter constructors MUST accept TimeManager as parameter
- **FR-018**: Main application MUST inject RealTimeManager into all components
- **FR-019**: Test files MUST inject MockTimeManager into all components
- **FR-020**: TimeManager MUST be stored as struct field in all components that use time

#### Code Migration
- **FR-021**: Zero direct `time.Now()` calls in `internal/core/*` (use `tm.Now()`)
- **FR-022**: Zero direct `time.Now()` calls in `internal/infrastructure/*` (use `tm.Now()`)
- **FR-023**: Zero `time.Sleep()` calls in test files (use mock time progression)
- **FR-024**: All time operations MUST go through TimeManager interface
- **FR-025**: Legacy time utilities in `utils` package MUST be deprecated

### Key Entities

- **TimeManager**: Interface for all time operations
- **RealTimeManager**: Production implementation using `time.Now()`
- **MockTimeManager**: Test implementation with configurable fixed time
- **Time Format**: String representation of time (ISO8601, RFC3339, custom formats)
- **UTC Timestamp**: All times stored and compared in UTC timezone

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Zero direct `time.Now()` calls in `internal/core/*` (verified by `grep -r "time.Now()" internal/core`)
- **SC-002**: Zero direct `time.Now()` calls in `internal/infrastructure/*` (verified by grep)
- **SC-003**: Zero `time.Sleep()` calls in `*_test.go` files (verified by grep)
- **SC-004**: 100% of use cases have TimeManager as constructor parameter
- **SC-005**: 100% of tests use MockTimeManager for time-dependent logic
- **SC-006**: Token expiry tests run instantly (<100ms) using mock time
- **SC-007**: golangci-lint custom rule blocks new `time.Now()` usage in prohibited packages
- **SC-008**: CLAUDE.md documentation updated with TimeManager usage examples
- **SC-009**: Test execution time reduces by >50% after removing `time.Sleep()` calls
- **SC-010**: All time-related tests are deterministic (same input produces same output every run)
