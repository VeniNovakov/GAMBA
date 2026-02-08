package models

import (
	"time"

	"github.com/google/uuid"
)

type FriendStatus string

const (
	FriendStatusPending  FriendStatus = "pending"
	FriendStatusAccepted FriendStatus = "accepted"
	FriendStatusRejected FriendStatus = "rejected"
	FriendStatusBlocked  FriendStatus = "blocked"
)

type Friend struct {
	ID        uuid.UUID    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    uuid.UUID    `json:"user_id" gorm:"type:uuid;not null;index"`
	FriendID  uuid.UUID    `json:"friend_id" gorm:"type:uuid;not null;index"`
	Status    FriendStatus `json:"status" gorm:"type:varchar(20);default:'pending'"`
	CreatedAt time.Time    `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time    `json:"updated_at" gorm:"autoUpdateTime"`

	User   User `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Friend User `json:"friend,omitempty" gorm:"foreignKey:FriendID"`
}

func (Friend) TableName() string { return "friends" }
