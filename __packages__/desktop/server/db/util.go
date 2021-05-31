package db

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// StringArray type that is compat with both SQLite and PostGres
type StringArray []string

// Scan (internal) to-external method for StringArray Type
func (sa *StringArray) Scan(value interface{}) error {
	s, ok := value.(string)
	if !ok {
		return errors.New(fmt.Sprint("Not a string:", value))
	}

	result := make([]string, 0)
	err := json.Unmarshal([]byte(s), &result)
	*sa = result
	return err
}

// Value (internal) to-database method for StringArray Type
func (sa StringArray) Value() (driver.Value, error) {
	bytes, err := json.Marshal(sa)
	return string(bytes), err
}

// GormDBDataType represents UserMeta's data type
func (StringArray) GormDBDataType(_db *gorm.DB, _ *schema.Field) string {
	return "JSON"
}
