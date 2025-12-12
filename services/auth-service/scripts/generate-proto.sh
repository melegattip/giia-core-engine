#!/bin/bash

# Script to generate Go code from protocol buffer definitions
# Usage: ./scripts/generate-proto.sh

set -e

echo "🔧 Generating Go code from proto files..."

# Check if protoc is installed
if ! command -v protoc &> /dev/null; then
    echo "❌ protoc not found. Please install Protocol Buffers compiler."
    echo "   Download from: https://github.com/protocolbuffers/protobuf/releases"
    exit 1
fi

# Check if protoc-gen-go is installed
if ! command -v protoc-gen-go &> /dev/null; then
    echo "❌ protoc-gen-go not found. Installing..."
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
fi

# Check if protoc-gen-go-grpc is installed
if ! command -v protoc-gen-go-grpc &> /dev/null; then
    echo "❌ protoc-gen-go-grpc not found. Installing..."
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
fi

# Clean previous generated files
rm -rf api/proto/gen/go/*

# Create output directory
mkdir -p api/proto/gen/go/auth/v1

# Generate Go code from proto files
protoc \
  --go_out=api/proto/gen/go \
  --go_opt=paths=source_relative \
  --go-grpc_out=api/proto/gen/go \
  --go-grpc_opt=paths=source_relative \
  --proto_path=api/proto \
  auth/v1/messages.proto \
  auth/v1/auth.proto

echo "✅ Proto generation completed successfully!"
echo "📁 Generated files in: api/proto/gen/go/auth/v1/"
