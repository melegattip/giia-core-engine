package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Run("should create logger with default log level", func(t *testing.T) {
		logger := New("test-service", "")
		assert.NotNil(t, logger)
		assert.Equal(t, "test-service", logger.serviceName)
	})

	t.Run("should create logger with debug level", func(t *testing.T) {
		logger := New("test-service", "debug")
		assert.NotNil(t, logger)
		assert.Equal(t, "test-service", logger.serviceName)
	})

	t.Run("should create logger with info level", func(t *testing.T) {
		logger := New("test-service", "info")
		assert.NotNil(t, logger)
		assert.Equal(t, "test-service", logger.serviceName)
	})

	t.Run("should create logger with warn level", func(t *testing.T) {
		logger := New("test-service", "warn")
		assert.NotNil(t, logger)
		assert.Equal(t, "test-service", logger.serviceName)
	})

	t.Run("should create logger with error level", func(t *testing.T) {
		logger := New("test-service", "error")
		assert.NotNil(t, logger)
		assert.Equal(t, "test-service", logger.serviceName)
	})
}

func TestParseLogLevel(t *testing.T) {
	testCases := []struct {
		name          string
		givenLevel    string
		expectedLevel zerolog.Level
	}{
		{"debug level", "debug", zerolog.DebugLevel},
		{"info level", "info", zerolog.InfoLevel},
		{"warn level", "warn", zerolog.WarnLevel},
		{"error level", "error", zerolog.ErrorLevel},
		{"fatal level", "fatal", zerolog.FatalLevel},
		{"unknown level defaults to info", "unknown", zerolog.InfoLevel},
		{"empty level defaults to info", "", zerolog.InfoLevel},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			level := parseLogLevel(tc.givenLevel)
			assert.Equal(t, tc.expectedLevel, level)
		})
	}
}

func TestLogger_Debug(t *testing.T) {
	var buf bytes.Buffer
	logger := &ZerologLogger{
		logger: zerolog.New(&buf).With().Timestamp().Str("service", "test-service").Logger(),
		serviceName: "test-service",
	}
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	ctx := context.Background()

	t.Run("should log debug message", func(t *testing.T) {
		buf.Reset()
		logger.Debug(ctx, "debug message", nil)

		var logEntry map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &logEntry)
		require.NoError(t, err)

		assert.Equal(t, "debug", logEntry["level"])
		assert.Equal(t, "debug message", logEntry["message"])
		assert.Equal(t, "test-service", logEntry["service"])
	})

	t.Run("should log debug message with tags", func(t *testing.T) {
		buf.Reset()
		logger.Debug(ctx, "debug with tags", Tags{
			"user_id": "123",
			"action":  "test",
		})

		var logEntry map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &logEntry)
		require.NoError(t, err)

		assert.Equal(t, "debug", logEntry["level"])
		assert.Equal(t, "debug with tags", logEntry["message"])
		assert.Equal(t, "123", logEntry["user_id"])
		assert.Equal(t, "test", logEntry["action"])
	})
}

func TestLogger_Info(t *testing.T) {
	var buf bytes.Buffer
	logger := &ZerologLogger{
		logger: zerolog.New(&buf).With().Timestamp().Str("service", "test-service").Logger(),
		serviceName: "test-service",
	}
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	ctx := context.Background()

	t.Run("should log info message", func(t *testing.T) {
		buf.Reset()
		logger.Info(ctx, "info message", nil)

		var logEntry map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &logEntry)
		require.NoError(t, err)

		assert.Equal(t, "info", logEntry["level"])
		assert.Equal(t, "info message", logEntry["message"])
		assert.Equal(t, "test-service", logEntry["service"])
	})

	t.Run("should log info message with tags", func(t *testing.T) {
		buf.Reset()
		logger.Info(ctx, "request processed", Tags{
			"user_id":    "456",
			"duration_ms": 150,
		})

		var logEntry map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &logEntry)
		require.NoError(t, err)

		assert.Equal(t, "info", logEntry["level"])
		assert.Equal(t, "request processed", logEntry["message"])
		assert.Equal(t, "456", logEntry["user_id"])
		assert.Equal(t, float64(150), logEntry["duration_ms"])
	})
}

func TestLogger_Warn(t *testing.T) {
	var buf bytes.Buffer
	logger := &ZerologLogger{
		logger: zerolog.New(&buf).With().Timestamp().Str("service", "test-service").Logger(),
		serviceName: "test-service",
	}
	zerolog.SetGlobalLevel(zerolog.WarnLevel)

	ctx := context.Background()

	t.Run("should log warn message", func(t *testing.T) {
		buf.Reset()
		logger.Warn(ctx, "warning message", nil)

		var logEntry map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &logEntry)
		require.NoError(t, err)

		assert.Equal(t, "warn", logEntry["level"])
		assert.Equal(t, "warning message", logEntry["message"])
	})

	t.Run("should log warn message with tags", func(t *testing.T) {
		buf.Reset()
		logger.Warn(ctx, "rate limit approaching", Tags{
			"current_requests": 950,
			"max_requests":     1000,
		})

		var logEntry map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &logEntry)
		require.NoError(t, err)

		assert.Equal(t, "warn", logEntry["level"])
		assert.Equal(t, "rate limit approaching", logEntry["message"])
		assert.Equal(t, float64(950), logEntry["current_requests"])
	})
}

func TestLogger_Error(t *testing.T) {
	var buf bytes.Buffer
	logger := &ZerologLogger{
		logger: zerolog.New(&buf).With().Timestamp().Str("service", "test-service").Logger(),
		serviceName: "test-service",
	}
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)

	ctx := context.Background()

	t.Run("should log error message", func(t *testing.T) {
		buf.Reset()
		testErr := errors.New("test error")
		logger.Error(ctx, testErr, "error occurred", nil)

		var logEntry map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &logEntry)
		require.NoError(t, err)

		assert.Equal(t, "error", logEntry["level"])
		assert.Equal(t, "error occurred", logEntry["message"])
		assert.Equal(t, "test error", logEntry["error"])
	})

	t.Run("should log error message with tags", func(t *testing.T) {
		buf.Reset()
		testErr := errors.New("database connection failed")
		logger.Error(ctx, testErr, "database error", Tags{
			"operation": "insert",
			"table":     "users",
		})

		var logEntry map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &logEntry)
		require.NoError(t, err)

		assert.Equal(t, "error", logEntry["level"])
		assert.Equal(t, "database error", logEntry["message"])
		assert.Equal(t, "database connection failed", logEntry["error"])
		assert.Equal(t, "insert", logEntry["operation"])
		assert.Equal(t, "users", logEntry["table"])
	})
}

func TestLogger_ContextExtraction(t *testing.T) {
	var buf bytes.Buffer
	logger := &ZerologLogger{
		logger: zerolog.New(&buf).With().Timestamp().Str("service", "test-service").Logger(),
		serviceName: "test-service",
	}
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	t.Run("should extract request ID from context", func(t *testing.T) {
		ctx := WithRequestID(context.Background(), "req-123")

		buf.Reset()
		logger.Info(ctx, "test message", nil)

		var logEntry map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &logEntry)
		require.NoError(t, err)

		assert.Equal(t, "req-123", logEntry["request_id"])
	})

	t.Run("should not add request ID when not in context", func(t *testing.T) {
		ctx := context.Background()

		buf.Reset()
		logger.Info(ctx, "test message", nil)

		var logEntry map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &logEntry)
		require.NoError(t, err)

		_, exists := logEntry["request_id"]
		assert.False(t, exists)
	})
}

func TestLogger_StructuredOutput(t *testing.T) {
	var buf bytes.Buffer
	logger := &ZerologLogger{
		logger: zerolog.New(&buf).With().Timestamp().Str("service", "test-service").Logger(),
		serviceName: "test-service",
	}
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	ctx := context.Background()

	t.Run("should produce valid JSON output", func(t *testing.T) {
		buf.Reset()
		logger.Info(ctx, "test message", Tags{
			"key1": "value1",
			"key2": 42,
			"key3": true,
		})

		var logEntry map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &logEntry)
		require.NoError(t, err)

		assert.NotNil(t, logEntry["time"])
		assert.Equal(t, "info", logEntry["level"])
		assert.Equal(t, "test message", logEntry["message"])
		assert.Equal(t, "test-service", logEntry["service"])
		assert.Equal(t, "value1", logEntry["key1"])
		assert.Equal(t, float64(42), logEntry["key2"])
		assert.Equal(t, true, logEntry["key3"])
	})
}

func TestNewWithConfig(t *testing.T) {
	t.Run("should create logger with file output", func(t *testing.T) {
		tempFile, err := os.CreateTemp("", "logger_test_*.log")
		require.NoError(t, err)
		tempFilePath := tempFile.Name()
		tempFile.Close()
		defer os.Remove(tempFilePath)

		logger, err := NewWithConfig("test-service", "info", true, tempFilePath)
		require.NoError(t, err)
		assert.NotNil(t, logger)
		assert.Equal(t, "test-service", logger.serviceName)
	})

	t.Run("should create logger with stdout when file disabled", func(t *testing.T) {
		logger, err := NewWithConfig("test-service", "info", false, "")
		require.NoError(t, err)
		assert.NotNil(t, logger)
		assert.Equal(t, "test-service", logger.serviceName)
	})

	t.Run("should fail when file path is invalid", func(t *testing.T) {
		logger, err := NewWithConfig("test-service", "info", true, "/invalid/path/test.log")
		assert.Error(t, err)
		assert.Nil(t, logger)
	})
}

func TestNewConsoleLogger(t *testing.T) {
	t.Run("should create console logger", func(t *testing.T) {
		logger := NewConsoleLogger("test-service")
		assert.NotNil(t, logger)
		assert.Equal(t, "test-service", logger.serviceName)
	})
}

func TestLogger_LogLevelFiltering(t *testing.T) {
	var buf bytes.Buffer

	t.Run("should filter debug logs when level is info", func(t *testing.T) {
		buf.Reset()
		zerolog.SetGlobalLevel(zerolog.InfoLevel)

		logger := &ZerologLogger{
			logger: zerolog.New(&buf).With().Timestamp().Str("service", "test-service").Logger(),
			serviceName: "test-service",
		}

		logger.Debug(context.Background(), "debug message", nil)

		output := buf.String()
		assert.Empty(t, output, "Debug message should not be logged when level is Info")
	})

	t.Run("should log info when level is info", func(t *testing.T) {
		buf.Reset()
		zerolog.SetGlobalLevel(zerolog.InfoLevel)

		logger := &ZerologLogger{
			logger: zerolog.New(&buf).With().Timestamp().Str("service", "test-service").Logger(),
			serviceName: "test-service",
		}

		logger.Info(context.Background(), "info message", nil)

		output := buf.String()
		assert.NotEmpty(t, output)
		assert.Contains(t, output, "info message")
	})
}

func TestLogger_ComplexTags(t *testing.T) {
	var buf bytes.Buffer
	logger := &ZerologLogger{
		logger: zerolog.New(&buf).With().Timestamp().Str("service", "test-service").Logger(),
		serviceName: "test-service",
	}
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	ctx := context.Background()

	t.Run("should handle nested map tags", func(t *testing.T) {
		buf.Reset()
		logger.Info(ctx, "complex data", Tags{
			"user": map[string]interface{}{
				"id":   "123",
				"name": "Test User",
			},
		})

		output := buf.String()
		assert.Contains(t, output, "complex data")
		assert.Contains(t, output, "user")
	})

	t.Run("should handle array tags", func(t *testing.T) {
		buf.Reset()
		logger.Info(ctx, "array data", Tags{
			"items": []string{"item1", "item2", "item3"},
		})

		output := buf.String()
		assert.Contains(t, output, "array data")
		assert.Contains(t, output, "items")
	})
}

func TestLogger_EmptyTags(t *testing.T) {
	var buf bytes.Buffer
	logger := &ZerologLogger{
		logger: zerolog.New(&buf).With().Timestamp().Str("service", "test-service").Logger(),
		serviceName: "test-service",
	}
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	ctx := context.Background()

	t.Run("should handle nil tags gracefully", func(t *testing.T) {
		buf.Reset()
		logger.Info(ctx, "message with nil tags", nil)

		var logEntry map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &logEntry)
		require.NoError(t, err)

		assert.Equal(t, "info", logEntry["level"])
		assert.Equal(t, "message with nil tags", logEntry["message"])
	})

	t.Run("should handle empty tags map", func(t *testing.T) {
		buf.Reset()
		logger.Info(ctx, "message with empty tags", Tags{})

		var logEntry map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &logEntry)
		require.NoError(t, err)

		assert.Equal(t, "info", logEntry["level"])
		assert.Equal(t, "message with empty tags", logEntry["message"])
	})
}

func TestLogger_Timestamp(t *testing.T) {
	var buf bytes.Buffer
	logger := &ZerologLogger{
		logger: zerolog.New(&buf).With().Timestamp().Str("service", "test-service").Logger(),
		serviceName: "test-service",
	}
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	ctx := context.Background()

	t.Run("should include timestamp in log entry", func(t *testing.T) {
		buf.Reset()
		logger.Info(ctx, "timestamped message", nil)

		var logEntry map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &logEntry)
		require.NoError(t, err)

		_, exists := logEntry["time"]
		assert.True(t, exists, "Timestamp should be present in log entry")
	})
}

func TestLogger_ServiceName(t *testing.T) {
	var buf bytes.Buffer
	logger := &ZerologLogger{
		logger: zerolog.New(&buf).With().Timestamp().Str("service", "my-custom-service").Logger(),
		serviceName: "my-custom-service",
	}
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	ctx := context.Background()

	t.Run("should include service name in log entry", func(t *testing.T) {
		buf.Reset()
		logger.Info(ctx, "service test", nil)

		output := strings.TrimSpace(buf.String())
		assert.Contains(t, output, "my-custom-service")
	})
}
