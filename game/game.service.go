package game

import (
	"errors"
	"math/rand"
	"time"

	"gamba/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrGameNotFound      = errors.New("game not found")
	ErrGameInactive      = errors.New("game is not active")
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrBetTooLow         = errors.New("bet amount below minimum")
	ErrBetTooHigh        = errors.New("bet amount above maximum")
)

// Slot symbols
var symbols = []string{"🍒", "🍋", "🍊", "🍇", "🔔", "⭐", "7️⃣"}

// Symbol multipliers (index matches symbols array)
var multipliers = []float64{2, 3, 4, 5, 10, 20, 50}

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	rand.Seed(time.Now().UnixNano())
	return &Service{db: db}
}

func (s *Service) GetAll() ([]models.Game, error) {
	var games []models.Game
	if err := s.db.Where("status = ?", models.GameStatusActive).Find(&games).Error; err != nil {
		return nil, err
	}
	return games, nil
}

func (s *Service) GetByID(id uuid.UUID) (*models.Game, error) {
	var game models.Game
	if err := s.db.First(&game, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrGameNotFound
		}
		return nil, err
	}
	return &game, nil
}

func (s *Service) Create(req *CreateRequest) (*models.Game, error) {
	game := models.Game{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		Category:    "slots",
		Status:      models.GameStatusActive,
		MinBet:      req.MinBet,
		MaxBet:      req.MaxBet,
		HouseEdge:   req.HouseEdge,
	}

	if err := s.db.Create(&game).Error; err != nil {
		return nil, err
	}
	return &game, nil
}

func (s *Service) Update(id uuid.UUID, req *UpdateRequest) (*models.Game, error) {
	var game models.Game
	if err := s.db.First(&game, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrGameNotFound
		}
		return nil, err
	}

	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.MinBet != nil {
		updates["min_bet"] = *req.MinBet
	}
	if req.MaxBet != nil {
		updates["max_bet"] = *req.MaxBet
	}
	if req.HouseEdge != nil {
		updates["house_edge"] = *req.HouseEdge
	}

	if len(updates) > 0 {
		if err := s.db.Model(&game).Updates(updates).Error; err != nil {
			return nil, err
		}
	}

	return &game, nil
}

func (s *Service) Delete(id uuid.UUID) error {
	result := s.db.Delete(&models.Game{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrGameNotFound
	}
	return nil
}

// Play spins the slot machine
func (s *Service) Play(userID uuid.UUID, req *PlayRequest) (*PlayResponse, error) {
	// Get game
	var game models.Game
	if err := s.db.First(&game, "id = ?", req.GameID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrGameNotFound
		}
		return nil, err
	}

	if game.Status != models.GameStatusActive {
		return nil, ErrGameInactive
	}

	// Validate bet
	if req.BetAmount < game.MinBet {
		return nil, ErrBetTooLow
	}
	if req.BetAmount > game.MaxBet {
		return nil, ErrBetTooHigh
	}

	// Get user
	var user models.User
	if err := s.db.First(&user, "id = ?", userID).Error; err != nil {
		return nil, err
	}

	if user.Balance < req.BetAmount {
		return nil, ErrInsufficientFunds
	}

	// Spin the reels
	reels := s.spin()

	// Calculate result
	won, multiplier := s.calculateWin(reels)
	payout := 0.0
	if won {
		payout = req.BetAmount * multiplier
	}

	// Update balance in transaction
	newBalance := user.Balance - req.BetAmount + payout

	err := s.db.Transaction(func(tx *gorm.DB) error {
		// Update user balance
		if err := tx.Model(&user).Update("balance", newBalance).Error; err != nil {
			return err
		}

		// Create bet record
		bet := models.Bet{
			ID:        uuid.New(),
			UserID:    userID,
			Type:      models.BetTypeGame,
			GameID:    &game.ID,
			Amount:    req.BetAmount,
			Odds:      multiplier,
			Status:    models.BetStatusLost,
			Payout:    payout,
			SettledAt: timePtr(time.Now()),
		}
		if won {
			bet.Status = models.BetStatusWon
		}
		if err := tx.Create(&bet).Error; err != nil {
			return err
		}

		// Create transaction record for bet
		betTx := models.Transaction{
			ID:            uuid.New(),
			UserID:        userID,
			Type:          models.TransactionTypeBet,
			Status:        models.TransactionStatusCompleted,
			Amount:        -req.BetAmount,
			ReferenceID:   &bet.ID,
			ReferenceType: strPtr("bet"),
			Description:   "Slot machine bet",
		}
		if err := tx.Create(&betTx).Error; err != nil {
			return err
		}

		// Create transaction record for win
		if won {
			winTx := models.Transaction{
				ID:            uuid.New(),
				UserID:        userID,
				Type:          models.TransactionTypeWin,
				Status:        models.TransactionStatusCompleted,
				Amount:        payout,
				ReferenceID:   &bet.ID,
				ReferenceType: strPtr("bet"),
				Description:   "Slot machine win",
			}
			if err := tx.Create(&winTx).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &PlayResponse{
		Reels:      reels,
		Won:        won,
		Payout:     payout,
		Multiplier: multiplier,
		NewBalance: newBalance,
	}, nil
}

// spin generates 3 random symbols
func (s *Service) spin() [3]string {
	return [3]string{
		symbols[rand.Intn(len(symbols))],
		symbols[rand.Intn(len(symbols))],
		symbols[rand.Intn(len(symbols))],
	}
}

// calculateWin checks if reels match and returns multiplier
func (s *Service) calculateWin(reels [3]string) (bool, float64) {
	// All three match
	if reels[0] == reels[1] && reels[1] == reels[2] {
		for i, sym := range symbols {
			if sym == reels[0] {
				return true, multipliers[i]
			}
		}
	}

	// Two match (smaller payout)
	if reels[0] == reels[1] || reels[1] == reels[2] || reels[0] == reels[2] {
		var matchSymbol string
		if reels[0] == reels[1] {
			matchSymbol = reels[0]
		} else if reels[1] == reels[2] {
			matchSymbol = reels[1]
		} else {
			matchSymbol = reels[0]
		}

		for i, sym := range symbols {
			if sym == matchSymbol {
				return true, multipliers[i] * 0.25 // 25% of full multiplier for 2 match
			}
		}
	}

	return false, 0
}

func timePtr(t time.Time) *time.Time {
	return &t
}

func strPtr(s string) *string {
	return &s
}
