//go:build integration

package events

import (
	"context"
	"testing"
	"time"

	"github.com/melegattip/giia-core-engine/pkg/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPublisher_Integration(t *testing.T) {
	ctx := context.Background()
	cm := testutil.NewContainerManager()
	defer cm.Cleanup(ctx)

	natsURL, cleanup := cm.StartNATS(ctx, t)
	defer cleanup()

	nc, cleanupConn := testutil.SetupTestNATS(t, natsURL)
	defer cleanupConn()

	js, err := nc.JetStream()
	require.NoError(t, err)

	testutil.CreateTestStream(t, js, "TEST_EVENTS", []string{"test.>"})
	defer testutil.DeleteStream(t, js, "TEST_EVENTS")

	t.Run("should create publisher successfully", func(t *testing.T) {
		publisher, err := NewPublisher(nc)
		assert.NoError(t, err)
		assert.NotNil(t, publisher)
	})

	t.Run("should fail to create publisher with nil connection", func(t *testing.T) {
		publisher, err := NewPublisher(nil)
		assert.Error(t, err)
		assert.Nil(t, publisher)
	})
}

func TestPublisher_Publish_Integration(t *testing.T) {
	ctx := context.Background()
	cm := testutil.NewContainerManager()
	defer cm.Cleanup(ctx)

	natsURL, cleanup := cm.StartNATS(ctx, t)
	defer cleanup()

	nc, cleanupConn := testutil.SetupTestNATS(t, natsURL)
	defer cleanupConn()

	js, err := nc.JetStream()
	require.NoError(t, err)

	testutil.CreateTestStream(t, js, "TEST_EVENTS", []string{"test.>"})
	defer testutil.DeleteStream(t, js, "TEST_EVENTS")

	publisher, err := NewPublisher(nc)
	require.NoError(t, err)

	t.Run("should publish event successfully", func(t *testing.T) {
		testutil.PurgeStream(t, js, "TEST_EVENTS")

		event := NewEvent(
			"test.created",
			"test-service",
			"org-123",
			time.Now(),
			map[string]interface{}{
				"item_id": "123",
				"name":    "Test Item",
			},
		)

		err := publisher.Publish(ctx, "test.items.created", event)
		assert.NoError(t, err)

		time.Sleep(200 * time.Millisecond)

		stream := testutil.GetStreamInfo(t, js, "TEST_EVENTS")
		assert.Equal(t, uint64(1), stream.State.Msgs)
	})

	t.Run("should publish multiple events", func(t *testing.T) {
		testutil.PurgeStream(t, js, "TEST_EVENTS")

		events := []*Event{
			NewEvent("test.created", "test-service", "org-123", time.Now(), map[string]interface{}{"id": "1"}),
			NewEvent("test.created", "test-service", "org-123", time.Now(), map[string]interface{}{"id": "2"}),
			NewEvent("test.created", "test-service", "org-123", time.Now(), map[string]interface{}{"id": "3"}),
		}

		for _, event := range events {
			err := publisher.Publish(ctx, "test.items.created", event)
			assert.NoError(t, err)
		}

		time.Sleep(200 * time.Millisecond)

		stream := testutil.GetStreamInfo(t, js, "TEST_EVENTS")
		assert.Equal(t, uint64(3), stream.State.Msgs)
	})

	t.Run("should fail with empty subject", func(t *testing.T) {
		event := NewEvent("test.created", "test-service", "org-123", time.Now(), map[string]interface{}{})

		err := publisher.Publish(ctx, "", event)
		assert.Error(t, err)
	})

	t.Run("should fail with nil event", func(t *testing.T) {
		err := publisher.Publish(ctx, "test.items.created", nil)
		assert.Error(t, err)
	})

	t.Run("should fail with invalid event", func(t *testing.T) {
		event := &Event{
			ID:   "123",
			Type: "",
		}

		err := publisher.Publish(ctx, "test.items.created", event)
		assert.Error(t, err)
	})
}

func TestPublisher_PublishAsync_Integration(t *testing.T) {
	ctx := context.Background()
	cm := testutil.NewContainerManager()
	defer cm.Cleanup(ctx)

	natsURL, cleanup := cm.StartNATS(ctx, t)
	defer cleanup()

	nc, cleanupConn := testutil.SetupTestNATS(t, natsURL)
	defer cleanupConn()

	js, err := nc.JetStream()
	require.NoError(t, err)

	testutil.CreateTestStream(t, js, "TEST_ASYNC_EVENTS", []string{"async.>"})
	defer testutil.DeleteStream(t, js, "TEST_ASYNC_EVENTS")

	publisher, err := NewPublisher(nc)
	require.NoError(t, err)

	t.Run("should publish event asynchronously", func(t *testing.T) {
		testutil.PurgeStream(t, js, "TEST_ASYNC_EVENTS")

		event := NewEvent(
			"async.created",
			"test-service",
			"org-123",
			time.Now(),
			map[string]interface{}{
				"test": "async",
			},
		)

		err := publisher.PublishAsync(ctx, "async.test", event)
		assert.NoError(t, err)

		success := testutil.WaitForMessages(t, js, "TEST_ASYNC_EVENTS", 1, 3*time.Second)
		assert.True(t, success, "Message should be published asynchronously")
	})

	t.Run("should publish multiple async events", func(t *testing.T) {
		testutil.PurgeStream(t, js, "TEST_ASYNC_EVENTS")

		for i := 1; i <= 5; i++ {
			event := NewEvent(
				"async.created",
				"test-service",
				"org-123",
				time.Now(),
				map[string]interface{}{"index": i},
			)

			err := publisher.PublishAsync(ctx, "async.test", event)
			assert.NoError(t, err)
		}

		success := testutil.WaitForMessages(t, js, "TEST_ASYNC_EVENTS", 5, 3*time.Second)
		assert.True(t, success, "All async messages should be published")
	})
}

func TestPublisher_Close_Integration(t *testing.T) {
	ctx := context.Background()
	cm := testutil.NewContainerManager()
	defer cm.Cleanup(ctx)

	natsURL, cleanup := cm.StartNATS(ctx, t)
	defer cleanup()

	nc, cleanupConn := testutil.SetupTestNATS(t, natsURL)
	defer cleanupConn()

	publisher, err := NewPublisher(nc)
	require.NoError(t, err)

	t.Run("should close publisher successfully", func(t *testing.T) {
		err := publisher.Close()
		assert.NoError(t, err)
	})
}
