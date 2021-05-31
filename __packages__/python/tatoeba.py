import sqlite3
import requests
import bz2
import tarfile

if __name__ == "__main__":
    sql = sqlite3.connect("generated/tatoeba-1.db")

    sql.executescript(
        """
    CREATE TABLE IF NOT EXISTS "sentence" (
        "id"      INT NOT NULL PRIMARY KEY,
        "lang"    TEXT NOT NULL,
        "text"    TEXT NOT NULL
    );
    CREATE TABLE IF NOT EXISTS "link" (
        "id1"     INT NOT NULL,
        "id2"     INT NOT NULL,
        PRIMARY KEY ("id1", "id2")
    );
    """
    )

    r = requests.get(
        "https://downloads.tatoeba.org/exports/per_language/cmn/cmn_sentences.tsv.bz2"
    )

    with open("generated/cmn_sentences.tsv.bz2", "wb") as f:
        f.write(r.content)

    with bz2.open("generated/cmn_sentences.tsv.bz2", "rb") as zf:
        for row in zf.read().decode("utf-8").splitlines():
            cols = row.split("\t")

            sql.execute(
                """
            INSERT OR REPLACE INTO "sentence" ("id", "lang", "text")
            VALUES (?, ?, ?)
            """,
                (
                    int(cols[0]),
                    cols[1],
                    cols[2],
                ),
            )

        sql.commit()

    r = requests.get(
        "https://downloads.tatoeba.org/exports/per_language/eng/eng_sentences.tsv.bz2"
    )

    with open("generated/eng_sentences.tsv.bz2", "wb") as f:
        f.write(r.content)

    with bz2.open("generated/eng_sentences.tsv.bz2", "rb") as zf:
        for row in zf.read().decode("utf-8").splitlines():
            cols = row.split("\t")

            sql.execute(
                """
            INSERT OR REPLACE INTO "sentence" ("id", "lang", "text")
            VALUES (?, ?, ?)
            """,
                (
                    int(cols[0]),
                    cols[1],
                    cols[2],
                ),
            )

        sql.commit()

    r = requests.get("https://downloads.tatoeba.org/exports/links.tar.bz2")

    with open("generated/links.tar.bz2", "wb") as f:
        f.write(r.content)

    with tarfile.open("generated/links.tar.bz2", "r:bz2") as zf:
        zf.extractall("generated")

    with open("generated/links.csv") as zf:
        for row in zf:
            cols = row.split("\t")
            if len(cols) < 2:
                continue

            sql.execute(
                """
            INSERT OR REPLACE INTO "link" ("id1", "id2")
            VALUES (?, ?)
            """,
                (
                    int(cols[0]),
                    int(cols[1]),
                ),
            )

        sql.commit()

    sql2 = sqlite3.connect("generated/tatoeba.db")

    sql2.executescript(
        """
    CREATE TABLE IF NOT EXISTS "cmn_eng" (
        "cmn"     TEXT NOT NULL PRIMARY KEY,
        "eng"     TEXT NOT NULL
    );
    """
    )

    for r in sql.execute(
        """
    SELECT
        s2.text     cmn,
        s1.text     eng
    FROM sentence s1
    JOIN link t       ON t.id1 = s1.id
    JOIN sentence s2  ON t.id2 = s2.id
    WHERE s1.lang = 'eng' AND s2.lang = 'cmn'
    """
    ):
        sql2.execute(
            """
        INSERT INTO "cmn_eng" (cmn, eng)
        VALUES (?, ?)
        ON CONFLICT DO NOTHING
        """,
            (r[0], r[1]),
        )

    sql2.commit()
    sql2.close()
    sql.close()
