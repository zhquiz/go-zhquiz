package db

import (
	"errors"
	"strings"
	"time"

	"github.com/jkomyno/nanoid"
	"gopkg.in/yaml.v2"
	"gorm.io/gorm"
)

// Extra is user database model for Extra
type Extra struct {
	ID        string    `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`

	Chinese     string `gorm:"index:,unique;not null" json:"chinese"`
	Pinyin      string `json:"pinyin"`
	English     string `gorm:"-" json:"english"`
	Type        string `gorm:"-" json:"type"`
	Description string `json:"description"`
	Tag         string `gorm:"-" json:"tag"`
}

// BeforeCreate generates ID if not exists
func (u *Extra) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == "" {
		for {
			id, err := nanoid.Nanoid(6)
			if err != nil {
				return err
			}

			var count int64
			if r := tx.Model(Extra{}).Where("id = ?", id).Count(&count); r.Error != nil {
				return err
			}

			if count == 0 {
				u.ID = id
				return nil
			}
		}
	}

	return
}

// AfterCreate hook
func (u *Extra) AfterCreate(tx *gorm.DB) (err error) {
	var old struct {
		ID          string
		Description string
		Tag         string
	}

	if r := tx.Raw(`
	SELECT extra.id ID, extra.description Description, extra_q.tag Tag
	FROM extra
	LEFT JOIN extra_q ON extra.id = extra_q.id
	WHERE extra.id = ?
	`, u.ID).Scan(&old); r.Error != nil && !errors.Is(r.Error, gorm.ErrRecordNotFound) {
		panic(r.Error)
	}

	descSet := map[string]bool{}
	tagSet := map[string]bool{}

	if old.ID != "" {
		if strings.TrimSpace(old.Description) != "" {
			var desc []string
			e := yaml.Unmarshal([]byte(old.Description), &desc)
			if e != nil {
				descSet[old.Description] = true
			} else {
				for _, d := range desc {
					descSet[d] = true
				}
			}
		}

		if strings.TrimSpace(old.Tag) != "" {
			for _, t := range strings.Split(old.Tag, " ") {
				tagSet[t] = true
			}
		}
	}

	if strings.TrimSpace(u.Description) != "" {
		descSet[u.Description] = true
	}

	if strings.TrimSpace(u.Tag) != "" {
		for _, t := range strings.Split(u.Tag, " ") {
			tagSet[t] = true
		}
	}

	description := func() string {
		desc := make([]string, 0)
		for k := range descSet {
			desc = append(desc, k)
		}

		descByte, e := yaml.Marshal(desc)
		if e != nil {
			panic(e)
		}

		return string(descByte)
	}()

	descriptionParsed := parseChinese(description)

	tags := make([]string, 0)
	for k := range tagSet {
		tags = append(tags, k)
	}

	old.Tag = strings.Join(tags, " ")

	if old.ID != "" {
		if r := tx.Exec(`
		UPDATE extra SET description = @description WHERE id = @id
		`, map[string]interface{}{
			"description": description,
			"id":          old.ID,
		}); r.Error != nil {
			panic(r.Error)
		}

		if r := tx.Exec(`
		UPDATE extra_q SET description = @description, tag = @tag WHERE id = @id
		`, map[string]interface{}{
			"description": descriptionParsed,
			"tag":         old.Tag,
			"id":          old.ID,
		}); r.Error != nil {
			panic(r.Error)
		}
	} else {
		if r := tx.Exec(`
		INSERT INTO extra_q (id, chinese, pinyin, english, [type], description, tag)
		SELECT @id, @chinese, @pinyin, @english, @type, @description, @tag
		`, map[string]interface{}{
			"id":          u.ID,
			"chinese":     parseChinese(u.Chinese),
			"pinyin":      parsePinyin(u.Pinyin),
			"english":     u.English,
			"type":        u.Type,
			"description": descriptionParsed,
			"tag":         old.Tag,
		}); r.Error != nil {
			panic(r.Error)
		}
	}

	return
}

// FullUpdate makes sure description and tag are always updated
func (u *Extra) FullUpdate(tx *gorm.DB) error {
	if u.Description == "" {
		u.Description = " "
	}

	if u.Tag == "" {
		u.Tag = " "
	}

	if r := tx.Where("id = ?", u.ID).Updates(&u); r.Error != nil {
		return r.Error
	}

	if r := tx.Exec(`
	UPDATE extra_q
	SET chinese = @chinese, pinyin = @pinyin, english = @english, [description] = @description, tag = @tag
	WHERE id = @id
	`, map[string]interface{}{
		"id":          u.ID,
		"chinese":     parseChinese(u.Chinese),
		"pinyin":      parsePinyin(u.Pinyin),
		"english":     u.English,
		"type":        u.Type,
		"description": parseChinese(u.Description),
		"tag":         u.Tag,
	}); r.Error != nil {
		return r.Error
	}

	return nil
}

// AfterDelete hook
func (u *Extra) AfterDelete(tx *gorm.DB) (err error) {
	tx.Exec(`
	DELETE FROM extra_q
	WHERE id = ?
	`, u.ID)
	return
}
