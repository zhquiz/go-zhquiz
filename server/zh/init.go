package zh

import (
	"log"
	"path"

	"github.com/zhquiz/go-server/shared"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// DB holds storage for current DB
type DB struct {
	Current *gorm.DB
}

// Connect connects to the database
func Connect() DB {
	db, err := gorm.Open(sqlite.Open(path.Join(shared.ExecDir, "assets", "zh.db")+"?mode=ro"), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		log.Fatalln(err)
	}

	return DB{
		Current: db,
	}
}
