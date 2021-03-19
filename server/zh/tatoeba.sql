CREATE TABLE "tatoeba" (
	id INT PRIMARY KEY,
	chinese TEXT NOT NULL UNIQUE,
	english TEXT NOT NULL,
	frequency FLOAT,
	"level" FLOAT,
	chinese_norm TEXT AS (jieba_search(chinese))
);

CREATE INDEX idx_tatoeba_frequency ON "tatoeba"(frequency);
CREATE INDEX idx_tatoeba_level ON "tatoeba"("level");

CREATE VIRTUAL TABLE tatoeba_q USING fts5 (
	chinese,
	english,
	content = 'tatoeba',
	content_rowid = 'id',
	tokenize = 'porter unicode61'
);
CREATE TRIGGER t_tatoeba_insert
AFTER
INSERT ON "tatoeba" BEGIN
INSERT INTO tatoeba_q (rowid, chinese, english)
VALUES (
		new.id,
		new.chinese_norm,
		new.english
	);
END;
CREATE TRIGGER t_tatoeba_delete
AFTER DELETE ON "tatoeba" BEGIN
INSERT INTO tatoeba_q (tatoeba_q, chinese, english)
VALUES (
		'delete',
		old.chinese_norm,
		old.english
	);
END;
CREATE TRIGGER t_tatoeba_update
AFTER
UPDATE ON "tatoeba" BEGIN
INSERT INTO tatoeba_q (
		tatoeba_q,
		chinese,
		english
	)
VALUES (
		'delete',
		old.chinese_norm,
		old.english
	);
INSERT INTO tatoeba_q (rowid, chinese, english)
VALUES (
		new.id,
		new.chinese_norm,
		new.english
	);
END;
