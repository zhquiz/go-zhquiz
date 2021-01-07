package db

import (
	"strconv"
	"time"

	"github.com/zhquiz/go-zhquiz/server/zh"
	"gorm.io/gorm"
)

// Quiz is the database model for quiz
type Quiz struct {
	ID        string `gorm:"primaryKey;check:length(id) > 0"`
	CreatedAt time.Time
	UpdatedAt time.Time

	// Relationships
	UserID string `gorm:"index:quiz_unique_idx,unique;not null"`
	User   User

	// Entry references
	Entry     string `gorm:"index:quiz_unique_idx,unique;not null;check:length(entry) > 0"`
	Type      string `gorm:"index:quiz_unique_idx,unique;not null;check:[type] in ('hanzi','vocab','sentence')"`
	Direction string `gorm:"index:quiz_unique_idx,unique;not null;check:direction in ('se','ec','te')"`

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

// AfterCreate hook
func (q *Quiz) AfterCreate(tx *gorm.DB) (err error) {
	entry := q.Entry
	pinyin := ""
	english := ""
	description := ""
	tag := ""
	level := ""

	switch q.Type {
	case "vocab":
		var vocabs []zh.Vocab
		var tokens []zh.Token
		zhDB.Current.Where("simplified = ? OR traditional = ?", q.Entry, q.Entry).Find(&vocabs)
		zhDB.Current.Where("entry = ?", q.Entry).Find(&tokens)

		entry = ""

		for _, v := range vocabs {
			entry += v.Simplified + " " + v.Traditional + " "
			pinyin += v.Pinyin + " "
			english += v.English + " "

			if v.Source != "" {
				tag += v.Source + " "
			}
		}

		for _, t := range tokens {
			description += t.Description + " "
			tag += t.Tag + " "

			if t.VocabLevel != 0 {
				level = strconv.Itoa(t.HanziLevel)
			}
		}
	case "sentence":
		var sentences []zh.Sentence
		zhDB.Current.Where("chinese = ?", q.Entry).Find(&sentences)

		for _, s := range sentences {
			pinyin += s.Pinyin + " "
			english += s.English + " "
			description += s.Description + " "
			tag += s.Tag + " "

			if s.Level != 0 {
				level = strconv.Itoa(int(s.Level))
			}
		}
	default:
		var tokens []zh.Token
		zhDB.Current.Where("entry = ?", q.Entry).Find(&tokens)

		for _, t := range tokens {
			pinyin += t.Pinyin + " "
			english += t.English + " "
			description += t.Description + " "
			tag += t.Tag + " "

			if t.HanziLevel != 0 {
				level = strconv.Itoa(t.HanziLevel)
			}
		}
	}

	tx.Exec(`
	INSERT INTO quiz_q (id, [entry], [level], [pinyin], [english], [description], [tag])
	VALUES (@id, @entry, @level, @pinyin, @english, @description, @tag)
	`, map[string]interface{}{
		"id":          q.ID,
		"entry":       parseChinese(entry),
		"level":       level,
		"pinyin":      parsePinyin(pinyin),
		"english":     english,
		"description": description,
		"tag":         tag,
	})
	return
}

// AfterDelete hook
func (q *Quiz) AfterDelete(tx *gorm.DB) (err error) {
	tx.Exec(`
	DELETE FROM quiz_q
	WHERE id = ?
	`, q.ID)
	return
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

	return time.Now().Add(10 * time.Minute)
}

// UpdateSRSLevel updates SRSLevel and also updates stats
func (q *Quiz) UpdateSRSLevel(dSRSLevel int8) {
	if dSRSLevel > 0 {
		if q.RightStreak == nil {
			var s uint = 0
			q.RightStreak = &s
		}

		*q.RightStreak++

		if q.MaxRight == nil || *q.MaxRight < *q.RightStreak {
			q.MaxRight = q.RightStreak
		}
	} else if dSRSLevel < 0 {
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
