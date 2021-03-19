// +build !alter

package zh

import (
	"fmt"
	"log"
	"path"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func TestLibraryRead(t *testing.T) {
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

	out := make([]map[string]interface{}, 0)

	if r := db.Raw(`
	SELECT * FROM library WHERE id IN (
		SELECT rowid FROM library_q('HSK6')
	) LIMIT 10;
	`).Find(&out); r.Error != nil {
		log.Fatalln(r.Error)
	}

	fmt.Println(out)
}
