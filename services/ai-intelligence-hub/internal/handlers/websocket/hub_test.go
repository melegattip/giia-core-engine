package websocket

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/melegattip/giia-core-engine/pkg/logger"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/core/providers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockLogger struct{ mock.Mock }

func (m *MockLogger) Debug(ctx context.Context, msg string, tags logger.Tags) {}
func (m *MockLogger) Info(ctx context.Context, msg string, tags logger.Tags)  {}
func (m *MockLogger) Warn(ctx context.Context, msg string, tags logger.Tags)  {}
func (m *MockLogger) Error(ctx context.Context, err error, msg string, tags logger.Tags) {
}
func (m *MockLogger) Fatal(ctx context.Context, err error, msg string, tags logger.Tags) {
}

type MockNotificationRepository struct{ mock.Mock }

func (m *MockNotificationRepository) Create(ctx context.Context, n *domain.AINotification) error {
	return m.Called(ctx, n).Error(0)
}
func (m *MockNotificationRepository) GetByID(ctx context.Context, id, orgID uuid.UUID) (*domain.AINotification, error) {
	args := m.Called(ctx, id, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AINotification), args.Error(1)
}
func (m *MockNotificationRepository) List(ctx context.Context, userID, orgID uuid.UUID, f *providers.NotificationFilters) ([]*domain.AINotification, error) {
	args := m.Called(ctx, userID, orgID, f)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.AINotification), args.Error(1)
}
func (m *MockNotificationRepository) Update(ctx context.Context, n *domain.AINotification) error {
	return m.Called(ctx, n).Error(0)
}
func (m *MockNotificationRepository) Delete(ctx context.Context, id, orgID uuid.UUID) error {
	return m.Called(ctx, id, orgID).Error(0)
}

func TestNewHub(t *testing.T) {
	hub := NewHub(nil, new(MockLogger))
	assert.NotNil(t, hub)
	assert.NotNil(t, hub.clients)
	assert.NotNil(t, hub.orgClients)
	assert.NotNil(t, hub.broadcast)
}

func TestHub_GetStats(t *testing.T) {
	hub := NewHub(nil, new(MockLogger))
	stats := hub.GetStats()
	assert.Contains(t, stats, "active_connections")
	assert.Contains(t, stats, "messages_sent")
}

func TestHub_BroadcastNotification(t *testing.T) {
	hub := NewHub(nil, new(MockLogger))
	userID, orgID := uuid.New(), uuid.New()
	notif := &domain.AINotification{
		ID: uuid.New(), UserID: userID, OrganizationID: orgID,
		Type: domain.NotificationTypeAlert, Priority: domain.NotificationPriorityHigh,
		Title: "Test", Summary: "Test", Status: domain.NotificationStatusUnread,
		CreatedAt: time.Now(),
	}
	hub.BroadcastNotification(notif)
	select {
	case msg := <-hub.broadcast:
		assert.Equal(t, notif.ID, msg.Notification.ID)
	default:
		t.Error("Expected message")
	}
}

func TestHub_Run(t *testing.T) {
	hub := NewHub(nil, new(MockLogger))
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	done := make(chan struct{})
	go func() { hub.Run(ctx); close(done) }()
	<-done
}

func TestHub_HandleWebSocket_MissingAuth(t *testing.T) {
	hub := NewHub(nil, new(MockLogger))
	req := httptest.NewRequest(http.MethodGet, "/ws/notifications", nil)
	rec := httptest.NewRecorder()
	hub.HandleWebSocket(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestNotificationPayload(t *testing.T) {
	notif := &domain.AINotification{
		ID: uuid.New(), UserID: uuid.New(), OrganizationID: uuid.New(),
		Type: domain.NotificationTypeAlert, Title: "Test", Summary: "Test",
		Status: domain.NotificationStatusUnread, CreatedAt: time.Now(),
	}
	payload := toNotificationPayload(notif)
	require.NotNil(t, payload)
	assert.Equal(t, notif.ID, payload.ID)
	assert.Nil(t, toNotificationPayload(nil))
}
