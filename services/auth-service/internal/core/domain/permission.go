package domain

import (
	"time"

	"github.com/google/uuid"
)

type Permission struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Code        string    `json:"code" gorm:"type:varchar(255);not null;uniqueIndex:idx_permissions_code"`
	Description string    `json:"description" gorm:"type:text"`
	Service     string    `json:"service" gorm:"type:varchar(50);not null;index:idx_permissions_service"`
	Resource    string    `json:"resource" gorm:"type:varchar(100);not null"`
	Action      string    `json:"action" gorm:"type:varchar(50);not null"`
	CreatedAt   time.Time `json:"created_at" gorm:"not null;default:CURRENT_TIMESTAMP"`
}

func (Permission) TableName() string {
	return "permissions"
}

type PermissionResponse struct {
	ID          uuid.UUID `json:"id"`
	Code        string    `json:"code"`
	Description string    `json:"description"`
	Service     string    `json:"service"`
	Resource    string    `json:"resource"`
	Action      string    `json:"action"`
	CreatedAt   time.Time `json:"created_at"`
}

func (p *Permission) ToResponse() *PermissionResponse {
	return &PermissionResponse{
		ID:          p.ID,
		Code:        p.Code,
		Description: p.Description,
		Service:     p.Service,
		Resource:    p.Resource,
		Action:      p.Action,
		CreatedAt:   p.CreatedAt,
	}
}

type CreatePermissionRequest struct {
	Code        string `json:"code" binding:"required,min=5,max=255"`
	Description string `json:"description"`
	Service     string `json:"service" binding:"required,min=2,max=50"`
	Resource    string `json:"resource" binding:"required,min=2,max=100"`
	Action      string `json:"action" binding:"required,min=2,max=50"`
}

type CheckPermissionRequest struct {
	UserID     string `json:"user_id" binding:"required,uuid"`
	Permission string `json:"permission" binding:"required,min=5"`
}

type CheckPermissionResponse struct {
	Allowed bool `json:"allowed"`
}

type BatchCheckPermissionRequest struct {
	UserID      string   `json:"user_id" binding:"required,uuid"`
	Permissions []string `json:"permissions" binding:"required,min=1"`
}

type BatchCheckPermissionResponse struct {
	Results map[string]bool `json:"results"`
}
