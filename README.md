# GIIA Core Engine - Monorepo

> **GIIA** (Gestión Inteligente de Inventario con IA) - AI-Powered DDMRP Inventory Management Platform

[![Go Version](https://img.shields.io/badge/Go-1.23-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Architecture](https://img.shields.io/badge/Architecture-Microservices-green.svg)](ctx/ARCHITECTURE_BALANCED.md)

## 📖 Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Monorepo Structure](#monorepo-structure)
- [Getting Started](#getting-started)
- [Development](#development)
- [Testing](#testing)
- [Deployment](#deployment)
- [Documentation](#documentation)
- [Contributing](#contributing)

---

## 🎯 Overview

GIIA is a SaaS platform that implements **DDMRP (Demand Driven Material Requirements Planning)** with AI-powered assistance. The platform helps manufacturing and distribution companies optimize their inventory levels, reduce stockouts, and improve supply chain efficiency.

### Key Features

- 📊 **DDMRP Buffer Management** - Automated calculation of buffer zones (Red/Yellow/Green)
- 🤖 **AI Assistant** - Intelligent chat interface for supply chain insights
- 📈 **Real-time Analytics** - KPI dashboards and variance analysis
- 🔄 **ERP Integration** - Connectors for SAP, Odoo, and custom systems
- 🏢 **Multi-tenancy** - Secure isolation for enterprise clients
- 🔐 **RBAC** - Role-based access control with fine-grained permissions

---

## 🏗️ Architecture

This project follows the **Balanced Microservices Architecture** as defined in [ARCHITECTURE_BALANCED.md](ctx/ARCHITECTURE_BALANCED.md).

### Microservices

```
┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐
│  Auth Service   │  │ Catalog Service │  │ DDMRP Engine    │
│  (Multi-tenant) │  │  (Master Data)  │  │ (Core Logic)    │
└─────────────────┘  └─────────────────┘  └─────────────────┘
┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐
│ Execution Svc   │  │ Analytics Svc   │  │  AI Agent Svc   │
│ (Orders/Inv)    │  │ (KPIs/Reports)  │  │ (ChatGPT)       │
└─────────────────┘  └─────────────────┘  └─────────────────┘
```

### Technology Stack

- **Language**: Go 1.23
- **API**: gRPC (internal), REST (external), WebSocket (AI chat)
- **Database**: PostgreSQL 16, Redis 7
- **Message Bus**: NATS Jetstream
- **Container Orchestration**: Kubernetes
- **CI/CD**: GitHub Actions
- **Observability**: Prometheus, Grafana, Loki

---

## 📁 Monorepo Structure

```
giia-core-engine/
├── services/                     # Microservices
│   ├── auth-service/            # Authentication, Multi-tenancy, RBAC
│   ├── catalog-service/         # Products, Suppliers, Buffer Profiles
│   ├── ddmrp-engine-service/    # Buffer calculations, CPD, Replenishment
│   ├── execution-service/       # Orders, Inventory, ERP integrations
│   ├── analytics-service/       # KPIs, Reports, Projections
│   └── ai-agent-service/        # AI Chat, Proactive Analysis
│
├── pkg/                         # Shared Libraries
│   ├── config/                  # Configuration management (Viper)
│   ├── logger/                  # Structured logging (Zerolog)
│   ├── database/                # Database connection pool (GORM)
│   ├── errors/                  # Typed error system
│   ├── events/                  # NATS event publisher/subscriber
│   ├── middleware/              # Common HTTP/gRPC middleware
│   ├── monitoring/              # Prometheus metrics
│   └── utils/                   # Common utilities
│
├── api/                         # API Definitions
│   └── proto/                   # Protocol Buffer definitions
│       ├── auth/v1/
│       ├── catalog/v1/
│       ├── ddmrp/v1/
│       ├── execution/v1/
│       ├── analytics/v1/
│       └── ai/v1/
│
├── deployments/                 # Kubernetes Manifests
│   ├── dev/                    # Development environment
│   ├── staging/                # Staging environment
│   └── prod/                   # Production environment
│
├── migrations/                  # Database Migrations
│   ├── auth/
│   ├── catalog/
│   └── ...
│
├── scripts/                     # Utility Scripts
│   ├── setup.sh
│   ├── seed-data.sh
│   └── backup-db.sh
│
├── docs/                        # Documentation
│   ├── architecture/           # Architecture diagrams
│   └── api/                    # API documentation
│
├── ctx/                         # Context (original docs, references)
│   ├── rules/                  # Development rules
│   └── ARCHITECTURE_BALANCED.md
│
├── go.work                      # Go workspace (monorepo magic!)
├── Makefile                     # Build automation
├── docker-compose.yml           # Local development stack
├── .gitignore
├── .editorconfig
└── README.md                    # You are here!
```

---

## 🚀 Getting Started

### Prerequisites

- **Go 1.23+** - [Install](https://go.dev/dl/)
- **Docker & Docker Compose** - [Install](https://docs.docker.com/get-docker/)
- **Make** - Usually pre-installed on macOS/Linux
- **Protocol Buffers** - `brew install protobuf` (macOS) or [Download](https://grpc.io/docs/protoc-installation/)
- **Git** - [Install](https://git-scm.com/downloads)

### Quick Start

```bash
# 1. Clone the repository
git clone https://github.com/yourusername/giia-core-engine.git
cd giia-core-engine

# 2. One-command setup (recommended)
make setup-local

# Or step-by-step:
# 2a. Setup development tools
make setup

# 2b. Start local infrastructure (PostgreSQL, Redis, NATS)
make run-local

# 3. Download dependencies
make deps

# 4. Build all services
make build

# 5. Run tests
make test

# 6. Run a service locally (example: auth-service)
./bin/auth-service
```

**📘 For detailed setup instructions, see [Local Development Guide](docs/LOCAL_DEVELOPMENT.md)**

---

## 💻 Development

### Makefile Commands

Run `make help` to see all available commands:

```bash
make help                # Show all commands
make build               # Build all services
make build-auth          # Build specific service
make test                # Run all tests
make test-coverage       # Generate coverage report
make lint                # Run linters
make fmt                 # Format code
make proto               # Generate protobuf code
make docker-build        # Build Docker images
make run-local           # Start local dev environment
make clean               # Clean build artifacts
```

### Working with Go Workspace

This monorepo uses **Go workspaces** (go.work). Benefits:

✅ **Local development** - Edit shared packages (`pkg/*`) and see changes immediately in services
✅ **No version management** - No need to publish/version shared libraries during development
✅ **Type safety** - Full IDE autocomplete across services

```bash
# Sync workspace after pulling changes
go work sync

# Add a new module to workspace
go work use ./services/new-service
```

### Adding a New Service

```bash
# 1. Create service directory
mkdir -p services/my-new-service/{cmd/server,internal/{domain,application,adapter,infrastructure}}

# 2. Initialize Go module
cd services/my-new-service
go mod init github.com/giia/giia-core-engine/services/my-new-service

# 3. Add to workspace
cd ../..
go work use ./services/my-new-service

# 4. Update Makefile SERVICES variable
```

### Code Style

- **Go**: Follow [Effective Go](https://go.dev/doc/effective_go) and project rules in [ctx/rules/02-go-standards.mdc](ctx/rules/02-go-standards.mdc)
- **Formatting**: Run `make fmt` before committing
- **Linting**: Run `make lint` and fix all issues
- **Testing**: Maintain >80% code coverage

---

## 🧪 Testing

### Unit Tests

```bash
# Run all tests
make test

# Run tests for specific service
make test-auth

# Run with coverage
make test-coverage
open coverage.html
```

### Integration Tests

```bash
# Start test database
make run-local

# Run integration tests
go test -tags=integration ./...
```

### End-to-End Tests

```bash
# Deploy to dev cluster
make k8s-dev-deploy

# Run E2E tests
go test -tags=e2e ./tests/e2e/...
```

---

## 🚢 Deployment

### Local Development

```bash
# Start infrastructure with health checks
make run-local

# (Optional) Start development tools (pgAdmin, Redis Commander)
make run-tools

# Run service locally
make run-service SERVICE=auth

# Or run directly
go run services/auth-service/cmd/api/main.go
```

**📘 See [Local Development Guide](docs/LOCAL_DEVELOPMENT.md) for complete instructions**

### Kubernetes (Development)

```bash
# Build Docker images
make docker-build

# Deploy to dev cluster
make k8s-dev-deploy

# Check status
kubectl get pods -n giia-dev

# View logs
make k8s-logs
```

### Kubernetes (Production)

See [deployments/README.md](deployments/README.md) for production deployment guide.

---

## 📚 Documentation

- **Local Development**: [docs/LOCAL_DEVELOPMENT.md](docs/LOCAL_DEVELOPMENT.md) ⭐ **Start Here!**
- **Architecture**: [ctx/ARCHITECTURE_BALANCED.md](ctx/ARCHITECTURE_BALANCED.md)
- **Development Rules**: [ctx/rules/](ctx/rules/)
- **API Documentation**: [docs/api/](docs/api/)
- **Service Docs**:
  - [Auth Service](services/auth-service/README.md)
  - [Catalog Service](services/catalog-service/README.md)
  - [DDMRP Engine](services/ddmrp-engine-service/README.md)
  - [Execution Service](services/execution-service/README.md)
  - [Analytics Service](services/analytics-service/README.md)
  - [AI Agent Service](services/ai-agent-service/README.md)

---

## 🤝 Contributing

### Branching Strategy

```bash
# Format: feature/PROJ-[number]-[description]
git checkout -b feature/GIIA-123-add-buffer-calculation

# Develop, test, commit
git add .
git commit -m "feat(ddmrp): implement buffer zone calculation"

# Push and create PR
git push origin feature/GIIA-123-add-buffer-calculation
```

### Commit Message Format

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
feat(scope): add new feature
fix(scope): bug fix
docs(scope): documentation change
test(scope): add tests
refactor(scope): code refactoring
```

### Pull Request Process

1. Ensure all tests pass: `make test`
2. Run linters: `make lint`
3. Update documentation if needed
4. Get minimum 2 code reviews
5. Squash and merge to develop

---

## 📊 Project Status

**Phase 1: Foundation (Months 1-3)** - 🚧 In Progress

- [x] Task 1: Setup monorepo structure ✅
- [x] Task 2: CI/CD pipeline ✅
- [x] Task 3: Local development environment ✅
- [ ] Task 4: Shared infrastructure packages
- [ ] Task 5: Auth/IAM service with multi-tenancy
- [ ] Task 6: RBAC implementation
- [ ] Task 7: gRPC server in Auth
- [ ] Task 8: NATS event system
- [ ] Task 9: Catalog service skeleton
- [ ] Task 10: Kubernetes dev cluster

---

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 👥 Team

- **Tech Lead**: [Your Name]
- **Backend Engineers**: [Team Members]
- **DevOps Engineer**: [Name]
- **QA Engineer**: [Name]

---

## 📞 Support

- **Email**: support@giia.io
- **Slack**: #giia-dev
- **Issues**: [GitHub Issues](https://github.com/yourusername/giia-core-engine/issues)

---

**Built with ❤️ by the GIIA Team**
