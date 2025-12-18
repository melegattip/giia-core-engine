# GIIA Specifications & Plans

This directory contains all feature specifications and implementation plans following the **Spec-Driven Development** methodology.

## Methodology

### 1. SPEC (WHAT to Build)
- **Purpose**: Define what needs to be built from a user/business perspective
- **Technology**: Agnostic - focuses on outcomes, not implementation
- **Contains**: User stories, requirements, acceptance criteria, success metrics
- **File**: `spec.md` in each feature folder

### 2. PLAN (HOW to Build)
- **Purpose**: Define the technical implementation approach
- **Technology**: Specific - includes frameworks, libraries, file structure
- **Contains**: Technical context, project structure, phased tasks, dependencies
- **File**: `plan.md` in each feature folder

## Directory Structure

```
specs/
├── README.md                           # This file
├── task-03-local-dev-environment/      # Spec & Plan for Task 3
│   ├── spec.md                         # WHAT: Dev environment requirements
│   └── plan.md                         # HOW: Docker Compose implementation
├── task-04-shared-packages/            # Spec & Plan for Task 4
│   ├── spec.md                         # WHAT: Shared package requirements
│   └── plan.md                         # HOW: Go package implementation
├── task-05-auth-service-migration/     # Spec & Plan for Task 5
│   ├── spec.md                         # WHAT: Auth/IAM requirements
│   └── plan.md                         # HOW: Migration & implementation
├── task-06-rbac-implementation/        # Spec & Plan for Task 6
│   ├── spec.md                         # WHAT: RBAC requirements
│   └── plan.md                         # HOW: RBAC implementation
├── task-07-grpc-server/                # Spec & Plan for Task 7
│   ├── spec.md                         # WHAT: gRPC requirements
│   └── plan.md                         # HOW: gRPC implementation
├── task-08-nats-jetstream/             # Spec & Plan for Task 8
│   ├── spec.md                         # WHAT: Event system requirements
│   └── plan.md                         # HOW: NATS Jetstream setup
├── task-09-catalog-service/            # Spec & Plan for Task 9
│   ├── spec.md                         # WHAT: Catalog service requirements
│   └── plan.md                         # HOW: Service implementation
└── task-10-kubernetes-cluster/         # Spec & Plan for Task 10
    ├── spec.md                         # WHAT: K8s requirements
    └── plan.md                         # HOW: K8s setup & deployment
```

## Workflow

### For Each New Feature/Task:

1. **Create SPEC First**
   ```bash
   # Copy template
   cp docs/templates\ spec-driven-development/spec-template.md specs/[feature-name]/spec.md

   # Fill out:
   - User scenarios (prioritized, independently testable)
   - Functional requirements
   - Key entities
   - Success criteria
   ```

2. **Then Create PLAN**
   ```bash
   # Copy template
   cp docs/templates\ spec-driven-development/plan-template.md specs/[feature-name]/plan.md

   # Fill out:
   - Technical context (Go 1.23, frameworks, etc.)
   - Project structure (actual file paths)
   - Phased tasks (Setup → Foundational → User Stories → Polish)
   - Dependencies & execution order
   ```

3. **Implement Following Plan**
   - Execute tasks in order
   - Check off tasks as completed
   - Stop at checkpoints to verify independently
   - Tests pass before moving forward

## Key Principles

### User Stories Must Be:
- **Prioritized** (P1, P2, P3...) by business value
- **Independently Testable** - Each story delivers standalone value
- **Incrementally Deliverable** - Can ship P1 alone as MVP

### Plans Must Include:
- **Phase 1**: Setup (project structure, dependencies)
- **Phase 2**: Foundational (blocking prerequisites - MUST complete before user stories)
- **Phase 3+**: User Stories (one phase per story, can run in parallel if staffed)
- **Final Phase**: Polish (cross-cutting concerns, optimization)

### Task Labeling:
- `[P]` = Test task (write test first)
- `[US1]` = Belongs to User Story 1
- `[US2]` = Belongs to User Story 2
- Task numbers (T001, T002...) for traceability

## Benefits

✅ **Clear Separation of Concerns**: What vs How
✅ **Technology Independence**: Can change implementation without changing spec
✅ **Testability**: Every user story is independently testable
✅ **Incremental Delivery**: Can ship P1 features first, add P2/P3 later
✅ **Traceability**: Tasks map back to user stories map back to requirements
✅ **Parallel Development**: Multiple devs can work on different user stories simultaneously

## Current Status

### Phase 1 Tasks (Foundation)

| Task | Spec | Plan | Implementation Status | Completion |
|------|------|------|----------------------|------------|
| Task 3: Local Dev Environment | ✅ Done | ✅ Done | ✅ **COMPLETE** - All .env files, scripts operational | 100% |
| Task 4: Shared Packages | ✅ Done | ✅ Done | 🟢 **ADVANCED** - All packages coded, integration tests pending | 85% |
| Task 5: Auth Service Migration | ✅ Done | ✅ Done | 🟢 **ADVANCED** - Clean Arch, RBAC, gRPC, multi-tenancy | 95% |
| Task 6: RBAC Implementation | ✅ Done | ✅ Done | 🟢 **ADVANCED** - Redis cache implemented, seed data pending | 95% |
| Task 7: gRPC Server | ✅ Done | ✅ Done | 🟢 **ADVANCED** - Proto files, generated code, full implementation | 95% |
| Task 8: NATS Jetstream | ✅ Done | ✅ Done | 🟢 **ADVANCED** - Stream config, event publishing active | 85% |
| Task 9: Catalog Service | ✅ Done | ✅ Done | 🟢 **ADVANCED** - Full Clean Architecture, REST API | 85% |
| Task 10: Kubernetes Cluster | ✅ Done | ✅ Done | ✅ **COMPLETE** - K8s setup with Helm charts | 100% |

**Phase 1 Status**: 93% Complete

### Phase 2A Tasks (Complete to 100%)

| Task | Spec | Plan | Implementation Status | Completion |
|------|------|------|----------------------|------------|
| Task 11: Auth Service Registration | ✅ Done | ✅ Done | ⏸️ **PENDING** - User registration, email verification, password reset | 0% → 100% |
| Task 12: Catalog Service Integration | ✅ Done | ⏸️ Pending | ⏸️ **PENDING** - gRPC, Supplier/BufferProfile, Auth integration | 0% → 100% |
| Task 13: Shared Packages Testing | ✅ Done | ⏸️ Pending | ⏸️ **PENDING** - Integration tests with PostgreSQL, NATS | 0% → 100% |

**Phase 2A Goal**: Bring Phase 1 from 93% to 100%

### Phase 2B Tasks (New Microservices)

| Task | Spec | Plan | Implementation Status | Completion |
|------|------|------|----------------------|------------|
| Task 14: DDMRP Engine Service | ✅ Done | ⏸️ Pending | ⏸️ **PENDING** - ADU, Buffer calculations, Net Flow Equation | 0% → 90% |
| Task 15: Execution Service | ✅ Done | ⏸️ Pending | ⏸️ **PENDING** - Orders, inventory transactions, replenishment | 0% → 90% |
| Task 16: Analytics Service | ✅ Done | ⏸️ Pending | ⏸️ **PENDING** - Dashboards, KPIs, reporting | 0% → 90% |
| Task 17: AI Agent Service | ✅ Done | ⏸️ Pending | ⏸️ **PENDING** - Forecasting, optimization, anomaly detection | 0% → 90% |

**Phase 2B Goal**: Implement remaining 4 microservices

**Legend**: ✅ Complete | 🟢 Advanced (>75%) | 🟡 Partial (<75%) | ⏸️ Pending

## Architecture: Monorepo Microservices

This project uses a **monorepo microservices architecture**:

- **6 independent microservices**: auth, catalog, ddmrp-engine, execution, analytics, ai-agent
- **Shared infrastructure packages**: config, logger, database, errors, events (in pkg/)
- **Service Communication**: gRPC for synchronous, NATS Jetstream for asynchronous
- **Independent Deployment**: Each service can be deployed separately
- **Coordinated Development**: All services in single repo with shared CI/CD

**Current Status**:
- Auth-service is most advanced (80% complete)
- Shared packages are 85% complete
- Other services are at skeleton stage awaiting implementation

See [PROJECT_STATUS.md](../PROJECT_STATUS.md) for detailed status of all services and tasks.

---

**Last Updated**: 2025-12-13 (Status audit completed, architecture clarified)
**Methodology**: Spec-Driven Development
**Project**: GIIA Core Engine
**Architecture**: Monorepo Microservices
