package tournament

import (
	"errors"
	"sort"

	"gamba/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrTournamentNotFound  = errors.New("tournament not found")
	ErrTournamentFull      = errors.New("tournament is full")
	ErrTournamentNotOpen   = errors.New("tournament is not open for registration")
	ErrAlreadyJoined       = errors.New("already joined this tournament")
	ErrInsufficientFunds   = errors.New("insufficient funds for entry fee")
	ErrNotParticipant      = errors.New("user is not a participant")
	ErrTournamentNotActive = errors.New("tournament is not active")
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// GetAll returns tournaments with optional filters
func (s *Service) GetAll(filter *TournamentFilter) ([]models.Tournament, error) {
	var tournaments []models.Tournament
	query := s.db.Preload("Participants")

	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	if filter.GameID != nil {
		query = query.Where("game_id = ?", *filter.GameID)
	}

	query = query.Order("starts_at DESC").Limit(filter.Limit).Offset(filter.Offset)

	if err := query.Find(&tournaments).Error; err != nil {
		return nil, err
	}
	return tournaments, nil
}

// GetByID returns a tournament by ID
func (s *Service) GetByID(id uuid.UUID) (*models.Tournament, error) {
	var tournament models.Tournament
	if err := s.db.Preload("Participants").First(&tournament, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTournamentNotFound
		}
		return nil, err
	}
	return &tournament, nil
}

// Create creates a new tournament (admin only)
func (s *Service) Create(req *CreateRequest) (*models.Tournament, error) {
	tournament := models.Tournament{
		ID:              uuid.New(),
		Name:            req.Name,
		Description:     req.Description,
		Status:          models.TournamentStatusDraft,
		GameID:          req.GameID,
		EntryFee:        req.EntryFee,
		PrizePool:       req.PrizePool,
		MaxParticipants: req.MaxParticipants,
		StartsAt:        req.StartsAt,
		EndsAt:          req.EndsAt,
	}

	if err := s.db.Create(&tournament).Error; err != nil {
		return nil, err
	}
	return &tournament, nil
}

// Update updates a tournament (admin only)
func (s *Service) Update(id uuid.UUID, req *UpdateRequest) (*models.Tournament, error) {
	var tournament models.Tournament
	if err := s.db.First(&tournament, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTournamentNotFound
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
	if req.GameID != nil {
		updates["game_id"] = *req.GameID
	}
	if req.EntryFee != nil {
		updates["entry_fee"] = *req.EntryFee
	}
	if req.PrizePool != nil {
		updates["prize_pool"] = *req.PrizePool
	}
	if req.MaxParticipants != nil {
		updates["max_participants"] = *req.MaxParticipants
	}
	if req.StartsAt != nil {
		updates["starts_at"] = *req.StartsAt
	}
	if req.EndsAt != nil {
		updates["ends_at"] = *req.EndsAt
	}

	if len(updates) > 0 {
		if err := s.db.Model(&tournament).Updates(updates).Error; err != nil {
			return nil, err
		}
	}

	return &tournament, nil
}

// Delete deletes a tournament (admin only)
func (s *Service) Delete(id uuid.UUID) error {
	result := s.db.Delete(&models.Tournament{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrTournamentNotFound
	}
	return nil
}

// Join allows a user to join a tournament
func (s *Service) Join(userID uuid.UUID, req *JoinRequest) (*models.TournamentParticipant, error) {
	var tournament models.Tournament
	if err := s.db.Preload("Participants").First(&tournament, "id = ?", req.TournamentID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTournamentNotFound
		}
		return nil, err
	}

	// Check if tournament is open
	if tournament.Status != models.TournamentStatusOpen {
		return nil, ErrTournamentNotOpen
	}

	// Check if full
	if len(tournament.Participants) >= tournament.MaxParticipants {
		return nil, ErrTournamentFull
	}

	// Check if already joined
	for _, p := range tournament.Participants {
		if p.UserID == userID {
			return nil, ErrAlreadyJoined
		}
	}

	// Get user for balance check
	var user models.User
	if err := s.db.First(&user, "id = ?", userID).Error; err != nil {
		return nil, err
	}

	if user.Balance < tournament.EntryFee {
		return nil, ErrInsufficientFunds
	}

	var participant *models.TournamentParticipant

	err := s.db.Transaction(func(tx *gorm.DB) error {
		// Deduct entry fee
		if tournament.EntryFee > 0 {
			newBalance := user.Balance - tournament.EntryFee
			if err := tx.Model(&user).Update("balance", newBalance).Error; err != nil {
				return err
			}

			// Create transaction record
			transaction := models.Transaction{
				ID:            uuid.New(),
				UserID:        userID,
				Type:          models.TransactionTypeTournamentEntry,
				Status:        models.TransactionStatusCompleted,
				Amount:        -tournament.EntryFee,
				ReferenceID:   &tournament.ID,
				ReferenceType: strPtr("tournament"),
				Description:   "Tournament entry: " + tournament.Name,
			}
			if err := tx.Create(&transaction).Error; err != nil {
				return err
			}
		}

		// Create participant
		participant = &models.TournamentParticipant{
			ID:           uuid.New(),
			TournamentID: tournament.ID,
			UserID:       userID,
			Score:        0,
			Rank:         0,
			PrizeWon:     0,
		}
		return tx.Create(participant).Error
	})

	if err != nil {
		return nil, err
	}

	return participant, nil
}

// Leave allows a user to leave a tournament (only if not started)
func (s *Service) Leave(userID uuid.UUID, tournamentID uuid.UUID) error {
	var tournament models.Tournament
	if err := s.db.First(&tournament, "id = ?", tournamentID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTournamentNotFound
		}
		return err
	}

	// Can only leave if tournament hasn't started
	if tournament.Status != models.TournamentStatusOpen && tournament.Status != models.TournamentStatusDraft {
		return ErrTournamentNotOpen
	}

	var participant models.TournamentParticipant
	if err := s.db.Where("tournament_id = ? AND user_id = ?", tournamentID, userID).First(&participant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotParticipant
		}
		return err
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		// Refund entry fee
		if tournament.EntryFee > 0 {
			var user models.User
			if err := tx.First(&user, "id = ?", userID).Error; err != nil {
				return err
			}

			newBalance := user.Balance + tournament.EntryFee
			if err := tx.Model(&user).Update("balance", newBalance).Error; err != nil {
				return err
			}

			// Create refund transaction
			transaction := models.Transaction{
				ID:            uuid.New(),
				UserID:        userID,
				Type:          models.TransactionTypeRefund,
				Status:        models.TransactionStatusCompleted,
				Amount:        tournament.EntryFee,
				ReferenceID:   &tournament.ID,
				ReferenceType: strPtr("tournament"),
				Description:   "Tournament refund: " + tournament.Name,
			}
			if err := tx.Create(&transaction).Error; err != nil {
				return err
			}
		}

		// Remove participant
		return tx.Delete(&participant).Error
	})
}

// UpdateScore updates a participant's score (admin only or game service)
func (s *Service) UpdateScore(tournamentID uuid.UUID, req *UpdateScoreRequest) error {
	var tournament models.Tournament
	if err := s.db.First(&tournament, "id = ?", tournamentID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTournamentNotFound
		}
		return err
	}

	if tournament.Status != models.TournamentStatusInProgress {
		return ErrTournamentNotActive
	}

	result := s.db.Model(&models.TournamentParticipant{}).
		Where("tournament_id = ? AND user_id = ?", tournamentID, req.UserID).
		Update("score", gorm.Expr("score + ?", req.Score))

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotParticipant
	}

	return nil
}

// GetLeaderboard returns the tournament leaderboard
func (s *Service) GetLeaderboard(tournamentID uuid.UUID) ([]LeaderboardEntry, error) {
	var tournament models.Tournament
	if err := s.db.Preload("Participants").First(&tournament, "id = ?", tournamentID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTournamentNotFound
		}
		return nil, err
	}

	// Get user names
	userIDs := make([]uuid.UUID, len(tournament.Participants))
	for i, p := range tournament.Participants {
		userIDs[i] = p.UserID
	}

	var users []models.User
	s.db.Where("id IN ?", userIDs).Find(&users)

	userMap := make(map[uuid.UUID]string)
	for _, u := range users {
		userMap[u.ID] = u.Username
	}

	// Sort by score descending
	participants := tournament.Participants
	sort.Slice(participants, func(i, j int) bool {
		return participants[i].Score > participants[j].Score
	})

	leaderboard := make([]LeaderboardEntry, len(participants))
	for i, p := range participants {
		leaderboard[i] = LeaderboardEntry{
			Rank:     i + 1,
			UserID:   p.UserID,
			UserName: userMap[p.UserID],
			Score:    p.Score,
			PrizeWon: p.PrizeWon,
		}
	}

	return leaderboard, nil
}

// EndTournament ends tournament and distributes prizes (admin only)
func (s *Service) EndTournament(tournamentID uuid.UUID) error {
	var tournament models.Tournament
	if err := s.db.Preload("Participants").First(&tournament, "id = ?", tournamentID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTournamentNotFound
		}
		return err
	}

	if tournament.Status != models.TournamentStatusInProgress {
		return ErrTournamentNotActive
	}

	// Sort participants by score
	participants := tournament.Participants
	sort.Slice(participants, func(i, j int) bool {
		return participants[i].Score > participants[j].Score
	})

	// Prize distribution (50% / 30% / 20% for top 3)
	prizeDistribution := []float64{0.5, 0.3, 0.2}

	return s.db.Transaction(func(tx *gorm.DB) error {
		for i, p := range participants {
			if i >= len(prizeDistribution) {
				break
			}

			prize := tournament.PrizePool * prizeDistribution[i]
			if prize <= 0 {
				continue
			}

			// Update participant prize
			if err := tx.Model(&p).Updates(map[string]interface{}{
				"rank":      i + 1,
				"prize_won": prize,
			}).Error; err != nil {
				return err
			}

			// Add prize to user balance
			var user models.User
			if err := tx.First(&user, "id = ?", p.UserID).Error; err != nil {
				return err
			}

			newBalance := user.Balance + prize
			if err := tx.Model(&user).Update("balance", newBalance).Error; err != nil {
				return err
			}

			// Create prize transaction
			transaction := models.Transaction{
				ID:            uuid.New(),
				UserID:        p.UserID,
				Type:          models.TransactionTypeTournamentPrize,
				Status:        models.TransactionStatusCompleted,
				Amount:        prize,
				ReferenceID:   &tournament.ID,
				ReferenceType: strPtr("tournament"),
				Description:   "Tournament prize: " + tournament.Name,
			}
			if err := tx.Create(&transaction).Error; err != nil {
				return err
			}
		}

		// Update remaining participant ranks
		for i := len(prizeDistribution); i < len(participants); i++ {
			tx.Model(&participants[i]).Update("rank", i+1)
		}

		// Mark tournament as completed
		return tx.Model(&tournament).Update("status", models.TournamentStatusCompleted).Error
	})
}

func strPtr(s string) *string {
	return &s
}
