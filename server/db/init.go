package db

import (
	"errors"
	"log"
	"path/filepath"

	"github.com/zhquiz/go-zhquiz/server/zh"
	"github.com/zhquiz/go-zhquiz/shared"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

// Connect connects to DATABASE_URL
func Connect() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(filepath.Join(shared.UserDataDir(), "data.db")), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		log.Fatalln(err)
	}

	db.AutoMigrate(
		&User{},
		&Tag{},
		&Quiz{},
		&Extra{},
		&Library{},
	)

	if r := db.FirstOrCreate(&User{
		ID:         1,
		Identifier: "-",
	}); r.Error != nil {
		panic(r.Error)
	}

	if r := db.First(&Library{}); r.Error != nil {
		if errors.Is(r.Error, gorm.ErrRecordNotFound) {
			zhDB := zh.Connect()

			libs := make([]zh.Library, 0)
			if r1 := zhDB.Find(&libs); r1.Error != nil {
				panic(r1.Error)
			}

			db.Transaction(func(db *gorm.DB) error {
				for _, lib := range libs {
					if r2 := db.Create(&Library{
						UserID:      1,
						Title:       lib.Title,
						Entries:     []string(lib.Entries),
						Type:        lib.Type,
						Tag:         []string(lib.Tag),
						Description: lib.Description,
					}); r2.Error != nil {
						return r2.Error
					}

					tags := make([]Tag, 0)

					for _, t := range lib.Tag {
						for _, ent := range lib.Entries {
							tags = append(tags, Tag{
								UserID: 1,
								Entry:  ent,
								Type:   lib.Type,
								Name:   t,
							})
						}
					}

					db.Clauses(clause.OnConflict{
						DoNothing: true,
					}).Create(&tags)
				}

				return nil
			})
		} else {
			panic(r.Error)
		}
	}

	return db
}
