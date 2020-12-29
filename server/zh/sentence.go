package zh

// Sentence represents Tatoeba sentence repository
type Sentence struct {
	ID        int `gorm:"primaryKey"`
	Chinese   string
	Pinyin    string
	English   string
	Frequency float64
	Level     int

	Token []Token `gorm:"many2many:sentence_token"`
	Tag   []Tag   `gorm:"many2many:sentence_tag"`
}

// SentenceToken is joint Table for Sentence-Token
// @internal
type SentenceToken struct {
	SentenceID int    `gorm:"primaryKey"`
	Entry      string `gorm:"primaryKey"`
	Sentence   Sentence
}

// SentenceTag is joint Table for Sentence-Tag
// @internal
type SentenceTag struct {
	SentenceID int `gorm:"primaryKey"`
	TagID      int `gorm:"primaryKey"`
	Sentence   Sentence
	Tag        Tag
}
