#!/bin/bash
set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}===================================================${NC}"
echo -e "${BLUE}   Deploying Infrastructure Services${NC}"
echo -e "${BLUE}===================================================${NC}"
echo ""

# Create namespace and base resources
echo -e "${YELLOW}Creating namespace and shared resources...${NC}"
kubectl apply -f k8s/base/namespace.yaml
kubectl apply -f k8s/base/shared-configmap.yaml
kubectl apply -f k8s/base/shared-secrets.yaml
echo -e "${GREEN}✓ Namespace and shared resources created${NC}"
echo ""

# Add Helm repositories
echo -e "${YELLOW}Adding Helm repositories...${NC}"
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo add nats https://nats-io.github.io/k8s/helm/charts/
helm repo update
echo -e "${GREEN}✓ Helm repositories added${NC}"
echo ""

# Deploy PostgreSQL
echo -e "${YELLOW}Deploying PostgreSQL...${NC}"
helm upgrade --install postgresql bitnami/postgresql \
  --namespace giia-dev \
  --values k8s/infrastructure/postgresql/values-dev.yaml \
  --wait \
  --timeout 5m

echo -e "${GREEN}✓ PostgreSQL deployed${NC}"
echo ""

# Deploy Redis
echo -e "${YELLOW}Deploying Redis...${NC}"
helm upgrade --install redis bitnami/redis \
  --namespace giia-dev \
  --values k8s/infrastructure/redis/values-dev.yaml \
  --wait \
  --timeout 5m

echo -e "${GREEN}✓ Redis deployed${NC}"
echo ""

# Deploy NATS
echo -e "${YELLOW}Deploying NATS with JetStream...${NC}"
helm upgrade --install nats nats/nats \
  --namespace giia-dev \
  --values k8s/infrastructure/nats/values-dev.yaml \
  --wait \
  --timeout 5m

echo -e "${GREEN}✓ NATS deployed${NC}"
echo ""

echo -e "${BLUE}===================================================${NC}"
echo -e "${GREEN}Infrastructure deployment complete!${NC}"
echo -e "${BLUE}===================================================${NC}"
echo ""

# Show status
kubectl get pods -n giia-dev
echo ""
echo -e "${YELLOW}Waiting for all pods to be ready...${NC}"
kubectl wait --for=condition=ready pod --all -n giia-dev --timeout=300s || true

echo ""
echo -e "${GREEN}✓ All infrastructure services are running${NC}"
echo ""
echo -e "${YELLOW}Connection Information:${NC}"
echo -e "  PostgreSQL: ${GREEN}postgresql.giia-dev.svc.cluster.local:5432${NC}"
echo -e "  Redis:      ${GREEN}redis-master.giia-dev.svc.cluster.local:6379${NC}"
echo -e "  NATS:       ${GREEN}nats.giia-dev.svc.cluster.local:4222${NC}"
echo ""
