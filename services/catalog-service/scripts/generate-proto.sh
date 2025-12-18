#!/bin/bash

# Generate Protocol Buffers for Catalog Service

PROTO_DIR="api/proto"
OUT_DIR="api/proto/gen/go"

# Create output directory
mkdir -p $OUT_DIR

# Generate Go code from proto files
protoc \
  --go_out=$OUT_DIR \
  --go_opt=paths=source_relative \
  --go-grpc_out=$OUT_DIR \
  --go-grpc_opt=paths=source_relative \
  --proto_path=$PROTO_DIR \
  $PROTO_DIR/catalog/v1/*.proto

echo "✅ Protocol Buffers generated successfully"