package transaction

import (
	"errors"

	"gamba/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrTransactionNotFound = errors.New("transaction not found")
	ErrInsufficientFunds   = errors.New("insufficient funds")
	ErrInvalidAmount       = errors.New("invalid amount")
	ErrUserNotFound        = errors.New("user not found")
	ErrCannotTransferSelf  = errors.New("cannot transfer to yourself")
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// GetByID returns a transaction by ID
func (s *Service) GetByID(id, userID uuid.UUID, isAdmin bool) (*models.Transaction, error) {
	var tx models.Transaction
	if err := s.db.First(&tx, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTransactionNotFound
		}
		return nil, err
	}

	if !isAdmin && tx.UserID != userID {
		return nil, ErrTransactionNotFound
	}

	return &tx, nil
}

// GetUserTransactions returns all transactions for a user
func (s *Service) GetUserTransactions(userID uuid.UUID, filter *TransactionFilter) ([]models.Transaction, error) {
	var transactions []models.Transaction
	query := s.db.Where("user_id = ?", userID)

	if filter.Type != nil {
		query = query.Where("type = ?", *filter.Type)
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}

	query = query.Order("created_at DESC").Limit(filter.Limit).Offset(filter.Offset)

	if err := query.Find(&transactions).Error; err != nil {
		return nil, err
	}
	return transactions, nil
}

// GetAll returns all transactions (admin only)
func (s *Service) GetAll(filter *TransactionFilter) ([]models.Transaction, error) {
	var transactions []models.Transaction
	query := s.db.Model(&models.Transaction{})

	if filter.Type != nil {
		query = query.Where("type = ?", *filter.Type)
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}

	query = query.Order("created_at DESC").Limit(filter.Limit).Offset(filter.Offset)

	if err := query.Find(&transactions).Error; err != nil {
		return nil, err
	}
	return transactions, nil
}

// Deposit adds funds to user balance
func (s *Service) Deposit(userID uuid.UUID, req *DepositRequest) (*models.Transaction, error) {
	if req.Amount <= 0 {
		return nil, ErrInvalidAmount
	}

	var user models.User
	if err := s.db.First(&user, "id = ?", userID).Error; err != nil {
		return nil, ErrUserNotFound
	}

	var tx *models.Transaction

	err := s.db.Transaction(func(db *gorm.DB) error {
		newBalance := user.Balance + req.Amount

		tx = &models.Transaction{
			ID:          uuid.New(),
			UserID:      userID,
			Type:        models.TransactionTypeDeposit,
			Status:      models.TransactionStatusCompleted,
			Amount:      req.Amount,
			Description: "Deposit",
		}

		if err := db.Create(tx).Error; err != nil {
			return err
		}

		return db.Model(&user).Update("balance", newBalance).Error
	})

	if err != nil {
		return nil, err
	}

	return tx, nil
}

// Withdraw removes funds from user balance
func (s *Service) Withdraw(userID uuid.UUID, req *WithdrawRequest) (*models.Transaction, error) {
	if req.Amount <= 0 {
		return nil, ErrInvalidAmount
	}

	var user models.User
	if err := s.db.First(&user, "id = ?", userID).Error; err != nil {
		return nil, ErrUserNotFound
	}

	if user.Balance < req.Amount {
		return nil, ErrInsufficientFunds
	}

	var tx *models.Transaction

	err := s.db.Transaction(func(db *gorm.DB) error {
		newBalance := user.Balance - req.Amount

		tx = &models.Transaction{
			ID:          uuid.New(),
			UserID:      userID,
			Type:        models.TransactionTypeWithdrawal,
			Status:      models.TransactionStatusCompleted,
			Amount:      -req.Amount,
			Description: "Withdrawal",
		}

		if err := db.Create(tx).Error; err != nil {
			return err
		}

		return db.Model(&user).Update("balance", newBalance).Error
	})

	if err != nil {
		return nil, err
	}

	return tx, nil
}

// Transfer moves funds between users
func (s *Service) Transfer(fromUserID uuid.UUID, req *TransferRequest) (*models.Transaction, error) {
	if req.Amount <= 0 {
		return nil, ErrInvalidAmount
	}

	if fromUserID == req.ToUserID {
		return nil, ErrCannotTransferSelf
	}

	var fromUser, toUser models.User
	if err := s.db.First(&fromUser, "id = ?", fromUserID).Error; err != nil {
		return nil, ErrUserNotFound
	}
	if err := s.db.First(&toUser, "id = ?", req.ToUserID).Error; err != nil {
		return nil, ErrUserNotFound
	}

	if fromUser.Balance < req.Amount {
		return nil, ErrInsufficientFunds
	}

	var tx *models.Transaction

	err := s.db.Transaction(func(db *gorm.DB) error {
		fromNewBalance := fromUser.Balance - req.Amount
		toNewBalance := toUser.Balance + req.Amount

		// Sender transaction
		tx = &models.Transaction{
			ID:            uuid.New(),
			UserID:        fromUserID,
			Type:          models.TransactionTypeTransfer,
			Status:        models.TransactionStatusCompleted,
			Amount:        -req.Amount,
			ReferenceID:   &req.ToUserID,
			ReferenceType: strPtr("user"),
			Description:   "Transfer sent",
		}
		if err := db.Create(tx).Error; err != nil {
			return err
		}

		// Receiver transaction
		receiverTx := &models.Transaction{
			ID:            uuid.New(),
			UserID:        req.ToUserID,
			Type:          models.TransactionTypeTransfer,
			Status:        models.TransactionStatusCompleted,
			Amount:        req.Amount,
			ReferenceID:   &fromUserID,
			ReferenceType: strPtr("user"),
			Description:   "Transfer received",
		}
		if err := db.Create(receiverTx).Error; err != nil {
			return err
		}

		// Update balances
		if err := db.Model(&fromUser).Update("balance", fromNewBalance).Error; err != nil {
			return err
		}
		return db.Model(&toUser).Update("balance", toNewBalance).Error
	})

	if err != nil {
		return nil, err
	}

	return tx, nil
}

// GetUserSummary returns transaction summary for a user
func (s *Service) GetUserSummary(userID uuid.UUID) (*TransactionSummary, error) {
	var summary TransactionSummary

	s.db.Model(&models.Transaction{}).
		Where("user_id = ? AND type = ? AND status = ?", userID, models.TransactionTypeDeposit, models.TransactionStatusCompleted).
		Select("COALESCE(SUM(amount), 0)").Scan(&summary.TotalDeposits)

	s.db.Model(&models.Transaction{}).
		Where("user_id = ? AND type = ? AND status = ?", userID, models.TransactionTypeWithdrawal, models.TransactionStatusCompleted).
		Select("COALESCE(ABS(SUM(amount)), 0)").Scan(&summary.TotalWithdrawals)

	s.db.Model(&models.Transaction{}).
		Where("user_id = ? AND type = ? AND status = ?", userID, models.TransactionTypeBet, models.TransactionStatusCompleted).
		Select("COALESCE(ABS(SUM(amount)), 0)").Scan(&summary.TotalBets)

	s.db.Model(&models.Transaction{}).
		Where("user_id = ? AND type = ? AND status = ?", userID, models.TransactionTypeWin, models.TransactionStatusCompleted).
		Select("COALESCE(SUM(amount), 0)").Scan(&summary.TotalWins)

	summary.NetBalance = summary.TotalDeposits - summary.TotalWithdrawals + summary.TotalWins - summary.TotalBets

	return &summary, nil
}

func strPtr(s string) *string {
	return &s
}
