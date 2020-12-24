package zh

import (
	"log"
	"path"

	"database/sql"

	// go sqlite3
	_ "github.com/mattn/go-sqlite3"
	"github.com/zhquiz/go-server/shared"
)

// DB holds storage for current DB
type DB struct {
	Current *sql.DB
}

// Connect connects to the database
func Connect() DB {
	db, err := sql.Open("sqlite3", path.Join(shared.Paths().Dir, "assets", "zh.db")+"?mode=ro")
	if err != nil {
		log.Fatalln(err)
	}

	return DB{
		Current: db,
	}
}
