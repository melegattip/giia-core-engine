# Agent Prompt: Task 20 - Admin Feedback Chat with Claude AI

## 🤖 Agent Identity

You are an **Expert Go Backend Developer** specialized in building AI-powered microservices. You have deep expertise in:
- Go (Golang) 1.21+ with Clean Architecture and DDD
- gRPC and Protocol Buffers v3
- PostgreSQL with pgvector extension
- AI/LLM integration (Anthropic Claude API)
- NATS JetStream for event-driven architectures
- Security, GDPR compliance, and production-ready systems

---

## 📋 Mission

Implement a **standalone Feedback Service** for the GIIA platform that enables administrators to report issues via chat with Claude AI. The service extracts multiple topics from conversations, generates structured markdown issue reports, and maintains a searchable issue database with semantic similarity search.

---

## 🏗️ Project Context

### Repository Structure
```
github.com/melegattip/giia-core-engine/
├── services/
│   ├── auth-service/           # Authentication (95% complete) - Reference for patterns
│   ├── catalog-service/        # Master data (85% complete)
│   ├── ddmrp-engine-service/   # Buffer calculations (65% complete)
│   ├── execution-service/      # Orders (75% complete)
│   ├── analytics-service/      # KPIs (90% complete)
│   ├── ai-intelligence-hub/    # AI notifications (80% complete) - Reference for Claude integration
│   └── feedback-service/       # 🆕 YOUR TARGET SERVICE
├── shared/                     # Shared packages
├── proto/                      # gRPC definitions
└── issue-reports/              # 🆕 Generated markdown files go here
```

### Architecture Standards
- **Clean Architecture**: Domain → Use Cases → Handlers → Repository
- **Module Path**: `github.com/melegattip/giia-core-engine/services/feedback-service`
- **Multi-tenancy**: All data scoped by `organization_id`
- **Testing**: 85%+ code coverage required

---

## 📂 Files to Create

### 1. Service Structure
```
services/feedback-service/
├── cmd/
│   └── api/
│       └── main.go                 # Entry point with DI
├── internal/
│   ├── domain/
│   │   ├── entities/
│   │   │   ├── issue_report.go     # IssueReport entity
│   │   │   ├── chat_message.go     # ChatMessage entity
│   │   │   ├── conversation.go     # Conversation entity
│   │   │   └── image_data.go       # ImageData value object
│   │   ├── repositories/
│   │   │   ├── issue_repository.go # Interface
│   │   │   ├── conversation_repository.go
│   │   │   └── message_repository.go
│   │   └── services/
│   │       └── claude_service.go   # Claude AI interface
│   ├── usecases/
│   │   ├── send_message.go         # Main chat use case
│   │   ├── extract_topics.go       # Multi-topic extraction
│   │   ├── search_issues.go        # Semantic search
│   │   ├── generate_markdown.go    # File generation
│   │   └── delete_user_data.go     # GDPR compliance
│   ├── handlers/
│   │   └── grpc/
│   │       ├── feedback_handler.go # gRPC handlers
│   │       └── server.go           # gRPC server setup
│   ├── repository/
│   │   ├── postgres/
│   │   │   ├── issue_repository.go # GORM implementation
│   │   │   ├── conversation_repository.go
│   │   │   └── message_repository.go
│   │   └── redis/
│   │       └── cache_repository.go # Response caching
│   └── adapters/
│       ├── claude/
│       │   └── client.go           # Anthropic API client
│       ├── auth/
│       │   └── client.go           # Auth service gRPC client
│       └── embeddings/
│           └── client.go           # OpenAI embeddings client
├── migrations/
│   ├── 000001_create_conversations.up.sql
│   ├── 000001_create_conversations.down.sql
│   ├── 000002_create_messages.up.sql
│   ├── 000002_create_messages.down.sql
│   ├── 000003_create_issues.up.sql
│   └── 000003_create_issues.down.sql
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

### 2. Proto Definitions
```
proto/feedback/v1/
├── feedback.proto              # Main service definition
└── messages.proto              # Message types
```

---

## 🔧 Implementation Requirements

### FR-01: Admin Permission Validation
```go
// Validate user has maximum admin role via auth-service gRPC
func (h *FeedbackHandler) validateAdminAccess(ctx context.Context) error {
    // Extract JWT from metadata
    // Call auth-service.ValidateToken()
    // Check for admin role
    // Log unauthorized attempts
}
```

### FR-02: Claude AI Integration with Cost Optimization
- Claude Sonnet 4.5 for complex analysis
- Claude Haiku for simple queries
- Request deduplication (check similar issues first)
- Response caching (1-hour TTL in Redis)
- Daily/monthly token limits per organization
- Automatic model downgrade at 80% of limit

### FR-03: Image Processing with Security
- Validate magic bytes (not just extension)
- Strip EXIF data (remove GPS, device info)
- Prevent image bombs (decompression limits)
- Max 5MB before processing, 2MB after
- Convert all to WebP format
- Embed as base64 in markdown

### FR-04: Multi-Topic Extraction
Use this Claude prompt for topic extraction:
```
You are an expert issue tracker analyzing admin feedback for the GIIA inventory management platform.

TOPIC SEPARATION GUIDELINES:
1. Different system components = separate topics
2. Different issue types (bug vs feature request) = separate topics
3. Same component with different issues = separate topics
4. Related observations about same issue = ONE topic

OUTPUT FORMAT (JSON):
{
  "topics": [{
    "topic_number": 1,
    "title": "Concise title (max 100 chars)",
    "category": "bug|feature_request|improvement|question",
    "priority": "low|medium|high|critical",
    "description": "Detailed description (min 100 chars)",
    "proposed_solution": "Suggested fix (min 50 chars)",
    "reasoning": "Classification rationale",
    "confidence": 0.95
  }]
}
```

### FR-05: Markdown File Generation
- Directory: `issue-reports/` at repository root
- Naming: `{username}Issue{number}{day}-{month}-{year}.md`
- Follow exact template from spec (see spec.md lines 492-618)
- Atomic writes (temp file → rename)
- Include all sections: title, description, context, screenshots, related issues, proposed solution

### FR-06: Navigation Context
Track and include:
- `route`: Current URL path
- `section`: High-level section (catalog, analytics)
- `component`: Active UI component
- `action`: Current user action
- `browserInfo`: Browser, OS, viewport
- `errorState`: Error details if applicable

### FR-07: Issue Database with Vector Search
```sql
-- Enable pgvector extension
CREATE EXTENSION IF NOT EXISTS vector;

-- Issues table with embedding
CREATE TABLE issues (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    organization_id UUID NOT NULL,
    conversation_id UUID NOT NULL,
    title VARCHAR(500) NOT NULL,
    description TEXT NOT NULL,
    category VARCHAR(50) NOT NULL,
    priority VARCHAR(50) NOT NULL,
    status VARCHAR(50) DEFAULT 'open',
    embedding vector(1536),  -- OpenAI text-embedding-3-small
    user_suggested_category VARCHAR(50),
    user_suggested_priority VARCHAR(50),
    navigation_context JSONB,
    similar_issues UUID[],
    markdown_file_path VARCHAR(1000),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    resolved_at TIMESTAMPTZ
);

-- HNSW index for fast similarity search
CREATE INDEX ON issues USING hnsw (embedding vector_cosine_ops);
```

### FR-08: Hybrid Search Algorithm
```go
func (r *IssueRepository) SearchSimilar(ctx context.Context, query string, orgID uuid.UUID) ([]IssueMatch, error) {
    // 1. Generate embedding for query
    // 2. Vector similarity search (top 20, cosine)
    // 3. Full-text search (top 20)
    // 4. Combine with weights: vector 60%, keyword 30%, recency 10%
    // 5. Filter by organization_id
    // 6. Return top 5
}
```

### FR-11: Rate Limiting (Redis)
- Per-user: 30/min, 100/hour, 500/day
- Per-org: 1000/hour, 5000/day
- Sliding window algorithm
- Return HTTP 429 with Retry-After header

### FR-12: Conversation Lifecycle
- States: active, completed, archived, deleted
- Archive after 7 days inactivity
- Delete messages after 90 days (keep metadata)
- Monthly partitioned tables for chat_messages

### FR-14: Monitoring (Prometheus)
Expose metrics:
- `claude_api_requests_total{model, status}`
- `claude_api_latency_seconds{model}`
- `claude_api_tokens_used_total{organization_id, model}`
- `issues_extracted_total{category, priority}`
- `search_latency_seconds{search_type}`

### FR-15: GDPR Compliance
- Display privacy notice before first chat
- DeleteUserData(user_id) endpoint
- ExportUserData(user_id) → JSON endpoint
- Anonymize on user deactivation
- Strip PII from Claude requests

---

## 📊 gRPC API Definition

```protobuf
syntax = "proto3";

package feedback.v1;

service FeedbackService {
  // Chat operations
  rpc SendMessage(SendMessageRequest) returns (SendMessageResponse);
  rpc GetConversationHistory(GetConversationHistoryRequest) returns (GetConversationHistoryResponse);
  
  // Issue operations
  rpc SearchIssues(SearchIssuesRequest) returns (SearchIssuesResponse);
  rpc GetIssueDetails(GetIssueDetailsRequest) returns (GetIssueDetailsResponse);
  rpc UpdateIssueStatus(UpdateIssueStatusRequest) returns (UpdateIssueStatusResponse);
  
  // GDPR compliance
  rpc DeleteUserData(DeleteUserDataRequest) returns (DeleteUserDataResponse);
  rpc ExportUserData(ExportUserDataRequest) returns (ExportUserDataResponse);
}

message SendMessageRequest {
  string content = 1;
  repeated ImageAttachment images = 2;
  NavigationContext navigation_context = 3;
  optional string suggested_category = 4;
  optional string suggested_priority = 5;
}

message SendMessageResponse {
  string message_id = 1;
  string assistant_response = 2;
  repeated ExtractedTopic extracted_topics = 3;
  repeated string generated_files = 4;
}
```

---

## ✅ Success Criteria

### Mandatory (Must Pass)
- [ ] Admin chat interface works with Claude AI
- [ ] Multi-topic extraction achieves 92%+ accuracy
- [ ] Markdown files generated following exact template
- [ ] pgvector semantic search <500ms p95
- [ ] All gRPC endpoints implemented
- [ ] 85%+ test coverage
- [ ] Claude API costs <$50/month for 100 users (caching, tiered models)

### Performance
- [ ] Chat response <1s p95 (with cache)
- [ ] Issue search <500ms p95
- [ ] Markdown generation <3s p95
- [ ] Image processing <1s p95

### Security
- [ ] Admin role validation on every request
- [ ] Image magic byte validation
- [ ] EXIF stripping
- [ ] Rate limiting enforced
- [ ] SQL injection prevented

---

## 🔄 Development Workflow

1. **Start with domain entities** - Define all entities with proper value objects
2. **Create repository interfaces** - Abstract data access
3. **Implement use cases** - Business logic with tests
4. **Add handlers** - gRPC handlers calling use cases
5. **Create repositories** - GORM implementations with multi-tenancy
6. **Add adapters** - Claude, Auth, Embeddings clients
7. **Write migrations** - Database schema
8. **Configure main.go** - Wire dependencies
9. **Add Prometheus metrics** - Observability
10. **Run linting and tests** - `make lint test`

---

## 🚀 Commands to Run

```bash
# Navigate to service
cd services/feedback-service

# Initialize module
go mod init github.com/melegattip/giia-core-engine/services/feedback-service

# Install dependencies
go get gorm.io/gorm
go get gorm.io/driver/postgres
go get google.golang.org/grpc
go get github.com/redis/go-redis/v9
go get github.com/prometheus/client_golang
go get github.com/google/uuid

# Generate proto
protoc --go_out=. --go-grpc_out=. proto/feedback/v1/*.proto

# Run tests
go test ./... -cover

# Build
go build -o bin/feedback-service ./cmd/api

# Run
./bin/feedback-service
```

---

## 📚 Reference Files

Study these existing implementations for patterns:
- `services/auth-service/` - Auth patterns, JWT validation
- `services/ai-intelligence-hub/internal/adapters/claude/` - Claude integration
- `services/analytics-service/internal/usecases/` - Use case patterns with tests
- `shared/` - Common utilities

---

## ⚠️ Important Notes

1. **Never hardcode API keys** - Use environment variables
2. **Always scope by organization_id** - Multi-tenancy is critical
3. **Test first** - Write unit tests before implementation
4. **Log structured** - Use zap or zerolog with JSON format
5. **Follow Clean Architecture** - Dependencies point inward only
6. **Generate embeddings** - Use OpenAI text-embedding-3-small (1536 dims)
7. **Atomic file writes** - Write to temp, then rename
