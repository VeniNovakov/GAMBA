package user

import (
	"errors"
	"strings"

	"gamba/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrUsernameExists   = errors.New("username already exists")
	ErrInvalidPassword  = errors.New("invalid password")
	ErrCannotFriendSelf = errors.New("cannot friend yourself")
	ErrAlreadyFriends   = errors.New("already friends or request pending")
	ErrRequestNotFound  = errors.New("friend request not found")
	ErrNotFriends       = errors.New("not friends")
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

func (s *Service) SendFriendRequest(userID, friendID uuid.UUID) (*models.Friend, error) {
	if userID == friendID {
		return nil, ErrCannotFriendSelf
	}

	var friend models.User
	if err := s.db.First(&friend, "id = ?", friendID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	var existing models.Friend
	err := s.db.Where(
		"(user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)",
		userID, friendID, friendID, userID,
	).First(&existing).Error

	if err == nil {
		return nil, ErrAlreadyFriends
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	request := models.Friend{
		ID:       uuid.New(),
		UserID:   userID,
		FriendID: friendID,
		Status:   models.FriendStatusPending,
	}

	if err := s.db.Create(&request).Error; err != nil {
		return nil, err
	}

	return &request, nil
}

func (s *Service) AcceptFriendRequest(userID, requestID uuid.UUID) error {
	var request models.Friend
	if err := s.db.First(&request, "id = ?", requestID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRequestNotFound
		}
		return err
	}

	if request.FriendID != userID {
		return ErrRequestNotFound
	}

	if request.Status != models.FriendStatusPending {
		return ErrRequestNotFound
	}

	return s.db.Model(&request).Update("status", models.FriendStatusAccepted).Error
}

func (s *Service) RejectFriendRequest(userID, requestID uuid.UUID) error {
	var request models.Friend
	if err := s.db.First(&request, "id = ?", requestID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRequestNotFound
		}
		return err
	}

	if request.FriendID != userID {
		return ErrRequestNotFound
	}

	return s.db.Model(&request).Update("status", models.FriendStatusRejected).Error
}

func (s *Service) RemoveFriend(userID, friendID uuid.UUID) error {
	result := s.db.Where(
		"((user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)) AND status = ?",
		userID, friendID, friendID, userID, models.FriendStatusAccepted,
	).Delete(&models.Friend{})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFriends
	}
	return nil
}

func (s *Service) GetFriends(userID uuid.UUID) ([]models.Friend, error) {
	var friends []models.Friend
	err := s.db.
		Preload("User").
		Preload("Friend").
		Where(
			"(user_id = ? OR friend_id = ?) AND status = ?",
			userID, userID, models.FriendStatusAccepted,
		).Find(&friends).Error

	return friends, err
}

func (s *Service) GetPendingRequests(userID uuid.UUID) ([]models.Friend, error) {
	var requests []models.Friend
	err := s.db.
		Preload("User").
		Where("friend_id = ? AND status = ?", userID, models.FriendStatusPending).
		Find(&requests).Error

	return requests, err
}

func (s *Service) GetSentRequests(userID uuid.UUID) ([]models.Friend, error) {
	var requests []models.Friend
	err := s.db.
		Preload("Friend").
		Where("user_id = ? AND status = ?", userID, models.FriendStatusPending).
		Find(&requests).Error

	return requests, err
}
