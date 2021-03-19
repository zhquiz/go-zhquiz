CREATE TABLE token_sub (
	parent TEXT NOT NULL REFERENCES "token",
	child TEXT NOT NULL REFERENCES "token",
	PRIMARY KEY (parent, child)
);
CREATE TABLE token_sup (
	parent TEXT NOT NULL REFERENCES "token",
	child TEXT NOT NULL REFERENCES "token",
	PRIMARY KEY (parent, child)
);
CREATE TABLE token_var (
	parent TEXT NOT NULL REFERENCES "token",
	child TEXT NOT NULL REFERENCES "token",
	PRIMARY KEY (parent, child)
);

CREATE TABLE "token" (
	"id" INTEGER PRIMARY KEY,
	"entry" TEXT NOT NULL UNIQUE,
	"frequency" FLOAT,
	"hanzi_level" INT,
	"vocab_level" INT,
	"pinyin" TEXT,
	"english" TEXT,
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
	content = 'token',
	content_rowid = 'id',
	tokenize = 'porter unicode61'
);

CREATE TRIGGER t_token_insert
AFTER
INSERT ON token BEGIN
INSERT INTO token_q (rowid, pinyin, english, tag)
VALUES (
		new.id,
		new.pinyin_norm,
		new.english_norm,
		new.tag
	);
END;

CREATE TRIGGER t_token_delete
AFTER DELETE ON token BEGIN
INSERT INTO token_q (token_q, pinyin, english, tag)
VALUES (
		'delete',
		old.pinyin_norm,
		old.english_norm,
		old.tag
	);
END;

CREATE TRIGGER t_token_update
AFTER
UPDATE ON token BEGIN
INSERT INTO token_q (token_q, pinyin, english, tag)
VALUES (
		'delete',
		old.pinyin_norm,
		old.english_norm,
		old.tag
	);
INSERT INTO token_q (rowid, pinyin, english, tag)
VALUES (
		new.id,
		new.pinyin_norm,
		new.english_norm,
		new.tag
	);
END;