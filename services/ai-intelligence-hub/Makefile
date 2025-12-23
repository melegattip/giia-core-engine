.PHONY: build run test test-coverage lint fmt clean migrate-up migrate-down

SERVICE_NAME=ai-intelligence-hub
BINARY_NAME=ai-intelligence-hub
GO_FILES=$(shell find . -name '*.go' -type f -not -path './vendor/*')

build:
	@echo "Building $(SERVICE_NAME)..."
	@go build -o bin/$(BINARY_NAME) cmd/api/main.go
	@echo "Build complete: bin/$(BINARY_NAME)"

run: build
	@echo "Running $(SERVICE_NAME)..."
	@./bin/$(BINARY_NAME)

test:
	@echo "Running tests..."
	@go test -v -race -count=1 ./...

test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

lint:
	@echo "Running linters..."
	@golangci-lint run

fmt:
	@echo "Formatting code..."
	@gofmt -w $(GO_FILES)
	@go vet ./...

clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

migrate-up:
	@echo "Running migrations up..."
	@migrate -path=migrations -database="$(DATABASE_URL)" up

migrate-down:
	@echo "Running migrations down..."
	@migrate -path=migrations -database="$(DATABASE_URL)" down 1

dev:
	@echo "Running in development mode with hot reload..."
	@air

deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies updated"

help:
	@echo "Available targets:"
	@echo "  build         - Build the service binary"
	@echo "  run           - Build and run the service"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  lint          - Run linters"
	@echo "  fmt           - Format code"
	@echo "  clean         - Remove build artifacts"
	@echo "  migrate-up    - Run database migrations up"
	@echo "  migrate-down  - Run database migrations down"
	@echo "  dev           - Run with hot reload (requires air)"
	@echo "  deps          - Update dependencies"
