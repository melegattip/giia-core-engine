# GIIA Quick Start Guide

**Get up and running in under 10 minutes!**

---

## ⚡ Prerequisites

Before you begin, ensure you have:

| Tool | Version | Installation |
|------|---------|--------------|
| **Go** | 1.23.4+ | [go.dev/dl](https://go.dev/dl/) |
| **Docker** | 24.0+ | [docker.com](https://docs.docker.com/get-docker/) |
| **Git** | 2.40+ | [git-scm.com](https://git-scm.com/downloads) |
| **Make** | Any | Windows: `choco install make`, macOS: pre-installed |

### Verify Installation

```bash
go version          # go version go1.23.x
docker --version    # Docker version 24.x.x
git --version       # git version 2.x.x
make --version      # GNU Make x.x
```

---

## 🚀 Quick Setup (5 Minutes)

### Step 1: Clone and Initialize

```bash
# Clone the repository
git clone <repository-url>
cd giia-core-engine

# Sync Go workspace
go work sync

# Download dependencies
go mod download
```

### Step 2: Start Infrastructure

```bash
# Start PostgreSQL, Redis, and NATS
docker-compose up -d

# Verify services are healthy
docker-compose ps
# Should show: giia-postgres, giia-redis, giia-nats (all healthy)
```

### Step 3: Build Services

```bash
# Build all services
make build

# Verify binaries
ls bin/
# Should see: auth-service.exe, catalog-service.exe, etc.
```

### Step 4: Run Tests

```bash
# Run all tests
make test

# Run with coverage
make test-coverage
```

---

## 🐳 Docker Commands

| Command | Description |
|---------|-------------|
| `docker-compose up -d` | Start all infrastructure |
| `docker-compose down` | Stop all services |
| `docker-compose logs -f` | View logs |
| `docker-compose ps` | Check status |
| `docker-compose down -v` | Clean everything (including data) |

### Database Access

```bash
# PostgreSQL CLI
docker exec -it giia-postgres psql -U giia -d giia_dev

# Redis CLI
docker exec -it giia-redis redis-cli -a giia_redis_password
```

---

## 🔧 Makefile Commands

Run `make help` for all available commands:

```bash
# Development
make build              # Build all services
make test               # Run tests
make lint               # Check code quality
make fmt                # Format code

# Local Environment
make run-local          # Start Docker Compose
make stop-local         # Stop Docker Compose
make logs-local         # View logs

# Specific Services
make build-auth         # Build auth service only
make test-auth          # Test auth service only
```

---

## 🏃 Running a Service

```bash
# Using Make
make run-service SERVICE=auth

# Or directly
cd services/auth-service
go run cmd/api/main.go
```

**Service Ports:**

| Service | HTTP | gRPC |
|---------|------|------|
| Auth | 8081 | 9081 |
| Catalog | 8082 | 9082 |
| DDMRP Engine | 8083 | 9083 |
| Execution | 8084 | 9084 |
| Analytics | 8085 | 9085 |
| AI Agent | 8086 | 9086 |

---

## 📁 Project Structure

```
giia-core-engine/
├── services/          # 6 microservices
│   ├── auth-service/
│   ├── catalog-service/
│   ├── ddmrp-engine-service/
│   ├── execution-service/
│   ├── analytics-service/
│   └── ai-intelligence-hub/
├── pkg/               # Shared libraries
│   ├── config/
│   ├── logger/
│   ├── database/
│   ├── errors/
│   └── events/
├── api/proto/         # gRPC definitions
├── k8s/               # Kubernetes manifests
├── docs/              # Documentation
└── scripts/           # Utility scripts
```

---

## 🐛 Troubleshooting

### Port Already in Use

```bash
# Windows
netstat -ano | findstr :5432

# Change port in docker-compose.yml
ports:
  - "5433:5432"  # Use different host port
```

### Docker Not Running

```bash
# Start Docker Desktop
# Then retry: docker-compose up -d
```

### Go Workspace Issues

```bash
go clean -modcache
go work sync
go mod download
```

---

## 📚 Next Steps

1. **Read Architecture**: [Architecture Overview](../architecture/OVERVIEW.md)
2. **Explore API**: [API Documentation](../api/PUBLIC_RFC.md)
3. **Development Guide**: [Development Standards](../development/DEVELOPMENT_GUIDE.md)
4. **Check Status**: [Project Status](../specifications/PROJECT_STATUS.md)

---

## 🆘 Need Help?

- **Documentation**: Browse the `/docs` folder
- **Makefile**: Run `make help`
- **Logs**: `docker-compose logs -f`
- **Team**: Contact the GIIA development team

---

**Ready to code! 🚀**
