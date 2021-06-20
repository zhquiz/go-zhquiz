import sqlite3
import requests
import bz2
import tarfile
import jieba
from regex import regex

if __name__ == "__main__":
    sql_tmp = sqlite3.connect("generated/tatoeba-1.db")

    sql_tmp.executescript(
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

            sql_tmp.execute(
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

        sql_tmp.commit()

    r = requests.get(
        "https://downloads.tatoeba.org/exports/per_language/eng/eng_sentences.tsv.bz2"
    )

    with open("generated/eng_sentences.tsv.bz2", "wb") as f:
        f.write(r.content)

    with bz2.open("generated/eng_sentences.tsv.bz2", "rb") as zf:
        for row in zf.read().decode("utf-8").splitlines():
            cols = row.split("\t")

            sql_tmp.execute(
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

        sql_tmp.commit()

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

            sql_tmp.execute(
                """
            INSERT OR REPLACE INTO "link" ("id1", "id2")
            VALUES (?, ?)
            """,
                (
                    int(cols[0]),
                    int(cols[1]),
                ),
            )

        sql_tmp.commit()

    sql_level = sqlite3.connect("../desktop/assets/zhlevel.db")

    def find_level(s: str) -> int:
        for r in sql_level.execute(
            "SELECT vLevel FROM zhlevel WHERE entry = ? LIMIT 1", (s,)
        ):
            if r[0]:
                return r[0]

        return 100

    sql_out = sqlite3.connect("generated/tatoeba.db")
    sql_out.executescript(
        """
    CREATE TABLE IF NOT EXISTS "cmn_eng" (
        "cmn"     TEXT NOT NULL PRIMARY KEY,
        "eng"     TEXT NOT NULL,
        "level"   FLOAT NOT NULL
    );
    """
    )

    for r in sql_tmp.execute(
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
        levels = list(
            map(
                find_level,
                filter(
                    lambda x: regex.search(r"\p{Han}", x), jieba.cut_for_search(r[0])
                ),
            )
        )
        lv = 100
        if len(levels):
            lv = sum(levels) / len(levels)

        sql_out.execute(
            """
        INSERT INTO "cmn_eng" (cmn, eng, "level")
        VALUES (?, ?, ?)
        ON CONFLICT DO NOTHING
        """,
            (r[0], r[1], lv),
        )

    sql_out.commit()
    sql_out.close()

    sql_level.close()
    sql_tmp.close()
