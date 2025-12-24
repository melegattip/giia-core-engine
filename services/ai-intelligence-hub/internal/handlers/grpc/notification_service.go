// Package grpc provides gRPC service implementations for the AI Intelligence Hub.
package grpc

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/melegattip/giia-core-engine/pkg/logger"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/core/providers"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/handlers/websocket"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// NotificationService implements the gRPC notification service.
type NotificationService struct {
	repo   providers.NotificationRepository
	wsHub  *websocket.Hub
	logger logger.Logger
}

// NewNotificationService creates a new gRPC notification service.
func NewNotificationService(
	repo providers.NotificationRepository,
	wsHub *websocket.Hub,
	logger logger.Logger,
) *NotificationService {
	return &NotificationService{
		repo:   repo,
		wsHub:  wsHub,
		logger: logger,
	}
}

// Notification represents a notification in gRPC messages.
type Notification struct {
	ID              string            `json:"id"`
	OrganizationID  string            `json:"organization_id"`
	UserID          string            `json:"user_id"`
	Type            string            `json:"type"`
	Priority        string            `json:"priority"`
	Title           string            `json:"title"`
	Summary         string            `json:"summary"`
	FullAnalysis    string            `json:"full_analysis"`
	Reasoning       string            `json:"reasoning"`
	Impact          *ImpactDetails    `json:"impact,omitempty"`
	Recommendations []*Recommendation `json:"recommendations,omitempty"`
	SourceEvents    []string          `json:"source_events"`
	Status          string            `json:"status"`
	CreatedAt       int64             `json:"created_at"`
	ReadAt          *int64            `json:"read_at,omitempty"`
	ActedAt         *int64            `json:"acted_at,omitempty"`
	DismissedAt     *int64            `json:"dismissed_at,omitempty"`
}

// ImpactDetails represents impact assessment in gRPC messages.
type ImpactDetails struct {
	RiskLevel           string  `json:"risk_level"`
	RevenueImpact       float64 `json:"revenue_impact"`
	CostImpact          float64 `json:"cost_impact"`
	TimeToImpactSeconds *int64  `json:"time_to_impact_seconds,omitempty"`
	AffectedOrders      int32   `json:"affected_orders"`
	AffectedProducts    int32   `json:"affected_products"`
}

// Recommendation represents a recommendation in gRPC messages.
type Recommendation struct {
	Action          string `json:"action"`
	Reasoning       string `json:"reasoning"`
	ExpectedOutcome string `json:"expected_outcome"`
	Effort          string `json:"effort"`
	Impact          string `json:"impact"`
	ActionURL       string `json:"action_url"`
	PriorityOrder   int32  `json:"priority_order"`
}

// GetNotificationRequest represents a request to get a notification.
type GetNotificationRequest struct {
	NotificationID string `json:"notification_id"`
	OrganizationID string `json:"organization_id"`
}

// GetNotificationResponse represents the response for getting a notification.
type GetNotificationResponse struct {
	Notification *Notification `json:"notification"`
}

// ListNotificationsRequest represents a request to list notifications.
type ListNotificationsRequest struct {
	UserID         string   `json:"user_id"`
	OrganizationID string   `json:"organization_id"`
	Types          []string `json:"types,omitempty"`
	Priorities     []string `json:"priorities,omitempty"`
	Statuses       []string `json:"statuses,omitempty"`
	PageSize       int32    `json:"page_size"`
	PageToken      string   `json:"page_token,omitempty"`
}

// ListNotificationsResponse represents the response for listing notifications.
type ListNotificationsResponse struct {
	Notifications []*Notification `json:"notifications"`
	NextPageToken string          `json:"next_page_token,omitempty"`
	TotalCount    int32           `json:"total_count"`
}

// CreateNotificationRequest represents a request to create a notification.
type CreateNotificationRequest struct {
	OrganizationID  string            `json:"organization_id"`
	UserID          string            `json:"user_id"`
	Type            string            `json:"type"`
	Priority        string            `json:"priority"`
	Title           string            `json:"title"`
	Summary         string            `json:"summary"`
	FullAnalysis    string            `json:"full_analysis,omitempty"`
	Reasoning       string            `json:"reasoning,omitempty"`
	Impact          *ImpactDetails    `json:"impact,omitempty"`
	Recommendations []*Recommendation `json:"recommendations,omitempty"`
	SourceEvents    []string          `json:"source_events,omitempty"`
}

// CreateNotificationResponse represents the response for creating a notification.
type CreateNotificationResponse struct {
	Notification *Notification `json:"notification"`
}

// UpdateNotificationStatusRequest represents a request to update notification status.
type UpdateNotificationStatusRequest struct {
	NotificationID string `json:"notification_id"`
	OrganizationID string `json:"organization_id"`
	Status         string `json:"status"`
}

// UpdateNotificationStatusResponse represents the response for updating notification status.
type UpdateNotificationStatusResponse struct {
	Notification *Notification `json:"notification"`
}

// DeleteNotificationRequest represents a request to delete a notification.
type DeleteNotificationRequest struct {
	NotificationID string `json:"notification_id"`
	OrganizationID string `json:"organization_id"`
}

// DeleteNotificationResponse represents the response for deleting a notification.
type DeleteNotificationResponse struct {
	Success bool `json:"success"`
}

// GetUnreadCountRequest represents a request to get unread notification count.
type GetUnreadCountRequest struct {
	UserID         string `json:"user_id"`
	OrganizationID string `json:"organization_id"`
}

// GetUnreadCountResponse represents the response for getting unread count.
type GetUnreadCountResponse struct {
	Count           int32            `json:"count"`
	CountByPriority map[string]int32 `json:"count_by_priority"`
}

// GetNotification retrieves a single notification by ID.
func (s *NotificationService) GetNotification(ctx context.Context, req *GetNotificationRequest) (*GetNotificationResponse, error) {
	if req.NotificationID == "" {
		return nil, status.Error(codes.InvalidArgument, "notification_id is required")
	}
	if req.OrganizationID == "" {
		return nil, status.Error(codes.InvalidArgument, "organization_id is required")
	}

	notifID, err := uuid.Parse(req.NotificationID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid notification_id format")
	}

	orgID, err := uuid.Parse(req.OrganizationID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid organization_id format")
	}

	notification, err := s.repo.GetByID(ctx, notifID, orgID)
	if err != nil {
		s.logger.Error(ctx, err, "Failed to get notification", logger.Tags{
			"notification_id": req.NotificationID,
		})
		return nil, status.Error(codes.NotFound, "notification not found")
	}

	return &GetNotificationResponse{
		Notification: s.toProtoNotification(notification),
	}, nil
}

// ListNotifications retrieves a paginated list of notifications.
func (s *NotificationService) ListNotifications(ctx context.Context, req *ListNotificationsRequest) (*ListNotificationsResponse, error) {
	if req.UserID == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if req.OrganizationID == "" {
		return nil, status.Error(codes.InvalidArgument, "organization_id is required")
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id format")
	}

	orgID, err := uuid.Parse(req.OrganizationID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid organization_id format")
	}

	pageSize := int(req.PageSize)
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	filters := &providers.NotificationFilters{
		Limit:  pageSize + 1, // Fetch one extra to determine if there are more results
		Offset: 0,
	}

	// Parse page token for offset
	if req.PageToken != "" {
		// In production, decode the page token
		// For simplicity, we'll skip pagination offset here
	}

	// Parse type filters
	if len(req.Types) > 0 {
		filters.Types = make([]domain.NotificationType, len(req.Types))
		for i, t := range req.Types {
			filters.Types[i] = domain.NotificationType(t)
		}
	}

	// Parse priority filters
	if len(req.Priorities) > 0 {
		filters.Priorities = make([]domain.NotificationPriority, len(req.Priorities))
		for i, p := range req.Priorities {
			filters.Priorities[i] = domain.NotificationPriority(p)
		}
	}

	// Parse status filters
	if len(req.Statuses) > 0 {
		filters.Statuses = make([]domain.NotificationStatus, len(req.Statuses))
		for i, st := range req.Statuses {
			filters.Statuses[i] = domain.NotificationStatus(st)
		}
	}

	notifications, err := s.repo.List(ctx, userID, orgID, filters)
	if err != nil {
		s.logger.Error(ctx, err, "Failed to list notifications", logger.Tags{
			"user_id": req.UserID,
		})
		return nil, status.Error(codes.Internal, "failed to list notifications")
	}

	hasMore := len(notifications) > pageSize
	if hasMore {
		notifications = notifications[:pageSize]
	}

	protoNotifications := make([]*Notification, len(notifications))
	for i, n := range notifications {
		protoNotifications[i] = s.toProtoNotification(n)
	}

	response := &ListNotificationsResponse{
		Notifications: protoNotifications,
		TotalCount:    int32(len(protoNotifications)),
	}

	if hasMore {
		// In production, encode the next page token
		response.NextPageToken = "next"
	}

	return response, nil
}

// CreateNotification creates a new notification and broadcasts it via WebSocket.
func (s *NotificationService) CreateNotification(ctx context.Context, req *CreateNotificationRequest) (*CreateNotificationResponse, error) {
	if req.OrganizationID == "" {
		return nil, status.Error(codes.InvalidArgument, "organization_id is required")
	}
	if req.UserID == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if req.Title == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}
	if req.Summary == "" {
		return nil, status.Error(codes.InvalidArgument, "summary is required")
	}

	orgID, err := uuid.Parse(req.OrganizationID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid organization_id format")
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id format")
	}

	// Create domain notification
	notification := domain.NewNotification(
		orgID,
		userID,
		domain.NotificationType(req.Type),
		domain.NotificationPriority(req.Priority),
		req.Title,
		req.Summary,
	)

	notification.FullAnalysis = req.FullAnalysis
	notification.Reasoning = req.Reasoning
	notification.SourceEvents = req.SourceEvents

	// Convert impact
	if req.Impact != nil {
		notification.Impact = domain.ImpactAssessment{
			RiskLevel:        req.Impact.RiskLevel,
			RevenueImpact:    req.Impact.RevenueImpact,
			CostImpact:       req.Impact.CostImpact,
			AffectedOrders:   int(req.Impact.AffectedOrders),
			AffectedProducts: int(req.Impact.AffectedProducts),
		}
		if req.Impact.TimeToImpactSeconds != nil {
			dur := time.Duration(*req.Impact.TimeToImpactSeconds) * time.Second
			notification.Impact.TimeToImpact = &dur
		}
	}

	// Convert recommendations
	if len(req.Recommendations) > 0 {
		notification.Recommendations = make([]domain.Recommendation, len(req.Recommendations))
		for i, r := range req.Recommendations {
			notification.Recommendations[i] = domain.Recommendation{
				Action:          r.Action,
				Reasoning:       r.Reasoning,
				ExpectedOutcome: r.ExpectedOutcome,
				Effort:          r.Effort,
				Impact:          r.Impact,
				ActionURL:       r.ActionURL,
				PriorityOrder:   int(r.PriorityOrder),
			}
		}
	}

	// Save notification
	if err := s.repo.Create(ctx, notification); err != nil {
		s.logger.Error(ctx, err, "Failed to create notification", logger.Tags{
			"user_id": req.UserID,
		})
		return nil, status.Error(codes.Internal, "failed to create notification")
	}

	s.logger.Info(ctx, "Notification created", logger.Tags{
		"notification_id": notification.ID.String(),
		"user_id":         req.UserID,
		"type":            req.Type,
	})

	// Broadcast via WebSocket
	if s.wsHub != nil {
		s.wsHub.BroadcastNotification(notification)
	}

	return &CreateNotificationResponse{
		Notification: s.toProtoNotification(notification),
	}, nil
}

// UpdateNotificationStatus updates the status of a notification.
func (s *NotificationService) UpdateNotificationStatus(ctx context.Context, req *UpdateNotificationStatusRequest) (*UpdateNotificationStatusResponse, error) {
	if req.NotificationID == "" {
		return nil, status.Error(codes.InvalidArgument, "notification_id is required")
	}
	if req.OrganizationID == "" {
		return nil, status.Error(codes.InvalidArgument, "organization_id is required")
	}
	if req.Status == "" {
		return nil, status.Error(codes.InvalidArgument, "status is required")
	}

	notifID, err := uuid.Parse(req.NotificationID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid notification_id format")
	}

	orgID, err := uuid.Parse(req.OrganizationID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid organization_id format")
	}

	// Validate status
	newStatus := domain.NotificationStatus(req.Status)
	if !isValidStatus(newStatus) {
		return nil, status.Error(codes.InvalidArgument, "invalid status value")
	}

	// Get notification
	notification, err := s.repo.GetByID(ctx, notifID, orgID)
	if err != nil {
		return nil, status.Error(codes.NotFound, "notification not found")
	}

	// Update status
	switch newStatus {
	case domain.NotificationStatusRead:
		notification.MarkAsRead()
	case domain.NotificationStatusActedUpon:
		notification.MarkAsActedUpon()
	case domain.NotificationStatusDismissed:
		notification.Dismiss()
	}

	// Save
	if err := s.repo.Update(ctx, notification); err != nil {
		s.logger.Error(ctx, err, "Failed to update notification", logger.Tags{
			"notification_id": req.NotificationID,
		})
		return nil, status.Error(codes.Internal, "failed to update notification")
	}

	s.logger.Info(ctx, "Notification status updated", logger.Tags{
		"notification_id": req.NotificationID,
		"status":          req.Status,
	})

	return &UpdateNotificationStatusResponse{
		Notification: s.toProtoNotification(notification),
	}, nil
}

// DeleteNotification deletes a notification.
func (s *NotificationService) DeleteNotification(ctx context.Context, req *DeleteNotificationRequest) (*DeleteNotificationResponse, error) {
	if req.NotificationID == "" {
		return nil, status.Error(codes.InvalidArgument, "notification_id is required")
	}
	if req.OrganizationID == "" {
		return nil, status.Error(codes.InvalidArgument, "organization_id is required")
	}

	notifID, err := uuid.Parse(req.NotificationID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid notification_id format")
	}

	orgID, err := uuid.Parse(req.OrganizationID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid organization_id format")
	}

	if err := s.repo.Delete(ctx, notifID, orgID); err != nil {
		s.logger.Error(ctx, err, "Failed to delete notification", logger.Tags{
			"notification_id": req.NotificationID,
		})
		return nil, status.Error(codes.Internal, "failed to delete notification")
	}

	s.logger.Info(ctx, "Notification deleted", logger.Tags{
		"notification_id": req.NotificationID,
	})

	return &DeleteNotificationResponse{
		Success: true,
	}, nil
}

// GetUnreadCount returns the count of unread notifications.
func (s *NotificationService) GetUnreadCount(ctx context.Context, req *GetUnreadCountRequest) (*GetUnreadCountResponse, error) {
	if req.UserID == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if req.OrganizationID == "" {
		return nil, status.Error(codes.InvalidArgument, "organization_id is required")
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id format")
	}

	orgID, err := uuid.Parse(req.OrganizationID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid organization_id format")
	}

	filters := &providers.NotificationFilters{
		Statuses: []domain.NotificationStatus{domain.NotificationStatusUnread},
		Limit:    1000,
	}

	notifications, err := s.repo.List(ctx, userID, orgID, filters)
	if err != nil {
		s.logger.Error(ctx, err, "Failed to get unread count", logger.Tags{
			"user_id": req.UserID,
		})
		return nil, status.Error(codes.Internal, "failed to get unread count")
	}

	// Count by priority
	countByPriority := make(map[string]int32)
	for _, n := range notifications {
		countByPriority[string(n.Priority)]++
	}

	return &GetUnreadCountResponse{
		Count:           int32(len(notifications)),
		CountByPriority: countByPriority,
	}, nil
}

// Helper methods

func (s *NotificationService) toProtoNotification(n *domain.AINotification) *Notification {
	if n == nil {
		return nil
	}

	proto := &Notification{
		ID:             n.ID.String(),
		OrganizationID: n.OrganizationID.String(),
		UserID:         n.UserID.String(),
		Type:           string(n.Type),
		Priority:       string(n.Priority),
		Title:          n.Title,
		Summary:        n.Summary,
		FullAnalysis:   n.FullAnalysis,
		Reasoning:      n.Reasoning,
		SourceEvents:   n.SourceEvents,
		Status:         string(n.Status),
		CreatedAt:      n.CreatedAt.Unix(),
	}

	if n.ReadAt != nil {
		ts := n.ReadAt.Unix()
		proto.ReadAt = &ts
	}
	if n.ActedAt != nil {
		ts := n.ActedAt.Unix()
		proto.ActedAt = &ts
	}
	if n.DismissedAt != nil {
		ts := n.DismissedAt.Unix()
		proto.DismissedAt = &ts
	}

	// Convert impact
	if n.Impact.RiskLevel != "" {
		proto.Impact = &ImpactDetails{
			RiskLevel:        n.Impact.RiskLevel,
			RevenueImpact:    n.Impact.RevenueImpact,
			CostImpact:       n.Impact.CostImpact,
			AffectedOrders:   int32(n.Impact.AffectedOrders),
			AffectedProducts: int32(n.Impact.AffectedProducts),
		}
		if n.Impact.TimeToImpact != nil {
			seconds := int64(n.Impact.TimeToImpact.Seconds())
			proto.Impact.TimeToImpactSeconds = &seconds
		}
	}

	// Convert recommendations
	if len(n.Recommendations) > 0 {
		proto.Recommendations = make([]*Recommendation, len(n.Recommendations))
		for i, r := range n.Recommendations {
			proto.Recommendations[i] = &Recommendation{
				Action:          r.Action,
				Reasoning:       r.Reasoning,
				ExpectedOutcome: r.ExpectedOutcome,
				Effort:          r.Effort,
				Impact:          r.Impact,
				ActionURL:       r.ActionURL,
				PriorityOrder:   int32(r.PriorityOrder),
			}
		}
	}

	return proto
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
