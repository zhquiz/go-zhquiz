// +build try

package zh

import (
	"fmt"
	"log"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func TestCreate(t *testing.T) {
	RegisterSQLiteCustom("sqlite_custom")

	db, err := gorm.Open(&sqlite.Dialector{
		DriverName: "sqlite_custom",
		DSN:        ":memory:",
	}, &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		log.Fatalln(err)
	}

	if r := db.Exec(`
	CREATE TABLE extra (
		id								INTEGER PRIMARY KEY,
		chinese						TEXT NOT NULL,
		alt								TEXT,		-- space separated
		pinyin						TEXT,
		pinyin_norm				TEXT AS (norm_pinyin(pinyin)) STORED,
		type							TEXT,
		english						TEXT,
		description				TEXT,
		description_norm	TEXT AS (jieba_search(description)),
		tag								TEXT		-- space separated
	);

	CREATE UNIQUE INDEX idx_extra_chinese ON extra(chinese);
	CREATE INDEX idx_extra_alt ON extra(alt);
	CREATE INDEX idx_extra_pinyin ON extra(pinyin);
	CREATE INDEX idx_extra_pinyin_norm ON extra(pinyin_norm);

	CREATE VIRTUAL TABLE extra_q USING fts5 (
		alt,
		english,
		description,
		tag,
		content='extra',
		content_rowid='id',
		tokenize='porter unicode61'
	);

	CREATE TRIGGER t_extra_insert AFTER INSERT ON extra BEGIN
		INSERT INTO extra_q (
			rowid,
			alt,
			english,
			description,
			tag
		)
		VALUES (
			new.id,
			new.alt,
			new.english,
			new.description_norm,
			new.tag
		);
	END;
	CREATE TRIGGER t_extra_delete AFTER DELETE ON extra BEGIN
		INSERT INTO extra_q (
			extra_q,
			alt,
			english,
			description,
			tag
		)
		VALUES (
			'delete',
			old.alt,
			old.english,
			old.description_norm,
			old.tag
		);
	END;
	CREATE TRIGGER t_extra_update AFTER UPDATE ON extra BEGIN
		INSERT INTO extra_q (
			extra_q,
			rowid,
			alt,
			english,
			description,
			tag
		)
		VALUES (
			'delete',
			old.id,
			old.alt,
			old.english,
			old.description_norm,
			old.tag
		);

		INSERT INTO extra_q (
			rowid,
			alt,
			english,
			description,
			tag
		)
		VALUES (
			new.id,
			new.alt,
			new.english,
			new.description_norm,
			new.tag
		);
	END;

	INSERT INTO extra (chinese, pinyin, description)
	VALUES ('你好', 'ni3 hao3', '老师你好！');
	`); r.Error != nil {
		log.Fatalln(r.Error)
	}

	out := make(map[string]interface{})

	if r := db.Raw(`
	SELECT chinese, pinyin, description_norm FROM extra WHERE id IN (
		SELECT id FROM extra_q(jieba_search('好你'))
	)
	`).First(&out); r.Error != nil {
		log.Fatalln(r.Error)
	}

	fmt.Println(out)
}
