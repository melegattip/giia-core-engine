# AI Intelligence Hub - Testing Strategy

**Version:** 1.0
**Last Updated:** 2025-12-23
**Target Coverage:** 80%+ overall, 90%+ critical paths

---

## Table of Contents

1. [Testing Pyramid](#testing-pyramid)
2. [Unit Testing](#unit-testing)
3. [Integration Testing](#integration-testing)
4. [End-to-End Testing](#end-to-end-testing)
5. [Load Testing](#load-testing)
6. [RAG System Testing](#rag-system-testing)
7. [AI Analysis Testing](#ai-analysis-testing)
8. [Test Data Management](#test-data-management)
9. [Continuous Integration](#continuous-integration)

---

## Testing Pyramid

```
                    /\
                   /  \
                  / E2E \          ~5% of tests
                 /  Tests \        Full system scenarios
                /──────────\
               /            \
              / Integration  \    ~25% of tests
             /     Tests      \   Component interactions
            /──────────────────\
           /                    \
          /     Unit Tests       \ ~70% of tests
         /________________________\ Fast, isolated tests

Target Distribution:
- Unit Tests: 70% (~300 tests)
- Integration Tests: 25% (~100 tests)
- E2E Tests: 5% (~20 tests)

Total: ~420 tests
```

---

## Unit Testing

### Coverage Targets by Component

| Component | Target Coverage | Rationale |
|-----------|----------------|-----------|
| **Event Processing** | 90%+ | Critical path - event routing must be reliable |
| **AI Analysis** | 85%+ | Core business logic |
| **RAG Retrieval** | 80%+ | Important but external dependency |
| **Notification Routing** | 90%+ | Critical - must deliver to right channels |
| **Domain Entities** | 95%+ | Business rules validation |
| **Repositories** | 75%+ | Database interactions |
| **Adapters** | 70%+ | External service wrappers |

### Unit Test Structure

Follow **Given-When-Then** pattern with descriptive test names:

```go
// ✅ GOOD: Clear test structure
func TestAnalyzeStockoutRisk_BelowRedZone_ReturnsCriticalPriority(t *testing.T) {
    // Given: Product with critical buffer status
    givenProductID := uuid.New()
    givenBuffer := &Buffer{
        CurrentStock:     15.0,
        DailyConsumption: 10.0,  // 1.5 days until stockout
        MinBuffer:        50.0,
    }

    mockDDMRPClient := new(MockDDMRPClient)
    mockDDMRPClient.On("GetBuffer", mock.Anything, givenProductID.String(), mock.Anything).
        Return(givenBuffer, nil)

    useCase := NewAnalyzeStockoutRiskUseCase(mockDDMRPClient, ...)

    // When: Analysis is executed
    notification, err := useCase.Execute(context.Background(), createTestEvent(givenProductID))

    // Then: Critical priority notification generated
    assert.NoError(t, err)
    assert.NotNil(t, notification)
    assert.Equal(t, domain.NotificationPriorityCritical, notification.Priority)
    assert.Contains(t, notification.Title, "Stockout")
    assert.GreaterOrEqual(t, len(notification.Recommendations), 1)

    mockDDMRPClient.AssertExpectations(t)
}
```

### Mock Best Practices

```go
// ✅ GOOD: Specific mock expectations
mockRepo.On("GetBuffer",
    mock.AnythingOfType("*context.Context"),
    "product-uuid-123",  // Specific product ID
    "org-uuid-456",      // Specific org ID
).Return(&Buffer{...}, nil).Once()

// ❌ BAD: Over-use of mock.Anything
mockRepo.On("GetBuffer",
    mock.Anything,
    mock.Anything,
    mock.Anything,
).Return(&Buffer{...}, nil)
```

### Critical Unit Tests

#### 1. Event Router Tests

```go
func TestEventRouter_BufferBelowMinimum_RoutesToBufferHandler(t *testing.T) {
    // Given
    mockBufferHandler := new(MockBufferEventHandler)
    router := NewEventRouter(mockBufferHandler, ...)

    givenEvent := &events.Event{
        Type: "buffer.below_minimum",
        Data: map[string]interface{}{"product_id": "test-123"},
    }

    mockBufferHandler.On("Handle", mock.Anything, givenEvent).Return(nil)

    // When
    err := router.Route(context.Background(), givenEvent)

    // Then
    assert.NoError(t, err)
    mockBufferHandler.AssertCalled(t, "Handle", mock.Anything, givenEvent)
}

func TestEventRouter_UnknownEventType_IgnoresGracefully(t *testing.T) {
    // Given
    router := NewEventRouter(...)
    givenEvent := &events.Event{Type: "unknown.event"}

    // When
    err := router.Route(context.Background(), givenEvent)

    // Then
    assert.NoError(t, err)  // Should not error on unknown events
}
```

#### 2. Priority Determination Tests

```go
func TestDeterminePriority_LessThan2Days_ReturnsCritical(t *testing.T) {
    testCases := []struct {
        name                string
        givenDaysToStockout float64
        expectedPriority    domain.NotificationPriority
    }{
        {"0.5 days", 0.5, domain.NotificationPriorityCritical},
        {"1.9 days", 1.9, domain.NotificationPriorityCritical},
        {"2.5 days", 2.5, domain.NotificationPriorityHigh},
        {"5.0 days", 5.0, domain.NotificationPriorityMedium},
        {"12.0 days", 12.0, domain.NotificationPriorityLow},
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            // When
            priority := determinePriority(tc.givenDaysToStockout)

            // Then
            assert.Equal(t, tc.expectedPriority, priority)
        })
    }
}
```

#### 3. Rate Limiter Tests

```go
func TestRateLimiter_UnderLimit_AllowsRequest(t *testing.T) {
    // Given
    mockRedis := newMockRedisClient()
    rateLimiter := NewRateLimiter(mockRedis)

    givenUserID := uuid.New()
    givenMaxPerHour := 10

    mockRedis.ExpectIncr("rate_limit:user:email:hour").SetVal(5)

    // When
    allowed, err := rateLimiter.CheckLimit(
        context.Background(),
        givenUserID,
        ChannelEmail,
        givenMaxPerHour,
    )

    // Then
    assert.NoError(t, err)
    assert.True(t, allowed)
}

func TestRateLimiter_OverLimit_RejectsRequest(t *testing.T) {
    // Given
    mockRedis := newMockRedisClient()
    rateLimiter := NewRateLimiter(mockRedis)

    givenUserID := uuid.New()
    givenMaxPerHour := 10

    mockRedis.ExpectIncr("rate_limit:user:email:hour").SetVal(11)

    // When
    allowed, err := rateLimiter.CheckLimit(
        context.Background(),
        givenUserID,
        ChannelEmail,
        givenMaxPerHour,
    )

    // Then
    assert.NoError(t, err)
    assert.False(t, allowed)
}
```

#### 4. Notification Router Tests

```go
func TestNotificationRouter_CriticalPriority_SendsToAllChannels(t *testing.T) {
    // Given
    router := NewNotificationRouter()

    givenNotification := &domain.AINotification{
        Priority: domain.NotificationPriorityCritical,
    }

    givenUserPrefs := &domain.UserNotificationPreferences{
        EnableInApp: true,
        EnableEmail: true,
        EnableSMS:   true,
        EnableSlack: false,
    }

    // When
    deliveries := router.Route(context.Background(), givenNotification, givenUserPrefs)

    // Then
    assert.Len(t, deliveries, 3)  // In-App, Email, SMS
    assert.Contains(t, getChannels(deliveries), ChannelInApp)
    assert.Contains(t, getChannels(deliveries), ChannelEmail)
    assert.Contains(t, getChannels(deliveries), ChannelSMS)
    assert.NotContains(t, getChannels(deliveries), ChannelSlack)  // Disabled
}

func TestNotificationRouter_QuietHours_SkipsNonCritical(t *testing.T) {
    // Given
    router := NewNotificationRouter()

    givenNotification := &domain.AINotification{
        Priority: domain.NotificationPriorityHigh,  // Not critical
    }

    givenUserPrefs := &domain.UserNotificationPreferences{
        EnableEmail:       true,
        QuietHoursStart:   parseTime("22:00"),
        QuietHoursEnd:     parseTime("07:00"),
        Timezone:          "America/New_York",
    }

    // Mock current time to be in quiet hours
    mockTime := parseTime("23:30")  // 11:30 PM

    // When
    deliveries := router.Route(
        withMockTime(context.Background(), mockTime),
        givenNotification,
        givenUserPrefs,
    )

    // Then
    assert.Empty(t, deliveries)  // No delivery during quiet hours
}
```

---

## Integration Testing

Integration tests verify component interactions without mocking.

### Test Environment Setup

```go
type TestEnvironment struct {
    DB              *sql.DB
    NATS            *nats.Conn
    Redis           *redis.Client
    ChromaDB        *chroma.Client
    EventPublisher  *events.Publisher
    TestServer      *httptest.Server
}

func setupTestEnvironment(t *testing.T) *TestEnvironment {
    // Use test containers for real dependencies
    postgresContainer := testcontainers.PostgresContainer(...)
    natsContainer := testcontainers.NATSContainer(...)
    redisContainer := testcontainers.RedisContainer(...)

    db := connectToPostgres(postgresContainer.ConnectionString())
    runMigrations(db)

    natsConn := connectToNATS(natsContainer.URL())
    redis := connectToRedis(redisContainer.URL())

    return &TestEnvironment{
        DB:    db,
        NATS:  natsConn,
        Redis: redis,
        ...
    }
}

func (env *TestEnvironment) Teardown() {
    env.DB.Close()
    env.NATS.Close()
    env.Redis.Close()
    // Stop containers
}
```

### Integration Test Examples

#### 1. Event Processing Integration

```go
func TestEventProcessing_BufferEvent_CreatesNotification(t *testing.T) {
    // Given: Full system with real dependencies
    env := setupTestEnvironment(t)
    defer env.Teardown()

    // Seed test data
    product := createTestProduct(t, env.DB)
    user := createTestUser(t, env.DB)

    // When: Publish buffer event
    event := &events.Event{
        ID:             uuid.New().String(),
        Type:           "buffer.below_minimum",
        OrganizationID: product.OrganizationID.String(),
        Data: map[string]interface{}{
            "product_id":    product.ID.String(),
            "current_stock": 15.0,
            "min_buffer":    50.0,
        },
    }

    err := env.EventPublisher.Publish(context.Background(), event)
    require.NoError(t, err)

    // Then: Notification created in database
    notification := waitForNotification(t, env.DB, user.ID, 5*time.Second)
    assert.NotNil(t, notification)
    assert.Equal(t, domain.NotificationTypeAlert, notification.Type)
    assert.Contains(t, notification.Title, product.Name)
}
```

#### 2. RAG Knowledge Retrieval Integration

```go
func TestRAGRetrieval_StockoutQuery_ReturnsRelevantKnowledge(t *testing.T) {
    // Given: ChromaDB with loaded knowledge base
    env := setupTestEnvironment(t)
    defer env.Teardown()

    // Load DDMRP knowledge into ChromaDB
    loadKnowledgeBase(t, env.ChromaDB)

    ragClient := NewChromaDBClient(env.ChromaDB, logger)

    // When: Query for stockout prevention knowledge
    query := "DDMRP buffer below minimum stockout prevention emergency order"
    docs, err := ragClient.RetrieveRelevantKnowledge(context.Background(), query, 5)

    // Then: Returns relevant DDMRP documents
    require.NoError(t, err)
    assert.Len(t, docs, 5)
    assert.Contains(t, strings.Join(docs, " "), "buffer calculation")
    assert.Contains(t, strings.Join(docs, " "), "execution best practices")
}
```

#### 3. Multi-Channel Delivery Integration

```go
func TestMultiChannelDelivery_CriticalAlert_DeliversToAll(t *testing.T) {
    // Given: Test environment with mock email/SMS services
    env := setupTestEnvironment(t)
    defer env.Teardown()

    mockEmailSender := newMockEmailSender()
    mockSMSSender := newMockSMSSender()

    notificationService := NewNotificationService(
        mockEmailSender,
        mockSMSSender,
        env.DB,
        logger,
    )

    user := createTestUser(t, env.DB, WithEmailNotifications, WithSMSNotifications)
    notification := createCriticalNotification(t, env.DB, user.ID)

    // When: Deliver notification
    err := notificationService.Deliver(context.Background(), notification, user)

    // Then: Delivered to all enabled channels
    require.NoError(t, err)
    assert.True(t, mockEmailSender.WasSent(user.Email))
    assert.True(t, mockSMSSender.WasSent(user.PhoneNumber))
}
```

---

## End-to-End Testing

E2E tests verify complete user scenarios from event to notification delivery.

### E2E Test Framework

```go
type E2ETestFramework struct {
    env            *TestEnvironment
    apiClient      *APIClient
    websocketConn  *websocket.Conn
    receivedNotifs chan *Notification
}

func setupE2E(t *testing.T) *E2ETestFramework {
    env := setupTestEnvironment(t)

    // Start service
    service := startIntelligenceHubService(env)

    // Connect WebSocket for real-time notifications
    ws := connectWebSocket(t, service.WebSocketURL())

    framework := &E2ETestFramework{
        env:            env,
        apiClient:      newAPIClient(service.HTTPURL()),
        websocketConn:  ws,
        receivedNotifs: make(chan *Notification, 100),
    }

    // Listen for WebSocket notifications
    go framework.listenForNotifications()

    return framework
}
```

### E2E Test Scenarios

#### Scenario 1: Complete Stockout Prevention Flow

```go
func TestE2E_StockoutPreventionFlow(t *testing.T) {
    e2e := setupE2E(t)
    defer e2e.Teardown()

    // Step 1: Create test data
    org := e2e.createOrganization(t)
    user := e2e.createUser(t, org, WithFullNotifications)
    product := e2e.createProduct(t, org, "Widget-A")
    buffer := e2e.createBuffer(t, product, BelowRedZone)

    // Step 2: Publish buffer event
    event := &events.Event{
        Type:           "buffer.below_minimum",
        OrganizationID: org.ID.String(),
        Data: map[string]interface{}{
            "product_id":    product.ID.String(),
            "current_stock": 15.0,
            "min_buffer":    140.0,
        },
    }

    err := e2e.publishEvent(event)
    require.NoError(t, err)

    // Step 3: Verify WebSocket notification received
    notification := e2e.waitForNotification(10 * time.Second)
    assert.NotNil(t, notification)
    assert.Equal(t, "alert", notification.Type)
    assert.Equal(t, "critical", notification.Priority)
    assert.Contains(t, notification.Title, "Widget-A")

    // Step 4: Verify notification stored in DB
    dbNotif := e2e.getNotificationFromDB(t, notification.ID)
    assert.Equal(t, "unread", dbNotif.Status)
    assert.GreaterOrEqual(t, len(dbNotif.Recommendations), 2)

    // Step 5: Mark as acted upon via API
    err = e2e.apiClient.MarkAsActedUpon(notification.ID, "Placed emergency order")
    require.NoError(t, err)

    // Step 6: Verify status updated
    updatedNotif := e2e.getNotificationFromDB(t, notification.ID)
    assert.Equal(t, "acted_upon", updatedNotif.Status)
    assert.NotNil(t, updatedNotif.ActedAt)

    // Step 7: Verify metrics updated
    analytics := e2e.apiClient.GetAnalytics(org.ID, last24Hours())
    assert.Equal(t, 1, analytics.TotalNotifications)
    assert.Equal(t, 1, analytics.ActedUpon)
}
```

#### Scenario 2: Daily Digest Generation

```go
func TestE2E_DailyDigestGeneration(t *testing.T) {
    e2e := setupE2E(t)
    defer e2e.Teardown()

    // Step 1: Create user with digest enabled
    user := e2e.createUser(t, org, WithDigestTime("06:00"))

    // Step 2: Create several notifications over past 24 hours
    e2e.createNotification(t, user, Critical, "Stockout alert")
    e2e.createNotification(t, user, High, "Buffer drift warning")
    e2e.createNotification(t, user, Medium, "Cost opportunity")

    // Step 3: Trigger digest generation (mock time to 6 AM)
    e2e.mockTime("2025-12-24T06:00:00Z")
    e2e.triggerScheduledJob("daily_digest")

    // Step 4: Verify digest email sent
    digestEmail := e2e.waitForEmail(user.Email, 5*time.Second)
    assert.NotNil(t, digestEmail)
    assert.Contains(t, digestEmail.Subject, "Daily Intelligence Digest")
    assert.Contains(t, digestEmail.Body, "3 notifications")
    assert.Contains(t, digestEmail.Body, "Stockout alert")
}
```

---

## Load Testing

### k6 Load Test Scripts

```javascript
// load-test.js
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

const errorRate = new Rate('errors');

export let options = {
  stages: [
    { duration: '2m', target: 50 },   // Ramp up to 50 publishers
    { duration: '5m', target: 100 },  // Sustained load
    { duration: '2m', target: 0 },    // Ramp down
  ],
  thresholds: {
    'http_req_duration': ['p(95)<1000', 'p(99)<3000'],  // 95% < 1s, 99% < 3s
    'errors': ['rate<0.01'],  // <1% error rate
  },
};

export default function () {
  // Publish event to NATS
  const event = {
    type: 'buffer.below_minimum',
    organization_id: __ENV.ORG_ID,
    data: {
      product_id: `product-${__VU}-${__ITER}`,
      current_stock: 15.0,
      min_buffer: 50.0,
    },
  };

  const res = http.post(
    'http://intelligence-hub.local/internal/events',
    JSON.stringify(event),
    {
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${__ENV.API_TOKEN}`,
      },
    }
  );

  check(res, {
    'status is 202': (r) => r.status === 202,
    'response time < 1s': (r) => r.timings.duration < 1000,
  }) || errorRate.add(1);

  sleep(1);
}
```

### Load Test Execution

```bash
# Run load test
k6 run --vus 100 --duration 10m load-test.js

# Expected results:
# - Event processing: <1s p95
# - Error rate: <1%
# - Throughput: 1000+ events/minute
```

---

## RAG System Testing

### Knowledge Base Quality Tests

```go
func TestRAGKnowledgeQuality_BufferCalculation_ReturnsAccurate(t *testing.T) {
    // Given: Loaded knowledge base
    rag := setupRAGSystem(t)

    testQueries := []struct {
        query           string
        expectedConcepts []string
    }{
        {
            query: "How to calculate DDMRP buffer green zone",
            expectedConcepts: []string{
                "green zone",
                "lead time factor",
                "average daily usage",
            },
        },
        {
            query: "Emergency order when stockout imminent",
            expectedConcepts: []string{
                "emergency order",
                "alternative supplier",
                "expedite",
                "lead time",
            },
        },
    }

    for _, tc := range testQueries {
        t.Run(tc.query, func(t *testing.T) {
            // When: Retrieve knowledge
            docs, err := rag.Retrieve(context.Background(), tc.query, 5)

            // Then: Contains expected concepts
            require.NoError(t, err)
            combined := strings.ToLower(strings.Join(docs, " "))

            for _, concept := range tc.expectedConcepts {
                assert.Contains(t, combined, strings.ToLower(concept),
                    fmt.Sprintf("Missing concept: %s", concept))
            }
        })
    }
}
```

---

## AI Analysis Testing

### Prompt Quality Tests

```go
func TestPromptBuilder_StockoutScenario_BuildsComprehensive(t *testing.T) {
    // Given: Stockout scenario data
    builder := NewPromptBuilder()

    context := &AnalysisContext{
        Event: &events.Event{Type: "buffer.below_minimum"},
        Buffer: &Buffer{
            CurrentStock:     15.0,
            DailyConsumption: 10.0,
            MinBuffer:        50.0,
        },
        Product: &Product{Name: "Widget-A", Price: 50.0},
        DDMRPKnowledge: []string{"Buffer calculation methodology..."},
    }

    // When: Build prompt
    prompt := builder.BuildStockoutPrompt(context)

    // Then: Contains all necessary sections
    assert.Contains(t, prompt, "DDMRP KNOWLEDGE")
    assert.Contains(t, prompt, "CURRENT EVENT")
    assert.Contains(t, prompt, "BUFFER STATUS")
    assert.Contains(t, prompt, "15.0 units")  // Current stock
    assert.Contains(t, prompt, "Widget-A")    // Product name
    assert.Contains(t, prompt, "JSON format") // Response format
}
```

### AI Response Validation Tests

```go
func TestResponseValidator_ValidResponse_AcceptsCorrectly(t *testing.T) {
    validator := NewAIResponseValidator()

    givenResponse := `{
        "summary": "Critical stockout in 1.5 days",
        "full_analysis": "Detailed analysis...",
        "reasoning": "DDMRP methodology indicates...",
        "recommendations": [
            {
                "action": "Place emergency order",
                "reasoning": "Lead time constraint",
                "expected_outcome": "Stockout prevented",
                "effort": "low",
                "impact": "high",
                "priority_order": 1
            }
        ],
        "impact_assessment": {
            "risk_level": "critical",
            "revenue_impact": 15000.00,
            "cost_impact": 200.00,
            "time_to_impact_hours": 36,
            "affected_orders": 5,
            "affected_products": 1
        },
        "confidence": 0.95
    }`

    // When: Validate
    parsed, err := validator.Validate(givenResponse)

    // Then: Accepted
    require.NoError(t, err)
    assert.NotNil(t, parsed)
    assert.Equal(t, "critical", parsed.ImpactAssessment.RiskLevel)
    assert.Equal(t, 0.95, parsed.Confidence)
}
```

---

## Test Data Management

### Test Data Builders

```go
// Product builder
type ProductBuilder struct {
    product *Product
}

func NewProductBuilder() *ProductBuilder {
    return &ProductBuilder{
        product: &Product{
            ID:    uuid.New(),
            Name:  "Test Product",
            SKU:   "TEST-001",
            Price: 50.0,
        },
    }
}

func (b *ProductBuilder) WithName(name string) *ProductBuilder {
    b.product.Name = name
    return b
}

func (b *ProductBuilder) WithPrice(price float64) *ProductBuilder {
    b.product.Price = price
    return b
}

func (b *ProductBuilder) Build() *Product {
    return b.product
}

// Usage
product := NewProductBuilder().
    WithName("Widget-A").
    WithPrice(75.0).
    Build()
```

---

## Continuous Integration

### GitHub Actions Workflow

```yaml
# .github/workflows/ai-intelligence-hub-test.yml
name: AI Intelligence Hub Tests

on:
  push:
    branches: [develop, main]
    paths:
      - 'services/ai-intelligence-hub/**'
  pull_request:
    paths:
      - 'services/ai-intelligence-hub/**'

jobs:
  test:
    runs-on: ubuntu-latest

    services:
      postgres:
        image: postgres:16
        env:
          POSTGRES_PASSWORD: test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s

      nats:
        image: nats:latest
        options: --name nats -p 4222:4222

      redis:
        image: redis:7
        options: >-
          --health-cmd "redis-cli ping"

    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.23.4'

      - name: Run unit tests
        run: |
          cd services/ai-intelligence-hub
          go test -v -race -coverprofile=coverage.out -covermode=atomic ./...

      - name: Check coverage
        run: |
          coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
          if (( $(echo "$coverage < 80" | bc -l) )); then
            echo "Coverage $coverage% is below 80% threshold"
            exit 1
          fi

      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out

      - name: Run integration tests
        env:
          DATABASE_URL: postgres://postgres:test@localhost:5432/intelligence_hub_test
          NATS_URL: nats://localhost:4222
          REDIS_URL: redis://localhost:6379
        run: |
          cd services/ai-intelligence-hub
          go test -v -tags=integration ./tests/integration/...

      - name: Run linters
        run: |
          cd services/ai-intelligence-hub
          golangci-lint run
```

---

## Summary

This testing strategy ensures:

✅ **80%+ code coverage** through comprehensive unit tests
✅ **Integration testing** with real dependencies via testcontainers
✅ **E2E testing** of complete user scenarios
✅ **Load testing** to verify performance targets (1000+ events/min)
✅ **RAG quality testing** to ensure knowledge retrieval accuracy
✅ **AI validation testing** to verify prompt and response quality
✅ **CI/CD integration** for automated testing on every commit

**Test Execution Time Targets:**
- Unit tests: <30 seconds
- Integration tests: <2 minutes
- E2E tests: <5 minutes
- Load tests: 10 minutes (run manually before releases)

**Total CI Pipeline:** <10 minutes for fast feedback loops

---

**Version:** 1.0
**Last Updated:** 2025-12-23
**Status:** ✅ Ready for Implementation