package db

import (
	"time"

	"github.com/jkomyno/nanoid"
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
	tx.Exec(`
	INSERT INTO extra_q (id, chinese, pinyin, english, [type], description, tag)
	SELECT @id, @chinese, @pinyin, @english, @type, @description, @tag
	WHERE NOT EXISTS (SELECT 1 FROM extra_q WHERE id = @id)
	`, map[string]interface{}{
		"id":          u.ID,
		"chinese":     parseChinese(u.Chinese),
		"pinyin":      parsePinyin(u.Pinyin),
		"english":     u.English,
		"type":        u.Type,
		"description": parseChinese(u.Description),
		"tag":         u.Tag,
	})
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
