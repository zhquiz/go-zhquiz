package db

import "time"

// Deck is user database model for Deck
type Deck struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time

	UserID uint `gorm:"index:deck_unique_idx,unique;not null"`

	Name string `gorm:"index:deck_unique_idx,unique;not null;check:length(name) > 0"`
	Q    string `gorm:"not null"`
}
