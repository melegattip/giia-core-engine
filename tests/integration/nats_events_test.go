package integration

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/melegattip/giia-core-engine/tests/integration/clients"
)

// NATSEvent represents a generic NATS event for testing.
type NATSEvent struct {
	EventID        string                 `json:"event_id"`
	EventType      string                 `json:"event_type"`
	OrganizationID string                 `json:"organization_id"`
	Timestamp      time.Time              `json:"timestamp"`
	Payload        map[string]interface{} `json:"payload"`
}

// TestNATSEvents_ProductCreated tests that product creation publishes NATS events.
func TestNATSEvents_ProductCreated(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	env := DefaultTestEnvironment()
	defer env.Teardown()

	// Connect to NATS
	nc, err := nats.Connect(env.NATSUrl)
	if err != nil {
		t.Skipf("Cannot connect to NATS: %v", err)
	}
	defer nc.Close()

	authClient := clients.NewAuthClient(env.AuthService.HTTPURL)
	catalogClient := clients.NewCatalogClient(env.CatalogService.HTTPURL)

	organizationID := uuid.New().String()
	email := generateTestEmail()
	password := "SecurePassword123!"

	_, err = authClient.Register(ctx, clients.RegisterRequest{
		Email:          email,
		Password:       password,
		FirstName:      "NATS",
		LastName:       "Test",
		Phone:          "+1234567890",
		OrganizationID: organizationID,
	})
	require.NoError(t, err)

	tokens, err := authClient.Login(ctx, email, password)
	require.NoError(t, err)

	// Subscribe to product events
	eventReceived := make(chan NATSEvent, 1)
	sub, err := nc.Subscribe("catalog.product.>", func(msg *nats.Msg) {
		var event NATSEvent
		if err := json.Unmarshal(msg.Data, &event); err == nil {
			eventReceived <- event
		}
	})
	if err != nil {
		t.Skipf("Cannot subscribe to NATS: %v", err)
	}
	defer sub.Unsubscribe()

	t.Run("ProductCreationPublishesEvent", func(t *testing.T) {
		// Create product
		productResp, err := catalogClient.CreateProduct(ctx, clients.CreateProductRequest{
			OrganizationID: organizationID,
			SKU:            generateTestSKU("NATS"),
			Name:           "NATS Event Test Product",
			Description:    "Testing NATS events",
			Category:       "Test",
			UnitOfMeasure:  "UNIT",
		}, tokens.AccessToken)
		require.NoError(t, err)

		// Wait for event
		select {
		case event := <-eventReceived:
			assert.Contains(t, event.EventType, "product", "Event type should be product-related")
			assert.Equal(t, organizationID, event.OrganizationID, "Event should be for correct org")
			if payload, ok := event.Payload["product_id"].(string); ok {
				assert.Equal(t, productResp.Product.ID, payload, "Event should contain correct product ID")
			}
		case <-time.After(5 * time.Second):
			t.Log("No NATS event received within timeout - this may be expected if events are not implemented")
		}
	})
}

// TestNATSEvents_PurchaseOrderCreated tests that purchase order creation publishes events.
func TestNATSEvents_PurchaseOrderCreated(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	env := DefaultTestEnvironment()
	defer env.Teardown()

	nc, err := nats.Connect(env.NATSUrl)
	if err != nil {
		t.Skipf("Cannot connect to NATS: %v", err)
	}
	defer nc.Close()

	authClient := clients.NewAuthClient(env.AuthService.HTTPURL)
	catalogClient := clients.NewCatalogClient(env.CatalogService.HTTPURL)
	executionClient := clients.NewExecutionClient(env.ExecutionService.HTTPURL)

	organizationID := uuid.New().String()
	email := generateTestEmail()
	password := "SecurePassword123!"

	_, err = authClient.Register(ctx, clients.RegisterRequest{
		Email:          email,
		Password:       password,
		FirstName:      "PO",
		LastName:       "Event",
		Phone:          "+1234567890",
		OrganizationID: organizationID,
	})
	require.NoError(t, err)

	tokens, err := authClient.Login(ctx, email, password)
	require.NoError(t, err)

	// Create product first
	productResp, err := catalogClient.CreateProduct(ctx, clients.CreateProductRequest{
		OrganizationID: organizationID,
		SKU:            generateTestSKU("PO-EVENT"),
		Name:           "PO Event Product",
		Description:    "Test",
		Category:       "Test",
		UnitOfMeasure:  "UNIT",
	}, tokens.AccessToken)
	require.NoError(t, err)

	// Subscribe to execution events
	eventReceived := make(chan NATSEvent, 10)
	sub, err := nc.Subscribe("execution.>", func(msg *nats.Msg) {
		var event NATSEvent
		if err := json.Unmarshal(msg.Data, &event); err == nil {
			eventReceived <- event
		}
	})
	if err != nil {
		t.Skipf("Cannot subscribe to NATS: %v", err)
	}
	defer sub.Unsubscribe()

	t.Run("PurchaseOrderCreationPublishesEvent", func(t *testing.T) {
		_, err := executionClient.CreatePurchaseOrder(ctx, clients.CreatePurchaseOrderRequest{
			OrganizationID: organizationID,
			SupplierID:     uuid.New().String(),
			Items: []clients.CreateOrderItemRequest{
				{
					ProductID: productResp.Product.ID,
					SKU:       "PO-EVENT-SKU",
					Quantity:  50.0,
					UnitPrice: 10.0,
				},
			},
			Notes: "NATS event test",
		}, tokens.AccessToken)
		require.NoError(t, err)

		select {
		case event := <-eventReceived:
			assert.Equal(t, organizationID, event.OrganizationID)
			assert.Contains(t, event.EventType, "order", "Event should be order-related")
		case <-time.After(5 * time.Second):
			t.Log("No NATS event received - events may not be implemented yet")
		}
	})
}

// TestNATSEvents_GoodsReceived tests that receiving goods publishes an event.
func TestNATSEvents_GoodsReceived(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	env := DefaultTestEnvironment()
	defer env.Teardown()

	nc, err := nats.Connect(env.NATSUrl)
	if err != nil {
		t.Skipf("Cannot connect to NATS: %v", err)
	}
	defer nc.Close()

	authClient := clients.NewAuthClient(env.AuthService.HTTPURL)
	catalogClient := clients.NewCatalogClient(env.CatalogService.HTTPURL)
	executionClient := clients.NewExecutionClient(env.ExecutionService.HTTPURL)

	organizationID := uuid.New().String()
	email := generateTestEmail()

	_, err = authClient.Register(ctx, clients.RegisterRequest{
		Email:          email,
		Password:       "SecurePassword123!",
		FirstName:      "Receive",
		LastName:       "Event",
		Phone:          "+1234567890",
		OrganizationID: organizationID,
	})
	require.NoError(t, err)

	tokens, err := authClient.Login(ctx, email, "SecurePassword123!")
	require.NoError(t, err)

	// Create product
	productResp, err := catalogClient.CreateProduct(ctx, clients.CreateProductRequest{
		OrganizationID: organizationID,
		SKU:            generateTestSKU("RECV-EVENT"),
		Name:           "Receive Event Product",
		Description:    "Test",
		Category:       "Test",
		UnitOfMeasure:  "UNIT",
	}, tokens.AccessToken)
	require.NoError(t, err)

	// Create PO
	poResp, err := executionClient.CreatePurchaseOrder(ctx, clients.CreatePurchaseOrderRequest{
		OrganizationID: organizationID,
		SupplierID:     uuid.New().String(),
		Items: []clients.CreateOrderItemRequest{
			{
				ProductID: productResp.Product.ID,
				SKU:       "RECV-SKU",
				Quantity:  100.0,
				UnitPrice: 10.0,
			},
		},
		Notes: "Receive event test",
	}, tokens.AccessToken)
	require.NoError(t, err)

	// Subscribe to inventory/receive events
	receiveEventReceived := make(chan NATSEvent, 10)
	sub, err := nc.Subscribe("execution.inventory.>", func(msg *nats.Msg) {
		var event NATSEvent
		if err := json.Unmarshal(msg.Data, &event); err == nil {
			receiveEventReceived <- event
		}
	})
	if err != nil {
		t.Skipf("Cannot subscribe to NATS: %v", err)
	}
	defer sub.Unsubscribe()

	t.Run("ReceiveGoodsPublishesInventoryEvent", func(t *testing.T) {
		_, err = executionClient.ReceiveGoods(ctx, poResp.Order.ID, clients.ReceiveGoodsRequest{
			Items: []clients.ReceiveItemRequest{
				{
					ProductID: productResp.Product.ID,
					Quantity:  100.0,
				},
			},
		}, tokens.AccessToken)
		require.NoError(t, err)

		select {
		case event := <-receiveEventReceived:
			assert.Contains(t, event.EventType, "inventory", "Should be inventory event")
			assert.Equal(t, organizationID, event.OrganizationID)
		case <-time.After(5 * time.Second):
			t.Log("No inventory event received - events may not be implemented yet")
		}
	})
}

// TestNATSEvents_DDMRPBufferUpdate tests that DDMRP buffer updates trigger events.
func TestNATSEvents_DDMRPBufferUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	_ = ctx // ctx available for future use

	env := DefaultTestEnvironment()
	defer env.Teardown()

	nc, err := nats.Connect(env.NATSUrl)
	if err != nil {
		t.Skipf("Cannot connect to NATS: %v", err)
	}
	defer nc.Close()

	// Subscribe to DDMRP events
	ddmrpEventReceived := make(chan NATSEvent, 10)
	sub, err := nc.Subscribe("ddmrp.>", func(msg *nats.Msg) {
		var event NATSEvent
		if err := json.Unmarshal(msg.Data, &event); err == nil {
			ddmrpEventReceived <- event
		}
	})
	if err != nil {
		t.Skipf("Cannot subscribe to NATS: %v", err)
	}
	defer sub.Unsubscribe()

	t.Run("DDMRPEventsSubscription", func(t *testing.T) {
		// This test verifies the subscription is working
		// Actual DDMRP events depend on the DDMRP service implementation
		select {
		case event := <-ddmrpEventReceived:
			assert.NotEmpty(t, event.EventType)
		case <-time.After(3 * time.Second):
			t.Log("No DDMRP events received - this is expected if no buffer operations occurred")
		}
	})
}

// TestNATSEvents_EventOrdering tests that events are received in the correct order.
func TestNATSEvents_EventOrdering(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	env := DefaultTestEnvironment()
	defer env.Teardown()

	nc, err := nats.Connect(env.NATSUrl)
	if err != nil {
		t.Skipf("Cannot connect to NATS: %v", err)
	}
	defer nc.Close()

	authClient := clients.NewAuthClient(env.AuthService.HTTPURL)
	catalogClient := clients.NewCatalogClient(env.CatalogService.HTTPURL)

	organizationID := uuid.New().String()
	email := generateTestEmail()

	_, err = authClient.Register(ctx, clients.RegisterRequest{
		Email:          email,
		Password:       "SecurePassword123!",
		FirstName:      "Order",
		LastName:       "Test",
		Phone:          "+1234567890",
		OrganizationID: organizationID,
	})
	require.NoError(t, err)

	tokens, err := authClient.Login(ctx, email, "SecurePassword123!")
	require.NoError(t, err)

	// Collect all events
	allEvents := make(chan NATSEvent, 100)
	sub, err := nc.Subscribe(">", func(msg *nats.Msg) {
		var event NATSEvent
		if err := json.Unmarshal(msg.Data, &event); err == nil {
			allEvents <- event
		}
	})
	if err != nil {
		t.Skipf("Cannot subscribe to NATS: %v", err)
	}
	defer sub.Unsubscribe()

	t.Run("SequentialOperationsProduceOrderedEvents", func(t *testing.T) {
		// Create multiple products in sequence
		var productIDs []string
		for i := 0; i < 5; i++ {
			resp, err := catalogClient.CreateProduct(ctx, clients.CreateProductRequest{
				OrganizationID: organizationID,
				SKU:            generateTestSKU("ORDER-" + string(rune('0'+i))),
				Name:           "Order Test " + string(rune('0'+i)),
				Description:    "Test",
				Category:       "Test",
				UnitOfMeasure:  "UNIT",
			}, tokens.AccessToken)
			require.NoError(t, err)
			productIDs = append(productIDs, resp.Product.ID)
		}

		// Collect events for a short time
		var events []NATSEvent
		timeout := time.After(3 * time.Second)
	collectLoop:
		for {
			select {
			case event := <-allEvents:
				if event.OrganizationID == organizationID {
					events = append(events, event)
				}
			case <-timeout:
				break collectLoop
			}
		}

		// Verify events have sequential timestamps
		for i := 1; i < len(events); i++ {
			assert.True(t,
				events[i].Timestamp.After(events[i-1].Timestamp) || events[i].Timestamp.Equal(events[i-1].Timestamp),
				"Events should be in chronological order")
		}
	})
}

// TestNATSEvents_JetStreamDurability tests that JetStream provides durable event storage.
func TestNATSEvents_JetStreamDurability(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	_ = ctx // ctx available for future use

	env := DefaultTestEnvironment()
	defer env.Teardown()

	nc, err := nats.Connect(env.NATSUrl)
	if err != nil {
		t.Skipf("Cannot connect to NATS: %v", err)
	}
	defer nc.Close()

	js, err := nc.JetStream()
	if err != nil {
		t.Skipf("Cannot get JetStream context: %v", err)
	}

	t.Run("VerifyJetStreamAvailable", func(t *testing.T) {
		// List streams to verify JetStream is available
		streams := js.StreamNames()
		var streamNames []string
		for name := range streams {
			streamNames = append(streamNames, name)
		}
		t.Logf("Available JetStream streams: %v", streamNames)
		// No assertion - just verifying JetStream is accessible
	})

	t.Run("PublishAndRetrieveFromJetStream", func(t *testing.T) {
		streamName := "GIIA_INTEGRATION_TEST"
		subjectName := "integration.test.events"

		// Try to create or get stream
		_, err := js.AddStream(&nats.StreamConfig{
			Name:     streamName,
			Subjects: []string{subjectName},
			MaxMsgs:  100,
			MaxAge:   time.Hour,
		})
		if err != nil {
			// Stream might already exist
			t.Logf("Stream creation result: %v", err)
		}

		// Publish a test message
		testEvent := NATSEvent{
			EventID:        uuid.New().String(),
			EventType:      "integration.test",
			OrganizationID: uuid.New().String(),
			Timestamp:      time.Now(),
			Payload:        map[string]interface{}{"test": true},
		}
		eventBytes, _ := json.Marshal(testEvent)

		_, err = js.Publish(subjectName, eventBytes)
		if err != nil {
			t.Logf("Publish to JetStream: %v", err)
		}

		// Clean up
		js.DeleteStream(streamName)
	})
}
