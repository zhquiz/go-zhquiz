package db

import (
	"time"
)

// Library is user database model for Library
type Library struct {
	ID        int `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time

	UserID      int    `gorm:"index:idx_library_u,unique;not null"`
	Title       string `gorm:"index:idx_library_u,unique;not null"`
	Type        string `gorm:"index:idx_library_u,unique;not null"`
	Entries     StringArray
	Description string
	Tag         StringArray
}
