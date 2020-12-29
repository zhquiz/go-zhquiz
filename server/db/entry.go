package db

import (
	"time"

	"github.com/zhquiz/go-server/server/types"
	"gorm.io/gorm"
)

// EntryItem (internal) are items of an entry
type EntryItem struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time

	Name    string `gorm:"index:name_entryId_idx,unique;not null;check:length(name) > 0"`
	EntryID uint   `gorm:"index:name_entryId_idx,unique;not null"`
}

// Entry is a custom dictionary entry
type Entry struct {
	gorm.Model

	// Relationships
	UserID uint  `gorm:"index;not null"`
	Tags   []Tag `gorm:"many2many:entry_tag"`

	// Entry
	Items        []EntryItem       `gorm:"foreignKey:EntryID"`
	Readings     types.StringArray `gorm:"type:text;not null;check:readings <> ''"`
	Translations types.StringArray `gorm:"type:text;not null;check:translations <> ''"`

	Type string `gorm:"index;not null;check:[type] in ('hanzi', 'vocab', 'sentence')"`
}
