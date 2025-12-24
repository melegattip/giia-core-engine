# Unit Testing Standards for GIIA Project

**Effective Date:** 2025-01-18
**Status:** MANDATORY for all development
**Applies to:** All Claude agents and developers working on this project

---

## Core Requirement

**ALL new features, use cases, and business logic MUST include comprehensive unit tests with a minimum of 85% code coverage.**

This is a non-negotiable requirement for all pull requests and code contributions.

---

## Testing Requirements

### Coverage Standards

1. **Minimum Coverage**: 85% code coverage for all new code
2. **Goal Coverage**: 90%+ for critical business logic
3. **Measurement**: Use `go test -coverprofile` to verify coverage

### What Must Be Tested

✅ **MUST TEST:**
- All public functions and methods in use cases
- All error handling paths
- All validation logic
- Business rule enforcement
- Edge cases and boundary conditions
- Success paths and failure paths
- Integration points with repositories and external services

❌ **DO NOT TEST:**
- Infrastructure code (routers, middleware) - integration tests handle these
- Third-party library code
- Simple getters/setters with no logic
- Generated code (protobuf, mocks)

---

## Test Structure Standards

### File Organization

```
internal/
  core/
    usecases/
      auth/
        request_password_reset.go
        request_password_reset_test.go     # Test file in same package
        confirm_password_reset.go
        confirm_password_reset_test.go
```

### Naming Conventions

```go
// Test function format: TestFunctionName_Scenario_ExpectedBehavior
func TestRequestPasswordResetUseCase_Execute_Success(t *testing.T)
func TestRequestPasswordResetUseCase_Execute_EmptyEmail_ReturnsError(t *testing.T)
func TestRequestPasswordResetUseCase_Execute_UserNotFound_ReturnsSuccessForSecurity(t *testing.T)
```

### Test Structure (Given-When-Then)

```go
func TestFunctionName_Scenario_ExpectedBehavior(t *testing.T) {
    // Given - Setup test data and mocks
    mockRepo := new(providers.MockRepository)
    logger := pkgLogger.New("test", "error")
    useCase := NewUseCase(mockRepo, logger)

    givenInput := &domain.Input{
        Field: "value",
    }

    mockRepo.On("Method", ctx, expectedParam).Return(expectedResult, nil)

    // When - Execute the function under test
    result, err := useCase.Execute(ctx, givenInput)

    // Then - Verify results and expectations
    assert.NoError(t, err)
    assert.Equal(t, expectedResult, result)
    mockRepo.AssertExpectations(t)
}
```

---

## Variable Naming in Tests

### Prefix `given` - Input Data

Use `given` prefix for all input data, mock configuration, and initial conditions:

```go
givenUserID := uuid.New()
givenEmail := "test@example.com"
givenUser := &domain.User{ID: userID}
givenRepositoryError := pkgErrors.NewNotFound("not found")
givenMockResponse := &domain.Result{Success: true}
```

### Prefix `expected` - Expected Results

Use `expected` prefix for results you expect from the test:

```go
expectedError := "invalid input"
expectedResult := &domain.Output{Status: "success"}
expectedRepositoryCalls := 2
expectedCacheWrites := 1
```

---

## Mock Usage

### Use Centralized Mocks

**Always use mocks from `internal/core/providers/mocks.go`:**

```go
// ✅ CORRECT - Use centralized mocks
mockUserRepo := new(providers.MockUserRepository)
mockTokenRepo := new(providers.MockTokenRepository)
mockEventPublisher := new(providers.MockEventPublisher)
mockTimeManager := new(providers.MockTimeManager)

// ❌ INCORRECT - Do not create duplicate mocks in test files
type MockUserRepository struct { ... }
```

### Be Specific with Mock Expectations

```go
// ✅ CORRECT - Specific parameters
mockRepo.On("GetByID", ctx, uuid.MustParse("..."), 123).Return(result, nil)

// ⚠️ USE SPARINGLY - mock.Anything only when value truly doesn't matter
mockRepo.On("Save", mock.AnythingOfType("*domain.User")).Return(nil)

// ❌ AVOID - Being vague with all parameters
mockRepo.On("Method", mock.Anything, mock.Anything).Return(result)
```

### Validate Mock Calls

```go
// ✅ Always assert expectations
mockRepo.AssertExpectations(t)
mockRepo.AssertNumberOfCalls(t, "Save", 1)
mockRepo.AssertNotCalled(t, "Delete")
```

---

## Test Coverage Verification

### Running Tests

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./internal/core/usecases/...

# Run with race detection
go test -race ./...
```

### Measuring Coverage

```bash
# Generate coverage for a package
go test -coverprofile=coverage.out -covermode=atomic ./internal/core/usecases/auth/

# View coverage report
go tool cover -html=coverage.out

# Check coverage percentage
go tool cover -func=coverage.out | grep total
```

### Coverage Enforcement

```bash
# Fail if coverage is below 85%
go test -coverprofile=coverage.out ./internal/core/usecases/... && \
  go tool cover -func=coverage.out | \
  awk '/total:/ {if ($3+0 < 85.0) exit 1}'
```

---

## Test Scenarios to Cover

### For Every Use Case, Test:

1. **Success Path**
   - Happy path with valid inputs
   - Expected result returned
   - All dependencies called correctly

2. **Input Validation**
   - Empty/null required fields
   - Invalid format (email, UUID, etc.)
   - Out of range values
   - Malformed data

3. **Permission/Authorization**
   - User has required permissions
   - User lacks permissions
   - Permission check failures

4. **Business Rule Violations**
   - State transition violations
   - Duplicate operations
   - Conflicting operations

5. **Dependency Failures**
   - Repository errors
   - External service failures
   - Network timeouts
   - Database connection issues

6. **Edge Cases**
   - Already processed items
   - Missing related entities
   - Concurrent modifications
   - Resource not found

7. **Event Publishing** (if applicable)
   - Event published successfully
   - Event publishing fails (should not block operation)

---

## Logger in Tests

**Always use a real logger in tests:**

```go
// ✅ CORRECT - Use real logger
logger := pkgLogger.New("test", "error")

// ❌ INCORRECT - Do not use NoOpLogger
logger := pkgLogger.NewNoOpLogger()  // This function doesn't exist!
```

---

## Pre-Commit Checklist

Before submitting a PR, verify:

- [ ] All new public functions have unit tests
- [ ] All error paths are tested
- [ ] Coverage is ≥ 85% for new code
- [ ] All tests pass: `go test ./...`
- [ ] No race conditions: `go test -race ./...`
- [ ] Mocks use centralized providers
- [ ] Tests follow Given-When-Then structure
- [ ] Variable names use `given`/`expected` prefixes
- [ ] Mock expectations are specific
- [ ] All mocks have `AssertExpectations(t)`

---

## Examples

### Complete Test Example

```go
func TestRequestPasswordResetUseCase_Execute_Success(t *testing.T) {
    // Given
    mockUserRepo := new(providers.MockUserRepository)
    mockTokenRepo := new(providers.MockTokenRepository)
    mockEmailService := new(MockEmailService) // Not in providers, so local mock OK
    logger := pkgLogger.New("test", "error")

    useCase := NewRequestPasswordResetUseCase(
        mockUserRepo,
        mockTokenRepo,
        mockEmailService,
        logger,
    )

    ctx := context.Background()
    givenEmail := "test@example.com"
    givenOrgID := uuid.New()
    givenUserID := uuid.New()

    givenUser := &domain.User{
        ID:             givenUserID,
        Email:          givenEmail,
        FirstName:      "Test",
        OrganizationID: givenOrgID,
    }

    mockUserRepo.On("GetByEmailAndOrg", ctx, givenEmail, givenOrgID).Return(givenUser, nil)
    mockTokenRepo.On("StorePasswordResetToken", ctx, mock.AnythingOfType("*domain.PasswordResetToken")).Return(nil)
    mockEmailService.On("SendPasswordResetEmail", ctx, givenEmail, mock.AnythingOfType("string"), "Test").Return(nil)

    // When
    err := useCase.Execute(ctx, givenEmail, givenOrgID)

    // Then
    assert.NoError(t, err)
    mockUserRepo.AssertExpectations(t)
    mockTokenRepo.AssertExpectations(t)
    mockEmailService.AssertExpectations(t)
}
```

---

## Enforcement

This standard is **MANDATORY** for:

- ✅ All Claude agents working on this project
- ✅ All human developers contributing code
- ✅ All pull requests (will be rejected if tests missing)
- ✅ All new features and bug fixes

**No exceptions without explicit approval from project lead.**

---

## References

- [Go Testing Package Documentation](https://pkg.go.dev/testing)
- [Testify Mock Documentation](https://pkg.go.dev/github.com/stretchr/testify/mock)
- [Project CLAUDE.md](./CLAUDE.md) - General Go development guidelines
- [Coverage Analysis](./docs/coverage-reports/) - Historical coverage data

---

**Last Updated:** 2025-01-18
**Approved By:** Project Team via User Request
**Version:** 1.0
