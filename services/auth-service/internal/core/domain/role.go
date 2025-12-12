package domain

import (
	"time"

	"github.com/google/uuid"
)

type Role struct {
	ID             uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name           string     `json:"name" gorm:"type:varchar(100);not null"`
	Description    string     `json:"description" gorm:"type:text"`
	OrganizationID *uuid.UUID `json:"organization_id,omitempty" gorm:"type:uuid;index:idx_roles_organization_id"`
	ParentRoleID   *uuid.UUID `json:"parent_role_id,omitempty" gorm:"type:uuid;index:idx_roles_parent_role_id"`
	IsSystem       bool       `json:"is_system" gorm:"not null;default:false;index:idx_roles_is_system"`
	CreatedAt      time.Time  `json:"created_at" gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt      time.Time  `json:"updated_at" gorm:"not null;default:CURRENT_TIMESTAMP"`

	Organization *Organization `json:"organization,omitempty" gorm:"foreignKey:OrganizationID"`
	ParentRole   *Role         `json:"parent_role,omitempty" gorm:"foreignKey:ParentRoleID"`
	Permissions  []Permission  `json:"permissions,omitempty" gorm:"many2many:role_permissions;"`
}

func (Role) TableName() string {
	return "roles"
}

type RoleResponse struct {
	ID             uuid.UUID            `json:"id"`
	Name           string               `json:"name"`
	Description    string               `json:"description"`
	OrganizationID *uuid.UUID           `json:"organization_id,omitempty"`
	ParentRoleID   *uuid.UUID           `json:"parent_role_id,omitempty"`
	IsSystem       bool                 `json:"is_system"`
	CreatedAt      time.Time            `json:"created_at"`
	UpdatedAt      time.Time            `json:"updated_at"`
	Permissions    []PermissionResponse `json:"permissions,omitempty"`
}

func (r *Role) ToResponse() *RoleResponse {
	response := &RoleResponse{
		ID:             r.ID,
		Name:           r.Name,
		Description:    r.Description,
		OrganizationID: r.OrganizationID,
		ParentRoleID:   r.ParentRoleID,
		IsSystem:       r.IsSystem,
		CreatedAt:      r.CreatedAt,
		UpdatedAt:      r.UpdatedAt,
	}

	if len(r.Permissions) > 0 {
		response.Permissions = make([]PermissionResponse, len(r.Permissions))
		for i, perm := range r.Permissions {
			response.Permissions[i] = *perm.ToResponse()
		}
	}

	return response
}

type CreateRoleRequest struct {
	Name           string   `json:"name" binding:"required,min=3,max=100"`
	Description    string   `json:"description"`
	OrganizationID *string  `json:"organization_id,omitempty" binding:"omitempty,uuid"`
	ParentRoleID   *string  `json:"parent_role_id,omitempty" binding:"omitempty,uuid"`
	PermissionIDs  []string `json:"permission_ids,omitempty"`
}

type UpdateRoleRequest struct {
	Name          string   `json:"name" binding:"omitempty,min=3,max=100"`
	Description   string   `json:"description"`
	ParentRoleID  *string  `json:"parent_role_id,omitempty" binding:"omitempty,uuid"`
	PermissionIDs []string `json:"permission_ids,omitempty"`
}

type AssignRoleRequest struct {
	UserID string `json:"user_id" binding:"required,uuid"`
	RoleID string `json:"role_id" binding:"required,uuid"`
}
