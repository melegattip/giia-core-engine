# ADR 001: Consolidate to Monolithic Architecture

**Status**: Accepted
**Date**: 2025-12-12
**Decision Makers**: Technical Lead, Development Team
**Tags**: architecture, microservices, monolith, deployment

---

## Context

The GIIA Core Engine project was initially designed with a microservices architecture consisting of 6 services:

1. **auth-service** (Fully developed - 77+ files)
2. **catalog-service** (Skeleton - 4 files)
3. **ddmrp-engine-service** (Skeleton - 4 files)
4. **execution-service** (Skeleton - 4 files)
5. **analytics-service** (Skeleton - 4 files)
6. **ai-agent-service** (Skeleton - 4 files)

### Current State Analysis

Only the auth-service has been fully developed with comprehensive functionality including:
- Authentication and authorization
- Multi-tenancy support
- RBAC (Role-Based Access Control)
- JWT token management
- User and organization management
- Comprehensive test suite (98.2% coverage)

The remaining 5 services are minimal skeletons containing only:
- Basic Dockerfile
- Environment configuration template
- Minimal main.go entry point
- go.mod module definition

### Challenges with Current Architecture

1. **Premature Decomposition**: Domain boundaries have not been validated through implementation
2. **Deployment Complexity**: Managing 6 services (5 of which are non-functional) adds unnecessary overhead
3. **Development Friction**: Multi-service setup slows down local development and testing
4. **Unclear Boundaries**: Without implementation, it's unclear if the service boundaries are correct
5. **Infrastructure Cost**: Running 6 services in development/staging environments wastes resources
6. **Cognitive Load**: Developers must navigate 6 services when only 1 is active

### Industry Best Practices

The Go community and industry experts recommend the **"monolith-first" approach**:

- **Sam Newman** (Building Microservices): "Start with a monolith and split when you have a reason"
- **Martin Fowler**: "MonolithFirst - build the application as a monolith initially"
- **Go Community**: Go's simplicity favors monolithic services that scale vertically first

Microservices should only be introduced when:
1. Domain boundaries are **proven** and stable
2. Teams are organized around bounded contexts
3. Independent scaling is **required** (not just desired)
4. Deployment independence provides **measurable** value
5. The monolith becomes a genuine bottleneck

---

## Decision

**We will consolidate the skeleton microservices into a monolithic architecture, preserving only the auth-service as the active service.**

### What This Means

1. **Archive Skeleton Services**: Move catalog, ddmrp-engine, execution, analytics, and ai-agent services to `archive/`
2. **Preserve Code**: Keep archived services for future reference and domain research
3. **Simplify Deployment**: Deploy single service + infrastructure (postgres, redis, nats)
4. **Single Go Module**: Consolidate development into auth-service (to be renamed later)
5. **Future Decomposition**: Split services when domain model is validated and boundaries are clear

### What This Does NOT Mean

- This is **not** abandoning microservices permanently
- This is **not** a failure of the original architecture design
- This is **not** preventing future service decomposition
- This is **strategic simplification** following industry best practices

---

## Consequences

### Positive Consequences

1. **Reduced Complexity**
   - Single service to build, test, and deploy
   - Simplified CI/CD pipeline
   - Fewer moving parts in development environment

2. **Faster Development**
   - No inter-service communication overhead during development
   - Easier to refactor across domain boundaries
   - Faster iteration cycles

3. **Better Domain Understanding**
   - Implement all features first to understand true boundaries
   - Easier to identify cohesive modules
   - Natural emergence of service boundaries through implementation

4. **Cost Savings**
   - Fewer infrastructure resources in dev/staging
   - Simpler monitoring and logging
   - Reduced DevOps overhead

5. **Simplified Testing**
   - Integration tests don't require multiple services
   - Easier to achieve high test coverage
   - Faster test execution

### Negative Consequences / Trade-offs

1. **Coupling Risk**
   - Need discipline to maintain clean module boundaries
   - Risk of creating tightly coupled code
   - **Mitigation**: Maintain Clean Architecture within the monolith

2. **Future Split Complexity**
   - Splitting a monolith requires effort
   - Need clear module boundaries from the start
   - **Mitigation**: Use internal packages with clear interfaces

3. **Single Point of Failure**
   - One service handles all requests
   - **Mitigation**: Implement proper error handling, health checks, graceful degradation

4. **Scaling Limitations**
   - Cannot scale individual components independently
   - **Mitigation**: Profile and optimize; vertical scaling is sufficient for initial phase

---

## Future Service Decomposition Criteria

We will consider splitting services when **all** of the following conditions are met:

### Technical Criteria

1. **Clear Domain Boundaries**
   - Bounded contexts are well-defined and stable
   - Module dependencies are minimal and one-directional
   - Each potential service has <5% shared code with others

2. **Performance Requirements**
   - Specific component requires independent scaling
   - Vertical scaling of monolith is insufficient
   - Different components have conflicting resource needs

3. **Team Organization**
   - Multiple teams working on different domains
   - Need for independent deployment schedules
   - Clear ownership boundaries established

### Business Criteria

1. **User Load**
   - Platform serves >10,000 daily active users
   - Different features have vastly different usage patterns
   - SLA requirements differ by feature

2. **Development Velocity**
   - Monolith deployment is slowing down releases
   - Different features require independent release cycles
   - Hotfixes in one domain are blocked by others

3. **Cost Justification**
   - Infrastructure cost of microservices < cost of scaling monolith
   - Developer productivity gains justify deployment complexity
   - Monitoring and operational costs are acceptable

---

## Implementation Plan

### Phase 1: Archive Services (Completed: TBD)
- [x] Move skeleton services to `archive/` directory
- [x] Create `archive/README.md` explaining preservation rationale
- [x] Update documentation to reflect current state

### Phase 2: Update Configuration (Completed: TBD)
- [x] Update `go.work` to reference only auth-service
- [x] Update `docker-compose.yml` for single service
- [x] Update `Makefile` and build scripts
- [x] Verify deployment works

### Phase 3: Documentation (Completed: TBD)
- [x] Update README.md to reflect monolithic architecture
- [x] Update architecture diagrams
- [x] Document future decomposition criteria
- [x] Communicate change to team

---

## Monitoring and Review

### Success Metrics

- **Development velocity**: Time to implement new features
- **Deployment frequency**: How often we can deploy safely
- **Test coverage**: Maintain >90% coverage
- **Build time**: CI/CD pipeline execution time
- **Onboarding time**: Time for new developers to become productive

### Review Schedule

- **3 months**: Review development velocity and team satisfaction
- **6 months**: Assess if domain boundaries are clear
- **12 months**: Evaluate if service decomposition is warranted

### Triggers for Reconsideration

If any of the following occur, we should reconsider microservices:
1. Monolith build time exceeds 5 minutes
2. Deployment frequency drops below weekly
3. Teams are blocked waiting for others' changes
4. Performance profiling shows clear independent scaling needs

---

## References

- [Building Microservices, 2nd Edition](https://samnewman.io/books/building_microservices_2nd_edition/) - Sam Newman
- [MonolithFirst](https://martinfowler.com/bliki/MonolithFirst.html) - Martin Fowler
- [Microservices Prerequisites](https://martinfowler.com/bliki/MicroservicePrerequisites.html) - Martin Fowler
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html) - Robert C. Martin
- [Go Monolith Best Practices](https://threedots.tech/post/microservices-or-monolith-its-detail/)

---

## Notes

- This ADR preserves the **option value** of microservices while reducing current complexity
- Archived services contain valuable domain research and should be preserved
- The monolith will maintain **internal module boundaries** to facilitate future decomposition
- Clean Architecture principles will be strictly enforced to prevent coupling
- This decision can be reversed when business and technical criteria are met

---

**Approved By**: Technical Lead
**Implementation Date**: 2025-12-12
**Review Date**: 2026-03-12 (3 months)
