package zh

// Token represents Token frequency and parts table
type Token struct {
	Entry       string `gorm:"primaryKey"`
	Pinyin      string
	English     string
	Frequency   float64
	HanziLevel  int
	VocabLevel  int
	Description string
	Tag         string

	Sub      []Token `gorm:"many2many:token_sub"`
	Sup      []Token `gorm:"many2many:token_sup"`
	Variants []Token `gorm:"many2many:token_var"`
}

// TokenSub is self-referential table for sub
// @internal
type TokenSub struct {
	Parent      string `gorm:"primaryKey"`
	Child       string `gorm:"primaryKey"`
	ParentToken Token  `gorm:"foreignKey:Parent"`
	ChildToken  Token  `gorm:"foreignKey:Child"`
}

// TokenSup is self-referential table for sup
// @internal
type TokenSup struct {
	Parent      string `gorm:"primaryKey"`
	Child       string `gorm:"primaryKey"`
	ParentToken Token  `gorm:"foreignKey:Parent"`
	ChildToken  Token  `gorm:"foreignKey:Child"`
}

// TokenVar is self-referential table for variants
// @internal
type TokenVar struct {
	Parent      string `gorm:"primaryKey"`
	Child       string `gorm:"primaryKey"`
	ParentToken Token  `gorm:"foreignKey:Parent"`
	ChildToken  Token  `gorm:"foreignKey:Child"`
}
