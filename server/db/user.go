package db

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// User holds user data
type User struct {
	ID        string `gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time

	Meta UserMeta
}

// BeforeCreate forces single user
func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	u.ID = "_"
	return
}

// UserMeta holds User's settings
type UserMeta struct {
	Forvo    *string `json:"forvo"`
	Level    *uint   `json:"level"`
	LevelMin *uint   `json:"levelMin"`
	Settings struct {
		Level struct {
			WhatToShow string `json:"whatToShow"`
		} `json:"level"`
		Quiz struct {
			Type         []string `json:"type"`
			Stage        []string `json:"stage"`
			Direction    []string `json:"direction"`
			IncludeUndue bool     `json:"includeUndue"`
			IncludeExtra bool     `json:"includeExtra"`
			Q            string   `json:"q"`
		} `json:"quiz"`
		Sentence struct {
			Min *uint `json:"min"`
			Max *uint `json:"max"`
		} `json:"sentence"`
	} `json:"settings"`
}

// Scan scan value into Jsonb, implements sql.Scanner interface
func (j *UserMeta) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}

	result := UserMeta{}
	err := json.Unmarshal(bytes, &result)
	*j = result
	return err
}

// Value return json value, implement driver.Valuer interface
func (j UserMeta) Value() (driver.Value, error) {
	return json.Marshal(j)
}

// GormDBDataType represents UserMeta's data type
func (UserMeta) GormDBDataType(db *gorm.DB, _ *schema.Field) string {
	switch db.Dialector.Name() {
	case "mysql", "sqlite":
		return "JSON"
	case "postgres":
		return "JSONB"
	}
	return "TEXT"
}
