package db

import (
	"log"

	"github.com/zhquiz/go-zhquiz/shared"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// DB is the storage for current DB
type DB struct {
	Current *gorm.DB
}

// Connect connects to DATABASE_URL
func Connect() DB {
	output := DB{}

	db, err := gorm.Open(sqlite.Open(shared.DatabaseURL()), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		log.Fatalln(err)
	}

	output = DB{
		Current: db,
	}

	output.Current.AutoMigrate(
		&User{},
		&Tag{},
		&Quiz{},
		&QuizTag{},
		// &Entry{},
		// &EntryTag{},
		// &EntryItem{},
		// &Preset{},
		&Extra{},
	)

	return output
}
