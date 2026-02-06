package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BetStatus string

const (
	BetStatusPending  BetStatus = "pending"
	BetStatusWon      BetStatus = "won"
	BetStatusLost     BetStatus = "lost"
	BetStatusRefunded BetStatus = "refunded"
)

type BetType string

const (
	BetTypeGame  BetType = "game"
	BetTypeEvent BetType = "event"
)

type Bet struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    uuid.UUID      `json:"user_id" gorm:"type:uuid;not null;index"`
	Type      BetType        `json:"type" gorm:"type:varchar(20);not null"`
	GameID    *uuid.UUID     `json:"game_id,omitempty" gorm:"type:uuid;index"`
	EventID   *uuid.UUID     `json:"event_id,omitempty" gorm:"type:uuid;index"`
	OutcomeID *uuid.UUID     `json:"outcome_id,omitempty" gorm:"type:uuid;index"`
	Amount    int64          `json:"amount" gorm:"not null"`
	Odds      int64          `json:"odds" gorm:"not null"`
	Status    BetStatus      `json:"status" gorm:"type:varchar(20);default:'pending'"`
	Payout    int64          `json:"payout" gorm:"default:0"`
	SettledAt *time.Time     `json:"settled_at,omitempty"`
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	//relationships
	User    User          `json:"-" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Game    *Game         `json:"-" gorm:"foreignKey:GameID;constraint:OnDelete:SET NULL"`
	Event   *Event        `json:"-" gorm:"foreignKey:EventID;constraint:OnDelete:SET NULL"`
	Outcome *EventOutcome `json:"-" gorm:"foreignKey:OutcomeID;constraint:OnDelete:SET NULL"`
}

func (Bet) TableName() string { return "bets" }
