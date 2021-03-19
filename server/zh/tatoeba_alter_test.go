// +build alter

package zh

import (
	"log"
	"path"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func TestTatoebaAlter(t *testing.T) {
	RegisterSQLiteCustom("sqlite_custom")

	db, err := gorm.Open(&sqlite.Dialector{
		DriverName: "sqlite_custom",
		DSN:        path.Join("../..", "assets", "zh.db"),
	}, &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		log.Fatalln(err)
	}

	if r := db.Exec(`
	INSERT OR REPLACE INTO tatoeba (id, chinese, english, frequency, level)
	SELECT id, chinese, english, frequency, level FROM tatoeba
	`); r.Error != nil {
		log.Fatalln(r.Error)
	}
}
