# Linting Guide

This guide provides comprehensive information about code linting in the GIIA Core Engine project.

## Table of Contents

- [Overview](#overview)
- [Quick Start](#quick-start)
- [Enabled Linters](#enabled-linters)
- [Custom Rules](#custom-rules)
- [Running Linters Locally](#running-linters-locally)
- [Pre-commit Hooks](#pre-commit-hooks)
- [CI/CD Integration](#cicd-integration)
- [Handling Linting Errors](#handling-linting-errors)
- [Requesting Exceptions](#requesting-exceptions)

---

## Overview

We use [golangci-lint](https://golangci-lint.run/) as our primary linting tool. It's a fast Go linters aggregator that runs multiple linters in parallel.

### Why Linting?

- **Early Error Detection**: Catches common mistakes before code review
- **Code Consistency**: Enforces consistent coding style across the team
- **Security**: Identifies potential security vulnerabilities
- **Best Practices**: Ensures adherence to Go best practices
- **Maintainability**: Makes code easier to read and maintain

---

## Quick Start

### Install golangci-lint

```bash
# Using make (recommended)
make setup

# Or install directly
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### Run Linters

```bash
# Run all linters
make lint

# Auto-fix issues where possible
make lint-fix

# Format code only
make fmt
```

---

## Enabled Linters

Our configuration enables the following linters:

### Core Linters

| Linter | Purpose |
|--------|---------|
| **errcheck** | Checks for unchecked errors |
| **gofmt** | Enforces standard Go formatting |
| **goimports** | Manages imports and removes unused ones |
| **govet** | Examines Go source code for suspicious constructs |
| **staticcheck** | Advanced static analysis |

### Code Quality Linters

| Linter | Purpose |
|--------|---------|
| **ineffassign** | Detects ineffectual assignments |
| **unused** | Finds unused code |
| **unconvert** | Removes unnecessary type conversions |
| **unparam** | Reports unused function parameters |
| **misspell** | Finds commonly misspelled English words |

### Style & Maintainability

| Linter | Purpose |
|--------|---------|
| **revive** | Fast, configurable linter for Go style |
| **exportloopref** | Checks for pointers to enclosing loop variables |

### Security

| Linter | Purpose |
|--------|---------|
| **gosec** | Inspects source code for security problems |

### Custom Rules

| Linter | Purpose |
|--------|---------|
| **forbidigo** | Bans specific function calls (see [Custom Rules](#custom-rules)) |

---

## Custom Rules

We enforce project-specific rules using the `forbidigo` linter:

### Rule 1: Ban fmt.Errorf

**Applies to**: `internal/core/*`, `internal/infrastructure/repositories/*`, `pkg/*`

**Rule**: Do not use `fmt.Errorf` - use typed errors from `pkg/errors` instead

**Why**: Typed errors provide better error handling, HTTP status code mapping, and structured logging.

**Example**:

```go
// ❌ BAD
func GetUser(id int64) (*User, error) {
    if id <= 0 {
        return nil, fmt.Errorf("invalid user ID: %d", id)
    }
    // ...
}

// ✅ GOOD
func GetUser(id int64) (*User, error) {
    if id <= 0 {
        return nil, errors.NewBadRequest("invalid user ID")
    }
    // ...
}
```

**Available error constructors**:
- `errors.NewBadRequest(message)` - 400
- `errors.NewUnauthorizedRequest(message)` - 401
- `errors.NewResourceNotFound(message)` - 404
- `errors.NewInternalServerError(message)` - 500
- `errors.NewTooManyRequests(message)` - 429

### Rule 2: Ban errors.New

**Applies to**: `internal/core/*`, `internal/infrastructure/repositories/*`, `pkg/*`

**Rule**: Do not use `errors.New` - use typed errors from `pkg/errors` instead

**Why**: Same reasoning as `fmt.Errorf` - we need typed errors for proper error handling.

### Rule 3: Ban time.Now()

**Applies to**: `internal/core/*`, `internal/infrastructure/*`

**Rule**: Do not use `time.Now()` directly - use injected `TimeManager` instead

**Why**: Direct `time.Now()` calls make code difficult to test. `TimeManager` allows mocking time in tests.

**Example**:

```go
// ❌ BAD
type Service struct {
    repo Repository
}

func (s *Service) CreateRecord() (*Record, error) {
    return &Record{
        CreatedAt: time.Now(), // Hard to test!
    }, nil
}

// ✅ GOOD
type Service struct {
    repo        Repository
    timeManager timeManager.TimeManager
}

func (s *Service) CreateRecord() (*Record, error) {
    return &Record{
        CreatedAt: s.timeManager.Now(), // Testable!
    }, nil
}
```

**Exceptions**: `time.Now()` is allowed in:
- Test files (`*_test.go`)
- Main entry points (`cmd/*/main.go`)
- Mock files

---

## Running Linters Locally

### Basic Commands

```bash
# Run all linters
golangci-lint run

# Run with timeout
golangci-lint run --timeout=5m

# Auto-fix issues
golangci-lint run --fix

# Run on specific directory
golangci-lint run ./services/auth-service/...

# Verbose output
golangci-lint run -v
```

### Using Makefile Targets

```bash
# Run all linters
make lint

# Auto-fix issues
make lint-fix

# Format code
make fmt

# Run go vet
make vet
```

---

## Pre-commit Hooks

Pre-commit hooks run linters automatically before each commit, catching issues early.

### Installation

```bash
# Install pre-commit
pip install pre-commit

# Or using homebrew on macOS
brew install pre-commit

# Install the git hook scripts
pre-commit install
```

### What Runs on Commit?

When you run `git commit`, the following checks run automatically:

1. **Go Formatting** (`gofmt`)
2. **Go Vet** (`go vet`)
3. **Import Formatting** (`goimports`)
4. **golangci-lint** (all configured linters)
5. **General File Checks**:
   - Trailing whitespace removal
   - End-of-file fixer
   - Large file detection
   - Merge conflict detection
   - Private key detection

### Running Manually

```bash
# Run all hooks on all files
pre-commit run --all-files

# Run on staged files only
pre-commit run

# Run specific hook
pre-commit run golangci-lint
```

### Bypassing Pre-commit Hooks

```bash
# Skip all hooks (use sparingly!)
git commit --no-verify

# Only bypass for emergencies
# CI will still enforce linting
```

---

## CI/CD Integration

### GitHub Actions Workflows

Linting runs automatically in CI/CD on:

1. **Pull Requests** - Lint changed files only
2. **Push to main/develop** - Lint entire codebase

### Workflow Files

- [`.github/workflows/ci.yml`](../.github/workflows/ci.yml) - Full codebase linting
- [`.github/workflows/pr-checks.yml`](../.github/workflows/pr-checks.yml) - Changed files only

### PR Requirements

All PRs must:
- ✅ Pass golangci-lint checks
- ✅ Have no linting errors
- ✅ Follow all custom rules

**PRs with linting errors will be blocked from merging.**

---

## Handling Linting Errors

### Step 1: Understand the Error

Read the linter output carefully. Example:

```
services/auth-service/internal/core/usecases/auth/login.go:45:2:
  Do not use fmt.Errorf. Use typed errors from pkg/errors instead (forbidigo)
```

### Step 2: Fix the Issue

Most issues can be auto-fixed:

```bash
# Try auto-fix first
make lint-fix
```

If auto-fix doesn't work, manually fix based on linter guidance.

### Step 3: Verify the Fix

```bash
# Run linters again
make lint

# Run on specific file
golangci-lint run services/auth-service/internal/core/usecases/auth/login.go
```

### Common Errors and Solutions

#### Unchecked Errors (errcheck)

```go
// ❌ BAD
file.Close()

// ✅ GOOD
if err := file.Close(); err != nil {
    logger.Error(ctx, err, "failed to close file")
}

// ✅ ALSO GOOD (when error truly doesn't matter)
_ = file.Close()
```

#### Unused Variables (unused)

```go
// ❌ BAD
func processData(input string, unusedParam int) {
    fmt.Println(input)
}

// ✅ GOOD - Remove it
func processData(input string) {
    fmt.Println(input)
}

// ✅ GOOD - Or use it
func processData(input string, count int) {
    for i := 0; i < count; i++ {
        fmt.Println(input)
    }
}
```

#### Shadow Variables (govet)

```go
// ❌ BAD
func process() error {
    err := doSomething()
    if err != nil {
        err := fmt.Errorf("wrapped: %w", err) // shadows err
        return err
    }
    return nil
}

// ✅ GOOD
func process() error {
    err := doSomething()
    if err != nil {
        return fmt.Errorf("wrapped: %w", err)
    }
    return nil
}
```

---

## Requesting Exceptions

Sometimes you need to disable a linter for specific code. Use this sparingly!

### Inline Directive (Single Line)

```go
// nolint:errcheck // Explain why error check isn't needed
doSomething()
```

### Block Directive (Multiple Lines)

```go
// nolint:gosec
func unsafeOperation() {
    // Security exception approved in JIRA-1234
    cmd := exec.Command("bash", "-c", userInput)
    cmd.Run()
}
```

### File-Level Exclusion

Add to `.golangci.yml`:

```yaml
issues:
  exclude-rules:
    - path: services/auth-service/internal/legacy/old_code.go
      linters:
        - staticcheck
      text: "specific error pattern to exclude"
```

### Process for Permanent Exceptions

1. **Discuss with team** - Ensure exception is justified
2. **Document reason** - Add comment explaining why
3. **Update `.golangci.yml`** - Add exclusion with comment
4. **PR review** - Get approval from 2+ reviewers
5. **Create tech debt ticket** - If temporary exception

---

## Configuration Files

### .golangci.yml

Main configuration file defining:
- Which linters are enabled
- Linter-specific settings
- Exclusion rules
- Custom rules (forbidigo)

Located at: [`.golangci.yml`](../.golangci.yml)

### .pre-commit-config.yaml

Defines pre-commit hooks configuration.

Located at: [`.pre-commit-config.yaml`](../.pre-commit-config.yaml)

---

## Troubleshooting

### golangci-lint Not Found

```bash
# Reinstall
make setup

# Or manually
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Verify installation
golangci-lint version
```

### Linter Takes Too Long

```bash
# Increase timeout
golangci-lint run --timeout=10m

# Or run on specific package
golangci-lint run ./services/auth-service/...
```

### False Positives

If you believe a linter error is incorrect:

1. Verify you understand the error correctly
2. Check if it's a known issue: https://github.com/golangci/golangci-lint/issues
3. Discuss with team
4. If confirmed false positive, add exclusion to `.golangci.yml`

### Pre-commit Hook Fails

```bash
# Update hooks
pre-commit autoupdate

# Clear cache
pre-commit clean

# Reinstall
pre-commit uninstall
pre-commit install
```

---

## Best Practices

1. **Run linters frequently** - Before committing, not just in CI
2. **Fix issues immediately** - Don't accumulate linting debt
3. **Understand the why** - Don't blindly fix, understand the rule
4. **Use auto-fix carefully** - Review auto-fix changes before committing
5. **Keep configuration updated** - Regularly update golangci-lint version
6. **Share learnings** - If you find a common issue, share with team

---

## Resources

- [golangci-lint Documentation](https://golangci-lint.run/)
- [Effective Go](https://golang.org/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Project CLAUDE.md](../CLAUDE.md) - Development guidelines
- [Project README](../README.md) - Getting started

---

## Getting Help

If you have questions about linting:

1. Check this guide first
2. Review [CLAUDE.md](../CLAUDE.md) for coding standards
3. Ask in team chat
4. Create a discussion in GitHub Discussions

---

**Last Updated**: 2025-12-12
**Maintained by**: GIIA Development Team
