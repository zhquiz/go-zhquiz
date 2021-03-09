package zh

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// Library is user database model for Library
type Library struct {
	ID        int `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time

	Title       string `gorm:"index:idx_library_u,unique;not null"`
	Type        string `gorm:"index:idx_library_u,unique;not null"`
	Entries     StringArray
	Description string
	Tag         StringArray
}

// StringArray type is stringified JSON array
type StringArray []string

// Scan (internal) to-external method for StringArray Type
func (sa *StringArray) Scan(value interface{}) error {
	s, ok := value.(string)
	if !ok {
		return errors.New(fmt.Sprint("Not a string:", value))
	}

	result := make(StringArray, 0)
	err := json.Unmarshal([]byte(s), &result)
	*sa = result
	return err
}

// Value (internal) to-database method for StringArray Type
func (sa StringArray) Value() (driver.Value, error) {
	ba, e := json.Marshal(sa)
	return string(ba), e
}
