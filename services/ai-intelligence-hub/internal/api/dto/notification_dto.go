package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/core/domain"
)

// NotificationResponse represents a notification in API responses
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

// ImpactAssessmentDTO represents impact assessment in API
type ImpactAssessmentDTO struct {
	RiskLevel        string   `json:"risk_level"`
	RevenueImpact    float64  `json:"revenue_impact"`
	CostImpact       float64  `json:"cost_impact"`
	TimeToImpactDays *float64 `json:"time_to_impact_days,omitempty"`
	AffectedOrders   int      `json:"affected_orders"`
	AffectedProducts int      `json:"affected_products"`
}

// RecommendationDTO represents a recommendation in API
type RecommendationDTO struct {
	Action          string `json:"action"`
	Reasoning       string `json:"reasoning"`
	ExpectedOutcome string `json:"expected_outcome"`
	Effort          string `json:"effort"`
	Impact          string `json:"impact"`
	ActionURL       string `json:"action_url,omitempty"`
	PriorityOrder   int    `json:"priority_order"`
}

// NotificationListResponse represents a paginated list of notifications
type NotificationListResponse struct {
	Notifications []NotificationResponse `json:"notifications"`
	Total         int                    `json:"total"`
	Page          int                    `json:"page"`
	PageSize      int                    `json:"page_size"`
	TotalPages    int                    `json:"total_pages"`
}

// NotificationFiltersRequest represents filter parameters for listing notifications
type NotificationFiltersRequest struct {
	Types      []string `json:"types" form:"types"`
	Priorities []string `json:"priorities" form:"priorities"`
	Statuses   []string `json:"statuses" form:"statuses"`
	Page       int      `json:"page" form:"page"`
	PageSize   int      `json:"page_size" form:"page_size"`
}

// UpdateNotificationStatusRequest represents a request to update notification status
type UpdateNotificationStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=read acted_upon dismissed"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string            `json:"error"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

// FromDomain converts a domain notification to DTO
func FromDomain(notif *domain.AINotification) *NotificationResponse {
	if notif == nil {
		return nil
	}

	response := &NotificationResponse{
		ID:              notif.ID,
		OrganizationID:  notif.OrganizationID,
		UserID:          notif.UserID,
		Type:            string(notif.Type),
		Priority:        string(notif.Priority),
		Title:           notif.Title,
		Summary:         notif.Summary,
		FullAnalysis:    notif.FullAnalysis,
		Reasoning:       notif.Reasoning,
		SourceEvents:    notif.SourceEvents,
		RelatedEntities: notif.RelatedEntities,
		Status:          string(notif.Status),
		CreatedAt:       notif.CreatedAt,
		ReadAt:          notif.ReadAt,
		ActedAt:         notif.ActedAt,
		DismissedAt:     notif.DismissedAt,
	}

	// Convert impact assessment
	if notif.Impact.RiskLevel != "" {
		var timeToImpactDays *float64
		if notif.Impact.TimeToImpact != nil {
			days := notif.Impact.TimeToImpact.Hours() / 24
			timeToImpactDays = &days
		}

		response.Impact = &ImpactAssessmentDTO{
			RiskLevel:        notif.Impact.RiskLevel,
			RevenueImpact:    notif.Impact.RevenueImpact,
			CostImpact:       notif.Impact.CostImpact,
			TimeToImpactDays: timeToImpactDays,
			AffectedOrders:   notif.Impact.AffectedOrders,
			AffectedProducts: notif.Impact.AffectedProducts,
		}
	}

	// Convert recommendations
	if len(notif.Recommendations) > 0 {
		response.Recommendations = make([]RecommendationDTO, len(notif.Recommendations))
		for i, rec := range notif.Recommendations {
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

// FromDomainList converts a list of domain notifications to DTOs
func FromDomainList(notifications []*domain.AINotification) []NotificationResponse {
	result := make([]NotificationResponse, len(notifications))
	for i, notif := range notifications {
		if dto := FromDomain(notif); dto != nil {
			result[i] = *dto
		}
	}
	return result
}
