package game

import "github.com/google/uuid"

type CreateRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Category    string  `json:"category"` // Added category field
	MinBet      float64 `json:"min_bet"`
	MaxBet      float64 `json:"max_bet"`
	HouseEdge   float64 `json:"house_edge"`
}

type UpdateRequest struct {
	Name        *string  `json:"name,omitempty"`
	Description *string  `json:"description,omitempty"`
	Category    *string  `json:"category,omitempty"` // Added category field
	Status      *string  `json:"status,omitempty"`
	MinBet      *float64 `json:"min_bet,omitempty"`
	MaxBet      *float64 `json:"max_bet,omitempty"`
	HouseEdge   *float64 `json:"house_edge,omitempty"`
}

type PlayRequest struct {
	GameID    uuid.UUID `json:"game_id"`
	BetAmount float64   `json:"bet_amount"`
}

// PlayResponse for slots
type PlayResponse struct {
	Reels      [3]string `json:"reels,omitempty"`      // For slots
	Dice       []int     `json:"dice,omitempty"`       // For dice
	Target     int       `json:"target,omitempty"`     // For dice
	Won        bool      `json:"won"`
	Payout     float64   `json:"payout"`
	Multiplier float64   `json:"multiplier"`
	NewBalance float64   `json:"new_balance"`
}
