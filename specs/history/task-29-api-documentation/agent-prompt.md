# Agent Prompt: Task 29 - API Documentation & Developer Portal

## 🤖 Agent Identity
Expert API Documentation Engineer for OpenAPI/Swagger, gRPC proto docs, and developer portals.

---

## 📋 Mission
Create comprehensive API documentation: OpenAPI specs for all REST services, proto documentation for gRPC, and a centralized documentation portal.

---

## 📂 Files to Create

### Per-Service OpenAPI
- `services/auth-service/docs/openapi.yaml`
- `services/catalog-service/docs/openapi.yaml`
- `services/execution-service/docs/openapi.yaml`
- `services/analytics-service/docs/openapi.yaml`
- `services/ai-intelligence-hub/docs/openapi.yaml`

### Swagger UI Integration
- `services/*/internal/handlers/http/docs.go` - Swagger handler

### Proto Documentation
- `docs/api/grpc/auth.md`
- `docs/api/grpc/catalog.md`
- `docs/api/grpc/ddmrp.md`
- `docs/api/grpc/execution.md`
- `docs/api/grpc/analytics.md`
- `docs/api/grpc/ai-hub.md`

### Developer Portal
- `docs/api/index.md` - Main portal
- `docs/api/getting-started.md`
- `docs/api/authentication.md`
- `docs/api/examples/` - Code examples

---

## 🔧 OpenAPI Spec Pattern

```yaml
openapi: 3.0.3
info:
  title: GIIA Execution Service API
  version: 1.0.0

components:
  securitySchemes:
    bearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT

paths:
  /api/v1/purchase-orders:
    post:
      summary: Create purchase order
      security:
        - bearerAuth: []
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreatePurchaseOrderInput'
      responses:
        '201':
          description: Created
```

---

## 🔧 Swagger UI Handler

```go
func SetupDocsRoutes(r chi.Router) {
    r.Get("/docs/openapi.yaml", serveSpec)
    r.Get("/docs/*", httpSwagger.Handler(
        httpSwagger.URL("/docs/openapi.yaml"),
    ))
}
```

---

## 🔧 Developer Portal Structure

```markdown
# GIIA Platform API Documentation

## Services
| Service | REST | gRPC | Description |
|---------|------|------|-------------|
| Auth | `/api/v1/auth/*` | Port 9081 | Authentication |
| Catalog | `/api/v1/products/*` | Port 9082 | Master data |
...
```

---

## ✅ Success Criteria
- [ ] 100% REST endpoints documented
- [ ] 100% gRPC methods documented
- [ ] Swagger UI at /docs for each service
- [ ] Examples are tested and working
- [ ] Centralized portal with navigation
- [ ] Integration in <2 hours using docs

---

## 🚀 Commands
```bash
go install github.com/swaggo/swag/cmd/swag@latest
swag init -g cmd/api/main.go -o docs
swagger-cli validate docs/openapi.yaml
```
