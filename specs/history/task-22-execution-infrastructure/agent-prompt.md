# Agent Prompt: Task 22 - Execution Service Infrastructure Layer

## 🤖 Agent Identity
Expert Go Infrastructure Engineer for database-driven microservices with GORM, PostgreSQL, and Clean Architecture.

---

## 📋 Mission
Build infrastructure for Execution Service: 5 GORM repositories, 7 database migrations, and main.go entry point.

---

## 📂 Files to Create

### Repositories (internal/repository/postgres/)
- `purchase_order_repository.go` + `_test.go`
- `sales_order_repository.go` + `_test.go`
- `inventory_transaction_repository.go` + `_test.go`
- `inventory_balance_repository.go` + `_test.go`
- `alert_repository.go` + `_test.go`

### Migrations (migrations/)
- `000001_create_purchase_orders.up.sql` / `.down.sql`
- `000002_create_purchase_order_lines.up.sql` / `.down.sql`
- `000003_create_sales_orders.up.sql` / `.down.sql`
- `000004_create_sales_order_lines.up.sql` / `.down.sql`
- `000005_create_inventory_transactions.up.sql` / `.down.sql`
- `000006_create_inventory_balances.up.sql` / `.down.sql`
- `000007_create_alerts.up.sql` / `.down.sql`

### Entry Point
- `cmd/api/main.go`

---

## 🔧 Key Requirements

### Multi-Tenancy
All queries MUST scope by `organization_id`:
```go
func (r *repo) scopeByOrg(orgID uuid.UUID) *gorm.DB {
    return r.db.Where("organization_id = ?", orgID)
}
```

### Repository Methods
- Create, GetByID, List (with filters), Update, Delete
- Specialized queries (e.g., GetPendingByProduct)
- Eager loading for associations (Preload)

### Migration Standards
- Use `gen_random_uuid()` for UUIDs
- Add indexes on foreign keys and query patterns
- Include UNIQUE constraints
- Support up AND down migrations

### main.go
- Database connection with pool config
- Wire repositories → use cases
- Health endpoints: `/health`, `/ready`
- Graceful shutdown

---

## ✅ Success Criteria
- [ ] 5 repositories with 85%+ test coverage
- [ ] 7 migrations (reversible)
- [ ] Service starts with `make run`
- [ ] Health check returns 200 OK
- [ ] Operations <50ms

---

## 🚀 Commands
```bash
cd services/execution-service
migrate -path migrations -database $DATABASE_URL up
go test ./internal/repository/... -cover
go build -o bin/execution-service ./cmd/api
curl http://localhost:8084/health
```
