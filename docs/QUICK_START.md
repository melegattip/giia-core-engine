# GIIA Quick Start Guide

**Start here after Task 1 completion!**

---

## ⚡ 5-Minute Setup

### Step 1: Install Prerequisites (Windows)

```powershell
# 1. Install Go 1.23+ from https://go.dev/dl/
# Download: go1.23.windows-amd64.msi
# Run installer and follow prompts

# 2. Verify Go installation
go version  # Should show: go version go1.23.x windows/amd64

# 3. Install Git (if not already installed)
# Download from: https://git-scm.com/download/win

# 4. Install Docker Desktop
# Download from: https://www.docker.com/products/docker-desktop
# Install and start Docker Desktop

# 5. Install Make (optional, for Windows)
# Option A: Use Git Bash (comes with Git for Windows)
# Option B: Install via Chocolatey: choco install make
# Option C: Use WSL2 Ubuntu
```

### Step 2: Initialize Workspace

```bash
# Open terminal in project directory
cd giia-core-engine

# Sync Go workspace (this downloads dependencies)
go work sync

# Download all dependencies
go mod download
# OR use make (if installed)
make deps

# Verify everything is working
go list -m all | head -20
```

### Step 3: Start Local Infrastructure

```bash
# Start Docker Desktop first!

# Then start PostgreSQL, Redis, and NATS
docker-compose up -d

# Verify services are running
docker-compose ps

# You should see:
# ✅ giia-postgres (healthy)
# ✅ giia-redis (healthy)
# ✅ giia-nats (healthy)

# Check logs
docker-compose logs -f

# To stop (when done):
# docker-compose down
```

### Step 4: Build Services

```bash
# Option A: Build all services (using Make)
make build

# Option B: Build manually (if Make not available)
cd services/auth-service
go build -o ../../bin/auth-service ./cmd/api/
cd ../..

# Verify binaries were created
ls bin/
# Should see: auth-service.exe (or auth-service on Linux/Mac)
```

### Step 5: Run Tests

```bash
# Run all tests
go test ./... -v -count=1

# OR using Make
make test

# Run with coverage
make test-coverage
```

---

## 🐳 Docker Commands Cheatsheet

```bash
# Start all services
docker-compose up -d

# Stop all services
docker-compose down

# View logs
docker-compose logs -f

# Restart a specific service
docker-compose restart postgres

# Check status
docker-compose ps

# Clean everything (including volumes)
docker-compose down -v

# PostgreSQL connection
docker exec -it giia-postgres psql -U giia -d giia_dev

# Redis CLI
docker exec -it giia-redis redis-cli -a giia_redis_password

# NATS monitoring
# Open browser: http://localhost:8222
```

---

## 🔧 Makefile Commands (if Make installed)

```bash
# See all commands
make help

# Setup development tools
make setup

# Build
make build              # Build all services
make build-auth         # Build specific service

# Testing
make test               # Run all tests
make test-coverage      # Generate HTML coverage report
make test-auth          # Test specific service

# Code Quality
make lint               # Run golangci-lint
make fmt                # Format all code
make vet                # Run go vet

# Local Development
make run-local          # Start Docker Compose
make stop-local         # Stop Docker Compose
make logs-local         # Tail logs

# Cleanup
make clean              # Remove binaries
make clean-all          # Deep clean (including generated code)
```

---

## 🚀 Running Services

### Auth Service (Example)

```bash
# Make sure local infrastructure is running
docker-compose ps

# Run auth service
./bin/auth-service

# OR directly with go run
cd services/auth-service
go run cmd/api/main.go
```

---

## 🐛 Troubleshooting

### Problem: `go: command not found`

**Solution:**
```bash
# Windows: Add Go to PATH
# System Properties → Environment Variables → Path
# Add: C:\Program Files\Go\bin

# Verify
go version
```

### Problem: `docker-compose: command not found`

**Solution:**
```bash
# Make sure Docker Desktop is installed and running
# In newer versions, use: docker compose (without hyphen)
docker compose up -d
```

### Problem: `Port 5432 already in use`

**Solution:**
```bash
# Check what's using the port
netstat -ano | findstr :5432

# Stop the process or change port in docker-compose.yml
# Change: "5432:5432" to "5433:5432"
```

### Problem: `make: command not found` (Windows)

**Solution:**
```bash
# Option 1: Use Git Bash instead of CMD
# Open Git Bash and run make commands

# Option 2: Run commands manually
# Example: Instead of "make build", run:
cd services/auth-service
go build -o ../../bin/auth-service ./cmd/api/
cd ../..

# Option 3: Install Make via Chocolatey
choco install make
```

### Problem: Go workspace not syncing

**Solution:**
```bash
# Clean and re-sync
go clean -modcache
go work sync
go mod download
```

---

## 📁 Project Structure Quick Reference

```
giia-core-engine/
├── services/          # 👈 All microservices here
│   ├── auth-service/         # Authentication & RBAC
│   ├── catalog-service/      # Products, Suppliers
│   ├── ddmrp-engine-service/ # Buffer calculations
│   ├── execution-service/    # Orders, Inventory
│   ├── analytics-service/    # Reports, KPIs
│   └── ai-agent-service/     # AI Chat
│
├── pkg/              # 👈 Shared libraries
│   ├── config/       # Viper configuration
│   ├── logger/       # Structured logging
│   ├── database/     # GORM PostgreSQL
│   ├── errors/       # Typed errors
│   └── events/       # NATS client
│
├── api/proto/        # 👈 gRPC definitions
├── deployments/      # 👈 Kubernetes manifests
├── scripts/          # 👈 Utility scripts
└── docs/             # 👈 Documentation
```

---

## 🔍 Useful Commands

### Check Service Dependencies

```bash
# See what packages a service imports
cd services/auth-service
go list -m all

# Check for outdated dependencies
go list -u -m all
```

### Verify Imports

```bash
# Verify all imports are correct
go mod verify

# Tidy up dependencies
go mod tidy
```

### Database Access

```bash
# PostgreSQL
docker exec -it giia-postgres psql -U giia -d giia_dev

# List databases
\l

# Connect to a specific database
\c giia_auth

# List tables
\dt

# Quit
\q
```

### Redis Access

```bash
# Redis CLI
docker exec -it giia-redis redis-cli -a giia_redis_password

# Test connection
PING
# Response: PONG

# Set a key
SET test "hello"

# Get a key
GET test

# Exit
exit
```

---

## 📚 Next Steps

1. **Read the Architecture** - [ctx/ARCHITECTURE_BALANCED.md](ctx/ARCHITECTURE_BALANCED.md)
2. **Review Development Rules** - [ctx/rules/](ctx/rules/)
3. **Setup GitHub Repository** - Follow instructions in [TASK_1_COMPLETE.md](TASK_1_COMPLETE.md)
4. **Move to Task 2** - Configure CI/CD Pipeline

---

## 🆘 Need Help?

- **Documentation**: See [README.md](README.md)
- **Architecture**: See [ctx/ARCHITECTURE_BALANCED.md](ctx/ARCHITECTURE_BALANCED.md)
- **Task 1 Details**: See [TASK_1_COMPLETE.md](TASK_1_COMPLETE.md)
- **Makefile Commands**: Run `make help`

---

**Ready to code!** 🚀
