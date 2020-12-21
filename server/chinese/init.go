package chinese

import (
	"log"
	"path"

	"github.com/zhquiz/go-server/shared"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// DB holds storage for current DB
type DB struct {
	Current *gorm.DB
}

// Connect connects to the database
func Connect() DB {
	db, err := gorm.Open(sqlite.Open(path.Join(shared.Paths().Dir, "assets", "zh.db")+"?mode=ro"), &gorm.Config{})
	if err != nil {
		log.Fatalln(err)
	}

	return DB{
		Current: db,
	}
}
