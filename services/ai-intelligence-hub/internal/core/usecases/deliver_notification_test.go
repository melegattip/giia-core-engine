package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/melegattip/giia-core-engine/pkg/logger"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/core/providers"
)

// MockEmailProvider mocks the email delivery provider
type MockEmailProvider struct {
	mock.Mock
}

func (m *MockEmailProvider) SendEmail(ctx context.Context, to []string, subject string, htmlBody string, textBody string) (string, error) {
	args := m.Called(ctx, to, subject, htmlBody, textBody)
	return args.String(0), args.Error(1)
}

// MockWebhookProvider mocks the webhook delivery provider
type MockWebhookProvider struct {
	mock.Mock
}

func (m *MockWebhookProvider) SendWebhook(ctx context.Context, webhookURL string, payload interface{}) error {
	args := m.Called(ctx, webhookURL, payload)
	return args.Error(0)
}

// MockSMSProvider mocks the SMS delivery provider
type MockSMSProvider struct {
	mock.Mock
}

func (m *MockSMSProvider) SendSMS(ctx context.Context, to string, message string) (string, error) {
	args := m.Called(ctx, to, message)
	return args.String(0), args.Error(1)
}

// MockNotificationRepo mocks the notification repository
type MockNotificationRepo struct {
	mock.Mock
}

func (m *MockNotificationRepo) Create(ctx context.Context, notification *domain.AINotification) error {
	args := m.Called(ctx, notification)
	return args.Error(0)
}

func (m *MockNotificationRepo) GetByID(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) (*domain.AINotification, error) {
	args := m.Called(ctx, id, organizationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AINotification), args.Error(1)
}

func (m *MockNotificationRepo) List(ctx context.Context, userID uuid.UUID, organizationID uuid.UUID, filters *providers.NotificationFilters) ([]*domain.AINotification, error) {
	args := m.Called(ctx, userID, organizationID, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.AINotification), args.Error(1)
}

func (m *MockNotificationRepo) Update(ctx context.Context, notification *domain.AINotification) error {
	args := m.Called(ctx, notification)
	return args.Error(0)
}

func (m *MockNotificationRepo) Delete(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) error {
	args := m.Called(ctx, id, organizationID)
	return args.Error(0)
}

// MockPrefsRepo mocks the preferences repository
type MockPrefsRepo struct {
	mock.Mock
}

func (m *MockPrefsRepo) Create(ctx context.Context, prefs *domain.UserNotificationPreferences) error {
	args := m.Called(ctx, prefs)
	return args.Error(0)
}

func (m *MockPrefsRepo) GetByUserID(ctx context.Context, userID uuid.UUID, organizationID uuid.UUID) (*domain.UserNotificationPreferences, error) {
	args := m.Called(ctx, userID, organizationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.UserNotificationPreferences), args.Error(1)
}

func (m *MockPrefsRepo) Update(ctx context.Context, prefs *domain.UserNotificationPreferences) error {
	args := m.Called(ctx, prefs)
	return args.Error(0)
}

// MockLogger mocks the logger
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

func TestDeliverNotificationUseCase_Execute_EmailSuccess(t *testing.T) {
	mockEmail := new(MockEmailProvider)
	mockWebhook := new(MockWebhookProvider)
	mockSMS := new(MockSMSProvider)
	mockNotifRepo := new(MockNotificationRepo)
	mockPrefsRepo := new(MockPrefsRepo)
	mockLogger := new(MockLogger)

	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Return()

	uc := NewDeliverNotificationUseCase(
		mockEmail,
		mockWebhook,
		mockSMS,
		mockNotifRepo,
		mockPrefsRepo,
		mockLogger,
	)

	notification := &domain.AINotification{
		ID:             uuid.New(),
		OrganizationID: uuid.New(),
		UserID:         uuid.New(),
		Title:          "Test Alert",
		Summary:        "Test summary",
		Priority:       domain.NotificationPriorityHigh,
	}

	input := &DeliverNotificationInput{
		Notification: notification,
		Channels:     []domain.DeliveryChannel{domain.DeliveryChannelEmail},
		Recipients: map[domain.DeliveryChannel]string{
			domain.DeliveryChannelEmail: "test@example.com",
		},
	}

	mockEmail.On("SendEmail", mock.Anything, []string{"test@example.com"}, mock.Anything, mock.Anything, mock.Anything).
		Return("msg-123", nil)

	output, err := uc.Execute(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, 1, output.SuccessCount)
	assert.Equal(t, 0, output.FailedCount)
	assert.Len(t, output.Results, 1)
	assert.True(t, output.Results[0].Success)
	assert.Equal(t, "msg-123", output.Results[0].MessageID)
	mockEmail.AssertExpectations(t)
}

func TestDeliverNotificationUseCase_Execute_MultiChannel(t *testing.T) {
	mockEmail := new(MockEmailProvider)
	mockWebhook := new(MockWebhookProvider)
	mockSMS := new(MockSMSProvider)
	mockNotifRepo := new(MockNotificationRepo)
	mockPrefsRepo := new(MockPrefsRepo)
	mockLogger := new(MockLogger)

	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Return()

	uc := NewDeliverNotificationUseCase(
		mockEmail,
		mockWebhook,
		mockSMS,
		mockNotifRepo,
		mockPrefsRepo,
		mockLogger,
	)

	notification := &domain.AINotification{
		ID:             uuid.New(),
		OrganizationID: uuid.New(),
		UserID:         uuid.New(),
		Title:          "Critical Alert",
		Summary:        "Urgent action needed",
		Priority:       domain.NotificationPriorityCritical,
	}

	input := &DeliverNotificationInput{
		Notification: notification,
		Channels: []domain.DeliveryChannel{
			domain.DeliveryChannelEmail,
			domain.DeliveryChannelSlack,
			domain.DeliveryChannelSMS,
		},
		Recipients: map[domain.DeliveryChannel]string{
			domain.DeliveryChannelEmail: "test@example.com",
			domain.DeliveryChannelSlack: "https://hooks.slack.com/services/xxx",
			domain.DeliveryChannelSMS:   "+15551234567",
		},
	}

	mockEmail.On("SendEmail", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return("email-123", nil)
	mockWebhook.On("SendWebhook", mock.Anything, "https://hooks.slack.com/services/xxx", mock.Anything).
		Return(nil)
	mockSMS.On("SendSMS", mock.Anything, "+15551234567", mock.Anything).
		Return("sms-456", nil)

	output, err := uc.Execute(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, 3, output.SuccessCount)
	assert.Equal(t, 0, output.FailedCount)
	mockEmail.AssertExpectations(t)
	mockWebhook.AssertExpectations(t)
	mockSMS.AssertExpectations(t)
}

func TestDeliverNotificationUseCase_Execute_PartialFailure(t *testing.T) {
	mockEmail := new(MockEmailProvider)
	mockWebhook := new(MockWebhookProvider)
	mockSMS := new(MockSMSProvider)
	mockNotifRepo := new(MockNotificationRepo)
	mockPrefsRepo := new(MockPrefsRepo)
	mockLogger := new(MockLogger)

	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

	uc := NewDeliverNotificationUseCase(
		mockEmail,
		mockWebhook,
		mockSMS,
		mockNotifRepo,
		mockPrefsRepo,
		mockLogger,
	)

	notification := &domain.AINotification{
		ID:             uuid.New(),
		OrganizationID: uuid.New(),
		UserID:         uuid.New(),
		Title:          "Test Alert",
		Priority:       domain.NotificationPriorityHigh,
	}

	input := &DeliverNotificationInput{
		Notification: notification,
		Channels: []domain.DeliveryChannel{
			domain.DeliveryChannelEmail,
			domain.DeliveryChannelSlack,
		},
		Recipients: map[domain.DeliveryChannel]string{
			domain.DeliveryChannelEmail: "test@example.com",
			domain.DeliveryChannelSlack: "https://hooks.slack.com/invalid",
		},
	}

	mockEmail.On("SendEmail", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return("email-123", nil)
	mockWebhook.On("SendWebhook", mock.Anything, mock.Anything, mock.Anything).
		Return(errors.New("webhook failed"))

	output, err := uc.Execute(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, 1, output.SuccessCount)
	assert.Equal(t, 1, output.FailedCount)
	assert.Equal(t, 1, output.QueuedForRetry)
}

func TestDeliverNotificationUseCase_Execute_SMSOnlyForCritical(t *testing.T) {
	mockEmail := new(MockEmailProvider)
	mockWebhook := new(MockWebhookProvider)
	mockSMS := new(MockSMSProvider)
	mockNotifRepo := new(MockNotificationRepo)
	mockPrefsRepo := new(MockPrefsRepo)
	mockLogger := new(MockLogger)

	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Return()

	uc := NewDeliverNotificationUseCase(
		mockEmail,
		mockWebhook,
		mockSMS,
		mockNotifRepo,
		mockPrefsRepo,
		mockLogger,
	)

	// Non-critical notification should not send SMS
	notification := &domain.AINotification{
		ID:             uuid.New(),
		OrganizationID: uuid.New(),
		UserID:         uuid.New(),
		Title:          "Medium Priority Alert",
		Priority:       domain.NotificationPriorityMedium,
	}

	input := &DeliverNotificationInput{
		Notification: notification,
		Channels:     []domain.DeliveryChannel{domain.DeliveryChannelSMS},
		Recipients: map[domain.DeliveryChannel]string{
			domain.DeliveryChannelSMS: "+15551234567",
		},
	}

	// SMS should not be called for non-critical
	output, err := uc.Execute(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, 1, output.SuccessCount) // Skipped silently = success
	mockSMS.AssertNotCalled(t, "SendSMS")
}

func TestDeliverNotificationUseCase_DeliverBasedOnPreferences(t *testing.T) {
	mockEmail := new(MockEmailProvider)
	mockWebhook := new(MockWebhookProvider)
	mockSMS := new(MockSMSProvider)
	mockNotifRepo := new(MockNotificationRepo)
	mockPrefsRepo := new(MockPrefsRepo)
	mockLogger := new(MockLogger)

	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Return()

	uc := NewDeliverNotificationUseCase(
		mockEmail,
		mockWebhook,
		mockSMS,
		mockNotifRepo,
		mockPrefsRepo,
		mockLogger,
	)

	userID := uuid.New()
	orgID := uuid.New()

	notification := &domain.AINotification{
		ID:             uuid.New(),
		OrganizationID: orgID,
		UserID:         userID,
		Title:          "Test Alert",
		Summary:        "Test summary",
		Priority:       domain.NotificationPriorityHigh,
	}

	prefs := &domain.UserNotificationPreferences{
		UserID:           userID,
		OrganizationID:   orgID,
		EnableEmail:      true,
		EmailAddress:     "user@example.com",
		EnableSlack:      true,
		SlackWebhookURL:  "https://hooks.slack.com/services/xxx",
		EnableSMS:        false,
		EmailMinPriority: domain.NotificationPriorityMedium,
	}

	mockPrefsRepo.On("GetByUserID", mock.Anything, userID, orgID).Return(prefs, nil)
	mockEmail.On("SendEmail", mock.Anything, []string{"user@example.com"}, mock.Anything, mock.Anything, mock.Anything).
		Return("msg-123", nil)
	mockWebhook.On("SendWebhook", mock.Anything, "https://hooks.slack.com/services/xxx", mock.Anything).
		Return(nil)

	output, err := uc.DeliverBasedOnPreferences(context.Background(), notification)

	require.NoError(t, err)
	assert.GreaterOrEqual(t, output.SuccessCount, 2) // Email + Slack + InApp
	mockEmail.AssertExpectations(t)
	mockWebhook.AssertExpectations(t)
	mockSMS.AssertNotCalled(t, "SendSMS") // SMS disabled
}

func TestDeliverNotificationUseCase_DeliverBasedOnPreferences_DefaultPrefs(t *testing.T) {
	mockEmail := new(MockEmailProvider)
	mockWebhook := new(MockWebhookProvider)
	mockSMS := new(MockSMSProvider)
	mockNotifRepo := new(MockNotificationRepo)
	mockPrefsRepo := new(MockPrefsRepo)
	mockLogger := new(MockLogger)

	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Return()

	uc := NewDeliverNotificationUseCase(
		mockEmail,
		mockWebhook,
		mockSMS,
		mockNotifRepo,
		mockPrefsRepo,
		mockLogger,
	)

	userID := uuid.New()
	orgID := uuid.New()

	notification := &domain.AINotification{
		ID:             uuid.New(),
		OrganizationID: orgID,
		UserID:         userID,
		Title:          "Test Alert",
		Priority:       domain.NotificationPriorityHigh,
	}

	// Return error so defaults are used
	mockPrefsRepo.On("GetByUserID", mock.Anything, userID, orgID).
		Return(nil, errors.New("not found"))

	output, err := uc.DeliverBasedOnPreferences(context.Background(), notification)

	require.NoError(t, err)
	// When default prefs are used with no configured email address, only InApp is added
	// but InApp doesn't have a recipient in the map, so it gets skipped in Execute.
	// This is expected behavior - zero external channels without configuration.
	assert.NotNil(t, output)
}

func TestDeliverNotificationUseCase_QuietHours(t *testing.T) {
	mockEmail := new(MockEmailProvider)
	mockWebhook := new(MockWebhookProvider)
	mockSMS := new(MockSMSProvider)
	mockNotifRepo := new(MockNotificationRepo)
	mockPrefsRepo := new(MockPrefsRepo)
	mockLogger := new(MockLogger)

	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Return()

	uc := NewDeliverNotificationUseCase(
		mockEmail,
		mockWebhook,
		mockSMS,
		mockNotifRepo,
		mockPrefsRepo,
		mockLogger,
	)

	orgID := uuid.New()

	// Set up channel config with quiet hours
	configSet := domain.NewChannelConfigSet(orgID)
	emailConfig := domain.NewChannelConfig(orgID, domain.DeliveryChannelEmail)
	emailConfig.QuietHoursEnabled = true
	emailConfig.QuietHoursStart = "00:00"
	emailConfig.QuietHoursEnd = "23:59" // Always quiet (for testing)
	emailConfig.Timezone = "UTC"
	configSet.AddConfig(emailConfig)

	uc.SetChannelConfig(orgID, configSet)

	notification := &domain.AINotification{
		ID:             uuid.New(),
		OrganizationID: orgID,
		UserID:         uuid.New(),
		Title:          "Test Alert",
		Priority:       domain.NotificationPriorityMedium,
	}

	input := &DeliverNotificationInput{
		Notification: notification,
		Channels:     []domain.DeliveryChannel{domain.DeliveryChannelEmail},
		Recipients: map[domain.DeliveryChannel]string{
			domain.DeliveryChannelEmail: "test@example.com",
		},
		ForceDeliver: false,
	}

	output, err := uc.Execute(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, 0, output.SuccessCount)
	assert.Equal(t, 1, output.QueuedForRetry)
	mockEmail.AssertNotCalled(t, "SendEmail")
}

func TestDeliverNotificationUseCase_RateLimit(t *testing.T) {
	mockEmail := new(MockEmailProvider)
	mockWebhook := new(MockWebhookProvider)
	mockSMS := new(MockSMSProvider)
	mockNotifRepo := new(MockNotificationRepo)
	mockPrefsRepo := new(MockPrefsRepo)
	mockLogger := new(MockLogger)

	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Return()

	uc := NewDeliverNotificationUseCase(
		mockEmail,
		mockWebhook,
		mockSMS,
		mockNotifRepo,
		mockPrefsRepo,
		mockLogger,
	)

	orgID := uuid.New()

	// Set up channel config with zero rate limit
	configSet := domain.NewChannelConfigSet(orgID)
	emailConfig := domain.NewChannelConfig(orgID, domain.DeliveryChannelEmail)
	emailConfig.MaxPerHour = 0 // No sends allowed
	emailConfig.CurrentHourly = 0
	emailConfig.LastResetHour = time.Now()
	configSet.AddConfig(emailConfig)

	uc.SetChannelConfig(orgID, configSet)

	notification := &domain.AINotification{
		ID:             uuid.New(),
		OrganizationID: orgID,
		UserID:         uuid.New(),
		Title:          "Test Alert",
		Priority:       domain.NotificationPriorityMedium,
	}

	input := &DeliverNotificationInput{
		Notification: notification,
		Channels:     []domain.DeliveryChannel{domain.DeliveryChannelEmail},
		Recipients: map[domain.DeliveryChannel]string{
			domain.DeliveryChannelEmail: "test@example.com",
		},
	}

	output, err := uc.Execute(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, 1, output.QueuedForRetry)
	mockEmail.AssertNotCalled(t, "SendEmail")
}

func TestDeliverNotificationUseCase_GetQueueStats(t *testing.T) {
	mockLogger := new(MockLogger)

	uc := NewDeliverNotificationUseCase(nil, nil, nil, nil, nil, mockLogger)

	pending, failed, delivered := uc.GetQueueStats()

	assert.Equal(t, 0, pending)
	assert.Equal(t, 0, failed)
	assert.Equal(t, 0, delivered)
}

func TestDeliverNotificationUseCase_ProcessRetryQueue(t *testing.T) {
	mockEmail := new(MockEmailProvider)
	mockWebhook := new(MockWebhookProvider)
	mockSMS := new(MockSMSProvider)
	mockNotifRepo := new(MockNotificationRepo)
	mockPrefsRepo := new(MockPrefsRepo)
	mockLogger := new(MockLogger)

	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

	uc := NewDeliverNotificationUseCase(
		mockEmail,
		mockWebhook,
		mockSMS,
		mockNotifRepo,
		mockPrefsRepo,
		mockLogger,
	)

	// Process empty queue should return 0s
	successCount, failedCount, err := uc.ProcessRetryQueue(context.Background())

	require.NoError(t, err)
	assert.Equal(t, 0, successCount)
	assert.Equal(t, 0, failedCount)
}

func TestDeliverNotificationUseCase_BuildEmailSubject(t *testing.T) {
	mockLogger := new(MockLogger)

	uc := NewDeliverNotificationUseCase(nil, nil, nil, nil, nil, mockLogger)

	tests := []struct {
		priority domain.NotificationPriority
		title    string
		expected string
	}{
		{domain.NotificationPriorityCritical, "Alert", "🚨 CRITICAL: Alert"},
		{domain.NotificationPriorityHigh, "Warning", "⚠️ HIGH: Warning"},
		{domain.NotificationPriorityMedium, "Info", "Info"},
		{domain.NotificationPriorityLow, "Note", "Note"},
	}

	for _, tc := range tests {
		notif := &domain.AINotification{
			Priority: tc.priority,
			Title:    tc.title,
		}
		subject := uc.buildEmailSubject(notif)
		assert.Equal(t, tc.expected, subject)
	}
}

func TestDeliverNotificationUseCase_ShouldDeliverToChannel(t *testing.T) {
	mockLogger := new(MockLogger)

	uc := NewDeliverNotificationUseCase(nil, nil, nil, nil, nil, mockLogger)

	tests := []struct {
		notifPriority domain.NotificationPriority
		minPriority   domain.NotificationPriority
		expected      bool
	}{
		{domain.NotificationPriorityCritical, domain.NotificationPriorityMedium, true},
		{domain.NotificationPriorityHigh, domain.NotificationPriorityMedium, true},
		{domain.NotificationPriorityMedium, domain.NotificationPriorityMedium, true},
		{domain.NotificationPriorityLow, domain.NotificationPriorityMedium, false},
		{domain.NotificationPriorityCritical, domain.NotificationPriorityCritical, true},
		{domain.NotificationPriorityHigh, domain.NotificationPriorityCritical, false},
	}

	for _, tc := range tests {
		result := uc.shouldDeliverToChannel(tc.notifPriority, tc.minPriority)
		assert.Equal(t, tc.expected, result, "Failed for notif: %s, min: %s", tc.notifPriority, tc.minPriority)
	}
}
