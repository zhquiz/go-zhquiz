package db

import "gorm.io/gorm"

// Sentence caches sentences from http://www.jukuu.com/search.php?q=%s
type Sentence struct {
	gorm.Model

	Chinese string `gorm:"uniqueIndex"`
	English string
}
