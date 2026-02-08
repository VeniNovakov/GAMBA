package chat

import (
	"time"

	"github.com/google/uuid"
)

type CreateChatRequest struct {
	UserID uuid.UUID `json:"user_id" binding:"required"`
}

type SendMessageRequest struct {
	Content string `json:"content" binding:"required"`
}

type ChatResponse struct {
	ID          uuid.UUID        `json:"id"`
	User1       *UserResponse    `json:"user1,omitempty"`
	User2       *UserResponse    `json:"user2,omitempty"`
	LastMessage *MessageResponse `json:"last_message,omitempty"`
	CreatedAt   time.Time        `json:"created_at"`
}

type MessageResponse struct {
	ID        uuid.UUID     `json:"id"`
	ChatID    uuid.UUID     `json:"chat_id"`
	SenderID  uuid.UUID     `json:"sender_id"`
	Sender    *UserResponse `json:"sender,omitempty"`
	Content   string        `json:"content"`
	ReadAt    *time.Time    `json:"read_at,omitempty"`
	CreatedAt time.Time     `json:"created_at"`
}

type UserResponse struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
}

type ChatFilter struct {
	Limit  int `form:"limit,default=20"`
	Offset int `form:"offset,default=0"`
}

type MessageFilter struct {
	Before *time.Time `form:"before"`
	Limit  int        `form:"limit,default=50"`
}

// WebSocket message types
type WSMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type WSChatMessage struct {
	ChatID  uuid.UUID `json:"chat_id"`
	Content string    `json:"content"`
}

type WSTypingEvent struct {
	ChatID uuid.UUID `json:"chat_id"`
	UserID uuid.UUID `json:"user_id"`
}

type WSReadEvent struct {
	ChatID    uuid.UUID `json:"chat_id"`
	MessageID uuid.UUID `json:"message_id"`
}
