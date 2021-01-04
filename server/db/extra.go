package db

import "time"

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
