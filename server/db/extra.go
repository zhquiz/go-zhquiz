package db

import (
	"gorm.io/gorm"
)

// Extra is user database model for Extra
type Extra struct {
	gorm.Model

	UserID uint `gorm:"index:idx_extra_user_chinese,unique;not null"`
	User   User `gorm:"foreignKey:UserID"`

	Chinese string `gorm:"index:idx_extra_user_chinese,unique;not null"`
	Pinyin  string `gorm:"not null"`
	English string `gorm:"not null"`
}
