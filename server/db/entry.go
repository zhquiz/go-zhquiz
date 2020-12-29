package db

import (
	"time"
)

// EntryItem are items of an entry
// @internal
type EntryItem struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time

	Name string `gorm:"index:name_entryId_idx,unique;not null;check:length(name) > 0"`

	EntryID uint `gorm:"index:name_entryId_idx,unique;not null"`
	Entry   Entry
}

// Entry is a custom dictionary entry
type Entry struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time

	UserID uint `gorm:"index;not null"`
	User   User

	Tags []Tag `gorm:"many2many:entry_tag"`

	// Entry
	Items        []EntryItem `gorm:"foreignKey:EntryID"`
	Readings     StringArray `gorm:"type:text;not null;check:length(readings) > 0"`
	Translations StringArray `gorm:"type:text;not null;check:length(translations) > 0"`

	Type string `gorm:"index;not null;check:[type] in ('hanzi', 'vocab', 'sentence')"`
}

// EntryTag is joint table of Entry-Tag
// @internal
type EntryTag struct {
	EntryID uint `gorm:"primaryKey"`
	TagID   uint `gorm:"primaryKey"`
	Entry   Entry
	Tag     Tag
}
