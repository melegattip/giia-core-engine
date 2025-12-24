---
description: Infrastructure and DevOps tool commands for GIIA Core Engine
---

# Infrastructure & DevOps Workflow

This workflow defines commands for infrastructure tooling including Docker, Kubernetes, and protobuf generation.

## Container Tools

// turbo
1. Validate Docker Compose configuration:
```bash
docker compose config
```

## Kubernetes Tools

// turbo
2. Check kubectl version:
```bash
kubectl version
```

// turbo
3. Check Helm version:
```bash
helm version
```

// turbo
4. Check Minikube version:
```bash
minikube version
```

## Protocol Buffers

// turbo
5. Generate protobuf code:
```bash
protoc <options>
```

// turbo
6. Run proto generation script:
```bash
bash scripts/generate-proto.sh
```

## Package Managers

// turbo
7. Install via winget:
```bash
winget install <package>
```

// turbo
8. Install via chocolatey:
```bash
choco install <package>
```
