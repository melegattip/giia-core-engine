# Feature Specification: Local Development Environment

**Created**: 2025-12-09

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Developer Onboarding (Priority: P1)

As a new developer joining the GIIA team, I need to set up my local development environment quickly so I can start contributing to the project within 30 minutes.

**Why this priority**: Critical for team productivity. Every developer must be able to start working immediately. Reduces friction and onboarding time from days to minutes.

**Independent Test**: Can be fully tested by running `make run-local` on a fresh machine and verifying all infrastructure services (PostgreSQL, Redis, NATS) start successfully with health checks passing. Delivers standalone value: working local infrastructure.

**Acceptance Scenarios**:

1. **Scenario**: Fresh machine setup
   - **Given** a developer has Go 1.23+ and Docker installed
   - **When** they clone the repository and run `make run-local`
   - **Then** all infrastructure services start within 2 minutes and health checks pass

2. **Scenario**: Database initialization
   - **Given** PostgreSQL container is starting for the first time
   - **When** the container initializes
   - **Then** database schemas for all services are created automatically

3. **Scenario**: Service connectivity verification
   - **Given** all infrastructure services are running
   - **When** developer runs connection tests
   - **Then** services can connect to PostgreSQL (port 5432), Redis (port 6379), and NATS (port 4222)

---

### User Story 2 - Service Development Workflow (Priority: P2)

As a backend developer, I need to run individual services locally while connected to shared infrastructure so I can develop and test features independently.

**Why this priority**: Essential for daily development work. Enables rapid iteration without needing full cluster deployment.

**Independent Test**: Can be tested by starting infrastructure with `docker-compose up`, then running a single service (e.g., `go run services/auth-service/cmd/api/main.go`) and verifying it connects to all dependencies and responds to requests.

**Acceptance Scenarios**:

1. **Scenario**: Running single service locally
   - **Given** infrastructure is running via Docker Compose
   - **When** developer runs a service from their IDE or terminal
   - **Then** service starts, connects to dependencies, and serves requests on its designated port

2. **Scenario**: Hot reload during development
   - **Given** a service is running locally
   - **When** developer modifies Go code
   - **Then** developer can rebuild and restart quickly (< 10 seconds)

3. **Scenario**: Debugging with IDE
   - **Given** VS Code launch configuration exists
   - **When** developer uses "Debug" in VS Code
   - **Then** breakpoints work and developer can step through code

---

### User Story 3 - Database Management (Priority: P3)

As a developer, I need visual tools to inspect and manage local databases so I can debug data issues and verify migrations.

**Why this priority**: Nice-to-have quality-of-life improvement. Developers can use CLI tools if needed, but GUI tools improve productivity.

**Independent Test**: Can be tested by starting tools profile (`docker-compose --profile tools up`) and accessing pgAdmin (port 5050) and Redis Commander (port 8081) in browser.

**Acceptance Scenarios**:

1. **Scenario**: Accessing database GUI
   - **Given** infrastructure is running with tools profile
   - **When** developer navigates to http://localhost:5050
   - **Then** pgAdmin interface loads and shows GIIA databases

2. **Scenario**: Inspecting Redis cache
   - **Given** infrastructure is running with tools profile
   - **When** developer navigates to http://localhost:8081
   - **Then** Redis Commander shows cached keys and values

---

### Edge Cases

- What happens when PostgreSQL port 5432 is already in use by another application?
- How does system handle starting services before infrastructure is ready?
- What happens if Docker daemon is not running?
- How to reset all data and start fresh (clean slate)?
- How to handle multiple developers working on different branches with different schema migrations?
- What happens when disk space is low and Docker volumes fill up?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST provide Docker Compose configuration for PostgreSQL 16, Redis 7, and NATS Jetstream 2
- **FR-002**: System MUST initialize PostgreSQL with separate schemas for each microservice (auth, catalog, ddmrp, execution, analytics, ai_agent)
- **FR-003**: System MUST expose PostgreSQL on port 5432, Redis on port 6379, NATS client on port 4222, and NATS monitoring on port 8222
- **FR-004**: System MUST implement health checks for all infrastructure services (PostgreSQL, Redis, NATS)
- **FR-005**: System MUST persist data across container restarts using Docker volumes
- **FR-006**: System MUST provide pgAdmin on port 5050 and Redis Commander on port 8081 (optional tools profile)
- **FR-007**: System MUST provide Makefile commands for common operations (start, stop, clean, logs)
- **FR-008**: System MUST include initialization scripts for database schemas and seed data
- **FR-009**: System MUST configure Redis with password authentication
- **FR-010**: System MUST enable NATS Jetstream for event streaming
- **FR-011**: System MUST use Docker network isolation for service communication
- **FR-012**: System MUST provide environment variable templates (.env.example) for each service

### Key Entities

- **PostgreSQL Database**: Main relational database with separate schemas per service
- **Redis Cache**: In-memory cache for session management and caching
- **NATS Jetstream**: Message broker for event-driven communication between services
- **Docker Network**: Isolated network for local service communication
- **Docker Volumes**: Persistent storage for database data, cache, and message queue

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: New developers can start all infrastructure services in under 2 minutes with a single command
- **SC-002**: All infrastructure services pass health checks within 30 seconds of starting
- **SC-003**: Infrastructure can run continuously for 8+ hours without issues during a development session
- **SC-004**: Developers can reset all data and restart fresh in under 1 minute
- **SC-005**: Documentation allows a new developer to complete full setup without assistance
- **SC-006**: Database initialization scripts run successfully on first startup
- **SC-007**: Services can connect to all infrastructure dependencies without manual configuration
- **SC-008**: Memory usage stays under 2GB for all infrastructure containers combined
