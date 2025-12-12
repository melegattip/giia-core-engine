# Implementation Plan: Local Development Environment

**Created**: 2025-12-09
**Spec**: [spec.md](./spec.md)

## Technical Context

### Technologies & Tools
- **Docker**: 24.0+ with Docker Compose v2
- **PostgreSQL**: 16-alpine image (Official)
- **Redis**: 7-alpine image (Official)
- **NATS**: 2-alpine image with Jetstream (Official)
- **pgAdmin**: dpage/pgadmin4 (Optional tool)
- **Redis Commander**: rediscommander/redis-commander (Optional tool)
- **Make**: Build automation tool

### Design Decisions
- **Docker Compose over Kubernetes locally**: Lighter weight, faster startup, easier debugging
- **Alpine images**: Smaller size, faster pulls, production-like
- **Named volumes**: Data persistence across container restarts
- **Health checks**: Ensures services are ready before dependent services start
- **Profiles for optional tools**: pgAdmin and Redis Commander only start when needed
- **Bridge network**: Isolated container communication

## Project Structure

```
giia-core-engine/
├── docker-compose.yml              # Main infrastructure definition (EXISTS - will enhance)
├── .env.example                    # Environment variable template (NEW)
├── Makefile                        # Automation commands (UPDATE)
│
├── scripts/                        # Setup and utility scripts
│   ├── init-db.sql                # Database initialization (NEW)
│   ├── seed-data.sql              # Sample data for testing (NEW)
│   ├── setup-local.sh             # One-command setup script (NEW)
│   └── clean-local.sh             # Reset script (NEW)
│
├── services/                       # Each service gets .env.example
│   ├── auth-service/
│   │   └── .env.example           # Service-specific env vars (NEW)
│   ├── catalog-service/
│   │   └── .env.example           # (NEW)
│   └── .../
│
└── docs/
    └── LOCAL_DEVELOPMENT.md        # Comprehensive setup guide (NEW)
```

## Phased Implementation

### Phase 1: Setup (Foundation)
**Goal**: Prepare project structure and documentation

- **T001** [Setup]: Create .env.example in project root with all infrastructure variables
- **T002** [Setup]: Create .env.example for each service with service-specific variables
- **T003** [Setup]: Update .gitignore to exclude .env files (keep .env.example tracked)
- **T004** [Setup]: Create scripts/ directory structure

**Checkpoint**: Project has environment templates and directory structure ready

---

### Phase 2: Foundational (Blocking Prerequisites)
**Goal**: Enhance Docker Compose configuration and database initialization

- **T005** [Foundational]: Review existing docker-compose.yml and identify enhancements needed
- **T006** [Foundational]: Add POSTGRES_INITDB_ARGS for UTF-8 encoding (if not present)
- **T007** [Foundational]: Verify health checks are configured for all services
- **T008** [Foundational]: Ensure volumes are named (not anonymous) for easy management
- **T009** [Foundational]: Create scripts/init-db.sql with schema creation for all services:
  ```sql
  CREATE SCHEMA IF NOT EXISTS auth;
  CREATE SCHEMA IF NOT EXISTS catalog;
  CREATE SCHEMA IF NOT EXISTS ddmrp;
  CREATE SCHEMA IF NOT EXISTS execution;
  CREATE SCHEMA IF NOT EXISTS analytics;
  CREATE SCHEMA IF NOT EXISTS ai_agent;
  ```
- **T010** [Foundational]: Mount init-db.sql in docker-compose.yml PostgreSQL service

**Checkpoint**: Infrastructure services start with proper initialization

---

### Phase 3: User Story 1 - Developer Onboarding (P1)
**Goal**: Enable new developers to start working in under 30 minutes

- **T011** [US1]: Update Makefile with `run-local` target:
  ```makefile
  run-local:
      @echo "Starting local infrastructure..."
      docker-compose up -d
      @echo "Waiting for services to be healthy..."
      ./scripts/wait-for-services.sh
      @echo "✅ All services ready!"
  ```
- **T012** [US1]: Create scripts/wait-for-services.sh to poll health checks
- **T013** [US1]: Add `stop-local` and `logs-local` Makefile targets
- **T014** [US1]: Create scripts/setup-local.sh for one-command setup:
  - Check Docker is running
  - Check prerequisites (Docker version, disk space)
  - Create .env from .env.example if not exists
  - Run docker-compose up
  - Wait for health checks
  - Display connection info
- **T015** [US1]: Test full onboarding flow on fresh machine (or fresh VM)

**Checkpoint**: Developer can run `./scripts/setup-local.sh` and have working environment

---

### Phase 4: User Story 2 - Service Development Workflow (P2)
**Goal**: Enable running services locally with proper configuration

- **T016** [US2]: Create .env.example for auth-service with:
  ```
  PORT=8081
  DATABASE_URL=postgresql://giia:giia_dev_password@localhost:5432/giia_dev?search_path=auth
  REDIS_URL=redis://:giia_redis_password@localhost:6379/0
  NATS_URL=nats://localhost:4222
  LOG_LEVEL=debug
  JWT_SECRET=dev_secret_change_in_production
  ```
- **T017** [US2]: Create .env.example for all other services (catalog, ddmrp, execution, analytics, ai-agent)
- **T018** [US2]: Verify .vscode/launch.json exists and is configured correctly (DONE in Task 2)
- **T019** [US2]: Add Makefile target `run-service SERVICE=auth` to run individual service
- **T020** [US2]: Create scripts/seed-data.sql with sample data for testing
- **T021** [US2]: Add Makefile target `seed-data` to populate test data

**Checkpoint**: Developer can start infrastructure, then run any service from IDE or terminal

---

### Phase 5: User Story 3 - Database Management (P3)
**Goal**: Provide visual tools for database inspection

- **T022** [US3]: Verify pgAdmin and Redis Commander are in docker-compose.yml under "tools" profile
- **T023** [US3]: Add Makefile target `run-tools`:
  ```makefile
  run-tools:
      docker-compose --profile tools up -d
  ```
- **T024** [US3]: Create pgAdmin server configuration JSON to auto-connect to giia-postgres
- **T025** [US3]: Update LOCAL_DEVELOPMENT.md with tool access instructions

**Checkpoint**: Developer can access pgAdmin (port 5050) and Redis Commander (port 8081)

---

### Phase 6: Polish (Cross-Cutting Concerns)
**Goal**: Documentation, troubleshooting, and edge cases

- **T026** [Polish]: Create comprehensive docs/LOCAL_DEVELOPMENT.md with:
  - Prerequisites checklist
  - Quick start (TL;DR section)
  - Detailed setup steps
  - Connection strings for each service
  - Troubleshooting section (common errors)
  - FAQ section
- **T027** [Polish]: Create Makefile target `clean-local` to remove volumes and reset data:
  ```makefile
  clean-local:
      docker-compose down -v
      @echo "⚠️  All data has been deleted. Run 'make run-local' to start fresh."
  ```
- **T028** [Polish]: Add Makefile target `status-local` to show service health
- **T029** [Polish]: Handle edge case: Port conflict detection script
- **T030** [Polish]: Handle edge case: Disk space check before starting
- **T031** [Polish]: Create scripts/backup-local.sh to export PostgreSQL data
- **T032** [Polish]: Create scripts/restore-local.sh to import PostgreSQL data
- **T033** [Polish]: Add health check visualization script (table showing status of each service)
- **T034** [Polish]: Test all Make targets and scripts
- **T035** [Polish]: Update main README.md with link to LOCAL_DEVELOPMENT.md

**Checkpoint**: Complete documentation and all edge cases handled

---

## Task Dependencies

```mermaid
graph TD
    T001[Setup env templates] --> T009[Create init scripts]
    T002[Service env examples] --> T016[Auth service config]
    T009 --> T011[Makefile run-local]
    T011 --> T014[Setup script]
    T014 --> T016
    T016 --> T019[Run service target]
    T019 --> T022[Optional tools]
    T022 --> T026[Documentation]
    T026 --> T027[Clean script]
```

**Critical Path**: T001 → T009 → T011 → T014 → T016 → T026

## Execution Strategy

### Recommended Order
1. **Day 1**: Phase 1 + Phase 2 (Foundation - all services can start)
2. **Day 2**: Phase 3 (Developer onboarding - complete setup flow)
3. **Day 3**: Phase 4 (Service development - run services locally)
4. **Day 4**: Phase 5 + Phase 6 (Tools and documentation)

### Validation Points
- After Phase 2: Run `docker-compose up` and verify all services are healthy
- After Phase 3: Fresh clone + run setup script, should work without errors
- After Phase 4: Start infrastructure, run auth-service from IDE, make HTTP request
- After Phase 6: Have another developer follow documentation and report issues

## Testing Strategy

### Manual Testing Checklist
- [ ] Fresh clone on new machine, run setup, verify all services start
- [ ] Stop services, start again, data persists
- [ ] Create product in catalog-service, restart PostgreSQL, product still exists
- [ ] Run all Make targets, verify no errors
- [ ] Simulate port conflict (run external PostgreSQL on 5432), verify error message
- [ ] Test with low disk space (if possible), verify error handling
- [ ] Test all service .env.example files load correctly

### Acceptance Tests
- Run `make run-local` → all services healthy in under 2 minutes
- Run `make clean-local` → all data deleted, can start fresh
- Follow docs/LOCAL_DEVELOPMENT.md → new developer completes setup without assistance
- Services can connect to PostgreSQL, Redis, NATS on standard ports

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Docker not installed | High | Setup script checks and provides install link |
| Insufficient resources (< 4GB RAM) | High | Pre-flight check with clear error message |
| Port conflicts (5432, 6379, 4222 in use) | Medium | Port detection script suggests alternatives |
| Docker daemon not running | High | Check docker ps first, provide clear error |
| Network issues (cannot pull images) | Low | Document how to use local image cache |

## Rollout Plan

1. **PR 1**: Foundation (Phase 1 + Phase 2) - Get docker-compose working
2. **PR 2**: Developer Onboarding (Phase 3) - Setup script and Make targets
3. **PR 3**: Service Development (Phase 4) - Service configs and seed data
4. **PR 4**: Tools + Documentation (Phase 5 + Phase 6) - Polish and edge cases

Each PR should be independently reviewable and testable.

---

**Estimated Effort**: 3-4 days
**Complexity**: Low-Medium
**Blockers**: None (Task 1-2 complete)
**Enables**: Task 4 (services need local infrastructure), Task 5+ (all service development)
