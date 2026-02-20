import os

from tools import read_csv, google_translate, write_csv, openai_query
from tqdm import tqdm

file_path = "../public/docs/EiP.csv"

entries, fieldnames = read_csv(file_path)

_TRANSLATE_GOOGLE = False
_TRANSLATE_OPENAI = False

if not os.path.exists("out"):
    os.makedirs("out")

for i in tqdm(range(len(entries)), desc="Processing entries"):
    entry = entries[i]

    if _TRANSLATE_GOOGLE and entry["language"] != "ENGLISH" and entry["title_EN"] == "":
        translations, languages = (
            google_translate([entry["title"], entry["colophon"], entry["imprint"]])
        )
        entries[i]["title_EN"], entries[i]["colophon_EN"], entries[i]["imprint_EN"] = translations

    if _TRANSLATE_OPENAI and entry["language"] != "ENGLISH" and entry["title_EN"] == "":
        for key in ["title", "colophon", "imprint"]:
            if entry[key] != "":
                translation = openai_query(
                    "Translate this text to English",
                    entry[key],
                    "Translate the text to English, preserving the original meaning and context. Do not add any additional information or context. Please preserve the original formatting, whitespace, line breaks (even at the middle of a sentence) and punctuation as much as possible.",
                    max_tokens=None
                )
                entries[i][f"{key}_EN"] = translation

write_csv(entries, file_path, fieldnames)
