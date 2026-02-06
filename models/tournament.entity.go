package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TournamentStatus string

const (
	TournamentStatusDraft      TournamentStatus = "draft"
	TournamentStatusOpen       TournamentStatus = "open"
	TournamentStatusInProgress TournamentStatus = "in_progress"
	TournamentStatusCompleted  TournamentStatus = "completed"
	TournamentStatusCancelled  TournamentStatus = "cancelled"
)

type Tournament struct {
	ID              uuid.UUID        `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name            string           `json:"name" gorm:"not null"`
	Description     string           `json:"description" gorm:"type:text"`
	Status          TournamentStatus `json:"status" gorm:"type:varchar(20);default:'draft'"`
	GameID          *uuid.UUID       `json:"game_id,omitempty" gorm:"type:uuid;index"`
	EntryFee        int64            `json:"entry_fee" gorm:"default:0"`
	PrizePool       int64            `json:"prize_pool" gorm:"default:0"`
	MaxParticipants int              `json:"max_participants" gorm:"not null"`
	StartsAt        time.Time        `json:"starts_at" gorm:"not null"`
	EndsAt          time.Time        `json:"ends_at" gorm:"not null"`
	CreatedAt       time.Time        `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time        `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt       gorm.DeletedAt   `json:"-" gorm:"index"`

	//relationships
	Game         *Game                   `json:"-" gorm:"foreignKey:GameID;constraint:OnDelete:SET NULL"`
	Participants []TournamentParticipant `json:"participants,omitempty" gorm:"foreignKey:TournamentID"`
}

type TournamentParticipant struct {
	ID           uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TournamentID uuid.UUID      `json:"tournament_id" gorm:"type:uuid;not null;index"`
	UserID       uuid.UUID      `json:"user_id" gorm:"type:uuid;not null;index"`
	Score        float64        `json:"score" gorm:"default:0"`
	Rank         int            `json:"rank" gorm:"default:0"`
	PrizeWon     float64        `json:"prize_won" gorm:"default:0"`
	JoinedAt     time.Time      `json:"joined_at" gorm:"autoCreateTime"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`

	//relationships
	Tournament Tournament `json:"-" gorm:"foreignKey:TournamentID;constraint:OnDelete:CASCADE"`
	User       User       `json:"-" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

func (Tournament) TableName() string            { return "tournaments" }
func (TournamentParticipant) TableName() string { return "tournament_participants" }
