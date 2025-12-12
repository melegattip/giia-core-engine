# Auth Service - Testing Strategy

## Testing Pyramid Approach

Following the testing pyramid methodology for comprehensive coverage:

```
                    ╱╲
                   ╱  ╲
                  ╱ E2E ╲
                 ╱────────╲
                ╱          ╲
               ╱ Integration╲
              ╱──────────────╲
             ╱                ╲
            ╱  Component Tests ╲
           ╱────────────────────╲
          ╱                      ╲
         ╱     Unit Tests (90%+)  ╲
        ╱────────────────────────────╲
```

## Testing Layers

### Layer 1: Unit Tests (Base - 90%+ Coverage)
**Scope**: Individual functions, methods, and use cases in isolation
**Tools**: Go testing, testify/assert, testify/mock
**Coverage Target**: 90%+

**What to Test:**
- ✅ All use cases (auth, rbac, role management)
- ✅ Domain entities validation logic
- ✅ Repository methods (with mocks)
- ✅ Adapters (JWT manager, cache)
- ✅ Utilities and helpers
- ✅ Error handling paths
- ✅ Edge cases and boundary conditions

**Test Files Structure:**
```
internal/core/usecases/auth/
├── login_test.go
├── register_test.go
├── refresh_test.go
├── validate_token_test.go
└── ...

internal/core/usecases/rbac/
├── check_permission_test.go
├── batch_check_test.go
├── get_user_permissions_test.go
├── resolve_inheritance_test.go
└── ...
```

### Layer 2: Component Tests
**Scope**: Multiple units working together (e.g., use case + repository)
**Tools**: Go testing with test containers for dependencies
**Coverage Target**: Critical paths

**What to Test:**
- ✅ Use case + Repository (real database)
- ✅ Use case + Cache (real Redis)
- ✅ JWT generation + validation flow
- ✅ RBAC permission resolution with hierarchy
- ✅ Multi-tenancy isolation

**Test Files Structure:**
```
internal/core/usecases/auth/
└── integration_test.go

internal/core/usecases/rbac/
└── integration_test.go
```

### Layer 3: Integration Tests
**Scope**: HTTP/gRPC APIs with real infrastructure
**Tools**: Go testing, httptest, grpc test clients
**Coverage Target**: All API endpoints

**What to Test:**
- ✅ HTTP REST API endpoints
- ✅ gRPC service methods
- ✅ Middleware (auth, permissions)
- ✅ Request/response formats
- ✅ Error responses
- ✅ Multi-service communication

**Test Files Structure:**
```
tests/integration/
├── http_api_test.go
├── grpc_api_test.go
├── auth_flow_test.go
└── rbac_flow_test.go
```

### Layer 4: End-to-End Tests
**Scope**: Complete user workflows across services
**Tools**: Go testing, Docker Compose test environment
**Coverage Target**: Critical user journeys

**What to Test:**
- ✅ Complete registration → login → access protected resource
- ✅ Role assignment → permission check → access grant/deny
- ✅ Token expiry and refresh flow
- ✅ Multi-tenancy isolation verification
- ✅ Cross-service gRPC communication

**Test Files Structure:**
```
tests/e2e/
├── user_journey_test.go
├── rbac_journey_test.go
└── multi_tenant_test.go
```

## Test Coverage Goals

### Overall Targets
- **Unit Tests**: 90%+ coverage
- **Component Tests**: 100% of critical paths
- **Integration Tests**: 100% of API endpoints
- **E2E Tests**: 100% of P1 user stories

### Per Package Targets
| Package | Target Coverage | Priority |
|---------|----------------|----------|
| `usecases/auth` | 95%+ | P0 |
| `usecases/rbac` | 95%+ | P0 |
| `usecases/role` | 90%+ | P1 |
| `repositories` | 85%+ | P1 |
| `adapters` | 90%+ | P1 |
| `entrypoints/http` | 85%+ | P1 |
| `grpc/server` | 90%+ | P0 |
| `grpc/interceptors` | 80%+ | P2 |

## Testing Conventions

### Test Naming
```go
func TestFunctionName_Scenario_ExpectedBehavior(t *testing.T)
```

**Examples:**
```go
func TestLogin_WithValidCredentials_ReturnsTokens(t *testing.T)
func TestLogin_WithInvalidPassword_ReturnsUnauthorizedError(t *testing.T)
func TestCheckPermission_WithWildcardPermission_AllowsAllActions(t *testing.T)
```

### Test Structure (Given-When-Then)
```go
func TestLogin_WithValidCredentials_ReturnsTokens(t *testing.T) {
    // Given - Setup test data and mocks
    mockRepo := new(mocks.UserRepository)
    useCase := auth.NewLoginUseCase(mockRepo, jwtManager, passwordService, logger)

    givenEmail := "user@example.com"
    givenPassword := "password123"
    givenUser := &domain.User{ID: uuid.New(), Email: givenEmail}

    mockRepo.On("GetByEmail", mock.Anything, givenEmail).Return(givenUser, nil)

    // When - Execute the function under test
    result, err := useCase.Execute(context.Background(), givenEmail, givenPassword)

    // Then - Verify results
    assert.NoError(t, err)
    assert.NotEmpty(t, result.AccessToken)
    assert.NotEmpty(t, result.RefreshToken)
    mockRepo.AssertExpectations(t)
}
```

### Variable Naming in Tests
- `given*` - Input data and configuration
- `expected*` - Expected results and outcomes

```go
// Given variables
givenUserID := uuid.New()
givenPermission := "catalog:products:read"
givenMockError := errors.New("database error")

// Expected variables
expectedAllowed := true
expectedCacheCalls := 1
expectedError := errors.New("permission denied")
```

### Mock Usage
- **Be specific** with mock parameters - avoid `mock.Anything` when possible
- **Verify behavior** - check that mocks were called correctly
- **One assertion per test** - focus on single behavior

```go
// ✅ Good - Specific parameters
mockRepo.On("GetByID", ctx, specificUserID).Return(user, nil)

// ❌ Bad - Too generic
mockRepo.On("GetByID", mock.Anything, mock.Anything).Return(user, nil)
```

## Test Data Management

### Test Fixtures
Create reusable test data in `testdata/` directory:

```
tests/testdata/
├── users.json          # Sample user data
├── roles.json          # Sample roles
├── permissions.json    # Sample permissions
└── organizations.json  # Sample orgs
```

### Factory Functions
```go
// tests/fixtures/user_factory.go
func CreateTestUser(t *testing.T, overrides ...func(*domain.User)) *domain.User {
    user := &domain.User{
        ID:             uuid.New(),
        Email:          "test@example.com",
        FirstName:      "Test",
        LastName:       "User",
        Status:         domain.UserStatusActive,
        OrganizationID: uuid.New(),
    }

    for _, override := range overrides {
        override(user)
    }

    return user
}
```

## Running Tests

### Unit Tests
```bash
# All unit tests
go test ./internal/... -v

# Specific package
go test ./internal/core/usecases/auth -v

# With coverage
go test ./internal/... -coverprofile=coverage.out
go tool cover -html=coverage.out

# With race detection
go test ./internal/... -race
```

### Component Tests
```bash
# Requires Docker for test containers
docker-compose -f docker-compose.test.yml up -d
go test ./internal/.../integration_test.go -tags=integration
docker-compose -f docker-compose.test.yml down
```

### Integration Tests
```bash
# Start test environment
make test-env-up

# Run integration tests
go test ./tests/integration/... -v

# Cleanup
make test-env-down
```

### E2E Tests
```bash
# Start full environment
docker-compose up -d

# Run E2E tests
go test ./tests/e2e/... -v

# Cleanup
docker-compose down
```

### Coverage Report
```bash
# Generate coverage for all packages
make test-coverage

# View HTML report
make test-coverage-html
```

## CI/CD Integration

### GitHub Actions Workflow
```yaml
name: Tests

on: [push, pull_request]

jobs:
  unit-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.23'
      - name: Run unit tests
        run: go test ./internal/... -coverprofile=coverage.out
      - name: Check coverage
        run: |
          coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
          if (( $(echo "$coverage < 90" | bc -l) )); then
            echo "Coverage $coverage% is below 90%"
            exit 1
          fi

  integration-tests:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15
      redis:
        image: redis:7
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
      - name: Run integration tests
        run: go test ./tests/integration/... -v
```

## Test Checklist

### Before Committing
- [ ] All unit tests pass
- [ ] Coverage is ≥90%
- [ ] No race conditions (`go test -race`)
- [ ] No skipped tests without justification
- [ ] Mocks are verified (`AssertExpectations`)
- [ ] Test names follow conventions

### Before Merging PR
- [ ] All tests pass in CI
- [ ] Integration tests pass
- [ ] No flaky tests
- [ ] Test documentation updated
- [ ] New features have tests

### Before Release
- [ ] E2E tests pass
- [ ] Performance tests pass
- [ ] Load tests pass
- [ ] Security tests pass

## Troubleshooting

### Common Issues

**Test Flakiness**
```go
// ❌ Bad - Time-dependent
time.Sleep(100 * time.Millisecond)

// ✅ Good - Deterministic
mockTime.On("Now").Return(fixedTime)
```

**Database State**
```go
// ✅ Good - Clean state per test
func TestSomething(t *testing.T) {
    t.Cleanup(func() {
        cleanupDatabase(t)
    })
}
```

**Mock Issues**
```go
// Verify mocks were called
mockRepo.AssertExpectations(t)
mockRepo.AssertNumberOfCalls(t, "GetByID", 1)
```

## Performance Testing

### Load Tests (Post-MVP)
```bash
# Using k6
k6 run tests/load/auth_load_test.js

# Target metrics
- Requests/sec: 10,000
- P95 latency: <50ms
- P99 latency: <100ms
- Error rate: <0.1%
```

## Security Testing

### Security Checks (Post-MVP)
```bash
# Static analysis
gosec ./...

# Dependency vulnerabilities
go list -json -m all | nancy sleuth

# SQL injection testing
sqlmap -u "http://localhost:8081/api/v1/users"
```

## Documentation

### Test Documentation Requirements
- README in each test directory
- Inline comments for complex test logic
- Table-driven test documentation
- Test failure troubleshooting guide

## Maintenance

### Test Maintenance Guidelines
- Update tests when requirements change
- Remove obsolete tests
- Refactor duplicated test code
- Keep test data up to date
- Review and fix flaky tests immediately

---

**Implementation Status**
- [ ] Layer 1: Unit Tests (90%+ coverage)
- [ ] Layer 2: Component Tests
- [ ] Layer 3: Integration Tests
- [ ] Layer 4: E2E Tests

**Current Coverage**: 0%
**Target Coverage**: 90%+
**Timeline**: TBD
