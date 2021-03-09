package zh

import (
	"log"
	"path"

	"github.com/zhquiz/go-zhquiz/shared"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// Connect connects to the database
func Connect() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(path.Join(shared.ExecDir, "assets", "zh.db")+"?mode=ro"), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		log.Fatalln(err)
	}

	return db
}
