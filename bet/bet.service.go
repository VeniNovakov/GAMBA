package bet

import (
	"errors"

	"gamba/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrBetNotFound = errors.New("bet not found")
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

func (s *Service) GetByID(id, userID uuid.UUID, isAdmin bool) (*models.Bet, error) {
	var bet models.Bet
	if err := s.db.First(&bet, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrBetNotFound
		}
		return nil, err
	}

	if !isAdmin && bet.UserID != userID {
		return nil, ErrBetNotFound
	}

	return &bet, nil
}

func (s *Service) GetUserBets(userID uuid.UUID, filter *BetFilter) ([]models.Bet, error) {
	var bets []models.Bet
	query := s.db.Where("user_id = ?", userID)

	if filter.Type != nil {
		query = query.Where("type = ?", *filter.Type)
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}

	query = query.Order("created_at DESC").Limit(filter.Limit).Offset(filter.Offset)

	if err := query.Find(&bets).Error; err != nil {
		return nil, err
	}
	return bets, nil
}

func (s *Service) GetAll(filter *BetFilter) ([]models.Bet, error) {
	var bets []models.Bet
	query := s.db.Model(&models.Bet{})

	if filter.Type != nil {
		query = query.Where("type = ?", *filter.Type)
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}

	query = query.Order("created_at DESC").Limit(filter.Limit).Offset(filter.Offset)

	if err := query.Find(&bets).Error; err != nil {
		return nil, err
	}
	return bets, nil
}

// GetUserSummary returns betting summary for a user
func (s *Service) GetUserSummary(userID uuid.UUID) (*BetSummary, error) {
	var summary BetSummary

	// Total bets
	s.db.Model(&models.Bet{}).Where("user_id = ?", userID).Count(&summary.TotalBets)

	// Total wagered
	s.db.Model(&models.Bet{}).Where("user_id = ?", userID).Select("COALESCE(SUM(amount), 0)").Scan(&summary.TotalWagered)

	// Total won (payout from winning bets)
	s.db.Model(&models.Bet{}).Where("user_id = ? AND status = ?", userID, models.BetStatusWon).Select("COALESCE(SUM(payout), 0)").Scan(&summary.TotalWon)

	// Total lost (amount from losing bets)
	s.db.Model(&models.Bet{}).Where("user_id = ? AND status = ?", userID, models.BetStatusLost).Select("COALESCE(SUM(amount), 0)").Scan(&summary.TotalLost)

	// Win rate
	var wonCount int64
	s.db.Model(&models.Bet{}).Where("user_id = ? AND status = ?", userID, models.BetStatusWon).Count(&wonCount)
	if summary.TotalBets > 0 {
		summary.WinRate = float64(wonCount) / float64(summary.TotalBets) * 100
	}

	return &summary, nil
}
