package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Role string

const (
	RolePlayer        Role = "player"
	RoleAdministrator Role = "administrator"
)

type User struct {
	ID           uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name         string         `json:"name" gorm:"uniqueIndex;not null"`
	PasswordHash string         `json:"-" gorm:"not null"`
	Role         Role           `json:"role" gorm:"type:varchar(20);default:'player'"`
	Balance      int64          `json:"balance" gorm:"default:0"`
	IsActive     bool           `json:"is_active" gorm:"default:true"`
	IsRestricted bool           `json:"is_restricted" gorm:"default:false"`
	CreatedAt    time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

func (User) TableName() string {
	return "users"
}
