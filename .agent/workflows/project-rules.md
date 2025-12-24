---
description: GIIA Core Engine project conventions and coding standards
---

# GIIA Core Engine - Project Rules

## Repository Information

- **Repository**: `melegattip/giia-core-engine`
- **Primary Branches**: `main`, `develop`
- **Environments**: `development`, `staging`, `production`

## Architecture

This project follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles:

### Directory Structure
```
services/
├── <service-name>/
│   ├── cmd/           # Entry points
│   ├── internal/
│   │   ├── domain/    # Entities, value objects, domain events
│   │   ├── usecases/  # Application business logic
│   │   ├── handlers/  # HTTP/gRPC handlers
│   │   └── repository/# Data access implementations
│   ├── go.mod
│   └── Makefile
└── ...
```

### Coding Standards

1. **Go workspace**: This project uses Go workspaces (`go.work`)
2. **Module path**: `github.com/melegattip/giia-core-engine`
3. **Test coverage**: Aim for 80%+ code coverage
4. **Linting**: Use `golangci-lint` for code quality

### Workflow Guidelines

1. Always run `go mod tidy` after adding dependencies
2. Run `go work sync` when updating workspace modules
3. Use `make build` to build all services
4. Use `make test` to run all tests
5. Format code with `go fmt ./...` before committing

### PR Requirements

- All PRs require 2 approving reviews
- Linear history required (no merge commits)
- Conversation resolution required before merge
- Stale reviews are dismissed on new commits

### Environment Deployments

- `development`: Deploys from `develop` branch
- `staging`: Deploys from version tags (`v*`)
- `production`: Deploys from version tags (`v*`)
