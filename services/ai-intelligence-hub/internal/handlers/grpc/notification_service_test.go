package grpc

import (
	"context"
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

func createTestNotification(userID, orgID uuid.UUID) *domain.AINotification {
	return &domain.AINotification{
		ID: uuid.New(), OrganizationID: orgID, UserID: userID,
		Type: domain.NotificationTypeAlert, Priority: domain.NotificationPriorityHigh,
		Title: "Test", Summary: "Test summary", Status: domain.NotificationStatusUnread,
		CreatedAt: time.Now(), SourceEvents: []string{"test"},
		Recommendations: []domain.Recommendation{{Action: "Act", PriorityOrder: 1}},
	}
}

func TestNotificationService_GetNotification(t *testing.T) {
	mockRepo := new(MockNotificationRepository)
	svc := NewNotificationService(mockRepo, nil, new(MockLogger))

	userID, orgID := uuid.New(), uuid.New()
	notif := createTestNotification(userID, orgID)

	t.Run("success", func(t *testing.T) {
		mockRepo.On("GetByID", mock.Anything, notif.ID, orgID).Return(notif, nil).Once()
		resp, err := svc.GetNotification(context.Background(), &GetNotificationRequest{
			NotificationID: notif.ID.String(), OrganizationID: orgID.String(),
		})
		require.NoError(t, err)
		assert.Equal(t, notif.ID.String(), resp.Notification.ID)
	})

	t.Run("missing notification_id", func(t *testing.T) {
		_, err := svc.GetNotification(context.Background(), &GetNotificationRequest{
			OrganizationID: orgID.String(),
		})
		assert.Error(t, err)
	})

	t.Run("invalid notification_id", func(t *testing.T) {
		_, err := svc.GetNotification(context.Background(), &GetNotificationRequest{
			NotificationID: "invalid", OrganizationID: orgID.String(),
		})
		assert.Error(t, err)
	})
}

func TestNotificationService_ListNotifications(t *testing.T) {
	mockRepo := new(MockNotificationRepository)
	svc := NewNotificationService(mockRepo, nil, new(MockLogger))

	userID, orgID := uuid.New(), uuid.New()
	notifications := []*domain.AINotification{createTestNotification(userID, orgID)}

	t.Run("success", func(t *testing.T) {
		mockRepo.On("List", mock.Anything, userID, orgID, mock.Anything).Return(notifications, nil).Once()
		resp, err := svc.ListNotifications(context.Background(), &ListNotificationsRequest{
			UserID: userID.String(), OrganizationID: orgID.String(), PageSize: 20,
		})
		require.NoError(t, err)
		assert.Len(t, resp.Notifications, 1)
	})

	t.Run("missing user_id", func(t *testing.T) {
		_, err := svc.ListNotifications(context.Background(), &ListNotificationsRequest{
			OrganizationID: orgID.String(),
		})
		assert.Error(t, err)
	})
}

func TestNotificationService_CreateNotification(t *testing.T) {
	mockRepo := new(MockNotificationRepository)
	svc := NewNotificationService(mockRepo, nil, new(MockLogger))

	userID, orgID := uuid.New(), uuid.New()

	t.Run("success", func(t *testing.T) {
		mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.AINotification")).Return(nil).Once()
		resp, err := svc.CreateNotification(context.Background(), &CreateNotificationRequest{
			OrganizationID: orgID.String(), UserID: userID.String(),
			Type: "alert", Priority: "high", Title: "Test", Summary: "Test",
		})
		require.NoError(t, err)
		assert.NotEmpty(t, resp.Notification.ID)
	})

	t.Run("missing required fields", func(t *testing.T) {
		_, err := svc.CreateNotification(context.Background(), &CreateNotificationRequest{
			OrganizationID: orgID.String(), UserID: userID.String(),
		})
		assert.Error(t, err)
	})
}

func TestNotificationService_UpdateNotificationStatus(t *testing.T) {
	mockRepo := new(MockNotificationRepository)
	svc := NewNotificationService(mockRepo, nil, new(MockLogger))

	userID, orgID := uuid.New(), uuid.New()
	notif := createTestNotification(userID, orgID)

	t.Run("mark as read", func(t *testing.T) {
		mockRepo.On("GetByID", mock.Anything, notif.ID, orgID).Return(notif, nil).Once()
		mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.AINotification")).Return(nil).Once()
		resp, err := svc.UpdateNotificationStatus(context.Background(), &UpdateNotificationStatusRequest{
			NotificationID: notif.ID.String(), OrganizationID: orgID.String(), Status: "read",
		})
		require.NoError(t, err)
		assert.Equal(t, "read", resp.Notification.Status)
	})

	t.Run("invalid status", func(t *testing.T) {
		_, err := svc.UpdateNotificationStatus(context.Background(), &UpdateNotificationStatusRequest{
			NotificationID: notif.ID.String(), OrganizationID: orgID.String(), Status: "invalid",
		})
		assert.Error(t, err)
	})
}

func TestNotificationService_DeleteNotification(t *testing.T) {
	mockRepo := new(MockNotificationRepository)
	svc := NewNotificationService(mockRepo, nil, new(MockLogger))

	orgID := uuid.New()
	notifID := uuid.New()

	t.Run("success", func(t *testing.T) {
		mockRepo.On("Delete", mock.Anything, notifID, orgID).Return(nil).Once()
		resp, err := svc.DeleteNotification(context.Background(), &DeleteNotificationRequest{
			NotificationID: notifID.String(), OrganizationID: orgID.String(),
		})
		require.NoError(t, err)
		assert.True(t, resp.Success)
	})
}

func TestNotificationService_GetUnreadCount(t *testing.T) {
	mockRepo := new(MockNotificationRepository)
	svc := NewNotificationService(mockRepo, nil, new(MockLogger))

	userID, orgID := uuid.New(), uuid.New()
	notifications := []*domain.AINotification{
		{ID: uuid.New(), Status: domain.NotificationStatusUnread, Priority: domain.NotificationPriorityHigh},
		{ID: uuid.New(), Status: domain.NotificationStatusUnread, Priority: domain.NotificationPriorityCritical},
	}

	t.Run("success", func(t *testing.T) {
		mockRepo.On("List", mock.Anything, userID, orgID, mock.Anything).Return(notifications, nil).Once()
		resp, err := svc.GetUnreadCount(context.Background(), &GetUnreadCountRequest{
			UserID: userID.String(), OrganizationID: orgID.String(),
		})
		require.NoError(t, err)
		assert.Equal(t, int32(2), resp.Count)
		assert.Equal(t, int32(1), resp.CountByPriority["high"])
	})
}

func TestToProtoNotification(t *testing.T) {
	svc := NewNotificationService(nil, nil, new(MockLogger))
	notif := createTestNotification(uuid.New(), uuid.New())
	notif.Impact = domain.ImpactAssessment{RiskLevel: "high", AffectedOrders: 5}

	proto := svc.toProtoNotification(notif)
	require.NotNil(t, proto)
	assert.Equal(t, notif.ID.String(), proto.ID)
	assert.Equal(t, notif.Title, proto.Title)
	assert.NotNil(t, proto.Impact)
	assert.Nil(t, svc.toProtoNotification(nil))
}

func TestIsValidStatus(t *testing.T) {
	assert.True(t, isValidStatus(domain.NotificationStatusRead))
	assert.True(t, isValidStatus(domain.NotificationStatusActedUpon))
	assert.True(t, isValidStatus(domain.NotificationStatusDismissed))
	assert.False(t, isValidStatus(domain.NotificationStatus("invalid")))
}
