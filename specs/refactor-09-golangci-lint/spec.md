# Feature Specification: Add Golangci-lint CI Enforcement

**Created**: 2025-12-10
**Priority**: 🟢 LOW
**Effort**: 0.5 day

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Configure Golangci-lint (Priority: P1)

As a developer, I need golangci-lint configured with project-specific rules so that common mistakes are caught before code review.

**Why this priority**: Automated linting catches errors early and enforces consistency. Golangci-lint is Go community standard.

**Independent Test**: Run `make lint` and verify linters execute with project configuration.

**Acceptance Scenarios**:

1. **Scenario**: Lint catches fmt.Errorf usage
   - **Given** Code using fmt.Errorf in prohibited package
   - **When** Running golangci-lint
   - **Then** Linter reports error and suggests typed error alternative

2. **Scenario**: Lint catches time.Now() usage
   - **Given** Code using time.Now() in prohibited package
   - **When** Running golangci-lint
   - **Then** Linter reports error and suggests TimeManager usage

3. **Scenario**: Lint passes for compliant code
   - **Given** Code following all project rules
   - **When** Running golangci-lint
   - **Then** Zero errors reported

---

### User Story 2 - CI/CD Integration (Priority: P1)

As a team lead, I need golangci-lint running in CI/CD so that PRs with linting errors are automatically blocked.

**Why this priority**: Prevents non-compliant code from being merged. Enforces standards at gate level.

**Independent Test**: Create PR with linting error and verify CI fails.

**Acceptance Scenarios**:

1. **Scenario**: CI fails on linting errors
   - **Given** PR with code violating lint rules
   - **When** CI pipeline runs
   - **Then** Lint stage fails and blocks merge

2. **Scenario**: CI passes for clean code
   - **Given** PR with compliant code
   - **When** CI pipeline runs
   - **Then** Lint stage passes and allows merge

---

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: .golangci.yml configuration file MUST exist in repository root
- **FR-002**: Linters MUST include: gofmt, goimports, govet, errcheck, staticcheck, gosec
- **FR-003**: Custom rules MUST ban fmt.Errorf in internal/core/* and internal/infrastructure/repositories/*
- **FR-004**: Custom rules MUST ban time.Now() in internal/core/* and internal/infrastructure/*
- **FR-005**: Makefile MUST include `make lint` target
- **FR-006**: GitHub Actions workflow MUST run lint on every PR
- **FR-007**: Pre-commit hooks MUST run golangci-lint
- **FR-008**: README MUST document linter setup and usage

### Key Entities

- **.golangci.yml**: Linter configuration file
- **GitHub Actions Workflow**: CI pipeline running lint
- **Pre-commit Hook**: Local git hook running lint before commit
- **Makefile Target**: `make lint` command

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: `.golangci.yml` exists with project-specific rules
- **SC-002**: `make lint` command runs successfully
- **SC-003**: GitHub Actions lint workflow exists and runs on PR
- **SC-004**: All existing linting violations fixed (<5 violations remaining)
- **SC-005**: Pre-commit hooks installed and functional
- **SC-006**: README includes linter documentation
- **SC-007**: CI fails when lint errors introduced
