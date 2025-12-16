#!/bin/bash
set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}===================================================${NC}"
echo -e "${BLUE}   Tearing Down Kubernetes Cluster${NC}"
echo -e "${BLUE}===================================================${NC}"
echo ""

echo -e "${YELLOW}⚠️  WARNING: This will destroy the local Kubernetes cluster and all data!${NC}"
echo -e "${YELLOW}Press Ctrl+C to cancel, or wait 5 seconds to continue...${NC}"
sleep 5

echo ""
echo -e "${YELLOW}Deleting all Helm releases in giia-dev namespace...${NC}"
helm list -n giia-dev --short | xargs -I {} helm uninstall {} -n giia-dev || true
echo -e "${GREEN}✓ Helm releases deleted${NC}"

echo ""
echo -e "${YELLOW}Deleting namespace...${NC}"
kubectl delete namespace giia-dev --ignore-not-found=true
echo -e "${GREEN}✓ Namespace deleted${NC}"

echo ""
echo -e "${YELLOW}Stopping Minikube...${NC}"
minikube stop
echo -e "${GREEN}✓ Minikube stopped${NC}"

echo ""
echo -e "${YELLOW}Deleting Minikube cluster...${NC}"
minikube delete
echo -e "${GREEN}✓ Minikube cluster deleted${NC}"

echo ""
echo -e "${BLUE}===================================================${NC}"
echo -e "${GREEN}Cluster destroyed successfully!${NC}"
echo -e "${BLUE}===================================================${NC}"
echo ""
echo -e "${YELLOW}To create a new cluster, run:${NC} ${GREEN}make k8s-setup${NC}"
echo ""
