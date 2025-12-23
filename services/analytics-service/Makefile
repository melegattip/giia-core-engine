.PHONY: help build run test coverage lint proto migrate-up migrate-down clean

help:
	@echo "Available targets:"
	@echo "  build        - Build the service binary"
	@echo "  run          - Run the service locally"
	@echo "  test         - Run unit tests"
	@echo "  coverage     - Generate test coverage report"
	@echo "  lint         - Run linters"
	@echo "  proto        - Generate protobuf code"
	@echo "  migrate-up   - Apply database migrations"
	@echo "  migrate-down - Rollback database migrations"
	@echo "  clean        - Remove build artifacts"

build:
	@echo "Building Analytics Service..."
	go build -o bin/analytics-service ./cmd/server

run:
	@echo "Running Analytics Service..."
	go run ./cmd/server/main.go

test:
	@echo "Running tests..."
	go test ./... -v -count=1

test-unit:
	@echo "Running unit tests..."
	go test ./internal/core/... -v -count=1

coverage:
	@echo "Generating coverage report..."
	go test ./... -coverprofile=coverage.out -covermode=atomic
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

coverage-core:
	@echo "Generating core package coverage..."
	go test ./internal/core/... -coverprofile=coverage-core.out -covermode=atomic
	go tool cover -func=coverage-core.out

lint:
	@echo "Running linters..."
	golangci-lint run ./...

proto:
	@echo "Generating protobuf code..."
	protoc --go_out=. --go-grpc_out=. \
		--go_opt=paths=source_relative \
		--go-grpc_opt=paths=source_relative \
		api/proto/analytics/v1/analytics.proto

migrate-up:
	@echo "Applying database migrations..."
	@echo "Note: Run migrations manually or use a migration tool"

migrate-down:
	@echo "Rolling back database migrations..."
	@echo "Note: Run rollback manually or use a migration tool"

clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -f coverage.out coverage.html coverage-core.out

deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

vendor:
	@echo "Vendoring dependencies..."
	go mod vendor
