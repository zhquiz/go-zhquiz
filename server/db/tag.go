package db

import "time"

// Tag is the database model for tag
type Tag struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time

	Name string `gorm:"index:,unique;not null;check:length(name) > 0"`
}
