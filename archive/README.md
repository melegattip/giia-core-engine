# Archive - Preserved Microservices

**Status**: Archived
**Date Archived**: 2025-12-12
**Reason**: Consolidation to monolithic architecture (see [ADR 001](../docs/architecture/adr/001-consolidate-to-monolithic-architecture.md))

---

## Overview

This directory contains **skeleton microservices** that were archived during the architectural consolidation to a monolithic approach. These services are preserved for:

1. **Future Reference**: Domain research and initial design decisions
2. **Historical Context**: Understanding the original microservices vision
3. **Future Decomposition**: When validated, these can serve as templates for service boundaries

---

## Archived Services

### 1. catalog-service
**Purpose**: Product catalog, supplier management, buffer profile configuration

**Planned Functionality**:
- Product master data (SKUs, descriptions, attributes)
- Supplier and vendor management
- Buffer profile definitions (Lead Time Factor, Variability Factor)
- Material categorization (Make-to-Stock, Make-to-Order, etc.)

**Status**: Skeleton (4 files)
**Future**: Will be implemented as module within monolith, may split when:
- Product catalog exceeds 100,000 SKUs
- Multiple teams work on catalog features
- Independent scaling of catalog queries needed

---

### 2. ddmrp-engine-service
**Purpose**: Core DDMRP calculation engine - buffer sizing, positioning, and replenishment

**Planned Functionality**:
- Average Daily Usage (ADU) calculation
- Decoupled Lead Time (DLT) determination
- Buffer zone calculation (Red, Yellow, Green)
- Net Flow Equation execution
- Planned Spike Adjustments
- Buffer status alerts (Top of Red, Top of Yellow)

**Status**: Skeleton (4 files)
**Future**: Core business logic - highest priority for implementation
- May remain in monolith for performance (tight coupling with execution)
- Split only if calculation latency requires dedicated scaling
- Candidate for worker pool pattern within monolith first

---

### 3. execution-service
**Purpose**: Order management, inventory transactions, ERP integration

**Planned Functionality**:
- Purchase Order creation and tracking
- Inventory adjustment transactions
- ERP connector framework (SAP, Odoo, etc.)
- Stock level synchronization
- Supply execution recommendations
- Order status workflow

**Status**: Skeleton (4 files)
**Future**: Tightly coupled with DDMRP engine
- Likely remains in monolith with ddmrp-engine
- Separate service only if ERP integrations become complex standalone system
- Integration patterns will emerge during implementation

---

### 4. analytics-service
**Purpose**: KPI dashboards, reporting, data visualization, variance analysis

**Planned Functionality**:
- DDMRP KPIs (On-Time Delivery, Buffer Status Distribution, etc.)
- Variance analysis (planned vs actual)
- Trend analysis and forecasting
- Custom report builder
- Data export (CSV, Excel, PDF)
- Historical data aggregation

**Status**: Skeleton (4 files)
**Future**: Clear candidate for eventual decomposition
- Read-heavy workload vs write-heavy execution service
- Different scaling characteristics
- Can be split when reporting queries impact transaction performance
- Good candidate for separate read replica database

---

### 5. ai-agent-service
**Purpose**: AI-powered chat assistant, proactive insights, anomaly detection

**Planned Functionality**:
- ChatGPT/Claude integration
- Natural language query interface
- Proactive buffer health alerts
- Supply chain anomaly detection
- Conversational analytics
- Context-aware recommendations

**Status**: Skeleton (4 files)
**Future**: Strong candidate for microservice
- Different technology stack (Python might be better for ML/AI)
- Independent scaling for AI workloads
- External API rate limiting isolation
- Experimental features without impacting core platform
- May use different deployment patterns (serverless, GPU instances)

---

## Why These Were Archived

### Problem Statement

At the time of archival (2025-12-12), the GIIA project had:
- **1 fully developed service**: auth-service (77+ files, 98% test coverage)
- **5 skeleton services**: 4 files each, no real functionality

This created:
- **Deployment complexity**: Managing 6 services when only 1 was functional
- **Development friction**: Multi-service setup for single-service development
- **Unclear boundaries**: Domain boundaries not validated through implementation
- **Premature optimization**: Microservices before understanding the domain

### The Monolith-First Approach

Industry best practices (Sam Newman, Martin Fowler) recommend:
1. Build as a monolith first
2. Validate domain model through implementation
3. Let natural boundaries emerge
4. Split when there's a **proven need**, not a hypothetical one

### What We're Doing Instead

The auth-service will become the **core monolithic service** where all functionality is implemented:
- Maintain **Clean Architecture** with clear internal module boundaries
- Organize code by **domain** (catalog, ddmrp, execution, analytics, ai)
- Use **internal packages** with explicit interfaces
- Preserve **option value** for future microservice decomposition

---

## When to Resurrect These Services

### Prerequisites for Service Decomposition

Before splitting ANY service out of the monolith, we must have:

#### Technical Prerequisites
- [ ] Clear, stable module boundaries validated through implementation
- [ ] <5% code sharing between potential services
- [ ] Well-defined interfaces between modules
- [ ] Comprehensive test suite (>90% coverage)
- [ ] Proven performance bottleneck that vertical scaling can't solve

#### Organizational Prerequisites
- [ ] Multiple teams working on different domains
- [ ] Need for independent deployment schedules
- [ ] Clear service ownership established
- [ ] DevOps capacity to manage multiple services

#### Business Prerequisites
- [ ] >10,000 daily active users
- [ ] Different features have vastly different usage patterns
- [ ] SLA requirements differ by feature
- [ ] Cost of microservices < cost of scaling monolith

### Recommended Decomposition Order

When the time comes, split services in this order:

1. **ai-agent-service** (First)
   - Least coupled to core business logic
   - Different tech stack might be beneficial
   - External API integration isolation
   - Experimental nature suits independent deployment

2. **analytics-service** (Second)
   - Clear read-heavy vs write-heavy boundary
   - Different scaling characteristics
   - Can use read replicas
   - Less critical for core operations

3. **catalog-service** (Third)
   - Relatively independent domain
   - Potential for high query volume
   - Multiple teams might work on product data
   - Good candidate for caching strategies

4. **execution-service + ddmrp-engine-service** (Last, if ever)
   - Tightly coupled business logic
   - Splitting might reduce performance
   - Only split if absolutely necessary
   - Consider keeping as single service even in microservices architecture

---

## Preserved Domain Knowledge

### Key Design Decisions in Skeletons

Even though these are skeletons, they contain important decisions:

1. **Service Naming**: Reflects domain language
2. **Module Structure**: cmd/server pattern established
3. **Containerization**: Docker-ready from start
4. **Configuration**: Environment-based config pattern

### Directory Structure Pattern

Each archived service follows:
```
service-name/
├── cmd/
│   └── server/
│       └── main.go          # Entry point
├── Dockerfile               # Container definition
├── .env.example            # Configuration template
└── go.mod                  # Module definition
```

This pattern can be reused when splitting services.

---

## Migration Path

### From Monolith to Microservice (Future)

When it's time to extract a service:

1. **Identify Module**: Find well-bounded module in monolith
2. **Extract Interface**: Define clean service API
3. **Copy Code**: Start with code from monolith module
4. **Add Infrastructure**: gRPC server, health checks, metrics
5. **Gradual Migration**: Use strangler fig pattern
6. **Monitor**: Ensure no performance regression
7. **Cleanup**: Remove old code from monolith once validated

### Use Archived Services As Templates

The archived services provide:
- Project structure template
- Naming conventions
- Docker configuration baseline
- Environment variable patterns

---

## References

- **ADR 001**: [Consolidate to Monolithic Architecture](../docs/architecture/adr/001-consolidate-to-monolithic-architecture.md)
- **Monolith First**: https://martinfowler.com/bliki/MonolithFirst.html
- **Microservices Prerequisites**: https://martinfowler.com/bliki/MicroservicePrerequisites.html
- **Building Microservices**: Sam Newman, 2nd Edition

---

## Notes

- These services are **not deleted** - they're preserved
- This is **not a failure** - it's strategic simplification
- Domain boundaries will be **validated through implementation**
- Microservices remain **an option**, not a mandate
- The goal is **business value**, not architectural purity

**Last Updated**: 2025-12-12
**Next Review**: When monolith reaches 50,000 LOC or performance bottleneck identified
