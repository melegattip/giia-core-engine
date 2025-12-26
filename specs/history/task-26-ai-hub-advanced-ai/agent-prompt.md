# Agent Prompt: Task 26 - AI Intelligence Hub Advanced AI Integration

## 🤖 Agent Identity
Expert AI/ML Engineer for LLM integration, RAG systems, and pattern detection with Go.

---

## 📋 Mission
Upgrade AI Intelligence Hub with real Claude API integration, ChromaDB vector RAG, and cross-event pattern detection.

---

## 📂 Files to Create/Modify

### Claude Integration
- `internal/adapters/claude/client.go` - Real API client (replace mock)
- `internal/adapters/claude/prompt_builder.go` - Structured prompts

### Vector RAG System
- `internal/adapters/chromadb/client.go` - ChromaDB integration
- `internal/adapters/embeddings/service.go` - Embedding generation
- `internal/usecases/rag_retrieval.go` - Knowledge retrieval
- `internal/domain/entities/knowledge_document.go`

### Pattern Detection
- `internal/usecases/pattern_detector.go` + `_test.go`
- `internal/domain/entities/pattern.go`

---

## 🔧 Claude API Integration

```go
type ClaudeClient struct {
    apiKey     string
    httpClient *http.Client
    model      string // claude-sonnet-4.5 or claude-haiku
}

func (c *ClaudeClient) Analyze(ctx context.Context, prompt string, context []string) (*AnalysisResult, error) {
    // Build request with system prompt + RAG context
    // Call Anthropic Messages API
    // Parse structured JSON response
    // Handle rate limits with exponential backoff
}
```

### Prompt Engineering
Build prompts with DDMRP context for stockout analysis, recommendations, and impact assessment.

---

## 🔧 ChromaDB RAG System

```go
type ChromaDBRetriever struct {
    client     *chromem.DB
    collection *chromem.Collection
}

func (r *ChromaDBRetriever) Retrieve(ctx context.Context, query string, topK int) ([]Document, error) {
    // Generate embedding for query
    // Perform similarity search
    // Return top K relevant documents
}
```

---

## 🔧 Pattern Detection

```go
type PatternDetector struct {
    eventStore EventStore
    alerter    AlertService
}

func (p *PatternDetector) DetectPatterns(ctx context.Context) error {
    // Pattern 1: Recurring stockouts (same product, 3+ times in 7 days)
    // Pattern 2: Supplier delays (same supplier, 3+ late deliveries)
    // Pattern 3: Demand spikes (unusual demand increases)
    // Generate pattern notifications with root cause analysis
}
```

---

## 🔧 Fallback Logic

When Claude API fails, fall back to rule-based analysis.

---

## ✅ Success Criteria
- [ ] Claude API integration with <2s response
- [ ] ChromaDB indexes all knowledge documents
- [ ] Semantic search >80% accuracy
- [ ] Pattern detection within 1 hour of occurrence
- [ ] 99.9% uptime with graceful degradation
- [ ] 85%+ test coverage

---

## 🚀 Commands
```bash
cd services/ai-intelligence-hub
export ANTHROPIC_API_KEY=your_key
docker run -d -p 8000:8000 chromadb/chroma
go test ./internal/adapters/... -cover
go build -o bin/ai-hub ./cmd/api
```
