#!/bin/bash
set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}===================================================${NC}"
echo -e "${BLUE}   Building Docker Images for Kubernetes${NC}"
echo -e "${BLUE}===================================================${NC}"
echo ""

# Services to build
SERVICES=("auth-service" "catalog-service" "ddmrp-engine-service" "execution-service" "analytics-service" "ai-agent-service")
REGISTRY="ghcr.io/giia"

for service in "${SERVICES[@]}"; do
    echo -e "${YELLOW}Building ${service}...${NC}"

    if [ ! -f "services/${service}/Dockerfile" ]; then
        echo -e "${RED}Warning: Dockerfile not found for ${service}, skipping...${NC}"
        continue
    fi

    # Build from project root (includes go.work and shared packages)
    docker build \
        -f "services/${service}/Dockerfile" \
        -t "${REGISTRY}/${service}:latest" \
        . || {
        echo -e "${RED}Error building ${service}${NC}"
        exit 1
    }

    # Load image into Minikube
    echo -e "${YELLOW}Loading ${service} into Minikube...${NC}"
    minikube image load "${REGISTRY}/${service}:latest"

    echo -e "${GREEN}✓ ${service} built and loaded${NC}"
    echo ""
done

echo -e "${BLUE}===================================================${NC}"
echo -e "${GREEN}All Docker images built and loaded successfully!${NC}"
echo -e "${BLUE}===================================================${NC}"
echo ""

# Show loaded images
echo -e "${YELLOW}Images in Minikube:${NC}"
minikube image ls | grep giia
echo ""
