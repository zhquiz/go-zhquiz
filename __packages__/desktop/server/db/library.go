package db

import (
	"time"

	"gorm.io/gorm"
)

// Library is fts5 model for Library
type Library struct {
	ID          uint        `json:"id"`
	UserID      uint        `gorm:"-" json:"-"` // nullable
	Title       string      `json:"title"`
	Entry       StringArray `json:"entries"`
	Description string      `json:"description"`
	Tag         StringArray `json:"tag"`
	Type        string      `json:"type"`
}

// LibraryBase is constraint indexed model for Library
type LibraryBase struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	UserID    uint      `json:"-"` // nullable
	Title     string    `gorm:"index:,unique;not null" json:"title"`
}

// BaseCreate ensures base create
func (Library) BaseCreate(tx *gorm.DB, items ...*Library) error {
	bases := make([]LibraryBase, len(items))
	for i, base := range bases {
		base.UserID = items[i].UserID
		base.Title = items[i].Title
	}

	if r := tx.Create(&bases); r.Error != nil {
		return r.Error
	}

	if r := tx.Create(&items); r.Error != nil {
		return r.Error
	}

	for i, base := range bases {
		items[i].ID = base.ID
	}

	return nil
}

// BaseUpdate ensures base update
func (Library) BaseUpdate(tx *gorm.DB, items ...*Library) error {
	bases := make([]LibraryBase, len(items))
	for i, base := range bases {
		base.ID = items[i].ID
		base.UserID = items[i].UserID
		base.Title = items[i].Title
	}

	if r := tx.Updates(&bases); r.Error != nil {
		return r.Error
	}

	if r := tx.Updates(&items); r.Error != nil {
		return r.Error
	}

	return nil
}

// BaseDelete ensures base delete
func (Library) BaseDelete(tx *gorm.DB, items ...*Library) error {
	bases := make([]LibraryBase, len(items))
	for i, base := range bases {
		base.ID = items[i].ID
	}

	if r := tx.Delete(&bases); r.Error != nil {
		return r.Error
	}

	if r := tx.Delete(&items); r.Error != nil {
		return r.Error
	}

	return nil
}
