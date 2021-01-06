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
		&Quiz{},
		&Extra{},
	)

	if r := output.Current.Raw("SELECT Name FROM sqlite_temp_master WHERE type='table' AND name='quiz_q'").First(&struct {
		Name string
	}{}); errors.Is(r.Error, gorm.ErrRecordNotFound) {
		output.Current.Exec(`
		CREATE VIRTUAL TABLE quiz_q USING fts5(
			[id],
			[entry],
			[type],
			[direction],
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
				tx.Exec(`
				INSERT INTO quiz_q (id, [entry], [type], [direction])
				VALUES (@id, @entry, @type, @direction)
				`, map[string]interface{}{
					"id":        q.ID,
					"entry":     parseChinese(q.Entry),
					"type":      q.Type,
					"direction": q.Direction,
				})
			}

			return nil
		})
	}

	output.Current.Exec(`
	CREATE VIRTUAL TABLE IF NOT EXISTS extra_q USING fts5(
		[id],
		[chinese],
		[pinyin],
		[english],
		[description],
		[tag]
	);
	`)

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
