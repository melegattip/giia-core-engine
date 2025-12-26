# =============================================================================
# GIIA AI Intelligence Hub - Multi-stage Docker Build
# =============================================================================
# Build from repository root:
#   docker build -f services/ai-intelligence-hub/Dockerfile -t giia/ai-intelligence-hub .
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
COPY services/ai-intelligence-hub/ ./services/ai-intelligence-hub/

WORKDIR /build/services/ai-intelligence-hub

# Disable Go workspace mode for Docker build
ENV GOWORK=off

# Resolve and download dependencies
RUN go mod tidy && go mod download

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /app/ai-intelligence-hub \
    ./cmd/api

# -----------------------------------------------------------------------------
# Stage 2: Runtime
# -----------------------------------------------------------------------------
FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata wget

RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /app/ai-intelligence-hub .

RUN chown -R appuser:appgroup /app

USER appuser

# HTTP API
EXPOSE 8086
# gRPC
EXPOSE 9094
# WebSocket
EXPOSE 8087

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget -q -O- http://localhost:8086/health || exit 1

ENV ENVIRONMENT=production

ENTRYPOINT ["./ai-intelligence-hub"]
