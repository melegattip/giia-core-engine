---
alwaysApply: true
---

# Go Development Guidelines for AI Assistants

## Project Overview

This is a Go 1.23.4 application implementing Clean Architecture principles. This document consolidates all development rules, standards, and best practices for AI assistants working on this codebase.

---

## Table of Contents

1. [Architecture Principles](#1-architecture-principles)
2. [Go Coding Standards](#2-go-coding-standards)
3. [Development Workflow](#3-development-workflow)
4. [Testing Standards](#4-testing-standards)
5. [Validation & Security](#5-validation--security)
6. [Local Development](#6-local-development)

---

## 1. Architecture Principles

### Project Structure

```
project-root/
├── cmd/                     # Entry points
│   └── api/                # Main HTTP server
│
├── internal/               # Private code
│   ├── core/              # 🧠 DOMAIN & BUSINESS LOGIC
│   │   ├── domain/        # Entities and value objects
│   │   ├── usecases/      # Use cases
│   │   ├── providers/     # Interfaces/contracts
│   │   ├── errors/        # Domain errors
│   │   └── logs/          # Core logging
│   │
│   ├── infrastructure/    # 🔌 EXTERNAL ADAPTERS
│   │   ├── adapters/      # Interface implementations
│   │   ├── repositories/  # Data access
│   │   ├── entrypoints/   # HTTP controllers
│   │   ├── middlewares/   # HTTP middlewares
│   │   ├── context/       # Infrastructure context
│   │   └── logger/        # Logging system
│   │
│   ├── app/              # 🏗️ CONFIGURATION
│   └── stub/             # 🧪 MOCKS & TESTING
│
├── docs/                 # Documentation
├── ctx/rules/           # Development rules
├── go.mod               # Dependencies
├── go.sum               # Checksums
├── Makefile            # Automation commands
└── README.md           # Main documentation
```

### Clean Architecture Principles

- **Clear separation** between core, infrastructure, and app
- **Dependencies point inward**: Outer layers depend on inner layers
- **Framework independence**: Core must not depend on external tools
- **Dependency Injection**: Decouple components
- **Repository Pattern**: For external data access
- **Interface Segregation**: Small, specific interfaces

---

## 2. Go Coding Standards

### Mandatory Tools

#### Linters
```bash
# Primary linter
golangci-lint run

# Custom configuration
golangci-lint run -c .code_quality/.golangci.yml

# Pre-commit
pre-commit run --all-files
```

#### Formatting
```bash
gofmt
go vet
```

### Error Handling

**CRITICAL RULES:**
- **Always explicit**: Never ignore errors
- **Typed errors**: Use constructors from `internal/core/errors` package
- **NO fmt.Errorf**: Prefer typed errors over generic wrapping
- **Early validation**: Validate parameters at function start

#### Available Error Types

```go
// ✅ Client errors (4xx)
errors.NewBadRequest("invalid input")
errors.NewResourceNotFound("user not found")
errors.NewUnauthorizedRequest("authentication required")

// ✅ Server errors (5xx)
errors.NewInternalServerError("database connection failed")
errors.NewTooManyRequests("rate limit exceeded")

// ✅ Domain-specific errors
errors.NewResourceParsingError("invalid date format")
errors.NewSkippableError("optional service unavailable")
```

#### Error Handling Example

```go
// ✅ CORRECT - Use typed errors
func (r *Repository) Search(ctx context.Context, userID int64, products []string) ([]entities.SearchResults, error) {
    if userID <= 0 {
        return nil, errors.NewBadRequest("invalid userID")
    }

    results, err := r.searchRecords(ctx, userID, products)
    if err != nil {
        return nil, errors.NewInternalServerError("failed to search records")
    }

    return results, nil
}
```

### Context Management

- **Use context.Context**: For all asynchronous operations
- **Propagation**: Pass context through all layers
- **Timeouts**: Configure appropriate timeouts

```go
// ✅ CORRECT
func (s *Service) ProcessRequest(ctx context.Context, req *Request) (*Response, error) {
    // Use context in all downstream calls
    data, err := s.repository.GetData(ctx, req.ID)
    if err != nil {
        return nil, errors.NewInternalServerError("failed to get data")
    }

    return s.processData(ctx, data)
}
```

### Naming Conventions

#### General Principles
- **Descriptive**: Names that explain purpose
- **Consistent**: Follow Go conventions
- **Avoid abbreviations**: Use full, clear names
- **Acronyms**: In uppercase (ID, HTTP, JSON)

#### Variables, Constants, and Structs
- **camelCase**: For variables, constants, struct fields, and functions
- **PascalCase**: For exported types (structs, interfaces)

```go
// ✅ CORRECT
type UserID int64
type HTTPClient interface{}
type APIResponse struct{}

var processingEngine string
const maxRetryAttempts = 3

type DataRequest struct {
    UserID       int64
    SiteID       string
    DataValue    float64
}

// ❌ INCORRECT
type UserId int64
type HttpClient interface{}
var processing_engine string  // ❌ snake_case
const MAX_RETRY_ATTEMPTS = 3      // ❌ UPPER_CASE
```

#### Directories and Packages
- **snake_case**: For directory and package names
- **camelCase**: For import aliases

```go
// ✅ CORRECT - Directories in snake_case
internal/core/usecases/data_processing/
internal/infrastructure/adapters/time_manager/

// ✅ CORRECT - Import aliases in camelCase
import (
    dataProcessing "github.com/project/internal/core/usecases/data_processing"
    dataEntity "github.com/project/internal/core/domain/entities/data"
    timeManager "github.com/project/internal/infrastructure/adapters/time_manager"
)

// ❌ INCORRECT
import (
    data_processing "..."  // ❌ snake_case in alias
    tm "..."              // ❌ unclear abbreviation
)
```

#### Standard Import Aliases
- **Descriptive names**: Use descriptive camelCase for all import aliases
- **Contracts**: Suffix `Contract` (e.g., `requestContract`, `responseContract`)
- **Entities**: Suffix `Entity` when necessary to avoid collisions
- **Consistency**: Maintain consistent naming patterns across the project

### Code Comments

**CRITICAL: DO NOT COMMENT CODE**
- Code should be self-explanatory
- Use descriptive names
- Break into smaller, descriptive functions

```go
// ❌ BAD - Commenting obvious functionality
// This function sums two numbers
func sum(a, b int) int {
    return a + b
}

// ✅ GOOD - Descriptive name without comment
func calculateTotalWithTax(amount, taxRate float64) float64 {
    return amount * (1 + taxRate)
}
```

### Function Structure

**Validation Order:**
1. Parameter validation
2. Main logic
3. Response handling

```go
// ✅ Recommended structure
func ProcessData(ctx context.Context, input *Input) (*Output, error) {
    // 1. Validations
    if input == nil {
        return nil, errors.NewBadRequest("input cannot be nil")
    }
    if input.UserID <= 0 {
        return nil, errors.NewBadRequest("invalid user ID")
    }

    // 2. Main logic
    result, err := processBusinessLogic(ctx, input)
    if err != nil {
        return nil, errors.NewInternalServerError("business logic failed")
    }

    // 3. Prepare response
    return &Output{
        Data: result,
        Timestamp: time.Now(),
    }, nil
}
```

### Structured Logging

```go
// ✅ Use structured logs
logger.Error(ctx, err, "Error getting data from API",
    logs.Tags{
        "api_name": apiName,
        "user_id": userID,
        "operation": "search",
    })

// ✅ Info logs
logger.Info(ctx, "Processing request",
    logs.Tags{
        "user_id": userID,
        "products": strings.Join(products, ","),
    })
```

### Date Management with TimeManager

**MANDATORY**: Always use `TimeManager` for date operations

```go
// ✅ CORRECT - Dependency injection with TimeManager
type MyService struct {
    TimeManager timeManager.TimeManager
}

// ✅ CORRECT - Production usage
func (s *MyService) FormatDate(date time.Time) string {
    return s.TimeManager.FormatToISO8601(date)
}

// ✅ CORRECT - Date parsing
func (s *MyService) ParseDate(dateString string) (time.Time, error) {
    return s.TimeManager.StringToUTC(dateString)
}

// ✅ CORRECT - Mock in tests
func TestService(t *testing.T) {
    mockTimeManager := new(timeManager.TimeManagerMock)
    mockTimeManager.On("Now").Return(time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC))

    service := &MyService{TimeManager: mockTimeManager}
    // ... test logic
}
```

#### Main TimeManager Functions
- `Now()` - Current time in UTC
- `FormatToISO8601(date)` - ISO8601 Zulu format
- `FormatToOffset(date)` - Format with offset
- `StringToUTC(dateString)` - Parse string to UTC
- `StringYearMonthDayToUTC(date)` - Parse YYYY-MM-DD format
- `FirstDayOfNextMonth(date, siteID)` - First day of next month
- `GetDateWithoutTime(date, siteID)` - Date without time by site

#### ❌ Obsolete Practices

**DO NOT use anymore:**
- ~~`utils.ParseDate()`~~ - utils package removed
- ~~`dates.NewDateAdapter()`~~ - dates package deprecated
- ~~`dates.StringYearMonthDayToUTC()`~~ - Migrated to TimeManager

---

## 3. Development Workflow

### Branching Strategy

#### Branch Format
```bash
# Recommended format
feature/[ticket-prefix]-[task_number]-[descriptive-title]

# Examples
feature/PROJ-1728-interface-time
feature/PROJ-648-epic-feature
feature/PROJ-790-add-new-functionality
```

#### Branch Flow
- **Base Branch**: All branches must be created from `develop`
- **Target Branch**: PRs must target `develop`
- **Approvals**: Minimum 2 approvals for merge
- **Merge Strategy**: Use "Squash and Merge" for features
- **Epic Branch**: For large features requiring multiple parallel branches:
  - Create `feature/CBA3-epic-[name]` from `develop`
  - Individual branches created from epic branch
  - PRs target epic branch
  - Once all features complete, epic branch merges to `develop`

### Semantic Versioning

Follow `MAJOR.MINOR.PATCH` format:

- **MAJOR**: Breaking API changes
- **MINOR**: New backwards-compatible functionality
- **PATCH**: Backwards-compatible bug fixes

#### Examples
```
1.0.0 → 1.0.1  (patch: bug fix)
1.0.1 → 1.1.0  (minor: new feature)
1.1.0 → 2.0.0  (major: breaking change)
```

### Deployment Strategy

#### Merge Strategies by Branch
- **feature → develop**: SQUASH
- **hotfix → master**: MERGE
- **release → master**: MERGE
- **backports → develop**: MERGE
- **backports → master**: MERGE

#### Release Process
1. Create release branch from develop
2. PR to master with format: `Release - X.Y.Z - dd-mm-yyyy`
3. Deploy RC to blue-green
4. Verify in Kibana logs
5. Merge to master with "Create a merge commit"

### Pre-commit Checklist

- [ ] Does the branch follow the agreed naming format?
- [ ] Does the PR target the correct base branch?
- [ ] Does it have minimum required reviewers?
- [ ] Does it follow architectural rules?
- [ ] Do all linters pass?
- [ ] Does it have unit tests?

---

## 4. Testing Standards

### Test Commands

#### Basic Testing
```bash
# Basic tests
go test ./... -count=1

# Tests with race detection
go test ./... -count=1 -race

# Verbose output
go test ./... -v
```

#### Coverage
```bash
# Generate coverage for specific package
go test -coverprofile=coverage.out [package_path]

# Example
go test -coverprofile=coverage.out ./internal/core/usecases/...

# Visualize coverage
go tool cover -html="coverage.out"

# Coverage for all packages
go test ./... -coverprofile=coverage.out -covermode=atomic
```

### Test Conventions

#### Test Structure
Use **Given-When-Then** pattern:

```go
// ✅ Recommended structure
func TestFunctionName_Scenario_ExpectedBehavior(t *testing.T) {
    // Given - Prepare test data
    input := setupTestData()
    expectedResult := "expected_value"

    // When - Execute function under test
    result, err := FunctionToTest(input)

    // Then - Verify results
    assert.NoError(t, err)
    assert.Equal(t, expectedResult, result)
}
```

### Test Case Variable Naming

#### Prefix `given` - Input Data and Configuration
Variables that **we establish** as input data, mock configuration, or initial conditions:

```go
// ✅ "given" variable examples
type testScenario struct {
    name                        string
    givenRequestString          string                    // Input data
    givenSiteID                 constants.Site           // Configuration parameters
    givenLatestRecord           *dependencies.DebtContext // Initial states
    givenKvsGetError            error                     // Simulated errors in mocks
    givenKvsSaveError           error                     // Dependency behavior
    givenTimestampToUTCResponse time.Time               // Mocked responses
}
```

#### Prefix `expected` - Expected Results
Variables representing **results we expect** from the test:

```go
// ✅ "expected" variable examples
type testScenario struct {
    expectedError               error  // Expected error as result
    expectedKvsSaveCalls        int    // Expected number of calls
    expectedKvsGetCalls         int    // Expected mock interactions
    expectedPublishCalls        int    // Behavior validations
    expectedResult              string // Expected return values
}
```

#### Complete Example
```go
type testScenario struct {
    name                        string
    // Given - What we establish
    givenUserID                 int64
    givenRequestData            string
    givenRepositoryError        error
    givenMockResponse           *entities.User

    // Expected - What we expect
    expectedError               error
    expectedResult              *entities.User
    expectedRepositoryCalls     int
    expectedCacheWrites         int
}
```

### Naming Convention
```go
// Format: TestFunctionName_Scenario_ExpectedBehavior
func TestSearch_WithValidUserID_ReturnsResults(t *testing.T) {}
func TestSearch_WithInvalidUserID_ReturnsError(t *testing.T) {}
func TestSearch_WithEmptyProducts_ReturnsAllProducts(t *testing.T) {}
```

### Mock Conventions

#### Location and Naming
- **Location**: Same package as what it mocks
- **Name**: Same as original + `Mock` suffix
- **File**: Ends in `_mock.go`

```go
// ✅ Example: internal/core/providers/repository_mock.go
type RepositoryMock struct {
    mock.Mock
}

func (m *RepositoryMock) Search(ctx context.Context, userID int64, statuses []string, filters []string) ([]entities.SearchResults, error) {
    args := m.Called(ctx, userID, statuses, filters)
    return args.Get(0).([]entities.SearchResults), args.Error(1)
}
```

#### Mock Configuration

**CRITICAL: BE SPECIFIC with parameters**

```go
// ✅ Setup in tests - BE SPECIFIC with parameters
func TestService_ProcessData_Success(t *testing.T) {
    // Given
    mockRepo := new(RepositoryMock)
    service := NewService(mockRepo)

    expectedResults := []entities.SearchResults{{ID: 1}}
    expectedStatuses := []string{"approved", "active"}
    expectedFilters := []string{"type_a", "type_b"}

    // ✅ Specify exact values instead of mock.Anything
    mockRepo.On("Search",
        context.Background(),
        int64(123),
        expectedStatuses,
        expectedFilters,
    ).Return(expectedResults, nil)

    // When
    results, err := service.ProcessData(context.Background(), 123)

    // Then
    assert.NoError(t, err)
    assert.Equal(t, expectedResults, results)
    mockRepo.AssertExpectations(t)
}
```

#### Rules for mock.Anything

```go
// ❌ BAD - Using mock.Anything indiscriminately
mockRepo.On("Search", mock.Anything, mock.Anything, mock.Anything, mock.Anything)

// ✅ GOOD - Only when value truly doesn't matter
mockRepo.On("Search",
    mock.AnythingOfType("*context.Context"), // When context isn't relevant
    int64(123),                              // Specific UserID
    []string{"approved"},                    // Specific Statuses
    mock.Anything,                          // Only if products truly can be anything
)

// 🎯 BEST - Be specific whenever possible
mockRepo.On("Search",
    context.Background(),
    int64(123),
    []string{"approved", "active"},
    []string{"type_a"},
).Return(expectedResults, nil)
```

### Mock Best Practices

#### Fundamental Principles
- **Be specific** is better than using `mock.Anything`
- **Validate behavior** not just results
- **One mock per responsibility** - don't mock everything
- **Clean mocks** between tests

#### When to Use mock.Anything

```go
// ✅ ACCEPTABLE - Context that doesn't affect business logic
mockService.On("ProcessData", mock.AnythingOfType("*context.Context"), specificUserID)

// ✅ ACCEPTABLE - Dynamically generated IDs that can't be predicted
mockRepo.On("Save", mock.AnythingOfType("*entities.User")).Return(mock.Anything, nil)

// ❌ AVOID - Parameters that matter for logic
mockRepo.On("FindByStatus", mock.Anything) // What status? Must be specific!
```

#### Mock Validations

```go
// ✅ Validate called with correct parameters
mockRepo.On("UpdateStatus", int64(123), "approved").Return(nil).Once()

// ✅ Validate number of calls
mockRepo.AssertNumberOfCalls(t, "UpdateStatus", 1)

// ✅ Validate method was NOT called
mockRepo.AssertNotCalled(t, "Delete")
```

### Testing Strategies

#### Test Types
- **Unit Tests**: For each public function/method
- **Integration Tests**: For complete flows between components
- **Repository Tests**: To validate integration with external APIs

#### Coverage Goals
- **Minimum**: 80% coverage
- **Goal**: 90%+ for critical code
- **Focus**: Prioritize main paths and error cases

#### Test Data Management

```go
// ✅ Use TimeManagerMock in tests
func setupTestData() *TestInput {
    mockTimeManager := new(timeManager.TimeManagerMock)
    testDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

    return &TestInput{
        UserID:    123,
        Filters:   []string{"type_a", "type_b"},
        CreatedAt: testDate,
    }
}
```

### Integration Testing

#### External Dependencies
For integration testing with external services:
- Use appropriate mocking frameworks for dependencies
- Create documented test cases
- Validate behavior in staging/test environments

#### Local Testing Limitations
- Some external services may have limited functionality in local development
- Complete end-to-end validation should be performed in staging/test environments
- Real external API testing should be done in appropriate test environments

### Testing Checklist

- [ ] Unit tests for new public functions?
- [ ] Tests for error cases and edge cases?
- [ ] Do mocks follow naming conventions?
- [ ] Minimum 80% coverage?
- [ ] Integration tests for critical flows?
- [ ] Documentation of test cases for beta?
- [ ] Are unit tests as atomic as possible?

---

## 5. Validation & Security

### Input Validation

#### Mandatory Validations
- **Null parameters**: Verify they are not nil/null
- **Numeric ranges**: Validate they are within expected ranges
- **String length**: Verify maximum/minimum limits
- **Formats**: Validate emails, dates, IDs, etc.

#### Validation Examples

```go
// ✅ Input parameter validation
func ValidateSearchRequest(userID int64, products []string) error {
    if userID <= 0 {
        return errors.NewBadRequest("userID must be positive")
    }

    if len(products) > 10 {
        return errors.NewBadRequest("too many products specified")
    }

    for _, product := range products {
        if len(product) == 0 {
            return errors.NewBadRequest("product name cannot be empty")
        }
    }

    return nil
}

// ✅ Validation in functions
func (r *Repository) Search(ctx context.Context, userID int64, products []string) ([]entities.SearchResults, error) {
    // Early validation
    if err := ValidateSearchRequest(userID, products); err != nil {
        return nil, err // Error already typed, no wrapping needed
    }

    // Main logic...
}
```

### Security Validations

#### Data Sanitization

```go
// ✅ Clean inputs before processing
func SanitizeUserInput(input string) string {
    // Remove dangerous characters
    cleaned := strings.TrimSpace(input)
    cleaned = strings.ReplaceAll(cleaned, "<", "")
    cleaned = strings.ReplaceAll(cleaned, ">", "")
    return cleaned
}

// ✅ Validate ID format
func ValidateUserID(userID string) error {
    if _, err := strconv.ParseInt(userID, 10, 64); err != nil {
        return errors.NewBadRequest("invalid user ID format")
    }
    return nil
}
```

#### Authentication & Authorization

```go
// ✅ Validate authentication headers
func ValidateAuthHeaders(headers map[string]string) error {
    callerID, exists := headers["X-Caller-Id"]
    if !exists || callerID == "" {
        return errors.NewBadRequest("X-Caller-Id header is required")
    }

    if _, err := strconv.ParseInt(callerID, 10, 64); err != nil {
        return errors.NewBadRequest("invalid X-Caller-Id format")
    }

    return nil
}
```

### Business Logic Validation

#### Domain Validations

```go
// ✅ Validate business rules
func ValidateDataRequest(req *DataRequest) error {
    if req.UserID <= 0 {
        return errors.NewBadRequest("invalid user ID")
    }

    // Validate available types
    validTypes := map[string]bool{
        "type_a": true,
        "type_b": true,
        "type_c": true,
    }

    for _, dataType := range req.Types {
        if !validTypes[dataType] {
            return errors.NewBadRequest("invalid type: " + dataType)
        }
    }

    return nil
}
```

#### States and Transitions

```go
// ✅ Validate state transitions
func ValidateStatusTransition(from, to string) error {
    validTransitions := map[string][]string{
        "pending":  {"approved", "rejected"},
        "approved": {"paused", "cancelled"},
        "paused":   {"approved", "cancelled"},
    }

    validToStates, exists := validTransitions[from]
    if !exists {
        return errors.NewBadRequest("invalid from status: " + from)
    }

    for _, validTo := range validToStates {
        if validTo == to {
            return nil
        }
    }

    return errors.NewBadRequest("invalid transition from " + from + " to " + to)
}
```

### Validation Logging

```go
// ✅ Structured log for validations
func LogValidationError(ctx context.Context, err error, operation string, userID int64) {
    logger.Error(ctx, err, "Validation failed", logs.Tags{
        "operation": operation,
        "user_id":   userID,
        "error_type": "validation",
    })
}

// ✅ Usage in functions
func (s *Service) ProcessRequest(ctx context.Context, req *Request) error {
    if err := ValidateRequest(req); err != nil {
        LogValidationError(ctx, err, "process_request", req.UserID)
        return err // Error already typed, no wrapping needed
    }

    // Continue processing...
}
```

### Validation Checklist

- [ ] Are all input parameters validated?
- [ ] Are null/empty input cases handled?
- [ ] Are appropriate numeric ranges verified?
- [ ] Are authentication headers validated?
- [ ] Are inputs sanitized before processing?
- [ ] Are specific business rules validated?
- [ ] Are validation errors logged appropriately?
- [ ] Are typed error constructors used?
- [ ] Are descriptive and typed errors returned to client?

---

## 6. Local Development

### Development Environment Setup

#### Prerequisites
- **Go 1.23.4** or later
- **Docker** (if using containerized dependencies)
- **Make** for automation commands

#### Environment Configuration
```bash
# Load environment variables
source .env

# Or use your IDE's environment configuration
# See IDE-specific sections below
```

#### Load Environment
Follow README instructions to load the environment file into your preferred IDE.

### Local Development Considerations

#### External Services
- Some external services may have limited functionality locally
- Use mocks or stubs for external dependencies when appropriate
- Document which services require real connections vs mocks

#### Local vs Staging Testing

#### Local Development
```bash
# For development and basic debugging
go test ./... -count=1

# Verify linters
golangci-lint run
pre-commit run --all-files

# Run application locally
make run
# or
go run cmd/api/main.go
```

#### Staging/Test Environment Testing
- **Complete integration**: Validate with real external services
- **End-to-end flows**: Complete use case validation
- **Performance testing**: Real-world performance metrics
- **Documentation**: Generate documented test cases

### IDE Configuration

#### VSCode/Cursor
```json
// .vscode/launch.json
{
   "version": "0.2.0",
   "configurations": [
       {
           "name": "Launch",
           "type": "go",
           "request": "launch",
           "mode": "auto",
           "program": "${workspaceFolder}/cmd/api",
           "envFile": "${workspaceFolder}/env"
       }
   ]
}
```

#### IntelliJ/GoLand

**Option A: Manual**
1. Run `make sandbox`
2. **Edit Configurations**
3. Copy contents of `env` file
4. **Paste** into "User environment variables"

**Option B: EnvFile Plugin**
1. Install **EnvFile** plugin
2. **Edit Configurations → EnvFile**
3. Add `env` file generated by sandbox

### Local Monitoring

#### Logs
```bash
# Verify application logs
tail -f logs/app.log

# Structured logs
grep "ERROR" logs/app.log | jq .
```

#### Health Check
```bash
# Verify service is running
curl http://localhost:8080/health

# Verify metrics
curl http://localhost:8080/metrics
```

### Debugging

#### Tools
- **Delve**: For Go debugging
- **IDE Debugger**: Use breakpoints in IDE
- **Logging**: Add temporary logs for debugging

#### Common Cases
```go
// ✅ Temporary debug logging
logger.Debug(ctx, "Processing request", logs.Tags{
    "user_id": userID,
    "step": "validation",
    "data": fmt.Sprintf("%+v", input),
})
```

### Troubleshooting

#### Common Issues
- **Environment variables not loading**: Verify IDE configuration and .env file
- **Dependency connection timeouts**: Check network connectivity and service availability
- **Build failures**: Ensure all dependencies are properly installed with `go mod download`

#### Hot Reload (Optional)
```bash
# Install air for hot reload during development
go install github.com/cosmtrek/air@latest

# Configure air
air init

# Run with hot reload
air
```

---

## Quick Reference

### Command Cheat Sheet

```bash
# Development
make run                      # Run application
go run cmd/api/main.go       # Run directly

# Testing
go test ./... -count=1
go test ./... -count=1 -race
go test ./... -v             # Verbose

# Linting
golangci-lint run
pre-commit run --all-files

# Coverage
go test -coverprofile=coverage.out [package_path]
go tool cover -html="coverage.out"
go test ./... -coverprofile=coverage.out -covermode=atomic

# Formatting
gofmt -w .
go vet ./...

# Dependencies
go mod download
go mod tidy
go mod verify
```

### Key Reminders

1. **Use typed errors** from `internal/core/errors`
2. **Use TimeManager** for all date operations
3. **Follow snake_case** for directories, **camelCase** for aliases
4. **Be specific in mocks** - avoid `mock.Anything` when possible
5. **Validate early** in all functions
6. **No code comments** - make code self-explanatory
7. **Test coverage minimum** 80%
8. **Branch format**: Follow agreed naming convention
9. **Minimum required approvals** for PRs
10. **Clean Architecture** - dependencies point inward

---

## Cross-References

- **01-architecture.mdc** — Architecture principles
- **02-go-standards.mdc** — Go coding standards
- **03-development.mdc** — Development workflow
- **04-testing.mdc** — Testing conventions
- **05-validation.mdc** — Validation and security
- **06-local-development.mdc** — Local development setup
