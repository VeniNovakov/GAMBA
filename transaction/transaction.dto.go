package transaction

import (
	"time"

	"github.com/google/uuid"
)

type DepositRequest struct {
	Amount float64 `json:"amount" binding:"required,gt=0"`
}

type WithdrawRequest struct {
	Amount float64 `json:"amount" binding:"required,gt=0"`
}

type TransferRequest struct {
	ToUserID uuid.UUID `json:"to_user_id" binding:"required"`
	Amount   float64   `json:"amount" binding:"required,gt=0"`
}

type TransactionFilter struct {
	Type   *string `form:"type"`
	Status *string `form:"status"`
	Limit  int     `form:"limit,default=20"`
	Offset int     `form:"offset,default=0"`
}

type TransactionSummary struct {
	TotalDeposits    float64 `json:"total_deposits"`
	TotalWithdrawals float64 `json:"total_withdrawals"`
	TotalBets        float64 `json:"total_bets"`
	TotalWins        float64 `json:"total_wins"`
	NetBalance       float64 `json:"net_balance"`
}

type TransactionResponse struct {
	ID            uuid.UUID  `json:"id"`
	UserID        uuid.UUID  `json:"user_id"`
	Type          string     `json:"type"`
	Status        string     `json:"status"`
	Amount        float64    `json:"amount"`
	ReferenceID   *uuid.UUID `json:"reference_id,omitempty"`
	ReferenceType *string    `json:"reference_type,omitempty"`
	Description   string     `json:"description"`
	CreatedAt     time.Time  `json:"created_at"`
}
