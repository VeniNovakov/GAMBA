package user

import (
	"errors"
	"strings"

	"gamba/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrUsernameExists  = errors.New("username already exists")
	ErrInvalidPassword = errors.New("invalid password")
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

func (s *Service) GetProfile(userID uuid.UUID) (*models.User, error) {
	var user models.User
	if err := s.db.First(&user, "id = ?", userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (s *Service) UpdateProfile(userID uuid.UUID, req *UpdateProfileRequest) (*models.User, error) {
	var user models.User
	if err := s.db.First(&user, "id = ?", userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	updates := make(map[string]interface{})
	if req.Username != nil {
		updates["username"] = *req.Username
	}

	if len(updates) > 0 {
		if err := s.db.Model(&user).Updates(updates).Error; err != nil {
			if strings.Contains(err.Error(), "duplicate") {
				return nil, ErrUsernameExists
			}
			return nil, err
		}
	}

	return &user, nil
}

func (s *Service) GetByID(userID uuid.UUID) (*models.User, error) {
	var user models.User
	if err := s.db.First(&user, "id = ?", userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (s *Service) GetByUsername(username string) (*models.User, error) {
	var user models.User
	if err := s.db.First(&user, "username = ?", username).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (s *Service) Search(query string, limit int) ([]models.User, error) {
	var users []models.User
	if limit == 0 {
		limit = 10
	}

	err := s.db.Where("username ILIKE ?", "%"+query+"%").
		Limit(limit).
		Find(&users).Error

	return users, err
}

func (s *Service) GetAll(filter *UserFilter) ([]models.User, error) {
	var users []models.User
	query := s.db.Model(&models.User{})

	if filter.Username != nil {
		query = query.Where("username ILIKE ?", "%"+*filter.Username+"%")
	}
	if filter.Role != nil {
		query = query.Where("role = ?", *filter.Role)
	}
	if filter.IsActive != nil {
		query = query.Where("is_active = ?", *filter.IsActive)
	}

	if filter.Limit == 0 {
		filter.Limit = 20
	}

	err := query.Order("created_at DESC").
		Limit(filter.Limit).
		Offset(filter.Offset).
		Find(&users).Error

	return users, err
}

func (s *Service) SetRestricted(userID uuid.UUID, restricted bool) error {
	result := s.db.Model(&models.User{}).Where("id = ?", userID).Update("is_restricted", restricted)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (s *Service) SetActive(userID uuid.UUID, active bool) error {
	result := s.db.Model(&models.User{}).Where("id = ?", userID).Update("is_active", active)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}
