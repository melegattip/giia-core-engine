package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/melegattip/giia-core-engine/pkg/logger"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/api/dto"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/core/providers"
)

// NotificationHandler handles HTTP requests for notifications
type NotificationHandler struct {
	repo   providers.NotificationRepository
	logger logger.Logger
}

// NewNotificationHandler creates a new notification handler
func NewNotificationHandler(repo providers.NotificationRepository, logger logger.Logger) *NotificationHandler {
	return &NotificationHandler{
		repo:   repo,
		logger: logger,
	}
}

// RegisterRoutes registers all notification routes
func (h *NotificationHandler) RegisterRoutes(router *mux.Router) {
	// Notification routes
	router.HandleFunc("/notifications", h.ListNotifications).Methods(http.MethodGet)
	router.HandleFunc("/notifications/{id}", h.GetNotification).Methods(http.MethodGet)
	router.HandleFunc("/notifications/{id}/status", h.UpdateNotificationStatus).Methods(http.MethodPatch)
	router.HandleFunc("/notifications/{id}", h.DeleteNotification).Methods(http.MethodDelete)
}

// ListNotifications handles GET /notifications
// @Summary List notifications
// @Description Get a paginated list of notifications for the current user
// @Tags notifications
// @Accept json
// @Produce json
// @Param types query []string false "Filter by types"
// @Param priorities query []string false "Filter by priorities"
// @Param statuses query []string false "Filter by statuses"
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Page size (default: 20)"
// @Success 200 {object} dto.NotificationListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /notifications [get]
func (h *NotificationHandler) ListNotifications(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract user and organization from context (set by auth middleware)
	userID, err := h.getUserIDFromContext(r)
	if err != nil {
		h.respondError(w, http.StatusUnauthorized, "unauthorized", "User not authenticated")
		return
	}

	orgID, err := h.getOrganizationIDFromContext(r)
	if err != nil {
		h.respondError(w, http.StatusUnauthorized, "unauthorized", "Organization not found")
		return
	}

	// Parse query parameters
	query := r.URL.Query()
	page, _ := strconv.Atoi(query.Get("page"))
	if page < 1 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(query.Get("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// Build filters
	filters := &providers.NotificationFilters{
		Limit:  pageSize,
		Offset: (page - 1) * pageSize,
	}

	// Parse type filters
	if types := query["types"]; len(types) > 0 {
		filters.Types = make([]domain.NotificationType, len(types))
		for i, t := range types {
			filters.Types[i] = domain.NotificationType(t)
		}
	}

	// Parse priority filters
	if priorities := query["priorities"]; len(priorities) > 0 {
		filters.Priorities = make([]domain.NotificationPriority, len(priorities))
		for i, p := range priorities {
			filters.Priorities[i] = domain.NotificationPriority(p)
		}
	}

	// Parse status filters
	if statuses := query["statuses"]; len(statuses) > 0 {
		filters.Statuses = make([]domain.NotificationStatus, len(statuses))
		for i, s := range statuses {
			filters.Statuses[i] = domain.NotificationStatus(s)
		}
	}

	// Fetch notifications
	notifications, err := h.repo.List(ctx, userID, orgID, filters)
	if err != nil {
		h.logger.Error(ctx, err, "Failed to list notifications", logger.Tags{
			"user_id": userID.String(),
			"org_id":  orgID.String(),
		})
		h.respondError(w, http.StatusInternalServerError, "internal_error", "Failed to fetch notifications")
		return
	}

	// Calculate total pages (simplified - in production, get actual count from repo)
	totalPages := 1
	if len(notifications) == pageSize {
		totalPages = page + 1 // Estimate
	}

	// Convert to DTOs
	response := dto.NotificationListResponse{
		Notifications: dto.FromDomainList(notifications),
		Total:         len(notifications), // In production, get actual count
		Page:          page,
		PageSize:      pageSize,
		TotalPages:    totalPages,
	}

	h.respondJSON(w, http.StatusOK, response)
}

// GetNotification handles GET /notifications/{id}
// @Summary Get notification by ID
// @Description Get a single notification by ID
// @Tags notifications
// @Accept json
// @Produce json
// @Param id path string true "Notification ID"
// @Success 200 {object} dto.NotificationResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /notifications/{id} [get]
func (h *NotificationHandler) GetNotification(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)

	// Parse notification ID
	notifID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid_id", "Invalid notification ID format")
		return
	}

	// Get organization from context
	orgID, err := h.getOrganizationIDFromContext(r)
	if err != nil {
		h.respondError(w, http.StatusUnauthorized, "unauthorized", "Organization not found")
		return
	}

	// Fetch notification
	notification, err := h.repo.GetByID(ctx, notifID, orgID)
	if err != nil {
		h.logger.Error(ctx, err, "Failed to get notification", logger.Tags{
			"notification_id": notifID.String(),
			"org_id":          orgID.String(),
		})
		h.respondError(w, http.StatusNotFound, "not_found", "Notification not found")
		return
	}

	// Convert to DTO
	response := dto.FromDomain(notification)
	h.respondJSON(w, http.StatusOK, response)
}

// UpdateNotificationStatus handles PATCH /notifications/{id}/status
// @Summary Update notification status
// @Description Mark notification as read, acted upon, or dismissed
// @Tags notifications
// @Accept json
// @Produce json
// @Param id path string true "Notification ID"
// @Param request body dto.UpdateNotificationStatusRequest true "Status update request"
// @Success 200 {object} dto.NotificationResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /notifications/{id}/status [patch]
func (h *NotificationHandler) UpdateNotificationStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)

	// Parse notification ID
	notifID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid_id", "Invalid notification ID format")
		return
	}

	// Get organization from context
	orgID, err := h.getOrganizationIDFromContext(r)
	if err != nil {
		h.respondError(w, http.StatusUnauthorized, "unauthorized", "Organization not found")
		return
	}

	// Parse request body
	var req dto.UpdateNotificationStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	// Validate status
	status := domain.NotificationStatus(req.Status)
	if !isValidStatus(status) {
		h.respondError(w, http.StatusBadRequest, "invalid_status", "Invalid status value")
		return
	}

	// Fetch notification
	notification, err := h.repo.GetByID(ctx, notifID, orgID)
	if err != nil {
		h.logger.Error(ctx, err, "Failed to get notification", logger.Tags{
			"notification_id": notifID.String(),
		})
		h.respondError(w, http.StatusNotFound, "not_found", "Notification not found")
		return
	}

	// Update status based on request
	switch status {
	case domain.NotificationStatusRead:
		notification.MarkAsRead()
	case domain.NotificationStatusActedUpon:
		notification.MarkAsActedUpon()
	case domain.NotificationStatusDismissed:
		notification.Dismiss()
	}

	// Save updated notification
	if err := h.repo.Update(ctx, notification); err != nil {
		h.logger.Error(ctx, err, "Failed to update notification", logger.Tags{
			"notification_id": notifID.String(),
		})
		h.respondError(w, http.StatusInternalServerError, "internal_error", "Failed to update notification")
		return
	}

	h.logger.Info(ctx, "Notification status updated", logger.Tags{
		"notification_id": notifID.String(),
		"status":          req.Status,
	})

	// Return updated notification
	response := dto.FromDomain(notification)
	h.respondJSON(w, http.StatusOK, response)
}

// DeleteNotification handles DELETE /notifications/{id}
// @Summary Delete notification
// @Description Delete a notification by ID
// @Tags notifications
// @Accept json
// @Produce json
// @Param id path string true "Notification ID"
// @Success 204 "No Content"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /notifications/{id} [delete]
func (h *NotificationHandler) DeleteNotification(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)

	// Parse notification ID
	notifID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid_id", "Invalid notification ID format")
		return
	}

	// Get organization from context
	orgID, err := h.getOrganizationIDFromContext(r)
	if err != nil {
		h.respondError(w, http.StatusUnauthorized, "unauthorized", "Organization not found")
		return
	}

	// Delete notification
	if err := h.repo.Delete(ctx, notifID, orgID); err != nil {
		h.logger.Error(ctx, err, "Failed to delete notification", logger.Tags{
			"notification_id": notifID.String(),
		})
		h.respondError(w, http.StatusInternalServerError, "internal_error", "Failed to delete notification")
		return
	}

	h.logger.Info(ctx, "Notification deleted", logger.Tags{
		"notification_id": notifID.String(),
	})

	w.WriteHeader(http.StatusNoContent)
}

// Helper methods

func (h *NotificationHandler) getUserIDFromContext(r *http.Request) (uuid.UUID, error) {
	// In production, extract from JWT or session
	// For now, use a header for testing
	userIDStr := r.Header.Get("X-User-ID")
	if userIDStr == "" {
		return uuid.Nil, ErrUnauthorized
	}
	return uuid.Parse(userIDStr)
}

func (h *NotificationHandler) getOrganizationIDFromContext(r *http.Request) (uuid.UUID, error) {
	// In production, extract from JWT or session
	// For now, use a header for testing
	orgIDStr := r.Header.Get("X-Organization-ID")
	if orgIDStr == "" {
		return uuid.Nil, ErrUnauthorized
	}
	return uuid.Parse(orgIDStr)
}

func (h *NotificationHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *NotificationHandler) respondError(w http.ResponseWriter, status int, errorCode, message string) {
	response := dto.ErrorResponse{
		Error:   errorCode,
		Message: message,
	}
	h.respondJSON(w, status, response)
}

func isValidStatus(status domain.NotificationStatus) bool {
	switch status {
	case domain.NotificationStatusRead,
		domain.NotificationStatusActedUpon,
		domain.NotificationStatusDismissed:
		return true
	default:
		return false
	}
}

var ErrUnauthorized = fmt.Errorf("unauthorized")
