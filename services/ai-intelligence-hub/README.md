# AI Intelligence Hub Service

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

- [ ] Claude API integration for AI analysis
- [ ] RAG system with ChromaDB
- [ ] Multi-channel notification delivery
- [ ] Pattern detection
- [ ] Daily digest generation
- [ ] HTTP/gRPC API endpoints

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
