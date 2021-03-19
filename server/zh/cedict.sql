CREATE TABLE IF NOT EXISTS "cedict" (
	"id" INTEGER,
	"simplified" TEXT NOT NULL,
	"traditional" TEXT,
	"pinyin" TEXT NOT NULL,
	"pinyin_norm" TEXT AS (norm_pinyin(pinyin)) STORED,
	"english" TEXT NOT NULL,
	"frequency" FLOAT,
	PRIMARY KEY("id")
);
CREATE INDEX "idx_cedict_frequency" ON "cedict" ("frequency");
CREATE UNIQUE INDEX "idx_unique_cedict" ON "cedict" ("simplified", "traditional", "pinyin");
CREATE INDEX "idx_cedict_simplified" ON "cedict" ("simplified");
CREATE INDEX "idx_cedict_traditional" ON "cedict" ("traditional");
CREATE INDEX "idx_cedict_pinyin" ON "cedict" ("pinyin");
CREATE INDEX "idx_cedict_pinyin_norm" ON "cedict" ("pinyin_norm");

CREATE VIRTUAL TABLE cedict_q USING fts5 (
	english,
	content = 'cedict',
	content_rowid = 'id',
	tokenize = 'porter unicode61'
);

CREATE TRIGGER t_cedict_insert
AFTER
INSERT ON cedict BEGIN
INSERT INTO cedict_q (rowid, english)
VALUES (new.id, new.english);
END;

CREATE TRIGGER t_cedict_delete
AFTER DELETE ON cedict BEGIN
INSERT INTO cedict_q (cedict_q, english)
VALUES ('delete', old.english);
END;

CREATE TRIGGER t_cedict_update
AFTER
UPDATE ON cedict BEGIN
INSERT INTO cedict_q (cedict_q, english)
VALUES ('delete', old.english);
INSERT INTO cedict_q (rowid, english)
VALUES (new.id, new.english);
END;
