package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/melegattip/giia-core-engine/pkg/errors"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/core/providers"
)

type notificationRepository struct {
	db *sql.DB
}

func NewNotificationRepository(db *sql.DB) providers.NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) Create(ctx context.Context, notification *domain.AINotification) error {
	if notification == nil {
		return errors.NewBadRequest("notification cannot be nil")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.NewInternalServerError("failed to begin transaction")
	}
	defer tx.Rollback()

	sourceEventsJSON, err := json.Marshal(notification.SourceEvents)
	if err != nil {
		return errors.NewInternalServerError("failed to marshal source events")
	}

	relatedEntitiesJSON, err := json.Marshal(notification.RelatedEntities)
	if err != nil {
		return errors.NewInternalServerError("failed to marshal related entities")
	}

	var timeToImpactSeconds *int
	if notification.Impact.TimeToImpact != nil {
		seconds := int(notification.Impact.TimeToImpact.Seconds())
		timeToImpactSeconds = &seconds
	}

	query := `
		INSERT INTO ai_notifications (
			id, organization_id, user_id,
			type, priority,
			title, summary, full_analysis, reasoning,
			risk_level, revenue_impact, cost_impact, time_to_impact_seconds,
			affected_orders, affected_products,
			source_events, related_entities,
			status, created_at
		) VALUES (
			$1, $2, $3,
			$4, $5,
			$6, $7, $8, $9,
			$10, $11, $12, $13,
			$14, $15,
			$16, $17,
			$18, $19
		)
	`

	_, err = tx.ExecContext(ctx, query,
		notification.ID, notification.OrganizationID, notification.UserID,
		notification.Type, notification.Priority,
		notification.Title, notification.Summary, notification.FullAnalysis, notification.Reasoning,
		notification.Impact.RiskLevel, notification.Impact.RevenueImpact, notification.Impact.CostImpact, timeToImpactSeconds,
		notification.Impact.AffectedOrders, notification.Impact.AffectedProducts,
		sourceEventsJSON, relatedEntitiesJSON,
		notification.Status, notification.CreatedAt,
	)
	if err != nil {
		return errors.NewInternalServerError(fmt.Sprintf("failed to insert notification: %v", err))
	}

	for i, rec := range notification.Recommendations {
		recQuery := `
			INSERT INTO ai_recommendations (
				id, notification_id,
				action, reasoning, expected_outcome,
				effort, impact, action_url,
				priority_order
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`

		_, err = tx.ExecContext(ctx, recQuery,
			uuid.New(), notification.ID,
			rec.Action, rec.Reasoning, rec.ExpectedOutcome,
			rec.Effort, rec.Impact, rec.ActionURL,
			i+1,
		)
		if err != nil {
			return errors.NewInternalServerError(fmt.Sprintf("failed to insert recommendation: %v", err))
		}
	}

	if err := tx.Commit(); err != nil {
		return errors.NewInternalServerError("failed to commit transaction")
	}

	return nil
}

func (r *notificationRepository) GetByID(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) (*domain.AINotification, error) {
	query := `
		SELECT
			id, organization_id, user_id,
			type, priority,
			title, summary, full_analysis, reasoning,
			risk_level, revenue_impact, cost_impact, time_to_impact_seconds,
			affected_orders, affected_products,
			source_events, related_entities,
			status, created_at, read_at, acted_at, dismissed_at
		FROM ai_notifications
		WHERE id = $1 AND organization_id = $2
	`

	notification := &domain.AINotification{}
	var sourceEventsJSON, relatedEntitiesJSON []byte
	var timeToImpactSeconds *int

	err := r.db.QueryRowContext(ctx, query, id, organizationID).Scan(
		&notification.ID, &notification.OrganizationID, &notification.UserID,
		&notification.Type, &notification.Priority,
		&notification.Title, &notification.Summary, &notification.FullAnalysis, &notification.Reasoning,
		&notification.Impact.RiskLevel, &notification.Impact.RevenueImpact, &notification.Impact.CostImpact, &timeToImpactSeconds,
		&notification.Impact.AffectedOrders, &notification.Impact.AffectedProducts,
		&sourceEventsJSON, &relatedEntitiesJSON,
		&notification.Status, &notification.CreatedAt, &notification.ReadAt, &notification.ActedAt, &notification.DismissedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.NewNotFound("notification not found")
	}
	if err != nil {
		return nil, errors.NewInternalServerError(fmt.Sprintf("failed to query notification: %v", err))
	}

	if err := json.Unmarshal(sourceEventsJSON, &notification.SourceEvents); err != nil {
		return nil, errors.NewInternalServerError("failed to unmarshal source events")
	}

	if err := json.Unmarshal(relatedEntitiesJSON, &notification.RelatedEntities); err != nil {
		return nil, errors.NewInternalServerError("failed to unmarshal related entities")
	}

	recommendations, err := r.getRecommendations(ctx, notification.ID)
	if err != nil {
		return nil, err
	}
	notification.Recommendations = recommendations

	return notification, nil
}

func (r *notificationRepository) getRecommendations(ctx context.Context, notificationID uuid.UUID) ([]domain.Recommendation, error) {
	query := `
		SELECT action, reasoning, expected_outcome, effort, impact, action_url, priority_order
		FROM ai_recommendations
		WHERE notification_id = $1
		ORDER BY priority_order
	`

	rows, err := r.db.QueryContext(ctx, query, notificationID)
	if err != nil {
		return nil, errors.NewInternalServerError("failed to query recommendations")
	}
	defer rows.Close()

	var recommendations []domain.Recommendation
	for rows.Next() {
		var rec domain.Recommendation
		if err := rows.Scan(&rec.Action, &rec.Reasoning, &rec.ExpectedOutcome, &rec.Effort, &rec.Impact, &rec.ActionURL, &rec.PriorityOrder); err != nil {
			return nil, errors.NewInternalServerError("failed to scan recommendation")
		}
		recommendations = append(recommendations, rec)
	}

	return recommendations, nil
}

func (r *notificationRepository) List(ctx context.Context, userID uuid.UUID, organizationID uuid.UUID, filters *providers.NotificationFilters) ([]*domain.AINotification, error) {
	query := `
		SELECT
			id, organization_id, user_id,
			type, priority,
			title, summary, full_analysis, reasoning,
			risk_level, revenue_impact, cost_impact, time_to_impact_seconds,
			affected_orders, affected_products,
			source_events, related_entities,
			status, created_at, read_at, acted_at, dismissed_at
		FROM ai_notifications
		WHERE user_id = $1 AND organization_id = $2
	`

	args := []interface{}{userID, organizationID}
	argIndex := 3

	if filters != nil {
		if len(filters.Types) > 0 {
			query += fmt.Sprintf(" AND type = ANY($%d)", argIndex)
			args = append(args, pq.Array(filters.Types))
			argIndex++
		}

		if len(filters.Priorities) > 0 {
			query += fmt.Sprintf(" AND priority = ANY($%d)", argIndex)
			args = append(args, pq.Array(filters.Priorities))
			argIndex++
		}

		if len(filters.Statuses) > 0 {
			query += fmt.Sprintf(" AND status = ANY($%d)", argIndex)
			args = append(args, pq.Array(filters.Statuses))
			argIndex++
		}
	}

	query += " ORDER BY created_at DESC"

	if filters != nil && filters.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filters.Limit)
		argIndex++

		if filters.Offset > 0 {
			query += fmt.Sprintf(" OFFSET $%d", argIndex)
			args = append(args, filters.Offset)
		}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.NewInternalServerError(fmt.Sprintf("failed to query notifications: %v", err))
	}
	defer rows.Close()

	var notifications []*domain.AINotification
	for rows.Next() {
		notification := &domain.AINotification{}
		var sourceEventsJSON, relatedEntitiesJSON []byte
		var timeToImpactSeconds *int

		err := rows.Scan(
			&notification.ID, &notification.OrganizationID, &notification.UserID,
			&notification.Type, &notification.Priority,
			&notification.Title, &notification.Summary, &notification.FullAnalysis, &notification.Reasoning,
			&notification.Impact.RiskLevel, &notification.Impact.RevenueImpact, &notification.Impact.CostImpact, &timeToImpactSeconds,
			&notification.Impact.AffectedOrders, &notification.Impact.AffectedProducts,
			&sourceEventsJSON, &relatedEntitiesJSON,
			&notification.Status, &notification.CreatedAt, &notification.ReadAt, &notification.ActedAt, &notification.DismissedAt,
		)
		if err != nil {
			return nil, errors.NewInternalServerError("failed to scan notification")
		}

		if err := json.Unmarshal(sourceEventsJSON, &notification.SourceEvents); err != nil {
			return nil, errors.NewInternalServerError("failed to unmarshal source events")
		}

		if err := json.Unmarshal(relatedEntitiesJSON, &notification.RelatedEntities); err != nil {
			return nil, errors.NewInternalServerError("failed to unmarshal related entities")
		}

		notifications = append(notifications, notification)
	}

	for _, notif := range notifications {
		recommendations, err := r.getRecommendations(ctx, notif.ID)
		if err != nil {
			return nil, err
		}
		notif.Recommendations = recommendations
	}

	return notifications, nil
}

func (r *notificationRepository) Update(ctx context.Context, notification *domain.AINotification) error {
	if notification == nil {
		return errors.NewBadRequest("notification cannot be nil")
	}

	query := `
		UPDATE ai_notifications
		SET status = $1, read_at = $2, acted_at = $3, dismissed_at = $4
		WHERE id = $5 AND organization_id = $6
	`

	result, err := r.db.ExecContext(ctx, query,
		notification.Status, notification.ReadAt, notification.ActedAt, notification.DismissedAt,
		notification.ID, notification.OrganizationID,
	)
	if err != nil {
		return errors.NewInternalServerError(fmt.Sprintf("failed to update notification: %v", err))
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.NewInternalServerError("failed to get rows affected")
	}

	if rowsAffected == 0 {
		return errors.NewNotFound("notification not found")
	}

	return nil
}

func (r *notificationRepository) Delete(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) error {
	query := `DELETE FROM ai_notifications WHERE id = $1 AND organization_id = $2`

	result, err := r.db.ExecContext(ctx, query, id, organizationID)
	if err != nil {
		return errors.NewInternalServerError(fmt.Sprintf("failed to delete notification: %v", err))
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.NewInternalServerError("failed to get rows affected")
	}

	if rowsAffected == 0 {
		return errors.NewNotFound("notification not found")
	}

	return nil
}
