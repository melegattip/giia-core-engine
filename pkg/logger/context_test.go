package logger

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithRequestID(t *testing.T) {
	t.Run("should add request ID to context", func(t *testing.T) {
		ctx := context.Background()
		requestID := "req-12345"

		newCtx := WithRequestID(ctx, requestID)

		value := newCtx.Value(requestIDKey)
		assert.NotNil(t, value)
		assert.Equal(t, requestID, value)
	})

	t.Run("should handle empty request ID", func(t *testing.T) {
		ctx := context.Background()

		newCtx := WithRequestID(ctx, "")

		value := newCtx.Value(requestIDKey)
		assert.Equal(t, "", value)
	})

	t.Run("should overwrite existing request ID", func(t *testing.T) {
		ctx := context.Background()
		ctx = WithRequestID(ctx, "req-old")

		newCtx := WithRequestID(ctx, "req-new")

		value := newCtx.Value(requestIDKey)
		assert.Equal(t, "req-new", value)
	})
}

func TestExtractRequestID(t *testing.T) {
	t.Run("should extract request ID from context", func(t *testing.T) {
		ctx := context.Background()
		expectedRequestID := "req-67890"
		ctx = WithRequestID(ctx, expectedRequestID)

		requestID := ExtractRequestID(ctx)

		assert.Equal(t, expectedRequestID, requestID)
	})

	t.Run("should return empty string when no request ID in context", func(t *testing.T) {
		ctx := context.Background()

		requestID := ExtractRequestID(ctx)

		assert.Equal(t, "", requestID)
	})

	t.Run("should return empty string for invalid type in context", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), requestIDKey, 12345)

		requestID := ExtractRequestID(ctx)

		assert.Equal(t, "", requestID)
	})

	t.Run("should handle nil context gracefully", func(t *testing.T) {
		requestID := ExtractRequestID(context.Background())

		assert.Equal(t, "", requestID)
	})
}

func TestRequestIDWorkflow(t *testing.T) {
	t.Run("should maintain request ID through context chain", func(t *testing.T) {
		ctx := context.Background()
		requestID := "req-workflow-123"

		ctx = WithRequestID(ctx, requestID)

		childCtx := context.WithValue(ctx, "other_key", "other_value")

		extractedID := ExtractRequestID(childCtx)
		assert.Equal(t, requestID, extractedID)
	})

	t.Run("should work with context cancellation", func(t *testing.T) {
		ctx := context.Background()
		requestID := "req-cancel-456"

		ctx = WithRequestID(ctx, requestID)
		cancelCtx, cancel := context.WithCancel(ctx)
		defer cancel()

		extractedID := ExtractRequestID(cancelCtx)
		assert.Equal(t, requestID, extractedID)
	})
}
