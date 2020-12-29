package db

import (
	"time"

	"github.com/zhquiz/go-server/server/types"
)

// Preset holds deck states
type Preset struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time

	UserID uint `gorm:"index:preset_unique_idx,unique;not null"`

	Name     string            `gorm:"index:preset_unique_idx,unique;not null;length(name) > 0"`
	Q        string            `gorm:"not null"`
	Status   PresetStatus      `gorm:"embedded"`
	Selected types.StringArray `gorm:"type:text;not null"`
	Opened   types.StringArray `gorm:"type:text;not null"`
}

// PresetStatus (internal) holds status of a Preset
type PresetStatus struct {
	New       bool `gorm:"not null"`
	Due       bool `gorm:"not null"`
	Leech     bool `gorm:"not null"`
	Graduated bool `gorm:"not null"`
}
