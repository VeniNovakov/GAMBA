package event

import (
	"gamba/models"
	"time"

	"github.com/google/uuid"
)

type CreateRequest struct {
	Name        string               `json:"name" binding:"required"`
	Description string               `json:"description"`
	Category    models.EventCategory `json:"category" binding:"required"`
	StartsAt    time.Time            `json:"starts_at" binding:"required"`
	EndsAt      *time.Time           `json:"ends_at,omitempty"`
}

type UpdateRequest struct {
	Name        *string               `json:"name,omitempty"`
	Description *string               `json:"description,omitempty"`
	Category    *models.EventCategory `json:"category,omitempty"`
	Status      *string               `json:"status,omitempty"`
	StartsAt    *time.Time            `json:"starts_at,omitempty"`
	EndsAt      *time.Time            `json:"ends_at,omitempty"`
}

type CreateOutcomeRequest struct {
	Name string  `json:"name" binding:"required"`
	Odds float64 `json:"odds" binding:"required,gt=1"`
}

type UpdateOutcomeRequest struct {
	Name     *string  `json:"name,omitempty"`
	Odds     *float64 `json:"odds,omitempty"`
	IsWinner *bool    `json:"is_winner,omitempty"`
}

type PlaceBetRequest struct {
	OutcomeID uuid.UUID `json:"outcome_id" binding:"required"`
	Amount    float64   `json:"amount" binding:"required,gt=0"`
}

type SettleRequest struct {
	WinningOutcomeID uuid.UUID `json:"winning_outcome_id" binding:"required"`
}

type EventFilter struct {
	Status   *string `form:"status"`
	Category *string `form:"category"`
	Limit    int     `form:"limit,default=20"`
	Offset   int     `form:"offset,default=0"`
}
