package db

import (
	"errors"
	"log"
	"path"
	"path/filepath"
	"strings"

	"github.com/zhquiz/zhquiz-desktop/shared"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
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
		&Quiz{},
		&ChineseBase{},
		&LibraryBase{},
	)

	// Create a user, if not exist yet
	func() {
		var nUser int64

		if r := db.Model(&User{}).Count(&nUser); r.Error != nil {
			panic(r.Error)
		}

		if nUser == 0 {
			if r := db.Create(&User{}); r.Error != nil {
				panic(r.Error)
			}
		}
	}()

	if r := db.Raw("SELECT Name FROM sqlite_master WHERE type='table' AND name='chinese'").First(&struct {
		Name string
	}{}); r.Error != nil {
		if errors.Is(r.Error, gorm.ErrRecordNotFound) {
			if r := db.Exec(`
			CREATE VIRTUAL TABLE chinese USING fts5 (
				id,
				source,
				[type],
				chinese,
				alt,
				pinyin,
				english,
				[description],
				[tag],
				tokenize = porter
			)
			`); r.Error != nil {
				panic(r.Error)
			}

			if e := db.Transaction(func(tx *gorm.DB) error {
				dbCedict, err := gorm.Open(sqlite.Open(path.Join(shared.ExecDir, "assets", "cedict.db") + "?mode=ro"))
				if err != nil {
					return err
				}
				rows, err := dbCedict.Raw(`
				SELECT
					simplified chinese,
					json_group_array(DISTINCT traditional) alt,
					json_group_array(DISTINCT pinyin) pinyin,
					json_group_array(DISTINCT json_each.value) english,
					frequency
				FROM cedict, json_each(english)
				GROUP BY simplified
				`).Rows()
				if err != nil {
					return err
				}
				defer rows.Close()

				items := make([]Chinese, 0)
				for rows.Next() {
					var c Chinese
					dbCedict.ScanRows(rows, &c)

					c.Type = "vocab"
					c.Source = "cedict"
					c.Alt = cleanStringArray(c.Alt)

					items = append(items, c)
				}

				if e := Chinese.BaseCreate(Chinese{}, tx, items...); e != nil {
					return e
				}

				return nil
			}); e != nil {
				panic(e)
			}

			if e := db.Transaction(func(tx *gorm.DB) error {
				dbCedict, err := gorm.Open(sqlite.Open(path.Join(shared.ExecDir, "assets", "junda.db") + "?mode=ro"))
				if err != nil {
					return err
				}
				rows, err := dbCedict.Raw(`
				SELECT
					[character] chinese,
					json_array(pinyin) pinyin,
					json_array(english) english,
					raw_freq frequency
				FROM hanzi
				`).Rows()
				if err != nil {
					return err
				}
				defer rows.Close()

				items := make([]Chinese, 0)
				for rows.Next() {
					var c Chinese
					dbCedict.ScanRows(rows, &c)

					c.Type = "hanzi"
					c.Source = "junda"
					c.Pinyin = strings.Split(c.Pinyin[0], "/")
					c.English = strings.Split(c.English[0], "/")

					items = append(items, c)
				}

				if e := Chinese.BaseCreate(Chinese{}, tx, items...); e != nil {
					return e
				}

				return nil
			}); e != nil {
				panic(e)
			}

			if e := db.Transaction(func(tx *gorm.DB) error {
				dbCedict, err := gorm.Open(sqlite.Open(path.Join(shared.ExecDir, "assets", "tatoeba.db") + "?mode=ro"))
				if err != nil {
					return err
				}
				rows, err := dbCedict.Raw(`
				SELECT
					cmn chinese,
					json_array(eng) english
				FROM cmn_eng
				`).Rows()
				if err != nil {
					return err
				}
				defer rows.Close()

				items := make([]Chinese, 0)
				for rows.Next() {
					var c Chinese
					dbCedict.ScanRows(rows, &c)

					c.Type = "sentence"
					c.Source = "tatoeba"

					items = append(items, c)
				}

				if e := Chinese.BaseCreate(Chinese{}, tx, items...); e != nil {
					return e
				}

				return nil
			}); e != nil {
				panic(e)
			}
		} else {
			log.Fatalln(r.Error)
		}
	}

	if r := db.Raw("SELECT Name FROM sqlite_master WHERE type='table' AND name='library'").First(&struct {
		Name string
	}{}); r.Error != nil {
		if errors.Is(r.Error, gorm.ErrRecordNotFound) {
			if e := db.Transaction(func(tx *gorm.DB) error {
				if r := tx.Exec(`
				CREATE VIRTUAL TABLE library USING fts5 (
					id,
					title,
					entry,
					[description],
					[tag],
					[type],
					tokenize = porter
				)
				`); r.Error != nil {
					return r.Error
				}

				return nil
			}); e != nil {
				panic(e)
			}
		} else {
			log.Fatalln(r.Error)
		}
	}

	return db
}

func cleanStringArray(s []string) []string {
	cleaned := make([]string, 0)
	for _, r := range s {
		if r != "" {
			cleaned = append(cleaned, r)
		}
	}

	return cleaned
}
