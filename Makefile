.PHONY: help setup build test clean lint proto docker-build docker-push run-local

# Variables
GO := go
DOCKER := docker
KUBECTL := kubectl
SERVICES := auth-service
ARCHIVED_SERVICES := catalog-service ddmrp-engine-service execution-service analytics-service ai-agent-service
DOCKER_REGISTRY := ghcr.io/giia

# Colors for output
COLOR_RESET := \033[0m
COLOR_BOLD := \033[1m
COLOR_GREEN := \033[32m
COLOR_YELLOW := \033[33m
COLOR_BLUE := \033[34m

##@ General

help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\n$(COLOR_BOLD)Usage:$(COLOR_RESET)\n  make $(COLOR_BLUE)<target>$(COLOR_RESET)\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  $(COLOR_BLUE)%-20s$(COLOR_RESET) %s\n", $$1, $$2 } /^##@/ { printf "\n$(COLOR_BOLD)%s$(COLOR_RESET)\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

setup: ## Setup development environment
	@echo "$(COLOR_GREEN)Setting up development environment...$(COLOR_RESET)"
	$(GO) install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	$(GO) install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GO) work sync
	@echo "$(COLOR_GREEN)Setup complete!$(COLOR_RESET)"

##@ Build

build: ## Build all services
	@echo "$(COLOR_GREEN)Building all services...$(COLOR_RESET)"
	@for service in $(SERVICES); do \
		echo "$(COLOR_BLUE)Building $$service...$(COLOR_RESET)"; \
		cd services/$$service && $(GO) build -o ../../bin/$$service ./cmd/server/ || exit 1; \
		cd ../..; \
	done
	@echo "$(COLOR_GREEN)Build complete!$(COLOR_RESET)"

build-auth: ## Build auth service
	@echo "$(COLOR_BLUE)Building auth-service...$(COLOR_RESET)"
	cd services/auth-service && $(GO) build -o ../../bin/auth-service ./cmd/api/

# Archived service build targets removed - see archive/ directory
# Services consolidated to monolithic architecture (ADR 001)

##@ Testing

test: ## Run all tests
	@echo "$(COLOR_GREEN)Running all tests...$(COLOR_RESET)"
	$(GO) test -v -race -count=1 ./...

test-coverage: ## Run tests with coverage
	@echo "$(COLOR_GREEN)Running tests with coverage...$(COLOR_RESET)"
	$(GO) test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "$(COLOR_GREEN)Coverage report generated: coverage.html$(COLOR_RESET)"

test-auth: ## Run auth service tests
	@echo "$(COLOR_BLUE)Testing auth-service...$(COLOR_RESET)"
	cd services/auth-service && $(GO) test -v -race -count=1 ./...

# Archived service test targets removed - see archive/ directory

##@ Code Quality

lint: ## Run linters on all code
	@echo "$(COLOR_GREEN)Running linters...$(COLOR_RESET)"
	golangci-lint run --timeout=5m ./...

lint-fix: ## Fix linting issues automatically
	@echo "$(COLOR_GREEN)Fixing linting issues...$(COLOR_RESET)"
	golangci-lint run --fix --timeout=5m ./...

fmt: ## Format all Go code
	@echo "$(COLOR_GREEN)Formatting code...$(COLOR_RESET)"
	$(GO) fmt ./...
	gofmt -s -w .

vet: ## Run go vet
	@echo "$(COLOR_GREEN)Running go vet...$(COLOR_RESET)"
	$(GO) vet ./...

##@ Protocol Buffers

proto: ## Generate protobuf code
	@echo "$(COLOR_GREEN)Generating protobuf code...$(COLOR_RESET)"
	@for service in auth catalog ddmrp execution analytics ai; do \
		echo "$(COLOR_BLUE)Generating protos for $$service...$(COLOR_RESET)"; \
		protoc --go_out=. --go_opt=paths=source_relative \
			--go-grpc_out=. --go-grpc_opt=paths=source_relative \
			api/proto/$$service/v1/*.proto || true; \
	done
	@echo "$(COLOR_GREEN)Protobuf generation complete!$(COLOR_RESET)"

proto-clean: ## Clean generated protobuf files
	@echo "$(COLOR_YELLOW)Cleaning generated protobuf files...$(COLOR_RESET)"
	find api/proto -name "*.pb.go" -type f -delete
	find api/proto -name "*_grpc.pb.go" -type f -delete

##@ Docker

docker-build: ## Build all Docker images
	@echo "$(COLOR_GREEN)Building Docker images...$(COLOR_RESET)"
	@for service in $(SERVICES); do \
		echo "$(COLOR_BLUE)Building Docker image for $$service...$(COLOR_RESET)"; \
		$(DOCKER) build -f services/$$service/Dockerfile -t $(DOCKER_REGISTRY)/$$service:latest . || exit 1; \
	done

docker-build-auth: ## Build auth service Docker image
	@echo "$(COLOR_BLUE)Building auth-service Docker image...$(COLOR_RESET)"
	$(DOCKER) build -f services/auth-service/Dockerfile -t $(DOCKER_REGISTRY)/auth-service:latest .

docker-push: ## Push all Docker images
	@echo "$(COLOR_GREEN)Pushing Docker images...$(COLOR_RESET)"
	@for service in $(SERVICES); do \
		echo "$(COLOR_BLUE)Pushing $$service...$(COLOR_RESET)"; \
		$(DOCKER) push $(DOCKER_REGISTRY)/$$service:latest || exit 1; \
	done

##@ Local Development

setup-local: ## Complete local development setup (one-command)
	@echo "$(COLOR_GREEN)Running complete local setup...$(COLOR_RESET)"
	@bash scripts/setup-local.sh

run-local: ## Run local development environment with Docker Compose
	@echo "$(COLOR_GREEN)Starting local development environment...$(COLOR_RESET)"
	docker-compose up -d
	@echo "$(COLOR_BLUE)Waiting for services to be healthy...$(COLOR_RESET)"
	@bash scripts/wait-for-services.sh || true
	@echo "$(COLOR_GREEN)✓ All services ready!$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_YELLOW)Connection Information:$(COLOR_RESET)"
	@echo "  PostgreSQL: localhost:5432 (user: giia, db: giia_dev)"
	@echo "  Redis:      localhost:6379 (password: giia_redis_password)"
	@echo "  NATS:       localhost:4222 (monitoring: http://localhost:8222)"

stop-local: ## Stop local development environment
	@echo "$(COLOR_YELLOW)Stopping local development environment...$(COLOR_RESET)"
	docker-compose down
	@echo "$(COLOR_GREEN)Services stopped$(COLOR_RESET)"

restart-local: stop-local run-local ## Restart local development environment

logs-local: ## Show logs from local development environment
	docker-compose logs -f

status-local: ## Show status of local infrastructure services
	@echo "$(COLOR_BLUE)Local Infrastructure Status:$(COLOR_RESET)"
	@echo ""
	@docker-compose ps
	@echo ""
	@echo "$(COLOR_BLUE)Health Checks:$(COLOR_RESET)"
	@docker exec giia-postgres pg_isready -U giia && echo "$(COLOR_GREEN)✓ PostgreSQL: Healthy$(COLOR_RESET)" || echo "$(COLOR_RED)✗ PostgreSQL: Unhealthy$(COLOR_RESET)"
	@docker exec giia-redis redis-cli -a giia_redis_password ping > /dev/null 2>&1 && echo "$(COLOR_GREEN)✓ Redis: Healthy$(COLOR_RESET)" || echo "$(COLOR_RED)✗ Redis: Unhealthy$(COLOR_RESET)"
	@curl -s http://localhost:8222/healthz > /dev/null 2>&1 && echo "$(COLOR_GREEN)✓ NATS: Healthy$(COLOR_RESET)" || echo "$(COLOR_RED)✗ NATS: Unhealthy$(COLOR_RESET)"

clean-local: ## Clean local environment (remove all data and containers)
	@echo "$(COLOR_YELLOW)⚠️  WARNING: This will delete all local data!$(COLOR_RESET)"
	@echo "$(COLOR_YELLOW)Press Ctrl+C to cancel, or wait 5 seconds to continue...$(COLOR_RESET)"
	@sleep 5
	@echo "$(COLOR_RED)Removing containers and volumes...$(COLOR_RESET)"
	docker-compose down -v
	@echo "$(COLOR_GREEN)✓ Local environment cleaned$(COLOR_RESET)"
	@echo "$(COLOR_BLUE)Run 'make run-local' to start fresh$(COLOR_RESET)"

run-tools: ## Start optional development tools (pgAdmin, Redis Commander)
	@echo "$(COLOR_GREEN)Starting development tools...$(COLOR_RESET)"
	docker-compose --profile tools up -d
	@echo "$(COLOR_GREEN)✓ Tools started!$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_YELLOW)Access Tools:$(COLOR_RESET)"
	@echo "  pgAdmin:         http://localhost:5050 (admin@giia.local / admin)"
	@echo "  Redis Commander: http://localhost:8081"

run-service: ## Run a specific service locally (usage: make run-service SERVICE=auth)
	@if [ -z "$(SERVICE)" ]; then \
		echo "$(COLOR_RED)Error: SERVICE not specified$(COLOR_RESET)"; \
		echo "$(COLOR_BLUE)Usage: make run-service SERVICE=auth$(COLOR_RESET)"; \
		exit 1; \
	fi
	@echo "$(COLOR_GREEN)Running $(SERVICE)-service...$(COLOR_RESET)"
	@cd services/$(SERVICE)-service && $(GO) run cmd/api/main.go || $(GO) run cmd/server/main.go

##@ Database

migrate-up: ## Run database migrations
	@echo "$(COLOR_GREEN)Running database migrations...$(COLOR_RESET)"
	@for service in $(SERVICES); do \
		if [ -d "services/$$service/migrations" ]; then \
			echo "$(COLOR_BLUE)Migrating $$service...$(COLOR_RESET)"; \
			# Add migration command here when ready; \
		fi \
	done

migrate-down: ## Rollback database migrations
	@echo "$(COLOR_YELLOW)Rolling back database migrations...$(COLOR_RESET)"
	# Add rollback command here when ready

seed-data: ## Load sample data into local database
	@echo "$(COLOR_GREEN)Loading seed data...$(COLOR_RESET)"
	@docker exec -i giia-postgres psql -U giia -d giia_dev < scripts/seed-data.sql
	@echo "$(COLOR_GREEN)✓ Seed data loaded successfully$(COLOR_RESET)"

##@ Kubernetes - Development Cluster

k8s-setup: ## Setup local Kubernetes cluster with Minikube
	@echo "$(COLOR_GREEN)Setting up local Kubernetes cluster...$(COLOR_RESET)"
	@bash scripts/k8s-setup-cluster.sh

k8s-deploy-infra: ## Deploy infrastructure services (PostgreSQL, Redis, NATS)
	@echo "$(COLOR_GREEN)Deploying infrastructure services...$(COLOR_RESET)"
	@bash scripts/k8s-deploy-infrastructure.sh

k8s-build-images: ## Build and load Docker images into Minikube
	@echo "$(COLOR_GREEN)Building and loading Docker images...$(COLOR_RESET)"
	@bash scripts/k8s-build-images.sh

k8s-deploy-services: ## Deploy all GIIA microservices
	@echo "$(COLOR_GREEN)Deploying GIIA services...$(COLOR_RESET)"
	@bash scripts/k8s-deploy-services.sh

k8s-deploy-auth: ## Deploy auth-service only
	@echo "$(COLOR_BLUE)Deploying auth-service...$(COLOR_RESET)"
	@helm upgrade --install auth-service k8s/services/auth-service/ \
		--namespace giia-dev \
		--values k8s/services/auth-service/values-dev.yaml \
		--wait

k8s-deploy-catalog: ## Deploy catalog-service only
	@echo "$(COLOR_BLUE)Deploying catalog-service...$(COLOR_RESET)"
	@helm upgrade --install catalog-service k8s/services/catalog-service/ \
		--namespace giia-dev \
		--values k8s/services/catalog-service/values-dev.yaml \
		--wait

k8s-status: ## Show Kubernetes cluster status
	@echo "$(COLOR_BLUE)Kubernetes Cluster Status:$(COLOR_RESET)"
	@echo ""
	@kubectl get pods,svc,ingress -n giia-dev

k8s-pods: ## List all pods in giia-dev namespace
	@kubectl get pods -n giia-dev

k8s-logs: ## Tail logs from a specific service (usage: make k8s-logs SERVICE=auth-service)
	@if [ -z "$(SERVICE)" ]; then \
		echo "$(COLOR_RED)Error: SERVICE not specified$(COLOR_RESET)"; \
		echo "$(COLOR_BLUE)Usage: make k8s-logs SERVICE=auth-service$(COLOR_RESET)"; \
		exit 1; \
	fi
	@kubectl logs -f deployment/$(SERVICE) -n giia-dev --all-containers=true

k8s-describe: ## Describe a specific service pod (usage: make k8s-describe SERVICE=auth-service)
	@if [ -z "$(SERVICE)" ]; then \
		echo "$(COLOR_RED)Error: SERVICE not specified$(COLOR_RESET)"; \
		echo "$(COLOR_BLUE)Usage: make k8s-describe SERVICE=auth-service$(COLOR_RESET)"; \
		exit 1; \
	fi
	@kubectl describe deployment/$(SERVICE) -n giia-dev

k8s-shell: ## Open shell in a service pod (usage: make k8s-shell SERVICE=auth-service)
	@if [ -z "$(SERVICE)" ]; then \
		echo "$(COLOR_RED)Error: SERVICE not specified$(COLOR_RESET)"; \
		echo "$(COLOR_BLUE)Usage: make k8s-shell SERVICE=auth-service$(COLOR_RESET)"; \
		exit 1; \
	fi
	@kubectl exec -it deployment/$(SERVICE) -n giia-dev -- /bin/sh

k8s-restart: ## Restart a specific service (usage: make k8s-restart SERVICE=auth-service)
	@if [ -z "$(SERVICE)" ]; then \
		echo "$(COLOR_RED)Error: SERVICE not specified$(COLOR_RESET)"; \
		echo "$(COLOR_BLUE)Usage: make k8s-restart SERVICE=auth-service$(COLOR_RESET)"; \
		exit 1; \
	fi
	@kubectl rollout restart deployment/$(SERVICE) -n giia-dev
	@kubectl rollout status deployment/$(SERVICE) -n giia-dev

k8s-tunnel: ## Start Minikube tunnel for accessing services (run in separate terminal)
	@echo "$(COLOR_YELLOW)Starting Minikube tunnel...$(COLOR_RESET)"
	@echo "$(COLOR_YELLOW)Keep this terminal open. Press Ctrl+C to stop.$(COLOR_RESET)"
	@echo ""
	@minikube tunnel

k8s-dashboard: ## Open Kubernetes dashboard
	@minikube dashboard

k8s-clean: ## Delete all services and infrastructure (keeps cluster running)
	@echo "$(COLOR_YELLOW)Deleting all Helm releases...$(COLOR_RESET)"
	@helm list -n giia-dev --short | xargs -I {} helm uninstall {} -n giia-dev || true
	@echo "$(COLOR_GREEN)All services deleted$(COLOR_RESET)"

k8s-teardown: ## Destroy local Kubernetes cluster completely
	@echo "$(COLOR_YELLOW)Tearing down Kubernetes cluster...$(COLOR_RESET)"
	@bash scripts/k8s-teardown-cluster.sh

k8s-full-deploy: k8s-setup k8s-deploy-infra k8s-build-images k8s-deploy-services ## Complete deployment (setup + infra + build + services)
	@echo ""
	@echo "$(COLOR_GREEN)✓ Full deployment complete!$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_YELLOW)Run in a separate terminal:$(COLOR_RESET) $(COLOR_GREEN)make k8s-tunnel$(COLOR_RESET)"
	@echo "$(COLOR_YELLOW)Then add to /etc/hosts:$(COLOR_RESET)"
	@echo "  $(COLOR_GREEN)127.0.0.1 auth.giia.local catalog.giia.local$(COLOR_RESET)"
	@echo ""

##@ Cleanup

clean: ## Clean build artifacts
	@echo "$(COLOR_YELLOW)Cleaning build artifacts...$(COLOR_RESET)"
	rm -rf bin/
	rm -rf dist/
	rm -f coverage.out coverage.html
	@for service in $(SERVICES); do \
		rm -rf services/$$service/bin/; \
		rm -rf services/$$service/dist/; \
	done
	@echo "$(COLOR_GREEN)Clean complete!$(COLOR_RESET)"

clean-all: clean proto-clean ## Clean everything including generated code
	@echo "$(COLOR_GREEN)Deep clean complete!$(COLOR_RESET)"

##@ Dependencies

deps: ## Download and tidy dependencies
	@echo "$(COLOR_GREEN)Downloading dependencies...$(COLOR_RESET)"
	$(GO) mod download
	$(GO) work sync
	@echo "$(COLOR_GREEN)Dependencies updated!$(COLOR_RESET)"

deps-update: ## Update dependencies
	@echo "$(COLOR_GREEN)Updating dependencies...$(COLOR_RESET)"
	@for service in $(SERVICES); do \
		echo "$(COLOR_BLUE)Updating $$service dependencies...$(COLOR_RESET)"; \
		cd services/$$service && $(GO) get -u ./... && $(GO) mod tidy; \
		cd ../..; \
	done
	$(GO) work sync
	@echo "$(COLOR_GREEN)Dependencies updated!$(COLOR_RESET)"

##@ Information

info: ## Show project information
	@echo "$(COLOR_BOLD)GIIA Core Engine Monorepo$(COLOR_RESET)"
	@echo "$(COLOR_BLUE)Services:$(COLOR_RESET)"
	@for service in $(SERVICES); do \
		echo "  - $$service"; \
	done
	@echo ""
	@echo "$(COLOR_BLUE)Go Version:$(COLOR_RESET)"
	@$(GO) version
	@echo ""
	@echo "$(COLOR_BLUE)Docker Version:$(COLOR_RESET)"
	@$(DOCKER) version --format '{{.Server.Version}}' 2>/dev/null || echo "Not installed"
	@echo ""
	@echo "$(COLOR_BLUE)Kubectl Version:$(COLOR_RESET)"
	@$(KUBECTL) version --client --short 2>/dev/null || echo "Not installed"
