package db

import (
	"time"
)

// Extra is user database model for Extra
type Extra struct {
	ID        int `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time

	UserID      int    `gorm:"index:idx_extra_u,unique;not null"`
	Entry       string `gorm:"index:idx_extra_u,unique;not null"`
	Type        string `gorm:"index:idx_extra_u,unique;not null"`
	Reading     string `gorm:"index"`
	English     string
	Description string
}

// Tag is user database model for Tag
type Tag struct {
	ID     int    `gorm:"primarykey" json:"id"`
	UserID int    `gorm:"index:idx_tag_u,unique;not null"`
	Entry  string `gorm:"index:idx_tag_u,unique;not null"`
	Type   string `gorm:"index:idx_tag_u,unique;not null"`
	Name   string `gorm:"index:idx_tag_u,unique;not null"`
}
