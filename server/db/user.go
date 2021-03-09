package db

import (
	"time"
)

// S4QuizObject holds settings for quiz
type S4QuizObject struct {
	Type         []string `json:"type"`
	Stage        []string `json:"stage"`
	Direction    []string `json:"direction"`
	IncludeUndue bool     `json:"includeUndue"`
	IncludeExtra bool     `json:"includeExtra"`
	Q            string   `json:"q"`
}

// User holds user data
type User struct {
	ID        int `gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time

	Identifier string `gorm:"index:,unique;not null"`
	Level      *uint
	LevelMin   *uint
	S4Level    *StringArray
	S4Quiz     *S4QuizObject
}
