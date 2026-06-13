#!/usr/bin/env python3
"""Analyze named mathematical/deductive parts on title pages.

This phase asks what title pages say mathematical knowledge is made of:
propositions, demonstrations, theorems, figures, scholia, definitions, etc.
It keeps the metadata-defined Elements corpus separate from the broader title
page ecology.
"""

from __future__ import annotations

import csv
import re
import unicodedata
from collections import Counter, defaultdict
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
DATA = ROOT / "derived_data"
SOURCE = DATA / "metadata_elements_corpus_ecology_matrix.csv"
MODE_SOURCE = DATA / "metadata_elements_natural_modes_matrix_with_format.csv"


def norm(text: str) -> str:
    text = unicodedata.normalize("NFKD", text or "")
    text = "".join(ch for ch in text if not unicodedata.combining(ch))
    text = text.lower()
    text = text.replace("ſ", "s").replace("ß", "ss").replace("æ", "ae").replace("œ", "oe")
    text = re.sub(r"[^a-z0-9]+", " ", text)
    return re.sub(r"\s+", " ", text).strip()


PART_PATTERNS = {
    "demonstrations_proofs": [
        r"\bdemonstrat\w*\b",
        r"\bdemonstration\w*\b",
        r"\bdimonstrat\w*\b",
        r"\bpreuve\w*\b",
        r"\bpreuves?\b",
        r"\bproofs?\b",
        r"\bproved?\b",
        r"\bprouv\w*\b",
        r"\bbeweis\w*\b",
        r"\bbewezen\b",
        r"\bdimostr\w*\b",
        r"\bprobatio\w*\b",
    ],
    "propositions": [
        r"\bproposition\w*\b",
        r"\bpropos\w*\b",
        r"\bprop\b",
        r"\bvoorstel\w*\b",
    ],
    "theorems": [
        r"\btheorem\w*\b",
        r"\btheoreme\w*\b",
        r"\btheorema\w*\b",
        r"\btheor\w*\b",
    ],
    "problems": [
        r"\bproblem\w*\b",
        r"\bprobleme\w*\b",
        r"\bproblema\w*\b",
    ],
    "figures_diagrams": [
        r"\bfigure\w*\b",
        r"\bfigures?\b",
        r"\bfigura\w*\b",
        r"\bdiagram\w*\b",
        r"\bschema\w*\b",
        r"\bplaat\w*\b",
        r"\bplanches?\b",
        r"\btabulis\b",
        r"\btabulae\b",
    ],
    "scholia_commentary": [
        r"\bscholi\w*\b",
        r"\bscholies?\b",
        r"\bcommentar\w*\b",
        r"\bcommentair\w*\b",
        r"\bcomment\w*\b",
        r"\bannotation\w*\b",
        r"\banmerk\w*\b",
    ],
    "notes_observations": [
        r"\bnotes?\b",
        r"\bnotis\b",
        r"\bnota\w*\b",
        r"\bobservation\w*\b",
        r"\bobservat\w*\b",
        r"\bremarq\w*\b",
        r"\bremarks?\b",
    ],
    "corollaries": [
        r"\bcorollar\w*\b",
        r"\bcorollaria\b",
    ],
    "definitions": [
        r"\bdefinition\w*\b",
        r"\bdefinitio\w*\b",
        r"\bdefin\w*\b",
        r"\bbepaling\w*\b",
    ],
    "axioms_common_notions": [
        r"\baxiom\w*\b",
        r"\bcommon notions?\b",
        r"\bnotiones communes\b",
        r"\bnotion\w* commun\w*\b",
        r"\bcommunes sententiae\b",
    ],
    "postulates": [
        r"\bpostulat\w*\b",
        r"\bpetition\w*\b",
        r"\bpostulata\b",
    ],
    "lemmas": [
        r"\blemma\w*\b",
        r"\blemmata\b",
    ],
    "enunciations": [
        r"\benunciation\w*\b",
        r"\benonce\w*\b",
        r"\benuntiat\w*\b",
    ],
    "examples": [
        r"\bexample\w*\b",
        r"\bexemples?\b",
        r"\bexempla\b",
        r"\bexempel\w*\b",
        r"\bexemplo\w*\b",
    ],
    "operations_constructions": [
        r"\boperation\w*\b",
        r"\bconstruct\w*\b",
        r"\bconstruction\w*\b",
        r"\boperat\w*\b",
    ],
    "principles": [
        r"\bprincip\w*\b",
        r"\bprincipia\b",
        r"\bbeginsel\w*\b",
    ],
    "paradoxes": [
        r"\bparadox\w*\b",
    ],
}


PART_ORDER = list(PART_PATTERNS)


def match_parts(text: str) -> dict[str, bool]:
    return {
        part: any(re.search(pat, text) for pat in patterns)
        for part, patterns in PART_PATTERNS.items()
    }


def period(year: str) -> str:
    try:
        y = int(float(year))
    except Exception:
        return "unknown"
    if y < 1550:
        return "pre-1550"
    if y < 1600:
        return "1550-1599"
    if y < 1650:
        return "1600-1649"
    if y < 1700:
        return "1650-1699"
    return "1700+"


def pct(num: int, den: int) -> str:
    return "" if den == 0 else f"{100 * num / den:.1f}"


def write_csv(path: Path, rows: list[dict], fields: list[str]) -> None:
    with path.open("w", newline="", encoding="utf-8") as f:
        w = csv.DictWriter(f, fieldnames=fields, extrasaction="ignore")
        w.writeheader()
        w.writerows(rows)


with SOURCE.open(newline="", encoding="utf-8-sig") as f:
    reader = csv.DictReader(f)
    rows = list(reader)

mode_by_key = {}
if MODE_SOURCE.exists():
    with MODE_SOURCE.open(newline="", encoding="utf-8-sig") as f:
        for mode_row in csv.DictReader(f):
            mode_by_key[mode_row.get("classification_key", "")] = mode_row


case_rows = []
for row in rows:
    mode_row = mode_by_key.get(row.get("classification_key", ""), {})
    evidence_fields = [
        row.get("short_title", ""),
        row.get("rich_claim_text", ""),
        row.get("int_text", ""),
        row.get("value_text", ""),
        row.get("content_description", ""),
        row.get("base_content", ""),
        row.get("enriched_with", ""),
        row.get("edition_details", ""),
        row.get("bound_with", ""),
        row.get("additional_content", ""),
        row.get("metadata_additional_data", ""),
        row.get("metadata_additional_optics_catoptrics", ""),
        row.get("metadata_additional_archimedes", ""),
    ]
    search_text = norm(" | ".join(evidence_fields))
    matches = match_parts(search_text)
    count = sum(matches.values())
    if not count:
        continue
    out = {
        "classification_key": row.get("classification_key", ""),
        "short_title": row.get("short_title", ""),
        "year": row.get("year", ""),
        "period": row.get("period") or period(row.get("year", "")),
        "city": row.get("city", ""),
        "language_first": row.get("language_first") or row.get("language", ""),
        "primary_subject_family": row.get("primary_subject_family", ""),
        "is_metadata_elements_representative": row.get("is_metadata_elements_representative", ""),
        "elements_books_group": row.get("elements_books_group", ""),
        "natural_dominant_mode": mode_row.get("natural_dominant_mode", "")
        or row.get("euclid_elements_dominant_mode", ""),
        "format_group": mode_row.get("format_group", ""),
        "format_first": mode_row.get("format_first", ""),
        "rich_claim_count": row.get("rich_claim_count", ""),
        "deductive_part_count": count,
        "deductive_parts": ";".join(part for part in PART_ORDER if matches[part]),
        "deductive_part_evidence": " | ".join(x for x in evidence_fields if x),
    }
    out.update({f"part_{part}": matches[part] for part in PART_ORDER})
    case_rows.append(out)


all_n = len(rows)
elements_n = sum(r.get("is_metadata_elements_representative") == "True" for r in rows)
non_elements_n = all_n - elements_n


summary_rows = []
for part in PART_ORDER:
    all_count = sum(1 for r in case_rows if r[f"part_{part}"])
    elem_count = sum(
        1
        for r in case_rows
        if r[f"part_{part}"] and r["is_metadata_elements_representative"] == "True"
    )
    non_count = all_count - elem_count
    summary_rows.append(
        {
            "part": part,
            "all_count": all_count,
            "all_pct": pct(all_count, all_n),
            "metadata_elements_count": elem_count,
            "metadata_elements_pct": pct(elem_count, elements_n),
            "non_elements_count": non_count,
            "non_elements_pct": pct(non_count, non_elements_n),
            "elements_minus_non_elements_pct_points": (
                f"{(100 * elem_count / elements_n) - (100 * non_count / non_elements_n):.1f}"
                if elements_n and non_elements_n
                else ""
            ),
        }
    )


combo_counter = Counter()
combo_examples = defaultdict(list)
for r in case_rows:
    parts = r["deductive_parts"]
    combo_counter[parts] += 1
    if len(combo_examples[parts]) < 8:
        combo_examples[parts].append(r["classification_key"])

combo_rows = [
    {
        "deductive_parts": parts,
        "count": count,
        "examples": ";".join(combo_examples[parts]),
    }
    for parts, count in combo_counter.most_common()
]


def stratify(field: str, corpus_filter: str | None = None) -> list[dict]:
    buckets = defaultdict(list)
    for r in rows:
        if corpus_filter == "elements" and r.get("is_metadata_elements_representative") != "True":
            continue
        if corpus_filter == "non_elements" and r.get("is_metadata_elements_representative") == "True":
            continue
        mode_row = mode_by_key.get(r.get("classification_key", ""), {})
        buckets[r.get(field, "") or mode_row.get(field, "") or "unknown"].append(
            r.get("classification_key", "")
        )

    case_by_key = {r["classification_key"]: r for r in case_rows}
    out = []
    for bucket, keys in sorted(buckets.items()):
        den = len(keys)
        if den < 3:
            continue
        for part in PART_ORDER:
            num = sum(1 for key in keys if key in case_by_key and case_by_key[key][f"part_{part}"])
            if num:
                out.append(
                    {
                        "corpus": corpus_filter or "all",
                        "field": field,
                        "bucket": bucket,
                        "part": part,
                        "count": num,
                        "denominator": den,
                        "pct": pct(num, den),
                    }
                )
    return sorted(out, key=lambda x: (x["field"], x["bucket"], -int(x["count"]), x["part"]))


strata_rows = []
for field in [
    "period",
    "language_first",
    "primary_subject_family",
    "elements_books_group",
    "natural_dominant_mode",
    "format_group",
]:
    strata_rows.extend(stratify(field, None))
    strata_rows.extend(stratify(field, "elements"))
    strata_rows.extend(stratify(field, "non_elements"))


pair_counter = Counter()
pair_examples = defaultdict(list)
for r in case_rows:
    parts = [part for part in PART_ORDER if r[f"part_{part}"]]
    for i, a in enumerate(parts):
        for b in parts[i + 1 :]:
            pair = f"{a} + {b}"
            pair_counter[pair] += 1
            if len(pair_examples[pair]) < 8:
                pair_examples[pair].append(r["classification_key"])

pair_rows = [
    {"pair": pair, "count": count, "examples": ";".join(pair_examples[pair])}
    for pair, count in pair_counter.most_common()
]


interesting_rows = []
interesting_parts = [
    "axioms_common_notions",
    "postulates",
    "definitions",
    "lemmas",
    "enunciations",
    "paradoxes",
    "corollaries",
    "operations_constructions",
    "principles",
]
for r in case_rows:
    if (
        r["is_metadata_elements_representative"] == "True"
        or any(r[f"part_{p}"] for p in interesting_parts)
        or int(r["deductive_part_count"]) >= 3
    ):
        interesting_rows.append(r)

interesting_rows = sorted(
    interesting_rows,
    key=lambda r: (
        r["is_metadata_elements_representative"] != "True",
        -int(r["deductive_part_count"]),
        r["period"],
        r["classification_key"],
    ),
)


case_fields = [
    "classification_key",
    "short_title",
    "year",
    "period",
    "city",
    "language_first",
    "primary_subject_family",
    "is_metadata_elements_representative",
    "elements_books_group",
    "natural_dominant_mode",
    "format_group",
    "format_first",
    "deductive_part_count",
    "deductive_parts",
    "deductive_part_evidence",
] + [f"part_{part}" for part in PART_ORDER]

write_csv(DATA / "deductive_parts_cases.csv", case_rows, case_fields)
write_csv(
    DATA / "deductive_parts_summary.csv",
    summary_rows,
    [
        "part",
        "all_count",
        "all_pct",
        "metadata_elements_count",
        "metadata_elements_pct",
        "non_elements_count",
        "non_elements_pct",
        "elements_minus_non_elements_pct_points",
    ],
)
write_csv(DATA / "deductive_parts_combinations.csv", combo_rows, ["deductive_parts", "count", "examples"])
write_csv(DATA / "deductive_parts_pairs.csv", pair_rows, ["pair", "count", "examples"])
write_csv(
    DATA / "deductive_parts_by_strata.csv",
    strata_rows,
    ["corpus", "field", "bucket", "part", "count", "denominator", "pct"],
)
write_csv(DATA / "deductive_parts_interesting_cases.csv", interesting_rows, case_fields)

print(f"Read {len(rows)} representative rows")
print(f"Detected deductive parts in {len(case_rows)} rows")
print(f"Metadata Elements representatives: {elements_n}")
for row in summary_rows[:8]:
    print(row)
