//go:build integration

package events

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/melegattip/giia-core-engine/pkg/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubscriber_Integration(t *testing.T) {
	ctx := context.Background()
	cm := testutil.NewContainerManager()
	defer cm.Cleanup(ctx)

	natsURL, cleanup := cm.StartNATS(ctx, t)
	defer cleanup()

	nc, cleanupConn := testutil.SetupTestNATS(t, natsURL)
	defer cleanupConn()

	t.Run("should create subscriber successfully", func(t *testing.T) {
		subscriber, err := NewSubscriber(nc)
		assert.NoError(t, err)
		assert.NotNil(t, subscriber)
	})

	t.Run("should fail to create subscriber with nil connection", func(t *testing.T) {
		subscriber, err := NewSubscriber(nil)
		assert.Error(t, err)
		assert.Nil(t, subscriber)
	})
}

func TestSubscriber_Subscribe_Integration(t *testing.T) {
	ctx := context.Background()
	cm := testutil.NewContainerManager()
	defer cm.Cleanup(ctx)

	natsURL, cleanup := cm.StartNATS(ctx, t)
	defer cleanup()

	nc, cleanupConn := testutil.SetupTestNATS(t, natsURL)
	defer cleanupConn()

	js, err := nc.JetStream()
	require.NoError(t, err)

	testutil.CreateTestStream(t, js, "TEST_SUB_EVENTS", []string{"sub.>"})
	defer testutil.DeleteStream(t, js, "TEST_SUB_EVENTS")

	publisher, err := NewPublisher(nc)
	require.NoError(t, err)

	subscriber, err := NewSubscriber(nc)
	require.NoError(t, err)

	t.Run("should receive published event", func(t *testing.T) {
		testutil.PurgeStream(t, js, "TEST_SUB_EVENTS")

		var receivedEvent *Event
		var wg sync.WaitGroup
		wg.Add(1)

		err := subscriber.Subscribe(ctx, "sub.>", func(ctx context.Context, event *Event) error {
			receivedEvent = event
			wg.Done()
			return nil
		})
		require.NoError(t, err)

		time.Sleep(200 * time.Millisecond)

		event := NewEvent(
			"sub.created",
			"test-service",
			"org-123",
			time.Now(),
			map[string]interface{}{
				"item_id": "456",
			},
		)
		err = publisher.Publish(ctx, "sub.items.created", event)
		require.NoError(t, err)

		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			assert.NotNil(t, receivedEvent)
			assert.Equal(t, "sub.created", receivedEvent.Type)
			assert.Equal(t, "456", receivedEvent.Data["item_id"])
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for event")
		}
	})

	t.Run("should handle multiple events", func(t *testing.T) {
		testutil.PurgeStream(t, js, "TEST_SUB_EVENTS")

		receivedEvents := []*Event{}
		var mu sync.Mutex
		var wg sync.WaitGroup
		wg.Add(3)

		ncSub, cleanupConnSub := testutil.SetupTestNATS(t, natsURL)
		defer cleanupConnSub()

		subscriberMulti, err := NewSubscriber(ncSub)
		require.NoError(t, err)

		err = subscriberMulti.Subscribe(ctx, "sub.>", func(ctx context.Context, event *Event) error {
			mu.Lock()
			receivedEvents = append(receivedEvents, event)
			mu.Unlock()
			wg.Done()
			return nil
		})
		require.NoError(t, err)

		time.Sleep(200 * time.Millisecond)

		for i := 1; i <= 3; i++ {
			event := NewEvent(
				"sub.created",
				"test-service",
				"org-123",
				time.Now(),
				map[string]interface{}{
					"id": fmt.Sprintf("%d", i),
				},
			)
			err = publisher.Publish(ctx, fmt.Sprintf("sub.items.%d", i), event)
			require.NoError(t, err)
		}

		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			assert.Len(t, receivedEvents, 3)
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for events")
		}
	})

	t.Run("should fail with empty subject", func(t *testing.T) {
		err := subscriber.Subscribe(ctx, "", func(ctx context.Context, event *Event) error {
			return nil
		})
		assert.Error(t, err)
	})

	t.Run("should fail with nil handler", func(t *testing.T) {
		err := subscriber.Subscribe(ctx, "sub.test", nil)
		assert.Error(t, err)
	})
}

func TestSubscriber_SubscribeDurable_Integration(t *testing.T) {
	ctx := context.Background()
	cm := testutil.NewContainerManager()
	defer cm.Cleanup(ctx)

	natsURL, cleanup := cm.StartNATS(ctx, t)
	defer cleanup()

	nc, cleanupConn := testutil.SetupTestNATS(t, natsURL)
	defer cleanupConn()

	js, err := nc.JetStream()
	require.NoError(t, err)

	testutil.CreateTestStream(t, js, "TEST_DURABLE_EVENTS", []string{"durable.>"})
	defer testutil.DeleteStream(t, js, "TEST_DURABLE_EVENTS")

	publisher, err := NewPublisher(nc)
	require.NoError(t, err)

	t.Run("should create durable subscription", func(t *testing.T) {
		testutil.PurgeStream(t, js, "TEST_DURABLE_EVENTS")

		ncSub, cleanupConnSub := testutil.SetupTestNATS(t, natsURL)
		defer cleanupConnSub()

		subscriber, err := NewSubscriber(ncSub)
		require.NoError(t, err)

		var receivedCount int
		var mu sync.Mutex

		err = subscriber.SubscribeDurable(ctx, "durable.>", "test-durable", func(ctx context.Context, event *Event) error {
			mu.Lock()
			receivedCount++
			mu.Unlock()
			return nil
		})
		require.NoError(t, err)

		time.Sleep(200 * time.Millisecond)

		for i := 1; i <= 3; i++ {
			event := NewEvent(
				"durable.created",
				"test-service",
				"org-123",
				time.Now(),
				map[string]interface{}{"id": i},
			)
			err = publisher.Publish(ctx, "durable.test", event)
			require.NoError(t, err)
		}

		time.Sleep(1 * time.Second)

		mu.Lock()
		count := receivedCount
		mu.Unlock()

		assert.Equal(t, 3, count)
	})

	t.Run("should fail with empty durable name", func(t *testing.T) {
		subscriber, err := NewSubscriber(nc)
		require.NoError(t, err)

		err = subscriber.SubscribeDurable(ctx, "durable.test", "", func(ctx context.Context, event *Event) error {
			return nil
		})
		assert.Error(t, err)
	})
}

func TestSubscriber_ErrorHandling_Integration(t *testing.T) {
	ctx := context.Background()
	cm := testutil.NewContainerManager()
	defer cm.Cleanup(ctx)

	natsURL, cleanup := cm.StartNATS(ctx, t)
	defer cleanup()

	nc, cleanupConn := testutil.SetupTestNATS(t, natsURL)
	defer cleanupConn()

	js, err := nc.JetStream()
	require.NoError(t, err)

	testutil.CreateTestStream(t, js, "TEST_ERROR_EVENTS", []string{"error.>"})
	defer testutil.DeleteStream(t, js, "TEST_ERROR_EVENTS")

	publisher, err := NewPublisher(nc)
	require.NoError(t, err)

	t.Run("should handle processing errors with retries", func(t *testing.T) {
		testutil.PurgeStream(t, js, "TEST_ERROR_EVENTS")

		ncSub, cleanupConnSub := testutil.SetupTestNATS(t, natsURL)
		defer cleanupConnSub()

		subscriberErr, err := NewSubscriber(ncSub)
		require.NoError(t, err)

		attemptCount := 0
		var mu sync.Mutex

		config := &SubscriberConfig{
			MaxDeliver: 3,
			AckWait:    1 * time.Second,
		}

		err = subscriberErr.SubscribeDurableWithConfig(ctx, "error.>", "error-handler", config, func(ctx context.Context, event *Event) error {
			mu.Lock()
			attemptCount++
			mu.Unlock()
			return fmt.Errorf("simulated processing error")
		})
		require.NoError(t, err)

		time.Sleep(200 * time.Millisecond)

		event := NewEvent(
			"error.test",
			"test-service",
			"org-123",
			time.Now(),
			map[string]interface{}{"test": "error"},
		)
		err = publisher.Publish(ctx, "error.test", event)
		require.NoError(t, err)

		time.Sleep(5 * time.Second)

		mu.Lock()
		count := attemptCount
		mu.Unlock()

		assert.GreaterOrEqual(t, count, 3, "Should retry at least MaxDeliver times")
	})
}

func TestSubscriber_Close_Integration(t *testing.T) {
	ctx := context.Background()
	cm := testutil.NewContainerManager()
	defer cm.Cleanup(ctx)

	natsURL, cleanup := cm.StartNATS(ctx, t)
	defer cleanup()

	nc, cleanupConn := testutil.SetupTestNATS(t, natsURL)
	defer cleanupConn()

	js, err := nc.JetStream()
	require.NoError(t, err)

	testutil.CreateTestStream(t, js, "TEST_CLOSE_EVENTS", []string{"close.>"})
	defer testutil.DeleteStream(t, js, "TEST_CLOSE_EVENTS")

	subscriber, err := NewSubscriber(nc)
	require.NoError(t, err)

	err = subscriber.Subscribe(ctx, "close.>", func(ctx context.Context, event *Event) error {
		return nil
	})
	require.NoError(t, err)

	t.Run("should close subscriber successfully", func(t *testing.T) {
		err := subscriber.Close()
		assert.NoError(t, err)
	})
}
