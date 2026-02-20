import json
import os

from features import *
from tools import read_csv, write_csv, openai_query, strip_surrounding_quotes
from tqdm import tqdm
import time
from tpparts import TitlePagePart

file_path = "../public/docs/EiP.csv"

entries, fieldnames = read_csv(file_path)

_DELETE_OLD_OUTPUT = True
_TITLE_FEATURES = True
_TITLE_FEATURES_MERGE = True

# _TITLE_FEATURES_FILTER = lambda curr_entry: curr_entry["key"].strip() in ["Strasbourg 1559", "Paris 1558a", "Rome 1594"]
_TITLE_FEATURES_FILTER = lambda curr_entry: True
_TITLE_PAGE_PART = TitlePagePart.TITLE_PAGE
_FEATURES = [
    IMPRINT_DATE,
    IMPRINT_PUBLISHER,
    IMPRINT_CITY,
    IMPRINT_PRIVILEGES,
    IMPRINT_DEDICATION,
    IMPRINT_ADAPTER_ATTRIBUTION,
    IMPRINT_ADAPTER_DESCRIPTION
]

if _TITLE_FEATURES and _DELETE_OLD_OUTPUT:
    if os.path.exists("out"):
        for f in os.listdir("out"):
            os.remove(os.path.join("out", f))
        print("Deleted old output files.")

if not os.path.exists("out"):
    os.makedirs("out")

run_id = time.strftime("%d_%H-%M")
print(f"Run ID: {run_id}")

instructions = prompt(_FEATURES)
if _TITLE_FEATURES:
    print(f"Using prompt:\n{instructions}\n")

if len(_FEATURES) == 0:
    raise ValueError("No features specified for title processing.")

title_page_part = _FEATURES[0].title_page_part
for f in _FEATURES:
    if f.title_page_part != title_page_part:
        raise ValueError(f"All features must have the same title page part. Found {f.title_page_part} for feature {f.name}, expected {title_page_part}.")

processing_count = len([e for e in entries if _TITLE_FEATURES_FILTER(e)])
print(f"Processing {processing_count} entries with features: {[f.name for f in _FEATURES]}")

for i in tqdm(range(len(entries)), desc="Processing entries"):
    entry = entries[i]

    if _TITLE_FEATURES and _TITLE_FEATURES_FILTER(entry):
        if len(entry[str(title_page_part.value)]) < 10:
            continue
        out_path = f'out/{entry["key"].replace("/", "_")}.json'
        result = openai_query(
            f"language: {entry['language']}{' and ' + entry['language 2'] if entry['language 2'] != '' else ''}",
            entry[str(title_page_part.value)],
            instructions,
            max_tokens=None
        )
        with open(out_path, "w") as f:
            f.write(result)

    if _TITLE_FEATURES_MERGE:
        out_path = f'out/{entry["key"].replace("/", "_")}.json'
        if not os.path.exists(out_path):
            continue
        if not _TITLE_FEATURES_FILTER(entry):
            continue
        entry["tagger"] = f"ai_{run_id}"
        with open(out_path, "r") as f:
            features = f.read().replace("```json", "").replace("```", "").strip()
        try:
            features_dict = json.loads(features)
            for key, value in features_dict.items():
                entries[i][key] = (
                    ", ".join([strip_surrounding_quotes(v) for v in value])
                    if isinstance(value, list)
                    else strip_surrounding_quotes(value))
        except Exception as e:
            print(f"Error processing features for {entry['key']}: {e}")
            continue

write_csv(entries, file_path, fieldnames)
