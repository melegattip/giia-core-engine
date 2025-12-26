package logger

import (
	"context"
	"os"
	"time"

	"github.com/rs/zerolog"
)

type Tags map[string]interface{}

type Logger interface {
	Debug(ctx context.Context, msg string, tags Tags)
	Info(ctx context.Context, msg string, tags Tags)
	Warn(ctx context.Context, msg string, tags Tags)
	Error(ctx context.Context, err error, msg string, tags Tags)
	Fatal(ctx context.Context, err error, msg string, tags Tags)
}

type ZerologLogger struct {
	logger      zerolog.Logger
	serviceName string
}

func New(serviceName string, logLevel string) *ZerologLogger {
	level := parseLogLevel(logLevel)
	zerolog.SetGlobalLevel(level)

	logger := zerolog.New(os.Stdout).
		With().
		Timestamp().
		Str("service", serviceName).
		Logger()

	return &ZerologLogger{
		logger:      logger,
		serviceName: serviceName,
	}
}

func parseLogLevel(level string) zerolog.Level {
	switch level {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	default:
		return zerolog.InfoLevel
	}
}

func (l *ZerologLogger) Debug(ctx context.Context, msg string, tags Tags) {
	event := l.logger.Debug()
	event = l.addContextFields(ctx, event, tags)
	event.Msg(msg)
}

func (l *ZerologLogger) Info(ctx context.Context, msg string, tags Tags) {
	event := l.logger.Info()
	event = l.addContextFields(ctx, event, tags)
	event.Msg(msg)
}

func (l *ZerologLogger) Warn(ctx context.Context, msg string, tags Tags) {
	event := l.logger.Warn()
	event = l.addContextFields(ctx, event, tags)
	event.Msg(msg)
}

func (l *ZerologLogger) Error(ctx context.Context, err error, msg string, tags Tags) {
	event := l.logger.Error().Err(err)
	event = l.addContextFields(ctx, event, tags)
	event.Msg(msg)
}

func (l *ZerologLogger) Fatal(ctx context.Context, err error, msg string, tags Tags) {
	event := l.logger.Fatal().Err(err)
	event = l.addContextFields(ctx, event, tags)
	event.Msg(msg)
}

func (l *ZerologLogger) addContextFields(ctx context.Context, event *zerolog.Event, tags Tags) *zerolog.Event {
	if ctx != nil {
		if requestID := ExtractRequestID(ctx); requestID != "" {
			event = event.Str("request_id", requestID)
		}
	}

	if tags != nil {
		for key, value := range tags {
			event = event.Interface(key, value)
		}
	}

	return event
}

func NewWithConfig(serviceName string, logLevel string, logToFile bool, logFilePath string) (*ZerologLogger, error) {
	level := parseLogLevel(logLevel)
	zerolog.SetGlobalLevel(level)

	var output = os.Stdout
	if logToFile && logFilePath != "" {
		file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, err
		}
		output = file
	}

	logger := zerolog.New(output).
		With().
		Timestamp().
		Str("service", serviceName).
		Logger()

	return &ZerologLogger{
		logger:      logger,
		serviceName: serviceName,
	}, nil
}

func NewConsoleLogger(serviceName string) *ZerologLogger {
	logger := zerolog.New(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}).
		With().
		Timestamp().
		Str("service", serviceName).
		Logger()

	return &ZerologLogger{
		logger:      logger,
		serviceName: serviceName,
	}
}
