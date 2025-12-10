package domain

import (
	"time"

	"github.com/google/uuid"
)

type RefreshToken struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TokenHash string    `json:"-" gorm:"type:varchar(255);unique;not null"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
	ExpiresAt time.Time `json:"expires_at" gorm:"not null;index"`
	Revoked   bool      `json:"revoked" gorm:"not null;default:false"`
	CreatedAt time.Time `json:"created_at" gorm:"not null;default:CURRENT_TIMESTAMP"`
}

func (RefreshToken) TableName() string {
	return "refresh_tokens"
}
