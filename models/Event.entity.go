package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EventStatus string

const (
	EventStatusUpcoming  EventStatus = "upcoming"
	EventStatusLive      EventStatus = "live"
	EventStatusCompleted EventStatus = "completed"
	EventStatusCancelled EventStatus = "cancelled"
)

type EventCategory string

const (
	EventCategorySports        EventCategory = "sports"
	EventCategoryEsports       EventCategory = "esports"
	EventCategoryPolitics      EventCategory = "politics"
	EventCategoryEntertainment EventCategory = "entertainment"
	EventCategoryOther         EventCategory = "other"
)

type Event struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name        string         `json:"name" gorm:"not null"`
	Description string         `json:"description" gorm:"type:text"`
	Category    EventCategory  `json:"category" gorm:"type:varchar(30);not null"`
	Status      EventStatus    `json:"status" gorm:"type:varchar(20);default:'upcoming'"`
	StartsAt    time.Time      `json:"starts_at" gorm:"not null"`
	EndsAt      *time.Time     `json:"ends_at,omitempty"`
	CreatedAt   time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	Outcomes []EventOutcome `json:"outcomes,omitempty" gorm:"foreignKey:EventID"`
}

type EventOutcome struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	EventID   uuid.UUID      `json:"event_id" gorm:"type:uuid;not null;index"`
	Name      string         `json:"name" gorm:"not null"`
	Odds      float64        `json:"odds" gorm:"not null"`
	IsWinner  *bool          `json:"is_winner,omitempty"`
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (Event) TableName() string {
	return "events"
}

func (EventOutcome) TableName() string {
	return "event_outcomes"
}
