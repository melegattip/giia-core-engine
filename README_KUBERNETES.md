# Kubernetes Development Cluster - GIIA Platform

Complete guide for deploying the GIIA platform to a local Kubernetes cluster using Minikube and Helm.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [Architecture Overview](#architecture-overview)
- [Detailed Setup Guide](#detailed-setup-guide)
- [Service Access](#service-access)
- [Common Operations](#common-operations)
- [Troubleshooting](#troubleshooting)
- [Cleanup](#cleanup)

---

## Prerequisites

### Required Tools

Install these tools before proceeding:

#### 1. kubectl (Kubernetes CLI)
```bash
# Windows
choco install kubernetes-cli

# macOS
brew install kubectl

# Linux
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl
```

#### 2. Helm (Package Manager)
```bash
# Windows
choco install kubernetes-helm

# macOS
brew install helm

# Linux
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
```

#### 3. Minikube (Local Cluster)
```bash
# Windows
choco install minikube

# macOS
brew install minikube

# Linux
curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64
sudo install minikube-linux-amd64 /usr/local/bin/minikube
```

#### 4. Docker
- Download from: https://docs.docker.com/get-docker/
- Minikube uses Docker as the container runtime

### System Requirements

- **CPU**: 4+ cores
- **RAM**: 8GB+ available
- **Disk**: 20GB+ free space
- **OS**: Windows 10/11, macOS 11+, Linux (Ubuntu 20.04+)

---

## Quick Start

Get the entire platform running in 3 commands:

```bash
# 1. Setup Kubernetes cluster
make k8s-setup

# 2. Deploy infrastructure (PostgreSQL, Redis, NATS)
make k8s-deploy-infra

# 3. Build and deploy all services
make k8s-build-images
make k8s-deploy-services
```

Or use the all-in-one command:
```bash
make k8s-full-deploy
```

### Access Services

In a **separate terminal**, start the Minikube tunnel:
```bash
make k8s-tunnel
```

Add to `/etc/hosts` (Windows: `C:\Windows\System32\drivers\etc\hosts`):
```
127.0.0.1 auth.giia.local catalog.giia.local
```

Access services:
- **Auth Service**: http://auth.giia.local
- **Catalog Service**: http://catalog.giia.local

---

## Architecture Overview

### Cluster Components

```
giia-dev Namespace
├── Infrastructure Services
│   ├── PostgreSQL (port 5432)
│   ├── Redis (port 6379)
│   └── NATS JetStream (port 4222)
│
└── Application Services
    ├── auth-service (HTTP: 8083, gRPC: 9091)
    └── catalog-service (HTTP: 8082)
```

### Ingress Routing

```
Internet → Minikube Tunnel → NGINX Ingress Controller
  ↓
  ├→ auth.giia.local → auth-service:8083
  └→ catalog.giia.local → catalog-service:8082
```

### Service Discovery

Services communicate using Kubernetes DNS:
- PostgreSQL: `postgresql.giia-dev.svc.cluster.local:5432`
- Redis: `redis-master.giia-dev.svc.cluster.local:6379`
- NATS: `nats.giia-dev.svc.cluster.local:4222`

---

## Detailed Setup Guide

### Step 1: Create Kubernetes Cluster

```bash
make k8s-setup
```

This command:
- Creates a Minikube cluster with 4 CPU cores and 8GB RAM
- Enables NGINX Ingress Controller
- Enables metrics-server for resource monitoring

**Verification:**
```bash
kubectl get nodes
kubectl cluster-info
```

### Step 2: Deploy Infrastructure

```bash
make k8s-deploy-infra
```

This deploys:
- **PostgreSQL 16** with 10GB persistent volume
- **Redis 7** with authentication
- **NATS JetStream** with 1GB file storage

**Verification:**
```bash
make k8s-status
```

Expected output:
```
NAME                              READY   STATUS    RESTARTS   AGE
pod/postgresql-0                  1/1     Running   0          2m
pod/redis-master-0                1/1     Running   0          2m
pod/nats-0                        1/1     Running   0          2m
```

### Step 3: Build Docker Images

```bash
make k8s-build-images
```

This:
- Builds Docker images for all services from Dockerfiles
- Loads images into Minikube (no registry needed)

**Verification:**
```bash
minikube image ls | grep giia
```

### Step 4: Deploy Services

```bash
make k8s-deploy-services
```

This:
- Deploys auth-service using Helm
- Deploys catalog-service using Helm
- Creates Ingress routes for external access

**Verification:**
```bash
make k8s-pods
```

Expected output:
```
NAME                                   READY   STATUS    RESTARTS   AGE
auth-service-xxxxxxxxx-xxxxx           1/1     Running   0          1m
catalog-service-xxxxxxxxx-xxxxx        1/1     Running   0          1m
```

---

## Service Access

### Enable Ingress Access

Minikube requires a tunnel for ingress to work:

```bash
# Terminal 1 - Keep this running
make k8s-tunnel
```

### Configure DNS

Add these entries to your hosts file:

**Windows**: `C:\Windows\System32\drivers\etc\hosts`
**macOS/Linux**: `/etc/hosts`

```
127.0.0.1 auth.giia.local catalog.giia.local
```

### Test Services

```bash
# Auth Service health check
curl http://auth.giia.local/health

# Catalog Service health check
curl http://catalog.giia.local/health
```

---

## Common Operations

### View Cluster Status

```bash
make k8s-status
```

### View Service Logs

```bash
# Auth service logs
make k8s-logs SERVICE=auth-service

# Catalog service logs
make k8s-logs SERVICE=catalog-service
```

### Restart a Service

```bash
make k8s-restart SERVICE=auth-service
```

### Access Service Shell

```bash
make k8s-shell SERVICE=auth-service
```

### Deploy Individual Services

```bash
# Deploy only auth-service
make k8s-deploy-auth

# Deploy only catalog-service
make k8s-deploy-catalog
```

### Rebuild and Redeploy

```bash
# Rebuild images after code changes
make k8s-build-images

# Redeploy services
make k8s-deploy-services
```

### Open Kubernetes Dashboard

```bash
make k8s-dashboard
```

---

## Troubleshooting

### Pods Not Starting

**Check pod status:**
```bash
kubectl get pods -n giia-dev
```

**Describe problematic pod:**
```bash
make k8s-describe SERVICE=auth-service
```

**View pod logs:**
```bash
make k8s-logs SERVICE=auth-service
```

### Service Not Accessible

**Verify Minikube tunnel is running:**
```bash
make k8s-tunnel
```

**Check ingress configuration:**
```bash
kubectl get ingress -n giia-dev
```

**Verify /etc/hosts entry:**
```bash
# Should return 127.0.0.1
ping auth.giia.local
```

### Image Pull Errors

**If you see `ImagePullBackOff`:**

The images must be loaded into Minikube:
```bash
make k8s-build-images
```

**Verify images are loaded:**
```bash
minikube image ls | grep giia
```

### Database Connection Issues

**Test PostgreSQL connectivity:**
```bash
kubectl run postgresql-client --rm --tty -i --restart='Never' \
  --namespace giia-dev \
  --image bitnami/postgresql:16 \
  --env="PGPASSWORD=giia_dev_password" \
  --command -- psql --host postgresql.giia-dev.svc.cluster.local -U giia -d giia_dev -c "SELECT 1"
```

**Test Redis connectivity:**
```bash
kubectl run redis-client --rm --tty -i --restart='Never' \
  --namespace giia-dev \
  --image bitnami/redis:7 \
  --env="REDISCLI_AUTH=giia_redis_password" \
  --command -- redis-cli -h redis-master.giia-dev.svc.cluster.local ping
```

### Resource Constraints

**If pods are pending due to insufficient resources:**

```bash
# Increase Minikube resources
minikube stop
minikube delete
minikube start --cpus=6 --memory=12288
```

### Port Conflicts

**If Minikube fails to start due to port conflicts:**

```bash
# Check what's using the port
# Windows
netstat -ano | findstr :8443

# macOS/Linux
lsof -i :8443

# Delete existing cluster and recreate
make k8s-teardown
make k8s-setup
```

---

## Cleanup

### Delete Services Only

Keep infrastructure and cluster running:
```bash
make k8s-clean
```

### Full Teardown

Delete everything (cluster, data, volumes):
```bash
make k8s-teardown
```

This will:
- Delete all Helm releases
- Delete giia-dev namespace
- Stop Minikube
- Delete Minikube cluster

---

## Directory Structure

```
k8s/
├── base/
│   ├── namespace.yaml              # giia-dev namespace
│   ├── shared-configmap.yaml       # Shared configuration
│   └── shared-secrets.yaml         # Shared secrets
│
├── infrastructure/
│   ├── postgresql/
│   │   └── values-dev.yaml        # PostgreSQL Helm values
│   ├── redis/
│   │   └── values-dev.yaml        # Redis Helm values
│   └── nats/
│       └── values-dev.yaml        # NATS Helm values
│
└── services/
    ├── auth-service/
    │   ├── Chart.yaml             # Helm chart metadata
    │   ├── values.yaml            # Default values
    │   ├── values-dev.yaml        # Dev overrides
    │   └── templates/
    │       ├── deployment.yaml
    │       ├── service.yaml
    │       ├── ingress.yaml
    │       ├── configmap.yaml
    │       └── serviceaccount.yaml
    │
    └── catalog-service/
        └── ...                    # Same structure
```

---

## Configuration Management

### Environment Variables

Services receive configuration from three sources:

1. **Shared ConfigMap** (`k8s/base/shared-configmap.yaml`)
   - Non-sensitive configuration
   - Database host, Redis URL, NATS URL
   - Environment type (development, staging, production)

2. **Shared Secrets** (`k8s/base/shared-secrets.yaml`)
   - Sensitive data (passwords, tokens)
   - Database password, Redis password, JWT secret

3. **Service-specific values** (`k8s/services/*/values.yaml`)
   - Service-specific environment variables
   - Resource limits, replica count, ports

### Modifying Configuration

**Change infrastructure credentials:**
```bash
# Edit secrets
kubectl edit secret shared-secrets -n giia-dev

# Or update the file and reapply
kubectl apply -f k8s/base/shared-secrets.yaml
```

**Change service configuration:**
```bash
# Edit Helm values
vim k8s/services/auth-service/values-dev.yaml

# Redeploy with new values
make k8s-deploy-auth
```

---

## Next Steps

1. **Add Remaining Services**: Create Helm charts for:
   - ddmrp-engine-service
   - execution-service
   - analytics-service
   - ai-agent-service

2. **Add Observability**: Deploy Prometheus and Grafana
   ```bash
   # Coming soon
   make k8s-deploy-observability
   ```

3. **Production Configuration**: Create values-prod.yaml files with:
   - Higher resource limits
   - Multiple replicas
   - Production secrets
   - Horizontal Pod Autoscaling

4. **CI/CD Integration**: Automate deployments with GitHub Actions

---

## Additional Resources

- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [Helm Documentation](https://helm.sh/docs/)
- [Minikube Documentation](https://minikube.sigs.k8s.io/docs/)
- [GIIA Project README](./README.md)

---

## Support

For questions or issues:
- Check [Troubleshooting](#troubleshooting) section
- View service logs: `make k8s-logs SERVICE=<name>`
- Check cluster status: `make k8s-status`
- Open an issue in the project repository
