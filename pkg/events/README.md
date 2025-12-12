# Events Package

NATS Jetstream event publishing and subscription with CloudEvents-like event structure.

## Features

- NATS Jetstream publisher and subscriber
- CloudEvents-inspired event structure
- Automatic retry with exponential backoff (max 3 retries)
- Durable subscriptions support
- At-least-once delivery guarantees
- Mock implementations for testing

## Installation

```go
import "github.com/giia/giia-core-engine/pkg/events"
```

## Usage

### Connect to NATS

```go
// Basic connection
nc, err := events.ConnectWithDefaults("nats://localhost:4222")
if err != nil {
    log.Fatal(err)
}
defer events.Disconnect(nc)

// Custom connection configuration
config := &events.ConnectionConfig{
    URL:            "nats://nats.example.com:4222",
    MaxReconnects:  10,
    ReconnectWait:  2 * time.Second,
    ConnectionName: "auth-service",
}
nc, err := events.Connect(config)
```

### Publishing Events

```go
publisher, err := events.NewPublisher(nc)
if err != nil {
    log.Fatal(err)
}
defer publisher.Close()

// Create and publish event
event := events.NewEvent(
    "user.created",           // Event type
    "auth-service",           // Source service
    "org-123",               // Organization ID
    map[string]interface{}{  // Event data
        "user_id": 12345,
        "email": "user@example.com",
        "role": "admin",
    },
)

err = publisher.Publish(context.Background(), "users.events", event)
if err != nil {
    log.Printf("Failed to publish event: %v", err)
}
```

### Async Publishing (Fire and Forget)

```go
err = publisher.PublishAsync(ctx, "users.events", event)
```

### Subscribing to Events

```go
subscriber, err := events.NewSubscriber(nc)
if err != nil {
    log.Fatal(err)
}
defer subscriber.Close()

// Define event handler
handler := func(ctx context.Context, event *events.Event) error {
    log.Printf("Received event: %s from %s", event.Type, event.Source)

    // Process event
    if event.Type == "user.created" {
        userID := event.Data["user_id"].(float64)
        return processNewUser(ctx, int64(userID))
    }

    return nil
}

// Subscribe to subject
err = subscriber.Subscribe(context.Background(), "users.events", handler)
```

### Durable Subscriptions

```go
// Create durable subscription that survives restarts
err = subscriber.SubscribeDurable(
    ctx,
    "users.events",           // Subject
    "user-processor-service", // Durable name
    handler,
)
```

### Event Structure

```go
type Event struct {
    ID             string                 // Unique event ID (UUID)
    Type           string                 // Event type (e.g., "user.created")
    Source         string                 // Source service name
    OrganizationID string                 // Organization/tenant ID
    Timestamp      time.Time              // Event timestamp (UTC)
    SchemaVersion  string                 // Schema version ("1.0")
    Data           map[string]interface{} // Event payload
}
```

**Note**: The `NewEvent` function now requires a `timestamp` parameter. Use `TimeManager.Now()` for testability:

```go
event := events.NewEvent(
    "user.created",
    "auth-service",
    "org-123",
    timeManager.Now(), // Use injected TimeManager
    map[string]interface{}{
        "user_id": 12345,
    },
)
```

### Event JSON Format

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "type": "user.created",
  "source": "auth-service",
  "organization_id": "org-123",
  "timestamp": "2024-01-15T10:30:45Z",
  "data": {
    "user_id": 12345,
    "email": "user@example.com",
    "role": "admin"
  }
}
```

### Testing with Mocks

```go
mockPublisher := new(events.PublisherMock)
mockPublisher.On("Publish", mock.Anything, "users.events", mock.Anything).Return(nil)

service := NewUserService(mockPublisher)
service.CreateUser(ctx, userRequest)

mockPublisher.AssertExpectations(t)
```

## Event Naming Conventions

Use dot-notation for event types:

- `{resource}.{action}` - e.g., `user.created`, `order.updated`, `payment.failed`
- `{domain}.{resource}.{action}` - e.g., `catalog.product.deleted`, `billing.invoice.sent`

### Examples

```go
// User domain
user.created
user.updated
user.deleted
user.password_reset

// Order domain
order.created
order.confirmed
order.shipped
order.delivered
order.cancelled

// Payment domain
payment.initiated
payment.succeeded
payment.failed
payment.refunded
```

## Stream Management

### Create Streams

```go
nc, _ := events.ConnectWithDefaults("nats://localhost:4222")
js, _ := nc.JetStream()

// Create all default streams (AUTH_EVENTS, CATALOG_EVENTS, etc.)
err := events.CreateDefaultStreams(js)

// Or create a custom stream
streamConfig := events.NewStreamConfig("MY_EVENTS", []string{"my.>"})
streamConfig.MaxAge = 14 * 24 * time.Hour  // 14 days retention
streamConfig.MaxBytes = 2 * 1024 * 1024 * 1024  // 2GB max size
err = events.CreateStream(js, streamConfig)
```

### Get Stream Info

```go
info, err := events.GetStreamInfo(js, "AUTH_EVENTS")
if err != nil {
    log.Fatal(err)
}

log.Printf("Stream: %s, Messages: %d, Bytes: %d",
    info.Config.Name, info.State.Msgs, info.State.Bytes)
```

### Update Stream Configuration

```go
streamConfig := events.NewStreamConfig("AUTH_EVENTS", []string{"auth.>"})
streamConfig.MaxAge = 30 * 24 * time.Hour  // Increase to 30 days
err = events.UpdateStream(js, streamConfig)
```

### Default Streams

The package provides pre-configured streams for all services:

- `AUTH_EVENTS` - Authentication and authorization events
- `CATALOG_EVENTS` - Product and catalog events
- `DDMRP_EVENTS` - DDMRP buffer calculation events
- `EXECUTION_EVENTS` - Order execution events
- `ANALYTICS_EVENTS` - Analytics and reporting events
- `AI_AGENT_EVENTS` - AI assistant events
- `DLQ_EVENTS` - Dead letter queue for failed events

All streams have:
- 7-day retention
- 1GB max size
- File storage (persistent)
- Old message discard policy

## Advanced Features

### Custom Subscriber Configuration

Configure max retries and acknowledgment timeout:

```go
config := &events.SubscriberConfig{
    MaxDeliver: 5,               // Retry up to 5 times
    AckWait:    30 * time.Second, // Wait 30s for acknowledgment
}

err = subscriber.SubscribeDurableWithConfig(
    ctx,
    "auth.user.*",
    "catalog-consumer",
    config,
    handler,
)
```

### Event Validation

Events are automatically validated before publishing:

```go
event := events.NewEvent("", "auth-service", "org-123", time.Now(), nil)
err := publisher.Publish(ctx, "auth.user.created", event)
// Returns: BadRequest error "event type is required"
```

### Typed Errors

All errors use typed errors from `pkg/errors`:

```go
err := publisher.Publish(ctx, "", event)
if err != nil {
    switch e := err.(type) {
    case *errors.BadRequest:
        log.Printf("Invalid input: %v", e)
    case *errors.InternalServerError:
        log.Printf("NATS error: %v", e)
    }
}
```

## Best Practices

1. **Use durable subscriptions** for critical event processing
2. **Idempotent handlers** - events may be delivered more than once
3. **Include organization_id** for multi-tenancy
4. **Version your events** if schema changes are expected
5. **Log all published and received events** for debugging
6. **Handle handler errors gracefully** (NAK will retry)
7. **Use dead letter queues** for failed events after max retries
8. **Monitor event lag** and processing time
9. **Use TimeManager** for timestamps (testability)
10. **Validate events** before publishing (automatic)

## Configuration

### Environment Variables

```bash
# NATS Configuration
export NATS_URL=nats://localhost:4222
export NATS_MAX_RECONNECTS=10
export NATS_RECONNECT_WAIT=2s

# Service Configuration
export SERVICE_NAME=auth-service
export ORGANIZATION_ID=org-123
```

## Delivery Guarantees

- **At-least-once delivery**: Events may be delivered multiple times
- **Ordered delivery**: Events on the same subject are delivered in order
- **Durable subscriptions**: Resume from last acknowledged message after restart
