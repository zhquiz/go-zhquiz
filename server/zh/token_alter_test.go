// +build alter

package zh

import (
	"encoding/json"
	"log"
	"path"
	"strings"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func TestTokenAlter(t *testing.T) {
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

	rows, e := db.Raw(`
	SELECT entries, tag FROM library
	`).Rows()

	if e != nil {
		log.Fatalln(rows)
	}

	entryToTag := make(map[string][]string)

	for rows.Next() {
		entriesString := ""
		tagsString := ""

		if e := rows.Scan(&entriesString, &tagsString); e != nil {
			log.Fatalln(e)
		}

		entries := make([]string, 0)
		if e := json.Unmarshal([]byte(entriesString), &entries); e != nil {
			log.Fatalln(e)
		}

		tags := make([]string, 0)
		if e := json.Unmarshal([]byte(tagsString), &tags); e != nil {
			log.Fatalln(e)
		}

		for _, ent := range entries {
			if entryToTag[ent] == nil {
				entryToTag[ent] = make([]string, 0)
			}

			for _, t := range tags {
				entryToTag[ent] = append(entryToTag[ent], t)
			}
		}
	}

	if r := db.Exec(`
	DROP TABLE token;
	DROP TABLE token_q;

	CREATE TABLE "token" (
		"id" INTEGER PRIMARY KEY,
		"entry"	TEXT NOT NULL UNIQUE,
		"frequency"	FLOAT,
		"hanzi_level"	INT,
		"vocab_level"	INT,
		"pinyin"	TEXT,
		"english"	TEXT,
		pinyin_norm TEXT AS (norm_pinyin(COALESCE(pinyin, ''))),
		english_norm TEXT AS (jieba_search(COALESCE(english, ''))),
		tag TEXT
	);
	CREATE INDEX "idx_token_frequency" ON "token" ("frequency");
	CREATE INDEX "idx_token_hanzi_level" ON "token" ("hanzi_level");
	CREATE INDEX "idx_token_vocab_level" ON "token" ("vocab_level");

	CREATE VIRTUAL TABLE token_q USING fts5 (
		pinyin,
		english,
		tag,
		content='token',
		content_rowid='id',
		tokenize='porter unicode61'
	);

	CREATE TRIGGER t_token_insert AFTER INSERT ON token BEGIN
		INSERT INTO token_q (
			rowid,
			pinyin,
			english,
			tag
		)
		VALUES (
			new.id,
			new.pinyin_norm,
			new.english_norm,
			new.tag
		);
	END;
	CREATE TRIGGER t_token_delete AFTER DELETE ON token BEGIN
		INSERT INTO token_q (
			token_q,
			pinyin,
			english,
			tag
		)
		VALUES (
			'delete',
			old.pinyin_norm,
			old.english_norm,
			old.tag
		);
	END;
	CREATE TRIGGER t_token_update AFTER UPDATE ON token BEGIN
		INSERT INTO token_q (
			token_q,
			pinyin,
			english,
			tag
		)
		VALUES (
			'delete',
			old.pinyin_norm,
			old.english_norm,
			old.tag
		);

		INSERT INTO token_q (
			rowid,
			pinyin,
			english,
			tag
		)
		VALUES (
			new.id,
			new.pinyin_norm,
			new.english_norm,
			new.tag
		);
	END;
	`); r.Error != nil {
		log.Fatalln(r.Error)
	}

	if e := db.Transaction(func(tx *gorm.DB) error {
		out := make([]map[string]interface{}, 0)
		if r := tx.Raw(`
		SELECT
			entry,
			frequency,
			hanzi_level,
			vocab_level,
			pinyin,
			english
		FROM token_;
		`).Find(&out); r.Error != nil {
			return r.Error
		}

		for _, it := range out {
			tags := entryToTag[it["entry"].(string)]
			if tags == nil {
				tags = make([]string, 0)
			}

			it["tag"] = strings.Join(tags, " ")

			if r := tx.Exec(`
			INSERT INTO token (
				entry,
				frequency,
				hanzi_level,
				vocab_level,
				pinyin,
				english,
				tag
			)
			VALUES (
				@entry,
				@frequency,
				@hanzi_level,
				@vocab_level,
				@pinyin,
				@english,
				@tag
			);
			`, it); r.Error != nil {
				return r.Error
			}
		}

		return nil
	}); e != nil {
		log.Fatalln(e)
	}
}
