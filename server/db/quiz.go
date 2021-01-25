package db

import (
	"errors"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/jkomyno/nanoid"
	"github.com/zhquiz/go-zhquiz/server/zh"
	"gorm.io/gorm"
)

// Quiz is the database model for quiz
type Quiz struct {
	ID        string `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time
	UpdatedAt time.Time

	// Entry references
	Entry     string `gorm:"index:quiz_unique_idx,unique;not null" json:"entry"`
	Type      string `gorm:"index:quiz_unique_idx,unique;not null;check:[type] in ('hanzi','vocab','sentence')" json:"type"`
	Direction string `gorm:"index:quiz_unique_idx,unique;not null;check:direction in ('se','ec','te')" json:"direction"`
	Source    string `gorm:"index;not null" json:"source"`

	Description string `gorm:"-"`
	Tag         string `gorm:"-"`

	// Quiz statistics
	SRSLevel    *int8      `gorm:"index"`
	NextReview  *time.Time `gorm:"index"`
	LastRight   *time.Time `gorm:"index" json:"lastRight"`
	LastWrong   *time.Time `gorm:"index"`
	RightStreak *uint      `gorm:"index"`
	WrongStreak *uint      `gorm:"index" json:"wrongStreak"`
	MaxRight    *uint      `gorm:"index"`
	MaxWrong    *uint      `gorm:"index"`
}

// Create ensures q update
func (q *Quiz) Create(tx *gorm.DB) (err error) {
	for q.ID == "" {
		id, err := nanoid.Nanoid(6)
		if err != nil {
			return err
		}

		var count int64
		if r := tx.Model(Quiz{}).Where("id = ?", id).Count(&count); r.Error != nil {
			return err
		}

		if count == 0 {
			q.ID = id
			return nil
		}
	}

	if r := tx.Create(q); r.Error != nil {
		return r.Error
	}

	var old struct {
		ID          string
		Description string
		Tag         string
	}

	if r := tx.Raw(`
	SELECT ID, Description, Tag
	FROM quiz_q
	WHERE id = ?
	`, q.ID).Scan(&old); r.Error != nil && !errors.Is(r.Error, gorm.ErrRecordNotFound) {
		panic(r.Error)
	}

	descSet := map[string]bool{}
	tagSet := map[string]bool{}

	if old.ID != "" {
		if strings.TrimSpace(old.Description) != "" {
			for _, d := range strings.Split(old.Description, " ") {
				descSet[d] = true
			}
		}

		if strings.TrimSpace(old.Tag) != "" {
			for _, t := range strings.Split(old.Tag, " ") {
				tagSet[t] = true
			}
		}
	}

	if strings.TrimSpace(q.Description) != "" {
		for _, d := range strings.Split(parseChinese(q.Description), " ") {
			descSet[d] = true
		}
	}

	if strings.TrimSpace(q.Tag) != "" {
		for _, d := range strings.Split(parseChinese(q.Tag), " ") {
			tagSet[d] = true
		}
	}

	entry := q.Entry
	pinyin := ""
	english := ""
	level := ""

	var getter []struct {
		Description string
		Tag         string
	}

	zhDB.Current.Raw(`
	SELECT [Description], [Tag] FROM token_q WHERE entry = ?
	`, q.Entry).Find(&getter)

	for _, d := range getter {
		if strings.TrimSpace(d.Description) != "" {
			for _, k := range strings.Split(parseChinese(d.Description), " ") {
				descSet[k] = true
			}
		}

		if strings.TrimSpace(d.Tag) != "" {
			for _, k := range strings.Split(parseChinese(d.Tag), " ") {
				tagSet[k] = true
			}
		}
	}

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
				tagSet[v.Source] = true
			}
		}

		for _, t := range tokens {
			if t.VocabLevel != 0 {
				level = strconv.Itoa(t.HanziLevel)
			}
		}

	case "sentence":
		type sen struct {
			Pinyin  string
			English string
			Level   float64
		}
		var sentences []sen

		rows, err := zhDB.Current.Raw(`
		SELECT sentence.pinyin Pinyin, English, Level
		FROM sentence
		LEFT JOIN sentence_q ON sentence_q.id = sentence.id
		WHERE chinese = ?
		GROUP BY sentence.id
		`, q.Entry).Rows()
		if err != nil {
			panic(err)
		}

		for rows.Next() {
			var s sen
			rows.Scan(&s)
			sentences = append(sentences, s)
		}

		for _, s := range sentences {
			pinyin += s.Pinyin + " "
			english += s.English + " "

			if s.Level != 0 {
				level = strconv.Itoa(int(math.Round(s.Level)))
			}
		}
	default:
		var tokens []zh.Token
		zhDB.Current.Where("entry = ?", q.Entry).Find(&tokens)

		for _, t := range tokens {
			pinyin += t.Pinyin + " "
			english += t.English + " "

			if t.HanziLevel != 0 {
				level = strconv.Itoa(t.HanziLevel)
			}
		}
	}

	if q.Source == "extra" {
		var extra Extra
		tx.Select(`
		extra.description [Description],
		extra_q.tag       [Tag]
		`).Joins("LEFT JOIN extra_q ON extra.id = extra_q.id").
			Where("extra.chinese = ? AND type = ?", q.Entry, q.Type).
			Group("extra.id").
			First(&extra)

		if strings.TrimSpace(extra.Description) != "" {
			for _, d := range strings.Split(parseChinese(extra.Description), " ") {
				descSet[d] = true
			}
		}

		if strings.TrimSpace(extra.Tag) != "" {
			for _, d := range strings.Split(parseChinese(extra.Tag), " ") {
				tagSet[d] = true
			}
		}
	}

	old.Description = func() string {
		desc := make([]string, 0)
		for k := range descSet {
			desc = append(desc, k)
		}

		return strings.Join(desc, " ")
	}()

	old.Tag = func() string {
		tags := make([]string, 0)
		for k := range tagSet {
			tags = append(tags, k)
		}

		tags = append(tags, "level"+level)

		return strings.Join(tags, " ")
	}()

	if old.ID != "" {
		if r := tx.Exec(`
		UPDATE quiz_q SET description = @Description, tag = @Tag WHERE id = @ID
		`, old); r.Error != nil {
			panic(r.Error)
		}
	} else {
		if r := tx.Exec(`
		INSERT INTO quiz_q (id, [entry], [pinyin], [english], [type], [direction], [source], [description], [tag])
		SELECT @id, @entry, @pinyin, @english, @type, @direction, @source, @description, @tag
		WHERE EXISTS (SELECT 1 FROM quiz WHERE id = @id)
		`, map[string]interface{}{
			"id":          q.ID,
			"entry":       parseChinese(entry),
			"type":        q.Type,
			"direction":   q.Direction,
			"source":      q.Source,
			"level":       level,
			"pinyin":      parsePinyin(pinyin),
			"english":     english,
			"description": old.Description,
			"tag":         old.Tag,
		}); r.Error != nil {
			panic(r.Error)
		}
	}

	return
}

// Delete ensures q delete
func (q *Quiz) Delete(tx *gorm.DB) error {
	if r := tx.Delete(q); r.Error != nil {
		return r.Error
	}

	if r := tx.Exec(`
	DELETE FROM quiz_q
	WHERE id = ?
	`, q.ID); r.Error != nil {
		return r.Error
	}

	return nil
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
