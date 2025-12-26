# =============================================================================
# GIIA Analytics Service - Multi-stage Docker Build
# =============================================================================
# Build from repository root:
#   docker build -f services/analytics-service/Dockerfile -t giia/analytics-service .
# =============================================================================

# -----------------------------------------------------------------------------
# Stage 1: Build
# -----------------------------------------------------------------------------
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata make

WORKDIR /build

# Copy shared packages
COPY pkg/ ./pkg/

# Copy service files
COPY services/analytics-service/ ./services/analytics-service/

WORKDIR /build/services/analytics-service

# Disable Go workspace mode for Docker build
ENV GOWORK=off

# Resolve and download dependencies
RUN go mod tidy && go mod download

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /app/analytics-service \
    ./cmd/server

# -----------------------------------------------------------------------------
# Stage 2: Runtime
# -----------------------------------------------------------------------------
FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata wget

RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /app/analytics-service .

RUN chown -R appuser:appgroup /app

USER appuser

EXPOSE 8085
EXPOSE 9093

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget -q -O- http://localhost:8085/health || exit 1

ENV ENVIRONMENT=production

ENTRYPOINT ["./analytics-service"]
