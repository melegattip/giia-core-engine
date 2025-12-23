# Task 13: Shared Packages Testing - Specification

**Task ID**: task-13-shared-packages-testing
**Phase**: 2A - Complete to 100%
**Priority**: P1 (High)
**Estimated Completion**: 15% remaining work on shared packages
**Dependencies**: Task 4 (85% complete)

---

## Overview

Complete the shared infrastructure packages (pkg/) by implementing comprehensive integration tests with real infrastructure (PostgreSQL, Redis, NATS), adding missing unit tests, and ensuring production-readiness. This task brings shared packages from 85% to 100% completion.

---

## User Scenarios

### US1: Database Package Integration Tests (P1)

**As a** service developer
**I want to** integration tests for the database package with real PostgreSQL
**So that** I can trust database connections, transactions, and retry logic work correctly

**Acceptance Criteria**:
- Integration tests with real PostgreSQL (Docker container)
- Test connection establishment and pooling
- Test transaction commit and rollback
- Test retry logic with temporary failures
- Test health check functionality
- Test connection timeout scenarios

**Success Metrics**:
- 90%+ integration test coverage for database package
- All tests pass consistently

---

### US2: Events Package Integration Tests (P1)

**As a** service developer
**I want to** integration tests for the events package with real NATS Jetstream
**So that** I can trust event publishing, subscribing, and stream management work correctly

**Acceptance Criteria**:
- Integration tests with real NATS Jetstream (Docker container)
- Test event publishing and receiving
- Test stream creation and management
- Test consumer groups and acknowledgments
- Test dead letter queue handling
- Test connection retry logic

**Success Metrics**:
- 90%+ integration test coverage for events package
- All tests pass consistently

---

### US3: Config Package Unit Tests (P2)

**As a** service developer
**I want to** comprehensive unit tests for the config package
**So that** I can trust configuration loading and validation works correctly

**Acceptance Criteria**:
- Test loading config from environment variables
- Test loading config from files (.env, .yaml)
- Test configuration validation
- Test missing required configuration handling
- Test default values

**Success Metrics**:
- 85%+ unit test coverage for config package

---

### US4: Logger Package Additional Tests (P3)

**As a** service developer
**I want to** additional tests for the logger package
**So that** I can ensure structured logging works in all scenarios

**Acceptance Criteria**:
- Test logging at different levels (Debug, Info, Warn, Error)
- Test structured tags in log output
- Test context extraction (request IDs)
- Test log formatting (JSON output)

**Success Metrics**:
- 90%+ test coverage for logger package

---

## Functional Requirements

### FR1: Integration Test Infrastructure
- Docker Compose setup for test infrastructure (PostgreSQL, Redis, NATS)
- Test fixtures and helper functions
- Cleanup after tests (drop databases, purge streams)
- Parallel test execution support

### FR2: Database Integration Tests
- Test suite covering all database.go functions
- Connection pooling tests with concurrent requests
- Transaction isolation tests
- Retry logic tests with simulated failures
- Health check tests with database down scenarios

### FR3: Events Integration Tests
- Test suite covering publisher.go and subscriber.go
- Stream creation and deletion tests
- Event publishing with acknowledgment
- Event subscribing with consumer groups
- Dead letter queue tests
- Connection retry with NATS server restarts

### FR4: Config Unit Tests
- Test environment variable parsing
- Test file loading (YAML, ENV formats)
- Test nested configuration structures
- Test validation with missing/invalid values

### FR5: Logger Unit Tests
- Test log output format and structure
- Test log level filtering
- Test context propagation
- Test error logging with stack traces

---

## Key Test Scenarios

### Database Package
```go
// Test connection retry
func TestDatabaseRetryOnFailure(t *testing.T) {
    // Start PostgreSQL
    // Stop PostgreSQL briefly
    // Verify retry attempts
    // Start PostgreSQL again
    // Verify successful connection
}

// Test transaction rollback
func TestDatabaseTransactionRollback(t *testing.T) {
    // Begin transaction
    // Insert record
    // Simulate error
    // Rollback
    // Verify record not persisted
}
```

### Events Package
```go
// Test event publish and subscribe
func TestEventPublishSubscribe(t *testing.T) {
    // Start NATS
    // Create stream
    // Publish event
    // Subscribe to stream
    // Verify event received
    // Verify acknowledgment
}

// Test dead letter queue
func TestEventDeadLetterQueue(t *testing.T) {
    // Publish event
    // Simulate processing failure
    // Verify max retry attempts
    // Verify event moved to DLQ
}
```

---

## Non-Functional Requirements

### Test Reliability
- Tests must pass consistently (99%+ pass rate)
- No flaky tests due to timing issues
- Proper cleanup to prevent test pollution

### Test Performance
- Integration test suite completes in <5 minutes
- Unit test suite completes in <30 seconds
- Parallel execution support

### Test Coverage
- pkg/database: 90%+ integration test coverage
- pkg/events: 90%+ integration test coverage
- pkg/config: 85%+ unit test coverage
- pkg/logger: 90%+ unit test coverage
- pkg/errors: 100% (already complete)

### CI/CD Integration
- All tests run in GitHub Actions
- Docker Compose for test infrastructure
- Test results and coverage reports

---

## Success Criteria

### Mandatory (Must Have)
- ✅ Database integration tests with real PostgreSQL
- ✅ Events integration tests with real NATS Jetstream
- ✅ Config unit tests with various input formats
- ✅ Logger unit tests with output verification
- ✅ All tests pass in CI/CD pipeline
- ✅ Overall shared package test coverage >85%
- ✅ Test infrastructure automated with Docker Compose

### Optional (Nice to Have)
- ⚪ Performance benchmarks for database operations
- ⚪ Performance benchmarks for event publishing
- ⚪ Load testing for high-volume scenarios
- ⚪ Chaos engineering tests (random failures)

---

## Out of Scope

- ❌ End-to-end testing across services - Separate task
- ❌ Performance optimization - Future task after metrics
- ❌ Redis package testing - Will be added when more Redis usage
- ❌ Tracing integration - Future observability task

---

## Dependencies

- **Task 4**: Shared Packages at 85% (all packages implemented)
- **Infrastructure**: Docker, Docker Compose for test containers
- **CI/CD**: GitHub Actions for automated testing

---

## Risk Assessment

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Flaky tests due to async operations | Medium | Medium | Proper synchronization, timeouts, retries |
| Docker not available in CI | High | Low | Use GitHub Actions services, fallback to mocks |
| Test suite too slow | Medium | Medium | Parallel execution, optimize setup/teardown |
| Resource leaks in tests | Medium | Medium | Proper cleanup, defer statements, test isolation |

---

## References

- [Task 4 Spec](../task-04-shared-packages/spec.md) - Shared packages foundation
- [Go Testing Documentation](https://pkg.go.dev/testing)
- [Testcontainers Go](https://github.com/testcontainers/testcontainers-go) - Docker containers for tests

---

**Document Version**: 1.0
**Last Updated**: 2025-12-16
**Status**: Ready for Planning
**Next Step**: Create implementation plan (plan.md)