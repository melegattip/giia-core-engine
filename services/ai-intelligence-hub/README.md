# AI Intelligence Hub Service

**Version**: 1.0.0  
**Status**: 🟢 80% Complete - MVP Operational  
**Phase**: 2B - New Microservices  
**Last Updated**: 2025-12-23

AI-powered notification and intelligence system for GIIA platform that monitors events and provides proactive insights.

## Quick Start

### Prerequisites
- Go 1.23.4+
- PostgreSQL 16
- NATS Server (JetStream enabled)

### Setup

1. **Install Dependencies**
```bash
go mod download
```

2. **Configure Environment**
```bash
cp .env.example .env
# Edit .env with your configuration
```

3. **Run Database Migrations**
```bash
make migrate-up DATABASE_URL="postgresql://postgres:postgres@localhost:5432/intelligence_hub?sslmode=disable"
```

4. **Build and Run**
```bash
make run
```

## Development

### Build
```bash
make build
```

### Run Tests
```bash
make test
```

### Test Coverage
```bash
make test-coverage
```

### Lint Code
```bash
make lint
```

### Format Code
```bash
make fmt
```

## Architecture

### Directory Structure
```
ai-intelligence-hub/
├── cmd/
│   └── api/              # Main service entry point
├── internal/
│   ├── core/
│   │   ├── domain/       # Domain entities
│   │   ├── providers/    # Interfaces
│   │   └── usecases/     # Business logic
│   └── infrastructure/
│       ├── adapters/     # External integrations
│       └── repositories/ # Data access
├── migrations/           # Database migrations
└── Makefile
```

### Key Components

#### Event Processing
- Subscribes to NATS JetStream events
- Routes events to appropriate handlers
- Processes buffer, execution, and user events

#### Notifications
- Creates AI-powered notifications
- Stores notifications in PostgreSQL
- Supports multiple priority levels and types

## Configuration

### Environment Variables

#### Required
- `DATABASE_URL` - PostgreSQL connection string
- `NATS_SERVERS` - NATS server URLs (comma-separated)

#### Optional
- `LOG_LEVEL` - Logging level (debug, info, warn, error)
- `CLAUDE_API_KEY` - For AI analysis (future feature)
- `CHROMADB_HOST` - For RAG system (future feature)

## Database

### Migrations

**Run migrations**
```bash
make migrate-up DATABASE_URL="your-db-url"
```

**Rollback**
```bash
make migrate-down DATABASE_URL="your-db-url"
```

### Tables
- `ai_notifications` - Notification storage
- `ai_recommendations` - Recommendation storage
- `user_notification_preferences` - User preferences

## Current MVP Features

✅ Event subscription from NATS
✅ Buffer event processing
✅ Notification creation and storage
✅ Database persistence
✅ Clean architecture structure

## Roadmap

### 🔨 Pending Items

**Testing** (~50% → 80% goal)
- Additional use case tests
- Repository integration tests
- Event handler tests

**API Layer**
- HTTP/REST API endpoints for notifications
- gRPC service definitions
- WebSocket for real-time push

**Advanced Features**
- Real Claude API integration (currently using mocks)
- ChromaDB RAG system (currently keyword-based)
- Multi-channel notification delivery (Email, SMS, Slack)
- Pattern detection across events
- Daily digest generation

---

## Implementation Status

**Current**: 🟢 80% Complete (MVP Operational)

| Component | Status | Notes |
|-----------|--------|-------|
| Domain Entities | ✅ 100% | Notification, Recommendation, UserPreferences |
| Domain Tests | ✅ 100% | notification_test.go |
| Use Cases | ✅ 100% | Analysis, Event Processing |
| Use Case Tests | ✅ 75% | 2 test files, more needed |
| Repositories | ✅ 100% | PostgreSQL notification repository |
| Adapters | ✅ 100% | Claude client, NATS subscriber, RAG retriever |
| Event Handlers | ✅ 100% | Buffer, Execution, User events |
| Event Handler Tests | ✅ 33% | 1 test file |
| Database Migrations | ✅ 100% | 4 migration files |
| Main Entry Point | ✅ 100% | Full DI in main.go |
| HTTP/gRPC API | ⏸️ 0% | Not started |
| Integration Tests | ⏸️ 0% | Not started |

**Next Steps**:
1. Add more unit tests (target 80%+)
2. Implement HTTP REST endpoints
3. Add gRPC service definitions
4. Integrate real Claude API
5. Implement ChromaDB for vector RAG

---

## Testing

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run specific test
go test -v ./internal/core/domain/...
```

## Contributing

Follow GIIA development standards (see [CLAUDE.md](../../CLAUDE.md))

- Clean Architecture principles
- 80%+ test coverage
- Go coding standards
- Typed errors from pkg/errors

## License

Proprietary - GIIA Platform © 2025
