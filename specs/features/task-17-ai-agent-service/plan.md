# Task 17: AI Intelligence Hub - Implementation Plan

**Task ID**: task-17-ai-intelligence-hub
**Phase**: 2B - AI-Powered Intelligence
**Priority**: P1 (HIGH - Core Value Differentiator)
**Estimated Duration**: 6-8 weeks
**Dependencies**: Task 08 (NATS JetStream - ✅ Complete)

---

## 1. Technical Context

### Current State
- **NATS JetStream**: ✅ Implemented and operational (Task 08 complete)
- **Event Publishing**: ✅ Auth service publishing user events
- **Event Infrastructure**: ✅ Streams configured, retry logic in place
- **AI Integration**: ❌ Not yet implemented
- **RAG System**: ❌ Not yet implemented
- **Notification System**: ❌ Not yet implemented

### Technology Stack
- **Language**: Go 1.23.4
- **Architecture**: Clean Architecture + Event-Driven
- **Database**: PostgreSQL 16 for notifications, preferences
- **Event Bus**: NATS JetStream (already configured)
- **AI**: Claude API (Anthropic)
- **RAG**: ChromaDB for vector embeddings
- **Notifications**:
  - In-app: WebSocket (Gorilla WebSocket)
  - Email: SendGrid or AWS SES
  - Slack: Slack Web API
  - SMS: Twilio (critical alerts only)
- **Testing**: testify, gomock

### Key Design Decisions
1. **Event-Driven**: Subscribe to all NATS streams for comprehensive monitoring
2. **Microservice**: Standalone service subscribed to all events
3. **RAG Pattern**: Combine real-time data + DDMRP knowledge for analysis
4. **Asynchronous**: Non-blocking event processing with goroutines
5. **Multi-tenancy**: organization_id filtering throughout
6. **Caching**: Redis for AI response caching (reduce costs)
7. **User Preferences**: Flexible notification channel configuration

---

## 2. Project Structure

### Files to Create

```
giia-core-engine/
├── services/ai-intelligence-hub/                [NEW SERVICE]
│   │
│   ├── api/proto/intelligence/v1/
│   │   ├── intelligence.proto                   [NEW] gRPC API definition
│   │   ├── intelligence.pb.go                   [GENERATED]
│   │   └── intelligence_grpc.pb.go              [GENERATED]
│   │
│   ├── cmd/
│   │   └── api/
│   │       └── main.go                          [NEW] Service entry point
│   │
│   ├── internal/
│   │   ├── core/
│   │   │   ├── domain/
│   │   │   │   ├── notification.go              [NEW] Core notification entity
│   │   │   │   ├── analysis_context.go          [NEW] AI analysis context
│   │   │   │   ├── user_preferences.go          [NEW] Notification preferences
│   │   │   │   ├── impact_assessment.go         [NEW] Impact calculation
│   │   │   │   └── recommendation.go            [NEW] AI recommendation
│   │   │   │
│   │   │   ├── providers/
│   │   │   │   ├── event_subscriber.go          [NEW] Interface for NATS
│   │   │   │   ├── ai_analyzer.go               [NEW] Interface for Claude
│   │   │   │   ├── rag_knowledge.go             [NEW] Interface for RAG
│   │   │   │   ├── notification_sender.go       [NEW] Interface for delivery
│   │   │   │   ├── service_client.go            [NEW] Interface for GIIA services
│   │   │   │   └── mocks.go                     [NEW] Test mocks
│   │   │   │
│   │   │   └── usecases/
│   │   │       ├── event_processing/
│   │   │       │   ├── process_buffer_event.go  [NEW] Handle buffer events
│   │   │       │   ├── process_execution_event.go [NEW] Handle execution events
│   │   │       │   ├── process_user_event.go    [NEW] Handle user events
│   │   │       │   └── detect_patterns.go       [NEW] Pattern detection
│   │   │       │
│   │   │       ├── analysis/
│   │   │       │   ├── analyze_stockout_risk.go [NEW] Stockout analysis
│   │   │       │   ├── analyze_cost_opportunity.go [NEW] Cost analysis
│   │   │       │   ├── analyze_seasonal_pattern.go [NEW] Seasonality
│   │   │       │   └── generate_daily_digest.go [NEW] Daily summary
│   │   │       │
│   │   │       └── notification/
│   │   │           ├── create_notification.go    [NEW] Generate notification
│   │   │           ├── route_notification.go     [NEW] Channel routing
│   │   │           └── manage_preferences.go     [NEW] User preferences
│   │   │
│   │   └── infrastructure/
│   │       ├── adapters/
│   │       │   ├── nats/
│   │       │   │   ├── event_subscriber.go      [NEW] NATS subscriber
│   │       │   │   └── event_processor.go       [NEW] Event handler
│   │       │   │
│   │       │   ├── ai/
│   │       │   │   ├── claude_client.go         [NEW] Claude API client
│   │       │   │   └── response_cache.go        [NEW] Redis cache
│   │       │   │
│   │       │   ├── rag/
│   │       │   │   ├── chromadb_client.go       [NEW] ChromaDB client
│   │       │   │   └── knowledge_retriever.go   [NEW] Knowledge search
│   │       │   │
│   │       │   ├── notifications/
│   │       │   │   ├── email_sender.go          [NEW] Email via SendGrid
│   │       │   │   ├── websocket_push.go        [NEW] In-app push
│   │       │   │   ├── slack_sender.go          [NEW] Slack integration
│   │       │   │   └── sms_sender.go            [NEW] Twilio SMS
│   │       │   │
│   │       │   └── services/
│   │       │       ├── ddmrp_client.go          [NEW] DDMRP service gRPC
│   │       │       ├── catalog_client.go        [NEW] Catalog service gRPC
│   │       │       ├── execution_client.go      [NEW] Execution service gRPC
│   │       │       └── analytics_client.go      [NEW] Analytics service gRPC
│   │       │
│   │       ├── repositories/
│   │       │   ├── notification_repository.go   [NEW] Notifications CRUD
│   │       │   └── preferences_repository.go    [NEW] Preferences CRUD
│   │       │
│   │       └── entrypoints/
│   │           ├── grpc/
│   │           │   └── intelligence_server.go   [NEW] gRPC server
│   │           │
│   │           └── http/
│   │               ├── handlers/
│   │               │   └── notification_handler.go [NEW] REST API
│   │               │
│   │               └── middleware/
│   │                   └── websocket.go          [NEW] WebSocket handler
│   │
│   ├── migrations/
│   │   ├── 000001_create_notifications.up.sql   [NEW]
│   │   ├── 000002_create_user_preferences.up.sql [NEW]
│   │   └── 000003_create_notification_history.up.sql [NEW]
│   │
│   ├── knowledge_base/                           [NEW] RAG documents
│   │   ├── ddmrp_methodology/
│   │   │   ├── buffer_calculation.md            [NEW]
│   │   │   ├── buffer_zones.md                  [NEW]
│   │   │   ├── demand_planning.md               [NEW]
│   │   │   └── execution_best_practices.md      [NEW]
│   │   │
│   │   └── giia_platform/
│   │       ├── platform_overview.md             [NEW]
│   │       └── troubleshooting.md               [NEW]
│   │
│   ├── .env.example                             [NEW]
│   ├── Dockerfile                               [NEW]
│   ├── Makefile                                 [NEW]
│   └── README.md                                [NEW]
│
└── docker-compose.yml                            [UPDATE] Add intelligence hub
```

---

## 3. Implementation Phases

### Phase 1: Foundation & Database (Week 1)

#### T001: Database Schema Design
**File**: `services/ai-intelligence-hub/migrations/000001_create_notifications.up.sql`

```sql
-- AI Notifications table
CREATE TABLE IF NOT EXISTS ai_notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL,
    user_id UUID NOT NULL,

    type VARCHAR(20) NOT NULL,  -- alert, warning, info, suggestion, insight, digest
    priority VARCHAR(20) NOT NULL,  -- critical, high, medium, low

    title VARCHAR(255) NOT NULL,
    summary TEXT NOT NULL,
    full_analysis TEXT,
    reasoning TEXT,

    -- Impact assessment
    risk_level VARCHAR(20),
    revenue_impact DECIMAL(15,2),
    cost_impact DECIMAL(15,2),
    time_to_impact_seconds INTEGER,
    affected_orders INTEGER,
    affected_products INTEGER,

    -- Source tracking
    source_events JSONB,  -- Array of event IDs
    related_entities JSONB,  -- product_ids, supplier_ids, etc.

    -- Status
    status VARCHAR(20) NOT NULL DEFAULT 'unread',  -- unread, read, acted_upon, dismissed

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    read_at TIMESTAMP,
    acted_at TIMESTAMP,
    dismissed_at TIMESTAMP,

    CONSTRAINT chk_notification_type CHECK (type IN (
        'alert', 'warning', 'info', 'suggestion', 'insight', 'digest'
    )),
    CONSTRAINT chk_notification_priority CHECK (priority IN (
        'critical', 'high', 'medium', 'low'
    )),
    CONSTRAINT chk_notification_status CHECK (status IN (
        'unread', 'read', 'acted_upon', 'dismissed'
    ))
);

CREATE INDEX idx_notifications_user ON ai_notifications(user_id, organization_id);
CREATE INDEX idx_notifications_status ON ai_notifications(status, created_at);
CREATE INDEX idx_notifications_priority ON ai_notifications(priority, created_at);
CREATE INDEX idx_notifications_type ON ai_notifications(type);

-- Recommendations sub-table
CREATE TABLE IF NOT EXISTS ai_recommendations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    notification_id UUID NOT NULL REFERENCES ai_notifications(id) ON DELETE CASCADE,

    action TEXT NOT NULL,
    reasoning TEXT NOT NULL,
    expected_outcome TEXT,
    effort VARCHAR(20),  -- low, medium, high
    impact VARCHAR(20),  -- low, medium, high
    action_url TEXT,

    priority_order INTEGER NOT NULL,

    CONSTRAINT chk_recommendation_effort CHECK (effort IN ('low', 'medium', 'high')),
    CONSTRAINT chk_recommendation_impact CHECK (impact IN ('low', 'medium', 'high'))
);

CREATE INDEX idx_recommendations_notification ON ai_recommendations(notification_id);
```

**File**: `services/ai-intelligence-hub/migrations/000002_create_user_preferences.up.sql`

```sql
-- User Notification Preferences
CREATE TABLE IF NOT EXISTS user_notification_preferences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    organization_id UUID NOT NULL,

    -- Channel preferences
    enable_in_app BOOLEAN NOT NULL DEFAULT true,
    enable_email BOOLEAN NOT NULL DEFAULT true,
    enable_sms BOOLEAN NOT NULL DEFAULT false,
    enable_slack BOOLEAN NOT NULL DEFAULT false,
    slack_webhook_url TEXT,

    -- Priority thresholds
    in_app_min_priority VARCHAR(20) NOT NULL DEFAULT 'low',
    email_min_priority VARCHAR(20) NOT NULL DEFAULT 'medium',
    sms_min_priority VARCHAR(20) NOT NULL DEFAULT 'critical',

    -- Timing preferences
    digest_time TIME NOT NULL DEFAULT '06:00:00',
    quiet_hours_start TIME,
    quiet_hours_end TIME,
    timezone VARCHAR(50) NOT NULL DEFAULT 'UTC',

    -- Frequency limits
    max_alerts_per_hour INTEGER NOT NULL DEFAULT 10,
    max_emails_per_day INTEGER NOT NULL DEFAULT 50,

    -- Content preferences
    detail_level VARCHAR(20) NOT NULL DEFAULT 'detailed',  -- brief, detailed, comprehensive
    include_charts BOOLEAN NOT NULL DEFAULT true,
    include_historical BOOLEAN NOT NULL DEFAULT true,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_user_preferences UNIQUE (user_id, organization_id),
    CONSTRAINT chk_detail_level CHECK (detail_level IN ('brief', 'detailed', 'comprehensive'))
);

CREATE INDEX idx_user_prefs_user ON user_notification_preferences(user_id, organization_id);
```

#### T002: Domain Entities

**File**: `services/ai-intelligence-hub/internal/core/domain/notification.go`

```go
package domain

import (
    "time"
    "github.com/google/uuid"
)

type NotificationType string

const (
    NotificationTypeAlert      NotificationType = "alert"
    NotificationTypeWarning    NotificationType = "warning"
    NotificationTypeInfo       NotificationType = "info"
    NotificationTypeSuggestion NotificationType = "suggestion"
    NotificationTypeInsight    NotificationType = "insight"
    NotificationTypeDigest     NotificationType = "digest"
)

type NotificationPriority string

const (
    NotificationPriorityCritical NotificationPriority = "critical"
    NotificationPriorityHigh     NotificationPriority = "high"
    NotificationPriorityMedium   NotificationPriority = "medium"
    NotificationPriorityLow      NotificationPriority = "low"
)

type NotificationStatus string

const (
    NotificationStatusUnread     NotificationStatus = "unread"
    NotificationStatusRead       NotificationStatus = "read"
    NotificationStatusActedUpon  NotificationStatus = "acted_upon"
    NotificationStatusDismissed  NotificationStatus = "dismissed"
)

type AINotification struct {
    ID              uuid.UUID
    OrganizationID  uuid.UUID
    UserID          uuid.UUID

    Type            NotificationType
    Priority        NotificationPriority

    Title           string
    Summary         string
    FullAnalysis    string
    Reasoning       string

    Impact          ImpactAssessment
    Recommendations []Recommendation

    SourceEvents    []string  // Event IDs
    RelatedEntities map[string][]string  // "product_ids": ["uuid1", "uuid2"]

    Status          NotificationStatus

    CreatedAt       time.Time
    ReadAt          *time.Time
    ActedAt         *time.Time
    DismissedAt     *time.Time
}

type ImpactAssessment struct {
    RiskLevel        string  // low, medium, high, critical
    RevenueImpact    float64
    CostImpact       float64
    TimeToImpact     *time.Duration
    AffectedOrders   int
    AffectedProducts int
}

type Recommendation struct {
    Action          string
    Reasoning       string
    ExpectedOutcome string
    Effort          string  // low, medium, high
    Impact          string  // low, medium, high
    ActionURL       string
    PriorityOrder   int
}

func (n *AINotification) MarkAsRead() {
    now := time.Now()
    n.ReadAt = &now
    n.Status = NotificationStatusRead
}

func (n *AINotification) MarkAsActedUpon() {
    now := time.Now()
    n.ActedAt = &now
    n.Status = NotificationStatusActedUpon
}

func (n *AINotification) Dismiss() {
    now := time.Now()
    n.DismissedAt = &now
    n.Status = NotificationStatusDismissed
}
```

#### T003: Service Scaffold

**File**: `services/ai-intelligence-hub/cmd/api/main.go`

```go
package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"

    "github.com/giia/giia-core-engine/pkg/config"
    "github.com/giia/giia-core-engine/pkg/database"
    "github.com/giia/giia-core-engine/pkg/events"
    "github.com/giia/giia-core-engine/pkg/logger"
)

func main() {
    ctx := context.Background()

    // Load configuration
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    // Initialize logger
    logger := logger.NewLogger(cfg.LogLevel)

    // Initialize database
    db, err := database.NewPostgresConnection(cfg.DatabaseURL)
    if err != nil {
        logger.Fatal(ctx, err, "Failed to connect to database", nil)
    }
    defer db.Close()

    // Initialize NATS connection
    natsConn, err := events.Connect(cfg.NATSServers)
    if err != nil {
        logger.Fatal(ctx, err, "Failed to connect to NATS", nil)
    }
    defer natsConn.Close()

    // Initialize event subscriber
    subscriber, err := events.NewSubscriber(natsConn)
    if err != nil {
        logger.Fatal(ctx, err, "Failed to create subscriber", nil)
    }

    // TODO: Initialize AI Intelligence Hub service

    logger.Info(ctx, "AI Intelligence Hub started successfully", nil)

    // Graceful shutdown
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
    <-sigChan

    logger.Info(ctx, "Shutting down gracefully...", nil)
}
```

---

### Phase 2: Event Processing (Week 2)

#### T004: NATS Event Subscriber

**File**: `services/ai-intelligence-hub/internal/infrastructure/adapters/nats/event_subscriber.go`

```go
package nats

import (
    "context"
    "github.com/giia/giia-core-engine/pkg/events"
    "github.com/giia/giia-core-engine/pkg/logger"
)

type EventSubscriber struct {
    subscriber      *events.Subscriber
    eventProcessor  EventProcessor
    logger          logger.Logger
}

type EventProcessor interface {
    ProcessEvent(ctx context.Context, event *events.Event) error
}

func NewEventSubscriber(
    subscriber *events.Subscriber,
    processor EventProcessor,
    logger logger.Logger,
) *EventSubscriber {
    return &EventSubscriber{
        subscriber:     subscriber,
        eventProcessor: processor,
        logger:         logger,
    }
}

func (es *EventSubscriber) Start(ctx context.Context) error {
    // Subscribe to all event subjects
    subjects := []string{
        "auth.>",
        "catalog.>",
        "ddmrp.>",
        "execution.>",
        "analytics.>",
    }

    for _, subject := range subjects {
        if err := es.subscribeToSubject(ctx, subject); err != nil {
            return err
        }
    }

    es.logger.Info(ctx, "Event subscriber started", logger.Tags{
        "subjects": subjects,
    })

    return nil
}

func (es *EventSubscriber) subscribeToSubject(ctx context.Context, subject string) error {
    config := &events.SubscriberConfig{
        MaxDeliver: 5,
        AckWait:    30 * time.Second,
    }

    return es.subscriber.SubscribeDurableWithConfig(
        ctx,
        subject,
        "ai-intelligence-hub-consumer",
        config,
        es.handleEvent,
    )
}

func (es *EventSubscriber) handleEvent(ctx context.Context, event *events.Event) error {
    es.logger.Debug(ctx, "Received event", logger.Tags{
        "event_type": event.Type,
        "event_id":   event.ID,
        "source":     event.Source,
    })

    // Process event asynchronously
    go func() {
        if err := es.eventProcessor.ProcessEvent(ctx, event); err != nil {
            es.logger.Error(ctx, err, "Failed to process event", logger.Tags{
                "event_type": event.Type,
                "event_id":   event.ID,
            })
        }
    }()

    return nil
}
```

#### T005: Event Router and Processor

**File**: `services/ai-intelligence-hub/internal/infrastructure/adapters/nats/event_processor.go`

```go
package nats

import (
    "context"
    "strings"

    "github.com/giia/giia-core-engine/pkg/events"
    "github.com/giia/giia-core-engine/pkg/logger"
    "github.com/giia/giia-core-engine/services/ai-intelligence-hub/internal/core/providers"
)

type EventProcessorImpl struct {
    bufferEventHandler    providers.BufferEventHandler
    executionEventHandler providers.ExecutionEventHandler
    userEventHandler      providers.UserEventHandler
    patternDetector       providers.PatternDetector
    logger                logger.Logger
}

func (ep *EventProcessorImpl) ProcessEvent(ctx context.Context, event *events.Event) error {
    // Route event to appropriate handler
    switch {
    case strings.HasPrefix(event.Type, "buffer."):
        return ep.bufferEventHandler.Handle(ctx, event)

    case strings.HasPrefix(event.Type, "order.execution"):
        return ep.executionEventHandler.Handle(ctx, event)

    case strings.HasPrefix(event.Type, "user."):
        return ep.userEventHandler.Handle(ctx, event)

    default:
        ep.logger.Debug(ctx, "No handler for event type", logger.Tags{
            "event_type": event.Type,
        })
        return nil
    }
}
```

---

### Phase 3: AI Analysis Engine (Week 3-4)

#### T006: Claude API Integration

**File**: `services/ai-intelligence-hub/internal/infrastructure/adapters/ai/claude_client.go`

```go
package ai

import (
    "context"
    "encoding/json"
    "fmt"

    anthropic "github.com/anthropics/anthropic-sdk-go"
    "github.com/giia/giia-core-engine/pkg/logger"
)

type ClaudeClient struct {
    client  *anthropic.Client
    cache   ResponseCache
    logger  logger.Logger
}

type AIAnalysisRequest struct {
    Event           *events.Event
    Context         map[string]interface{}
    DDMRPKnowledge  []string
    Prompt          string
}

type AIAnalysisResponse struct {
    Summary             string
    FullAnalysis        string
    Reasoning           string
    Recommendations     []string
    ImpactAssessment    map[string]interface{}
    Confidence          float64
}

func NewClaudeClient(apiKey string, cache ResponseCache, logger logger.Logger) *ClaudeClient {
    client := anthropic.NewClient(
        anthropic.WithAPIKey(apiKey),
    )

    return &ClaudeClient{
        client: client,
        cache:  cache,
        logger: logger,
    }
}

func (c *ClaudeClient) Analyze(ctx context.Context, req *AIAnalysisRequest) (*AIAnalysisResponse, error) {
    // Check cache first
    cacheKey := c.generateCacheKey(req)
    if cached, found := c.cache.Get(ctx, cacheKey); found {
        c.logger.Debug(ctx, "Cache hit for AI analysis", logger.Tags{
            "cache_key": cacheKey,
        })
        return cached, nil
    }

    // Build comprehensive prompt
    prompt := c.buildPrompt(req)

    // Call Claude API
    message, err := c.client.Messages.New(ctx, anthropic.MessageNewParams{
        Model: anthropic.F(anthropic.ModelClaude3_5SonnetLatest),
        Messages: anthropic.F([]anthropic.MessageParam{
            anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
        }),
        MaxTokens: anthropic.F(int64(2000)),
    })

    if err != nil {
        return nil, fmt.Errorf("claude API error: %w", err)
    }

    // Parse response
    response, err := c.parseResponse(message)
    if err != nil {
        return nil, err
    }

    // Cache response
    c.cache.Set(ctx, cacheKey, response, 1*time.Hour)

    c.logger.Info(ctx, "AI analysis completed", logger.Tags{
        "event_type": req.Event.Type,
        "confidence": response.Confidence,
    })

    return response, nil
}

func (c *ClaudeClient) buildPrompt(req *AIAnalysisRequest) string {
    return fmt.Sprintf(`
You are an expert DDMRP supply chain advisor analyzing a real-time event.

DDMRP KNOWLEDGE BASE:
%s

CURRENT EVENT:
Type: %s
Source: %s
Data: %s

RELATED CONTEXT:
%s

ANALYSIS REQUEST:
%s

Provide your analysis in the following JSON format:
{
  "summary": "Brief 1-2 sentence summary",
  "full_analysis": "Detailed analysis with reasoning",
  "reasoning": "Why this matters and root cause",
  "recommendations": ["Action 1", "Action 2", "Action 3"],
  "impact_assessment": {
    "risk_level": "low|medium|high|critical",
    "revenue_impact": 0.0,
    "cost_impact": 0.0,
    "time_to_impact_hours": 0,
    "affected_orders": 0,
    "affected_products": 0
  },
  "confidence": 0.0  // 0-1 confidence score
}

Be specific, actionable, and use natural conversational language.
`,
        strings.Join(req.DDMRPKnowledge, "\n\n"),
        req.Event.Type,
        req.Event.Source,
        c.formatEventData(req.Event.Data),
        c.formatContext(req.Context),
        req.Prompt,
    )
}
```

#### T007: RAG Knowledge Base

**File**: `services/ai-intelligence-hub/internal/infrastructure/adapters/rag/chromadb_client.go`

```go
package rag

import (
    "context"

    chroma "github.com/amikos-tech/chroma-go"
    "github.com/giia/giia-core-engine/pkg/logger"
)

type ChromaDBClient struct {
    client     *chroma.Client
    collection *chroma.Collection
    logger     logger.Logger
}

func NewChromaDBClient(host string, port int, logger logger.Logger) (*ChromaDBClient, error) {
    client, err := chroma.NewClient(chroma.WithHost(host), chroma.WithPort(port))
    if err != nil {
        return nil, err
    }

    // Get or create collection
    collection, err := client.GetOrCreateCollection(
        context.Background(),
        "ddmrp_knowledge",
        map[string]interface{}{
            "description": "DDMRP methodology and best practices",
        },
        true,
        chroma.NewOpenAIEmbeddingFunction("text-embedding-ada-002"),
    )

    if err != nil {
        return nil, err
    }

    return &ChromaDBClient{
        client:     client,
        collection: collection,
        logger:     logger,
    }, nil
}

func (c *ChromaDBClient) RetrieveRelevantKnowledge(ctx context.Context, query string, topK int) ([]string, error) {
    results, err := c.collection.Query(
        context.Background(),
        []string{query},
        topK,
        nil,
        nil,
        nil,
    )

    if err != nil {
        return nil, err
    }

    // Extract documents
    var docs []string
    for _, doc := range results.Documents[0] {
        docs = append(docs, doc)
    }

    c.logger.Debug(ctx, "Retrieved knowledge from RAG", logger.Tags{
        "query":        query,
        "results_count": len(docs),
    })

    return docs, nil
}
```

---

### Phase 4: Notification System (Week 5)

#### T008: Multi-Channel Notification Delivery

**File**: `services/ai-intelligence-hub/internal/infrastructure/adapters/notifications/email_sender.go`

```go
package notifications

import (
    "context"
    "fmt"

    "github.com/sendgrid/sendgrid-go"
    "github.com/sendgrid/sendgrid-go/helpers/mail"
)

type EmailSender struct {
    client  *sendgrid.Client
    fromEmail string
    fromName  string
}

func NewEmailSender(apiKey, fromEmail, fromName string) *EmailSender {
    return &EmailSender{
        client:    sendgrid.NewSendClient(apiKey),
        fromEmail: fromEmail,
        fromName:  fromName,
    }
}

func (es *EmailSender) SendNotification(ctx context.Context, notification *domain.AINotification, userEmail string) error {
    from := mail.NewEmail(es.fromName, es.fromEmail)
    to := mail.NewEmail("", userEmail)

    subject := fmt.Sprintf("[%s] %s", notification.Priority, notification.Title)

    htmlContent := es.buildHTMLContent(notification)
    textContent := es.buildTextContent(notification)

    message := mail.NewSingleEmail(from, subject, to, textContent, htmlContent)

    response, err := es.client.Send(message)
    if err != nil {
        return fmt.Errorf("sendgrid error: %w", err)
    }

    if response.StatusCode >= 400 {
        return fmt.Errorf("sendgrid returned %d: %s", response.StatusCode, response.Body)
    }

    return nil
}

func (es *EmailSender) buildHTMLContent(notification *domain.AINotification) string {
    // Build beautiful HTML email template
    return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; }
        .priority-%s { border-left: 4px solid #%s; }
        .recommendations { background: #f5f5f5; padding: 15px; }
    </style>
</head>
<body>
    <div class="priority-%s">
        <h2>%s</h2>
        <p>%s</p>
        <div class="recommendations">
            <h3>Recommended Actions:</h3>
            <ol>
                %s
            </ol>
        </div>
    </div>
</body>
</html>
    `,
        notification.Priority,
        es.getPriorityColor(notification.Priority),
        notification.Priority,
        notification.Title,
        notification.Summary,
        es.formatRecommendations(notification.Recommendations),
    )
}
```

---

### Phase 5: Use Case Implementation (Week 6)

#### T009: Stockout Risk Analysis

**File**: `services/ai-intelligence-hub/internal/core/usecases/analysis/analyze_stockout_risk.go`

```go
package analysis

import (
    "context"
    "fmt"
    "time"

    "github.com/giia/giia-core-engine/pkg/events"
    "github.com/giia/giia-core-engine/services/ai-intelligence-hub/internal/core/domain"
    "github.com/giia/giia-core-engine/services/ai-intelligence-hub/internal/core/providers"
)

type AnalyzeStockoutRiskUseCase struct {
    aiAnalyzer      providers.AIAnalyzer
    ragKnowledge    providers.RAGKnowledge
    ddmrpClient     providers.DDMRPServiceClient
    catalogClient   providers.CatalogServiceClient
    executionClient providers.ExecutionServiceClient
}

func (uc *AnalyzeStockoutRiskUseCase) Execute(ctx context.Context, event *events.Event) (*domain.AINotification, error) {
    // 1. Extract product ID from event
    productID := event.Data["product_id"].(string)
    orgID := event.OrganizationID

    // 2. Gather context from multiple services
    buffer, err := uc.ddmrpClient.GetBuffer(ctx, productID, orgID)
    if err != nil {
        return nil, err
    }

    product, err := uc.catalogClient.GetProduct(ctx, productID)
    if err != nil {
        return nil, err
    }

    pendingOrders, err := uc.executionClient.GetPendingOrders(ctx, productID, orgID)
    if err != nil {
        return nil, err
    }

    // 3. Calculate days until stockout
    daysUntilStockout := buffer.CurrentStock / buffer.DailyConsumption

    // 4. Retrieve DDMRP knowledge
    knowledge, err := uc.ragKnowledge.Retrieve(ctx, "buffer below minimum stockout prevention", 5)
    if err != nil {
        return nil, err
    }

    // 5. Build AI analysis request
    aiReq := &providers.AIAnalysisRequest{
        Event: event,
        Context: map[string]interface{}{
            "buffer":         buffer,
            "product":        product,
            "pending_orders": pendingOrders,
            "days_to_stockout": daysUntilStockout,
        },
        DDMRPKnowledge: knowledge,
        Prompt: fmt.Sprintf(`
Analyze this buffer below minimum situation for product "%s".

Current State:
- On-hand stock: %.2f units
- Daily consumption: %.2f units/day
- Days until stockout: %.1f days
- Min buffer: %.2f units
- Pending orders: %d

Provide:
1. Risk assessment (is stockout imminent?)
2. Revenue impact estimation (pending orders × product price)
3. Root cause analysis (why below minimum?)
4. Immediate action recommendations (emergency order? supplier switch?)
5. Long-term prevention (buffer recalculation?)

Use DDMRP best practices from the provided knowledge.
        `,
            product.Name,
            buffer.CurrentStock,
            buffer.DailyConsumption,
            daysUntilStockout,
            buffer.MinBuffer,
            len(pendingOrders),
        ),
    }

    // 6. Get AI analysis
    aiResponse, err := uc.aiAnalyzer.Analyze(ctx, aiReq)
    if err != nil {
        return nil, err
    }

    // 7. Build notification
    notification := &domain.AINotification{
        OrganizationID: orgID,
        UserID:         // TODO: Get relevant users
        Type:           domain.NotificationTypeAlert,
        Priority:       uc.determinePriority(daysUntilStockout),
        Title:          fmt.Sprintf("Imminent Stockout: %s", product.Name),
        Summary:        aiResponse.Summary,
        FullAnalysis:   aiResponse.FullAnalysis,
        Reasoning:      aiResponse.Reasoning,
        Impact:         uc.buildImpact(aiResponse.ImpactAssessment),
        Recommendations: uc.buildRecommendations(aiResponse.Recommendations),
        SourceEvents:   []string{event.ID},
        RelatedEntities: map[string][]string{
            "product_ids": {productID},
        },
        Status:    domain.NotificationStatusUnread,
        CreatedAt: time.Now(),
    }

    return notification, nil
}

func (uc *AnalyzeStockoutRiskUseCase) determinePriority(daysUntilStockout float64) domain.NotificationPriority {
    switch {
    case daysUntilStockout < 2:
        return domain.NotificationPriorityCritical
    case daysUntilStockout < 5:
        return domain.NotificationPriorityHigh
    case daysUntilStockout < 10:
        return domain.NotificationPriorityMedium
    default:
        return domain.NotificationPriorityLow
    }
}
```

---

### Phase 6: Testing & Integration (Week 7-8)

#### T010: Integration Tests

**File**: `services/ai-intelligence-hub/internal/core/usecases/analysis/analyze_stockout_risk_test.go`

```go
package analysis_test

import (
    "context"
    "testing"
    "time"

    "github.com/google/uuid"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"

    "github.com/giia/giia-core-engine/pkg/events"
    "github.com/giia/giia-core-engine/services/ai-intelligence-hub/internal/core/domain"
    "github.com/giia/giia-core-engine/services/ai-intelligence-hub/internal/core/usecases/analysis"
)

func TestAnalyzeStockoutRisk_Critical_Success(t *testing.T) {
    // Given
    mockAIAnalyzer := new(MockAIAnalyzer)
    mockRAG := new(MockRAGKnowledge)
    mockDDMRPClient := new(MockDDMRPServiceClient)
    mockCatalogClient := new(MockCatalogServiceClient)
    mockExecutionClient := new(MockExecutionServiceClient)

    useCase := analysis.NewAnalyzeStockoutRiskUseCase(
        mockAIAnalyzer,
        mockRAG,
        mockDDMRPClient,
        mockCatalogClient,
        mockExecutionClient,
    )

    givenProductID := uuid.New().String()
    givenOrgID := uuid.New().String()
    givenEvent := &events.Event{
        ID:             uuid.New().String(),
        Type:           "buffer.below_minimum",
        OrganizationID: givenOrgID,
        Data: map[string]interface{}{
            "product_id": givenProductID,
        },
    }

    givenBuffer := &Buffer{
        CurrentStock:      15.0,
        DailyConsumption:  10.0,  // 1.5 days until stockout!
        MinBuffer:         50.0,
    }

    mockDDMRPClient.On("GetBuffer", mock.Anything, givenProductID, givenOrgID).
        Return(givenBuffer, nil)

    mockCatalogClient.On("GetProduct", mock.Anything, givenProductID).
        Return(&Product{Name: "Widget-A", Price: 50.0}, nil)

    mockExecutionClient.On("GetPendingOrders", mock.Anything, givenProductID, givenOrgID).
        Return([]Order{{Amount: 100}}, nil)

    mockRAG.On("Retrieve", mock.Anything, mock.Anything, 5).
        Return([]string{"DDMRP knowledge..."}, nil)

    mockAIAnalyzer.On("Analyze", mock.Anything, mock.Anything).
        Return(&AIAnalysisResponse{
            Summary:      "Critical stockout in 1.5 days",
            FullAnalysis: "Detailed analysis...",
            Reasoning:    "Daily consumption exceeds stock",
            Recommendations: []string{
                "Place emergency order",
                "Contact Supplier B",
            },
            ImpactAssessment: map[string]interface{}{
                "risk_level":      "critical",
                "revenue_impact":  5000.0,
            },
            Confidence: 0.95,
        }, nil)

    // When
    notification, err := useCase.Execute(context.Background(), givenEvent)

    // Then
    assert.NoError(t, err)
    assert.NotNil(t, notification)
    assert.Equal(t, domain.NotificationTypeAlert, notification.Type)
    assert.Equal(t, domain.NotificationPriorityCritical, notification.Priority)
    assert.Contains(t, notification.Title, "Widget-A")
    assert.Equal(t, 2, len(notification.Recommendations))

    mockAIAnalyzer.AssertExpectations(t)
    mockDDMRPClient.AssertExpectations(t)
}
```

---

## 4. Success Criteria Checklist

### Mandatory Features
- [ ] Real-time event processing from NATS (<1s p95)
- [ ] Critical alert generation (stockouts, execution failures)
- [ ] AI-powered analysis with Claude API
- [ ] RAG knowledge base with ChromaDB
- [ ] Multi-channel notification delivery
- [ ] User preference management
- [ ] Daily digest generation
- [ ] Pattern detection (3+ related events)
- [ ] 80%+ test coverage
- [ ] Multi-tenancy support

### Performance Metrics
- [ ] Event processing: <1s p95
- [ ] AI analysis: <5s p95
- [ ] Notification delivery: <10s p95
- [ ] Handle 1000+ events/minute
- [ ] False positive rate: <10%

### Quality Metrics
- [ ] 90%+ notification accuracy
- [ ] 60%+ recommendation acceptance rate
- [ ] 85%+ pattern detection accuracy
- [ ] User satisfaction: 4+/5

---

## 5. Dependencies & Timeline

```
Week 1: Foundation & Database
Week 2: Event Processing
Week 3: AI Analysis Engine (Part 1)
Week 4: AI Analysis Engine (Part 2) + RAG
Week 5: Notification System
Week 6: Use Case Implementation
Week 7: Integration Testing
Week 8: Polish, Performance, Documentation
```

---

## 6. Cost Estimation

### Claude API Costs (Monthly per org)
- ~1000 events/day analyzed
- ~500 tokens/analysis avg
- ~$30-50/month per organization

### Infrastructure
- ChromaDB: $0 (self-hosted)
- Redis cache: $20/month
- SendGrid: $15/month (up to 40K emails)
- Twilio SMS: Pay-as-you-go

**Total**: ~$65-85/month per org

---

## 7. Risk Mitigation

| Risk | Mitigation Strategy |
|------|---------------------|
| High Claude API costs | Aggressive caching (1-hour TTL), batch analysis |
| False positive fatigue | User feedback loop, confidence thresholds, A/B testing |
| Event processing lag | Horizontal scaling, async processing, priority queues |
| RAG knowledge drift | Monthly updates, version control |

---

**Document Version**: 2.0 (Complete Rewrite)
**Last Updated**: 2025-12-22
**Status**: Ready for Implementation
**Estimated Start Date**: TBD
**Estimated Completion**: 6-8 weeks from start
