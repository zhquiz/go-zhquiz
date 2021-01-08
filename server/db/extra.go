package db

import (
	"time"

	"gorm.io/gorm"
)

// Extra is user database model for Extra
type Extra struct {
	ID        string `gorm:"primarykey" json:"id"`
	CreatedAt time.Time
	UpdatedAt time.Time

	UserID string `gorm:"index:idx_extra_user_chinese,unique;not null" json:"-"`
	User   User   `gorm:"foreignKey:UserID" json:"-"`

	Chinese     string `gorm:"index:idx_extra_user_chinese,unique;not null" json:"chinese"`
	Pinyin      string `json:"pinyin"`
	English     string `json:"english"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Tag         string `json:"tag"`
}

// AfterCreate hook
func (u *Extra) AfterCreate(tx *gorm.DB) (err error) {
	tx.Exec(`
	INSERT INTO extra_q (id, chinese, pinyin, english, description, tag)
	VALUES (@id, @chinese, @pinyin, @english, @description, @tag)
	`, map[string]interface{}{
		"id":          u.ID,
		"chinese":     parseChinese(u.Chinese),
		"pinyin":      parsePinyin(u.Pinyin),
		"english":     u.English,
		"description": parseChinese(u.Description),
		"tag":         u.Tag,
	})
	return
}

// AfterUpdate hook
func (u *Extra) AfterUpdate(tx *gorm.DB) (err error) {
	tx.Exec(`
	UPDATE extra_q
	SET chinese = @chinese, pinyin = @pinyin, english = @english, [description] = @description, tag = @tag
	WHERE id = @id
	`, map[string]interface{}{
		"id":          u.ID,
		"chinese":     parseChinese(u.Chinese),
		"pinyin":      parsePinyin(u.Pinyin),
		"english":     u.English,
		"description": parseChinese(u.Description),
		"tag":         u.Tag,
	})
	return
}

// AfterDelete hook
func (u *Extra) AfterDelete(tx *gorm.DB) (err error) {
	tx.Exec(`
	DELETE FROM extra_q
	WHERE id = ?
	`, u.ID)
	return
}
