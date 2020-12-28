package db

import (
	"log"
	"time"

	"github.com/zhquiz/go-server/server/rand"
	"gorm.io/gorm"
)

// User holds user data
type User struct {
	ID        string `gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Email  string `gorm:"index:,unique;not null;check:email <> ''"`
	APIKey string `gorm:"index,not null;check:api_key <> ''"`

	Meta UserMeta `gorm:"type:json"`

	// Relations
	Decks   []Deck   `gorm:"constraint:OnDelete:CASCADE"`
	Entries []Entry  `gorm:"constraint:OnDelete:CASCADE"`
	Presets []Preset `gorm:"constraint:OnDelete:CASCADE"`
	Quizzes []Quiz   `gorm:"constraint:OnDelete:CASCADE"`
}

// UserMeta holds User's settings
type UserMeta struct {
	Forvo    *string
	Level    *uint
	LevelMin *uint
	Quiz     struct {
		Type      []string
		Stage     []string
		Direction []string
		IsDue     bool
	}
}

// New creates new User record
func (u *User) New(email string) {
	u.Email = email

	u.NewAPIKey()
}

// NewAPIKey generates a new API key to the User, and returns it
func (u *User) NewAPIKey() string {
	apiKey, err := rand.GenerateRandomString(64)
	if err != nil {
		log.Fatalln(err)
	}

	u.APIKey = apiKey

	return apiKey
}
