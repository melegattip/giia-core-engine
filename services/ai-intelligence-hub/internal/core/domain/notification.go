package domain

import (
	"time"

	"github.com/google/uuid"
)

type NotificationType string

const (
	NotificationTypeAlert      NotificationType = "alert"
	NotificationTypeWarning    NotificationType = "warning"
	NotificationTypeInfo       NotificationType = "info"
	NotificationTypeSuggestion NotificationType = "suggestion"
	NotificationTypeInsight    NotificationType = "insight"
	NotificationTypeDigest     NotificationType = "digest"
)

type NotificationPriority string

const (
	NotificationPriorityCritical NotificationPriority = "critical"
	NotificationPriorityHigh     NotificationPriority = "high"
	NotificationPriorityMedium   NotificationPriority = "medium"
	NotificationPriorityLow      NotificationPriority = "low"
)

type NotificationStatus string

const (
	NotificationStatusUnread     NotificationStatus = "unread"
	NotificationStatusRead       NotificationStatus = "read"
	NotificationStatusActedUpon  NotificationStatus = "acted_upon"
	NotificationStatusDismissed  NotificationStatus = "dismissed"
)

type AINotification struct {
	ID              uuid.UUID
	OrganizationID  uuid.UUID
	UserID          uuid.UUID
	Type            NotificationType
	Priority        NotificationPriority
	Title           string
	Summary         string
	FullAnalysis    string
	Reasoning       string
	Impact          ImpactAssessment
	Recommendations []Recommendation
	SourceEvents    []string
	RelatedEntities map[string][]string
	Status          NotificationStatus
	CreatedAt       time.Time
	ReadAt          *time.Time
	ActedAt         *time.Time
	DismissedAt     *time.Time
}

type ImpactAssessment struct {
	RiskLevel        string
	RevenueImpact    float64
	CostImpact       float64
	TimeToImpact     *time.Duration
	AffectedOrders   int
	AffectedProducts int
}

type Recommendation struct {
	Action          string
	Reasoning       string
	ExpectedOutcome string
	Effort          string
	Impact          string
	ActionURL       string
	PriorityOrder   int
}

func (n *AINotification) MarkAsRead() {
	now := time.Now()
	n.ReadAt = &now
	n.Status = NotificationStatusRead
}

func (n *AINotification) MarkAsActedUpon() {
	now := time.Now()
	n.ActedAt = &now
	n.Status = NotificationStatusActedUpon
}

func (n *AINotification) Dismiss() {
	now := time.Now()
	n.DismissedAt = &now
	n.Status = NotificationStatusDismissed
}

func NewNotification(
	organizationID uuid.UUID,
	userID uuid.UUID,
	notifType NotificationType,
	priority NotificationPriority,
	title string,
	summary string,
) *AINotification {
	return &AINotification{
		ID:              uuid.New(),
		OrganizationID:  organizationID,
		UserID:          userID,
		Type:            notifType,
		Priority:        priority,
		Title:           title,
		Summary:         summary,
		Status:          NotificationStatusUnread,
		CreatedAt:       time.Now(),
		SourceEvents:    []string{},
		RelatedEntities: make(map[string][]string),
		Recommendations: []Recommendation{},
	}
}
