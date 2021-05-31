package db

import (
	"errors"
	"log"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/wangbin/jiebago"
	"github.com/zhquiz/zhquiz-desktop/shared"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var jieba jiebago.Segmenter

// Connect connects to DATABASE_URL
func Connect() *gorm.DB {
	jieba.LoadDictionary(filepath.Join(shared.ExecDir, "assets", "dict.txt"))

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
			if e := db.Transaction(func(tx *gorm.DB) error {
				if r := tx.Exec(`
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
					return r.Error
				}

				dbCedict, err := gorm.Open(sqlite.Open(path.Join(shared.ExecDir, "assets", "cedict.db") + "?mode=ro"))
				if err != nil {
					return err
				}
				if r := dbCedict.Raw("SELECT simplified, traditional, pinyin, english FROM cedict"); r.Error != nil {
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

func parseChinese(s string) string {
	out := make([]string, 0)
	func(ch <-chan string) {
		for word := range ch {
			out = append(out, word)
		}
	}(jieba.CutAll(s))

	if len(out) == 0 {
		out = append(out, s)
	}

	return strings.Join(out, " ")
}

func parsePinyin(s string) string {
	out := make([]string, 0)
	re := regexp.MustCompile(`\d+$`)

	for _, c := range strings.Split(s, " ") {
		out = append(out, re.ReplaceAllString(c, ""))
	}

	return strings.Join(out, " ")
}
