#!/bin/bash
# =============================================================================
# Wait for Services - Health Check Script
# =============================================================================
# This script polls infrastructure services until they're all healthy
# =============================================================================

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
MAX_WAIT=120  # Maximum wait time in seconds
CHECK_INTERVAL=2  # Check every 2 seconds

# Service endpoints
POSTGRES_HOST=${POSTGRES_HOST:-localhost}
POSTGRES_PORT=${POSTGRES_PORT:-5432}
REDIS_HOST=${REDIS_HOST:-localhost}
REDIS_PORT=${REDIS_PORT:-6379}
NATS_HOST=${NATS_HOST:-localhost}
NATS_MONITORING_PORT=${NATS_MONITORING_PORT:-8222}

echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}  Waiting for GIIA infrastructure services to be ready...${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

# Function to check if a service is ready
check_postgres() {
    docker exec giia-postgres pg_isready -U giia -h localhost > /dev/null 2>&1
}

check_redis() {
    docker exec giia-redis redis-cli -a giia_redis_password ping > /dev/null 2>&1
}

check_nats() {
    curl -s http://${NATS_HOST}:${NATS_MONITORING_PORT}/healthz > /dev/null 2>&1
}

# Wait for PostgreSQL
echo -ne "${YELLOW}⏳ PostgreSQL...${NC}"
elapsed=0
while ! check_postgres; do
    if [ $elapsed -ge $MAX_WAIT ]; then
        echo -e "\r${RED}✗ PostgreSQL - TIMEOUT${NC}"
        exit 1
    fi
    sleep $CHECK_INTERVAL
    elapsed=$((elapsed + CHECK_INTERVAL))
done
echo -e "\r${GREEN}✓ PostgreSQL is ready${NC}"

# Wait for Redis
echo -ne "${YELLOW}⏳ Redis...${NC}"
elapsed=0
while ! check_redis; do
    if [ $elapsed -ge $MAX_WAIT ]; then
        echo -e "\r${RED}✗ Redis - TIMEOUT${NC}"
        exit 1
    fi
    sleep $CHECK_INTERVAL
    elapsed=$((elapsed + CHECK_INTERVAL))
done
echo -e "\r${GREEN}✓ Redis is ready${NC}"

# Wait for NATS
echo -ne "${YELLOW}⏳ NATS Jetstream...${NC}"
elapsed=0
while ! check_nats; do
    if [ $elapsed -ge $MAX_WAIT ]; then
        echo -e "\r${RED}✗ NATS - TIMEOUT${NC}"
        exit 1
    fi
    sleep $CHECK_INTERVAL
    elapsed=$((elapsed + CHECK_INTERVAL))
done
echo -e "\r${GREEN}✓ NATS Jetstream is ready${NC}"

echo ""
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}  All services are ready! 🚀${NC}"
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo -e "${BLUE}Connection Information:${NC}"
echo -e "  PostgreSQL: ${POSTGRES_HOST}:${POSTGRES_PORT} (user: giia, db: giia_dev)"
echo -e "  Redis:      ${REDIS_HOST}:${REDIS_PORT} (password: giia_redis_password)"
echo -e "  NATS:       ${NATS_HOST}:4222 (monitoring: ${NATS_HOST}:${NATS_MONITORING_PORT})"
echo ""