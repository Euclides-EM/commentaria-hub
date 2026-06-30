#!/usr/bin/env python3
"""Build first follow-up experiment outputs for the old-presentation audit."""

from __future__ import annotations

from pathlib import Path
import re

import pandas as pd


ROOT = Path(__file__).resolve().parents[3]
DERIVED = ROOT / "derived_data"
OUT = ROOT / "exploration" / "old_presentation_audit" / "experiment_outputs"
OUT.mkdir(parents=True, exist_ok=True)


def bool_series(df: pd.DataFrame, col: str) -> pd.Series:
    if col not in df:
        return pd.Series(False, index=df.index)
    return df[col].astype(str).str.lower().isin(["true", "1", "yes", "primary", "secondary"])


def classify_designation(row: pd.Series) -> str:
    text = " | ".join(
        str(row.get(col, ""))
        for col in ["elements_designation", "base_content", "short_title", "content_description"]
        if pd.notna(row.get(col, ""))
    ).lower()

    if not text.strip():
        return "missing/unclear designation"
    if any(x in text for x in ["phaenomena", "catoptrica", "optica", "data", "levi", "ponderoso"]):
        return "Euclidean corpus beyond Elements"
    if re.search(r"(six|sex|ses|sechs|sei|premiers|primi|prioribus|first six|1[-–]6)", text):
        return "first-six-books Elements"
    if re.search(r"(eleventh|twelfth|onzi|douzi|solid|solide|solids|11[-–]12)", text):
        return "plane-plus-solid Elements"
    if re.search(r"(fifth|quinto|dixiesme|tenth|decim|later|postremos|posteriores)", text):
        return "selected/later-book Elements"
    if any(x in text for x in ["element"]):
        if any(x in text for x in ["euclid", "euclide", "euclidis", "evclid", "eukleid", "ευκλει"]):
            return "Euclid + Elements designation"
        return "Elements language without Euclid"
    if any(x in text for x in ["geometria", "geometrie", "geometry", "géométrie"]):
        if any(x in text for x in ["euclid", "euclide", "euclidis", "evclid", "eukleid", "ευκλει"]):
            return "Euclid + geometry designation"
        return "geometry-only designation"
    if any(x in text for x in ["euclid", "euclide", "euclidis", "evclid", "eukleid", "ευκλει"]):
        return "Euclid without Elements wording"
    return "other designation"


def core_pre1700_tables(df: pd.DataFrame) -> None:
    df = df.copy()
    df["year_num"] = pd.to_numeric(df["year"], errors="coerce")
    pre = df[df["year_num"].lt(1700)].copy()
    ded = pd.read_csv(DERIVED / "deductive_parts_cases.csv")
    part_cols = [c for c in ded.columns if c.startswith("part_")]
    ded["_any_named_part"] = ded[part_cols].apply(lambda r: r.astype(str).str.lower().isin(["true", "1", "yes"]).any(), axis=1)
    pre = pre.merge(ded[["classification_key", "_any_named_part"]], on="classification_key", how="left")

    markers = {
        "Ancient authority/restoration": "claim_ancient_authority_restoration",
        "Explicit Euclid/book identity": "claim_canonical_textual_identity",
        "Method/demonstration/order": "claim_method_demonstration_order",
        "Any named deductive part": "_any_named_part",
        "Translation/transfer": "claim_translation_vernacularization_transfer",
        "Augmentation/composition": "claim_augmentation_enrichment_composition",
        "Selection/extraction": "claim_selection_extraction_abridgment",
        "Utility/practice/application": "claim_utility_practice_application",
        "Access/clarity/pedagogy": "claim_accessibility_clarity_pedagogy",
        "Jesuit": "inst_jesuit",
        "Students/learners": "aud_students_learners",
        "Universities/academies": "inst_universities_academies_colleges",
        "Math professor/lecturer": "role_mathematics_professor_lecturer",
        "Military users": "aud_military_users",
        "General readers/lovers": "aud_general_readers_lovers",
    }

    pre["corpus"] = pre["is_metadata_elements_representative"].map(
        lambda x: "metadata Elements" if str(x).lower() == "true" else "non-Elements"
    )
    rows = []
    for label, col in markers.items():
        for corpus, sub in pre.groupby("corpus"):
            n = int(bool_series(sub, col).sum())
            rows.append(
                {
                    "marker": label,
                    "corpus": corpus,
                    "count": n,
                    "denominator": len(sub),
                    "pct": round(100 * n / len(sub), 1),
                }
            )
    long = pd.DataFrame(rows)
    long.to_csv(OUT / "pre1700_core_contrasts_long.csv", index=False)
    matrix = long.pivot(index="marker", columns="corpus", values="pct").reset_index()
    matrix["elements_minus_non_elements_pct_points"] = (
        matrix["metadata Elements"] - matrix["non-Elements"]
    ).round(1)
    matrix.to_csv(OUT / "pre1700_core_contrasts_matrix.csv", index=False)


def designation_taxonomy(df: pd.DataFrame) -> None:
    elements = df[df["is_metadata_elements_representative"].astype(str).str.lower().eq("true")].copy()
    elements["designation_type"] = elements.apply(classify_designation, axis=1)
    elements[
        [
            "classification_key",
            "year",
            "city",
            "language",
            "short_title",
            "elements_books_group",
            "elements_designation",
            "base_content",
            "designation_type",
        ]
    ].to_csv(OUT / "elements_designation_taxonomy_cases.csv", index=False)
    summary = (
        elements["designation_type"]
        .value_counts()
        .rename_axis("designation_type")
        .reset_index(name="count")
    )
    summary["denominator"] = len(elements)
    summary["pct"] = (100 * summary["count"] / len(elements)).round(1)
    summary.to_csv(OUT / "elements_designation_taxonomy_summary.csv", index=False)

    cross = pd.crosstab(elements["designation_type"], elements["elements_books_group"])
    cross.to_csv(OUT / "elements_designation_taxonomy_by_bookgroup.csv")


def main() -> None:
    df = pd.read_csv(DERIVED / "metadata_elements_corpus_ecology_matrix.csv")
    core_pre1700_tables(df)
    designation_taxonomy(df)


if __name__ == "__main__":
    main()
