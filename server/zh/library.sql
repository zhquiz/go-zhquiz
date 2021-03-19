CREATE TABLE library (
		id 								INTEGER PRIMARY KEY,
		title 						TEXT NOT NULL UNIQUE,
		entries 					JSON NOT NULL,
		description 			TEXT NULL DEFAULT '',
		type 							TEXT NOT NULL,
		tag 							TEXT NOT NULL DEFAULT '',
		title_norm				TEXT AS (jieba_search(title)),
		description_norm	TEXT AS (jieba_search(description))
	);
CREATE VIRTUAL TABLE library_q USING fts5 (
		title,
		description,
		tag,
		content = 'library',
		content_rowid = 'id',
		tokenize = 'porter unicode61'
	);
CREATE TRIGGER t_library_insert
	AFTER
	INSERT ON "library" BEGIN
	INSERT INTO library_q (rowid, title, description, tag)
	VALUES (
			new.id,
			new.title_norm,
			new.description_norm,
			new.tag
		);
	END;
CREATE TRIGGER t_library_delete
	AFTER DELETE ON "library" BEGIN
	INSERT INTO library_q (library_q, title, description, tag)
	VALUES (
			'delete',
			old.title_norm,
			old.description_norm,
			old.tag
		);
	END;
CREATE TRIGGER t_library_update
	AFTER
	UPDATE ON "library" BEGIN
	INSERT INTO library_q (
			library_q,
			title,
			description,
			tag
		)
	VALUES (
			'delete',
			old.title_norm,
			old.description_norm,
			old.tag
		);
	INSERT INTO library_q (rowid, title, description, tag)
	VALUES (
			new.id,
			new.title_norm,
			new.description_norm,
			new.tag_norm
		);
	END;
