package events

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type PublisherMock struct {
	mock.Mock
}

func (m *PublisherMock) Publish(ctx context.Context, subject string, event *Event) error {
	args := m.Called(ctx, subject, event)
	return args.Error(0)
}

func (m *PublisherMock) PublishAsync(ctx context.Context, subject string, event *Event) error {
	args := m.Called(ctx, subject, event)
	return args.Error(0)
}

func (m *PublisherMock) Close() error {
	args := m.Called()
	return args.Error(0)
}
