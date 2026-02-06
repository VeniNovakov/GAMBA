package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TicketStatus string

const (
	TicketStatusOpen       TicketStatus = "open"
	TicketStatusInProgress TicketStatus = "in_progress"
	TicketStatusResolved   TicketStatus = "resolved"
	TicketStatusClosed     TicketStatus = "closed"
)

type TicketPriority string

const (
	TicketPriorityLow    TicketPriority = "low"
	TicketPriorityMedium TicketPriority = "medium"
	TicketPriorityHigh   TicketPriority = "high"
)

type Ticket struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID      uuid.UUID      `json:"user_id" gorm:"type:uuid;not null;index"`
	AssignedTo  *uuid.UUID     `json:"assigned_to,omitempty" gorm:"type:uuid;index"`
	Subject     string         `json:"subject" gorm:"not null"`
	Description string         `json:"description" gorm:"type:text;not null"`
	Status      TicketStatus   `json:"status" gorm:"type:varchar(20);default:'open'"`
	Priority    TicketPriority `json:"priority" gorm:"type:varchar(20);default:'medium'"`
	ResolvedAt  *time.Time     `json:"resolved_at,omitempty"`
	CreatedAt   time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	//relationships
	User     User            `json:"-" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Assignee *User           `json:"-" gorm:"foreignKey:AssignedTo;constraint:OnDelete:SET NULL"`
	Messages []TicketMessage `json:"messages,omitempty" gorm:"foreignKey:TicketID"`
}

type TicketMessage struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TicketID  uuid.UUID      `json:"ticket_id" gorm:"type:uuid;not null;index"`
	SenderID  uuid.UUID      `json:"sender_id" gorm:"type:uuid;not null;index"`
	Content   string         `json:"content" gorm:"type:text;not null"`
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	//relationships
	Ticket Ticket `json:"-" gorm:"foreignKey:TicketID;constraint:OnDelete:CASCADE"`
	Sender User   `json:"-" gorm:"foreignKey:SenderID;constraint:OnDelete:CASCADE"`
}

func (Ticket) TableName() string        { return "tickets" }
func (TicketMessage) TableName() string { return "ticket_messages" }
