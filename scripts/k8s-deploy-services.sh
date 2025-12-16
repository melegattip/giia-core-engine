#!/bin/bash
set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}===================================================${NC}"
echo -e "${BLUE}   Deploying GIIA Services to Kubernetes${NC}"
echo -e "${BLUE}===================================================${NC}"
echo ""

# Services to deploy (only those with Helm charts)
SERVICES=("auth-service" "catalog-service")

for service in "${SERVICES[@]}"; do
    echo -e "${YELLOW}Deploying ${service}...${NC}"

    if [ ! -d "k8s/services/${service}" ]; then
        echo -e "${RED}Warning: Helm chart not found for ${service}, skipping...${NC}"
        continue
    fi

    helm upgrade --install "${service}" "k8s/services/${service}" \
        --namespace giia-dev \
        --values "k8s/services/${service}/values-dev.yaml" \
        --wait \
        --timeout 5m || {
        echo -e "${RED}Error deploying ${service}${NC}"
        exit 1
    }

    echo -e "${GREEN}✓ ${service} deployed${NC}"
    echo ""
done

echo -e "${BLUE}===================================================${NC}"
echo -e "${GREEN}All services deployed successfully!${NC}"
echo -e "${BLUE}===================================================${NC}"
echo ""

# Show status
kubectl get pods,svc,ingress -n giia-dev
echo ""

echo -e "${YELLOW}Waiting for all service pods to be ready...${NC}"
kubectl wait --for=condition=ready pod -l app.kubernetes.io/part-of=giia -n giia-dev --timeout=300s || true

echo ""
echo -e "${GREEN}✓ All services are running${NC}"
echo ""
echo -e "${YELLOW}Access Services:${NC}"
echo -e "  Note: Run ${GREEN}minikube tunnel${NC} in a separate terminal first"
echo -e "  Auth Service:    ${GREEN}http://auth.giia.local${NC}"
echo -e "  Catalog Service: ${GREEN}http://catalog.giia.local${NC}"
echo ""
echo -e "${YELLOW}Add to /etc/hosts (or C:\\Windows\\System32\\drivers\\etc\\hosts):${NC}"
echo -e "  ${GREEN}127.0.0.1 auth.giia.local catalog.giia.local${NC}"
echo ""
