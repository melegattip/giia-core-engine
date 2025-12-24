# GIIA Core Engine - Documentation Hub

> **GIIA** (Gestión Inteligente de Inventario con IA) - AI-Powered DDMRP Inventory Management Platform

**Last Updated**: 2025-12-23  
**Version**: 1.0  
**Phase 1 Completion**: 93%

---

## 📖 Documentation Index

This folder contains all comprehensive documentation for the GIIA Core Engine project, organized by category for easy navigation.

### Quick Navigation

| Category | Description | Start Here |
|----------|-------------|------------|
| **🚀 Getting Started** | Setup and onboarding | [Quick Start](./getting-started/QUICK_START.md) |
| **🏗️ Architecture** | System design and patterns | [Architecture Overview](./architecture/OVERVIEW.md) |
| **📡 API Reference** | All endpoints and contracts | [API Documentation](./api/PUBLIC_RFC.md) |
| **🔧 Development** | Coding standards and guides | [Development Guide](./development/DEVELOPMENT_GUIDE.md) |
| **📊 Specifications** | Features and requirements | [Specs Index](./specifications/README.md) |

---

## 📁 Documentation Structure

```
docs/
├── README.md                           # This file - Documentation Index
│
├── getting-started/                    # 🚀 Setup and Onboarding
│   ├── QUICK_START.md                 # 5-minute setup guide
│   ├── LOCAL_DEVELOPMENT.md           # Detailed local dev environment
│   └── KUBERNETES_SETUP.md            # Kubernetes deployment guide
│
├── architecture/                       # 🏗️ System Architecture
│   ├── OVERVIEW.md                    # High-level architecture overview
│   ├── MICROSERVICES.md               # Microservices design
│   ├── CLEAN_ARCHITECTURE.md          # Clean Architecture patterns
│   ├── DATA_MODEL.md                  # Database schema and entities
│   └── DDMRP_METHODOLOGY.md           # DDMRP implementation approach
│
├── api/                               # 📡 API Documentation
│   ├── PUBLIC_RFC.md                  # Complete API reference (RFC format)
│   ├── AUTH_SERVICE_API.md            # Auth endpoints
│   ├── CATALOG_SERVICE_API.md         # Catalog endpoints
│   └── GRPC_CONTRACTS.md              # gRPC service definitions
│
├── development/                        # 🔧 Development Standards
│   ├── DEVELOPMENT_GUIDE.md           # Main development guidelines
│   ├── CODING_STANDARDS.md            # Go coding conventions
│   ├── ERROR_HANDLING.md              # Error handling patterns
│   ├── TESTING_STANDARDS.md           # Unit testing requirements
│   ├── LINTING_GUIDE.md               # Code quality tools
│   └── GIT_WORKFLOW.md                # Branching and versioning
│
├── specifications/                     # 📊 Feature Specifications
│   ├── README.md                      # Specifications methodology
│   ├── PROJECT_STATUS.md              # Current project status
│   └── REQUIREMENTS_ANALYSIS.md       # Business requirements gaps
│
└── templates/                          # 📝 Document Templates
    ├── spec-template.md               # Feature specification template
    └── plan-template.md               # Implementation plan template
```

---

## 🎯 Project Overview

GIIA is a **SaaS platform** implementing **DDMRP (Demand Driven Material Requirements Planning)** with AI-powered assistance. The platform helps manufacturing and distribution companies:

- 📊 **Optimize inventory levels** using buffer zone management
- 🤖 **Leverage AI insights** for supply chain decisions
- 📈 **Track real-time KPIs** and performance metrics
- 🔄 **Integrate with ERPs** like SAP and Odoo
- 🏢 **Support multi-tenant operations** with role-based access

### Current Status

| Service | Status | Completion |
|---------|--------|------------|
| **Auth Service** | 🟢 Advanced | 95% |
| **Catalog Service** | 🟢 Advanced | 85% |
| **DDMRP Engine** | ⏸️ Skeleton | 0% |
| **Execution Service** | ⏸️ Skeleton | 0% |
| **Analytics Service** | ⏸️ Skeleton | 0% |
| **AI Intelligence Hub** | 🟡 In Progress | 40% |

**Legend**: ✅ Complete | 🟢 Advanced (>75%) | 🟡 Partial (<75%) | ⏸️ Pending

---

## 🏗️ Technology Stack

| Layer | Technology |
|-------|------------|
| **Language** | Go 1.23.4 |
| **API** | gRPC (internal), REST (external), WebSocket (AI chat) |
| **Database** | PostgreSQL 16, Redis 7 |
| **Message Bus** | NATS Jetstream |
| **Container** | Docker, Kubernetes |
| **CI/CD** | GitHub Actions |
| **Observability** | Prometheus, Grafana, Loki |

---

## 📚 How to Use This Documentation

### For New Team Members
1. Start with [Quick Start](./getting-started/QUICK_START.md)
2. Read [Architecture Overview](./architecture/OVERVIEW.md)
3. Review [Development Guide](./development/DEVELOPMENT_GUIDE.md)
4. Check [Project Status](./specifications/PROJECT_STATUS.md)

### For API Consumers
1. Read [API Documentation](./api/PUBLIC_RFC.md)
2. Review authentication in [Auth Service API](./api/AUTH_SERVICE_API.md)
3. Check [gRPC Contracts](./api/GRPC_CONTRACTS.md) for internal communication

### For Developers
1. Follow [Coding Standards](./development/CODING_STANDARDS.md)
2. Use [Error Handling](./development/ERROR_HANDLING.md) patterns
3. Write tests per [Testing Standards](./development/TESTING_STANDARDS.md)
4. Run [Linting Guide](./development/LINTING_GUIDE.md) before commits

---

## 🔗 Related Resources

- **📘 Main README**: [/README.md](../README.md)
- **📋 Specifications**: [/specs](../specs/)
- **🛠️ Scripts**: [/scripts](../scripts/)
- **☸️ Kubernetes**: [/k8s](../k8s/)

---

## 📝 Contributing to Documentation

When updating documentation:

1. Follow the established structure
2. Use Markdown formatting consistently
3. Keep content up-to-date with code changes
4. Include code examples where appropriate
5. Update the index (this file) when adding new documents

---

**Maintained by the GIIA Development Team** 🚀
