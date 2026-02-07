package ticket

import (
	"errors"

	"gamba/models" // adjust import path

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrTicketNotFound = errors.New("ticket not found")
	ErrUnauthorized   = errors.New("unauthorized")
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// GetAll returns all tickets (admin) or user's tickets
func (s *Service) GetAll(userID uuid.UUID, isAdmin bool) ([]models.Ticket, error) {
	var tickets []models.Ticket
	query := s.db.Preload("Messages").Preload("User").Preload("Assignee")

	if !isAdmin {
		query = query.Where("user_id = ?", userID)
	}

	if err := query.Find(&tickets).Error; err != nil {
		return nil, err
	}
	return tickets, nil
}

// GetByID returns a ticket by ID
func (s *Service) GetByID(id, userID uuid.UUID, isAdmin bool) (*models.Ticket, error) {
	var ticket models.Ticket
	if err := s.db.Preload("Messages.Sender").Preload("User").Preload("Assignee").First(&ticket, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTicketNotFound
		}
		return nil, err
	}

	// Check authorization
	if !isAdmin && ticket.UserID != userID {
		return nil, ErrUnauthorized
	}

	return &ticket, nil
}

// Create creates a new ticket
func (s *Service) Create(userID uuid.UUID, req *CreateRequest) (*models.Ticket, error) {
	priority := models.TicketPriority(req.Priority)
	if req.Priority == "" {
		priority = models.TicketPriorityMedium
	}

	ticket := models.Ticket{
		ID:          uuid.New(),
		UserID:      userID,
		Subject:     req.Subject,
		Description: req.Description,
		Status:      models.TicketStatusOpen,
		Priority:    priority,
	}

	if err := s.db.Create(&ticket).Error; err != nil {
		return nil, err
	}
	return &ticket, nil
}

// Update updates a ticket (admin only for status/assignment, user can update priority)
func (s *Service) Update(id, userID uuid.UUID, isAdmin bool, req *UpdateRequest) (*models.Ticket, error) {
	var ticket models.Ticket
	if err := s.db.First(&ticket, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTicketNotFound
		}
		return nil, err
	}

	// Check authorization
	if !isAdmin && ticket.UserID != userID {
		return nil, ErrUnauthorized
	}

	updates := make(map[string]interface{})

	// Admin-only fields
	if isAdmin {
		if req.Status != nil {
			updates["status"] = *req.Status
		}
		if req.AssignedTo != nil {
			updates["assigned_to"] = *req.AssignedTo
		}
	}

	// User can update priority
	if req.Priority != nil {
		updates["priority"] = *req.Priority
	}

	if len(updates) > 0 {
		if err := s.db.Model(&ticket).Updates(updates).Error; err != nil {
			return nil, err
		}
	}

	return &ticket, nil
}

// AddMessage adds a message to a ticket
func (s *Service) AddMessage(ticketID, senderID uuid.UUID, isAdmin bool, req *AddMessageRequest) (*models.TicketMessage, error) {
	var ticket models.Ticket
	if err := s.db.First(&ticket, "id = ?", ticketID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTicketNotFound
		}
		return nil, err
	}

	// Check authorization
	if !isAdmin && ticket.UserID != senderID {
		return nil, ErrUnauthorized
	}

	message := models.TicketMessage{
		ID:       uuid.New(),
		TicketID: ticketID,
		SenderID: senderID,
		Content:  req.Content,
	}

	if err := s.db.Create(&message).Error; err != nil {
		return nil, err
	}

	// Update ticket status if user replies to resolved ticket
	if !isAdmin && ticket.Status == models.TicketStatusResolved {
		s.db.Model(&ticket).Update("status", models.TicketStatusOpen)
	}

	return &message, nil
}

// Close closes a ticket
func (s *Service) Close(id, userID uuid.UUID, isAdmin bool) error {
	var ticket models.Ticket
	if err := s.db.First(&ticket, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTicketNotFound
		}
		return err
	}

	// Check authorization
	if !isAdmin && ticket.UserID != userID {
		return ErrUnauthorized
	}

	return s.db.Model(&ticket).Update("status", models.TicketStatusClosed).Error
}
