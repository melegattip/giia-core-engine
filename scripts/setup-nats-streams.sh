#!/bin/bash

# NATS Jetstream Stream Setup Script
# Creates streams for all services in GIIA Core Engine

set -e

NATS_URL="${NATS_URL:-nats://localhost:4222}"

echo "Setting up NATS Jetstream streams..."
echo "NATS URL: $NATS_URL"
echo ""

nats_stream_add() {
    local stream_name=$1
    local subjects=$2

    echo "Creating stream: $stream_name"
    nats stream add "$stream_name" \
        --subjects="$subjects" \
        --storage=file \
        --retention=limits \
        --max-age=7d \
        --max-bytes=1G \
        --replicas=1 \
        --discard=old \
        --max-msg-size=8MB \
        --dupe-window=2m \
        --server="$NATS_URL" \
        --defaults || echo "Stream $stream_name may already exist"
    echo ""
}

nats_stream_add "AUTH_EVENTS" "auth.>"
nats_stream_add "CATALOG_EVENTS" "catalog.>"
nats_stream_add "DDMRP_EVENTS" "ddmrp.>"
nats_stream_add "EXECUTION_EVENTS" "execution.>"
nats_stream_add "ANALYTICS_EVENTS" "analytics.>"
nats_stream_add "AI_AGENT_EVENTS" "ai_agent.>"
nats_stream_add "DLQ_EVENTS" "dlq.>"

echo "✅ All streams created successfully!"
echo ""
echo "Verify streams with: nats stream list --server=$NATS_URL"
