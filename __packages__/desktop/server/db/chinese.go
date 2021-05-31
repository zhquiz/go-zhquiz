package db

import (
	"time"

	"gorm.io/gorm"
)

// Chinese is fts5 model for Chinese
type Chinese struct {
	ID          uint        `json:"id"`
	UserID      uint        `gorm:"-" json:"-"`
	Source      string      `json:"source"`
	Type        string      `json:"type"`
	Chinese     string      `json:"chinese"`
	Alt         StringArray `json:"alt"`
	Pinyin      StringArray `json:"pinyin"`
	English     StringArray `json:"english"`
	Description string      `json:"description"`
	Tag         StringArray `json:"tag"`
}

// ChineseBase is constraint indexed model for Chinese
type ChineseBase struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	UserID    uint      `json:"-"` // nullable
	Source    string    `gorm:"index:idx_chinese_u,unique;not null" json:"source"`
	Type      string    `gorm:"index:idx_chinese_u,unique;not null" json:"type"`
	Chinese   string    `gorm:"index:idx_chinese_u,unique;not null" json:"chinese"`
}

// BaseCreate ensures base create
func (Chinese) BaseCreate(tx *gorm.DB, items ...*Chinese) error {
	bases := make([]ChineseBase, len(items))
	for i, base := range bases {
		base.UserID = items[i].UserID
		base.Source = items[i].Source
		base.Type = items[i].Type
		base.Chinese = items[i].Chinese
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
func (Chinese) BaseUpdate(tx *gorm.DB, items ...*Chinese) error {
	bases := make([]ChineseBase, len(items))
	for i, base := range bases {
		base.ID = items[i].ID
		base.UserID = items[i].UserID
		base.Source = items[i].Source
		base.Type = items[i].Type
		base.Chinese = items[i].Chinese
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
func (Chinese) BaseDelete(tx *gorm.DB, items ...*Chinese) error {
	bases := make([]ChineseBase, len(items))
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
