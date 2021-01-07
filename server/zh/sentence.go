package zh

// Sentence represents Tatoeba sentence repository
type Sentence struct {
	ID        int `gorm:"primaryKey"`
	Chinese   string
	Pinyin    string
	English   string
	Frequency float64
	Level     float64
}
