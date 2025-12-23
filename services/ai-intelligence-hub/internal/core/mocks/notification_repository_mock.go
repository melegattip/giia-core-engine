package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/core/providers"
	"github.com/stretchr/testify/mock"
)

// MockNotificationRepository is a mock implementation of providers.NotificationRepository
type MockNotificationRepository struct {
	mock.Mock
}

func NewMockNotificationRepository() *MockNotificationRepository {
	return &MockNotificationRepository{}
}

func (m *MockNotificationRepository) Create(ctx context.Context, notification *domain.AINotification) error {
	args := m.Called(ctx, notification)
	return args.Error(0)
}

func (m *MockNotificationRepository) GetByID(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) (*domain.AINotification, error) {
	args := m.Called(ctx, id, organizationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AINotification), args.Error(1)
}

func (m *MockNotificationRepository) List(ctx context.Context, userID uuid.UUID, organizationID uuid.UUID, filters *providers.NotificationFilters) ([]*domain.AINotification, error) {
	args := m.Called(ctx, userID, organizationID, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.AINotification), args.Error(1)
}

func (m *MockNotificationRepository) Update(ctx context.Context, notification *domain.AINotification) error {
	args := m.Called(ctx, notification)
	return args.Error(0)
}

func (m *MockNotificationRepository) Delete(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) error {
	args := m.Called(ctx, id, organizationID)
	return args.Error(0)
}
