# Feature Specification: Kubernetes Development Cluster

**Created**: 2025-12-09

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Local Kubernetes Development Environment (Priority: P1)

As a backend developer, I need a local Kubernetes cluster (Minikube or kind) so that I can test service deployments and configurations before pushing to staging.

**Why this priority**: Critical for development workflow. Enables testing Kubernetes-specific features (ingress, services, config maps) locally. Prevents "works on my machine" issues.

**Independent Test**: Can be fully tested by running cluster setup script, deploying a service, and verifying it's accessible via kubectl and browser. Delivers standalone value: local Kubernetes environment ready for testing.

**Acceptance Scenarios**:

1. **Scenario**: Cluster initialization
   - **Given** developer has Docker and kubectl installed
   - **When** developer runs `make k8s-dev-setup`
   - **Then** local Kubernetes cluster starts with all addons (ingress, metrics-server)

2. **Scenario**: Deploy service to local cluster
   - **Given** local cluster is running
   - **When** developer runs `make k8s-dev-deploy-auth`
   - **Then** Auth service pod starts and passes readiness checks

3. **Scenario**: Service accessibility
   - **Given** Auth service is deployed
   - **When** developer accesses http://auth.giia.local
   - **Then** service responds to HTTP requests

---

### User Story 2 - Infrastructure Services in Kubernetes (Priority: P1)

As a DevOps engineer, I need PostgreSQL, Redis, and NATS deployed in Kubernetes so that services can connect to dependencies in a production-like environment.

**Why this priority**: Critical for realistic testing. Services must connect to infrastructure services same way as production. Required before deploying application services.

**Independent Test**: Can be tested by deploying infrastructure Helm charts, verifying all pods are running, and connecting to services from application pods.

**Acceptance Scenarios**:

1. **Scenario**: Deploy PostgreSQL with Helm
   - **Given** Kubernetes cluster is running
   - **When** DevOps engineer deploys PostgreSQL Helm chart
   - **Then** PostgreSQL pod starts with persistent volume and is accessible

2. **Scenario**: Deploy Redis with Helm
   - **Given** Kubernetes cluster is running
   - **When** DevOps engineer deploys Redis Helm chart
   - **Then** Redis pod starts and services can connect on port 6379

3. **Scenario**: Deploy NATS Jetstream
   - **Given** Kubernetes cluster is running
   - **When** DevOps engineer deploys NATS Helm chart with Jetstream enabled
   - **Then** NATS pod starts with persistent streams configured

---

### User Story 3 - Service Deployment with Helm (Priority: P2)

As a DevOps engineer, I need Helm charts for all microservices so that deployments are consistent, versioned, and easy to manage across environments.

**Why this priority**: Important for deployment consistency and automation. Can deploy with kubectl initially but Helm provides better configuration management.

**Independent Test**: Can be tested by installing service Helm chart with custom values, upgrading to new version, and verifying rolling update works correctly.

**Acceptance Scenarios**:

1. **Scenario**: Deploy service with Helm
   - **Given** Helm chart exists for Auth service
   - **When** DevOps engineer runs `helm install auth-service ./charts/auth-service`
   - **Then** service deploys with all resources (deployment, service, ingress, configmap, secret)

2. **Scenario**: Upgrade service version
   - **Given** Auth service v1.0.0 is deployed
   - **When** DevOps engineer runs `helm upgrade auth-service ./charts/auth-service --set image.tag=v1.1.0`
   - **Then** service performs rolling update to new version

3. **Scenario**: Environment-specific configuration
   - **Given** Helm chart has values files for dev, staging, prod
   - **When** chart is installed with `--values values-dev.yaml`
   - **Then** service uses dev-specific configuration (replicas, resources, env vars)

---

### User Story 4 - Observability Stack (Priority: P3)

As a DevOps engineer, I need Prometheus and Grafana deployed in the cluster so that I can monitor service metrics and set up alerts.

**Why this priority**: Important for production but not blocking for initial development. Can view logs with kubectl initially. Critical before production deployment.

**Independent Test**: Can be tested by deploying Prometheus and Grafana, verifying they scrape metrics from services, and accessing Grafana dashboards.

**Acceptance Scenarios**:

1. **Scenario**: Deploy Prometheus
   - **Given** Kubernetes cluster is running
   - **When** DevOps engineer deploys Prometheus Helm chart
   - **Then** Prometheus scrapes metrics from all service pods

2. **Scenario**: Deploy Grafana with dashboards
   - **Given** Prometheus is collecting metrics
   - **When** DevOps engineer deploys Grafana with pre-configured dashboards
   - **Then** Grafana displays service metrics and health status

3. **Scenario**: Service discovery
   - **Given** new service is deployed with `/metrics` endpoint
   - **When** Prometheus performs service discovery
   - **Then** Prometheus automatically starts scraping new service

---

### Edge Cases

- What happens when Kubernetes cluster runs out of resources (CPU/memory)?
- How to handle persistent volume failures (database data loss)?
- What happens when service deployment fails (rollback strategy)?
- How to test disaster recovery (cluster failure, backup restore)?
- What happens when ingress controller is unavailable?
- How to handle secrets management (avoid committing secrets)?
- What happens when service cannot connect to dependency (circuit breaker)?
- How to handle cluster upgrades (Kubernetes version updates)?

## Requirements *(mandatory)*

### Functional Requirements

#### Cluster Setup
- **FR-001**: System MUST support local Kubernetes cluster using Minikube or kind
- **FR-002**: System MUST configure cluster with minimum 4 CPU cores and 8GB RAM
- **FR-003**: System MUST install NGINX Ingress Controller for external access
- **FR-004**: System MUST install metrics-server for resource monitoring
- **FR-005**: System MUST configure local DNS (*.giia.local) or use /etc/hosts entries
- **FR-006**: System MUST provide Makefile commands for cluster lifecycle (setup, start, stop, destroy)

#### Infrastructure Services
- **FR-007**: System MUST deploy PostgreSQL 16 using Bitnami Helm chart with persistence
- **FR-008**: System MUST deploy Redis 7 using Bitnami Helm chart with password authentication
- **FR-009**: System MUST deploy NATS 2 with Jetstream enabled using official Helm chart
- **FR-010**: System MUST configure persistent volumes for database data (10GB PostgreSQL, 1GB Redis, 1GB NATS)
- **FR-011**: System MUST create Kubernetes secrets for database credentials

#### Service Deployment
- **FR-012**: System MUST provide Helm charts for all 6 microservices
- **FR-013**: Helm charts MUST support configurable replicas, resources, environment variables
- **FR-014**: Helm charts MUST include Kubernetes resources: Deployment, Service, Ingress, ConfigMap, Secret
- **FR-015**: Deployments MUST implement readiness and liveness probes
- **FR-016**: Deployments MUST use rolling update strategy with max unavailable = 1
- **FR-017**: Services MUST be exposed via ClusterIP (internal) and Ingress (external)

#### Configuration Management
- **FR-018**: System MUST use ConfigMaps for non-sensitive configuration
- **FR-019**: System MUST use Secrets for sensitive data (passwords, tokens, API keys)
- **FR-020**: System MUST support environment-specific values files (values-dev.yaml, values-staging.yaml, values-prod.yaml)
- **FR-021**: System MUST inject shared configuration from namespace-level ConfigMap

#### Observability (Optional)
- **FR-022**: System SHOULD deploy Prometheus for metrics collection
- **FR-023**: System SHOULD deploy Grafana for metrics visualization
- **FR-024**: System SHOULD configure Prometheus to scrape /metrics endpoints from all services
- **FR-025**: System SHOULD provide pre-configured Grafana dashboards for service metrics

### Key Entities

- **Cluster**: Local Kubernetes cluster (Minikube or kind)
- **Namespace**: Logical grouping of resources (giia-dev, giia-staging, giia-prod)
- **Helm Chart**: Package containing Kubernetes manifests and configuration templates
- **Deployment**: Kubernetes resource managing pod replicas and updates
- **Service**: Kubernetes resource providing stable network endpoint for pods
- **Ingress**: Kubernetes resource routing external HTTP traffic to services
- **ConfigMap**: Non-sensitive configuration data
- **Secret**: Sensitive configuration data (encrypted at rest)
- **PersistentVolume**: Storage for stateful services (PostgreSQL, Redis, NATS)

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Local cluster initializes in under 5 minutes from fresh state
- **SC-002**: All infrastructure services (PostgreSQL, Redis, NATS) deploy successfully and pass health checks
- **SC-003**: All 6 microservices deploy successfully to cluster and pass readiness probes
- **SC-004**: Services can discover and connect to infrastructure dependencies via Kubernetes DNS
- **SC-005**: Ingress routing works correctly for all services (*.giia.local domain)
- **SC-006**: Cluster can run all services simultaneously without resource exhaustion
- **SC-007**: Helm charts successfully install, upgrade, and rollback without errors
- **SC-008**: Developer can tear down and recreate cluster in under 10 minutes
