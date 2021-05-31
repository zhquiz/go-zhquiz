package db

import (
	"time"
)

// Quiz is the database model for quiz
type Quiz struct {
	ID          uint        `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time   `json:"createdAt"`
	UpdatedAt   time.Time   `json:"updatedAt"`
	UserID      uint        `json:"-"`
	Entry       string      `gorm:"index:idx_quiz_u,unique;not null" json:"entry"`
	Type        string      `gorm:"index:idx_quiz_u,unique;not null;check:[type] in ('hanzi','vocab','sentence')" json:"type"`
	Direction   string      `gorm:"index:idx_quiz_u,unique;not null;check:direction in ('se','ec','te')" json:"direction"`
	Description string      `json:"description"`
	Tag         StringArray `json:"tag"`
	SRSLevel    *int        `gorm:"index" json:"srsLevel"`
	NextReview  *time.Time  `gorm:"index" json:"nextReview"`
	LastRight   *time.Time  `gorm:"index" json:"lastRight"`
	LastWrong   *time.Time  `gorm:"index" json:"lastWrong"`
	RightStreak *int        `gorm:"index" json:"rightStreak"`
	WrongStreak *int        `gorm:"index" json:"wrongStreak"`
	MaxRight    *int        `gorm:"index" json:"maxRight"`
	MaxWrong    *int        `gorm:"index" json:"maxWrong"`
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

func getNextReview(srsLevel int) time.Time {
	if srsLevel >= 0 && srsLevel < len(srsMap) {
		return time.Now().Add(srsMap[srsLevel])
	}

	return time.Now().Add(1 * time.Hour)
}

// UpdateSRSLevel updates SRSLevel and also updates stats
func (q *Quiz) UpdateSRSLevel(dSRSLevel int) {
	now := time.Now()

	if dSRSLevel > 0 {
		q.LastRight = &now

		if q.RightStreak == nil {
			s := 0
			q.RightStreak = &s
		}

		*q.RightStreak++

		if q.MaxRight == nil || *q.MaxRight < *q.RightStreak {
			q.MaxRight = q.RightStreak
		}
	} else if dSRSLevel < 0 {
		q.LastWrong = &now

		if q.WrongStreak == nil {
			s := 0
			q.WrongStreak = &s
		}

		*q.WrongStreak++

		if q.MaxWrong == nil || *q.MaxWrong < *q.WrongStreak {
			q.MaxWrong = q.WrongStreak
		}
	}

	if q.SRSLevel == nil {
		s := 0
		q.SRSLevel = &s
	}

	*q.SRSLevel += dSRSLevel

	if *q.SRSLevel >= len(srsMap) {
		*q.SRSLevel = len(srsMap) - 1
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
