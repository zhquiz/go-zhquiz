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
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time

	Meta UserMeta
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
	s, ok := value.(string)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JSON value:", value))
	}

	result := UserMeta{}
	err := json.Unmarshal([]byte(s), &result)
	*j = result
	return err
}

// Value return json value, implement driver.Valuer interface
func (j UserMeta) Value() (driver.Value, error) {
	bytes, err := json.Marshal(j)
	return string(bytes), err
}

// GormDBDataType represents UserMeta's data type
func (UserMeta) GormDBDataType(_db *gorm.DB, _ *schema.Field) string {
	return "JSON"
}
