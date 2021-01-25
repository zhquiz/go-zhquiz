import sqlite3 from 'better-sqlite3'

async function main() {
  const db = sqlite3('../../data.db')

  db.exec(/* sql */ `
  ALTER TABLE extra RENAME TO extra1;
  `)

  // db.exec(/* sql */ `
  // CREATE VIRTUAL TABLE IF NOT EXISTS extra USING fts5(
  //   id,
  //   created_at  UNINDEXED,
  //   updated_at  UNINDEXED,
  //   chinese,
  //   word,
  //   pinyin,
  //   english,
  //   [type],
  //   [description],
  //   tag,
  //   tokenize = "unicode61 separators '01234567890'"
  // );
  // `)

  db.exec(/* sql */ `
  CREATE TABLE IF NOT EXISTS extra_q (
    id      TEXT PRIMARY KEY,
    chinese TEXT UNIQUE NOT NULL
  );

  CREATE TRIGGER IF NOT EXISTS t_extra_insert
    BEFORE INSERT ON extra
  BEGIN
    INSERT INTO extra_q (id, chinese)
    VALUES (NEW.id, NEW.chinese);
  END;

  CREATE TRIGGER IF NOT EXISTS t_extra_update
    BEFORE INSERT ON extra
  BEGIN
    UPDATE extra_q SET chinese = NEW.chinese
    WHERE id = NEW.id;
  END;
  `)

  db.close()
}

if (require.main === module) {
  main()
}
