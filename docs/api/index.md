# GIIA Platform API Documentation

**Version:** 1.0  
**Last Updated:** 2025-12-24

Welcome to the GIIA Platform API documentation. This portal provides comprehensive documentation for all GIIA services, including REST APIs, gRPC contracts, and integration guides.

---

## 🚀 Quick Links

- [Getting Started](./getting-started.md)
- [Authentication Guide](./authentication.md)
- [Code Examples](./examples/)
- [Public API RFC](./PUBLIC_RFC.md)
- [gRPC Contracts](./GRPC_CONTRACTS.md)

---

## 📋 Service Overview

| Service | REST Port | gRPC Port | Description |
|---------|-----------|-----------|-------------|
| **Auth Service** | 8081 | 9081 | Authentication, authorization, multi-tenancy |
| **Catalog Service** | 8082 | 9082 | Products, suppliers, buffer profiles |
| **DDMRP Engine** | 8083 | 9083 | Buffer calculations, demand adjustments |
| **Execution Service** | 8084 | 9084 | Orders, inventory, transactions |
| **Analytics Service** | 8085 | 9085 | KPIs, metrics, reporting |
| **AI Intelligence Hub** | 8086 | 9086 | Notifications, recommendations, insights |

---

## 📡 REST API Documentation

### OpenAPI Specifications

Each service exposes an OpenAPI 3.0 specification:

| Service | OpenAPI Spec | Swagger UI |
|---------|--------------|------------|
| Auth | [openapi.yaml](/services/auth-service/docs/openapi.yaml) | `/docs` |
| Catalog | [openapi.yaml](/services/catalog-service/docs/openapi.yaml) | `/docs` |
| Execution | [openapi.yaml](/services/execution-service/docs/openapi.yaml) | `/docs` |
| Analytics | [openapi.yaml](/services/analytics-service/docs/openapi.yaml) | `/docs` |
| AI Hub | [openapi.yaml](/services/ai-intelligence-hub/docs/openapi.yaml) | `/docs` |

### Base URLs

```
# Local Development
Auth:       http://localhost:8081/api/v1
Catalog:    http://localhost:8082/api/v1
DDMRP:      http://localhost:8083/api/v1
Execution:  http://localhost:8084/api/v1
Analytics:  http://localhost:8085/api/v1
AI Hub:     http://localhost:8086/api/v1

# Production
All:        https://api.giia.io/{service}/api/v1
```

---

## 🔗 gRPC API Documentation

Internal service-to-service communication uses gRPC for performance:

| Service | Package | Port | Documentation |
|---------|---------|------|---------------|
| Auth | `auth.v1` | 9081 | [auth.md](./grpc/auth.md) |
| Catalog | `catalog.v1` | 9082 | [catalog.md](./grpc/catalog.md) |
| DDMRP | `ddmrp.v1` | 9083 | [ddmrp.md](./grpc/ddmrp.md) |
| Execution | `execution.v1` | 9084 | [execution.md](./grpc/execution.md) |
| Analytics | `analytics.v1` | 9085 | [analytics.md](./grpc/analytics.md) |
| AI Hub | `ai.v1` | 9086 | [ai-hub.md](./grpc/ai-hub.md) |

### Proto File Locations

```
services/
├── auth-service/api/proto/auth/v1/auth.proto
├── catalog-service/api/proto/catalog/v1/catalog.proto
├── ddmrp-engine-service/api/proto/ddmrp/v1/ddmrp.proto
├── execution-service/api/proto/execution/v1/execution.proto
├── analytics-service/api/proto/analytics/v1/analytics.proto
└── ai-intelligence-hub/api/proto/ai/v1/ai.proto
```

---

## 🔐 Authentication

All API requests require authentication via JWT Bearer tokens:

```bash
# 1. Login to get tokens
curl -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "password"}'

# 2. Use access token in subsequent requests
curl http://localhost:8082/api/v1/products \
  -H "Authorization: Bearer <access_token>" \
  -H "X-Organization-ID: <org_id>"
```

See [Authentication Guide](./authentication.md) for details.

---

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    API Gateway (Kong)                        │
│                  Rate Limiting, Auth, Routing                │
└─────────────────────────────────────────────────────────────┘
                              │
        ┌─────────────────────┼─────────────────────┐
        │                     │                     │
        ▼                     ▼                     ▼
┌───────────────┐   ┌───────────────┐   ┌───────────────┐
│  Auth Service │   │Catalog Service│   │   DDMRP       │
│   REST+gRPC   │   │   REST+gRPC   │   │   Engine      │
│    :8081/9081 │   │    :8082/9082 │   │  :8083/9083   │
└───────────────┘   └───────────────┘   └───────────────┘
        │                     │                     │
        │         ┌───────────┼───────────┐        │
        │         │           │           │        │
        ▼         ▼           ▼           ▼        ▼
┌───────────────┐   ┌───────────────┐   ┌───────────────┐
│  Execution    │   │  Analytics    │   │   AI Hub      │
│   Service     │   │   Service     │   │   Service     │
│  :8084/9084   │   │  :8085/9085   │   │  :8086/9086   │
└───────────────┘   └───────────────┘   └───────────────┘
        │                     │                     │
        └─────────────────────┼─────────────────────┘
                              │
                              ▼
                    ┌───────────────┐
                    │  NATS JetStream│
                    │  Event Bus    │
                    └───────────────┘
```

---

## 📦 Key Concepts

### Multi-Tenancy

All data is isolated by `organization_id`. Include the organization context:
- REST: `X-Organization-ID` header
- gRPC: `organization_id` field in requests
- JWT: Encoded in token claims

### DDMRP Buffer Zones

The platform implements DDMRP (Demand Driven MRP):

```
┌─────────────────────┐
│     Green Zone      │  ← Healthy stock
├─────────────────────┤
│    Yellow Zone      │  ← Attention needed
├─────────────────────┤
│     Red Zone        │  ← Critical, reorder
└─────────────────────┘

NFP = On Hand + On Order - Qualified Demand
```

### Event-Driven Architecture

Services communicate via NATS JetStream events:

```
inventory.transaction.created  → Updates buffer NFP
buffer.zone.changed           → Triggers AI alerts
order.status.changed          → Updates analytics
```

---

## 🛠️ Development Tools

### Generate API Clients

```bash
# Go gRPC client
protoc --go_out=. --go-grpc_out=. api/proto/**/*.proto

# OpenAPI client (any language)
openapi-generator generate -i openapi.yaml -g go -o ./client
```

### Testing APIs

```bash
# gRPC
grpcurl -plaintext localhost:9081 list
grpcurl -plaintext -d '{"token":"..."}' localhost:9081 auth.v1.AuthService/ValidateToken

# REST
curl http://localhost:8081/health
```

### Swagger UI

Each service exposes Swagger UI at `/docs`:
- http://localhost:8081/docs
- http://localhost:8082/docs
- etc.

---

## 📚 Additional Resources

- [Architecture Overview](/docs/architecture/OVERVIEW.md)
- [Microservices Documentation](/docs/architecture/MICROSERVICES.md)
- [Development Guide](/docs/development/DEVELOPMENT_GUIDE.md)
- [DDMRP Concepts](/docs/specifications/DDMRP_CONCEPTS.md)

---

## 🆘 Support

- **Documentation Issues**: Open an issue in the repository
- **API Questions**: Contact the GIIA Team
- **Security Issues**: security@giia.io

---

**GIIA Platform API Documentation** - Built with ❤️ by the GIIA Team
