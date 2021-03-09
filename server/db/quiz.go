package db

import (
	"time"
)

// Quiz is the database model for quiz
type Quiz struct {
	ID        int `gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time

	// Entry references
	UserID    int    `gorm:"index:idx_quiz_u,unique;not null"`
	Entry     string `gorm:"index:idx_quiz_u,unique;not null"`
	Type      string `gorm:"index:idx_quiz_u,unique;not null;check:[type] in ('character','vocabulary','sentence')"`
	Direction string `gorm:"index:idx_quiz_u,unique;not null;check:direction in ('forward','backward','traditional')"`

	// Quiz statistics
	SRSLevel    *int8      `gorm:"index"`
	NextReview  *time.Time `gorm:"index"`
	LastRight   *time.Time `gorm:"index"`
	LastWrong   *time.Time `gorm:"index"`
	RightStreak *uint      `gorm:"index"`
	WrongStreak *uint      `gorm:"index"`
	MaxRight    *uint      `gorm:"index"`
	MaxWrong    *uint      `gorm:"index"`
}

var srsMap []time.Duration = []time.Duration{
	4 * time.Hour,
	8 * time.Hour,
	24 * time.Hour,
	3 * 24 * time.Hour,
	7 * 24 * time.Hour,
	2 * 7 * 24 * time.Hour,
	4 * 7 * 24 * time.Hour,
	16 * 7 * 24 * time.Hour,
}

func getNextReview(srsLevel int8) time.Time {
	if srsLevel >= 0 && srsLevel < int8(len(srsMap)) {
		return time.Now().Add(srsMap[srsLevel])
	}

	return time.Now().Add(1 * time.Hour)
}

// UpdateSRSLevel updates SRSLevel and also updates stats
func (q *Quiz) UpdateSRSLevel(dSRSLevel int8) {
	now := time.Now()

	if dSRSLevel > 0 {
		q.LastRight = &now

		if q.RightStreak == nil {
			var s uint = 0
			q.RightStreak = &s
		}

		*q.RightStreak++

		if q.MaxRight == nil || *q.MaxRight < *q.RightStreak {
			q.MaxRight = q.RightStreak
		}
	} else if dSRSLevel < 0 {
		q.LastWrong = &now

		if q.WrongStreak == nil {
			var s uint = 0
			q.WrongStreak = &s
		}

		*q.WrongStreak++

		if q.MaxWrong == nil || *q.MaxWrong < *q.WrongStreak {
			q.MaxWrong = q.WrongStreak
		}
	}

	if q.SRSLevel == nil {
		var s int8 = 0
		q.SRSLevel = &s
	}

	*q.SRSLevel += dSRSLevel

	if *q.SRSLevel >= int8(len(srsMap)) {
		*q.SRSLevel = int8(len(srsMap) - 1)
	}

	if *q.SRSLevel < 0 {
		*q.SRSLevel = 0
		nextReview := getNextReview(-1)
		q.NextReview = &nextReview
	} else {
		nextReview := getNextReview(*q.SRSLevel)
		q.NextReview = &nextReview
	}
}
