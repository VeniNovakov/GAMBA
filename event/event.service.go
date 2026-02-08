package event

import (
	"errors"
	"time"

	"gamba/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrEventNotFound       = errors.New("event not found")
	ErrOutcomeNotFound     = errors.New("outcome not found")
	ErrEventNotBettable    = errors.New("event is not accepting bets")
	ErrEventAlreadySettled = errors.New("event is already settled")
	ErrInsufficientFunds   = errors.New("insufficient funds")
	ErrInvalidAmount       = errors.New("invalid bet amount")
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// GetAll returns events with optional filters
func (s *Service) GetAll(filter *EventFilter) ([]models.Event, error) {
	var events []models.Event
	query := s.db.Preload("Outcomes")

	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	if filter.Category != nil {
		query = query.Where("category = ?", *filter.Category)
	}

	query = query.Order("starts_at DESC").Limit(filter.Limit).Offset(filter.Offset)

	if err := query.Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}

// GetByID returns an event by ID
func (s *Service) GetByID(id uuid.UUID) (*models.Event, error) {
	var event models.Event
	if err := s.db.Preload("Outcomes").First(&event, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrEventNotFound
		}
		return nil, err
	}
	return &event, nil
}

// Create creates a new event (admin only)
func (s *Service) Create(req *CreateRequest) (*models.Event, error) {
	event := models.Event{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		Status:      models.EventStatusUpcoming,
		StartsAt:    req.StartsAt,
		EndsAt:      req.EndsAt,
	}

	if err := s.db.Create(&event).Error; err != nil {
		return nil, err
	}
	return &event, nil
}

// Update updates an event (admin only)
func (s *Service) Update(id uuid.UUID, req *UpdateRequest) (*models.Event, error) {
	var event models.Event
	if err := s.db.First(&event, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrEventNotFound
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
	if req.Category != nil {
		updates["category"] = *req.Category
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.StartsAt != nil {
		updates["starts_at"] = *req.StartsAt
	}
	if req.EndsAt != nil {
		updates["ends_at"] = *req.EndsAt
	}

	if len(updates) > 0 {
		if err := s.db.Model(&event).Updates(updates).Error; err != nil {
			return nil, err
		}
	}

	return &event, nil
}

// Delete deletes an event (admin only)
func (s *Service) Delete(id uuid.UUID) error {
	result := s.db.Delete(&models.Event{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrEventNotFound
	}
	return nil
}

// AddOutcome adds an outcome to an event (admin only)
func (s *Service) AddOutcome(eventID uuid.UUID, req *CreateOutcomeRequest) (*models.EventOutcome, error) {
	var event models.Event
	if err := s.db.First(&event, "id = ?", eventID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrEventNotFound
		}
		return nil, err
	}

	outcome := models.EventOutcome{
		ID:      uuid.New(),
		EventID: eventID,
		Name:    req.Name,
		Odds:    req.Odds,
	}

	if err := s.db.Create(&outcome).Error; err != nil {
		return nil, err
	}
	return &outcome, nil
}

// UpdateOutcome updates an outcome (admin only)
func (s *Service) UpdateOutcome(outcomeID uuid.UUID, req *UpdateOutcomeRequest) (*models.EventOutcome, error) {
	var outcome models.EventOutcome
	if err := s.db.First(&outcome, "id = ?", outcomeID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOutcomeNotFound
		}
		return nil, err
	}

	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Odds != nil {
		updates["odds"] = *req.Odds
	}
	if req.IsWinner != nil {
		updates["is_winner"] = *req.IsWinner
	}

	if len(updates) > 0 {
		if err := s.db.Model(&outcome).Updates(updates).Error; err != nil {
			return nil, err
		}
	}

	return &outcome, nil
}

// DeleteOutcome deletes an outcome (admin only)
func (s *Service) DeleteOutcome(outcomeID uuid.UUID) error {
	result := s.db.Delete(&models.EventOutcome{}, "id = ?", outcomeID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrOutcomeNotFound
	}
	return nil
}

// PlaceBet places a bet on an event outcome
func (s *Service) PlaceBet(userID uuid.UUID, eventID uuid.UUID, req *PlaceBetRequest) (*models.Bet, error) {
	if req.Amount <= 0 {
		return nil, ErrInvalidAmount
	}

	// Get event
	var event models.Event
	if err := s.db.First(&event, "id = ?", eventID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrEventNotFound
		}
		return nil, err
	}

	// Check if event is bettable
	if event.Status != models.EventStatusUpcoming && event.Status != models.EventStatusLive {
		return nil, ErrEventNotBettable
	}

	// Get outcome
	var outcome models.EventOutcome
	if err := s.db.First(&outcome, "id = ? AND event_id = ?", req.OutcomeID, eventID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOutcomeNotFound
		}
		return nil, err
	}

	// Get user
	var user models.User
	if err := s.db.First(&user, "id = ?", userID).Error; err != nil {
		return nil, err
	}

	if user.Balance < req.Amount {
		return nil, ErrInsufficientFunds
	}

	var bet *models.Bet

	err := s.db.Transaction(func(tx *gorm.DB) error {
		newBalance := user.Balance - req.Amount

		// Update user balance
		if err := tx.Model(&user).Update("balance", newBalance).Error; err != nil {
			return err
		}

		// Create bet
		bet = &models.Bet{
			ID:        uuid.New(),
			UserID:    userID,
			Type:      models.BetTypeEvent,
			EventID:   &eventID,
			OutcomeID: &req.OutcomeID,
			Amount:    req.Amount,
			Odds:      outcome.Odds,
			Status:    models.BetStatusPending,
		}
		if err := tx.Create(bet).Error; err != nil {
			return err
		}

		// Create transaction
		transaction := models.Transaction{
			ID:            uuid.New(),
			UserID:        userID,
			Type:          models.TransactionTypeBet,
			Status:        models.TransactionStatusCompleted,
			Amount:        -req.Amount,
			ReferenceID:   &bet.ID,
			ReferenceType: strPtr("bet"),
			Description:   "Event bet: " + event.Name,
		}
		return tx.Create(&transaction).Error
	})

	if err != nil {
		return nil, err
	}

	return bet, nil
}

// Settle settles an event and pays out winning bets (admin only)
func (s *Service) Settle(eventID uuid.UUID, req *SettleRequest) error {
	var event models.Event
	if err := s.db.First(&event, "id = ?", eventID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrEventNotFound
		}
		return err
	}

	if event.Status == models.EventStatusCompleted || event.Status == models.EventStatusCancelled {
		return ErrEventAlreadySettled
	}

	// Verify outcome belongs to event
	var winningOutcome models.EventOutcome
	if err := s.db.First(&winningOutcome, "id = ? AND event_id = ?", req.WinningOutcomeID, eventID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrOutcomeNotFound
		}
		return err
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		// Mark winning outcome
		if err := tx.Model(&winningOutcome).Update("is_winner", true).Error; err != nil {
			return err
		}

		// Get all pending bets for this event
		var bets []models.Bet
		if err := tx.Where("event_id = ? AND status = ?", eventID, models.BetStatusPending).Find(&bets).Error; err != nil {
			return err
		}

		now := time.Now()

		for _, bet := range bets {
			if bet.OutcomeID != nil && *bet.OutcomeID == req.WinningOutcomeID {
				// Winner
				payout := bet.Amount * bet.Odds

				// Update bet
				if err := tx.Model(&bet).Updates(map[string]interface{}{
					"status":     models.BetStatusWon,
					"payout":     payout,
					"settled_at": now,
				}).Error; err != nil {
					return err
				}

				// Update user balance
				var user models.User
				if err := tx.First(&user, "id = ?", bet.UserID).Error; err != nil {
					return err
				}

				newBalance := user.Balance + payout
				if err := tx.Model(&user).Update("balance", newBalance).Error; err != nil {
					return err
				}

				// Create win transaction
				winTx := models.Transaction{
					ID:            uuid.New(),
					UserID:        bet.UserID,
					Type:          models.TransactionTypeWin,
					Status:        models.TransactionStatusCompleted,
					Amount:        payout,
					ReferenceID:   &bet.ID,
					ReferenceType: strPtr("bet"),
					Description:   "Event win: " + event.Name,
				}
				if err := tx.Create(&winTx).Error; err != nil {
					return err
				}
			} else {
				// Loser
				if err := tx.Model(&bet).Updates(map[string]interface{}{
					"status":     models.BetStatusLost,
					"settled_at": now,
				}).Error; err != nil {
					return err
				}
			}
		}

		// Mark event as completed
		return tx.Model(&event).Update("status", models.EventStatusCompleted).Error
	})
}

// Cancel cancels an event and refunds all bets (admin only)
func (s *Service) Cancel(eventID uuid.UUID) error {
	var event models.Event
	if err := s.db.First(&event, "id = ?", eventID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrEventNotFound
		}
		return err
	}

	if event.Status == models.EventStatusCompleted || event.Status == models.EventStatusCancelled {
		return ErrEventAlreadySettled
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		// Get all pending bets
		var bets []models.Bet
		if err := tx.Where("event_id = ? AND status = ?", eventID, models.BetStatusPending).Find(&bets).Error; err != nil {
			return err
		}

		now := time.Now()

		for _, bet := range bets {
			// Update bet
			if err := tx.Model(&bet).Updates(map[string]interface{}{
				"status":     models.BetStatusRefunded,
				"settled_at": now,
			}).Error; err != nil {
				return err
			}

			// Refund user
			var user models.User
			if err := tx.First(&user, "id = ?", bet.UserID).Error; err != nil {
				return err
			}

			newBalance := user.Balance + bet.Amount
			if err := tx.Model(&user).Update("balance", newBalance).Error; err != nil {
				return err
			}

			// Create refund transaction
			refundTx := models.Transaction{
				ID:            uuid.New(),
				UserID:        bet.UserID,
				Type:          models.TransactionTypeRefund,
				Status:        models.TransactionStatusCompleted,
				Amount:        bet.Amount,
				ReferenceID:   &bet.ID,
				ReferenceType: strPtr("bet"),
				Description:   "Event cancelled: " + event.Name,
			}
			if err := tx.Create(&refundTx).Error; err != nil {
				return err
			}
		}

		// Mark event as cancelled
		return tx.Model(&event).Update("status", models.EventStatusCancelled).Error
	})
}

func strPtr(s string) *string {
	return &s
}
