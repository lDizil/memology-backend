package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID           uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Username     string         `json:"username" gorm:"unique;not null" validate:"required,min=3,max=50"`
	Email        string         `json:"email" gorm:"unique" validate:"email"`
	PasswordHash string         `json:"-" gorm:"not null"`
	AvatarURL    string         `json:"avatar_url,omitempty"`
	IsActive     bool           `json:"is_active" gorm:"default:true"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`

	Sessions []UserSession `json:"-"`
	Memes    []Meme        `json:"-"`
}

type UserSession struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID    uuid.UUID `json:"user_id" gorm:"not null"`
	TokenHash string    `json:"-" gorm:"not null"`
	ExpiresAt time.Time `json:"expires_at" gorm:"not null"`
	CreatedAt time.Time `json:"created_at"`

	User User `json:"-" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

type Meme struct {
	ID               uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID           uuid.UUID      `json:"user_id" gorm:"not null"`
	Prompt           string         `json:"prompt" gorm:"not null"`
	ImageURL         string         `json:"image_url" gorm:"not null"`
	GenerationTimeMs int            `json:"generation_time_ms,omitempty"`
	Status           string         `json:"status" gorm:"default:completed"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `json:"-" gorm:"index"`

	User    User         `json:"-" gorm:"foreignKey:UserID"`
	Metrics *MemeMetrics `json:"metrics,omitempty" gorm:"foreignKey:MemeID"`
}

type MemeMetrics struct {
	ID                uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	MemeID            uuid.UUID `json:"meme_id" gorm:"unique;not null"`
	RatingScore       int       `json:"rating_score" gorm:"default:0"`
	ClickCount        int       `json:"click_count" gorm:"default:0"`
	DownloadCount     int       `json:"download_count" gorm:"default:0"`
	OtherInteractions int       `json:"other_interactions" gorm:"default:0"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`

	Meme Meme `json:"-" gorm:"foreignKey:MemeID;constraint:OnDelete:CASCADE"`
}
