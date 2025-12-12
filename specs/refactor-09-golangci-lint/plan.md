# Implementation Plan: Add Golangci-lint CI Enforcement

**Date**: 2025-12-10
**Spec**: [spec.md](./spec.md)

## Summary

Configure golangci-lint with project-specific rules, add to CI/CD pipeline, and set up pre-commit hooks. Enforce bans on `fmt.Errorf` and `time.Now()` in prohibited packages. Fix existing violations before enforcement.

## Technical Context

**Language/Version**: Go 1.23.4
**Primary Dependencies**: golangci-lint v1.55+, pre-commit framework
**Testing**: Lint enforcement in CI/CD, pre-commit hooks
**Project Type**: Microservice backend
**Performance Goals**: Lint execution <2 minutes for full codebase
**Constraints**: Must not block development unnecessarily
**Scale/Scope**: Configure 10+ linters, create 2 custom rules

## Project Structure

### Files to Create/Modify

```text
.
├── .golangci.yml                           # NEW - linter configuration
├── .github/
│   └── workflows/
│       └── lint.yml                        # NEW - CI/CD workflow
├── .pre-commit-config.yaml                 # NEW - pre-commit hooks
├── Makefile                                # UPDATE - add lint targets
├── README.md                               # UPDATE - add lint documentation
└── docs/
    └── linting-guide.md                    # NEW - linting documentation
```

---

## Phase 1: Install & Configure Golangci-lint

**Purpose**: Set up golangci-lint locally and configure rules

- [ ] T001 Install golangci-lint: `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`
- [ ] T002 Verify installation: `golangci-lint version`
- [ ] T003 Create .golangci.yml configuration file
- [ ] T004 Enable standard linters: gofmt, goimports, govet, errcheck, staticcheck, gosec
- [ ] T005 Enable additional linters: unconvert, unparam, unused, ineffassign, misspell
- [ ] T006 Configure linter settings: errcheck (check-blank: true), govet (check-shadowing: true)
- [ ] T007 Set timeout: 5 minutes
- [ ] T008 Configure exclude rules for tests (relax some rules in *_test.go)

**Checkpoint**: Golangci-lint installed and basic configuration created

---

## Phase 2: Create Custom Rules - Ban fmt.Errorf

**Goal**: Custom rule to prevent fmt.Errorf in prohibited packages

- [ ] T009 Research forbidigo linter: supports custom forbidden patterns
- [ ] T010 Add forbidigo to .golangci.yml linters
- [ ] T011 Configure forbidigo rule: `forbid "fmt\\.Errorf" in "internal/core/**"` (ban in core layer)
- [ ] T012 Configure forbidigo rule: `forbid "fmt\\.Errorf" in "internal/infrastructure/repositories/**"`
- [ ] T013 Add custom error message: "Use typed errors from pkg/errors instead of fmt.Errorf"
- [ ] T014 Test rule: run `golangci-lint run` and verify fmt.Errorf detected
- [ ] T015 Add exclusions if needed: allow fmt.Errorf in specific legacy files temporarily

**Checkpoint**: fmt.Errorf banned in core and repository layers

---

## Phase 3: Create Custom Rules - Ban time.Now()

**Goal**: Custom rule to prevent direct time.Now() usage

- [ ] T016 Configure forbidigo rule: `forbid "time\\.Now" in "internal/core/**"`
- [ ] T017 Configure forbidigo rule: `forbid "time\\.Now" in "internal/infrastructure/**"`
- [ ] T018 Add custom error message: "Use injected TimeManager instead of time.Now()"
- [ ] T019 Test rule: run `golangci-lint run` and verify time.Now() detected
- [ ] T020 Add exclusions for test files: allow time.Now() in *_test.go with justification

**Checkpoint**: time.Now() banned in production code

---

## Phase 4: Fix Existing Violations

**Purpose**: Clean up codebase before enforcing rules

- [ ] T021 Run golangci-lint: `golangci-lint run --fix` (auto-fix what's possible)
- [ ] T022 Review violations: `golangci-lint run | tee lint-report.txt`
- [ ] T023 Fix fmt.Errorf violations (covered in refactor-02, may already be done)
- [ ] T024 Fix time.Now() violations (covered in refactor-03, may already be done)
- [ ] T025 Fix other linter errors: unused variables, shadowing, etc.
- [ ] T026 Target: <5 violations remaining (document any exceptions in .golangci.yml)
- [ ] T027 Verify: `golangci-lint run` shows minimal errors

**Checkpoint**: Codebase clean, ready for enforcement

---

## Phase 5: Makefile Integration

**Goal**: Add lint targets to Makefile

- [ ] T028 Add `make lint` target: runs `golangci-lint run`
- [ ] T029 Add `make lint-fix` target: runs `golangci-lint run --fix`
- [ ] T030 Add `make lint-fast` target: runs quick linters only (gofmt, goimports)
- [ ] T031 Add `make lint-verbose` target: runs with verbose output for debugging
- [ ] T032 Test targets: verify all make commands work
- [ ] T033 Update `make help`: document lint targets

**Checkpoint**: Lint accessible via Makefile

---

## Phase 6: GitHub Actions CI/CD Integration

**Goal**: Enforce linting in CI/CD pipeline

- [ ] T034 Create .github/workflows/lint.yml
- [ ] T035 Configure workflow trigger: on pull_request and push to main/develop
- [ ] T036 Add job step: Checkout code
- [ ] T037 Add job step: Setup Go 1.23.4
- [ ] T038 Add job step: Install golangci-lint (use golangci/golangci-lint-action)
- [ ] T039 Add job step: Run golangci-lint with --timeout 5m
- [ ] T040 Configure to fail PR if linting errors found
- [ ] T041 Add caching: cache golangci-lint and Go modules for speed
- [ ] T042 Test workflow: create PR with lint violation, verify CI fails
- [ ] T043 Add status badge to README: "Lint Status"

**Checkpoint**: CI/CD enforces linting, blocks non-compliant PRs

---

## Phase 7: Pre-commit Hooks

**Goal**: Run linters before git commit (shift-left quality)

- [ ] T044 Install pre-commit: `pip install pre-commit` (or brew install pre-commit)
- [ ] T045 Create .pre-commit-config.yaml
- [ ] T046 Add golangci-lint hook: uses golangci/golangci-lint@latest
- [ ] T047 Configure hook to run on staged .go files only
- [ ] T048 Add other hooks: gofmt, goimports, go-vet
- [ ] T049 Install hooks: `pre-commit install`
- [ ] T050 Test hooks: stage file with lint error, attempt commit, verify blocked
- [ ] T051 Document hook installation in README
- [ ] T052 Add to developer onboarding: install pre-commit hooks

**Checkpoint**: Pre-commit hooks installed, prevent bad commits locally

---

## Phase 8: Documentation & Team Rollout

**Purpose**: Document linting process and educate team

- [ ] T053 Create docs/linting-guide.md
- [ ] T054 Document which linters are enabled and why
- [ ] T055 Document custom rules: fmt.Errorf ban, time.Now() ban
- [ ] T056 Document how to run linters locally: `make lint`
- [ ] T057 Document how to fix auto-fixable issues: `make lint-fix`
- [ ] T058 Document how to request exceptions (update .golangci.yml with justification)
- [ ] T059 Update README.md: add "Code Quality" section with lint badge
- [ ] T060 Update CLAUDE.md: add linting to development workflow
- [ ] T061 Create team announcement: explain new linting requirements
- [ ] T062 Hold team training: demonstrate lint workflow
- [ ] T063 Add lint check to PR template: "[ ] Lint passes (make lint)"
- [ ] T064 Monitor initial enforcement: help team with any issues

**Checkpoint**: Team onboarded, linting part of standard workflow

---

## Dependencies & Execution Order

### Sequential Phases

1. **Install (Phase 1)** - must complete first
2. **Custom Rules (Phase 2-3)** - depends on Phase 1, can run in parallel
3. **Fix Violations (Phase 4)** - depends on Phase 2-3 (needs rules defined)
4. **Makefile (Phase 5)** - depends on Phase 1, can run parallel with Phase 2-4
5. **CI/CD (Phase 6)** - depends on Phase 4 (violations must be fixed first)
6. **Pre-commit (Phase 7)** - depends on Phase 1, can run parallel with Phase 5-6
7. **Documentation (Phase 8)** - depends on all previous phases

## Notes

- Golangci-lint is fast: typically <2 minutes for full codebase
- Use `--fast` flag for quick feedback during development
- Pre-commit hooks can be bypassed with `--no-verify` (document when appropriate)
- CI/CD enforcement is final gate - cannot be bypassed
- Start with warnings, escalate to errors after team adaptation
- Monitor false positives: adjust rules if too many exceptions needed
- Golangci-lint has ~100 linters - only enable relevant ones
- Update golangci-lint regularly: `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`
- Consider golangci-lint config inheritance for multi-service repos
- Estimated effort: 0.5 day (quick setup, longer if many violations to fix)
