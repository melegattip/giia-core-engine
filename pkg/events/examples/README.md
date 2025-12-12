# NATS JetStream Events - Usage Examples

This directory contains practical examples demonstrating how to use the `pkg/events` package in real-world scenarios.

## Examples Overview

### 1. Publisher Example ([publisher_example.go](publisher_example.go))

Demonstrates how to publish events from a service using the NATS JetStream publisher.

**Key Features**:
- Creating events with proper structure
- Publishing synchronous events with `Publish()`
- Publishing asynchronous events with `PublishAsync()`
- TimeManager injection for testability
- Proper error handling and logging
- Service integration pattern

**Events Published**:
- `user.created` - When a new user is created
- `user.role.updated` - When a user's role changes
- `user.deleted` - When a user is deleted

**Run the example**:
```bash
# Start NATS JetStream first
docker-compose up -d nats

# Create streams
./scripts/setup-nats-streams.sh

# Run publisher
go run pkg/events/examples/publisher_example.go
```

### 2. Subscriber Example ([subscriber_example.go](subscriber_example.go))

Demonstrates how to consume events using durable subscriptions with proper error handling.

**Key Features**:
- Durable consumer configuration
- Event routing pattern
- Custom subscriber configuration (max retries, ack timeout)
- Graceful shutdown handling
- Event handler pattern
- Type-safe event processing

**Events Consumed**:
- `auth.user.*` - All user-related events from auth service

**Run the example**:
```bash
# Start NATS JetStream first
docker-compose up -d nats

# Create streams
./scripts/setup-nats-streams.sh

# Run subscriber
go run pkg/events/examples/subscriber_example.go
```

## Testing the Examples Together

### Terminal 1 - Start Subscriber
```bash
go run pkg/events/examples/subscriber_example.go
```

### Terminal 2 - Publish Events
```bash
go run pkg/events/examples/publisher_example.go
```

You should see the subscriber receive and process the events published by the publisher.

## Event Structure

All events follow the CloudEvents-inspired structure:

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "type": "user.created",
  "source": "auth-service",
  "organization_id": "org-123",
  "timestamp": "2024-01-15T10:30:45Z",
  "schema_version": "1.0",
  "data": {
    "user_id": 12345,
    "email": "user@example.com",
    "role": "admin"
  }
}
```

## Best Practices Demonstrated

### Publisher Best Practices
1. **Inject TimeManager** for timestamp generation (testability)
2. **Use typed errors** from `pkg/errors`
3. **Log all published events** for debugging
4. **Handle publish errors** appropriately
5. **Use PublishAsync** for fire-and-forget scenarios
6. **Close publisher** on shutdown

### Subscriber Best Practices
1. **Use durable subscriptions** for critical processing
2. **Configure appropriate retry limits** (MaxDeliver)
3. **Set reasonable ack timeouts** (AckWait)
4. **Route events by type** for clean handler separation
5. **Make handlers idempotent** (events may be delivered multiple times)
6. **Handle graceful shutdown** with signal handling
7. **Log all received events** for debugging

## Integration Patterns

### Service Integration

```go
// In your service constructor
func NewMyService(natsConn *nats.Conn, timeManager TimeManager) (*MyService, error) {
    publisher, err := events.NewPublisher(natsConn)
    if err != nil {
        return nil, err
    }

    return &MyService{
        publisher:   publisher,
        timeManager: timeManager,
    }, nil
}

// In your business logic
func (s *MyService) CreateResource(ctx context.Context, data *ResourceData) error {
    // Business logic...
    resource := createResource(data)

    // Publish event
    event := events.NewEvent(
        "resource.created",
        "my-service",
        data.OrganizationID,
        s.timeManager.Now(),
        map[string]interface{}{
            "resource_id": resource.ID,
            "type":        resource.Type,
        },
    )

    return s.publisher.Publish(ctx, "my-service.resource.created", event)
}
```

### Testing with Mocks

```go
func TestService_CreateResource(t *testing.T) {
    // Given
    mockPublisher := new(events.PublisherMock)
    mockTimeManager := new(TimeManagerMock)

    testTime := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
    mockTimeManager.On("Now").Return(testTime)

    mockPublisher.On("Publish",
        mock.Anything,
        "my-service.resource.created",
        mock.MatchedBy(func(e *events.Event) bool {
            return e.Type == "resource.created" &&
                   e.Source == "my-service" &&
                   e.Timestamp.Equal(testTime)
        }),
    ).Return(nil)

    service := NewMyService(mockPublisher, mockTimeManager)

    // When
    err := service.CreateResource(context.Background(), testData)

    // Then
    assert.NoError(t, err)
    mockPublisher.AssertExpectations(t)
}
```

## Troubleshooting

### Connection Issues

```bash
# Verify NATS is running
docker ps | grep nats

# Check NATS logs
docker logs giia-nats

# Test connection
nats server check --server=nats://localhost:4222
```

### Stream Issues

```bash
# List streams
nats stream list --server=nats://localhost:4222

# View stream info
nats stream info AUTH_EVENTS --server=nats://localhost:4222

# View consumers
nats consumer list AUTH_EVENTS --server=nats://localhost:4222
```

### Event Debugging

```bash
# Monitor events in real-time
nats sub "auth.>" --server=nats://localhost:4222

# View specific subject
nats sub "auth.user.created" --server=nats://localhost:4222
```

## Next Steps

1. Review the [main README](../README.md) for complete API documentation
2. Read the [NATS Architecture Guide](../../../docs/NATS_ARCHITECTURE.md) (if available)
3. Check the [Event Catalog](../../../docs/EVENTS.md) for all event types (if available)
4. Integrate event publishing into your service
5. Write unit tests using mocks
6. Test in staging environment with real NATS
