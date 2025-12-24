// Package http provides HTTP handlers for the AI Intelligence Hub API.
package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/melegattip/giia-core-engine/pkg/logger"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/core/providers"
)

// PreferencesHandler handles HTTP requests for user notification preferences.
type PreferencesHandler struct {
	repo   providers.PreferencesRepository
	logger logger.Logger
	cache  *NotificationCache
}

// NewPreferencesHandler creates a new preferences handler.
func NewPreferencesHandler(repo providers.PreferencesRepository, logger logger.Logger) *PreferencesHandler {
	return &PreferencesHandler{
		repo:   repo,
		logger: logger,
		cache:  NewNotificationCache(5 * time.Minute),
	}
}

// PreferencesResponse represents user notification preferences in API responses.
type PreferencesResponse struct {
	ID             uuid.UUID `json:"id"`
	UserID         uuid.UUID `json:"user_id"`
	OrganizationID uuid.UUID `json:"organization_id"`

	// Channel settings
	EnableInApp     bool   `json:"enable_in_app"`
	EnableEmail     bool   `json:"enable_email"`
	EnableSMS       bool   `json:"enable_sms"`
	EnableSlack     bool   `json:"enable_slack"`
	SlackWebhookURL string `json:"slack_webhook_url,omitempty"`
	EmailAddress    string `json:"email_address,omitempty"`
	PhoneNumber     string `json:"phone_number,omitempty"`

	// Priority thresholds
	InAppMinPriority string `json:"in_app_min_priority"`
	EmailMinPriority string `json:"email_min_priority"`
	SMSMinPriority   string `json:"sms_min_priority"`

	// Timing settings
	DigestTime      string `json:"digest_time"`
	QuietHoursStart string `json:"quiet_hours_start,omitempty"`
	QuietHoursEnd   string `json:"quiet_hours_end,omitempty"`
	Timezone        string `json:"timezone"`

	// Rate limits
	MaxAlertsPerHour int `json:"max_alerts_per_hour"`
	MaxEmailsPerDay  int `json:"max_emails_per_day"`

	// Display settings
	DetailLevel       string `json:"detail_level"`
	IncludeCharts     bool   `json:"include_charts"`
	IncludeHistorical bool   `json:"include_historical"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UpdatePreferencesRequest represents a request to update user preferences.
type UpdatePreferencesRequest struct {
	// Channel settings
	EnableInApp     *bool   `json:"enable_in_app,omitempty"`
	EnableEmail     *bool   `json:"enable_email,omitempty"`
	EnableSMS       *bool   `json:"enable_sms,omitempty"`
	EnableSlack     *bool   `json:"enable_slack,omitempty"`
	SlackWebhookURL *string `json:"slack_webhook_url,omitempty"`
	EmailAddress    *string `json:"email_address,omitempty"`
	PhoneNumber     *string `json:"phone_number,omitempty"`

	// Priority thresholds
	InAppMinPriority *string `json:"in_app_min_priority,omitempty"`
	EmailMinPriority *string `json:"email_min_priority,omitempty"`
	SMSMinPriority   *string `json:"sms_min_priority,omitempty"`

	// Timing settings
	DigestTime      *string `json:"digest_time,omitempty"`
	QuietHoursStart *string `json:"quiet_hours_start,omitempty"`
	QuietHoursEnd   *string `json:"quiet_hours_end,omitempty"`
	Timezone        *string `json:"timezone,omitempty"`

	// Rate limits
	MaxAlertsPerHour *int `json:"max_alerts_per_hour,omitempty"`
	MaxEmailsPerDay  *int `json:"max_emails_per_day,omitempty"`

	// Display settings
	DetailLevel       *string `json:"detail_level,omitempty"`
	IncludeCharts     *bool   `json:"include_charts,omitempty"`
	IncludeHistorical *bool   `json:"include_historical,omitempty"`
}

// GetPreferences handles GET /api/v1/notifications/preferences
func (h *PreferencesHandler) GetPreferences(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := h.getUserIDFromContext(r)
	if err != nil {
		h.respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	orgID, err := h.getOrganizationIDFromContext(r)
	if err != nil {
		h.respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Organization not found")
		return
	}

	// Check cache
	cacheKey := fmt.Sprintf("prefs:%s:%s", userID, orgID)
	if cached, found := h.cache.Get(cacheKey); found {
		h.respondJSON(w, http.StatusOK, cached)
		return
	}

	prefs, err := h.repo.GetByUserID(ctx, userID, orgID)
	if err != nil {
		// If not found, create default preferences
		h.logger.Info(ctx, "Creating default preferences for user", logger.Tags{
			"user_id": userID.String(),
			"org_id":  orgID.String(),
		})

		prefs = domain.NewUserPreferences(userID, orgID)
		if err := h.repo.Create(ctx, prefs); err != nil {
			h.logger.Error(ctx, err, "Failed to create default preferences", logger.Tags{
				"user_id": userID.String(),
			})
			h.respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create preferences")
			return
		}
	}

	response := h.toPreferencesResponse(prefs)
	h.cache.Set(cacheKey, response)
	h.respondJSON(w, http.StatusOK, response)
}

// UpdatePreferences handles PUT /api/v1/notifications/preferences
func (h *PreferencesHandler) UpdatePreferences(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := h.getUserIDFromContext(r)
	if err != nil {
		h.respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	orgID, err := h.getOrganizationIDFromContext(r)
	if err != nil {
		h.respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Organization not found")
		return
	}

	var req UpdatePreferencesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// Validate request
	if err := h.validatePreferencesRequest(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	// Get existing preferences or create new ones
	prefs, err := h.repo.GetByUserID(ctx, userID, orgID)
	if err != nil {
		prefs = domain.NewUserPreferences(userID, orgID)
	}

	// Apply updates
	h.applyPreferencesUpdate(prefs, &req)
	prefs.UpdatedAt = time.Now()

	// Save preferences
	if prefs.ID == uuid.Nil {
		prefs.ID = uuid.New()
		if err := h.repo.Create(ctx, prefs); err != nil {
			h.logger.Error(ctx, err, "Failed to create preferences", logger.Tags{
				"user_id": userID.String(),
			})
			h.respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create preferences")
			return
		}
	} else {
		if err := h.repo.Update(ctx, prefs); err != nil {
			h.logger.Error(ctx, err, "Failed to update preferences", logger.Tags{
				"user_id": userID.String(),
			})
			h.respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update preferences")
			return
		}
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("prefs:%s:%s", userID, orgID)
	h.cache.Delete(cacheKey)

	h.logger.Info(ctx, "Preferences updated", logger.Tags{
		"user_id": userID.String(),
		"org_id":  orgID.String(),
	})

	h.respondJSON(w, http.StatusOK, h.toPreferencesResponse(prefs))
}

// Helper methods

func (h *PreferencesHandler) getUserIDFromContext(r *http.Request) (uuid.UUID, error) {
	userIDStr := r.Header.Get("X-User-ID")
	if userIDStr == "" {
		userIDStr = r.URL.Query().Get("user_id")
	}
	if userIDStr == "" {
		return uuid.Nil, fmt.Errorf("user ID not provided")
	}
	return uuid.Parse(userIDStr)
}

func (h *PreferencesHandler) getOrganizationIDFromContext(r *http.Request) (uuid.UUID, error) {
	orgIDStr := r.Header.Get("X-Organization-ID")
	if orgIDStr == "" {
		orgIDStr = r.URL.Query().Get("organization_id")
	}
	if orgIDStr == "" {
		return uuid.Nil, fmt.Errorf("organization ID not provided")
	}
	return uuid.Parse(orgIDStr)
}

func (h *PreferencesHandler) validatePreferencesRequest(req *UpdatePreferencesRequest) error {
	// Validate priority values
	validPriorities := map[string]bool{
		"critical": true,
		"high":     true,
		"medium":   true,
		"low":      true,
	}

	if req.InAppMinPriority != nil && !validPriorities[*req.InAppMinPriority] {
		return fmt.Errorf("invalid in_app_min_priority: must be critical, high, medium, or low")
	}
	if req.EmailMinPriority != nil && !validPriorities[*req.EmailMinPriority] {
		return fmt.Errorf("invalid email_min_priority: must be critical, high, medium, or low")
	}
	if req.SMSMinPriority != nil && !validPriorities[*req.SMSMinPriority] {
		return fmt.Errorf("invalid sms_min_priority: must be critical, high, medium, or low")
	}

	// Validate detail level
	if req.DetailLevel != nil {
		validLevels := map[string]bool{"minimal": true, "summary": true, "detailed": true}
		if !validLevels[*req.DetailLevel] {
			return fmt.Errorf("invalid detail_level: must be minimal, summary, or detailed")
		}
	}

	// Validate rate limits
	if req.MaxAlertsPerHour != nil && (*req.MaxAlertsPerHour < 1 || *req.MaxAlertsPerHour > 100) {
		return fmt.Errorf("max_alerts_per_hour must be between 1 and 100")
	}
	if req.MaxEmailsPerDay != nil && (*req.MaxEmailsPerDay < 1 || *req.MaxEmailsPerDay > 500) {
		return fmt.Errorf("max_emails_per_day must be between 1 and 500")
	}

	// Validate time format (HH:MM)
	if req.DigestTime != nil {
		if _, err := time.Parse("15:04", *req.DigestTime); err != nil {
			return fmt.Errorf("invalid digest_time format: use HH:MM")
		}
	}
	if req.QuietHoursStart != nil {
		if _, err := time.Parse("15:04", *req.QuietHoursStart); err != nil {
			return fmt.Errorf("invalid quiet_hours_start format: use HH:MM")
		}
	}
	if req.QuietHoursEnd != nil {
		if _, err := time.Parse("15:04", *req.QuietHoursEnd); err != nil {
			return fmt.Errorf("invalid quiet_hours_end format: use HH:MM")
		}
	}

	return nil
}

func (h *PreferencesHandler) applyPreferencesUpdate(prefs *domain.UserNotificationPreferences, req *UpdatePreferencesRequest) {
	// Channel settings
	if req.EnableInApp != nil {
		prefs.EnableInApp = *req.EnableInApp
	}
	if req.EnableEmail != nil {
		prefs.EnableEmail = *req.EnableEmail
	}
	if req.EnableSMS != nil {
		prefs.EnableSMS = *req.EnableSMS
	}
	if req.EnableSlack != nil {
		prefs.EnableSlack = *req.EnableSlack
	}
	if req.SlackWebhookURL != nil {
		prefs.SlackWebhookURL = *req.SlackWebhookURL
	}
	if req.EmailAddress != nil {
		prefs.EmailAddress = *req.EmailAddress
	}
	if req.PhoneNumber != nil {
		prefs.PhoneNumber = *req.PhoneNumber
	}

	// Priority thresholds
	if req.InAppMinPriority != nil {
		prefs.InAppMinPriority = domain.NotificationPriority(*req.InAppMinPriority)
	}
	if req.EmailMinPriority != nil {
		prefs.EmailMinPriority = domain.NotificationPriority(*req.EmailMinPriority)
	}
	if req.SMSMinPriority != nil {
		prefs.SMSMinPriority = domain.NotificationPriority(*req.SMSMinPriority)
	}

	// Timing settings
	if req.DigestTime != nil {
		prefs.DigestTime = *req.DigestTime
	}
	if req.QuietHoursStart != nil {
		t, _ := time.Parse("15:04", *req.QuietHoursStart)
		prefs.QuietHoursStart = &t
	}
	if req.QuietHoursEnd != nil {
		t, _ := time.Parse("15:04", *req.QuietHoursEnd)
		prefs.QuietHoursEnd = &t
	}
	if req.Timezone != nil {
		prefs.Timezone = *req.Timezone
	}

	// Rate limits
	if req.MaxAlertsPerHour != nil {
		prefs.MaxAlertsPerHour = *req.MaxAlertsPerHour
	}
	if req.MaxEmailsPerDay != nil {
		prefs.MaxEmailsPerDay = *req.MaxEmailsPerDay
	}

	// Display settings
	if req.DetailLevel != nil {
		prefs.DetailLevel = *req.DetailLevel
	}
	if req.IncludeCharts != nil {
		prefs.IncludeCharts = *req.IncludeCharts
	}
	if req.IncludeHistorical != nil {
		prefs.IncludeHistorical = *req.IncludeHistorical
	}
}

func (h *PreferencesHandler) toPreferencesResponse(prefs *domain.UserNotificationPreferences) *PreferencesResponse {
	response := &PreferencesResponse{
		ID:                prefs.ID,
		UserID:            prefs.UserID,
		OrganizationID:    prefs.OrganizationID,
		EnableInApp:       prefs.EnableInApp,
		EnableEmail:       prefs.EnableEmail,
		EnableSMS:         prefs.EnableSMS,
		EnableSlack:       prefs.EnableSlack,
		SlackWebhookURL:   prefs.SlackWebhookURL,
		EmailAddress:      prefs.EmailAddress,
		PhoneNumber:       prefs.PhoneNumber,
		InAppMinPriority:  string(prefs.InAppMinPriority),
		EmailMinPriority:  string(prefs.EmailMinPriority),
		SMSMinPriority:    string(prefs.SMSMinPriority),
		DigestTime:        prefs.DigestTime,
		Timezone:          prefs.Timezone,
		MaxAlertsPerHour:  prefs.MaxAlertsPerHour,
		MaxEmailsPerDay:   prefs.MaxEmailsPerDay,
		DetailLevel:       prefs.DetailLevel,
		IncludeCharts:     prefs.IncludeCharts,
		IncludeHistorical: prefs.IncludeHistorical,
		CreatedAt:         prefs.CreatedAt,
		UpdatedAt:         prefs.UpdatedAt,
	}

	if prefs.QuietHoursStart != nil {
		response.QuietHoursStart = prefs.QuietHoursStart.Format("15:04")
	}
	if prefs.QuietHoursEnd != nil {
		response.QuietHoursEnd = prefs.QuietHoursEnd.Format("15:04")
	}

	return response
}

func (h *PreferencesHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func (h *PreferencesHandler) respondError(w http.ResponseWriter, status int, code, message string) {
	h.respondJSON(w, status, ErrorResponse{
		ErrorCode: code,
		Message:   message,
	})
}
