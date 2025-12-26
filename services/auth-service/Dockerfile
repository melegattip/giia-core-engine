# =============================================================================
# GIIA Auth Service - Multi-stage Docker Build
# =============================================================================
# This Dockerfile builds the auth-service as part of the GIIA monorepo.
# It copies shared packages from pkg/ and builds in standalone mode (GOWORK=off).
#
# Build from repository root:
#   docker build -f services/auth-service/Dockerfile -t giia/auth-service .
# =============================================================================

# -----------------------------------------------------------------------------
# Stage 1: Build
# -----------------------------------------------------------------------------
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory at repository root level
WORKDIR /build

# Copy shared packages first (for better caching)
COPY pkg/ ./pkg/

# Copy auth-service files
COPY services/auth-service/ ./services/auth-service/

# Set working directory to the service
WORKDIR /build/services/auth-service

# Disable Go workspace mode for Docker build
# The service has replace directives in go.mod for local packages
ENV GOWORK=off

# Resolve and download dependencies
# go mod tidy is needed to capture transitive deps from pkg/
RUN go mod tidy && go mod download

# Build the application
# CGO_ENABLED=0 for static binary
# -ldflags="-w -s" for smaller binary (strips debug info)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
  -ldflags="-w -s -X main.version=$(date +%Y%m%d)" \
  -o /app/auth-service \
  ./cmd/api

# -----------------------------------------------------------------------------
# Stage 2: Runtime
# -----------------------------------------------------------------------------
FROM alpine:3.20

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata wget

# Create non-root user for security
RUN addgroup -g 1001 -S appgroup && \
  adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/auth-service .

# Copy migrations if they exist (for auto-migration support)
# COPY --from=builder /build/services/auth-service/migrations ./migrations 2>/dev/null || true

# Change ownership
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose ports
# HTTP API
EXPOSE 8080
# gRPC
EXPOSE 9091

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Set default environment variables
ENV GIN_MODE=release
ENV ENVIRONMENT=production

# Run the application
ENTRYPOINT ["./auth-service"]