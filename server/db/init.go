package db

import (
	"errors"
	"log"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/wangbin/jiebago"
	"github.com/zhquiz/go-zhquiz/server/zh"
	"github.com/zhquiz/go-zhquiz/shared"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var jieba jiebago.Segmenter
var zhDB zh.DB

// DB is the storage for current DB
type DB struct {
	Current *gorm.DB
}

// Connect connects to DATABASE_URL
func Connect() DB {
	jieba.LoadDictionary(filepath.Join(shared.ExecDir, "assets", "dict.txt"))
	zhDB = zh.Connect()

	output := DB{}

	db, err := gorm.Open(sqlite.Open(filepath.Join(shared.UserDataDir(), "data.db")), &gorm.Config{
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
		&Quiz{},
		&Extra{},
		&Library{},
		&Sentence{},
	)

	var count int64
	output.Current.Find(&User{}).Count(&count)
	if count == 0 {
		if r := output.Current.Create(&User{}); r.Error != nil {
			log.Fatalln(r.Error)
		}
	}

	if r := output.Current.Raw("SELECT Name FROM sqlite_master WHERE type='table' AND name='quiz_q'").First(&struct {
		Name string
	}{}); r.Error != nil {
		if errors.Is(r.Error, gorm.ErrRecordNotFound) {
			output.Current.Exec(`
			CREATE VIRTUAL TABLE quiz_q USING fts5(
				[id],
				[entry],
				[level],
				[pinyin],
				[english],
				[description],
				[tag]
			);
			`)

			var quizzes []Quiz
			output.Current.Find(&quizzes)

			output.Current.Transaction(func(tx *gorm.DB) error {
				for _, q := range quizzes {
					q.Create(tx)
				}

				return nil
			})
		} else {
			log.Fatalln(r.Error)
		}
	}

	if r := output.Current.Raw("SELECT Name FROM sqlite_master WHERE type='table' AND name='extra_q'").First(&struct {
		Name string
	}{}); r.Error != nil {
		if errors.Is(r.Error, gorm.ErrRecordNotFound) {
			output.Current.Exec(`
			CREATE VIRTUAL TABLE extra_q USING fts5(
				[id],
				[chinese],
				[pinyin],
				[english],
				[type],
				[description],
				[tag]
			);
			`)

			var extras []Extra
			output.Current.Find(&extras)

			output.Current.Transaction(func(tx *gorm.DB) error {
				for _, ex := range extras {
					ex.Create(tx)
				}

				return nil
			})
		} else {
			log.Fatalln(r.Error)
		}
	}

	if r := output.Current.Raw("SELECT Name FROM sqlite_master WHERE type='table' AND name='library_q'").First(&struct {
		Name string
	}{}); r.Error != nil {
		if errors.Is(r.Error, gorm.ErrRecordNotFound) {
			output.Current.Exec(`
			CREATE VIRTUAL TABLE library_q USING fts5(
				[id],
				[title],
				[entry],
				[description],
				[tag]
			);
			`)

			if e := output.Current.Transaction(func(tx *gorm.DB) error {
				var libs []map[string]interface{}
				if r := zhDB.Current.Raw("SELECT title, entries FROM library").Find(&libs); r.Error != nil {
					log.Fatalln(r.Error)
				}

				for _, a := range libs {
					lib := Library{
						ID:    " " + a["title"].(string),
						Title: a["title"].(string),
					}

					if e := lib.Entries.Scan(a["entries"]); e != nil {
						return e
					}

					if e := lib.Create(tx); e != nil {
						return e
					}
				}

				return nil
			}); e != nil {
				panic(e)
			}
		} else {
			log.Fatalln(r.Error)
		}
	}

	return output
}

func parseChinese(s string) string {
	out := make([]string, 0)
	func(ch <-chan string) {
		for word := range ch {
			out = append(out, word)
		}
	}(jieba.CutAll(s))

	return strings.Join(out, " ")
}

func parsePinyin(s string) string {
	out := make([]string, 0)
	re := regexp.MustCompile("\\d+$")

	for _, c := range strings.Split(s, " ") {
		out = append(out, re.ReplaceAllString(c, ""))
	}

	return strings.Join(out, " ")
}
