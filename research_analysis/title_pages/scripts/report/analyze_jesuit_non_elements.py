#!/usr/bin/env python3
"""Compare Jesuit-marked non-Elements books with Elements and other non-Elements."""

from __future__ import annotations

from pathlib import Path
import os

import pandas as pd


ROOT = Path(__file__).resolve().parents[2]
OUT = ROOT / "results" / "report" / "tables"
FIG = ROOT / "results" / "report" / "figures"
FIG.mkdir(parents=True, exist_ok=True)
MATRIX = ROOT / "data" / "analysis_ready" / "elements_ecology.csv"


def b(df: pd.DataFrame, col: str) -> pd.Series:
    if col not in df.columns:
        return pd.Series(False, index=df.index)
    return df[col].map(lambda value: bool(value) if not pd.isna(value) else False)


def compare_bool(df: pd.DataFrame, group: str, cols: list[tuple[str, str]]) -> pd.DataFrame:
    rows = []
    for label, col in cols:
        for value, sub in df.groupby(group):
            marker = b(sub, col)
            rows.append(
                {
                    "marker": label,
                    group: value,
                    "count": int(marker.sum()),
                    "denominator": len(sub),
                    "share": marker.mean() if len(sub) else 0,
                }
            )
    return pd.DataFrame(rows)


def compare_category(df: pd.DataFrame, group: str, col: str, label: str) -> pd.DataFrame:
    rows = []
    for value, sub in df.groupby(group):
        counts = sub[col].fillna("missing").value_counts()
        for category, count in counts.items():
            rows.append(
                {
                    "variable": label,
                    group: value,
                    "category": category,
                    "count": int(count),
                    "denominator": len(sub),
                    "share": count / len(sub) if len(sub) else 0,
                }
            )
    return pd.DataFrame(rows)


def main() -> None:
    df = pd.read_csv(MATRIX)
    df["year_num"] = pd.to_numeric(df["year"], errors="coerce")
    df = df[df["year_num"].notna() & df["year_num"].le(1705)].copy()
    df["is_elements"] = df["is_metadata_elements_representative"].fillna(False).astype(bool)
    df["jesuit_marked"] = b(df, "inst_jesuit") | b(df, "jesuit_religious")

    def group_label(row: pd.Series) -> str:
        if row["is_elements"] and row["jesuit_marked"]:
            return "Elements: Jesuit-marked"
        if row["is_elements"]:
            return "Elements: other"
        if row["jesuit_marked"]:
            return "Non-Elements: Jesuit-marked"
        return "Non-Elements: other"

    df["comparison_group"] = df.apply(group_label, axis=1)

    bool_cols = [
        ("practical / utility", "claim_utility_practice_application"),
        ("access / pedagogy", "claim_accessibility_clarity_pedagogy"),
        ("method / order", "claim_method_demonstration_order"),
        ("augmentation / composition", "claim_augmentation_enrichment_composition"),
        ("restoration / ancient authority", "claim_ancient_authority_restoration"),
        ("correction / revision", "claim_correction_revision_accuracy"),
        ("translation / transfer", "claim_translation_vernacularization_transfer"),
        ("visual aids", "claim_visual_material_aids"),
        ("selection / abridgment", "claim_selection_extraction_abridgment"),
        ("completeness / system", "claim_completeness_totality_system"),
        ("students / schools", "aud_students_learners"),
        ("universities / colleges", "inst_universities_academies_colleges"),
        ("institution mentioned", "institutions_has"),
        ("Jesuit", "inst_jesuit"),
    ]

    scope = (
        df.groupby("comparison_group")
        .size()
        .reset_index(name="count")
        .sort_values("comparison_group")
    )
    scope.to_csv(OUT / "to_1705_jesuit_elements_non_elements_scope.csv", index=False)
    compare_bool(df, "comparison_group", bool_cols).to_csv(
        OUT / "jesuit_corpus_marker_rates_pre_1706.csv", index=False
    )

    cat_tables = [
        compare_category(df, "comparison_group", "primary_subject_family", "subject family"),
        compare_category(df, "comparison_group", "language_first", "language"),
        compare_category(df, "comparison_group", "period", "period"),
        compare_category(df, "comparison_group", "city", "city"),
    ]
    pd.concat(cat_tables, ignore_index=True).to_csv(
        OUT / "to_1705_jesuit_elements_non_elements_category_comparison.csv", index=False
    )

    cases = df[df["jesuit_marked"] & ~df["is_elements"]].sort_values(["year_num", "city"])[
        [
            "classification_key",
            "year",
            "city",
            "language_first",
            "author_or_editor",
            "primary_subject_family",
            "primary_classes",
            "claim_utility_practice_application",
            "claim_method_demonstration_order",
            "claim_augmentation_enrichment_composition",
            "claim_correction_revision_accuracy",
            "claim_translation_vernacularization_transfer",
            "institutions_has",
            "inst_jesuit",
            "jesuit_religious",
            "short_title",
        ]
    ]
    cases.to_csv(OUT / "to_1705_jesuit_marked_non_elements_cases.csv", index=False)

    os.environ.setdefault("MPLCONFIGDIR", "/private/tmp/matplotlib-cache-elements-dh")
    os.environ.setdefault("XDG_CACHE_HOME", "/private/tmp/elements-dh-cache")
    try:
        import matplotlib.pyplot as plt
    except Exception:
        return

    bools = pd.read_csv(OUT / "jesuit_corpus_marker_rates_pre_1706.csv")
    markers = [
        "practical / utility",
        "access / pedagogy",
        "method / order",
        "augmentation / composition",
        "correction / revision",
        "translation / transfer",
        "visual aids",
        "students / schools",
        "institution mentioned",
    ]
    groups = ["Non-Elements: other", "Non-Elements: Jesuit-marked", "Elements: other", "Elements: Jesuit-marked"]
    plot = bools[bools["marker"].isin(markers)].pivot(index="comparison_group", columns="marker", values="share")
    plot = plot.reindex(groups)[markers]
    fig, ax = plt.subplots(figsize=(14, 6.2))
    im = ax.imshow(plot.values, vmin=0, vmax=1, cmap="YlGnBu", aspect="auto")
    labels = ["Other non-Elements", "Jesuit non-Elements", "Other Elements", "Jesuit Elements"]
    ax.set_yticks(range(len(plot.index)), labels=labels)
    ax.set_xticks(range(len(plot.columns)), labels=plot.columns, rotation=35, ha="right")
    ax.set_title("Jesuit-Marked Books: Elements vs Other Mathematics (to 1705)", loc="left", fontsize=16, weight="bold")
    for i in range(len(plot.index)):
        for j in range(len(plot.columns)):
            ax.text(j, i, f"{plot.iat[i, j]*100:.0f}%", ha="center", va="center", fontsize=8)
    fig.colorbar(im, ax=ax, fraction=0.025, pad=0.02, label="Share of editions")
    fig.tight_layout()
    fig.savefig(FIG / "slide_jesuit_elements_vs_non_elements_markers.png", dpi=300)
    plt.close(fig)


if __name__ == "__main__":
    main()
