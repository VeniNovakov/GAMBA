package models

import (
	"time"

	"github.com/google/uuid"
)

type RefreshToken struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    uuid.UUID  `json:"user_id" gorm:"type:uuid;not null;index"`
	TokenHash string     `json:"-" gorm:"not null"`
	ExpiresAt time.Time  `json:"expires_at" gorm:"not null"`
	IsRevoked bool       `json:"is_revoked" gorm:"default:false"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
	CreatedAt time.Time  `json:"created_at" gorm:"autoCreateTime"`
}

func (RefreshToken) TableName() string { return "refresh_tokens" }
