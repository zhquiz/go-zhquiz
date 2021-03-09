package zh

// Tatoeba represents Tatoeba sentence repository
type Tatoeba struct {
	ID        int `gorm:"primaryKey"`
	Chinese   string
	Pinyin    string
	English   string
	Frequency float64
	Level     float64
}
