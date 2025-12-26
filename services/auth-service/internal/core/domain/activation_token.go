package domain

import (
	"time"

	"github.com/google/uuid"
)

type ActivationToken struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TokenHash string    `json:"-" gorm:"type:varchar(255);unique;not null"`
	UserID    int       `json:"user_id" gorm:"not null;index"`
	ExpiresAt time.Time `json:"expires_at" gorm:"not null"`
	Used      bool      `json:"used" gorm:"not null;default:false"`
	CreatedAt time.Time `json:"created_at" gorm:"not null;default:CURRENT_TIMESTAMP"`
}

func (ActivationToken) TableName() string {
	return "activation_tokens"
}
