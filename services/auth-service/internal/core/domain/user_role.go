package domain

import (
	"time"

	"github.com/google/uuid"
)

type UserRole struct {
	ID         uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID     uuid.UUID  `json:"user_id" gorm:"type:uuid;not null;index:idx_user_roles_user_id"`
	RoleID     uuid.UUID  `json:"role_id" gorm:"type:uuid;not null;index:idx_user_roles_role_id"`
	AssignedAt time.Time  `json:"assigned_at" gorm:"not null;default:CURRENT_TIMESTAMP;index:idx_user_roles_assigned_at"`
	AssignedBy *uuid.UUID `json:"assigned_by,omitempty" gorm:"type:uuid;index:idx_user_roles_assigned_by"`

	User       *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Role       *Role `json:"role,omitempty" gorm:"foreignKey:RoleID"`
	Assigner   *User `json:"assigner,omitempty" gorm:"foreignKey:AssignedBy"`
}

func (UserRole) TableName() string {
	return "user_roles"
}

type UserRoleResponse struct {
	ID         uuid.UUID     `json:"id"`
	UserID     uuid.UUID     `json:"user_id"`
	RoleID     uuid.UUID     `json:"role_id"`
	AssignedAt time.Time     `json:"assigned_at"`
	AssignedBy *uuid.UUID    `json:"assigned_by,omitempty"`
	Role       *RoleResponse `json:"role,omitempty"`
}

func (ur *UserRole) ToResponse() *UserRoleResponse {
	response := &UserRoleResponse{
		ID:         ur.ID,
		UserID:     ur.UserID,
		RoleID:     ur.RoleID,
		AssignedAt: ur.AssignedAt,
		AssignedBy: ur.AssignedBy,
	}

	if ur.Role != nil {
		response.Role = ur.Role.ToResponse()
	}

	return response
}