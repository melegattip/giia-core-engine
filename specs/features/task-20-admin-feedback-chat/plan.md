# Task 20: Admin Feedback Chat with Claude AI - Implementation Plan

**Task ID**: task-20-admin-feedback-chat
**Phase**: 2B - New Microservices (Extension)
**Priority**: P2 (Medium)
**Estimated Duration**: 2-3 weeks
**Dependencies**: Task 17 (AI Agent Service), Task 5 (Auth Service - RBAC)

---

## 1. Technical Context

### Current State
- **AI Agent Service**: Exists but focused on demand forecasting and optimization
- **Auth Service**: Complete with RBAC, permission checking, gRPC endpoints
- **External APIs**: Need to integrate Anthropic Claude API
- **Navigation Tracking**: Not yet implemented (needs to be added)

### Technology Stack
- **Language**: Go 1.23.4
- **Architecture**: Clean Architecture (Domain, Use Cases, Infrastructure) - extends AI Agent Service
- **Database**: PostgreSQL 16 for issue storage and search
- **gRPC**: Protocol Buffers v3 (extends existing ai-agent-service proto)
- **External AI**: Anthropic Claude API (claude-3-opus or claude-3-sonnet)
- **File Storage**: Markdown files in `issue-reports/` directory
- **Testing**: testify for Go tests

### Key Design Decisions
- **Integration Approach**: Extend AI Agent Service rather than create new service
- **Permission Validation**: Reuse auth-service gRPC client for permission checks
- **Navigation Context**: Track via frontend metadata sent with each message
- **Multi-Topic Extraction**: Use Claude's ability to analyze and separate topics
- **File Generation**: Atomic writes to ensure consistency
- **Search**: Use PostgreSQL full-text search + semantic similarity (embeddings optional)

---

## 2. Project Structure

### Files to Create/Modify

```
giia-core-engine/
├── services/ai-agent-service/
│   ├── api/proto/ai_agent/v1/
│   │   ├── ai_agent.proto                    [MODIFY] - Add chat endpoints
│   │   ├── ai_agent.pb.go                    [GENERATED]
│   │   └── ai_agent_grpc.pb.go              [GENERATED]
│   │
│   ├── internal/
│   │   ├── core/
│   │   │   ├── domain/
│   │   │   │   ├── issue_report.go           [NEW]
│   │   │   │   ├── chat_message.go           [NEW]
│   │   │   │   ├── conversation.go           [NEW]
│   │   │   │   └── image_data.go             [NEW]
│   │   │   │
│   │   │   ├── providers/
│   │   │   │   ├── claude_client.go          [NEW]
│   │   │   │   ├── auth_client.go            [NEW] - Reuse auth-service client
│   │   │   │   ├── navigation_tracker.go    [NEW]
│   │   │   │   └── issue_repository.go      [NEW]
│   │   │   │
│   │   │   └── usecases/
│   │   │       ├── chat/
│   │   │       │   ├── send_message.go       [NEW]
│   │   │       │   ├── get_conversation.go   [NEW]
│   │   │       │   └── validate_admin.go    [NEW]
│   │   │       │
│   │   │       ├── issue/
│   │   │       │   ├── extract_topics.go     [NEW]
│   │   │       │   ├── generate_markdown.go  [NEW]
│   │   │       │   ├── search_issues.go     [NEW]
│   │   │       │   └── classify_issue.go    [NEW]
│   │   │       │
│   │   │       └── context/
│   │   │           └── collect_navigation.go [NEW]
│   │   │
│   │   └── infrastructure/
│   │       ├── repositories/
│   │       │   ├── issue_repository_impl.go  [NEW]
│   │       │   ├── conversation_repository_impl.go [NEW]
│   │       │   └── message_repository_impl.go [NEW]
│   │       │
│   │       ├── adapters/
│   │       │   ├── claude_api_client.go      [NEW]
│   │       │   ├── auth_grpc_client.go      [NEW]
│   │       │   └── markdown_generator.go    [NEW]
│   │       │
│   │       └── grpc/
│   │           └── server/
│   │               └── chat_service.go       [NEW] - gRPC handlers
│   │
│   ├── migrations/
│   │   ├── 000005_create_issue_reports.up.sql    [NEW]
│   │   ├── 000006_create_conversations.up.sql    [NEW]
│   │   ├── 000007_create_chat_messages.up.sql    [NEW]
│   │   └── 000008_create_image_data.up.sql       [NEW]
│   │
│   └── issue-reports/                      [NEW DIRECTORY]
│       └── .gitkeep                        [NEW]
│
└── issue-reports/                          [NEW DIRECTORY AT ROOT]
    └── .gitkeep                            [NEW]
```

---

## 3. Implementation Steps

### Phase 1: Database Schema & Domain Entities (Week 1, Days 1-2)

#### Migrations

**File**: `services/ai-agent-service/migrations/000005_create_issue_reports.up.sql`

```sql
-- Issue Reports table
CREATE TABLE IF NOT EXISTS issue_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    organization_id UUID NOT NULL,
    conversation_id UUID NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    category VARCHAR(50) NOT NULL,
    priority VARCHAR(20) NOT NULL,
    user_suggested_category VARCHAR(50),
    user_suggested_priority VARCHAR(20),
    status VARCHAR(20) NOT NULL DEFAULT 'open',
    markdown_file_path VARCHAR(500) NOT NULL,
    navigation_context JSONB,
    similar_issues UUID[],
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMP,
    CONSTRAINT chk_category CHECK (category IN ('bug', 'feature_request', 'improvement', 'question')),
    CONSTRAINT chk_priority CHECK (priority IN ('low', 'medium', 'high', 'critical')),
    CONSTRAINT chk_status CHECK (status IN ('open', 'in_progress', 'resolved', 'closed'))
);

CREATE INDEX idx_issue_reports_user ON issue_reports(user_id, organization_id);
CREATE INDEX idx_issue_reports_category ON issue_reports(category);
CREATE INDEX idx_issue_reports_priority ON issue_reports(priority);
CREATE INDEX idx_issue_reports_status ON issue_reports(status);
CREATE INDEX idx_issue_reports_created ON issue_reports(created_at DESC);
CREATE INDEX idx_issue_reports_search ON issue_reports USING gin(to_tsvector('english', title || ' ' || description));
```

**File**: `services/ai-agent-service/migrations/000006_create_conversations.up.sql`

```sql
-- Conversations table
CREATE TABLE IF NOT EXISTS conversations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    organization_id UUID NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    last_message_at TIMESTAMP,
    CONSTRAINT chk_conversation_status CHECK (status IN ('active', 'completed', 'archived'))
);

CREATE INDEX idx_conversations_user ON conversations(user_id, organization_id);
CREATE INDEX idx_conversations_status ON conversations(status);
```

**File**: `services/ai-agent-service/migrations/000007_create_chat_messages.up.sql`

```sql
-- Chat Messages table
CREATE TABLE IF NOT EXISTS chat_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    role VARCHAR(20) NOT NULL,
    content TEXT NOT NULL,
    navigation_context JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_message_role CHECK (role IN ('user', 'assistant'))
);

CREATE INDEX idx_chat_messages_conversation ON chat_messages(conversation_id, created_at);
CREATE INDEX idx_chat_messages_user ON chat_messages(user_id);
```

**File**: `services/ai-agent-service/migrations/000008_create_image_data.up.sql`

```sql
-- Image Data table
CREATE TABLE IF NOT EXISTS image_data (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    message_id UUID NOT NULL REFERENCES chat_messages(id) ON DELETE CASCADE,
    filename VARCHAR(255) NOT NULL,
    content_type VARCHAR(100) NOT NULL,
    base64_data TEXT NOT NULL,
    size_bytes INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_image_size CHECK (size_bytes <= 5242880) -- 5MB max
);

CREATE INDEX idx_image_data_message ON image_data(message_id);
```

#### Domain Entities

**File**: `services/ai-agent-service/internal/core/domain/issue_report.go`

```go
package domain

import (
	"time"
	"github.com/google/uuid"
)

type IssueReport struct {
	ID                   uuid.UUID
	UserID               uuid.UUID
	OrganizationID       uuid.UUID
	ConversationID       uuid.UUID
	Title                string
	Description          string
	Category             IssueCategory
	Priority             IssuePriority
	UserSuggestedCategory *string
	UserSuggestedPriority  *string
	Status               IssueStatus
	MarkdownFilePath     string
	NavigationContext    map[string]interface{}
	SimilarIssues        []uuid.UUID
	CreatedAt            time.Time
	UpdatedAt            time.Time
	ResolvedAt           *time.Time
}

type IssueCategory string

const (
	IssueCategoryBug            IssueCategory = "bug"
	IssueCategoryFeatureRequest IssueCategory = "feature_request"
	IssueCategoryImprovement    IssueCategory = "improvement"
	IssueCategoryQuestion       IssueCategory = "question"
)

type IssuePriority string

const (
	IssuePriorityLow      IssuePriority = "low"
	IssuePriorityMedium   IssuePriority = "medium"
	IssuePriorityHigh     IssuePriority = "high"
	IssuePriorityCritical IssuePriority = "critical"
)

type IssueStatus string

const (
	IssueStatusOpen       IssueStatus = "open"
	IssueStatusInProgress IssueStatus = "in_progress"
	IssueStatusResolved   IssueStatus = "resolved"
	IssueStatusClosed     IssueStatus = "closed"
)
```

**File**: `services/ai-agent-service/internal/core/domain/chat_message.go`

```go
package domain

import (
	"time"
	"github.com/google/uuid"
)

type ChatMessage struct {
	ID                uuid.UUID
	ConversationID    uuid.UUID
	UserID            uuid.UUID
	Role              MessageRole
	Content           string
	Images            []ImageData
	NavigationContext map[string]interface{}
	CreatedAt         time.Time
}

type MessageRole string

const (
	MessageRoleUser      MessageRole = "user"
	MessageRoleAssistant MessageRole = "assistant"
)
```

**File**: `services/ai-agent-service/internal/core/domain/conversation.go`

```go
package domain

import (
	"time"
	"github.com/google/uuid"
)

type Conversation struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	OrganizationID uuid.UUID
	Status        ConversationStatus
	CreatedAt     time.Time
	UpdatedAt     time.Time
	LastMessageAt *time.Time
}

type ConversationStatus string

const (
	ConversationStatusActive    ConversationStatus = "active"
	ConversationStatusCompleted ConversationStatus = "completed"
	ConversationStatusArchived  ConversationStatus = "archived"
)
```

**File**: `services/ai-agent-service/internal/core/domain/image_data.go`

```go
package domain

import (
	"time"
	"github.com/google/uuid"
)

type ImageData struct {
	ID          uuid.UUID
	MessageID   uuid.UUID
	Filename    string
	ContentType string
	Base64Data  string
	SizeBytes   int
	CreatedAt   time.Time
}
```

---

### Phase 2: Claude API Integration (Week 1, Days 3-4)

#### Claude Client Provider

**File**: `services/ai-agent-service/internal/core/providers/claude_client.go`

```go
package providers

import (
	"context"
	"giia-core-engine/services/ai-agent-service/internal/core/domain"
)

type ClaudeClient interface {
	SendMessage(ctx context.Context, messages []domain.ChatMessage, systemPrompt string) (string, error)
	ExtractTopics(ctx context.Context, conversation []domain.ChatMessage) ([]domain.IssueReport, error)
	ClassifyIssue(ctx context.Context, description string, userCategory *string, userPriority *string) (domain.IssueCategory, domain.IssuePriority, string, error)
	SearchSimilarIssues(ctx context.Context, description string, existingIssues []domain.IssueReport) ([]uuid.UUID, error)
	GenerateMarkdown(ctx context.Context, issue domain.IssueReport, conversation []domain.ChatMessage) (string, error)
}
```

#### Claude API Adapter

**File**: `services/ai-agent-service/internal/infrastructure/adapters/claude_api_client.go`

```go
package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"bytes"
	
	"giia-core-engine/services/ai-agent-service/internal/core/domain"
	"giia-core-engine/services/ai-agent-service/internal/core/providers"
)

type ClaudeAPIClient struct {
	apiKey     string
	baseURL    string
	model      string
	httpClient *http.Client
}

func NewClaudeAPIClient(apiKey, baseURL, model string) providers.ClaudeClient {
	return &ClaudeAPIClient{
		apiKey:     apiKey,
		baseURL:    baseURL,
		model:      model,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *ClaudeAPIClient) SendMessage(ctx context.Context, messages []domain.ChatMessage, systemPrompt string) (string, error) {
	// Convert domain messages to Claude API format
	apiMessages := make([]map[string]interface{}, 0)
	for _, msg := range messages {
		apiMsg := map[string]interface{}{
			"role": msg.Role,
			"content": msg.Content,
		}
		// Add images if present
		if len(msg.Images) > 0 {
			// Handle image content
		}
		apiMessages = append(apiMessages, apiMsg)
	}
	
	payload := map[string]interface{}{
		"model":       c.model,
		"max_tokens":  4096,
		"messages":    apiMessages,
		"system":      systemPrompt,
	}
	
	// Make HTTP request to Claude API
	// Parse response and return content
	// Implementation details...
}
```

---

### Phase 3: Permission Validation & Auth Integration (Week 1, Day 5)

#### Auth Client Provider

**File**: `services/ai-agent-service/internal/core/providers/auth_client.go`

```go
package providers

import (
	"context"
	"github.com/google/uuid"
)

type AuthClient interface {
	ValidateAdminPermission(ctx context.Context, userID, organizationID uuid.UUID) (bool, error)
	GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]string, error)
}
```

#### Auth gRPC Adapter

**File**: `services/ai-agent-service/internal/infrastructure/adapters/auth_grpc_client.go`

```go
package adapters

import (
	"context"
	"github.com/google/uuid"
	"giia-core-engine/api/proto/auth/v1"
	"giia-core-engine/services/ai-agent-service/internal/core/providers"
	"google.golang.org/grpc"
)

type AuthGRPCClient struct {
	client authv1.AuthServiceClient
}

func NewAuthGRPCClient(conn *grpc.ClientConn) providers.AuthClient {
	return &AuthGRPCClient{
		client: authv1.NewAuthServiceClient(conn),
	}
}

func (c *AuthGRPCClient) ValidateAdminPermission(ctx context.Context, userID, organizationID uuid.UUID) (bool, error) {
	// Call auth-service gRPC CheckPermission with admin permission
	// Return true if user has maximum admin role
}
```

---

### Phase 4: Use Cases Implementation (Week 2, Days 1-3)

#### Validate Admin Use Case

**File**: `services/ai-agent-service/internal/core/usecases/chat/validate_admin.go`

```go
package chat

import (
	"context"
	"github.com/google/uuid"
	"giia-core-engine/services/ai-agent-service/internal/core/providers"
)

type ValidateAdminUseCase struct {
	authClient providers.AuthClient
}

func (uc *ValidateAdminUseCase) Execute(ctx context.Context, userID, organizationID uuid.UUID) error {
	hasPermission, err := uc.authClient.ValidateAdminPermission(ctx, userID, organizationID)
	if err != nil {
		return err
	}
	if !hasPermission {
		return errors.NewForbidden("admin permission required")
	}
	return nil
}
```

#### Send Message Use Case

**File**: `services/ai-agent-service/internal/core/usecases/chat/send_message.go`

```go
package chat

import (
	"context"
	"github.com/google/uuid"
	"giia-core-engine/services/ai-agent-service/internal/core/domain"
	"giia-core-engine/services/ai-agent-service/internal/core/providers"
)

type SendMessageUseCase struct {
	conversationRepo providers.ConversationRepository
	messageRepo      providers.MessageRepository
	claudeClient     providers.ClaudeClient
	contextCollector providers.NavigationTracker
}

func (uc *SendMessageUseCase) Execute(ctx context.Context, userID, organizationID uuid.UUID, content string, images []domain.ImageData, navContext map[string]interface{}) (*domain.ChatMessage, *domain.ChatMessage, error) {
	// 1. Get or create active conversation
	// 2. Save user message
	// 3. Collect navigation context
	// 4. Get conversation history
	// 5. Call Claude API
	// 6. Save assistant response
	// 7. Return both messages
}
```

#### Extract Topics Use Case

**File**: `services/ai-agent-service/internal/core/usecases/issue/extract_topics.go`

```go
package issue

import (
	"context"
	"giia-core-engine/services/ai-agent-service/internal/core/domain"
	"giia-core-engine/services/ai-agent-service/internal/core/providers"
)

type ExtractTopicsUseCase struct {
	claudeClient     providers.ClaudeClient
	issueRepo        providers.IssueRepository
	markdownGen      providers.MarkdownGenerator
}

func (uc *ExtractTopicsUseCase) Execute(ctx context.Context, conversationID uuid.UUID, conversation []domain.ChatMessage) ([]domain.IssueReport, error) {
	// 1. Call Claude to extract topics
	// 2. For each topic, create IssueReport
	// 3. Search for similar issues
	// 4. Generate markdown for each issue
	// 5. Save to database
	// 6. Write markdown files
}
```

#### Generate Markdown Use Case

**File**: `services/ai-agent-service/internal/core/usecases/issue/generate_markdown.go`

```go
package issue

import (
	"context"
	"giia-core-engine/services/ai-agent-service/internal/core/domain"
	"giia-core-engine/services/ai-agent-service/internal/core/providers"
)

type GenerateMarkdownUseCase struct {
	claudeClient providers.ClaudeClient
	markdownGen  providers.MarkdownGenerator
}

func (uc *GenerateMarkdownUseCase) Execute(ctx context.Context, issue domain.IssueReport, conversation []domain.ChatMessage) (string, error) {
	// 1. Call Claude to generate structured markdown
	// 2. Use markdown generator to format according to spec template
	// 3. Return markdown content
}
```

---

### Phase 5: gRPC API Implementation (Week 2, Days 4-5)

#### Protocol Buffer Definitions

**File**: `services/ai-agent-service/api/proto/ai_agent/v1/ai_agent.proto` (modify existing)

```protobuf
// Add to existing ai_agent.proto

service AIChatService {
  rpc SendMessage(SendMessageRequest) returns (SendMessageResponse);
  rpc GetConversation(GetConversationRequest) returns (GetConversationResponse);
  rpc SearchIssues(SearchIssuesRequest) returns (SearchIssuesResponse);
}

message SendMessageRequest {
  string conversation_id = 1; // Optional, creates new if empty
  string content = 2;
  repeated ImageAttachment images = 3;
  NavigationContext navigation_context = 4;
  string suggested_category = 5; // Optional
  string suggested_priority = 6; // Optional
}

message ImageAttachment {
  string filename = 1;
  string content_type = 2;
  string base64_data = 3;
}

message NavigationContext {
  string route = 1;
  string section = 2;
  string component = 3;
  map<string, string> metadata = 4;
}

message SendMessageResponse {
  string conversation_id = 1;
  ChatMessage user_message = 2;
  ChatMessage assistant_message = 3;
  repeated string extracted_issue_ids = 4; // Issues extracted from this message
}

message ChatMessage {
  string id = 1;
  string role = 2; // "user" or "assistant"
  string content = 3;
  repeated ImageAttachment images = 4;
  string created_at = 5;
}

message GetConversationRequest {
  string conversation_id = 1;
}

message GetConversationResponse {
  string conversation_id = 1;
  repeated ChatMessage messages = 2;
  string status = 3;
}

message SearchIssuesRequest {
  string query = 1;
  string category = 2; // Optional filter
  string priority = 3; // Optional filter
  int32 limit = 4;
}

message SearchIssuesResponse {
  repeated IssueSummary issues = 1;
}

message IssueSummary {
  string id = 1;
  string title = 2;
  string category = 3;
  string priority = 4;
  string status = 5;
  string created_at = 6;
  string markdown_file_path = 7;
}
```

#### gRPC Service Implementation

**File**: `services/ai-agent-service/internal/infrastructure/grpc/server/chat_service.go`

```go
package server

import (
	"context"
	"giia-core-engine/api/proto/ai_agent/v1"
	"giia-core-engine/services/ai-agent-service/internal/core/usecases/chat"
	"giia-core-engine/services/ai-agent-service/internal/core/usecases/issue"
)

type ChatService struct {
	ai_agentv1.UnimplementedAIChatServiceServer
	sendMessageUseCase *chat.SendMessageUseCase
	// ... other use cases
}

func (s *ChatService) SendMessage(ctx context.Context, req *ai_agentv1.SendMessageRequest) (*ai_agentv1.SendMessageResponse, error) {
	// 1. Extract user ID from context (from auth interceptor)
	// 2. Validate admin permission
	// 3. Call send message use case
	// 4. Return response
}
```

---

### Phase 6: Testing & Integration (Week 3)

#### Unit Tests
- Test all use cases with mocks
- Test Claude client adapter with test responses
- Test markdown generation
- Test permission validation

#### Integration Tests
- Test full chat flow with real database
- Test issue extraction and markdown file creation
- Test search functionality
- Test navigation context collection

---

## 4. Success Criteria

### Mandatory
- ✅ Admin chat interface with Claude AI
- ✅ Multi-topic extraction working
- ✅ Markdown files generated in `issue-reports/`
- ✅ Navigation context tracking
- ✅ Issue database with search
- ✅ Category/priority classification
- ✅ gRPC API fully implemented
- ✅ Permission validation working
- ✅ Image attachment support
- ✅ 80%+ test coverage

---

## 5. Dependencies

- **Task 17**: AI Agent Service (base service)
- **Task 5**: Auth Service with RBAC
- **External**: Anthropic Claude API
- **Shared packages**: pkg/events, pkg/database, pkg/logger, pkg/errors

---

## 6. Risk Mitigation

| Risk | Mitigation |
|------|------------|
| Claude API costs | Rate limiting, caching, cost monitoring |
| Multi-topic extraction accuracy | Manual validation, iterative prompt improvement |
| Navigation context overhead | Optimize collection, cache where possible |
| Permission validation | Comprehensive testing, reuse existing auth-service |

---

**Document Version**: 1.0
**Last Updated**: 2025-12-16
**Status**: Ready for Implementation

