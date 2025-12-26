#!/bin/bash
#
# GIIA Integration Test Suite Runner
#
# This script starts the test environment, waits for services,
# runs integration tests, and cleans up.
#
# Usage:
#   ./run-tests.sh              # Run all tests
#   ./run-tests.sh -v           # Run with verbose output
#   ./run-tests.sh -run "Auth"  # Run specific test pattern
#

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Default values
VERBOSE=""
TEST_PATTERN=""
TIMEOUT="10m"
SKIP_SETUP=false
SKIP_TEARDOWN=false

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -v|--verbose)
            VERBOSE="-v"
            shift
            ;;
        -run)
            TEST_PATTERN="-run $2"
            shift 2
            ;;
        --timeout)
            TIMEOUT="$2"
            shift 2
            ;;
        --skip-setup)
            SKIP_SETUP=true
            shift
            ;;
        --skip-teardown)
            SKIP_TEARDOWN=true
            shift
            ;;
        -h|--help)
            echo "Usage: $0 [options]"
            echo ""
            echo "Options:"
            echo "  -v, --verbose       Run tests with verbose output"
            echo "  -run PATTERN        Run only tests matching pattern"
            echo "  --timeout DURATION  Set test timeout (default: 10m)"
            echo "  --skip-setup        Skip docker-compose up"
            echo "  --skip-teardown     Skip docker-compose down"
            echo "  -h, --help          Show this help message"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

echo -e "${GREEN}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║     GIIA Integration Test Suite                          ║${NC}"
echo -e "${GREEN}╚════════════════════════════════════════════════════════════╝${NC}"
echo ""

# Start test environment
if [ "$SKIP_SETUP" = false ]; then
    echo -e "${YELLOW}🚀 Starting test environment...${NC}"
    docker-compose -f docker-compose.yml up -d
    
    echo -e "${YELLOW}⏳ Waiting for services to be healthy...${NC}"
    
    # Wait for services with timeout
    MAX_WAIT=120
    WAITED=0
    
    services=("giia-test-postgres" "giia-test-redis" "giia-test-nats" "giia-test-auth-service" "giia-test-catalog-service")
    
    for service in "${services[@]}"; do
        echo -n "  Waiting for $service... "
        while [ $WAITED -lt $MAX_WAIT ]; do
            if docker inspect --format='{{.State.Health.Status}}' "$service" 2>/dev/null | grep -q "healthy"; then
                echo -e "${GREEN}✓${NC}"
                break
            fi
            sleep 2
            WAITED=$((WAITED + 2))
        done
        
        if [ $WAITED -ge $MAX_WAIT ]; then
            echo -e "${RED}✗ Timeout waiting for $service${NC}"
            docker-compose -f docker-compose.yml logs "$service"
            exit 1
        fi
    done
    
    echo -e "${GREEN}✅ All services are healthy!${NC}"
    echo ""
else
    echo -e "${YELLOW}⏭️  Skipping setup (--skip-setup)${NC}"
fi

# Run integration tests
echo -e "${YELLOW}🧪 Running integration tests...${NC}"
echo ""

# Build test command
TEST_CMD="go test $VERBOSE -timeout $TIMEOUT $TEST_PATTERN ./..."

echo "Command: $TEST_CMD"
echo ""

# Run tests and capture exit code
set +e
$TEST_CMD
TEST_EXIT_CODE=$?
set -e

echo ""

# Show results
if [ $TEST_EXIT_CODE -eq 0 ]; then
    echo -e "${GREEN}╔════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║     ✅ All tests passed!                                 ║${NC}"
    echo -e "${GREEN}╚════════════════════════════════════════════════════════════╝${NC}"
else
    echo -e "${RED}╔════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${RED}║     ❌ Some tests failed!                                 ║${NC}"
    echo -e "${RED}╚════════════════════════════════════════════════════════════╝${NC}"
fi

# Cleanup
if [ "$SKIP_TEARDOWN" = false ]; then
    echo ""
    echo -e "${YELLOW}🧹 Cleaning up test environment...${NC}"
    docker-compose -f docker-compose.yml down -v
    echo -e "${GREEN}✅ Cleanup complete!${NC}"
else
    echo -e "${YELLOW}⏭️  Skipping teardown (--skip-teardown)${NC}"
fi

exit $TEST_EXIT_CODE
