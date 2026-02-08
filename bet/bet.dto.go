package bet

import (
	"time"

	"github.com/google/uuid"
)

type BetResponse struct {
	ID        uuid.UUID  `json:"id"`
	UserID    uuid.UUID  `json:"user_id"`
	Type      string     `json:"type"`
	GameID    *uuid.UUID `json:"game_id,omitempty"`
	EventID   *uuid.UUID `json:"event_id,omitempty"`
	OutcomeID *uuid.UUID `json:"outcome_id,omitempty"`
	Amount    float64    `json:"amount"`
	Odds      float64    `json:"odds"`
	Status    string     `json:"status"`
	Payout    float64    `json:"payout"`
	SettledAt *time.Time `json:"settled_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

type BetFilter struct {
	Type   *string `form:"type"`
	Status *string `form:"status"`
	Limit  int     `form:"limit,default=20"`
	Offset int     `form:"offset,default=0"`
}

type BetSummary struct {
	TotalBets    int64   `json:"total_bets"`
	TotalWagered float64 `json:"total_wagered"`
	TotalWon     float64 `json:"total_won"`
	TotalLost    float64 `json:"total_lost"`
	WinRate      float64 `json:"win_rate"`
}
