package zh

// Sentence represents Tatoeba sentence repository
type Sentence struct {
	ID        int `gorm:"primaryKey"`
	Chinese   string
	Pinyin    string
	English   string
	Frequency float64
	Level     float64

	Tag []Tag `gorm:"many2many:sentence_tag"`
}

// SentenceTag is joint Table for Sentence-Tag
// @internal
type SentenceTag struct {
	SentenceID int `gorm:"primaryKey"`
	TagID      int `gorm:"primaryKey"`
	Sentence   Sentence
	Tag        Tag
}
