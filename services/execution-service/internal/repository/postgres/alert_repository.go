package postgres

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/melegattip/giia-core-engine/services/execution-service/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/execution-service/internal/core/providers"
	"gorm.io/gorm"
)

// AlertModel represents the database model for alerts
type AlertModel struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	OrganizationID uuid.UUID  `gorm:"type:uuid;not null;index:idx_alert_org"`
	AlertType      string     `gorm:"type:varchar(50);not null;index:idx_alert_type"`
	Severity       string     `gorm:"type:varchar(20);not null;index:idx_alert_severity"`
	ResourceType   string     `gorm:"type:varchar(50);not null"`
	ResourceID     uuid.UUID  `gorm:"type:uuid;not null;index:idx_alert_resource"`
	Title          string     `gorm:"type:varchar(255);not null"`
	Message        string     `gorm:"type:text;not null"`
	Data           string     `gorm:"type:jsonb"`
	AcknowledgedAt *time.Time `gorm:""`
	AcknowledgedBy *uuid.UUID `gorm:"type:uuid"`
	ResolvedAt     *time.Time `gorm:""`
	ResolvedBy     *uuid.UUID `gorm:"type:uuid"`
	CreatedAt      time.Time  `gorm:"not null;default:now();index:idx_alert_created"`
}

func (AlertModel) TableName() string {
	return "alerts"
}

type alertRepository struct {
	db *gorm.DB
}

// NewAlertRepository creates a new alert repository
func NewAlertRepository(db *gorm.DB) providers.AlertRepository {
	return &alertRepository{db: db}
}

func (r *alertRepository) scopeByOrg(orgID uuid.UUID) *gorm.DB {
	return r.db.Where("organization_id = ?", orgID)
}

func (r *alertRepository) Create(ctx context.Context, alert *domain.Alert) error {
	model, err := r.toModel(alert)
	if err != nil {
		return err
	}

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}
	return nil
}

func (r *alertRepository) GetByID(ctx context.Context, id, organizationID uuid.UUID) (*domain.Alert, error) {
	var model AlertModel
	err := r.scopeByOrg(organizationID).
		WithContext(ctx).
		Where("id = ?", id).
		First(&model).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return r.toDomain(&model)
}

func (r *alertRepository) Update(ctx context.Context, alert *domain.Alert) error {
	model, err := r.toModel(alert)
	if err != nil {
		return err
	}

	result := r.db.WithContext(ctx).
		Where("id = ? AND organization_id = ?", model.ID, model.OrganizationID).
		Updates(model)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *alertRepository) List(ctx context.Context, organizationID uuid.UUID, filters map[string]interface{}, page, pageSize int) ([]*domain.Alert, int64, error) {
	var models []AlertModel
	var total int64

	query := r.scopeByOrg(organizationID).WithContext(ctx).Model(&AlertModel{})

	// Apply filters
	for key, value := range filters {
		if value != nil && value != "" {
			switch key {
			case "alert_type":
				query = query.Where("alert_type = ?", value)
			case "severity":
				query = query.Where("severity = ?", value)
			case "resource_type":
				query = query.Where("resource_type = ?", value)
			case "acknowledged":
				if value.(bool) {
					query = query.Where("acknowledged_at IS NOT NULL")
				} else {
					query = query.Where("acknowledged_at IS NULL")
				}
			case "resolved":
				if value.(bool) {
					query = query.Where("resolved_at IS NOT NULL")
				} else {
					query = query.Where("resolved_at IS NULL")
				}
			case "from_date":
				query = query.Where("created_at >= ?", value)
			case "to_date":
				query = query.Where("created_at <= ?", value)
			}
		}
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Paginate
	offset := (page - 1) * pageSize
	if err := query.
		Offset(offset).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&models).Error; err != nil {
		return nil, 0, err
	}

	// Convert to domain
	result := make([]*domain.Alert, 0, len(models))
	for i := range models {
		alert, err := r.toDomain(&models[i])
		if err != nil {
			return nil, 0, err
		}
		result = append(result, alert)
	}

	return result, total, nil
}

func (r *alertRepository) GetActiveAlerts(ctx context.Context, organizationID uuid.UUID) ([]*domain.Alert, error) {
	var models []AlertModel

	err := r.scopeByOrg(organizationID).
		WithContext(ctx).
		Where("acknowledged_at IS NULL").
		Where("resolved_at IS NULL").
		Order("CASE severity WHEN 'critical' THEN 1 WHEN 'high' THEN 2 WHEN 'medium' THEN 3 WHEN 'low' THEN 4 ELSE 5 END").
		Order("created_at DESC").
		Find(&models).Error

	if err != nil {
		return nil, err
	}

	result := make([]*domain.Alert, 0, len(models))
	for i := range models {
		alert, err := r.toDomain(&models[i])
		if err != nil {
			return nil, err
		}
		result = append(result, alert)
	}

	return result, nil
}

func (r *alertRepository) GetByResourceID(ctx context.Context, resourceType string, resourceID, organizationID uuid.UUID) ([]*domain.Alert, error) {
	var models []AlertModel

	err := r.scopeByOrg(organizationID).
		WithContext(ctx).
		Where("resource_type = ?", resourceType).
		Where("resource_id = ?", resourceID).
		Order("created_at DESC").
		Find(&models).Error

	if err != nil {
		return nil, err
	}

	result := make([]*domain.Alert, 0, len(models))
	for i := range models {
		alert, err := r.toDomain(&models[i])
		if err != nil {
			return nil, err
		}
		result = append(result, alert)
	}

	return result, nil
}

// toModel converts domain entity to database model
func (r *alertRepository) toModel(alert *domain.Alert) (*AlertModel, error) {
	dataJSON := ""
	if alert.Data != nil {
		data, err := json.Marshal(alert.Data)
		if err != nil {
			return nil, err
		}
		dataJSON = string(data)
	}

	return &AlertModel{
		ID:             alert.ID,
		OrganizationID: alert.OrganizationID,
		AlertType:      string(alert.AlertType),
		Severity:       string(alert.Severity),
		ResourceType:   alert.ResourceType,
		ResourceID:     alert.ResourceID,
		Title:          alert.Title,
		Message:        alert.Message,
		Data:           dataJSON,
		AcknowledgedAt: alert.AcknowledgedAt,
		AcknowledgedBy: alert.AcknowledgedBy,
		ResolvedAt:     alert.ResolvedAt,
		ResolvedBy:     alert.ResolvedBy,
		CreatedAt:      alert.CreatedAt,
	}, nil
}

// toDomain converts database model to domain entity
func (r *alertRepository) toDomain(model *AlertModel) (*domain.Alert, error) {
	var data map[string]interface{}
	if model.Data != "" {
		if err := json.Unmarshal([]byte(model.Data), &data); err != nil {
			return nil, err
		}
	}

	return &domain.Alert{
		ID:             model.ID,
		OrganizationID: model.OrganizationID,
		AlertType:      domain.AlertType(model.AlertType),
		Severity:       domain.AlertSeverity(model.Severity),
		ResourceType:   model.ResourceType,
		ResourceID:     model.ResourceID,
		Title:          model.Title,
		Message:        model.Message,
		Data:           data,
		AcknowledgedAt: model.AcknowledgedAt,
		AcknowledgedBy: model.AcknowledgedBy,
		ResolvedAt:     model.ResolvedAt,
		ResolvedBy:     model.ResolvedBy,
		CreatedAt:      model.CreatedAt,
	}, nil
}
