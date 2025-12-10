package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID             uuid.UUID   `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Email          string      `json:"email" gorm:"type:varchar(255);not null;index:idx_users_email_org"`
	Password       string      `json:"-" gorm:"type:varchar(255);not null"`
	FirstName      string      `json:"first_name" gorm:"type:varchar(100)"`
	LastName       string      `json:"last_name" gorm:"type:varchar(100)"`
	Phone          string      `json:"phone" gorm:"type:varchar(20)"`
	Avatar         string      `json:"avatar,omitempty" gorm:"type:varchar(500)"`
	Status         UserStatus  `json:"status" gorm:"type:varchar(20);not null;default:'inactive'"`
	OrganizationID uuid.UUID   `json:"organization_id" gorm:"type:uuid;not null;index:idx_users_organization_id,idx_users_email_org"`
	LastLoginAt    *time.Time  `json:"last_login_at,omitempty" gorm:"type:timestamp"`
	CreatedAt      time.Time   `json:"created_at" gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt      time.Time   `json:"updated_at" gorm:"not null;default:CURRENT_TIMESTAMP"`
	Organization   Organization `json:"organization,omitempty" gorm:"foreignKey:OrganizationID"`
}

type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusInactive UserStatus = "inactive"
	UserStatusSuspended UserStatus = "suspended"
)

func (User) TableName() string {
	return "users"
}

type UserResponse struct {
	ID             uuid.UUID  `json:"id"`
	Email          string     `json:"email"`
	FirstName      string     `json:"first_name"`
	LastName       string     `json:"last_name"`
	Phone          string     `json:"phone"`
	Avatar         string     `json:"avatar,omitempty"`
	Status         UserStatus `json:"status"`
	OrganizationID uuid.UUID  `json:"organization_id"`
	LastLoginAt    *time.Time `json:"last_login_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		ID:             u.ID,
		Email:          u.Email,
		FirstName:      u.FirstName,
		LastName:       u.LastName,
		Phone:          u.Phone,
		Avatar:         u.Avatar,
		Status:         u.Status,
		OrganizationID: u.OrganizationID,
		LastLoginAt:    u.LastLoginAt,
		CreatedAt:      u.CreatedAt,
	}
}

type RegisterRequest struct {
	Email          string `json:"email" binding:"required,email"`
	Password       string `json:"password" binding:"required,min=8"`
	FirstName      string `json:"first_name" binding:"required"`
	LastName       string `json:"last_name" binding:"required"`
	Phone          string `json:"phone"`
	OrganizationID string `json:"organization_id" binding:"required,uuid"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token"`
	ExpiresIn    int           `json:"expires_in"`
	User         *UserResponse `json:"user"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

type UpdateProfileRequest struct {
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Phone     string `json:"phone"`
	Avatar    string `json:"avatar,omitempty"`
}

type PasswordResetRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type PasswordResetComplete struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

type ActivateAccountRequest struct {
	Token string `json:"token" binding:"required"`
}
