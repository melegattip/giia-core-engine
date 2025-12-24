---
description: Go development commands and workflows for GIIA Core Engine
---

# Go Development Workflow

This workflow defines safe commands for Go development in the GIIA Core Engine project.

## Module Management

// turbo
1. Sync Go workspace dependencies:
```bash
go work sync
```

// turbo
2. Download module dependencies:
```bash
go mod download
```

// turbo
3. Tidy module dependencies:
```bash
go mod tidy
```

// turbo
4. Vendor workspace dependencies:
```bash
go work vendor
```

// turbo
5. Verify module checksums:
```bash
go mod verify
```

// turbo
6. Edit module file:
```bash
go mod edit <args>
```

// turbo
7. Add workspace module:
```bash
go work use <path>
```

## Building

// turbo
8. Build Go packages:
```bash
go build ./...
```

// turbo
9. Build with workspace disabled:
```bash
GOWORK=off go build ./...
```

// turbo
10. Build auth service via make:
```bash
make build-auth
```

// turbo
11. Build all services via make:
```bash
make build
```

## Testing

// turbo
12. Run Go tests:
```bash
go test ./...
```

// turbo
13. Run tests with workspace disabled:
```bash
GOWORK=off go test ./...
```

// turbo
14. Run tests via make:
```bash
make test
```

// turbo
15. Generate test coverage:
```bash
go tool cover -html=coverage.out
```

## Code Quality

// turbo
16. Vet Go packages:
```bash
go vet ./...
```

// turbo
17. Run golangci-lint:
```bash
golangci-lint run
```

// turbo
18. Run golangci-lint from home:
```bash
~/go/bin/golangci-lint run
```

// turbo
19. Format Go code:
```bash
go fmt ./...
```

// turbo
20. Format with gofmt:
```bash
gofmt -w .
```

## Utilities

// turbo
21. Check Go version:
```bash
go version
```

// turbo
22. List Go packages:
```bash
go list ./...
```

// turbo
23. Install Go tools:
```bash
go install <package>
```

// turbo
24. Get Go dependencies:
```bash
go get <package>
```

// turbo
25. Run Go programs:
```bash
go run <file>
```
