package http

import (
	"bytes"
	"context"
	"encoding/json"
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

// MockLogger is a mock implementation of logger.Logger
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Debug(ctx context.Context, msg string, tags logger.Tags) {
	m.Called(ctx, msg, tags)
}

func (m *MockLogger) Info(ctx context.Context, msg string, tags logger.Tags) {
	m.Called(ctx, msg, tags)
}

func (m *MockLogger) Warn(ctx context.Context, msg string, tags logger.Tags) {
	m.Called(ctx, msg, tags)
}

func (m *MockLogger) Error(ctx context.Context, err error, msg string, tags logger.Tags) {
	m.Called(ctx, err, msg, tags)
}

func (m *MockLogger) Fatal(ctx context.Context, err error, msg string, tags logger.Tags) {
	m.Called(ctx, err, msg, tags)
}

// MockNotificationRepository is a mock implementation of providers.NotificationRepository
type MockNotificationRepository struct {
	mock.Mock
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

func createTestNotification(userID, orgID uuid.UUID) *domain.AINotification {
	return &domain.AINotification{
		ID:             uuid.New(),
		OrganizationID: orgID,
		UserID:         userID,
		Type:           domain.NotificationTypeAlert,
		Priority:       domain.NotificationPriorityHigh,
		Title:          "Test Notification",
		Summary:        "This is a test notification",
		Status:         domain.NotificationStatusUnread,
		CreatedAt:      time.Now(),
		SourceEvents:   []string{"test.event"},
		RelatedEntities: map[string][]string{
			"products": {"product-1"},
		},
		Recommendations: []domain.Recommendation{
			{
				Action:          "Take action",
				Reasoning:       "Because it helps",
				ExpectedOutcome: "Better results",
				Effort:          "low",
				Impact:          "high",
				PriorityOrder:   1,
			},
		},
	}
}

func TestNotificationHandler_ListNotifications(t *testing.T) {
	mockRepo := new(MockNotificationRepository)
	mockLogger := new(MockLogger)
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

	handler := NewNotificationHandler(mockRepo, mockLogger)

	userID := uuid.New()
	orgID := uuid.New()

	notifications := []*domain.AINotification{
		createTestNotification(userID, orgID),
		createTestNotification(userID, orgID),
	}

	t.Run("successful list", func(t *testing.T) {
		mockRepo.On("List", mock.Anything, userID, orgID, mock.AnythingOfType("*providers.NotificationFilters")).
			Return(notifications, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/v1/notifications?page=1&page_size=20", nil)
		req.Header.Set("X-User-ID", userID.String())
		req.Header.Set("X-Organization-ID", orgID.String())

		rec := httptest.NewRecorder()
		handler.ListNotifications(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var response NotificationListResponse
		err := json.NewDecoder(rec.Body).Decode(&response)
		require.NoError(t, err)

		assert.Len(t, response.Notifications, 2)
		assert.Equal(t, 1, response.Page)
	})

	t.Run("missing user ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/notifications", nil)
		req.Header.Set("X-Organization-ID", orgID.String())

		rec := httptest.NewRecorder()
		handler.ListNotifications(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("missing organization ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/notifications", nil)
		req.Header.Set("X-User-ID", userID.String())

		rec := httptest.NewRecorder()
		handler.ListNotifications(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("with filters", func(t *testing.T) {
		mockRepo.On("List", mock.Anything, userID, orgID, mock.AnythingOfType("*providers.NotificationFilters")).
			Return(notifications, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/v1/notifications?type=alert&priority=high&status=unread", nil)
		req.Header.Set("X-User-ID", userID.String())
		req.Header.Set("X-Organization-ID", orgID.String())

		rec := httptest.NewRecorder()
		handler.ListNotifications(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

func TestNotificationHandler_GetNotification(t *testing.T) {
	mockRepo := new(MockNotificationRepository)
	mockLogger := new(MockLogger)
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

	handler := NewNotificationHandler(mockRepo, mockLogger)

	userID := uuid.New()
	orgID := uuid.New()
	notification := createTestNotification(userID, orgID)

	t.Run("successful get", func(t *testing.T) {
		mockRepo.On("GetByID", mock.Anything, notification.ID, orgID).
			Return(notification, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/v1/notifications/"+notification.ID.String(), nil)
		req.Header.Set("X-Organization-ID", orgID.String())

		rec := httptest.NewRecorder()
		handler.GetNotification(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var response NotificationResponse
		err := json.NewDecoder(rec.Body).Decode(&response)
		require.NoError(t, err)

		assert.Equal(t, notification.ID, response.ID)
		assert.Equal(t, notification.Title, response.Title)
	})

	t.Run("not found", func(t *testing.T) {
		notFoundID := uuid.New()
		mockRepo.On("GetByID", mock.Anything, notFoundID, orgID).
			Return(nil, assert.AnError).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/v1/notifications/"+notFoundID.String(), nil)
		req.Header.Set("X-Organization-ID", orgID.String())

		rec := httptest.NewRecorder()
		handler.GetNotification(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("invalid ID format", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/notifications/invalid-id", nil)
		req.Header.Set("X-Organization-ID", orgID.String())

		rec := httptest.NewRecorder()
		handler.GetNotification(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestNotificationHandler_UpdateNotification(t *testing.T) {
	mockRepo := new(MockNotificationRepository)
	mockLogger := new(MockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

	handler := NewNotificationHandler(mockRepo, mockLogger)

	userID := uuid.New()
	orgID := uuid.New()
	notification := createTestNotification(userID, orgID)

	t.Run("mark as read", func(t *testing.T) {
		mockRepo.On("GetByID", mock.Anything, notification.ID, orgID).
			Return(notification, nil).Once()
		mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.AINotification")).
			Return(nil).Once()

		body, _ := json.Marshal(UpdateNotificationRequest{Status: "read"})
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/notifications/"+notification.ID.String(), bytes.NewReader(body))
		req.Header.Set("X-Organization-ID", orgID.String())
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		handler.UpdateNotification(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("mark as acted upon", func(t *testing.T) {
		mockRepo.On("GetByID", mock.Anything, notification.ID, orgID).
			Return(notification, nil).Once()
		mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.AINotification")).
			Return(nil).Once()

		body, _ := json.Marshal(UpdateNotificationRequest{Status: "acted_upon"})
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/notifications/"+notification.ID.String(), bytes.NewReader(body))
		req.Header.Set("X-Organization-ID", orgID.String())
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		handler.UpdateNotification(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("dismiss", func(t *testing.T) {
		mockRepo.On("GetByID", mock.Anything, notification.ID, orgID).
			Return(notification, nil).Once()
		mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.AINotification")).
			Return(nil).Once()

		body, _ := json.Marshal(UpdateNotificationRequest{Status: "dismissed"})
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/notifications/"+notification.ID.String(), bytes.NewReader(body))
		req.Header.Set("X-Organization-ID", orgID.String())
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		handler.UpdateNotification(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("invalid status", func(t *testing.T) {
		body, _ := json.Marshal(UpdateNotificationRequest{Status: "invalid"})
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/notifications/"+notification.ID.String(), bytes.NewReader(body))
		req.Header.Set("X-Organization-ID", orgID.String())
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		handler.UpdateNotification(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("invalid request body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/notifications/"+notification.ID.String(), bytes.NewReader([]byte("invalid json")))
		req.Header.Set("X-Organization-ID", orgID.String())
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		handler.UpdateNotification(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestNotificationHandler_DeleteNotification(t *testing.T) {
	mockRepo := new(MockNotificationRepository)
	mockLogger := new(MockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

	handler := NewNotificationHandler(mockRepo, mockLogger)

	orgID := uuid.New()
	notifID := uuid.New()

	t.Run("successful delete", func(t *testing.T) {
		mockRepo.On("Delete", mock.Anything, notifID, orgID).
			Return(nil).Once()

		req := httptest.NewRequest(http.MethodDelete, "/api/v1/notifications/"+notifID.String(), nil)
		req.Header.Set("X-Organization-ID", orgID.String())

		rec := httptest.NewRecorder()
		handler.DeleteNotification(rec, req)

		assert.Equal(t, http.StatusNoContent, rec.Code)
	})

	t.Run("delete error", func(t *testing.T) {
		failID := uuid.New()
		mockRepo.On("Delete", mock.Anything, failID, orgID).
			Return(assert.AnError).Once()

		req := httptest.NewRequest(http.MethodDelete, "/api/v1/notifications/"+failID.String(), nil)
		req.Header.Set("X-Organization-ID", orgID.String())

		rec := httptest.NewRecorder()
		handler.DeleteNotification(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestNotificationHandler_GetUnreadCount(t *testing.T) {
	mockRepo := new(MockNotificationRepository)
	mockLogger := new(MockLogger)
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

	handler := NewNotificationHandler(mockRepo, mockLogger)

	userID := uuid.New()
	orgID := uuid.New()

	notifications := []*domain.AINotification{
		{
			ID:       uuid.New(),
			UserID:   userID,
			Status:   domain.NotificationStatusUnread,
			Priority: domain.NotificationPriorityHigh,
		},
		{
			ID:       uuid.New(),
			UserID:   userID,
			Status:   domain.NotificationStatusUnread,
			Priority: domain.NotificationPriorityCritical,
		},
	}

	t.Run("successful count", func(t *testing.T) {
		mockRepo.On("List", mock.Anything, userID, orgID, mock.AnythingOfType("*providers.NotificationFilters")).
			Return(notifications, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/v1/notifications/unread-count", nil)
		req.Header.Set("X-User-ID", userID.String())
		req.Header.Set("X-Organization-ID", orgID.String())

		rec := httptest.NewRecorder()
		handler.GetUnreadCount(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var response UnreadCountResponse
		err := json.NewDecoder(rec.Body).Decode(&response)
		require.NoError(t, err)

		assert.Equal(t, 2, response.UnreadCount)
		assert.Equal(t, 1, response.CountByPriority["high"])
		assert.Equal(t, 1, response.CountByPriority["critical"])
	})
}

func TestNotificationCache(t *testing.T) {
	cache := NewNotificationCache(100 * time.Millisecond)

	t.Run("set and get", func(t *testing.T) {
		cache.Set("key1", "value1")

		value, found := cache.Get("key1")
		assert.True(t, found)
		assert.Equal(t, "value1", value)
	})

	t.Run("get non-existent key", func(t *testing.T) {
		_, found := cache.Get("non-existent")
		assert.False(t, found)
	})

	t.Run("expired entry", func(t *testing.T) {
		cache.Set("expiring", "value")

		time.Sleep(150 * time.Millisecond)

		_, found := cache.Get("expiring")
		assert.False(t, found)
	})

	t.Run("delete", func(t *testing.T) {
		cache.Set("delete-me", "value")
		cache.Delete("delete-me")

		_, found := cache.Get("delete-me")
		assert.False(t, found)
	})

	t.Run("clear", func(t *testing.T) {
		cache.Set("key1", "value1")
		cache.Set("key2", "value2")
		cache.Clear()

		_, found1 := cache.Get("key1")
		_, found2 := cache.Get("key2")

		assert.False(t, found1)
		assert.False(t, found2)
	})
}
