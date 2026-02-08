package tournament

import (
	"time"

	"github.com/google/uuid"
)

type CreateRequest struct {
	Name            string     `json:"name" binding:"required"`
	Description     string     `json:"description"`
	GameID          *uuid.UUID `json:"game_id"`
	EntryFee        float64    `json:"entry_fee"`
	PrizePool       float64    `json:"prize_pool"`
	MaxParticipants int        `json:"max_participants" binding:"required"`
	StartsAt        time.Time  `json:"starts_at" binding:"required"`
	EndsAt          time.Time  `json:"ends_at" binding:"required"`
}

type UpdateRequest struct {
	Name            *string    `json:"name,omitempty"`
	Description     *string    `json:"description,omitempty"`
	Status          *string    `json:"status,omitempty"`
	GameID          *uuid.UUID `json:"game_id,omitempty"`
	EntryFee        *float64   `json:"entry_fee,omitempty"`
	PrizePool       *float64   `json:"prize_pool,omitempty"`
	MaxParticipants *int       `json:"max_participants,omitempty"`
	StartsAt        *time.Time `json:"starts_at,omitempty"`
	EndsAt          *time.Time `json:"ends_at,omitempty"`
}

type JoinRequest struct {
	TournamentID uuid.UUID `json:"tournament_id" binding:"required"`
}

type UpdateScoreRequest struct {
	UserID uuid.UUID `json:"user_id" binding:"required"`
	Score  float64   `json:"score" binding:"required"`
}

type TournamentFilter struct {
	Status *string `form:"status"`
	GameID *string `form:"game_id"`
	Limit  int     `form:"limit,default=20"`
	Offset int     `form:"offset,default=0"`
}

type LeaderboardEntry struct {
	Rank     int       `json:"rank"`
	UserID   uuid.UUID `json:"user_id"`
	UserName string    `json:"user_name"`
	Score    float64   `json:"score"`
	PrizeWon float64   `json:"prize_won"`
}
