package logger

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type LoggerMock struct {
	mock.Mock
}

func (m *LoggerMock) Debug(ctx context.Context, msg string, tags Tags) {
	m.Called(ctx, msg, tags)
}

func (m *LoggerMock) Info(ctx context.Context, msg string, tags Tags) {
	m.Called(ctx, msg, tags)
}

func (m *LoggerMock) Warn(ctx context.Context, msg string, tags Tags) {
	m.Called(ctx, msg, tags)
}

func (m *LoggerMock) Error(ctx context.Context, err error, msg string, tags Tags) {
	m.Called(ctx, err, msg, tags)
}

func (m *LoggerMock) Fatal(ctx context.Context, err error, msg string, tags Tags) {
	m.Called(ctx, err, msg, tags)
}
