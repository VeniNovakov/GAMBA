package models

import (
	"time"

	"github.com/google/uuid"
)

type Chat struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	User1ID   uuid.UUID `json:"user1_id" gorm:"type:uuid;not null;uniqueIndex:idx_chat_users"`
	User2ID   uuid.UUID `json:"user2_id" gorm:"type:uuid;not null;uniqueIndex:idx_chat_users"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	User1    User      `json:"user1,omitempty" gorm:"foreignKey:User1ID"`
	User2    User      `json:"user2,omitempty" gorm:"foreignKey:User2ID"`
	Messages []Message `json:"messages,omitempty" gorm:"foreignKey:ChatID"`
}

type Message struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ChatID    uuid.UUID  `json:"chat_id" gorm:"type:uuid;not null;index"`
	SenderID  uuid.UUID  `json:"sender_id" gorm:"type:uuid;not null;index"`
	Content   string     `json:"content" gorm:"type:text;not null"`
	ReadAt    *time.Time `json:"read_at,omitempty"`
	CreatedAt time.Time  `json:"created_at" gorm:"autoCreateTime"`

	// Relationships
	Chat   Chat `json:"-" gorm:"foreignKey:ChatID"`
	Sender User `json:"sender,omitempty" gorm:"foreignKey:SenderID"`
}

func (Chat) TableName() string    { return "chats" }
func (Message) TableName() string { return "messages" }
