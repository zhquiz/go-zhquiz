package db

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
)

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

// Scan (internal) to-external method for StringArray Type
func (q *S4QuizObject) Scan(value interface{}) error {
	s, ok := value.(string)
	if !ok {
		return errors.New(fmt.Sprint("Not a string:", value))
	}

	result := S4QuizObject{}
	err := json.Unmarshal([]byte(s), &result)
	*q = result
	return err
}

// Value (internal) to-database method for StringArray Type
func (q S4QuizObject) Value() (driver.Value, error) {
	ba, e := json.Marshal(q)
	return string(ba), e
}
