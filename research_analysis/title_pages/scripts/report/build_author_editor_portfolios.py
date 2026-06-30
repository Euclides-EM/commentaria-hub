#!/usr/bin/env python3
"""Build exploratory author/editor portfolio tables.

The metadata field is `author_or_editor`, so this is intentionally exploratory:
it groups title pages by named people, not by a fully disambiguated authority
file or by exact historical role.
"""

from __future__ import annotations

import csv
import re
from collections import Counter, defaultdict
from pathlib import Path


ROOT = Path(__file__).resolve().parents[2]
DERIVED = ROOT / "data" / "analysis_ready"
RESULTS = ROOT / "results" / "report" / "tables"
RICH_MATRIX = DERIVED / "social_arena_representatives.csv"
TITLE_MATRIX = DERIVED / "title_page_corpus.csv"
ELEMENTS_META = ROOT.parents[2] / "ocrflow" / "store" / "items_metadata" / "metadata_elements_print.csv"


VALUE_COLS = [
    "claim_method_demonstration_order",
    "claim_accessibility_clarity_pedagogy",
    "claim_utility_practice_application",
    "claim_correction_revision_accuracy",
    "claim_augmentation_enrichment_composition",
    "claim_translation_vernacularization_transfer",
    "claim_ancient_authority_restoration",
    "claim_visual_material_aids",
    "claim_selection_extraction_abridgment",
]

SOCIAL_COLS = [
    "pedagogical_school_arena",
    "court_state_service_arena",
    "military_fortification_arena",
    "professional_practical_arena",
    "religious_institutional_arena",
    "learned_scholarly_arena",
    "general_public_lovers_arena",
    "patronage_prestige_arena",
    "no_visible_social_arena",
]

MODE_COLS = [
    "sparse_canonical_identity",
    "procedural_pedagogical_identity",
    "composite_workshop_book",
    "utility_public_book",
    "humanist_transfer_book",
    "method_access_book",
]


def read_csv(path: Path) -> list[dict[str, str]]:
    with path.open(encoding="utf-8-sig", newline="") as f:
        return list(csv.DictReader(f))


def write_csv(path: Path, rows: list[dict[str, object]], fieldnames: list[str]) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    with path.open("w", encoding="utf-8", newline="") as f:
        writer = csv.DictWriter(f, fieldnames=fieldnames)
        writer.writeheader()
        writer.writerows(rows)


def truthy(value: str | int | bool | None) -> bool:
    if isinstance(value, bool):
        return value
    if value is None:
        return False
    return str(value).strip().lower() in {"1", "true", "yes", "y"}


def pct(num: int, den: int) -> str:
    if not den:
        return ""
    return f"{100 * num / den:.1f}"


def uniq_join(values: list[str]) -> str:
    clean = sorted({v.strip() for v in values if v and v.strip()})
    return " | ".join(clean)


def split_people(raw: str) -> list[str]:
    raw = (raw or "").strip()
    if not raw:
        return []
    # Most entries use commas between people. Keep compound surnames and
    # particles intact; split only on punctuation that functions as a list.
    parts = re.split(r"\s*(?:,|;|\band\b|\bet\b|&)\s*", raw)
    people = []
    for part in parts:
        name = re.sub(r"\s+", " ", part).strip()
        if not name:
            continue
        people.append(name)
    return people


def first_language(language: str) -> str:
    return (language or "").split(",")[0].strip()


def period(year: str) -> str:
    try:
        y = int(float(year))
    except ValueError:
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


def score(row: dict[str, str], cols: list[str]) -> int:
    return sum(1 for col in cols if truthy(row.get(col)))


def flag_has(case: dict[str, object], col: str) -> bool:
    if col in VALUE_COLS:
        return col in str(case.get("claim_flags", "")).split("; ")
    if col in SOCIAL_COLS:
        return col in str(case.get("social_flags", "")).split("; ")
    return col in str(case.get("mode_flags", "")).split("; ")


def main() -> None:
    rich_rows = read_csv(RICH_MATRIX)
    title_rows = read_csv(TITLE_MATRIX)
    title_by_key = {r["classification_key"]: r for r in title_rows}
    elements_rows = read_csv(ELEMENTS_META)
    element_keys = {r["key"] for r in elements_rows}
    element_bookgroup = {r["key"]: r.get("elements_books", "") for r in elements_rows}

    case_rows: list[dict[str, object]] = []
    person_cases: dict[str, list[dict[str, object]]] = defaultdict(list)
    combo_cases: dict[str, list[dict[str, object]]] = defaultdict(list)

    for row in rich_rows:
        key = row["classification_key"]
        title_row = title_by_key.get(key, {})
        is_elements = key in element_keys
        people = split_people(row.get("author_or_editor", ""))
        if not people:
            people = ["[No named author/editor]"]
        normalized = " | ".join(people)
        base = {
            "classification_key": key,
            "is_metadata_elements": int(is_elements),
            "elements_books": element_bookgroup.get(key, ""),
            "author_or_editor_raw": row.get("author_or_editor", ""),
            "person_count": len(people),
            "people_split": normalized,
            "short_title": row.get("short_title", ""),
            "year": row.get("year", ""),
            "period": period(row.get("year", "")),
            "city": row.get("city", ""),
            "language": row.get("language", ""),
            "language_first": first_language(row.get("language", "")),
            "format": title_row.get("format", ""),
            "volumes": title_row.get("volumes", ""),
            "has_diagrams_metadata": title_row.get("has_diagrams", ""),
            "publisher": row.get("publisher", ""),
            "primary_classes": row.get("primary_classes", ""),
            "primary_subject_family": row.get("primary_subject_family", ""),
            "intellectual_value_count": row.get("intellectual_value_count", ""),
            "social_arena_count": row.get("social_arena_count", ""),
            "value_flags_count": score(row, VALUE_COLS),
            "social_flags_count": score(row, SOCIAL_COLS),
            "claim_flags": "; ".join(col for col in VALUE_COLS if truthy(row.get(col))),
            "social_flags": "; ".join(col for col in SOCIAL_COLS if truthy(row.get(col))),
            "mode_flags": "; ".join(col for col in MODE_COLS if truthy(row.get(col))),
            "rich_claim_text": row.get("rich_claim_text", ""),
            "rich_social_text": row.get("rich_social_text", ""),
        }
        case_rows.append(base)
        combo_cases[row.get("author_or_editor", "").strip() or "[No named author/editor]"].append(base)
        for person in people:
            person_cases[person].append(base)

    def summarize_cases(name: str, cases: list[dict[str, object]], kind: str) -> dict[str, object]:
        elements = [c for c in cases if c["is_metadata_elements"]]
        non = [c for c in cases if not c["is_metadata_elements"]]
        all_claims = {c for case in cases for c in str(case["claim_flags"]).split("; ") if c}
        e_claims = {c for case in elements for c in str(case["claim_flags"]).split("; ") if c}
        n_claims = {c for case in non for c in str(case["claim_flags"]).split("; ") if c}
        all_social = {c for case in cases for c in str(case["social_flags"]).split("; ") if c}
        e_social = {c for case in elements for c in str(case["social_flags"]).split("; ") if c}
        n_social = {c for case in non for c in str(case["social_flags"]).split("; ") if c}

        row: dict[str, object] = {
            "portfolio_kind": kind,
            "name": name,
            "total_representative_cases": len(cases),
            "metadata_elements_cases": len(elements),
            "non_elements_cases": len(non),
            "elements_share_pct": pct(len(elements), len(cases)),
            "cities": uniq_join([str(c["city"]) for c in cases]),
            "languages_first": uniq_join([str(c["language_first"]) for c in cases]),
            "formats": uniq_join([str(c["format"]) for c in cases]),
            "periods": uniq_join([str(c["period"]) for c in cases]),
            "elements_bookgroups": uniq_join([str(c["elements_books"]) for c in elements]),
            "primary_subject_families": uniq_join([str(c["primary_subject_family"]) for c in cases]),
            "elements_keys": uniq_join([str(c["classification_key"]) for c in elements]),
            "non_elements_keys": uniq_join([str(c["classification_key"]) for c in non]),
            "elements_only_claims": "; ".join(sorted(e_claims - n_claims)),
            "non_elements_only_claims": "; ".join(sorted(n_claims - e_claims)),
            "shared_claims": "; ".join(sorted(e_claims & n_claims)),
            "elements_only_social": "; ".join(sorted(e_social - n_social)),
            "non_elements_only_social": "; ".join(sorted(n_social - e_social)),
            "shared_social": "; ".join(sorted(e_social & n_social)),
        }

        for col in VALUE_COLS + SOCIAL_COLS + MODE_COLS:
            row[f"{col}_elements_pct"] = pct(sum(flag_has(c, col) for c in elements), len(elements))
            row[f"{col}_non_elements_pct"] = pct(sum(flag_has(c, col) for c in non), len(non))
        return row

    person_summary = [
        summarize_cases(person, cases, "split_person")
        for person, cases in sorted(person_cases.items(), key=lambda kv: (-len(kv[1]), kv[0]))
    ]
    combo_summary = [
        summarize_cases(combo, cases, "raw_author_or_editor")
        for combo, cases in sorted(combo_cases.items(), key=lambda kv: (-len(kv[1]), kv[0]))
    ]

    interesting = []
    for row in person_summary:
        e = int(row["metadata_elements_cases"])
        n = int(row["non_elements_cases"])
        if e >= 2 and n >= 2:
            category = "bridge_portfolio"
        elif e >= 3 and n == 0:
            category = "elements_heavy_within_sample"
        elif e >= 1 and n >= 5:
            category = "few_elements_many_non_elements_within_sample"
        elif e >= 1 and n >= 1:
            category = "small_crossover"
        elif e >= 1:
            category = "elements_only_low_count"
        else:
            category = "non_elements_only"
        if category != "non_elements_only":
            out = dict(row)
            out["portfolio_category"] = category
            interesting.append(out)

    # Pair each Elements case with nearby non-Elements cases by the same person.
    pair_rows = []
    for person, cases in person_cases.items():
        elems = [c for c in cases if c["is_metadata_elements"]]
        nons = [c for c in cases if not c["is_metadata_elements"]]
        for e in elems:
            for n in nons:
                try:
                    gap = abs(int(str(e["year"])) - int(str(n["year"])))
                except ValueError:
                    gap = 9999
                pair_rows.append({
                    "person": person,
                    "elements_key": e["classification_key"],
                    "elements_year": e["year"],
                    "elements_city": e["city"],
                    "elements_language_first": e["language_first"],
                    "elements_format": e["format"],
                    "elements_books": e["elements_books"],
                    "elements_claim_flags": e["claim_flags"],
                    "elements_social_flags": e["social_flags"],
                    "non_elements_key": n["classification_key"],
                    "non_elements_year": n["year"],
                    "non_elements_city": n["city"],
                    "non_elements_language_first": n["language_first"],
                    "non_elements_format": n["format"],
                    "non_elements_subject_family": n["primary_subject_family"],
                    "non_elements_claim_flags": n["claim_flags"],
                    "non_elements_social_flags": n["social_flags"],
                    "year_gap": gap,
                    "same_city": int(e["city"] == n["city"] and bool(e["city"])),
                    "same_language_first": int(e["language_first"] == n["language_first"] and bool(e["language_first"])),
                    "same_format": int(e["format"] == n["format"] and bool(e["format"])),
                })
    pair_rows.sort(key=lambda r: (int(r["year_gap"]), r["person"], str(r["elements_key"])))

    summary_fields = list(person_summary[0].keys()) if person_summary else []
    interesting_fields = ["portfolio_category"] + summary_fields
    case_fields = list(case_rows[0].keys()) if case_rows else []
    pair_fields = list(pair_rows[0].keys()) if pair_rows else []

    write_csv(RESULTS / "author_editor_portfolio_case_matrix.csv", case_rows, case_fields)
    write_csv(RESULTS / "author_editor_portfolio_person_summary.csv", person_summary, summary_fields)
    write_csv(RESULTS / "author_editor_portfolio_raw_combo_summary.csv", combo_summary, summary_fields)
    write_csv(RESULTS / "author_editor_portfolio_interesting_people.csv", interesting, interesting_fields)
    write_csv(RESULTS / "author_editor_portfolio_elements_non_elements_pairs.csv", pair_rows, pair_fields)

    cats = Counter(r["portfolio_category"] for r in interesting)
    print(f"Representative title-page rows: {len(rich_rows)}")
    print(f"Metadata Elements rows represented: {sum(1 for r in case_rows if r['is_metadata_elements'])}")
    print(f"People with at least one represented Elements case: {sum(1 for r in person_summary if int(r['metadata_elements_cases']) > 0)}")
    print(f"People with Elements and non-Elements represented: {sum(1 for r in person_summary if int(r['metadata_elements_cases']) > 0 and int(r['non_elements_cases']) > 0)}")
    print("Interesting portfolio categories:")
    for k, v in cats.most_common():
        print(f"  {k}: {v}")
    print(f"Nearby/crossover pairs: {len(pair_rows)}")


if __name__ == "__main__":
    main()
