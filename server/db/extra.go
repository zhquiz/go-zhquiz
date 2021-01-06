package db

import (
	"time"

	"gorm.io/gorm"
)

// Extra is user database model for Extra
type Extra struct {
	ID        string `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time

	UserID string `gorm:"index:idx_extra_user_chinese,unique;not null"`
	User   User   `gorm:"foreignKey:UserID"`

	Chinese string `gorm:"index:idx_extra_user_chinese,unique;not null"`
	Pinyin  string
	English string
}

// AfterCreate hook
func (u *Extra) AfterCreate(tx *gorm.DB) (err error) {
	tx.Exec(`
	INSERT INTO extra_q (id, chinese, pinyin, english)
	VALUES (@id, @chinese, @pinyin, @english)
	`, map[string]interface{}{
		"id":      u.ID,
		"chinese": parseChinese(u.Chinese),
		"pinyin":  parsePinyin(u.Pinyin),
		"english": u.English,
	})
	return
}

// AfterUpdate hook
func (u *Extra) AfterUpdate(tx *gorm.DB) (err error) {
	tx.Exec(`
	UPDATE extra_q
	SET chinese = @chinese, pinyin = @pinyin, english = @english
	WHERE id = @id
	`, map[string]interface{}{
		"id":      u.ID,
		"chinese": parseChinese(u.Chinese),
		"pinyin":  parsePinyin(u.Pinyin),
		"english": u.English,
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
