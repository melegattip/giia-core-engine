package events

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type SubscriberMock struct {
	mock.Mock
}

func (m *SubscriberMock) Subscribe(ctx context.Context, subject string, handler EventHandler) error {
	args := m.Called(ctx, subject, handler)
	return args.Error(0)
}

func (m *SubscriberMock) SubscribeDurable(ctx context.Context, subject, durableName string, handler EventHandler) error {
	args := m.Called(ctx, subject, durableName, handler)
	return args.Error(0)
}

func (m *SubscriberMock) Close() error {
	args := m.Called()
	return args.Error(0)
}
