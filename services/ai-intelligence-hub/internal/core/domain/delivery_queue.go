package domain

import (
	"time"

	"github.com/google/uuid"
)

// DeliveryStatus represents the status of a delivery attempt
type DeliveryStatus string

const (
	DeliveryStatusPending   DeliveryStatus = "pending"
	DeliveryStatusInFlight  DeliveryStatus = "in_flight"
	DeliveryStatusDelivered DeliveryStatus = "delivered"
	DeliveryStatusFailed    DeliveryStatus = "failed"
	DeliveryStatusRetrying  DeliveryStatus = "retrying"
)

// DeliveryChannel represents a notification delivery channel
type DeliveryChannel string

const (
	DeliveryChannelEmail   DeliveryChannel = "email"
	DeliveryChannelSlack   DeliveryChannel = "slack"
	DeliveryChannelSMS     DeliveryChannel = "sms"
	DeliveryChannelWebhook DeliveryChannel = "webhook"
	DeliveryChannelInApp   DeliveryChannel = "in_app"
)

// DeliveryQueueItem represents an item in the delivery queue
type DeliveryQueueItem struct {
	ID             uuid.UUID
	NotificationID uuid.UUID
	OrganizationID uuid.UUID
	UserID         uuid.UUID
	Channel        DeliveryChannel
	Recipient      string // Email address, phone number, webhook URL, etc.
	Status         DeliveryStatus
	RetryCount     int
	MaxRetries     int
	NextRetryAt    *time.Time
	LastError      string
	MessageID      string // External message ID from provider
	Metadata       map[string]string
	Priority       int // Higher priority = process first
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeliveredAt    *time.Time
}

// NewDeliveryQueueItem creates a new delivery queue item
func NewDeliveryQueueItem(
	notificationID uuid.UUID,
	organizationID uuid.UUID,
	userID uuid.UUID,
	channel DeliveryChannel,
	recipient string,
	priority int,
) *DeliveryQueueItem {
	now := time.Now()
	return &DeliveryQueueItem{
		ID:             uuid.New(),
		NotificationID: notificationID,
		OrganizationID: organizationID,
		UserID:         userID,
		Channel:        channel,
		Recipient:      recipient,
		Status:         DeliveryStatusPending,
		RetryCount:     0,
		MaxRetries:     3,
		Priority:       priority,
		Metadata:       make(map[string]string),
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

// CanRetry returns true if the item can be retried
func (d *DeliveryQueueItem) CanRetry() bool {
	return d.RetryCount < d.MaxRetries && d.Status != DeliveryStatusDelivered
}

// MarkAsDelivered marks the item as successfully delivered
func (d *DeliveryQueueItem) MarkAsDelivered(messageID string) {
	now := time.Now()
	d.Status = DeliveryStatusDelivered
	d.MessageID = messageID
	d.DeliveredAt = &now
	d.UpdatedAt = now
}

// MarkAsFailed marks the item as failed
func (d *DeliveryQueueItem) MarkAsFailed(err string) {
	d.Status = DeliveryStatusFailed
	d.LastError = err
	d.UpdatedAt = time.Now()
}

// ScheduleRetry schedules a retry with exponential backoff
func (d *DeliveryQueueItem) ScheduleRetry(err string) {
	d.RetryCount++
	d.LastError = err
	d.Status = DeliveryStatusRetrying

	// Exponential backoff: 1min, 5min, 15min
	backoffMinutes := []int{1, 5, 15}
	backoffIndex := d.RetryCount - 1
	if backoffIndex >= len(backoffMinutes) {
		backoffIndex = len(backoffMinutes) - 1
	}

	nextRetry := time.Now().Add(time.Duration(backoffMinutes[backoffIndex]) * time.Minute)
	d.NextRetryAt = &nextRetry
	d.UpdatedAt = time.Now()
}

// MarkAsInFlight marks the item as being processed
func (d *DeliveryQueueItem) MarkAsInFlight() {
	d.Status = DeliveryStatusInFlight
	d.UpdatedAt = time.Now()
}

// GetBackoffDuration returns the backoff duration for the current retry count
func (d *DeliveryQueueItem) GetBackoffDuration() time.Duration {
	backoffMinutes := []int{1, 5, 15}
	backoffIndex := d.RetryCount
	if backoffIndex >= len(backoffMinutes) {
		backoffIndex = len(backoffMinutes) - 1
	}
	return time.Duration(backoffMinutes[backoffIndex]) * time.Minute
}

// DeliveryQueue represents a collection of delivery items for batch processing
type DeliveryQueue struct {
	Items          []*DeliveryQueueItem
	TotalCount     int
	PendingCount   int
	FailedCount    int
	DeliveredCount int
}

// NewDeliveryQueue creates a new delivery queue
func NewDeliveryQueue() *DeliveryQueue {
	return &DeliveryQueue{
		Items: make([]*DeliveryQueueItem, 0),
	}
}

// AddItem adds an item to the queue
func (q *DeliveryQueue) AddItem(item *DeliveryQueueItem) {
	q.Items = append(q.Items, item)
	q.TotalCount++
	q.PendingCount++
}

// GetPendingItems returns all pending items sorted by priority
func (q *DeliveryQueue) GetPendingItems() []*DeliveryQueueItem {
	pending := make([]*DeliveryQueueItem, 0)
	for _, item := range q.Items {
		if item.Status == DeliveryStatusPending || item.Status == DeliveryStatusRetrying {
			pending = append(pending, item)
		}
	}
	return pending
}

// GetRetryableItems returns items that can be retried
func (q *DeliveryQueue) GetRetryableItems() []*DeliveryQueueItem {
	retryable := make([]*DeliveryQueueItem, 0)
	now := time.Now()
	for _, item := range q.Items {
		if item.Status == DeliveryStatusRetrying &&
			item.NextRetryAt != nil &&
			now.After(*item.NextRetryAt) &&
			item.CanRetry() {
			retryable = append(retryable, item)
		}
	}
	return retryable
}

// UpdateStats recalculates the queue statistics
func (q *DeliveryQueue) UpdateStats() {
	q.TotalCount = len(q.Items)
	q.PendingCount = 0
	q.FailedCount = 0
	q.DeliveredCount = 0

	for _, item := range q.Items {
		switch item.Status {
		case DeliveryStatusPending, DeliveryStatusRetrying, DeliveryStatusInFlight:
			q.PendingCount++
		case DeliveryStatusFailed:
			q.FailedCount++
		case DeliveryStatusDelivered:
			q.DeliveredCount++
		}
	}
}
