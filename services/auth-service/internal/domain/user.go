package domain

import "time"

type User struct {
	ID                       uint       `json:"id"`
	Email                    string     `json:"email"`
	Password                 string     `json:"-"`
	FirstName                string     `json:"first_name"`
	LastName                 string     `json:"last_name"`
	Phone                    string     `json:"phone"`
	Avatar                   string     `json:"avatar,omitempty"`
	IsActive                 bool       `json:"is_active"`
	IsVerified               bool       `json:"is_verified"`
	EmailVerificationToken   string     `json:"-"`
	EmailVerificationExpires *time.Time `json:"-"`
	PasswordResetToken       string     `json:"-"`
	PasswordResetExpires     *time.Time `json:"-"`
	LastLogin                *time.Time `json:"last_login,omitempty"`
	FailedLoginAttempts      int        `json:"-"`
	LockedUntil              *time.Time `json:"-"`
	CreatedAt                time.Time  `json:"created_at"`
	UpdatedAt                time.Time  `json:"updated_at"`
}

type Preferences struct {
	UserID     uint   `json:"user_id"`
	Currency   string `json:"currency"`
	Language   string `json:"language"`
	Theme      string `json:"theme"`
	DateFormat string `json:"date_format"`
	Timezone   string `json:"timezone"`
}

type NotificationSettings struct {
	UserID                   uint `json:"user_id"`
	EmailNotifications       bool `json:"email_notifications"`
	PushNotifications        bool `json:"push_notifications"`
	WeeklyReports            bool `json:"weekly_reports"`
	ExpenseAlerts            bool `json:"expense_alerts"`
	BudgetAlerts             bool `json:"budget_alerts"`
	AchievementNotifications bool `json:"achievement_notifications"`
}

type TwoFA struct {
	UserID       uint     `json:"user_id"`
	Secret       string   `json:"secret"`
	Enabled      bool     `json:"enabled"`
	BackupCodes  []string `json:"backup_codes,omitempty"`
	LastUsedCode string   `json:"-"`
}

// Request/Response DTOs
type RegisterRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Phone     string `json:"phone"`
}

type LoginRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required"`
	TwoFACode string `json:"twofa_code,omitempty"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

type UserResponse struct {
	ID         uint       `json:"id"`
	Email      string     `json:"email"`
	FirstName  string     `json:"first_name"`
	LastName   string     `json:"last_name"`
	Phone      string     `json:"phone"`
	Avatar     string     `json:"avatar,omitempty"`
	IsActive   bool       `json:"is_active"`
	IsVerified bool       `json:"is_verified"`
	LastLogin  *time.Time `json:"last_login,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

// UpdateProfileRequest DTO específico para actualización de perfil
type UpdateProfileRequest struct {
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Phone     string `json:"phone"`
	Avatar    string `json:"avatar,omitempty"`
}
