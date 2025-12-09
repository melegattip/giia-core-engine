#!/bin/bash
# =============================================================================
# GIIA Local Development Setup Script
# =============================================================================
# One-command setup for local development environment
# =============================================================================

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
BOLD='\033[1m'
NC='\033[0m'

# Script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo -e "${BLUE}${BOLD}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  GIIA Core Engine - Local Development Setup"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${NC}"

# Function to check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to print status
print_status() {
    echo -e "${BLUE}➜${NC} $1"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

# =============================================================================
# 1. Check Prerequisites
# =============================================================================

print_status "Checking prerequisites..."

# Check Docker
if ! command_exists docker; then
    print_error "Docker is not installed!"
    echo "Please install Docker from: https://docs.docker.com/get-docker/"
    exit 1
fi
print_success "Docker found: $(docker --version | head -n 1)"

# Check if Docker daemon is running
if ! docker ps >/dev/null 2>&1; then
    print_error "Docker daemon is not running!"
    echo "Please start Docker and try again."
    exit 1
fi
print_success "Docker daemon is running"

# Check Docker Compose
if ! command_exists docker-compose; then
    if ! docker compose version >/dev/null 2>&1; then
        print_error "Docker Compose is not installed!"
        exit 1
    else
        DOCKER_COMPOSE="docker compose"
    fi
else
    DOCKER_COMPOSE="docker-compose"
fi
print_success "Docker Compose found"

# Check Go
if command_exists go; then
    print_success "Go found: $(go version | awk '{print $3}')"
else
    print_warning "Go is not installed (required for running services locally)"
fi

# Check disk space (at least 5GB free)
if command_exists df; then
    AVAILABLE_SPACE=$(df -BG "$PROJECT_ROOT" | tail -1 | awk '{print $4}' | sed 's/G//')
    if [ "$AVAILABLE_SPACE" -lt 5 ]; then
        print_warning "Low disk space: ${AVAILABLE_SPACE}GB available (recommended: 5GB+)"
    else
        print_success "Sufficient disk space: ${AVAILABLE_SPACE}GB available"
    fi
fi

echo ""

# =============================================================================
# 2. Create Environment Files
# =============================================================================

print_status "Setting up environment files..."

# Create root .env if it doesn't exist
if [ ! -f "$PROJECT_ROOT/.env" ]; then
    if [ -f "$PROJECT_ROOT/.env.example" ]; then
        cp "$PROJECT_ROOT/.env.example" "$PROJECT_ROOT/.env"
        print_success "Created .env from .env.example"
    else
        print_warning ".env.example not found, skipping root .env creation"
    fi
else
    print_success ".env already exists"
fi

# Create service .env files
SERVICES=("auth-service" "catalog-service" "ddmrp-engine-service" "execution-service" "analytics-service" "ai-agent-service")
for service in "${SERVICES[@]}"; do
    SERVICE_DIR="$PROJECT_ROOT/services/$service"
    if [ -d "$SERVICE_DIR" ]; then
        if [ ! -f "$SERVICE_DIR/.env" ] && [ -f "$SERVICE_DIR/.env.example" ]; then
            cp "$SERVICE_DIR/.env.example" "$SERVICE_DIR/.env"
            print_success "Created .env for $service"
        fi
    fi
done

echo ""

# =============================================================================
# 3. Start Infrastructure Services
# =============================================================================

print_status "Starting infrastructure services..."

cd "$PROJECT_ROOT"

# Stop any existing containers
$DOCKER_COMPOSE down > /dev/null 2>&1 || true

# Pull latest images
print_status "Pulling Docker images..."
$DOCKER_COMPOSE pull

# Start services
print_status "Starting containers..."
$DOCKER_COMPOSE up -d

echo ""

# =============================================================================
# 4. Wait for Services to Be Ready
# =============================================================================

print_status "Waiting for services to be healthy..."

if [ -f "$SCRIPT_DIR/wait-for-services.sh" ]; then
    bash "$SCRIPT_DIR/wait-for-services.sh"
else
    print_warning "wait-for-services.sh not found, waiting 10 seconds..."
    sleep 10
fi

echo ""

# =============================================================================
# 5. Display Connection Information
# =============================================================================

echo -e "${GREEN}${BOLD}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  Setup Complete! 🎉"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${NC}"

echo -e "${BOLD}Infrastructure Services:${NC}"
echo "  PostgreSQL:  localhost:5432"
echo "    User:      giia"
echo "    Password:  giia_dev_password"
echo "    Database:  giia_dev"
echo ""
echo "  Redis:       localhost:6379"
echo "    Password:  giia_redis_password"
echo ""
echo "  NATS:        localhost:4222"
echo "    Monitoring: http://localhost:8222"
echo ""

echo -e "${BOLD}Optional Tools:${NC}"
echo "  pgAdmin:           http://localhost:5050"
echo "  Redis Commander:   http://localhost:8081"
echo "  (Start with: docker-compose --profile tools up -d)"
echo ""

echo -e "${BOLD}Next Steps:${NC}"
echo "  1. Run a service: cd services/auth-service && go run cmd/api/main.go"
echo "  2. Run tests:     make test"
echo "  3. View logs:     docker-compose logs -f"
echo "  4. Stop services: docker-compose down"
echo ""

echo -e "${BOLD}Documentation:${NC}"
echo "  See docs/LOCAL_DEVELOPMENT.md for detailed setup instructions"
echo ""