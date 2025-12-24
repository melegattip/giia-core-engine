# Agent Prompt: Task 30 - NATS Event Publishing & Subscription

## 🤖 Agent Identity
Expert Go Engineer for event-driven architectures with NATS JetStream, message patterns, and distributed systems.

---

## 📋 Mission
Implement NATS event publishing in Execution and DDMRP services, event subscriptions, and ensure reliable message delivery.

---

## 📂 Files to Create

### Execution Service Events
- `internal/events/publisher.go` - Event publisher
- `internal/events/types.go` - Event type definitions
- `internal/events/subscriber.go` - Event subscriber
- `internal/events/handlers/catalog_handler.go`

### DDMRP Service Events
- `internal/events/publisher.go`
- `internal/events/types.go`

### Shared Event Package
- `shared/events/envelope.go` - Standard wrapper
- `shared/events/subjects.go` - Subject constants

---

## 🔧 Event Envelope

```go
type EventEnvelope struct {
    ID             string          `json:"id"`
    Subject        string          `json:"subject"`
    CorrelationID  string          `json:"correlation_id"`
    OrganizationID string          `json:"organization_id"`
    Timestamp      time.Time       `json:"timestamp"`
    Payload        json.RawMessage `json:"payload"`
}
```

---

## 🔧 Event Publisher

```go
type Publisher struct {
    js nats.JetStreamContext
}

func (p *Publisher) Publish(ctx context.Context, subject string, envelope *EventEnvelope) error {
    // Publish with retry (3 attempts)
}
```

---

## 🔧 Event Types

### Execution Service
- `execution.purchase_order.created`
- `execution.purchase_order.received`
- `execution.sales_order.created`
- `execution.sales_order.shipped`
- `execution.inventory.updated`

### DDMRP Service
- `ddmrp.buffer.calculated`
- `ddmrp.buffer.status_changed`
- `ddmrp.buffer.alert_triggered`

---

## 🔧 Event Catalog

| Subject | Publisher | Consumers | Payload |
|---------|-----------|-----------|---------|
| `execution.purchase_order.created` | Execution | DDMRP, Analytics | PO details |
| `execution.inventory.updated` | Execution | DDMRP, AI Hub | Balance change |
| `ddmrp.buffer.status_changed` | DDMRP | AI Hub, Analytics | Zone transition |
| `ddmrp.buffer.alert_triggered` | DDMRP | AI Hub, Execution | Replenishment |

---

## 🔧 Event Subscription

```go
func (s *EventSubscriber) Subscribe(ctx context.Context) error {
    sub, err := s.js.PullSubscribe("catalog.>", "execution-service")
    // Process messages, ack/nak appropriately
}
```

---

## ✅ Success Criteria
- [ ] All order ops publish events <100ms
- [ ] All buffer changes publish events <100ms
- [ ] Delivery success >99.9%
- [ ] No events lost during brief NATS downtime
- [ ] E2E event processing <500ms
- [ ] 85%+ test coverage

---

## 🚀 Commands
```bash
docker run -d --name nats -p 4222:4222 nats:2.9 -js
nats stream add GIIA --subjects "execution.>" --subjects "ddmrp.>" --subjects "catalog.>"
nats sub "execution.>"
cd services/execution-service
go test ./internal/events/... -cover
```
