#!/usr/bin/env python3
"""Build report-ready format/material-use tables and figures."""

from __future__ import annotations

import os
from pathlib import Path

os.environ.setdefault("MPLCONFIGDIR", "/tmp/elements_dh_matplotlib")

import matplotlib.pyplot as plt
import pandas as pd
import seaborn as sns


ROOT = Path(__file__).resolve().parents[2]
DERIVED = ROOT / "derived_data"
REPORT = ROOT / "report"
TABLES = REPORT / "tables"
FIGURES = REPORT / "figures"

TABLES.mkdir(parents=True, exist_ok=True)
FIGURES.mkdir(parents=True, exist_ok=True)

FULL = DERIVED / "metadata_elements_corpus_ecology_matrix.csv"
ELEMENTS = DERIVED / "metadata_elements_natural_modes_matrix_with_format.csv"
TITLE_MATRIX = DERIVED / "title_page_analysis_matrix.csv"
ITEMS_PRINT = ROOT.parents[1] / "ocrflow" / "store" / "items_metadata" / "items_print.csv"


FORMAT_ORDER = ["folio", "quarto", "octavo", "duodecimo", "6", "18", "missing/unknown"]
MODE_MARKERS = {
    "sparse authoritative": "mode_sparse_canonical",
    "pedagogical/method": "mode_pedagogical_method",
    "vernacular transfer": "mode_vernacular_transfer",
    "institutional authority": "mode_institutional_authority",
    "composite apparatus": "mode_composite_apparatus",
    "practical/public": "mode_practical_public",
    "corrected/updated": "mode_corrected_updated",
    "humanist/ancient": "mode_humanist_ancient",
}
REPORT_MARKERS = {
    "explicit Euclid/book identity": "claim_canonical_textual_identity",
    "ancient authority/restoration": "claim_ancient_authority_restoration",
    "method/demonstration/order": "claim_method_demonstration_order",
    "access/clarity/pedagogy": "claim_accessibility_clarity_pedagogy",
    "utility/practice/application": "claim_utility_practice_application",
    "correction/revision": "claim_correction_revision_accuracy",
    "augmentation/composition": "claim_augmentation_enrichment_composition",
    "translation/transfer": "claim_translation_vernacularization_transfer",
    "visual aids": "claim_visual_material_aids",
    "students/learners": "aud_students_learners",
    "general readers/lovers": "aud_general_readers_lovers",
    "universities/academies": "inst_universities_academies_colleges",
    "Jesuit": "inst_jesuit",
    "math professor/lecturer": "role_mathematics_professor_lecturer",
    "professional/practical arena": "professional_practical_arena",
    "no visible social arena": "no_visible_social_arena",
}


def bool_series(df: pd.DataFrame, col: str) -> pd.Series:
    if col not in df:
        return pd.Series(False, index=df.index)
    s = df[col]
    if s.dtype == bool:
        return s.fillna(False)
    return s.astype(str).str.lower().isin(["true", "1", "yes", "primary", "secondary"])


def ordered(df: pd.DataFrame) -> pd.DataFrame:
    available = [x for x in FORMAT_ORDER if x in df.index]
    rest = [x for x in df.index if x not in available]
    return df.loc[available + rest]


def format_group(value: object) -> str:
    if pd.isna(value) or str(value).strip() == "":
        return "missing/unknown"
    try:
        value = int(float(value))
    except ValueError:
        return str(value)
    return {2: "folio", 4: "quarto", 8: "octavo", 12: "duodecimo"}.get(value, str(value))


def save_heatmap(matrix: pd.DataFrame, path: Path, title: str, figsize=(12, 7)) -> None:
    plt.figure(figsize=figsize)
    vmax = max(25, min(100, float(matrix.to_numpy().max()) + 5))
    sns.heatmap(
        matrix,
        cmap="YlGnBu",
        annot=True,
        fmt=".1f",
        linewidths=0.4,
        linecolor="#eeeeee",
        cbar_kws={"label": "% of format group"},
        vmin=0,
        vmax=vmax,
    )
    plt.title(title, fontsize=13, pad=12)
    plt.xlabel("")
    plt.ylabel("")
    plt.xticks(rotation=35, ha="right")
    plt.yticks(rotation=0)
    plt.tight_layout()
    plt.savefig(path, dpi=220)
    plt.close()


def rate_matrix(df: pd.DataFrame, group_col: str, markers: dict[str, str]) -> pd.DataFrame:
    rows = []
    for group, sub in df.groupby(group_col, dropna=False):
        group = "missing/unknown" if pd.isna(group) or str(group).strip() == "" else str(group)
        den = len(sub)
        row = {"group": group, "n": den}
        for label, col in markers.items():
            row[label] = round(100 * int(bool_series(sub, col).sum()) / den, 1) if den else 0.0
        rows.append(row)
    out = pd.DataFrame(rows).set_index("group")
    return ordered(out)


def distribution_by_format(df: pd.DataFrame, corpus_label_col: str) -> pd.DataFrame:
    counts = pd.crosstab(df["format_group"], df[corpus_label_col])
    counts = ordered(counts)
    pcts = counts.div(counts.sum(axis=0), axis=1).mul(100).round(1)
    out = counts.add_prefix("count_").join(pcts.add_prefix("pct_of_"))
    return out.reset_index()


def add_format_to_full(full: pd.DataFrame) -> pd.DataFrame:
    meta = pd.read_csv(ITEMS_PRINT)
    meta["format_numeric"] = pd.to_numeric(meta.get("format"), errors="coerce")
    grouped = (
        meta.rename(columns={"key": "classification_key"})
        .groupby("classification_key", dropna=False)
        .agg(
            format_first=("format_numeric", "first"),
            formats_all=("format_numeric", lambda s: " | ".join(str(int(x)) for x in sorted(s.dropna().unique()))),
            has_diagrams_metadata=("has_diagrams", "max"),
            metadata_rows=("classification_key", "count"),
        )
        .reset_index()
    )
    grouped["format_group"] = grouped["format_first"].map(format_group)
    return full.merge(grouped, on="classification_key", how="left")


def main() -> None:
    full = add_format_to_full(pd.read_csv(FULL))
    elements = pd.read_csv(ELEMENTS)
    elements["format_group"] = elements["format_group"].fillna("missing/unknown").astype(str)
    full["format_group"] = full["format_group"].fillna("missing/unknown").astype(str)
    full["corpus"] = full["is_metadata_elements_representative"].map(
        lambda x: "metadata Elements" if str(x).lower() == "true" else "non-Elements"
    )

    distribution_by_format(full, "corpus").to_csv(TABLES / "report_format_corpus_distribution.csv", index=False)

    book_dist = pd.crosstab(elements["format_group"], elements["elements_books_group"])
    book_dist = ordered(book_dist)
    book_pct = book_dist.div(book_dist.sum(axis=1), axis=0).mul(100).round(1)
    book_dist.add_prefix("count_").join(book_pct.add_prefix("pct_within_format_")).reset_index().to_csv(
        TABLES / "report_format_elements_bookgroup_distribution.csv", index=False
    )

    mode_matrix = rate_matrix(elements, "format_group", MODE_MARKERS)
    mode_matrix.to_csv(TABLES / "report_format_elements_mode_rates_matrix.csv")
    save_heatmap(
        mode_matrix.drop(columns=["n"]),
        FIGURES / "heatmap_format_elements_modes.png",
        "Elements title-page modes by format",
        figsize=(12, 6.5),
    )

    marker_matrix = rate_matrix(elements, "format_group", REPORT_MARKERS)
    marker_matrix.to_csv(TABLES / "report_format_elements_marker_rates_matrix.csv")
    save_heatmap(
        marker_matrix.drop(columns=["n"]),
        FIGURES / "heatmap_format_elements_markers.png",
        "Elements social and intellectual markers by format",
        figsize=(13, 7),
    )

    density_rows = []
    for group, sub in elements.groupby("format_group", dropna=False):
        density_rows.append(
            {
                "format_group": group,
                "n": len(sub),
                "avg_rich_claim_count": round(pd.to_numeric(sub["rich_claim_count"], errors="coerce").mean(), 2),
                "avg_intellectual_value_count": round(
                    pd.to_numeric(sub["intellectual_value_count"], errors="coerce").mean(), 2
                ),
                "avg_social_arena_count": round(pd.to_numeric(sub["social_arena_count"], errors="coerce").mean(), 2),
                "pct_no_visible_social_arena": round(100 * bool_series(sub, "no_visible_social_arena").mean(), 1),
                "pct_sparse_authoritative": round(100 * bool_series(sub, "mode_sparse_canonical").mean(), 1),
                "pct_has_metadata_diagrams": round(100 * bool_series(sub, "has_diagrams_metadata").mean(), 1),
            }
        )
    pd.DataFrame(density_rows).set_index("format_group").pipe(ordered).reset_index().to_csv(
        TABLES / "report_format_elements_density_summary.csv", index=False
    )

    subject_rows = []
    for group, sub in full.groupby("format_group", dropna=False):
        row = {"format_group": group, "n": len(sub)}
        for subject in [
            "Geometry/Theory",
            "Arithmetic/Commerce",
            "Practical Geometry",
            "Architecture",
            "Military Engineering",
            "Instrument Use",
            "Navigation",
            "Astronomy",
        ]:
            if subject in sub:
                row[subject] = round(100 * sub[subject].fillna("").astype(str).isin(["primary", "secondary"]).mean(), 1)
        subject_rows.append(row)
    pd.DataFrame(subject_rows).set_index("format_group").pipe(ordered).reset_index().to_csv(
        TABLES / "report_format_full_corpus_subject_rates.csv", index=False
    )


if __name__ == "__main__":
    main()
