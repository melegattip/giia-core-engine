# Local Development Environment Guide

This guide will help you set up and run the GIIA Core Engine locally for development.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [Detailed Setup](#detailed-setup)
- [Infrastructure Services](#infrastructure-services)
- [Running Services](#running-services)
- [Development Tools](#development-tools)
- [Common Operations](#common-operations)
- [Troubleshooting](#troubleshooting)
- [FAQ](#faq)

---

## Prerequisites

Before you begin, ensure you have the following installed:

### Required

- **Docker 24.0+** with Docker Compose v2
  - [Install Docker Desktop](https://docs.docker.com/get-docker/) (includes Docker Compose)
  - Verify: `docker --version && docker-compose --version`

- **Go 1.23.4+** for running services locally
  - [Install Go](https://golang.org/doc/install)
  - Verify: `go version`

- **Git** for version control
  - Verify: `git --version`

### Optional (Recommended)

- **Make** for running automation commands
  - Windows: Install via Chocolatey `choco install make` or use Git Bash
  - macOS: Included with Xcode Command Line Tools
  - Linux: Usually pre-installed

- **Visual Studio Code** or your preferred IDE
  - [Install VS Code](https://code.visualstudio.com/)

### System Requirements

- **RAM**: Minimum 4GB available (8GB recommended)
- **Disk Space**: At least 5GB free
- **Network**: Internet connection for pulling Docker images

---

## Quick Start

**TL;DR** - Get up and running in under 2 minutes:

```bash
# 1. Clone the repository
git clone <repository-url>
cd giia-core-engine

# 2. One-command setup (recommended)
make setup-local

# Or manually:
# 2a. Start infrastructure
make run-local

# 2b. (Optional) Start development tools
make run-tools
```

That's it! Your local development environment is ready. 🚀

---

## Detailed Setup

### Step 1: Clone the Repository

```bash
git clone <repository-url>
cd giia-core-engine
```

### Step 2: Configure Environment Variables

The `.env.example` files are already included. You can use them as-is for local development, or customize them:

```bash
# Optional: Create root .env file
cp .env.example .env

# Optional: Create service-specific .env files
cp services/auth-service/.env.example services/auth-service/.env
# Repeat for other services...
```

**Note**: The default values in `.env.example` files are already configured to work with the Docker Compose setup.

### Step 3: Start Infrastructure Services

#### Option A: Using the Setup Script (Recommended)

```bash
./scripts/setup-local.sh
```

This script will:
- ✅ Check all prerequisites
- ✅ Create environment files if needed
- ✅ Pull Docker images
- ✅ Start all infrastructure services
- ✅ Wait for health checks to pass
- ✅ Display connection information

#### Option B: Using Make

```bash
make run-local
```

#### Option C: Using Docker Compose Directly

```bash
docker-compose up -d
```

### Step 4: Verify Services Are Running

```bash
# Check status
make status-local

# Or check manually
docker-compose ps
```

You should see three services running and healthy:
- ✅ `giia-postgres` (PostgreSQL 16)
- ✅ `giia-redis` (Redis 7)
- ✅ `giia-nats` (NATS Jetstream 2)

---

## Infrastructure Services

### PostgreSQL Database

The shared PostgreSQL instance uses a **multi-schema approach** where each microservice has its own schema within the `giia_dev` database.

**Connection Details:**
```
Host:     localhost
Port:     5432
User:     giia
Password: giia_dev_password
Database: giia_dev
```

**Schemas:**
- `auth` - Authentication & user management
- `catalog` - Product catalog
- `ddmrp` - DDMRP engine calculations
- `execution` - Order execution
- `analytics` - Analytics and reporting
- `ai_agent` - AI agent data and embeddings

**Connection Strings:**

```bash
# Generic
postgresql://giia:giia_dev_password@localhost:5432/giia_dev

# Service-specific (with schema)
postgresql://giia:giia_dev_password@localhost:5432/giia_dev?search_path=auth
postgresql://giia:giia_dev_password@localhost:5432/giia_dev?search_path=catalog
# ... and so on
```

### Redis Cache

**Connection Details:**
```
Host:     localhost
Port:     6379
Password: giia_redis_password
```

**Connection String:**
```bash
redis://:giia_redis_password@localhost:6379/0
```

**Database Assignment by Service:**
- `0` - Auth Service
- `1` - Catalog Service
- `2` - DDMRP Engine Service
- `3` - Execution Service
- `4` - Analytics Service
- `5` - AI Agent Service

### NATS Jetstream

**Connection Details:**
```
Client Port:     localhost:4222
Monitoring:      http://localhost:8222
```

**Connection String:**
```bash
nats://localhost:4222
```

**Monitoring Dashboard:**
Open http://localhost:8222 in your browser to view NATS statistics.

---

## Running Services

### Running a Single Service

Each service can be run independently while connected to the shared infrastructure.

#### Using Make (Recommended)

```bash
# Run auth service
make run-service SERVICE=auth

# Run catalog service
make run-service SERVICE=catalog

# Run any service
make run-service SERVICE=<service-name>
```

#### Using Go Directly

```bash
# Navigate to service directory
cd services/auth-service

# Run the service
go run cmd/api/main.go
```

#### Using Your IDE (VS Code)

1. Open the project in VS Code
2. Go to the service's `main.go` file (e.g., `services/auth-service/cmd/api/main.go`)
3. Press `F5` or click "Run and Debug"
4. Set breakpoints and debug as needed

**VS Code Launch Configuration:**

The project includes a `.vscode/launch.json` file with pre-configured debug targets for each service.

### Running Multiple Services

You can run multiple services simultaneously in different terminal windows/tabs:

```bash
# Terminal 1
make run-service SERVICE=auth

# Terminal 2
make run-service SERVICE=catalog

# Terminal 3
make run-service SERVICE=ddmrp
```

### Service Ports

| Service | HTTP Port | gRPC Port |
|---------|-----------|-----------|
| Auth Service | 8081 | 9081 |
| Catalog Service | 8082 | 9082 |
| DDMRP Engine Service | 8083 | 9083 |
| Execution Service | 8084 | 9084 |
| Analytics Service | 8085 | 9085 |
| AI Agent Service | 8086 | 9086 |

---

## Development Tools

Optional GUI tools for database and cache inspection.

### Starting Development Tools

```bash
make run-tools
```

This starts:
- **pgAdmin** - PostgreSQL GUI
- **Redis Commander** - Redis GUI

### pgAdmin (PostgreSQL GUI)

**Access:** http://localhost:5050

**Login Credentials:**
```
Email:    admin@giia.local
Password: admin
```

**Connecting to GIIA Database:**

1. Click "Add New Server"
2. General Tab:
   - Name: `GIIA Local`
3. Connection Tab:
   - Host: `giia-postgres` (Docker network name)
   - Port: `5432`
   - Database: `giia_dev`
   - Username: `giia`
   - Password: `giia_dev_password`
4. Click "Save"

### Redis Commander (Redis GUI)

**Access:** http://localhost:8081

No login required. You'll see all Redis databases and can inspect keys, values, and TTLs.

---

## Common Operations

### View Logs

```bash
# All services
make logs-local

# Specific service
docker-compose logs -f postgres
docker-compose logs -f redis
docker-compose logs -f nats
```

### Check Service Status

```bash
make status-local
```

### Restart Infrastructure

```bash
make restart-local
```

### Stop Infrastructure

```bash
make stop-local
```

### Clean Everything (Delete All Data)

```bash
make clean-local
```

⚠️ **Warning**: This will delete all data including databases, cache, and message queues. You'll start with a fresh environment.

### Load Sample Data

```bash
make seed-data
```

This loads test users, products, and other sample data for development and testing.

### Run Tests

```bash
# All tests
make test

# With coverage
make test-coverage

# Specific service
make test-auth
make test-catalog
```

### Build Services

```bash
# Build all services
make build

# Build specific service
make build-auth
make build-catalog
```

---

## Troubleshooting

### Issue: Docker daemon is not running

**Symptoms:**
```
Cannot connect to the Docker daemon at unix:///var/run/docker.sock
```

**Solution:**
1. Start Docker Desktop
2. Wait for Docker to fully initialize (whale icon should be steady)
3. Verify: `docker ps`

---

### Issue: Port already in use

**Symptoms:**
```
Error: Bind for 0.0.0.0:5432 failed: port is already allocated
```

**Solution:**

Check what's using the port:

```bash
# Windows
netstat -ano | findstr :5432

# macOS/Linux
lsof -i :5432
```

Options:
1. **Stop the conflicting service** (e.g., local PostgreSQL)
2. **Change the port** in `docker-compose.yml`:
   ```yaml
   ports:
     - "5433:5432"  # Map to a different host port
   ```
3. **Update service .env files** with the new port

---

### Issue: Services are slow or unresponsive

**Symptoms:**
- Health checks timeout
- Services take forever to start
- System is sluggish

**Solution:**

1. **Check system resources:**
   ```bash
   docker stats
   ```

2. **Verify sufficient disk space:**
   ```bash
   df -h  # macOS/Linux
   ```

3. **Increase Docker resources:**
   - Docker Desktop → Settings → Resources
   - Increase CPUs, Memory, and Disk

4. **Clean up Docker:**
   ```bash
   # Remove unused containers, images, networks
   docker system prune -a

   # Remove volumes (⚠️ deletes data)
   docker system prune -a --volumes
   ```

---

### Issue: Cannot connect to PostgreSQL from service

**Symptoms:**
```
dial tcp [::1]:5432: connect: connection refused
```

**Solution:**

1. **Check PostgreSQL is running:**
   ```bash
   docker exec giia-postgres pg_isready -U giia
   ```

2. **Verify connection string in service `.env`:**
   ```
   DB_HOST=localhost
   DB_PORT=5432
   DB_USER=giia
   DB_PASSWORD=giia_dev_password
   DB_NAME=giia_dev
   ```

3. **Check schema is specified:**
   ```
   DB_SCHEMA=auth  # or catalog, ddmrp, etc.
   ```

---

### Issue: Redis authentication failed

**Symptoms:**
```
NOAUTH Authentication required
```

**Solution:**

Ensure your service `.env` includes the Redis password:
```
REDIS_PASSWORD=giia_redis_password
```

Or use the connection string:
```
REDIS_URL=redis://:giia_redis_password@localhost:6379/0
```

---

### Issue: NATS connection refused

**Symptoms:**
```
dial tcp [::1]:4222: connect: connection refused
```

**Solution:**

1. **Check NATS is running:**
   ```bash
   curl http://localhost:8222/healthz
   ```

2. **Verify NATS URL in service `.env`:**
   ```
   NATS_URL=nats://localhost:4222
   ```

3. **Check NATS logs:**
   ```bash
   docker-compose logs nats
   ```

---

### Issue: Database schema not found

**Symptoms:**
```
ERROR: schema "auth" does not exist
```

**Solution:**

The initialization script should create all schemas automatically. If not:

```bash
# Restart PostgreSQL with clean state
docker-compose down -v
docker-compose up -d postgres

# Manually run init script
docker exec -i giia-postgres psql -U giia -d giia_dev < scripts/init-db.sql
```

---

### Issue: Permission denied on scripts

**Symptoms:**
```
bash: ./scripts/setup-local.sh: Permission denied
```

**Solution:**

```bash
# Make scripts executable
chmod +x scripts/*.sh

# Or run with bash explicitly
bash scripts/setup-local.sh
```

---

## FAQ

### Q: Do I need to run all services to develop on one?

**A:** No! You only need to run the infrastructure (PostgreSQL, Redis, NATS) and the specific service you're working on. Other services are optional.

---

### Q: Can I use my own PostgreSQL/Redis instead of Docker?

**A:** Yes, but you'll need to:
1. Create the schemas manually (see `scripts/init-db.sql`)
2. Update the `.env` files with your connection details
3. Ensure extensions like `uuid-ossp` and `pgcrypto` are installed

---

### Q: How do I reset my local database?

**A:**
```bash
make clean-local  # Deletes everything
make run-local    # Starts fresh
make seed-data    # (Optional) Load sample data
```

---

### Q: Can I run this on Windows?

**A:** Yes! The setup works on:
- ✅ Windows 10/11 with Docker Desktop
- ✅ Windows with WSL2
- ✅ Git Bash for running shell scripts

For PowerShell users, you can run `docker-compose` commands directly.

---

### Q: How do I update Docker images?

**A:**
```bash
docker-compose pull
docker-compose up -d
```

---

### Q: Can I run services in Docker instead of locally?

**A:** Yes, each service has a `Dockerfile`. You can add them to `docker-compose.yml` or build and run manually:

```bash
# Build service image
docker build -f services/auth-service/Dockerfile -t giia-auth .

# Run service container
docker run -p 8081:8081 --network giia-network giia-auth
```

---

### Q: How do I debug a service?

**A:**

**VS Code:**
1. Open service's `main.go`
2. Press `F5` to start debugging
3. Set breakpoints by clicking left of line numbers

**Delve (CLI):**
```bash
cd services/auth-service
dlv debug cmd/api/main.go
```

---

### Q: What if I need different environment variables per developer?

**A:** Create a `.env.local` file (not tracked by Git):

```bash
# Copy example
cp .env.example .env.local

# Modify .env.local with your settings
```

Then update the service to load `.env.local` if it exists.

---

### Q: How do I add a new microservice?

**A:**

1. Create service directory: `services/new-service/`
2. Add `.env.example` with connection details
3. Add schema to `scripts/init-db.sql`:
   ```sql
   CREATE SCHEMA IF NOT EXISTS new_service;
   GRANT USAGE ON SCHEMA new_service TO giia;
   ```
4. Assign a unique Redis DB number in `.env.example`
5. Choose unique HTTP and gRPC ports
6. Update `SERVICES` variable in `Makefile`

---

## Additional Resources

- [Main README](../README.md) - Project overview
- [CLAUDE.md](../CLAUDE.md) - Development guidelines
- [Architecture Documentation](../docs/architecture/) - System architecture
- [API Documentation](../docs/api/) - API references

---

## Need Help?

If you encounter issues not covered here:

1. **Check the logs**: `make logs-local`
2. **Verify service status**: `make status-local`
3. **Review Docker Compose setup**: `docker-compose.yml`
4. **Ask your team** - Someone may have encountered the same issue!

---

**Happy Coding! 🚀**
