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
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/core/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockPreferencesRepository is a mock implementation of providers.PreferencesRepository
type MockPreferencesRepository struct {
	mock.Mock
}

func (m *MockPreferencesRepository) Create(ctx context.Context, prefs *domain.UserNotificationPreferences) error {
	args := m.Called(ctx, prefs)
	return args.Error(0)
}

func (m *MockPreferencesRepository) GetByUserID(ctx context.Context, userID uuid.UUID, organizationID uuid.UUID) (*domain.UserNotificationPreferences, error) {
	args := m.Called(ctx, userID, organizationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.UserNotificationPreferences), args.Error(1)
}

func (m *MockPreferencesRepository) Update(ctx context.Context, prefs *domain.UserNotificationPreferences) error {
	args := m.Called(ctx, prefs)
	return args.Error(0)
}

func createTestPreferences(userID, orgID uuid.UUID) *domain.UserNotificationPreferences {
	return &domain.UserNotificationPreferences{
		ID:                uuid.New(),
		UserID:            userID,
		OrganizationID:    orgID,
		EnableInApp:       true,
		EnableEmail:       true,
		EnableSMS:         false,
		EnableSlack:       false,
		InAppMinPriority:  domain.NotificationPriorityLow,
		EmailMinPriority:  domain.NotificationPriorityMedium,
		SMSMinPriority:    domain.NotificationPriorityCritical,
		DigestTime:        "06:00",
		Timezone:          "UTC",
		MaxAlertsPerHour:  10,
		MaxEmailsPerDay:   50,
		DetailLevel:       "detailed",
		IncludeCharts:     true,
		IncludeHistorical: true,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
}

func TestPreferencesHandler_GetPreferences(t *testing.T) {
	mockRepo := new(MockPreferencesRepository)
	mockLogger := new(MockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

	handler := NewPreferencesHandler(mockRepo, mockLogger)

	userID := uuid.New()
	orgID := uuid.New()
	prefs := createTestPreferences(userID, orgID)

	t.Run("successful get", func(t *testing.T) {
		mockRepo.On("GetByUserID", mock.Anything, userID, orgID).
			Return(prefs, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/v1/notifications/preferences", nil)
		req.Header.Set("X-User-ID", userID.String())
		req.Header.Set("X-Organization-ID", orgID.String())

		rec := httptest.NewRecorder()
		handler.GetPreferences(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var response PreferencesResponse
		err := json.NewDecoder(rec.Body).Decode(&response)
		require.NoError(t, err)

		assert.Equal(t, userID, response.UserID)
		assert.True(t, response.EnableInApp)
		assert.True(t, response.EnableEmail)
		assert.Equal(t, "detailed", response.DetailLevel)
	})

	t.Run("create default on not found", func(t *testing.T) {
		newUserID := uuid.New()
		mockRepo.On("GetByUserID", mock.Anything, newUserID, orgID).
			Return(nil, assert.AnError).Once()
		mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.UserNotificationPreferences")).
			Return(nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/v1/notifications/preferences", nil)
		req.Header.Set("X-User-ID", newUserID.String())
		req.Header.Set("X-Organization-ID", orgID.String())

		rec := httptest.NewRecorder()
		handler.GetPreferences(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("missing user ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/notifications/preferences", nil)
		req.Header.Set("X-Organization-ID", orgID.String())

		rec := httptest.NewRecorder()
		handler.GetPreferences(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("missing organization ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/notifications/preferences", nil)
		req.Header.Set("X-User-ID", userID.String())

		rec := httptest.NewRecorder()
		handler.GetPreferences(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}

func TestPreferencesHandler_UpdatePreferences(t *testing.T) {
	mockRepo := new(MockPreferencesRepository)
	mockLogger := new(MockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

	handler := NewPreferencesHandler(mockRepo, mockLogger)

	userID := uuid.New()
	orgID := uuid.New()
	prefs := createTestPreferences(userID, orgID)

	t.Run("successful update", func(t *testing.T) {
		mockRepo.On("GetByUserID", mock.Anything, userID, orgID).
			Return(prefs, nil).Once()
		mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.UserNotificationPreferences")).
			Return(nil).Once()

		enableEmail := false
		updateReq := UpdatePreferencesRequest{
			EnableEmail: &enableEmail,
		}
		body, _ := json.Marshal(updateReq)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/notifications/preferences", bytes.NewReader(body))
		req.Header.Set("X-User-ID", userID.String())
		req.Header.Set("X-Organization-ID", orgID.String())
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		handler.UpdatePreferences(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("update priority thresholds", func(t *testing.T) {
		mockRepo.On("GetByUserID", mock.Anything, userID, orgID).
			Return(prefs, nil).Once()
		mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.UserNotificationPreferences")).
			Return(nil).Once()

		emailPriority := "high"
		updateReq := UpdatePreferencesRequest{
			EmailMinPriority: &emailPriority,
		}
		body, _ := json.Marshal(updateReq)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/notifications/preferences", bytes.NewReader(body))
		req.Header.Set("X-User-ID", userID.String())
		req.Header.Set("X-Organization-ID", orgID.String())
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		handler.UpdatePreferences(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("update timing settings", func(t *testing.T) {
		mockRepo.On("GetByUserID", mock.Anything, userID, orgID).
			Return(prefs, nil).Once()
		mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.UserNotificationPreferences")).
			Return(nil).Once()

		digestTime := "08:00"
		quietStart := "22:00"
		quietEnd := "07:00"
		updateReq := UpdatePreferencesRequest{
			DigestTime:      &digestTime,
			QuietHoursStart: &quietStart,
			QuietHoursEnd:   &quietEnd,
		}
		body, _ := json.Marshal(updateReq)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/notifications/preferences", bytes.NewReader(body))
		req.Header.Set("X-User-ID", userID.String())
		req.Header.Set("X-Organization-ID", orgID.String())
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		handler.UpdatePreferences(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("invalid priority value", func(t *testing.T) {
		invalidPriority := "invalid"
		updateReq := UpdatePreferencesRequest{
			EmailMinPriority: &invalidPriority,
		}
		body, _ := json.Marshal(updateReq)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/notifications/preferences", bytes.NewReader(body))
		req.Header.Set("X-User-ID", userID.String())
		req.Header.Set("X-Organization-ID", orgID.String())
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		handler.UpdatePreferences(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("invalid detail level", func(t *testing.T) {
		invalidLevel := "invalid"
		updateReq := UpdatePreferencesRequest{
			DetailLevel: &invalidLevel,
		}
		body, _ := json.Marshal(updateReq)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/notifications/preferences", bytes.NewReader(body))
		req.Header.Set("X-User-ID", userID.String())
		req.Header.Set("X-Organization-ID", orgID.String())
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		handler.UpdatePreferences(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("invalid rate limits", func(t *testing.T) {
		invalidRate := 200 // Max is 100
		updateReq := UpdatePreferencesRequest{
			MaxAlertsPerHour: &invalidRate,
		}
		body, _ := json.Marshal(updateReq)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/notifications/preferences", bytes.NewReader(body))
		req.Header.Set("X-User-ID", userID.String())
		req.Header.Set("X-Organization-ID", orgID.String())
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		handler.UpdatePreferences(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("invalid time format", func(t *testing.T) {
		invalidTime := "invalid-time"
		updateReq := UpdatePreferencesRequest{
			DigestTime: &invalidTime,
		}
		body, _ := json.Marshal(updateReq)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/notifications/preferences", bytes.NewReader(body))
		req.Header.Set("X-User-ID", userID.String())
		req.Header.Set("X-Organization-ID", orgID.String())
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		handler.UpdatePreferences(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("invalid request body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/api/v1/notifications/preferences", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("X-User-ID", userID.String())
		req.Header.Set("X-Organization-ID", orgID.String())
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		handler.UpdatePreferences(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("create new if not found", func(t *testing.T) {
		newUserID := uuid.New()
		mockRepo.On("GetByUserID", mock.Anything, newUserID, orgID).
			Return(nil, assert.AnError).Once()
		// NewUserPreferences sets an ID, so the handler calls Update instead of Create
		mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.UserNotificationPreferences")).
			Return(nil).Once()

		enableEmail := true
		updateReq := UpdatePreferencesRequest{
			EnableEmail: &enableEmail,
		}
		body, _ := json.Marshal(updateReq)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/notifications/preferences", bytes.NewReader(body))
		req.Header.Set("X-User-ID", newUserID.String())
		req.Header.Set("X-Organization-ID", orgID.String())
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		handler.UpdatePreferences(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

func TestPreferencesHandler_Validation(t *testing.T) {
	mockRepo := new(MockPreferencesRepository)
	mockLogger := new(MockLogger)
	handler := NewPreferencesHandler(mockRepo, mockLogger)

	t.Run("validate valid priorities", func(t *testing.T) {
		validPriorities := []string{"critical", "high", "medium", "low"}
		for _, priority := range validPriorities {
			p := priority
			req := &UpdatePreferencesRequest{
				EmailMinPriority: &p,
			}
			err := handler.validatePreferencesRequest(req)
			assert.NoError(t, err, "Priority %s should be valid", priority)
		}
	})

	t.Run("validate valid detail levels", func(t *testing.T) {
		validLevels := []string{"minimal", "summary", "detailed"}
		for _, level := range validLevels {
			l := level
			req := &UpdatePreferencesRequest{
				DetailLevel: &l,
			}
			err := handler.validatePreferencesRequest(req)
			assert.NoError(t, err, "Detail level %s should be valid", level)
		}
	})

	t.Run("validate rate limit boundaries", func(t *testing.T) {
		// Valid rate limits
		validAlerts := 50
		validEmails := 100
		req := &UpdatePreferencesRequest{
			MaxAlertsPerHour: &validAlerts,
			MaxEmailsPerDay:  &validEmails,
		}
		err := handler.validatePreferencesRequest(req)
		assert.NoError(t, err)

		// Invalid - too low
		invalidLow := 0
		req = &UpdatePreferencesRequest{
			MaxAlertsPerHour: &invalidLow,
		}
		err = handler.validatePreferencesRequest(req)
		assert.Error(t, err)

		// Invalid - too high for emails
		invalidHigh := 600
		req = &UpdatePreferencesRequest{
			MaxEmailsPerDay: &invalidHigh,
		}
		err = handler.validatePreferencesRequest(req)
		assert.Error(t, err)
	})

	t.Run("validate time formats", func(t *testing.T) {
		// Valid time format
		validTime := "14:30"
		req := &UpdatePreferencesRequest{
			DigestTime: &validTime,
		}
		err := handler.validatePreferencesRequest(req)
		assert.NoError(t, err)

		// Invalid time format
		invalidTime := "2:30 PM"
		req = &UpdatePreferencesRequest{
			DigestTime: &invalidTime,
		}
		err = handler.validatePreferencesRequest(req)
		assert.Error(t, err)
	})
}
