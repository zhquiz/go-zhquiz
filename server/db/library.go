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

// Create creates along with q
func (u *Library) Create(tx *gorm.DB) error {
	for u.ID == "" {
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
		}
	}

	if r := tx.Create(u); r.Error != nil {
		return r.Error
	}

	if r := tx.Exec(`
	INSERT INTO library_q (id, title, [entry], [description], tag)
	SELECT @id, @title, @entry, @description, @tag
	WHERE EXISTS (SELECT 1 FROM library WHERE id = @id)
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

// Update makes sure description and tag are always updated
func (u *Library) Update(tx *gorm.DB) error {
	if u.Description == "" {
		u.Description = " "
	}

	if r := tx.Updates(u); r.Error != nil {
		return r.Error
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

// Delete ensures q delete
func (u *Library) Delete(tx *gorm.DB) error {
	if r := tx.Delete(u); r.Error != nil {
		return r.Error
	}

	if r := tx.Exec(`
	DELETE FROM library_q
	WHERE id = ?
	`, u.ID); r.Error != nil {
		return r.Error
	}

	return nil
}
