package db

import (
	"strings"
	"time"

	"github.com/zhquiz/go-zhquiz/server/rand"
	"gorm.io/gorm"
)

// Library is user database model for Library
type Library struct {
	ID        string `gorm:"primarykey" json:"-"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	UserID string `gorm:"index:idx_library_user_title,unique" json:"-"`
	User   User   `gorm:"foreignKey:UserID" json:"-"`

	Title   string      `gorm:"index:idx_library_user_title,unique;not null" json:"title"`
	Entries StringArray `json:"entries"`
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
	INSERT INTO library_q (id, title, entry)
	VALUES (@id, @title, @entry)
	`, map[string]interface{}{
		"id":    u.ID,
		"title": u.Title,
		"entry": strings.Join(u.Entries, " "),
	})
	return
}

// AfterUpdate hook
func (u *Library) AfterUpdate(tx *gorm.DB) (err error) {
	tx.Exec(`
	UPDATE library_q
	SET title = @title, entry = @entry
	WHERE id = @id
	`, map[string]interface{}{
		"id":    u.ID,
		"title": u.Title,
		"entry": strings.Join(u.Entries, " "),
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
