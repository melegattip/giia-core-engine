.PHONY: help setup build test clean lint proto docker-build docker-push run-local

# Variables
GO := go
DOCKER := docker
KUBECTL := kubectl
SERVICES := auth-service catalog-service ddmrp-engine-service execution-service analytics-service ai-agent-service
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

build-catalog: ## Build catalog service
	@echo "$(COLOR_BLUE)Building catalog-service...$(COLOR_RESET)"
	cd services/catalog-service && $(GO) build -o ../../bin/catalog-service ./cmd/server/

build-ddmrp: ## Build ddmrp service
	@echo "$(COLOR_BLUE)Building ddmrp-engine-service...$(COLOR_RESET)"
	cd services/ddmrp-engine-service && $(GO) build -o ../../bin/ddmrp-engine-service ./cmd/server/

build-execution: ## Build execution service
	@echo "$(COLOR_BLUE)Building execution-service...$(COLOR_RESET)"
	cd services/execution-service && $(GO) build -o ../../bin/execution-service ./cmd/server/

build-analytics: ## Build analytics service
	@echo "$(COLOR_BLUE)Building analytics-service...$(COLOR_RESET)"
	cd services/analytics-service && $(GO) build -o ../../bin/analytics-service ./cmd/server/

build-ai: ## Build ai-agent service
	@echo "$(COLOR_BLUE)Building ai-agent-service...$(COLOR_RESET)"
	cd services/ai-agent-service && $(GO) build -o ../../bin/ai-agent-service ./cmd/server/

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

test-catalog: ## Run catalog service tests
	@echo "$(COLOR_BLUE)Testing catalog-service...$(COLOR_RESET)"
	cd services/catalog-service && $(GO) test -v -race -count=1 ./...

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

##@ Kubernetes

k8s-dev-deploy: ## Deploy to development Kubernetes cluster
	@echo "$(COLOR_GREEN)Deploying to Kubernetes (dev)...$(COLOR_RESET)"
	$(KUBECTL) apply -f deployments/dev/

k8s-dev-delete: ## Delete from development Kubernetes cluster
	@echo "$(COLOR_YELLOW)Deleting from Kubernetes (dev)...$(COLOR_RESET)"
	$(KUBECTL) delete -f deployments/dev/

k8s-logs: ## Tail logs from Kubernetes pods
	@echo "$(COLOR_BLUE)Tailing Kubernetes logs...$(COLOR_RESET)"
	$(KUBECTL) logs -f -l app=giia --all-containers=true -n giia-dev

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
