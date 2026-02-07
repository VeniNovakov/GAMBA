package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GameStatus string

const (
	GameStatusActive   GameStatus = "active"
	GameStatusInactive GameStatus = "inactive"
	GameStatusHidden   GameStatus = "hidden"
)

type GameCategory string

const (
	GameCategorySlots GameCategory = "slots"
	GameCategoryDice  GameCategory = "dice"
	GameCategoryOther GameCategory = "other"
)

type Game struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name        string         `json:"name" gorm:"not null"`
	Description string         `json:"description" gorm:"type:text"`
	Category    GameCategory   `json:"category" gorm:"type:varchar(30);not null"`
	Status      GameStatus     `json:"status" gorm:"type:varchar(20);default:'active'"`
	MinBet      int64          `json:"min_bet" gorm:"not null"`
	MaxBet      int64          `json:"max_bet" gorm:"not null"`
	HouseEdge   int64          `json:"house_edge" gorm:"default:0"`
	CreatedAt   time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

func (Game) TableName() string {
	return "games"
}
