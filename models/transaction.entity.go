package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TransactionType string

const (
	TransactionTypeDeposit         TransactionType = "deposit"
	TransactionTypeWithdrawal      TransactionType = "withdrawal"
	TransactionTypeBet             TransactionType = "bet"
	TransactionTypeWin             TransactionType = "win"
	TransactionTypeRefund          TransactionType = "refund"
	TransactionTypeTournamentEntry TransactionType = "tournament_entry"
	TransactionTypeTournamentPrize TransactionType = "tournament_prize"
	TransactionTypeTransfer        TransactionType = "transfer"
)

type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "pending"
	TransactionStatusCompleted TransactionStatus = "completed"
	TransactionStatusFailed    TransactionStatus = "failed"
	TransactionStatusCancelled TransactionStatus = "cancelled"
)

type Transaction struct {
	ID            uuid.UUID         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID        uuid.UUID         `json:"user_id" gorm:"type:uuid;not null;index"`
	Type          TransactionType   `json:"type" gorm:"type:varchar(30);not null"`
	Status        TransactionStatus `json:"status" gorm:"type:varchar(20);default:'pending'"`
	Amount        int64             `json:"amount" gorm:"not null"` // positive = credit, negative = debit
	ReferenceID   *uuid.UUID        `json:"reference_id,omitempty" gorm:"type:uuid;index"`
	ReferenceType *string           `json:"reference_type,omitempty" gorm:"type:varchar(30)"`
	Description   string            `json:"description" gorm:"type:text"`
	CreatedAt     time.Time         `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time         `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt     gorm.DeletedAt    `json:"-" gorm:"index"`

	// relationships
	User User `json:"-" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

func (Transaction) TableName() string { return "transactions" }
