package chat

import (
	"errors"
	"gamba/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrChatNotFound      = errors.New("chat not found")
	ErrMessageNotFound   = errors.New("message not found")
	ErrUserNotFound      = errors.New("user not found")
	ErrCannotChatSelf    = errors.New("cannot create chat with yourself")
	ErrChatAlreadyExists = errors.New("chat already exists")
	ErrUnauthorized      = errors.New("unauthorized")
)

type Service struct {
	db  *gorm.DB
	hub *Hub
}

func NewService(db *gorm.DB, hub *Hub) *Service {
	return &Service{db: db, hub: hub}
}

// GetUserChats returns all chats for a user
func (s *Service) GetUserChats(userID uuid.UUID, filter *ChatFilter) ([]models.Chat, error) {
	var chats []models.Chat

	if err := s.db.
		Preload("User1").
		Preload("User2").
		Preload("Messages", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC").Limit(1)
		}).
		Where("user1_id = ? OR user2_id = ?", userID, userID).
		Order("updated_at DESC").
		Limit(filter.Limit).
		Offset(filter.Offset).
		Find(&chats).Error; err != nil {
		return nil, err
	}

	return chats, nil
}

// GetChatByID returns a chat by ID
func (s *Service) GetChatByID(chatID, userID uuid.UUID) (*models.Chat, error) {
	var chat models.Chat

	if err := s.db.
		Preload("User1").
		Preload("User2").
		First(&chat, "id = ?", chatID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrChatNotFound
		}
		return nil, err
	}

	if chat.User1ID != userID && chat.User2ID != userID {
		return nil, ErrUnauthorized
	}

	return &chat, nil
}

func (s *Service) CreateChat(userID uuid.UUID, req *CreateChatRequest) (*models.Chat, error) {
	if userID == req.UserID {
		return nil, ErrCannotChatSelf
	}

	// Check if other user exists
	var otherUser models.User
	if err := s.db.First(&otherUser, "id = ?", req.UserID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	// Check if chat already exists (in either direction)
	var existingChat models.Chat
	err := s.db.Where(
		"(user1_id = ? AND user2_id = ?) OR (user1_id = ? AND user2_id = ?)",
		userID, req.UserID, req.UserID, userID,
	).First(&existingChat).Error

	if err == nil {
		return nil, ErrChatAlreadyExists
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// Create chat
	chat := models.Chat{
		ID:      uuid.New(),
		User1ID: userID,
		User2ID: req.UserID,
	}

	if err := s.db.Create(&chat).Error; err != nil {
		return nil, err
	}

	// Load relationships
	s.db.Preload("User1").Preload("User2").First(&chat, "id = ?", chat.ID)

	return &chat, nil
}

// GetOrCreateChat gets existing chat or creates new one
func (s *Service) GetOrCreateChat(userID, otherUserID uuid.UUID) (*models.Chat, error) {
	var chat models.Chat

	err := s.db.
		Preload("User1").
		Preload("User2").
		Where(
			"(user1_id = ? AND user2_id = ?) OR (user1_id = ? AND user2_id = ?)",
			userID, otherUserID, otherUserID, userID,
		).First(&chat).Error

	if err == nil {
		return &chat, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// Create new chat
	return s.CreateChat(userID, &CreateChatRequest{UserID: otherUserID})
}

// GetMessages returns messages for a chat
func (s *Service) GetMessages(chatID, userID uuid.UUID, filter *MessageFilter) ([]models.Message, error) {
	// Verify user is part of chat
	var chat models.Chat
	if err := s.db.First(&chat, "id = ?", chatID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrChatNotFound
		}
		return nil, err
	}

	if chat.User1ID != userID && chat.User2ID != userID {
		return nil, ErrUnauthorized
	}

	var messages []models.Message
	query := s.db.Preload("Sender").Where("chat_id = ?", chatID)

	if filter.Before != nil {
		query = query.Where("created_at < ?", *filter.Before)
	}

	if err := query.Order("created_at DESC").Limit(filter.Limit).Find(&messages).Error; err != nil {
		return nil, err
	}

	return messages, nil
}

// SendMessage sends a message in a chat
func (s *Service) SendMessage(chatID, senderID uuid.UUID, req *SendMessageRequest) (*models.Message, error) {
	// Verify user is part of chat
	var chat models.Chat
	if err := s.db.First(&chat, "id = ?", chatID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrChatNotFound
		}
		return nil, err
	}

	if chat.User1ID != senderID && chat.User2ID != senderID {
		return nil, ErrUnauthorized
	}

	message := models.Message{
		ID:       uuid.New(),
		ChatID:   chatID,
		SenderID: senderID,
		Content:  req.Content,
	}

	if err := s.db.Create(&message).Error; err != nil {
		return nil, err
	}

	// Update chat's updated_at
	s.db.Model(&chat).Update("updated_at", time.Now())

	// Load sender
	s.db.Preload("Sender").First(&message, "id = ?", message.ID)

	// Broadcast via WebSocket
	if s.hub != nil {
		recipientID := chat.User1ID
		if recipientID == senderID {
			recipientID = chat.User2ID
		}

		s.hub.SendToUser(recipientID, WSMessage{
			Type:    "new_message",
			Payload: message,
		})
	}

	return &message, nil
}

// MarkAsRead marks a message as read
func (s *Service) MarkAsRead(messageID, userID uuid.UUID) error {
	var message models.Message
	if err := s.db.Preload("Chat").First(&message, "id = ?", messageID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrMessageNotFound
		}
		return err
	}

	// Only recipient can mark as read
	if message.SenderID == userID {
		return nil // Sender can't mark their own message as read
	}

	// Verify user is part of chat
	if message.Chat.User1ID != userID && message.Chat.User2ID != userID {
		return ErrUnauthorized
	}

	now := time.Now()
	if err := s.db.Model(&message).Update("read_at", now).Error; err != nil {
		return err
	}

	// Notify sender via WebSocket
	if s.hub != nil {
		s.hub.SendToUser(message.SenderID, WSMessage{
			Type: "message_read",
			Payload: WSReadEvent{
				ChatID:    message.ChatID,
				MessageID: messageID,
			},
		})
	}

	return nil
}

// MarkChatAsRead marks all messages in a chat as read
func (s *Service) MarkChatAsRead(chatID, userID uuid.UUID) error {
	var chat models.Chat
	if err := s.db.First(&chat, "id = ?", chatID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrChatNotFound
		}
		return err
	}

	if chat.User1ID != userID && chat.User2ID != userID {
		return ErrUnauthorized
	}

	now := time.Now()
	return s.db.Model(&models.Message{}).
		Where("chat_id = ? AND sender_id != ? AND read_at IS NULL", chatID, userID).
		Update("read_at", now).Error
}

// SendTypingIndicator sends typing indicator via WebSocket
func (s *Service) SendTypingIndicator(chatID, userID uuid.UUID) error {
	var chat models.Chat
	if err := s.db.First(&chat, "id = ?", chatID).Error; err != nil {
		return err
	}

	if chat.User1ID != userID && chat.User2ID != userID {
		return ErrUnauthorized
	}

	if s.hub != nil {
		recipientID := chat.User1ID
		if recipientID == userID {
			recipientID = chat.User2ID
		}

		s.hub.SendToUser(recipientID, WSMessage{
			Type: "typing",
			Payload: WSTypingEvent{
				ChatID: chatID,
				UserID: userID,
			},
		})
	}

	return nil
}
