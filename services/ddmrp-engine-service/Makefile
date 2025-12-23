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
	@echo "Building DDMRP Engine Service..."
	go build -o bin/ddmrp-engine-service ./cmd/server

run:
	@echo "Running DDMRP Engine Service..."
	go run ./cmd/server/main.go

test:
	@echo "Running tests..."
	go test ./... -v -count=1

coverage:
	@echo "Generating coverage report..."
	go test ./... -coverprofile=coverage.out -covermode=atomic
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

lint:
	@echo "Running linters..."
	golangci-lint run ./...

proto:
	@echo "Generating protobuf code..."
	protoc --go_out=. --go-grpc_out=. \
		--go_opt=paths=source_relative \
		--go-grpc_opt=paths=source_relative \
		api/proto/ddmrp/v1/ddmrp.proto

migrate-up:
	@echo "Applying database migrations..."
	@echo "Note: Run migrations manually or use a migration tool"

migrate-down:
	@echo "Rolling back database migrations..."
	@echo "Note: Run rollback manually or use a migration tool"

clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -f coverage.out coverage.html

deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

vendor:
	@echo "Vendoring dependencies..."
	go mod vendor
