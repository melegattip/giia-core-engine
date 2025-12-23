# Testing Strategy - AI Intelligence Hub

## Overview

This document outlines the testing strategy for the AI Intelligence Hub service, including test coverage goals, testing patterns, and how to run tests.

## Test Coverage Goals

**Target: 80%+ code coverage**

### Coverage by Layer

- **Domain Layer**: 95%+ (critical business logic)
- **Use Cases**: 90%+ (core application logic)
- **Handlers**: 85%+ (event processing)
- **Adapters**: 70%+ (external integrations)
- **Repositories**: 80%+ (data access)

## Testing Structure

```
services/ai-intelligence-hub/
├── internal/
│   ├── core/
│   │   ├── domain/
│   │   │   └── notification_test.go          # Domain entity tests
│   │   ├── mocks/                             # Mock implementations
│   │   │   ├── ai_analyzer_mock.go
│   │   │   ├── notification_repository_mock.go
│   │   │   └── stockout_analyzer_mock.go
│   │   └── usecases/
│   │       ├── analysis/
│   │       │   └── analyze_stockout_risk_test.go
│   │       └── event_processing/
│   │           └── buffer_event_handler_test.go
│   └── testhelpers/
│       └── fixtures.go                        # Test data helpers
```

## Test Categories

### 1. Unit Tests

**Location**: `*_test.go` files next to the code being tested

**Purpose**: Test individual functions and methods in isolation

**Examples**:
- Domain entity behavior (creation, state transitions)
- Use case business logic
- Event handler routing

**Running**:
```bash
make test
```

### 2. Integration Tests

**Location**: `internal/integration_tests/`

**Purpose**: Test interaction between components with real dependencies

**Examples**:
- Repository with real database
- NATS event publishing/subscribing
- AI client with mock HTTP server

**Running**:
```bash
go test -tags=integration ./internal/integration_tests/...
```

### 3. End-to-End Tests

**Location**: `e2e_tests/`

**Purpose**: Test complete workflows from event to notification storage

**Running**:
```bash
go test -tags=e2e ./e2e_tests/...
```

## Testing Patterns

### Using Mocks

We use `testify/mock` for creating mock implementations of interfaces.

**Example**:
```go
mockRepo := mocks.NewMockNotificationRepository()
mockRepo.On("Create", ctx, mock.AnythingOfType("*domain.AINotification")).
    Return(nil)
```

### Test Fixtures

Use test helpers from `internal/testhelpers` to create consistent test data:

```go
event := testhelpers.CreateBufferBelowMinimumEvent("PROD-123", 50.0, 100.0, 10.0)
notif := testhelpers.CreateTestNotification()
aiResponse := testhelpers.CreateTestAIResponse()
```

### Table-Driven Tests

For testing multiple scenarios:

```go
tests := []struct {
    name     string
    input    float64
    expected domain.NotificationPriority
}{
    {"Critical", 1.5, domain.NotificationPriorityCritical},
    {"High", 3.0, domain.NotificationPriorityHigh},
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        // Test logic
    })
}
```

## What We Test

### Domain Layer (`internal/core/domain`)

✅ **Tested**:
- Entity creation with NewNotification()
- State transitions (MarkAsRead, MarkAsActedUpon, Dismiss)
- All notification types, priorities, and statuses
- Complete notification workflow
- Impact assessment structure
- Recommendations structure

### Use Cases (`internal/core/usecases`)

✅ **Tested**:
- **analyze_stockout_risk.go**:
  - Successful execution with full workflow
  - Error handling (nil event, missing data, invalid org ID)
  - RAG failure doesn't stop processing
  - AI analyzer failures
  - Repository failures
  - Priority determination logic
  - Time to impact conversion
  - Complete notification structure creation

### Event Handlers (`internal/core/usecases/event_processing`)

✅ **Tested**:
- **buffer_event_handler.go**:
  - Handles buffer.below_minimum events
  - Ignores other buffer events
  - Error propagation from analyzer
  - Multiple events processing

### Mocks (`internal/core/mocks`)

✅ **Created**:
- MockAIAnalyzer
- MockRAGKnowledge
- MockNotificationRepository
- MockStockoutRiskAnalyzer

## Running Tests

### All Tests
```bash
make test
```

### With Coverage
```bash
make test-coverage
# Opens coverage.html in browser
```

### Specific Package
```bash
go test -v ./internal/core/domain/...
go test -v ./internal/core/usecases/analysis/...
```

### With Race Detection
```bash
go test -race ./...
```

### Verbose Output
```bash
go test -v ./...
```

## Test Coverage Report

To generate and view coverage:

```bash
make test-coverage
```

This generates:
- `coverage.out` - Coverage data
- `coverage.html` - Visual coverage report

## Continuous Integration

Tests run automatically on:
- Every pull request
- Push to main/develop branches
- Pre-commit hooks (if configured)

**CI Requirements**:
- All tests must pass
- Coverage must be ≥ 80%
- No race conditions

## Pending Tests

### To Be Implemented

- [ ] Integration tests for NotificationRepository with PostgreSQL
- [ ] Integration tests for NATS event subscriber
- [ ] Unit tests for execution_event_handler.go
- [ ] Unit tests for user_event_handler.go
- [ ] Unit tests for claude_client.go
- [ ] Unit tests for simple_knowledge_retriever.go
- [ ] Integration test for complete event-to-notification flow
- [ ] Performance tests (load testing)
- [ ] Benchmarks for critical paths

## Best Practices

### ✅ DO

- Mock external dependencies (database, NATS, AI API)
- Use table-driven tests for multiple scenarios
- Test both success and error paths
- Use descriptive test names
- Assert all relevant fields
- Use testhelpers for common test data
- Test edge cases and boundary conditions

### ❌ DON'T

- Test implementation details
- Create tests that depend on external services (use mocks)
- Write tests that depend on execution order
- Ignore test failures
- Skip error handling tests
- Use production data in tests

## Troubleshooting

### Tests Failing Locally

1. **Check dependencies**: `go mod download && go mod tidy`
2. **Clean build cache**: `go clean -testcache`
3. **Check for race conditions**: `go test -race ./...`

### Mock Expectations Not Met

```go
// Make sure all expected calls are defined
mockRepo.On("Create", ctx, mock.Anything).Return(nil)

// And that they're asserted
mockRepo.AssertExpectations(t)
```

### Coverage Too Low

1. Check which files are not covered: `go tool cover -html=coverage.out`
2. Add tests for untested code paths
3. Focus on error handling and edge cases

## Examples

See test files for complete examples:
- `internal/core/domain/notification_test.go`
- `internal/core/usecases/analysis/analyze_stockout_risk_test.go`
- `internal/core/usecases/event_processing/buffer_event_handler_test.go`

## Resources

- [Go Testing Package](https://pkg.go.dev/testing)
- [Testify Documentation](https://github.com/stretchr/testify)
- [Table Driven Tests](https://go.dev/wiki/TableDrivenTests)
- [Go Test Coverage](https://go.dev/blog/cover)

---

**Last Updated**: 2025-12-23
**Coverage Status**: Tests Implemented for Core Functionality
**Next Milestone**: 80%+ Coverage Goal
