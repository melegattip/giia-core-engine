// Package http provides HTTP handlers for the AI Intelligence Hub API.
package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/melegattip/giia-core-engine/pkg/logger"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/core/providers"
)

// NotificationHandler handles HTTP requests for notification operations.
type NotificationHandler struct {
	repo   providers.NotificationRepository
	logger logger.Logger
	cache  *NotificationCache
}

// NotificationCache provides simple in-memory caching for notifications.
type NotificationCache struct {
	items    map[string]cacheItem
	ttl      time.Duration
	maxItems int
}

type cacheItem struct {
	data      interface{}
	expiresAt time.Time
}

// NewNotificationCache creates a new notification cache.
func NewNotificationCache(ttl time.Duration) *NotificationCache {
	return &NotificationCache{
		items:    make(map[string]cacheItem),
		ttl:      ttl,
		maxItems: 1000,
	}
}

// Get retrieves an item from the cache.
func (c *NotificationCache) Get(key string) (interface{}, bool) {
	item, ok := c.items[key]
	if !ok || time.Now().After(item.expiresAt) {
		delete(c.items, key)
		return nil, false
	}
	return item.data, true
}

// Set stores an item in the cache.
func (c *NotificationCache) Set(key string, value interface{}) {
	if len(c.items) >= c.maxItems {
		c.evictOldest()
	}
	c.items[key] = cacheItem{
		data:      value,
		expiresAt: time.Now().Add(c.ttl),
	}
}

// Delete removes an item from the cache.
func (c *NotificationCache) Delete(key string) {
	delete(c.items, key)
}

// Clear removes all items from the cache.
func (c *NotificationCache) Clear() {
	c.items = make(map[string]cacheItem)
}

func (c *NotificationCache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, item := range c.items {
		if oldestKey == "" || item.expiresAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = item.expiresAt
		}
	}

	if oldestKey != "" {
		delete(c.items, oldestKey)
	}
}

// NewNotificationHandler creates a new notification handler.
func NewNotificationHandler(repo providers.NotificationRepository, logger logger.Logger) *NotificationHandler {
	return &NotificationHandler{
		repo:   repo,
		logger: logger,
		cache:  NewNotificationCache(5 * time.Minute),
	}
}

// NotificationListResponse represents a paginated list of notifications.
type NotificationListResponse struct {
	Notifications []NotificationResponse `json:"notifications"`
	Total         int                    `json:"total"`
	TotalCount    int                    `json:"total_count"`
	UnreadCount   int                    `json:"unread_count"`
	Page          int                    `json:"page"`
	PageSize      int                    `json:"page_size"`
	TotalPages    int                    `json:"total_pages"`
	HasNext       bool                   `json:"has_next"`
	HasPrevious   bool                   `json:"has_previous"`
}

// NotificationResponse represents a notification in API responses.
type NotificationResponse struct {
	ID              uuid.UUID            `json:"id"`
	OrganizationID  uuid.UUID            `json:"organization_id"`
	UserID          uuid.UUID            `json:"user_id"`
	Type            string               `json:"type"`
	Priority        string               `json:"priority"`
	Title           string               `json:"title"`
	Summary         string               `json:"summary"`
	FullAnalysis    string               `json:"full_analysis,omitempty"`
	Reasoning       string               `json:"reasoning,omitempty"`
	Impact          *ImpactAssessmentDTO `json:"impact,omitempty"`
	Recommendations []RecommendationDTO  `json:"recommendations,omitempty"`
	SourceEvents    []string             `json:"source_events,omitempty"`
	RelatedEntities map[string][]string  `json:"related_entities,omitempty"`
	Status          string               `json:"status"`
	CreatedAt       time.Time            `json:"created_at"`
	ReadAt          *time.Time           `json:"read_at,omitempty"`
	ActedAt         *time.Time           `json:"acted_at,omitempty"`
	DismissedAt     *time.Time           `json:"dismissed_at,omitempty"`
}

// ImpactAssessmentDTO represents impact assessment in API.
type ImpactAssessmentDTO struct {
	RiskLevel        string   `json:"risk_level"`
	RevenueImpact    float64  `json:"revenue_impact"`
	CostImpact       float64  `json:"cost_impact"`
	TimeToImpactDays *float64 `json:"time_to_impact_days,omitempty"`
	AffectedOrders   int      `json:"affected_orders"`
	AffectedProducts int      `json:"affected_products"`
}

// RecommendationDTO represents a recommendation in API.
type RecommendationDTO struct {
	Action          string `json:"action"`
	Reasoning       string `json:"reasoning"`
	ExpectedOutcome string `json:"expected_outcome"`
	Effort          string `json:"effort"`
	Impact          string `json:"impact"`
	ActionURL       string `json:"action_url,omitempty"`
	PriorityOrder   int    `json:"priority_order"`
}

// UpdateNotificationRequest represents a request to update notification status.
type UpdateNotificationRequest struct {
	Status string `json:"status"`
}

// UnreadCountResponse represents the unread notification count.
type UnreadCountResponse struct {
	UnreadCount     int            `json:"unread_count"`
	CountByPriority map[string]int `json:"count_by_priority"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	ErrorCode string            `json:"error_code"`
	Message   string            `json:"message"`
	Details   map[string]string `json:"details,omitempty"`
}

// ListNotifications handles GET /api/v1/notifications
func (h *NotificationHandler) ListNotifications(w http.ResponseWriter, r *http.Request) {
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

	// Parse query parameters
	query := r.URL.Query()
	page := h.parseIntParam(query.Get("page"), 1)
	pageSize := h.parseIntParam(query.Get("page_size"), 20)
	if pageSize > 100 {
		pageSize = 100
	}

	// Build filters
	filters := &providers.NotificationFilters{
		Limit:  pageSize,
		Offset: (page - 1) * pageSize,
	}

	// Parse type filters
	if types := query["type"]; len(types) > 0 {
		filters.Types = make([]domain.NotificationType, len(types))
		for i, t := range types {
			filters.Types[i] = domain.NotificationType(t)
		}
	}

	// Parse priority filters
	if priorities := query["priority"]; len(priorities) > 0 {
		filters.Priorities = make([]domain.NotificationPriority, len(priorities))
		for i, p := range priorities {
			filters.Priorities[i] = domain.NotificationPriority(p)
		}
	}

	// Parse status filters
	if statuses := query["status"]; len(statuses) > 0 {
		filters.Statuses = make([]domain.NotificationStatus, len(statuses))
		for i, s := range statuses {
			filters.Statuses[i] = domain.NotificationStatus(s)
		}
	}

	// Check cache
	cacheKey := fmt.Sprintf("list:%s:%s:%d:%d", userID, orgID, page, pageSize)
	if cached, found := h.cache.Get(cacheKey); found {
		h.respondJSON(w, http.StatusOK, cached)
		return
	}

	// Fetch notifications
	notifications, err := h.repo.List(ctx, userID, orgID, filters)
	if err != nil {
		h.logger.Error(ctx, err, "Failed to list notifications", logger.Tags{
			"user_id": userID.String(),
			"org_id":  orgID.String(),
		})
		h.respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch notifications")
		return
	}

	// Count unread
	unreadCount := 0
	for _, n := range notifications {
		if n.Status == domain.NotificationStatusUnread {
			unreadCount++
		}
	}

	// Calculate pagination
	total := len(notifications)
	totalPages := (total + pageSize - 1) / pageSize
	if totalPages < 1 {
		totalPages = 1
	}

	response := NotificationListResponse{
		Notifications: h.toNotificationResponseList(notifications),
		Total:         total,
		TotalCount:    total,
		UnreadCount:   unreadCount,
		Page:          page,
		PageSize:      pageSize,
		TotalPages:    totalPages,
		HasNext:       page < totalPages,
		HasPrevious:   page > 1,
	}

	h.cache.Set(cacheKey, response)
	h.respondJSON(w, http.StatusOK, response)
}

// GetNotification handles GET /api/v1/notifications/{id}
func (h *NotificationHandler) GetNotification(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	notifID, err := h.parseNotificationID(r)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "INVALID_ID", "Invalid notification ID format")
		return
	}

	orgID, err := h.getOrganizationIDFromContext(r)
	if err != nil {
		h.respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Organization not found")
		return
	}

	// Check cache
	cacheKey := fmt.Sprintf("notif:%s:%s", notifID, orgID)
	if cached, found := h.cache.Get(cacheKey); found {
		h.respondJSON(w, http.StatusOK, cached)
		return
	}

	notification, err := h.repo.GetByID(ctx, notifID, orgID)
	if err != nil {
		h.logger.Error(ctx, err, "Failed to get notification", logger.Tags{
			"notification_id": notifID.String(),
			"org_id":          orgID.String(),
		})
		h.respondError(w, http.StatusNotFound, "NOT_FOUND", "Notification not found")
		return
	}

	response := h.toNotificationResponse(notification)
	h.cache.Set(cacheKey, response)
	h.respondJSON(w, http.StatusOK, response)
}

// UpdateNotification handles PATCH /api/v1/notifications/{id}
func (h *NotificationHandler) UpdateNotification(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	notifID, err := h.parseNotificationID(r)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "INVALID_ID", "Invalid notification ID format")
		return
	}

	orgID, err := h.getOrganizationIDFromContext(r)
	if err != nil {
		h.respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Organization not found")
		return
	}

	var req UpdateNotificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// Validate status
	status := domain.NotificationStatus(req.Status)
	if !h.isValidStatus(status) {
		h.respondError(w, http.StatusBadRequest, "INVALID_STATUS", "Invalid status value. Must be: read, acted_upon, or dismissed")
		return
	}

	notification, err := h.repo.GetByID(ctx, notifID, orgID)
	if err != nil {
		h.respondError(w, http.StatusNotFound, "NOT_FOUND", "Notification not found")
		return
	}

	// Update status
	switch status {
	case domain.NotificationStatusRead:
		notification.MarkAsRead()
	case domain.NotificationStatusActedUpon:
		notification.MarkAsActedUpon()
	case domain.NotificationStatusDismissed:
		notification.Dismiss()
	}

	if err := h.repo.Update(ctx, notification); err != nil {
		h.logger.Error(ctx, err, "Failed to update notification", logger.Tags{
			"notification_id": notifID.String(),
		})
		h.respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update notification")
		return
	}

	// Invalidate cache
	h.cache.Delete(fmt.Sprintf("notif:%s:%s", notifID, orgID))

	h.logger.Info(ctx, "Notification status updated", logger.Tags{
		"notification_id": notifID.String(),
		"status":          req.Status,
	})

	h.respondJSON(w, http.StatusOK, h.toNotificationResponse(notification))
}

// DeleteNotification handles DELETE /api/v1/notifications/{id}
func (h *NotificationHandler) DeleteNotification(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	notifID, err := h.parseNotificationID(r)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "INVALID_ID", "Invalid notification ID format")
		return
	}

	orgID, err := h.getOrganizationIDFromContext(r)
	if err != nil {
		h.respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Organization not found")
		return
	}

	if err := h.repo.Delete(ctx, notifID, orgID); err != nil {
		h.logger.Error(ctx, err, "Failed to delete notification", logger.Tags{
			"notification_id": notifID.String(),
		})
		h.respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete notification")
		return
	}

	// Invalidate cache
	h.cache.Delete(fmt.Sprintf("notif:%s:%s", notifID, orgID))

	h.logger.Info(ctx, "Notification deleted", logger.Tags{
		"notification_id": notifID.String(),
	})

	w.WriteHeader(http.StatusNoContent)
}

// GetUnreadCount handles GET /api/v1/notifications/unread-count
func (h *NotificationHandler) GetUnreadCount(w http.ResponseWriter, r *http.Request) {
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
	cacheKey := fmt.Sprintf("unread:%s:%s", userID, orgID)
	if cached, found := h.cache.Get(cacheKey); found {
		h.respondJSON(w, http.StatusOK, cached)
		return
	}

	// Fetch unread notifications
	filters := &providers.NotificationFilters{
		Statuses: []domain.NotificationStatus{domain.NotificationStatusUnread},
		Limit:    1000,
	}

	notifications, err := h.repo.List(ctx, userID, orgID, filters)
	if err != nil {
		h.logger.Error(ctx, err, "Failed to get unread count", logger.Tags{
			"user_id": userID.String(),
		})
		h.respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get unread count")
		return
	}

	// Count by priority
	countByPriority := make(map[string]int)
	for _, n := range notifications {
		countByPriority[string(n.Priority)]++
	}

	response := UnreadCountResponse{
		UnreadCount:     len(notifications),
		CountByPriority: countByPriority,
	}

	h.cache.Set(cacheKey, response)
	h.respondJSON(w, http.StatusOK, response)
}

// Helper methods

func (h *NotificationHandler) parseNotificationID(r *http.Request) (uuid.UUID, error) {
	// Extract from URL path - expecting format: /api/v1/notifications/{id}
	path := r.URL.Path
	parts := splitPath(path)
	if len(parts) >= 4 && parts[2] == "notifications" {
		return uuid.Parse(parts[3])
	}
	return uuid.Nil, fmt.Errorf("notification ID not found in path")
}

func splitPath(path string) []string {
	var parts []string
	start := 0
	for i := 0; i <= len(path); i++ {
		if i == len(path) || path[i] == '/' {
			if i > start {
				parts = append(parts, path[start:i])
			}
			start = i + 1
		}
	}
	return parts
}

func (h *NotificationHandler) getUserIDFromContext(r *http.Request) (uuid.UUID, error) {
	userIDStr := r.Header.Get("X-User-ID")
	if userIDStr == "" {
		userIDStr = r.URL.Query().Get("user_id")
	}
	if userIDStr == "" {
		return uuid.Nil, fmt.Errorf("user ID not provided")
	}
	return uuid.Parse(userIDStr)
}

func (h *NotificationHandler) getOrganizationIDFromContext(r *http.Request) (uuid.UUID, error) {
	orgIDStr := r.Header.Get("X-Organization-ID")
	if orgIDStr == "" {
		orgIDStr = r.URL.Query().Get("organization_id")
	}
	if orgIDStr == "" {
		return uuid.Nil, fmt.Errorf("organization ID not provided")
	}
	return uuid.Parse(orgIDStr)
}

func (h *NotificationHandler) parseIntParam(value string, defaultVal int) int {
	if value == "" {
		return defaultVal
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 1 {
		return defaultVal
	}
	return parsed
}

func (h *NotificationHandler) isValidStatus(status domain.NotificationStatus) bool {
	switch status {
	case domain.NotificationStatusRead,
		domain.NotificationStatusActedUpon,
		domain.NotificationStatusDismissed:
		return true
	default:
		return false
	}
}

func (h *NotificationHandler) toNotificationResponse(n *domain.AINotification) *NotificationResponse {
	if n == nil {
		return nil
	}

	response := &NotificationResponse{
		ID:              n.ID,
		OrganizationID:  n.OrganizationID,
		UserID:          n.UserID,
		Type:            string(n.Type),
		Priority:        string(n.Priority),
		Title:           n.Title,
		Summary:         n.Summary,
		FullAnalysis:    n.FullAnalysis,
		Reasoning:       n.Reasoning,
		SourceEvents:    n.SourceEvents,
		RelatedEntities: n.RelatedEntities,
		Status:          string(n.Status),
		CreatedAt:       n.CreatedAt,
		ReadAt:          n.ReadAt,
		ActedAt:         n.ActedAt,
		DismissedAt:     n.DismissedAt,
	}

	// Convert impact assessment
	if n.Impact.RiskLevel != "" {
		var timeToImpactDays *float64
		if n.Impact.TimeToImpact != nil {
			days := n.Impact.TimeToImpact.Hours() / 24
			timeToImpactDays = &days
		}

		response.Impact = &ImpactAssessmentDTO{
			RiskLevel:        n.Impact.RiskLevel,
			RevenueImpact:    n.Impact.RevenueImpact,
			CostImpact:       n.Impact.CostImpact,
			TimeToImpactDays: timeToImpactDays,
			AffectedOrders:   n.Impact.AffectedOrders,
			AffectedProducts: n.Impact.AffectedProducts,
		}
	}

	// Convert recommendations
	if len(n.Recommendations) > 0 {
		response.Recommendations = make([]RecommendationDTO, len(n.Recommendations))
		for i, rec := range n.Recommendations {
			response.Recommendations[i] = RecommendationDTO{
				Action:          rec.Action,
				Reasoning:       rec.Reasoning,
				ExpectedOutcome: rec.ExpectedOutcome,
				Effort:          rec.Effort,
				Impact:          rec.Impact,
				ActionURL:       rec.ActionURL,
				PriorityOrder:   rec.PriorityOrder,
			}
		}
	}

	return response
}

func (h *NotificationHandler) toNotificationResponseList(notifications []*domain.AINotification) []NotificationResponse {
	result := make([]NotificationResponse, 0, len(notifications))
	for _, n := range notifications {
		if resp := h.toNotificationResponse(n); resp != nil {
			result = append(result, *resp)
		}
	}
	return result
}

func (h *NotificationHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func (h *NotificationHandler) respondError(w http.ResponseWriter, status int, code, message string) {
	h.respondJSON(w, status, ErrorResponse{
		ErrorCode: code,
		Message:   message,
	})
}
