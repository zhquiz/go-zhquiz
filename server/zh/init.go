package zh

import (
	"database/sql"
	"log"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/mattn/go-sqlite3"
	"github.com/wangbin/jiebago"
	"github.com/zhquiz/go-zhquiz/shared"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var jieba jiebago.Segmenter

// Connect connects to the database
func Connect() *gorm.DB {
	db, err := gorm.Open(&sqlite.Dialector{
		DriverName: "sqlite_custom",
		DSN:        path.Join(shared.ExecDir, "assets", "zh.db") + "?mode=ro",
	}, &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})

	if err != nil {
		log.Fatalln(err)
	}

	return db
}

// RegisterSQLiteCustom registers custom SQLite driver
func RegisterSQLiteCustom(driverName string) string {
	jieba.LoadDictionary(filepath.Join(shared.ExecDir, "assets", "dict.txt"))

	sql.Register(driverName, &sqlite3.SQLiteDriver{
		ConnectHook: func(conn *sqlite3.SQLiteConn) error {
			if err := conn.RegisterFunc("norm_pinyin", func(pin1yin1 string) string {
				out := make([]string, 0)
				re := regexp.MustCompile("\\d+$")

				for _, c := range strings.Split(pin1yin1, " ") {
					out = append(out, re.ReplaceAllString(c, ""))
				}

				return strings.Join(out, " ")
			}, true); err != nil {
				return err
			}

			if err := conn.RegisterFunc("jieba_search", func(s string) string {
				out := make([]string, 0)
				func(ch <-chan string) {
					for word := range ch {
						out = append(out, word)
					}
				}(jieba.CutForSearch(s, true))

				if len(out) == 0 {
					out = append(out, s)
				}

				return strings.Join(out, " ")
			}, true); err != nil {
				return err
			}
			return nil
		},
	})

	return driverName
}
