package db

import (
	"strings"
	"time"

	"github.com/zhquiz/go-zhquiz/server/rand"
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
	Tag         string      `json:"tag"`
}

// BeforeCreate generates ID if not exists
func (u *Library) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == "" {
		u.ID = rand.NewULID()
	}

	return
}

// AfterCreate hook
func (u *Library) AfterCreate(tx *gorm.DB) (err error) {
	tx.Exec(`
	INSERT INTO library_q (id, title, [entry], [description], tag)
	VALUES (@id, @title, @entry, @description, @tag)
	`, map[string]interface{}{
		"id":          u.ID,
		"title":       u.Title,
		"entry":       strings.Join(u.Entries, " "),
		"description": u.Description,
		"tag":         u.Tag,
	})
	return
}

// BeforeUpdate makes sure description and tag are always updated
func (u *Library) BeforeUpdate(tx *gorm.DB) (err error) {
	if u.Description == "" {
		u.Description = " "
	}

	if u.Tag == "" {
		u.Tag = " "
	}

	return
}

// AfterUpdate hook
func (u *Library) AfterUpdate(tx *gorm.DB) (err error) {

	tx.Exec(`
	UPDATE library_q
	SET title = @title, entry = @entry, [description] = @description, tag = @tag
	WHERE id = @id
	`, map[string]interface{}{
		"id":          u.ID,
		"title":       u.Title,
		"entry":       strings.Join(u.Entries, " "),
		"description": u.Description,
		"tag":         u.Tag,
	})
	return
}

// AfterDelete hook
func (u *Library) AfterDelete(tx *gorm.DB) (err error) {
	tx.Exec(`
	DELETE FROM library_q
	WHERE id = ?
	`, u.ID)
	return
}
