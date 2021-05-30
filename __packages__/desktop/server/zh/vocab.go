package zh

// Vocab represents MDBG CEDict and extras
type Vocab struct {
	Simplified  string
	Traditional string
	Pinyin      string
	English     string
	Frequency   float64
	Source      string
}
