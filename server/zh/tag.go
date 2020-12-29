package zh

// Tag is the database model for tag
type Tag struct {
	ID   uint   `gorm:"primarykey"`
	Name string `gorm:"index:,unique;not null"`
}
