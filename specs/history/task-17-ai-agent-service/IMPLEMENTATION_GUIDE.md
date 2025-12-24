# Task 17: AI Intelligence Hub - Complete Implementation Guide

**Task ID**: task-17-ai-intelligence-hub
**Version**: 1.0
**Last Updated**: 2025-12-23
**Status**: Ready for Development

---

## Table of Contents

1. [Architecture Overview](#1-architecture-overview)
2. [RAG System Deep Dive](#2-rag-system-deep-dive)
3. [AI Analysis Patterns](#3-ai-analysis-patterns)
4. [Event Processing Strategies](#4-event-processing-strategies)
5. [Notification Routing Logic](#5-notification-routing-logic)
6. [Testing Strategy](#6-testing-strategy)
7. [Deployment Guide](#7-deployment-guide)
8. [Monitoring & Observability](#8-monitoring--observability)

---

## 1. Architecture Overview

### System Context Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                      GIIA Platform Ecosystem                     │
│                                                                  │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐       │
│  │   Auth   │  │ Catalog  │  │  DDMRP   │  │Execution │       │
│  │ Service  │  │ Service  │  │ Service  │  │ Service  │       │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘       │
│       │             │              │              │              │
│       └─────────────┴──────────────┴──────────────┘              │
│                           │                                      │
│                           ▼ (Events via NATS)                   │
│                                                                  │
│              ┌───────────────────────────────┐                  │
│              │  AI Intelligence Hub Service  │                  │
│              │                               │                  │
│              │  ┌─────────────────────────┐ │                  │
│              │  │  Event Subscriber       │ │                  │
│              │  │  (NATS Consumer)        │ │                  │
│              │  └──────────┬──────────────┘ │                  │
│              │             │                 │                  │
│              │             ▼                 │                  │
│              │  ┌─────────────────────────┐ │                  │
│              │  │  Event Router           │ │                  │
│              │  │  (Pattern Matcher)      │ │                  │
│              │  └──────────┬──────────────┘ │                  │
│              │             │                 │                  │
│              │    ┌────────┴────────┐        │                  │
│              │    ▼                 ▼        │                  │
│              │  ┌─────┐         ┌─────┐     │                  │
│              │  │Buffer│        │Exec │     │                  │
│              │  │Event │        │Event│     │                  │
│              │  │Handler│      │Handler│    │                  │
│              │  └──┬──┘         └──┬──┘     │                  │
│              │     │               │         │                  │
│              │     └───────┬───────┘         │                  │
│              │             ▼                 │                  │
│              │  ┌─────────────────────────┐ │                  │
│              │  │  Analysis Orchestrator  │ │                  │
│              │  │  • Context Builder      │ │                  │
│              │  │  • RAG Retrieval        │ │                  │
│              │  │  • AI Invocation        │ │                  │
│              │  └──────────┬──────────────┘ │                  │
│              │             │                 │                  │
│              │             ▼                 │                  │
│              │  ┌─────────────────────────┐ │                  │
│              │  │  Notification Generator │ │                  │
│              │  │  • Priority Assignment  │ │                  │
│              │  │  • Channel Routing      │ │                  │
│              │  └──────────┬──────────────┘ │                  │
│              │             │                 │                  │
│              └─────────────┼─────────────────┘                  │
│                            │                                     │
└────────────────────────────┼─────────────────────────────────────┘
                             │
                             ▼
                  ┌────────────────────┐
                  │  Notification      │
                  │  Delivery Services │
                  │                    │
                  │  • WebSocket       │
                  │  • Email           │
                  │  • Slack           │
                  │  • SMS             │
                  └────────────────────┘
```

### Component Interaction Flow

```
┌──────────────┐
│ Buffer Event │ buffer.below_minimum
└──────┬───────┘
       │
       ▼
┌──────────────────────────────────────────────────────┐
│ Step 1: Event Received                               │
│ • Validate event schema                              │
│ • Extract organization_id, product_id                │
│ • Check event deduplication (prevent duplicates)     │
└──────┬───────────────────────────────────────────────┘
       │
       ▼
┌──────────────────────────────────────────────────────┐
│ Step 2: Context Gathering (Parallel)                 │
│ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐  │
│ │ DDMRP Client │ │Catalog Client│ │ Exec Client  │  │
│ │ GetBuffer()  │ │GetProduct()  │ │GetPending()  │  │
│ └──────────────┘ └──────────────┘ └──────────────┘  │
│       │                  │                 │         │
│       └──────────────────┴─────────────────┘         │
│                          │                           │
│                  ┌───────▼────────┐                  │
│                  │ Context Object │                  │
│                  │ • Buffer state │                  │
│                  │ • Product info │                  │
│                  │ • Pending ord. │                  │
│                  └────────────────┘                  │
└──────┬───────────────────────────────────────────────┘
       │
       ▼
┌──────────────────────────────────────────────────────┐
│ Step 3: RAG Knowledge Retrieval                      │
│ • Build semantic query:                              │
│   "buffer below minimum stockout prevention DDMRP"   │
│ • ChromaDB similarity search (top-5)                 │
│ • Retrieved documents:                               │
│   1. buffer_calculation.md (chunk 3)                 │
│   2. execution_best_practices.md (chunk 7)           │
│   3. lead_time_management.md (chunk 2)               │
│   4. historical_pattern_nov2024.md                   │
│   5. supplier_selection_criteria.md                  │
└──────┬───────────────────────────────────────────────┘
       │
       ▼
┌──────────────────────────────────────────────────────┐
│ Step 4: AI Analysis (Claude API)                     │
│ Prompt Construction:                                 │
│ ┌──────────────────────────────────────────────┐    │
│ │ System: You are DDMRP expert                 │    │
│ │                                              │    │
│ │ DDMRP Knowledge (from RAG):                  │    │
│ │ [5 retrieved documents]                      │    │
│ │                                              │    │
│ │ Current Event:                               │    │
│ │ Type: buffer.below_minimum                   │    │
│ │ Product: Widget-A                            │    │
│ │ Current Stock: 15 units                      │    │
│ │ Min Buffer: 50 units                         │    │
│ │ Daily Consumption: 10 units/day              │    │
│ │ Pending Orders: 0                            │    │
│ │                                              │    │
│ │ Task: Analyze and recommend actions          │    │
│ └──────────────────────────────────────────────┘    │
│                                                      │
│ Claude Response:                                     │
│ ┌──────────────────────────────────────────────┐    │
│ │ {                                            │    │
│ │   "summary": "Critical: 1.5 days to stockout"│    │
│ │   "analysis": "...",                         │    │
│ │   "recommendations": [                       │    │
│ │     "Emergency order from Supplier B",       │    │
│ │     "Increase buffer by 20%"                 │    │
│ │   ],                                         │    │
│ │   "impact": {                                │    │
│ │     "revenue_at_risk": 15000,                │    │
│ │     "time_to_impact": "36 hours"             │    │
│ │   }                                          │    │
│ │ }                                            │    │
│ └──────────────────────────────────────────────┘    │
└──────┬───────────────────────────────────────────────┘
       │
       ▼
┌──────────────────────────────────────────────────────┐
│ Step 5: Notification Creation                        │
│ • Map AI response → AINotification entity            │
│ • Determine priority (Critical - <2 days stockout)   │
│ • Store in PostgreSQL                                │
│ • Generate notification ID                           │
└──────┬───────────────────────────────────────────────┘
       │
       ▼
┌──────────────────────────────────────────────────────┐
│ Step 6: User Preference Lookup                       │
│ • Query user_notification_preferences                │
│ • Check: enable_email=true, min_priority=medium      │
│ • Check: enable_sms=true, min_priority=critical ✓    │
│ • Check: quiet_hours (not applicable - critical)     │
│ • Route to: WebSocket + Email + SMS                  │
└──────┬───────────────────────────────────────────────┘
       │
       ▼
┌──────────────────────────────────────────────────────┐
│ Step 7: Multi-Channel Delivery (Parallel)            │
│ ┌──────────┐  ┌──────────┐  ┌──────────┐            │
│ │WebSocket │  │  Email   │  │   SMS    │            │
│ │  Push    │  │SendGrid  │  │ Twilio   │            │
│ └────┬─────┘  └────┬─────┘  └────┬─────┘            │
│      │             │              │                  │
│      ▼             ▼              ▼                  │
│    User         User            User                 │
│   Browser       Inbox           Phone                │
└──────────────────────────────────────────────────────┘

Total Time: ~3-5 seconds (target <10s p95)
```

---

## 2. RAG System Deep Dive

### 2.1 Knowledge Base Structure

```
knowledge_base/
├── ddmrp_methodology/
│   ├── 01_buffer_calculation.md          # 3,500 words
│   ├── 02_buffer_zones_explained.md      # 2,800 words
│   ├── 03_demand_planning.md             # 4,200 words
│   ├── 04_lead_time_management.md        # 3,100 words
│   ├── 05_execution_best_practices.md    # 3,900 words
│   └── 06_decoupling_points.md           # 2,600 words
│
├── industry_best_practices/
│   ├── inventory_optimization.md         # 3,500 words
│   ├── supplier_management.md            # 3,200 words
│   ├── cost_reduction_strategies.md      # 2,900 words
│   └── seasonal_demand_handling.md       # 2,700 words
│
├── giia_platform/
│   ├── platform_overview.md              # 2,500 words
│   ├── buffer_management_features.md     # 2,000 words
│   ├── execution_engine_guide.md         # 2,800 words
│   └── troubleshooting_common_issues.md  # 3,100 words
│
└── historical_patterns/
    ├── successful_resolutions/
    │   ├── stockout_prevention_nov2024.md
    │   ├── cost_optimization_dec2024.md
    │   └── seasonal_adjustment_q4_2024.md
    │
    └── common_issues/
        ├── execution_failure_patterns.md
        ├── buffer_drift_causes.md
        └── supplier_delay_handling.md

Total: ~50,000 words across 25 documents
```

### 2.2 Document Chunking Strategy

Each document is split into semantic chunks for efficient retrieval:

```python
# Chunking Configuration (implemented in Go)
CHUNK_SIZE = 500 tokens (~375 words)
CHUNK_OVERLAP = 50 tokens (~37 words)
CHUNK_METHOD = "semantic"  # Split by headings, then paragraphs

# Example: buffer_calculation.md chunking
Original Document: 3,500 words = ~4,500 tokens

Chunks Generated:
1. Introduction + Overview (tokens 0-500)
2. Zone Calculation Formulas (tokens 450-950) [50 overlap]
3. Green Zone Details (tokens 900-1400)
4. Yellow Zone Details (tokens 1350-1850)
5. Red Zone Details (tokens 1800-2300)
6. Practical Examples (tokens 2250-2750)
7. Common Mistakes (tokens 2700-3200)
8. Advanced Topics (tokens 3150-3650)
9. Conclusion + Resources (tokens 3600-4100)

Total Chunks: 9 chunks per document
Total Chunks in System: ~225 chunks
```

### 2.3 Embedding Generation

```go
// Embedding Configuration
type EmbeddingConfig struct {
    Provider   string  // "openai"
    Model      string  // "text-embedding-3-small"
    Dimensions int     // 1536
    BatchSize  int     // 100 chunks per API call
}

// Cost Estimation
// OpenAI text-embedding-3-small: $0.02 per 1M tokens
// Total tokens for knowledge base: ~112,500 (225 chunks × 500 tokens)
// One-time embedding cost: $0.00225 (~$0.002)
// Re-embedding on updates: ~$0.002 per full update
```

### 2.4 ChromaDB Collection Schema

```go
type ChromaCollection struct {
    Name     string                 // "giia_ddmrp_knowledge"
    Metadata map[string]interface{} // Collection metadata
}

type ChromaDocument struct {
    ID        string                 // "buffer_calculation_chunk_3"
    Document  string                 // Chunk text content
    Embedding []float32              // 1536-dimensional vector
    Metadata  map[string]interface{} // Document metadata
}

// Metadata Schema
type DocumentMetadata struct {
    Source       string   // "ddmrp_methodology/buffer_calculation.md"
    Category     string   // "ddmrp_methodology"
    ChunkIndex   int      // 3
    TotalChunks  int      // 9
    Title        string   // "Buffer Calculation - Green Zone Details"
    Tags         []string // ["buffer", "calculation", "green_zone"]
    LastUpdated  string   // "2025-12-23T00:00:00Z"
    Version      string   // "1.0"
    WordCount    int      // 375
    Confidence   float64  // 1.0 (manual review score)
}
```

### 2.5 Query Strategy

```go
type RAGQueryStrategy struct {
    // Step 1: Build contextual query from event
    BuildQuery func(event *Event, context map[string]interface{}) string

    // Step 2: Perform similarity search
    TopK       int     // 5 chunks
    MinScore   float64 // 0.7 similarity threshold

    // Step 3: Re-rank results (optional)
    ReRank     bool    // true - use cross-encoder
    MaxResults int     // 3 final chunks after re-ranking
}

// Example Query Building
func buildStockoutQuery(event *Event, ctx map[string]interface{}) string {
    productName := ctx["product_name"].(string)
    daysToStockout := ctx["days_to_stockout"].(float64)

    // Semantic query construction
    query := fmt.Sprintf(
        "DDMRP buffer management stockout prevention for product %s. "+
        "Critical situation with %.1f days until stockout. "+
        "Need emergency ordering procedures, supplier selection criteria, "+
        "and buffer recalculation methodology based on DDMRP best practices.",
        productName,
        daysToStockout,
    )

    return query
}
```

### 2.6 RAG Retrieval Performance

```
Benchmark Results (Expected):

ChromaDB Query (Similarity Search):
• 50-100ms p50
• 150-200ms p95
• 300-400ms p99

Embedding Generation (Query):
• OpenAI API: 100-200ms

Total RAG Retrieval Time:
• 150-300ms p50
• 350-600ms p95
• 700-1000ms p99

Caching Strategy:
• Cache frequent queries (1-hour TTL)
• Cache hit rate target: 40-60%
• Cached query time: <5ms
```

---

## 3. AI Analysis Patterns

### 3.1 Prompt Engineering Templates

#### Template 1: Stockout Risk Analysis

```go
const StockoutAnalysisPromptTemplate = `
You are an expert DDMRP (Demand Driven Material Requirements Planning) supply chain consultant.
You're analyzing a real-time event from the GIIA platform to help prevent supply chain disruptions.

ROLE:
Act as a senior supply chain advisor with 15+ years of DDMRP expertise.
Your goal is to provide actionable, specific recommendations that prevent stockouts while optimizing costs.

DDMRP METHODOLOGY KNOWLEDGE:
The following documents contain DDMRP best practices and methodologies:

{{range .DDMRPKnowledge}}
---
{{.}}
---
{{end}}

CURRENT SITUATION:
Event Type: {{.Event.Type}}
Product: {{.Context.Product.Name}} (SKU: {{.Context.Product.SKU}})
Organization: {{.Context.Organization.Name}}

BUFFER STATUS:
• Current On-Hand Stock: {{.Context.Buffer.CurrentStock}} units
• Minimum Buffer (Red Zone): {{.Context.Buffer.MinBuffer}} units
• Maximum Buffer (Green Zone): {{.Context.Buffer.MaxBuffer}} units
• Top of Yellow: {{.Context.Buffer.TopOfYellow}} units
• Top of Red: {{.Context.Buffer.TopOfRed}} units
• Current Zone: {{.Context.Buffer.CurrentZone}}
• Daily Consumption (ADU): {{.Context.Buffer.DailyConsumption}} units/day
• Days Until Stockout: {{.Context.Analysis.DaysUntilStockout}} days

SUPPLIER & ORDERING:
• Primary Supplier: {{.Context.Supplier.Name}}
• Lead Time: {{.Context.Supplier.LeadTime}} days
• Minimum Order Quantity: {{.Context.Supplier.MinOrderQty}} units
• Unit Price: ${{.Context.Product.Price}}
• Pending Orders: {{len .Context.PendingOrders}} orders
{{range .Context.PendingOrders}}
  - Order #{{.ID}}: {{.Quantity}} units, Expected: {{.ExpectedDate}}
{{end}}

FINANCIAL IMPACT:
• Product Value per Unit: ${{.Context.Product.Price}}
• Total Stock Value: ${{.Context.Analysis.StockValue}}
• Revenue at Risk (if stockout): ${{.Context.Analysis.RevenueAtRisk}}
• Cost of Emergency Order: ${{.Context.Analysis.EmergencyCost}}

TASK:
Analyze this buffer situation and provide recommendations following DDMRP methodology.

Respond in the following JSON format:
{
  "summary": "A clear, concise 1-2 sentence summary for busy executives",
  "full_analysis": "Detailed multi-paragraph analysis covering:
                    1. Current state assessment
                    2. Root cause analysis (why we're below minimum)
                    3. Risk evaluation (immediate vs medium-term)
                    4. Impact on operations and revenue
                    Use natural, conversational language.",
  "reasoning": "Explain the 'why' behind your analysis using DDMRP principles.
                Reference specific methodology from the knowledge base.",
  "recommendations": [
    {
      "action": "Specific action to take (be concrete, not vague)",
      "reasoning": "Why this action is recommended",
      "expected_outcome": "What will happen if this action is taken",
      "effort": "low|medium|high - implementation difficulty",
      "impact": "low|medium|high - business impact",
      "priority_order": 1
    }
  ],
  "impact_assessment": {
    "risk_level": "low|medium|high|critical",
    "revenue_impact": 15000.00,
    "cost_impact": 500.00,
    "time_to_impact_hours": 36,
    "affected_orders": 5,
    "affected_products": 1
  },
  "confidence": 0.95  // Your confidence in this analysis (0-1)
}

IMPORTANT:
• Be specific with numbers and dates
• Reference DDMRP methodology explicitly
• Prioritize preventing stockouts over cost optimization
• Consider lead times in all recommendations
• Account for pending orders in calculations
• Provide 2-4 recommendations, ranked by priority
`

#### Template 2: Cost Optimization Analysis

```go
const CostOptimizationPromptTemplate = `
You are a supply chain cost optimization expert specializing in DDMRP methodology.

DDMRP KNOWLEDGE:
{{range .DDMRPKnowledge}}
{{.}}
{{end}}

COST OPPORTUNITY DETECTED:
Product: {{.Context.Product.Name}}
Current Ordering Pattern:
• Frequency: {{.Context.Pattern.OrderFrequency}} orders per month
• Average Order Size: {{.Context.Pattern.AvgOrderSize}} units
• Current Supplier: {{.Context.Supplier.Name}}
• Current Unit Price: ${{.Context.Supplier.UnitPrice}}

ALTERNATIVE IDENTIFIED:
• Alternative Supplier: {{.Context.Alternative.Name}}
• Alternative Price: ${{.Context.Alternative.UnitPrice}}
• Lead Time: {{.Context.Alternative.LeadTime}} days (current: {{.Context.Supplier.LeadTime}})
• Reliability: {{.Context.Alternative.Reliability}}% (current: {{.Context.Supplier.Reliability}}%)

FINANCIAL ANALYSIS:
• Monthly Volume: {{.Context.Analysis.MonthlyVolume}} units
• Current Monthly Cost: ${{.Context.Analysis.CurrentMonthlyCost}}
• Projected Savings: ${{.Context.Analysis.ProjectedSavings}}/month

TASK:
Analyze this cost optimization opportunity considering:
1. Total Cost of Ownership (not just unit price)
2. DDMRP buffer implications (longer lead time = larger buffer)
3. Risk vs reward tradeoff
4. Implementation complexity

Provide a balanced recommendation in JSON format:
{
  "summary": "Executive summary of the opportunity",
  "full_analysis": "Comprehensive cost-benefit analysis",
  "reasoning": "DDMRP-based reasoning for buffer adjustments",
  "recommendations": [
    {
      "action": "Specific recommendation (trial, switch, or no action)",
      "reasoning": "Detailed justification",
      "expected_outcome": "Projected results with numbers",
      "effort": "low|medium|high",
      "impact": "low|medium|high",
      "priority_order": 1
    }
  ],
  "impact_assessment": {
    "risk_level": "low|medium|high",
    "revenue_impact": 0,
    "cost_impact": -1050.00,  // Negative = savings
    "time_to_impact_hours": 720,  // 30 days
    "affected_orders": 12,
    "affected_products": 1
  },
  "confidence": 0.85
}

Include:
• Buffer recalculation requirements
• Transitional risks
• Recommended trial period
• Success metrics
`
```

### 3.2 Response Parsing & Validation

```go
type AIResponseValidator struct {
    RequiredFields []string
    MaxRetries     int
    FallbackLogic  func(*Event) *AINotification
}

func (v *AIResponseValidator) Validate(response string) (*AIAnalysisResponse, error) {
    var parsed AIAnalysisResponse

    // Parse JSON
    if err := json.Unmarshal([]byte(response), &parsed); err != nil {
        return nil, fmt.Errorf("invalid JSON: %w", err)
    }

    // Validate required fields
    if parsed.Summary == "" {
        return nil, errors.New("missing summary")
    }
    if len(parsed.Recommendations) == 0 {
        return nil, errors.New("no recommendations provided")
    }

    // Validate confidence score
    if parsed.Confidence < 0 || parsed.Confidence > 1 {
        return nil, errors.New("confidence must be between 0 and 1")
    }

    // Validate risk level
    validRiskLevels := map[string]bool{
        "low": true, "medium": true, "high": true, "critical": true,
    }
    if !validRiskLevels[parsed.ImpactAssessment.RiskLevel] {
        return nil, errors.New("invalid risk level")
    }

    // Validate recommendations
    for i, rec := range parsed.Recommendations {
        if rec.Action == "" {
            return nil, fmt.Errorf("recommendation %d missing action", i)
        }
        if rec.Effort != "low" && rec.Effort != "medium" && rec.Effort != "high" {
            return nil, fmt.Errorf("recommendation %d invalid effort level", i)
        }
    }

    return &parsed, nil
}
```

### 3.3 Confidence Thresholds

```go
type ConfidencePolicy struct {
    MinimumConfidence float64 // 0.7
    Actions           map[float64]Action
}

// Example policy
var DefaultConfidencePolicy = ConfidencePolicy{
    MinimumConfidence: 0.7,
    Actions: map[float64]Action{
        0.95: Action{
            Name:   "Auto-execute",
            Desc:   "Very high confidence - can auto-execute low-risk actions",
            Notify: true,
        },
        0.80: Action{
            Name:   "Standard notification",
            Desc:   "High confidence - send notification with recommendations",
            Notify: true,
        },
        0.70: Action{
            Name:   "Notify with caveat",
            Desc:   "Moderate confidence - notify but flag uncertainty",
            Notify: true,
        },
        0.00: Action{
            Name:   "Escalate to human",
            Desc:   "Low confidence - escalate for manual review",
            Notify: false,
        },
    },
}

func determineAction(confidence float64) Action {
    switch {
    case confidence >= 0.95:
        return DefaultConfidencePolicy.Actions[0.95]
    case confidence >= 0.80:
        return DefaultConfidencePolicy.Actions[0.80]
    case confidence >= 0.70:
        return DefaultConfidencePolicy.Actions[0.70]
    default:
        return DefaultConfidencePolicy.Actions[0.00]
    }
}
```

---

## 4. Event Processing Strategies

### 4.1 Event Deduplication

Problem: Same event published multiple times due to retries or race conditions.

```go
type EventDeduplicator struct {
    cache *redis.Client
    ttl   time.Duration  // 5 minutes
}

func (ed *EventDeduplicator) IsDuplicate(ctx context.Context, event *Event) (bool, error) {
    key := fmt.Sprintf("event:processed:%s", event.ID)

    // Try to set key with NX (only if not exists)
    result, err := ed.cache.SetNX(ctx, key, "1", ed.ttl).Result()
    if err != nil {
        return false, err
    }

    // If result is false, key already exists = duplicate
    return !result, nil
}

// Usage in event handler
func (h *EventHandler) Handle(ctx context.Context, event *Event) error {
    isDuplicate, err := h.deduplicator.IsDuplicate(ctx, event)
    if err != nil {
        h.logger.Warn(ctx, "Deduplication check failed, processing anyway", nil)
    } else if isDuplicate {
        h.logger.Debug(ctx, "Duplicate event detected, skipping", logger.Tags{
            "event_id": event.ID,
        })
        return nil  // Skip duplicate
    }

    // Process event...
}
```

### 4.2 Pattern Detection

Detect patterns across multiple events (e.g., 3+ execution failures from same supplier).

```go
type PatternDetector struct {
    eventStore     EventStore
    patternRules   []PatternRule
    windowDuration time.Duration  // 6 hours
}

type PatternRule struct {
    Name          string
    EventTypes    []string
    MinOccurrence int
    Window        time.Duration
    Matcher       func(events []*Event) bool
    Priority      NotificationPriority
}

// Example: Execution Failure Pattern
var ExecutionFailurePattern = PatternRule{
    Name:          "repeated_execution_failures",
    EventTypes:    []string{"order.execution_failed"},
    MinOccurrence: 3,
    Window:        6 * time.Hour,
    Matcher: func(events []*Event) bool {
        // Check if all failures are from same supplier
        supplierID := events[0].Data["supplier_id"]
        for _, e := range events[1:] {
            if e.Data["supplier_id"] != supplierID {
                return false
            }
        }
        return true
    },
    Priority: NotificationPriorityHigh,
}

func (pd *PatternDetector) DetectPatterns(ctx context.Context, event *Event) ([]Pattern, error) {
    var detectedPatterns []Pattern

    for _, rule := range pd.patternRules {
        // Check if event type matches
        if !contains(rule.EventTypes, event.Type) {
            continue
        }

        // Get recent events of same type
        recentEvents, err := pd.eventStore.GetRecentEvents(
            ctx,
            event.OrganizationID,
            rule.EventTypes,
            rule.Window,
        )
        if err != nil {
            return nil, err
        }

        // Check if pattern threshold met
        if len(recentEvents) < rule.MinOccurrence {
            continue
        }

        // Run matcher function
        if rule.Matcher(recentEvents) {
            detectedPatterns = append(detectedPatterns, Pattern{
                Rule:   rule,
                Events: recentEvents,
            })
        }
    }

    return detectedPatterns, nil
}
```

### 4.3 Event Buffering & Batching

For high-volume events, batch processing improves efficiency.

```go
type EventBatcher struct {
    buffer    chan *Event
    batchSize int           // 10 events
    timeout   time.Duration // 5 seconds
    processor func([]*Event) error
}

func (eb *EventBatcher) Start(ctx context.Context) {
    go func() {
        batch := make([]*Event, 0, eb.batchSize)
        ticker := time.NewTicker(eb.timeout)
        defer ticker.Stop()

        for {
            select {
            case event := <-eb.buffer:
                batch = append(batch, event)

                // Process when batch full
                if len(batch) >= eb.batchSize {
                    eb.processBatch(batch)
                    batch = make([]*Event, 0, eb.batchSize)
                }

            case <-ticker.C:
                // Process on timeout (even if not full)
                if len(batch) > 0 {
                    eb.processBatch(batch)
                    batch = make([]*Event, 0, eb.batchSize)
                }

            case <-ctx.Done():
                // Graceful shutdown - process remaining
                if len(batch) > 0 {
                    eb.processBatch(batch)
                }
                return
            }
        }
    }()
}

func (eb *EventBatcher) Add(event *Event) {
    eb.buffer <- event
}
```

---

## 5. Notification Routing Logic

### 5.1 Priority-Based Routing

```go
type NotificationRouter struct {
    channelMatrix map[NotificationPriority][]Channel
}

type Channel string

const (
    ChannelInApp    Channel = "in_app"
    ChannelEmail    Channel = "email"
    ChannelSlack    Channel = "slack"
    ChannelSMS      Channel = "sms"
    ChannelWebhook  Channel = "webhook"
)

// Default routing matrix
var DefaultChannelMatrix = map[NotificationPriority][]Channel{
    NotificationPriorityCritical: {
        ChannelInApp,
        ChannelEmail,
        ChannelSMS,
        ChannelSlack,
    },
    NotificationPriorityHigh: {
        ChannelInApp,
        ChannelEmail,
        ChannelSlack,
    },
    NotificationPriorityMedium: {
        ChannelInApp,
        ChannelEmail,
    },
    NotificationPriorityLow: {
        ChannelInApp,
    },
}

func (nr *NotificationRouter) Route(
    ctx context.Context,
    notification *AINotification,
    userPrefs *UserNotificationPreferences,
) []ChannelDelivery {
    // Get default channels for priority
    defaultChannels := nr.channelMatrix[notification.Priority]

    var deliveries []ChannelDelivery

    for _, channel := range defaultChannels {
        // Check user preferences
        if !nr.isChannelEnabled(channel, userPrefs) {
            continue
        }

        // Check priority threshold
        if !nr.meetsPriorityThreshold(channel, notification.Priority, userPrefs) {
            continue
        }

        // Check quiet hours (except critical)
        if notification.Priority != NotificationPriorityCritical {
            if nr.isQuietHours(userPrefs) {
                continue
            }
        }

        // Check rate limits
        if !nr.checkRateLimit(ctx, userPrefs.UserID, channel) {
            continue
        }

        deliveries = append(deliveries, ChannelDelivery{
            Channel:      channel,
            Notification: notification,
            UserPrefs:    userPrefs,
        })
    }

    return deliveries
}

func (nr *NotificationRouter) isChannelEnabled(
    channel Channel,
    prefs *UserNotificationPreferences,
) bool {
    switch channel {
    case ChannelInApp:
        return prefs.EnableInApp
    case ChannelEmail:
        return prefs.EnableEmail
    case ChannelSMS:
        return prefs.EnableSMS
    case ChannelSlack:
        return prefs.EnableSlack
    default:
        return false
    }
}
```

### 5.2 Rate Limiting

```go
type RateLimiter struct {
    cache *redis.Client
}

func (rl *RateLimiter) CheckLimit(
    ctx context.Context,
    userID uuid.UUID,
    channel Channel,
    maxPerHour int,
) (bool, error) {
    key := fmt.Sprintf("rate_limit:%s:%s:%s",
        userID,
        channel,
        time.Now().Format("2006-01-02-15"),  // Hourly bucket
    )

    // Increment counter
    count, err := rl.cache.Incr(ctx, key).Result()
    if err != nil {
        return false, err
    }

    // Set expiry on first increment
    if count == 1 {
        rl.cache.Expire(ctx, key, 1*time.Hour)
    }

    // Check if under limit
    return count <= int64(maxPerHour), nil
}
```

### 5.3 Quiet Hours Logic

```go
func isQuietHours(prefs *UserNotificationPreferences) bool {
    if prefs.QuietHoursStart == nil || prefs.QuietHoursEnd == nil {
        return false  // No quiet hours configured
    }

    // Get current time in user's timezone
    loc, err := time.LoadLocation(prefs.Timezone)
    if err != nil {
        loc = time.UTC  // Fallback to UTC
    }

    now := time.Now().In(loc)
    currentTime := now.Format("15:04")

    start := prefs.QuietHoursStart.Format("15:04")
    end := prefs.QuietHoursEnd.Format("15:04")

    // Handle overnight quiet hours (e.g., 22:00 - 07:00)
    if start > end {
        return currentTime >= start || currentTime < end
    }

    // Handle same-day quiet hours (e.g., 13:00 - 14:00)
    return currentTime >= start && currentTime < end
}
```

---

## 6. Testing Strategy

### 6.1 Unit Test Coverage

```
Target Coverage: 80%+

Critical Components (90%+ coverage):
├── Event Processing
│   ├── Event Router
│   ├── Event Handlers
│   └── Pattern Detector
│
├── AI Analysis
│   ├── Prompt Builder
│   ├── Response Parser
│   └── Confidence Evaluator
│
├── RAG System
│   ├── Knowledge Retriever
│   ├── Query Builder
│   └── Cache Manager
│
└── Notification System
    ├── Priority Router
    ├── Channel Delivery
    └── Rate Limiter
```

### 6.2 Integration Test Scenarios

```go
// Test 1: End-to-End Stockout Prevention
func TestE2E_StockoutPrevention(t *testing.T) {
    // Given: Complete system setup
    testEnv := setupTestEnvironment(t)
    defer testEnv.Teardown()

    // Seed data
    product := createTestProduct(t, testEnv.DB)
    buffer := createTestBuffer(t, testEnv.DB, BufferBelowMinimum)
    user := createTestUser(t, testEnv.DB, WithEmailNotifications)

    // When: Buffer below minimum event published
    event := &Event{
        Type:           "buffer.below_minimum",
        OrganizationID: product.OrganizationID,
        Data: map[string]interface{}{
            "product_id": product.ID,
            "current_stock": 15.0,
            "min_buffer": 50.0,
        },
    }

    err := testEnv.EventPublisher.Publish(context.Background(), event)
    require.NoError(t, err)

    // Then: Notification delivered within 10 seconds
    notification := waitForNotification(t, testEnv.DB, user.ID, 10*time.Second)
    assert.NotNil(t, notification)
    assert.Equal(t, NotificationTypeAlert, notification.Type)
    assert.Equal(t, NotificationPriorityCritical, notification.Priority)
    assert.Contains(t, notification.Title, product.Name)
    assert.GreaterOrEqual(t, len(notification.Recommendations), 2)

    // Verify email sent
    emailSent := testEnv.EmailMock.WasSent(user.Email)
    assert.True(t, emailSent)

    // Verify RAG knowledge used
    assert.Contains(t, notification.Reasoning, "DDMRP")
}
```

### 6.3 Load Testing

```yaml
# k6 load test script
scenarios:
  event_processing:
    executor: ramping-vus
    startVUs: 0
    stages:
      - duration: 2m
        target: 50    # Ramp to 50 concurrent publishers
      - duration: 5m
        target: 100   # Sustained load
      - duration: 2m
        target: 0     # Ramp down

    exec: publishEvents

thresholds:
  event_processing_duration:
    - p(95) < 1000   # 95% under 1 second
    - p(99) < 3000   # 99% under 3 seconds

  ai_analysis_duration:
    - p(95) < 5000   # 95% under 5 seconds

  notification_delivery:
    - p(95) < 10000  # 95% under 10 seconds

  http_req_failed:
    - rate < 0.01    # Error rate <1%
```

---

## 7. Deployment Guide

### 7.1 Infrastructure Requirements

```yaml
# Kubernetes Deployment Spec
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ai-intelligence-hub
  namespace: giia-platform
spec:
  replicas: 3  # HA setup
  selector:
    matchLabels:
      app: ai-intelligence-hub
  template:
    spec:
      containers:
      - name: intelligence-hub
        image: giia/ai-intelligence-hub:v1.0.0

        resources:
          requests:
            memory: "512Mi"
            cpu: "500m"
          limits:
            memory: "1Gi"
            cpu: "1000m"

        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: intelligence-hub-secrets
              key: database-url

        - name: NATS_SERVERS
          value: "nats://nats.giia-platform.svc.cluster.local:4222"

        - name: CLAUDE_API_KEY
          valueFrom:
            secretKeyRef:
              name: intelligence-hub-secrets
              key: claude-api-key

        - name: CHROMADB_HOST
          value: "chromadb.giia-platform.svc.cluster.local"

        - name: REDIS_URL
          valueFrom:
            secretKeyRef:
              name: intelligence-hub-secrets
              key: redis-url

        ports:
        - containerPort: 8080
          name: http
        - containerPort: 9090
          name: grpc

        livenessProbe:
          httpGet:
            path: /health/live
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10

        readinessProbe:
          httpGet:
            path: /health/ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5

---
# ChromaDB StatefulSet
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: chromadb
  namespace: giia-platform
spec:
  serviceName: chromadb
  replicas: 1
  selector:
    matchLabels:
      app: chromadb
  template:
    spec:
      containers:
      - name: chromadb
        image: chromadb/chroma:0.4.24
        ports:
        - containerPort: 8000
        volumeMounts:
        - name: chromadb-data
          mountPath: /chroma/chroma

  volumeClaimTemplates:
  - metadata:
      name: chromadb-data
    spec:
      accessModes: ["ReadWriteOnce"]
      resources:
        requests:
          storage: 10Gi
```

### 7.2 Environment Configuration

```bash
# .env.production
SERVICE_NAME=ai-intelligence-hub
SERVICE_VERSION=1.0.0
ENVIRONMENT=production

# Database
DATABASE_URL=postgresql://user:pass@postgres.giia.internal:5432/intelligence_hub?sslmode=require
DATABASE_MAX_CONNECTIONS=25
DATABASE_MAX_IDLE_CONNECTIONS=5

# NATS
NATS_SERVERS=nats://nats-1.giia.internal:4222,nats://nats-2.giia.internal:4222,nats://nats-3.giia.internal:4222
NATS_CONSUMER_NAME=ai-intelligence-hub-prod
NATS_MAX_DELIVER=5
NATS_ACK_WAIT=30s

# Claude API
CLAUDE_API_KEY=sk-ant-api03-xxxxx
CLAUDE_MODEL=claude-3-5-sonnet-20250929
CLAUDE_MAX_TOKENS=2000
CLAUDE_TEMPERATURE=0.3
CLAUDE_TIMEOUT=30s

# ChromaDB (RAG)
CHROMADB_HOST=chromadb.giia.internal
CHROMADB_PORT=8000
CHROMADB_COLLECTION=giia_ddmrp_knowledge
CHROMADB_EMBEDDING_MODEL=text-embedding-3-small

# Redis (Caching)
REDIS_URL=redis://:password@redis.giia.internal:6379/0
REDIS_MAX_RETRIES=3
REDIS_POOL_SIZE=10

# Notification Channels
SENDGRID_API_KEY=SG.xxxxx
SENDGRID_FROM_EMAIL=notifications@giia.com
SENDGRID_FROM_NAME=GIIA Intelligence Hub

TWILIO_ACCOUNT_SID=ACxxxxx
TWILIO_AUTH_TOKEN=xxxxx
TWILIO_FROM_NUMBER=+1234567890

SLACK_BOT_TOKEN=xoxb-xxxxx

# Performance Tuning
EVENT_WORKER_POOL_SIZE=20
EVENT_BUFFER_SIZE=1000
AI_ANALYSIS_TIMEOUT=30s
NOTIFICATION_DELIVERY_TIMEOUT=10s

# Rate Limiting
MAX_ALERTS_PER_HOUR=10
MAX_EMAILS_PER_DAY=50
MAX_SMS_PER_DAY=10

# Feature Flags
ENABLE_PATTERN_DETECTION=true
ENABLE_DAILY_DIGEST=true
ENABLE_AUTO_RESOLUTION=false  # Future feature

# Monitoring
PROMETHEUS_PORT=9091
METRICS_PATH=/metrics
LOG_LEVEL=info
LOG_FORMAT=json
```

### 7.3 Database Migration

```bash
# Apply migrations in production
migrate -path=services/ai-intelligence-hub/migrations \
        -database="postgresql://user:pass@postgres:5432/intelligence_hub?sslmode=require" \
        up

# Rollback if needed
migrate -path=services/ai-intelligence-hub/migrations \
        -database="postgresql://user:pass@postgres:5432/intelligence_hub?sslmode=require" \
        down 1
```

### 7.4 Knowledge Base Initialization

```bash
# Step 1: Chunk documents
go run cmd/knowledge-loader/main.go \
  --source=knowledge_base/ \
  --chunk-size=500 \
  --chunk-overlap=50 \
  --output=chunks.jsonl

# Step 2: Generate embeddings
go run cmd/knowledge-loader/main.go \
  --mode=embed \
  --input=chunks.jsonl \
  --provider=openai \
  --model=text-embedding-3-small

# Step 3: Load into ChromaDB
go run cmd/knowledge-loader/main.go \
  --mode=load \
  --input=chunks_embedded.jsonl \
  --chromadb-host=chromadb.giia.internal \
  --collection=giia_ddmrp_knowledge

# Step 4: Verify
go run cmd/knowledge-loader/main.go \
  --mode=test \
  --query="buffer below minimum DDMRP methodology" \
  --top-k=5
```

---

## 8. Monitoring & Observability

### 8.1 Key Metrics

```go
// Prometheus metrics
var (
    // Event processing
    eventsProcessedTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "intelligence_hub_events_processed_total",
            Help: "Total events processed by type",
        },
        []string{"event_type", "status"},  // status: success, error
    )

    eventProcessingDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "intelligence_hub_event_processing_duration_seconds",
            Help:    "Event processing duration",
            Buckets: []float64{0.1, 0.5, 1.0, 2.0, 5.0, 10.0},
        },
        []string{"event_type"},
    )

    // AI Analysis
    aiAnalysisTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "intelligence_hub_ai_analysis_total",
            Help: "Total AI analyses performed",
        },
        []string{"analysis_type", "status"},
    )

    aiAnalysisDuration = promauto.NewHistogram(
        prometheus.HistogramOpts{
            Name:    "intelligence_hub_ai_analysis_duration_seconds",
            Help:    "AI analysis duration",
            Buckets: []float64{1.0, 3.0, 5.0, 10.0, 15.0, 30.0},
        },
    )

    aiAnalysisConfidence = promauto.NewHistogram(
        prometheus.HistogramOpts{
            Name:    "intelligence_hub_ai_confidence_score",
            Help:    "AI analysis confidence scores",
            Buckets: []float64{0.5, 0.6, 0.7, 0.8, 0.9, 0.95, 1.0},
        },
    )

    // RAG
    ragQueryTotal = promauto.NewCounter(
        prometheus.CounterOpts{
            Name: "intelligence_hub_rag_queries_total",
            Help: "Total RAG knowledge retrieval queries",
        },
    )

    ragCacheHitRate = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "intelligence_hub_rag_cache_hit_rate",
            Help: "RAG query cache hit rate",
        },
    )

    // Notifications
    notificationsCreatedTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "intelligence_hub_notifications_created_total",
            Help: "Total notifications created",
        },
        []string{"type", "priority"},
    )

    notificationsDeliveredTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "intelligence_hub_notifications_delivered_total",
            Help: "Total notifications delivered",
        },
        []string{"channel", "status"},  // status: success, failed, rate_limited
    )

    notificationDeliveryDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "intelligence_hub_notification_delivery_duration_seconds",
            Help:    "Notification delivery duration by channel",
            Buckets: []float64{0.5, 1.0, 2.0, 5.0, 10.0},
        },
        []string{"channel"},
    )

    // Cost tracking
    claudeAPICallsTotal = promauto.NewCounter(
        prometheus.CounterOpts{
            Name: "intelligence_hub_claude_api_calls_total",
            Help: "Total Claude API calls made",
        },
    )

    claudeTokensUsedTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "intelligence_hub_claude_tokens_used_total",
            Help: "Total Claude API tokens consumed",
        },
        []string{"type"},  // type: input, output
    )

    claudeEstimatedCostUSD = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "intelligence_hub_claude_estimated_cost_usd",
            Help: "Estimated Claude API cost in USD (monthly)",
        },
    )
)
```

### 8.2 Health Checks

```go
type HealthChecker struct {
    db          *sql.DB
    nats        *nats.Conn
    redis       *redis.Client
    chromaDB    *chroma.Client
    claudeAPI   *anthropic.Client
}

func (hc *HealthChecker) LivenessCheck(ctx context.Context) error {
    // Simple check - is the service running?
    return nil
}

func (hc *HealthChecker) ReadinessCheck(ctx context.Context) error {
    // Comprehensive check - can we serve traffic?

    checks := []struct {
        name string
        fn   func(context.Context) error
    }{
        {"database", hc.checkDatabase},
        {"nats", hc.checkNATS},
        {"redis", hc.checkRedis},
        {"chromadb", hc.checkChromaDB},
    }

    for _, check := range checks {
        if err := check.fn(ctx); err != nil {
            return fmt.Errorf("%s check failed: %w", check.name, err)
        }
    }

    return nil
}

func (hc *HealthChecker) checkDatabase(ctx context.Context) error {
    ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
    defer cancel()

    return hc.db.PingContext(ctx)
}

func (hc *HealthChecker) checkNATS(ctx context.Context) error {
    if !hc.nats.IsConnected() {
        return errors.New("NATS not connected")
    }
    return nil
}

// Health endpoint
func (s *Server) HealthHandler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    if err := s.healthChecker.ReadinessCheck(ctx); err != nil {
        w.WriteHeader(http.StatusServiceUnavailable)
        json.NewEncoder(w).Encode(map[string]interface{}{
            "status": "unhealthy",
            "error":  err.Error(),
        })
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status": "healthy",
        "timestamp": time.Now(),
    })
}
```

### 8.3 Alerting Rules

```yaml
# Prometheus alerting rules
groups:
  - name: intelligence_hub
    interval: 30s
    rules:
      # High error rate
      - alert: HighEventProcessingErrorRate
        expr: |
          sum(rate(intelligence_hub_events_processed_total{status="error"}[5m]))
          /
          sum(rate(intelligence_hub_events_processed_total[5m]))
          > 0.05
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High event processing error rate"
          description: "Error rate is {{ $value | humanizePercentage }} over the last 5 minutes"

      # Slow event processing
      - alert: SlowEventProcessing
        expr: |
          histogram_quantile(0.95,
            rate(intelligence_hub_event_processing_duration_seconds_bucket[5m])
          ) > 1.0
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "Event processing is slow"
          description: "P95 latency is {{ $value }}s (target: <1s)"

      # AI analysis timeout
      - alert: AIAnalysisTimeout
        expr: |
          histogram_quantile(0.95,
            rate(intelligence_hub_ai_analysis_duration_seconds_bucket[5m])
          ) > 30.0
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "AI analysis is timing out"
          description: "P95 AI analysis time is {{ $value }}s"

      # Low AI confidence
      - alert: LowAIConfidence
        expr: |
          histogram_quantile(0.50,
            rate(intelligence_hub_ai_confidence_score_bucket[1h])
          ) < 0.7
        for: 30m
        labels:
          severity: warning
        annotations:
          summary: "AI confidence scores are low"
          description: "Median confidence is {{ $value }} (target: >0.7)"

      # High Claude API cost
      - alert: HighClaudeAPICost
        expr: intelligence_hub_claude_estimated_cost_usd > 500
        for: 1h
        labels:
          severity: warning
        annotations:
          summary: "Claude API costs are high"
          description: "Estimated monthly cost: ${{ $value }}"

      # Notification delivery failures
      - alert: NotificationDeliveryFailures
        expr: |
          sum(rate(intelligence_hub_notifications_delivered_total{status="failed"}[5m]))
          /
          sum(rate(intelligence_hub_notifications_delivered_total[5m]))
          > 0.1
        for: 10m
        labels:
          severity: critical
        annotations:
          summary: "High notification delivery failure rate"
          description: "{{ $value | humanizePercentage }} of notifications are failing"

      # NATS disconnected
      - alert: NATSDisconnected
        expr: up{job="intelligence-hub"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Intelligence Hub is down"
          description: "Service has been down for 1 minute"
```

### 8.4 Logging Strategy

```go
// Structured logging with correlation
type LogContext struct {
    TraceID        string
    SpanID         string
    EventID        string
    OrganizationID string
    UserID         string
    EventType      string
}

func (lc *LogContext) Tags() logger.Tags {
    return logger.Tags{
        "trace_id":        lc.TraceID,
        "span_id":         lc.SpanID,
        "event_id":        lc.EventID,
        "organization_id": lc.OrganizationID,
        "user_id":         lc.UserID,
        "event_type":      lc.EventType,
    }
}

// Usage in handlers
func (h *BufferEventHandler) Handle(ctx context.Context, event *Event) error {
    logCtx := &LogContext{
        TraceID:        generateTraceID(),
        EventID:        event.ID,
        OrganizationID: event.OrganizationID,
        EventType:      event.Type,
    }

    h.logger.Info(ctx, "Processing buffer event", logCtx.Tags())

    // Process...

    h.logger.Info(ctx, "Buffer event processed successfully", logger.Tags{
        "trace_id":   logCtx.TraceID,
        "event_id":   logCtx.EventID,
        "duration_ms": time.Since(startTime).Milliseconds(),
    })

    return nil
}
```

---

## Summary

This implementation guide provides:

1. **Complete architecture** with detailed component interactions
2. **RAG system** design with chunking, embedding, and retrieval strategies
3. **AI analysis patterns** with prompt templates and validation
4. **Event processing** strategies including deduplication and pattern detection
5. **Notification routing** with priority-based, rate-limited delivery
6. **Comprehensive testing** approach with unit, integration, and load tests
7. **Production deployment** guide with K8s configs and migrations
8. **Monitoring & observability** with metrics, alerts, and logging

**Next Steps:**
1. Review and approve this guide
2. Begin Phase 1 implementation (Foundation & Database)
3. Set up development environment
4. Create knowledge base documents
5. Implement core components iteratively

**Estimated Timeline:** 6-8 weeks to MVP
**Team Size:** 2-3 engineers
**Dependencies:** NATS JetStream (✅ Complete)