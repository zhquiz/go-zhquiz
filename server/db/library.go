package db

import (
	"strings"
	"time"

	"github.com/jkomyno/nanoid"
	"gorm.io/gorm"
)

// Library is user database model for Library
type Library struct {
	ID        string `gorm:"primarykey" json:"id"`
	CreatedAt time.Time
	UpdatedAt time.Time

	Title       string      `gorm:"index:idx_library_user_title,unique;not null" json:"title"`
	Entries     StringArray `json:"entries"`
	Description string      `json:"description"`
	Tag         string      `gorm:"-" json:"tag"`
}

// BeforeCreate generates ID if not exists
func (u *Library) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == "" {
		for {
			id, err := nanoid.Nanoid(6)
			if err != nil {
				return err
			}

			var count int64
			if r := tx.Model(Library{}).Where("id = ?", id).Count(&count); r.Error != nil {
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
func (u *Library) AfterCreate(tx *gorm.DB) (err error) {
	tx.Exec(`
	INSERT INTO library_q (id, title, [entry], [description], tag)
	SELECT @id, @title, @entry, @description, @tag
	WHERE EXISTS (SELECT 1 FROM library WHERE id = @id)
	`, map[string]interface{}{
		"id":          u.ID,
		"title":       u.Title,
		"entry":       strings.Join(u.Entries, " "),
		"description": parseChinese(u.Description),
		"tag":         u.Tag,
	})
	return
}

// FullUpdate makes sure description and tag are always updated
func (u *Library) FullUpdate(tx *gorm.DB) error {
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
	UPDATE library_q
	SET title = @title, entry = @entry, [description] = @description, tag = @tag
	WHERE id = @id
	`, map[string]interface{}{
		"id":          u.ID,
		"title":       u.Title,
		"entry":       strings.Join(u.Entries, " "),
		"description": parseChinese(u.Description),
		"tag":         u.Tag,
	}); r.Error != nil {
		return r.Error
	}

	return nil
}

// AfterDelete hook
func (u *Library) AfterDelete(tx *gorm.DB) (err error) {
	tx.Exec(`
	DELETE FROM library_q
	WHERE id = ?
	`, u.ID)
	return
}
