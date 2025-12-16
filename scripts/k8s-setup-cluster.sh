#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}===================================================${NC}"
echo -e "${BLUE}   GIIA Kubernetes Development Cluster Setup${NC}"
echo -e "${BLUE}===================================================${NC}"
echo ""

# Check prerequisites
echo -e "${YELLOW}Checking prerequisites...${NC}"

if ! command -v kubectl &> /dev/null; then
    echo -e "${RED}Error: kubectl is not installed${NC}"
    echo -e "${YELLOW}Install kubectl: https://kubernetes.io/docs/tasks/tools/${NC}"
    exit 1
fi
echo -e "${GREEN}✓ kubectl installed${NC}"

if ! command -v helm &> /dev/null; then
    echo -e "${RED}Error: Helm is not installed${NC}"
    echo -e "${YELLOW}Install Helm: https://helm.sh/docs/intro/install/${NC}"
    echo -e "${YELLOW}Windows: choco install kubernetes-helm${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Helm installed${NC}"

if ! command -v minikube &> /dev/null; then
    echo -e "${RED}Error: Minikube is not installed${NC}"
    echo -e "${YELLOW}Install Minikube: https://minikube.sigs.k8s.io/docs/start/${NC}"
    echo -e "${YELLOW}Windows: choco install minikube${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Minikube installed${NC}"

if ! command -v docker &> /dev/null; then
    echo -e "${RED}Error: Docker is not installed${NC}"
    echo -e "${YELLOW}Install Docker: https://docs.docker.com/get-docker/${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Docker installed${NC}"

echo ""
echo -e "${BLUE}Creating Minikube cluster...${NC}"
minikube start \
  --cpus=4 \
  --memory=8192 \
  --disk-size=20g \
  --driver=docker \
  --kubernetes-version=v1.28.0

echo ""
echo -e "${BLUE}Enabling addons...${NC}"
minikube addons enable ingress
minikube addons enable metrics-server

echo ""
echo -e "${GREEN}✓ Cluster created successfully!${NC}"
echo ""

kubectl cluster-info
kubectl get nodes

echo ""
echo -e "${BLUE}===================================================${NC}"
echo -e "${GREEN}Kubernetes cluster is ready!${NC}"
echo -e "${BLUE}===================================================${NC}"
echo ""
echo -e "${YELLOW}Next steps:${NC}"
echo -e "  1. Deploy infrastructure: ${GREEN}make k8s-deploy-infra${NC}"
echo -e "  2. Deploy services: ${GREEN}make k8s-deploy-services${NC}"
echo -e "  3. Check status: ${GREEN}make k8s-status${NC}"
echo ""
