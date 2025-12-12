# Feature Specification: Consolidate Microservices - Defer Splitting

**Created**: 2025-12-10
**Priority**: 🟡 MEDIUM
**Effort**: 2-3 days

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Archive Skeleton Services (Priority: P1)

As a technical lead, I need skeleton microservices (catalog, ddmrp, execution, analytics) archived so that deployment complexity is reduced and domain boundaries can be validated first.

**Why this priority**: Premature microservice decomposition adds deployment complexity without proven benefit. Monolith-first is Go community best practice - split after domain model validation.

**Independent Test**: Verify only auth-service exists in `services/` and archived services moved to `archive/`.

**Acceptance Scenarios**:

1. **Scenario**: Skeleton services archived
   - **Given** 4 skeleton services with <10 files each
   - **When** Consolidation is performed
   - **Then** Services moved to `archive/` directory with documentation explaining decision

2. **Scenario**: Go workspace updated
   - **Given** go.work references 5 services
   - **When** Services archived
   - **Then** go.work only references auth-service

3. **Scenario**: Deployment simplified
   - **Given** Docker Compose with 5 service definitions
   - **When** Consolidation complete
   - **Then** Docker Compose defines single auth-service container

---

### User Story 2 - Document Service Consolidation Decision (Priority: P1)

As a team member, I need an Architecture Decision Record (ADR) explaining why services were consolidated so that the rationale is clear for future planning.

**Why this priority**: Important for team alignment and preventing premature re-splitting. Documents lessons learned.

**Independent Test**: ADR document exists in `docs/architecture/adr/` explaining consolidation decision.

**Acceptance Scenarios**:

1. **Scenario**: ADR created
   - **Given** Service consolidation decision
   - **When** ADR is written
   - **Then** Document explains: context, decision, consequences, future split criteria

2. **Scenario**: README updated
   - **Given** Root README references 6 microservices
   - **When** Consolidation complete
   - **Then** README describes monolithic approach with planned future decomposition

3. **Scenario**: Team notified
   - **Given** Service architecture change
   - **When** Consolidation complete
   - **Then** Team informed via Slack/email with link to ADR

---

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Directories `services/catalog-service/`, `services/ddmrp-engine-service/`, `services/execution-service/`, `services/analytics-service/` MUST be moved to `archive/`
- **FR-002**: go.work MUST only reference `services/auth-service`
- **FR-003**: Docker Compose MUST define single service container
- **FR-004**: ADR MUST be created in `docs/architecture/adr/001-consolidate-services.md`
- **FR-005**: README MUST be updated to reflect monolithic structure

### Key Entities

- **Archived Service**: Skeleton microservice moved to archive/ for future reference
- **ADR (Architecture Decision Record)**: Document explaining consolidation rationale
- **Monolithic Service**: Single service containing all functionality until domain boundaries validated

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Only `services/auth-service/` exists in services/ directory
- **SC-002**: Archived services exist in `archive/` with README explaining preservation
- **SC-003**: Application deploys as single service successfully
- **SC-004**: ADR document published and accessible to team
- **SC-005**: Deployment complexity reduced (single container vs 5)
