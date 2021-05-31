import sqlite3
import requests
from zipfile import ZipFile
import re
from wordfreq import word_frequency

if __name__ == "__main__":
    r = requests.get(
        "https://www.mdbg.net/chinese/export/cedict/cedict_1_0_ts_utf-8_mdbg.zip"
    )

    with open("generated/cedict.zip", "wb") as f:
        f.write(r.content)

    with ZipFile("generated/cedict.zip") as zf:
        sql = sqlite3.connect("generated/cedict.db")

        sql.executescript(
            """
        CREATE TABLE IF NOT EXISTS cedict (
            simplified      TEXT NOT NULL,
            traditional     TEXT,
            pinyin          TEXT NOT NULL,
            english         TEXT NOT NULL,
            frequency       FLOAT
        );

        CREATE UNIQUE INDEX IF NOT EXISTS idx_u_cedict ON cedict (simplified, traditional, pinyin);
        """
        )

        for row in zf.read("cedict_ts.u8").decode("utf-8").splitlines():
            simplified = ""
            traditional = ""
            pinyin = ""
            english = ""

            if row.startswith("#"):
                continue

            m = re.match(r"(.+?)\s+", row)
            if not m:
                continue

            traditional = m.group(1)
            row = row[len(m.group(0)) :]

            m = re.match(r"(.+?)\s+", row)
            if not m:
                continue

            simplified = m.group(1)
            if simplified == traditional:
                traditional = None

            row = row[len(m.group(0)) :]

            m = re.match(r"\[([^\]]+?)\]\s+", row)
            if not m:
                continue

            pinyin = m.group(1)
            row = row[len(m.group(0)) :]

            m = re.match(r"/(.+)/", row)
            if not m:
                continue

            english = " / ".join(m.group(1).split("/"))

            sql.execute(
                """
            INSERT OR REPLACE INTO cedict (simplified, traditional, pinyin, english, frequency)
            VALUES (?, ?, ?, ?, ?)
            """,
                (
                    simplified,
                    traditional,
                    pinyin,
                    english,
                    word_frequency(simplified, "zh") * 10 ** 6,
                ),
            )

        sql.commit()
