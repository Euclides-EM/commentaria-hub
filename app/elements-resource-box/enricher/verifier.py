from tools import read_csv
from collections import defaultdict
import re

file_paths = [
    "../public/docs/EiP.csv",
    "../public/docs/EiP-secondary.csv"
]

split_fields = [
    "EUCLID REF",
    "VERBS",
    "OTHER NAMES",
    "EXPLICITLY STATED: TRANSLATED FROM",
    "EXPLICITLY STATED: TRANSLATED TO",
    "ELEMENTS DESIGNATION",
    "INSTITUTIONS",
    "BOUND WITH",
    "ENRICHED WITH",
    "IMPRINT PUBLISHER",
    "IMPRINT PRIVILEGES",
    "IMPRINT PLACE",
]


def normalize(s):
    s = s.lower()
    return re.sub(r"\s+|-", "", s)


problems = defaultdict(set)
all_entries = []
all_fieldnames = set()

for file_path in file_paths:
    entries, fieldnames = read_csv(file_path)
    all_entries.extend(entries)
    all_fieldnames.update(fieldnames)

all_fieldnames = list(all_fieldnames)

for entry in all_entries:
    title = entry.get("title", "").strip()
    if title == "?" or title == "":
        continue
    key = entry.get("key", "")

    if not title or not key:
        continue

    norm_title = normalize(title)
    imprint = entry.get("imprint", "").strip() if entry.get("imprint") else ""
    norm_imprint = normalize(imprint)

    for field in all_fieldnames:
        if not field.isupper():
            continue

        value = entry.get(field, "").strip()
        if not value or value == "none":
            continue

        if field in split_fields:
            parts = value.split(", ")
            for part in parts:
                if part and normalize(part) not in norm_title and not any(normalize(part) in normalize(word) for word in title.split()) and normalize(part) not in norm_imprint and not any(normalize(part) in normalize(word) for word in imprint.split()):
                    problems[key].add(field)
        else:
            if normalize(value) not in norm_title and not any(normalize(value) in normalize(word) for word in title.split()) and normalize(value) not in norm_imprint and not any(normalize(value) in normalize(word) for word in imprint.split()):
                problems[key].add(field)

for key, fields in problems.items():
    title = next((entry.get("title", "") for entry in all_entries if entry.get("key", "") == key), "")
    norm_title = normalize(title)
    imprint = next((entry.get("imprint", "") for entry in all_entries if entry.get("key", "") == key), "")
    print(f">>> {key}:")
    print("  Title: ")
    print(f"    {title.replace('\n', '\\n')}")
    if imprint:
        print("  Imprint: ")
        print(f"    {imprint.replace('\n', '\\n')}")
    print("  Problematic fields:")
    for field in sorted(fields):
        entry_value = next((entry.get(field, "") for entry in all_entries if entry.get("key", "") == key), "")
        if field in split_fields:
            parts = entry_value.split(", ")
            problematic_parts = [part for part in parts if part and normalize(part) not in norm_title]
            problematic_value = ", ".join(problematic_parts)
            print(f"    {field}: \"{problematic_value.replace('\n', '\\n')}\"")
        else:
            print(f"    {field}: \"{entry_value.replace('\n', '\\n')}\"")
