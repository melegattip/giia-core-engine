# Implementation Plan: Consolidate Microservices - Defer Splitting

**Date**: 2025-12-10
**Spec**: [spec.md](./spec.md)

## Summary

Archive 4 skeleton microservices (catalog, ddmrp, execution, analytics) to reduce deployment complexity. Document decision in ADR. Follow monolith-first approach - validate domain model before microservice decomposition.

## Technical Context

**Language/Version**: Go 1.23
**Project Type**: Monorepo microservices → Monolithic service
**Performance Goals**: Simpler deployment, faster iteration
**Constraints**: Must preserve skeleton code for future reference
**Scale/Scope**: 4 services with <10 files each to archive

## Project Structure

### Before

```text
services/
├── auth-service/           # 77 files - fully developed
├── catalog-service/        # 8 files - skeleton
├── ddmrp-engine-service/   # 6 files - skeleton
├── execution-service/      # 7 files - skeleton
├── analytics-service/      # 4 files - skeleton
└── ai-agent-service/       # 5 files - skeleton
```

### After

```text
services/
└── auth-service/           # Active development

archive/
├── catalog-service/        # Preserved for future
├── ddmrp-engine-service/
├── execution-service/
├── analytics-service/
├── ai-agent-service/
└── README.md               # Explains why archived
```

---

## Phase 1: Create ADR (Architecture Decision Record)

**Goal**: Document consolidation rationale

- [ ] T001 [P] [US2] Create docs/architecture/adr/ directory
- [ ] T002 [P] [US2] Create ADR template for 001-consolidate-services.md
- [ ] T003 [P] [US2] Document context: premature microservice decomposition
- [ ] T004 [P] [US2] Document decision: consolidate to monolith until domain validated
- [ ] T005 [P] [US2] Document consequences: simpler deployment, faster development
- [ ] T006 [P] [US2] Document future split criteria: when domain boundaries clear, traffic justifies scaling
- [ ] T007 [P] [US2] Review ADR with team lead

**Checkpoint**: ADR created and reviewed

---

## Phase 2: Archive Skeleton Services

**Goal**: Move skeleton services to archive/

- [ ] T008 [P] [US1] Create archive/ directory in repo root
- [ ] T009 [P] [US1] Move services/catalog-service/ to archive/catalog-service/
- [ ] T010 [P] [US1] Move services/ddmrp-engine-service/ to archive/ddmrp-engine-service/
- [ ] T011 [P] [US1] Move services/execution-service/ to archive/execution-service/
- [ ] T012 [P] [US1] Move services/analytics-service/ to archive/analytics-service/
- [ ] T013 [P] [US1] Move services/ai-agent-service/ to archive/ai-agent-service/
- [ ] T014 [P] [US1] Create archive/README.md explaining preservation rationale
- [ ] T015 [P] [US1] Verify only services/auth-service/ remains in services/

**Checkpoint**: Skeleton services archived with documentation

---

## Phase 3: Update Go Workspace

**Goal**: Remove archived services from go.work

- [ ] T016 Open go.work file
- [ ] T017 Remove line: `use ./services/catalog-service`
- [ ] T018 Remove line: `use ./services/ddmrp-engine-service`
- [ ] T019 Remove line: `use ./services/execution-service`
- [ ] T020 Remove line: `use ./services/analytics-service`
- [ ] T021 Remove line: `use ./services/ai-agent-service`
- [ ] T022 Verify go.work only includes: `use ./services/auth-service` and `use ./pkg`
- [ ] T023 Run: `go work sync`
- [ ] T024 Verify build: `go build ./...`

**Checkpoint**: Go workspace updated, build succeeds

---

## Phase 4: Update Docker Compose

**Goal**: Simplify deployment to single service

- [ ] T025 Open docker-compose.yml
- [ ] T026 Remove service definitions for catalog, ddmrp, execution, analytics, ai-agent
- [ ] T027 Keep only: auth-service, postgres, redis, nats (infrastructure)
- [ ] T028 Update network configuration if needed
- [ ] T029 Update volume mounts if needed
- [ ] T030 Test: `docker-compose up --build`
- [ ] T031 Verify auth-service starts successfully
- [ ] T032 Verify all infrastructure services (postgres, redis, nats) healthy

**Checkpoint**: Docker Compose simplified to single application service

---

## Phase 5: Update Documentation

**Goal**: Reflect monolithic architecture in README

- [ ] T033 [P] [US2] Update root README.md: change "6 microservices" to "monolithic service (future microservices planned)"
- [ ] T034 [P] [US2] Update architecture diagram: show auth-service as monolith
- [ ] T035 [P] [US2] Add section: "Future Microservice Decomposition" with split criteria
- [ ] T036 [P] [US2] Update deployment guide: reference single service deployment
- [ ] T037 [P] [US2] Update local development guide: simplified setup
- [ ] T038 [P] [US2] Link to ADR from README for rationale
- [ ] T039 [P] [US2] Update services/auth-service/README.md if needed

**Checkpoint**: Documentation reflects monolithic approach

---

## Phase 6: Update Makefile & Scripts

**Goal**: Remove multi-service build targets

- [ ] T040 Open Makefile
- [ ] T041 Remove build targets for archived services
- [ ] T042 Update `build-all` to only build auth-service
- [ ] T043 Update `test-all` to only test auth-service
- [ ] T044 Update `docker-build-all` for single service
- [ ] T045 Simplify deployment scripts if needed
- [ ] T046 Test: `make build-all`
- [ ] T047 Test: `make test-all`
- [ ] T048 Update CI/CD pipeline if multi-service references exist

**Checkpoint**: Build system simplified

---

## Phase 7: Notify Team & Finalize

**Purpose**: Communicate change and finalize consolidation

- [ ] T049 Create announcement: explain consolidation decision
- [ ] T050 Share ADR link with team (Slack/email)
- [ ] T051 Update development onboarding: single service setup
- [ ] T052 Update project roadmap: clarify when to split services
- [ ] T053 Git commit: "refactor: consolidate to monolithic service architecture"
- [ ] T054 Create CHANGELOG entry
- [ ] T055 Tag release: v2.0.0-monolith
- [ ] T056 Deploy consolidated service to dev environment
- [ ] T057 Verify production deployment plan updated
- [ ] T058 Archive old Kubernetes manifests for multi-service if they exist

**Checkpoint**: Team notified, consolidation complete

---

## Dependencies & Execution Order

### Sequential Phases

1. **ADR (Phase 1)** - must complete first for documentation
2. **Archive Services (Phase 2)** - core action
3. **Go Workspace (Phase 3)** - depends on Phase 2
4. **Docker Compose (Phase 4)** - can run parallel with Phase 3
5. **Documentation (Phase 5)** - depends on Phase 2-4
6. **Makefile (Phase 6)** - depends on Phase 2
7. **Finalize (Phase 7)** - depends on all previous phases

## Notes

- This is an architectural decision, not just code refactoring
- Preserve archived services - they contain domain research
- ADR is critical - explains "why" for future team members
- Monolith-first is Go community best practice (see: Sam Newman, Martin Fowler)
- Microservices should be split when:
  1. Domain boundaries are clear and stable
  2. Teams are organized around services
  3. Independent scaling is needed
  4. Deployment independence provides value
- Estimated effort: 1-2 days including documentation and deployment updates
- Consider this a strategic simplification, not a failure
- Easier to split monolith later than merge microservices
