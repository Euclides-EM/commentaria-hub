#!/usr/bin/env python3
"""Analyze Dutch plain 1-6 Elements as a possible practical-vernacular route."""

from __future__ import annotations

import csv
from collections import Counter
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
DERIVED = ROOT / "derived_data"
ELEMENTS_MATRIX = DERIVED / "metadata_elements_natural_modes_matrix_with_format.csv"
RICH_MATRIX = DERIVED / "representative_analysis_matrix_rich_social_arenas.csv"
TITLE_MATRIX = DERIVED / "title_page_analysis_matrix.csv"


FEATURES = [
    "mode_sparse_canonical",
    "mode_pedagogical_method",
    "mode_vernacular_transfer",
    "mode_institutional_authority",
    "mode_composite_apparatus",
    "mode_practical_public",
    "mode_corrected_updated",
    "mode_humanist_ancient",
    "claim_method_demonstration_order",
    "claim_accessibility_clarity_pedagogy",
    "claim_utility_practice_application",
    "claim_correction_revision_accuracy",
    "claim_augmentation_enrichment_composition",
    "claim_translation_vernacularization_transfer",
    "claim_ancient_authority_restoration",
    "claim_selection_extraction_abridgment",
    "pedagogical_school_arena",
    "professional_practical_arena",
    "religious_institutional_arena",
    "learned_scholarly_arena",
    "general_public_lovers_arena",
    "no_visible_social_arena",
]


def read_csv(path: Path) -> list[dict[str, str]]:
    with path.open(encoding="utf-8-sig", newline="") as f:
        return list(csv.DictReader(f))


def write_csv(path: Path, rows: list[dict[str, object]], fieldnames: list[str]) -> None:
    with path.open("w", encoding="utf-8", newline="") as f:
        writer = csv.DictWriter(f, fieldnames=fieldnames, extrasaction="ignore")
        writer.writeheader()
        writer.writerows(rows)


def truthy(value: str | None) -> bool:
    return str(value or "").strip().lower() in {"1", "true", "yes", "y"}


def pct(num: int, den: int) -> float:
    return 0.0 if not den else round(100 * num / den, 1)


def norm_books(value: str) -> str:
    return (value or "").replace("–", "-").replace(" ", "")


def is_plain_1_6(row: dict[str, str]) -> bool:
    return row.get("elements_books_group") == "books_1_6" or norm_books(row.get("elements_books", "")) == "1-6"


def is_plus_solids(row: dict[str, str]) -> bool:
    b = norm_books(row.get("elements_books", ""))
    return row.get("elements_books_group") == "books_1_6_plus_solids" or b in {"1-6,11-12", "1-6,11-12"}


def first_language(row: dict[str, str]) -> str:
    return row.get("language_first") or (row.get("language", "").split(",")[0].strip())


def summarize_group(name: str, rows: list[dict[str, str]]) -> dict[str, object]:
    out: dict[str, object] = {
        "group": name,
        "n": len(rows),
        "keys": " | ".join(sorted(r["classification_key"] for r in rows)),
        "cities": " | ".join(sorted({r.get("city", "") for r in rows if r.get("city", "")})),
        "authors": " | ".join(sorted({r.get("author_or_editor", "") for r in rows if r.get("author_or_editor", "")})),
        "formats": " | ".join(sorted({r.get("format_group", "") for r in rows if r.get("format_group", "")})),
        "periods": " | ".join(sorted({r.get("period", "") for r in rows if r.get("period", "")})),
        "subjects": " | ".join(sorted({r.get("primary_subject_family", "") for r in rows if r.get("primary_subject_family", "")})),
    }
    for feat in FEATURES:
        out[f"{feat}_count"] = sum(truthy(r.get(feat)) for r in rows)
        out[f"{feat}_pct"] = pct(out[f"{feat}_count"], len(rows))
    return out


def main() -> None:
    elements = read_csv(ELEMENTS_MATRIX)
    rich = read_csv(RICH_MATRIX)
    title = {r["classification_key"]: r for r in read_csv(TITLE_MATRIX)}

    dutch_plain = [r for r in elements if is_plain_1_6(r) and first_language(r) == "DUTCH"]
    non_dutch_plain = [r for r in elements if is_plain_1_6(r) and first_language(r) != "DUTCH"]
    plus_solids = [r for r in elements if is_plus_solids(r)]
    non_plain = [r for r in elements if not is_plain_1_6(r)]

    # Nearby Dutch non-Elements practical/surveying/measurement books in the same broad period.
    neighboring = []
    for r in rich:
        key = r["classification_key"]
        if key in {e["classification_key"] for e in elements}:
            continue
        title_row = title.get(key, {})
        lang = (r.get("language", "").split(",")[0].strip())
        if lang != "DUTCH":
            continue
        try:
            year = int(float(r.get("year", "")))
        except ValueError:
            continue
        if not 1580 <= year <= 1700:
            continue
        subject_text = " | ".join([r.get("primary_classes", ""), r.get("secondary_classes", ""), r.get("primary_subject_family", "")]).lower()
        social_claim_text = " ".join([r.get("rich_claim_text", ""), r.get("rich_social_text", "")]).lower()
        if not any(term in subject_text + " " + social_claim_text for term in [
            "practical geometry", "surveying", "instrument", "land", "landt", "geometry/theory", "instruments/measurement"
        ]):
            continue
        rr = dict(r)
        rr["format_group"] = title_row.get("format", "")
        rr["elements_books_group"] = "non_elements_dutch_neighbor"
        neighboring.append(rr)

    groups = [
        ("dutch_plain_1_6", dutch_plain),
        ("non_dutch_plain_1_6", non_dutch_plain),
        ("all_plain_1_6", dutch_plain + non_dutch_plain),
        ("plus_solids_1_6_11_12", plus_solids),
        ("non_plain_elements", non_plain),
        ("dutch_non_elements_practical_neighbors_1580_1700", neighboring),
    ]

    group_rows = [summarize_group(name, rows) for name, rows in groups]
    write_csv(DERIVED / "dutch_plain_1_6_group_profiles.csv", group_rows, list(group_rows[0].keys()))

    case_fields = [
        "classification_key", "year", "city", "language", "author_or_editor", "publisher", "format_group",
        "elements_books", "elements_books_group", "primary_classes", "primary_subject_family",
        "natural_dominant_mode", "mode_pedagogical_method", "mode_vernacular_transfer",
        "mode_institutional_authority", "mode_composite_apparatus", "mode_practical_public",
        "mode_corrected_updated", "mode_humanist_ancient", "rich_claim_text", "rich_social_text",
        "action_verbs", "audience", "content_description", "edition_details", "editor_description",
        "enriched_with", "institutions",
    ]
    write_csv(DERIVED / "dutch_plain_1_6_cases.csv", dutch_plain, case_fields)

    neighbor_fields = [
        "classification_key", "short_title", "year", "city", "language", "author_or_editor", "publisher",
        "primary_classes", "secondary_classes", "primary_subject_family", "rich_claim_text", "rich_social_text",
        "claim_method_demonstration_order", "claim_utility_practice_application",
        "claim_correction_revision_accuracy", "claim_augmentation_enrichment_composition",
        "claim_translation_vernacularization_transfer", "professional_practical_arena",
        "pedagogical_school_arena", "general_public_lovers_arena", "no_visible_social_arena",
    ]
    neighbor_fields = [f for f in neighbor_fields if neighboring and f in neighboring[0]]
    write_csv(DERIVED / "dutch_non_elements_practical_neighbors_1580_1700.csv", neighboring, neighbor_fields)

    author_counts = Counter(r.get("author_or_editor", "") for r in dutch_plain)
    city_counts = Counter(r.get("city", "") for r in dutch_plain)
    print(f"Dutch plain 1-6 Elements: {len(dutch_plain)}")
    print(f"Non-Dutch plain 1-6 Elements: {len(non_dutch_plain)}")
    print(f"1-6 + 11-12 Elements: {len(plus_solids)}")
    print(f"Dutch non-Elements practical neighbors: {len(neighboring)}")
    print("Dutch plain 1-6 authors:", "; ".join(f"{k}:{v}" for k, v in author_counts.most_common()))
    print("Dutch plain 1-6 cities:", "; ".join(f"{k}:{v}" for k, v in city_counts.most_common()))


if __name__ == "__main__":
    main()
