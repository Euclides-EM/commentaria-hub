#!/usr/bin/env python3
"""Compare Jesuit-marked Elements editions with the rest of the Elements corpus."""

from __future__ import annotations

from pathlib import Path
import os

import pandas as pd


ROOT = Path(__file__).resolve().parents[3]
OUT = ROOT / "exploration" / "old_presentation_audit" / "experiment_outputs" / "local_circulation_network"
FIG = OUT / "named_adapter_figures"
FIG.mkdir(parents=True, exist_ok=True)
MATRIX = ROOT / "derived_data" / "metadata_elements_corpus_ecology_matrix.csv"
PROFILE = OUT / "to_1705_named_adapter_geographic_reach_item_features.csv"


def b(df: pd.DataFrame, col: str) -> pd.Series:
    if col not in df.columns:
        return pd.Series(False, index=df.index)
    return df[col].map(lambda value: bool(value) if not pd.isna(value) else False)


def pct(x: float) -> str:
    return f"{x * 100:.1f}%"


def compare_bool(df: pd.DataFrame, group: str, cols: list[tuple[str, str]]) -> pd.DataFrame:
    rows = []
    for label, col in cols:
        for value, sub in df.groupby(group):
            rows.append(
                {
                    "marker": label,
                    group: value,
                    "count": int(b(sub, col).sum()),
                    "denominator": len(sub),
                    "share": b(sub, col).mean() if len(sub) else 0,
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


def scope_df(df: pd.DataFrame, scope: str) -> pd.DataFrame:
    if scope == "to_1705":
        year_num = pd.to_numeric(df["year"], errors="coerce")
        return df[year_num.notna() & year_num.le(1705)].copy()
    return df.copy()


def main() -> None:
    matrix = pd.read_csv(MATRIX)
    elements = matrix[matrix["is_metadata_elements_representative"].fillna(False).astype(bool)].copy()
    elements["year_num"] = pd.to_numeric(elements["year"], errors="coerce")
    elements["jesuit_marked"] = b(elements, "inst_jesuit") | b(elements, "jesuit_religious")
    elements["not_jesuit_marked"] = ~elements["jesuit_marked"]
    elements["group"] = elements["jesuit_marked"].map({True: "Jesuit-marked", False: "Other Elements"})

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
        ("religious identity", "role_religious_identity"),
        ("royal / court patronage", "royal_court_patronage"),
        ("institution mentioned", "institutions_has"),
    ]

    summary_rows = []
    for scope in ["all", "to_1705"]:
        sub = scope_df(elements, scope)
        summary_rows.append(
            {
                "scope": scope,
                "elements_representatives": len(sub),
                "jesuit_marked_count": int(sub["jesuit_marked"].sum()),
                "jesuit_marked_share": sub["jesuit_marked"].mean(),
                "other_elements_count": int((~sub["jesuit_marked"]).sum()),
            }
        )
        compare_bool(sub, "group", bool_cols).assign(scope=scope).to_csv(
            OUT / f"{scope}_jesuit_marked_elements_bool_comparison.csv", index=False
        )
        cat_tables = [
            compare_category(sub, "group", "elements_books_group", "book coverage"),
            compare_category(sub, "group", "language_first", "language"),
            compare_category(sub, "group", "period", "period"),
            compare_category(sub, "group", "city", "city"),
        ]
        pd.concat(cat_tables, ignore_index=True).assign(scope=scope).to_csv(
            OUT / f"{scope}_jesuit_marked_elements_category_comparison.csv", index=False
        )
    pd.DataFrame(summary_rows).to_csv(OUT / "jesuit_marked_elements_scope_summary.csv", index=False)

    case_cols = [
        "classification_key",
        "year",
        "city",
        "language_first",
        "author_or_editor",
        "elements_books_group",
        "claim_utility_practice_application",
        "claim_method_demonstration_order",
        "claim_augmentation_enrichment_composition",
        "claim_ancient_authority_restoration",
        "claim_correction_revision_accuracy",
        "institutions_has",
        "inst_jesuit",
        "jesuit_religious",
    ]
    cases = elements[elements["jesuit_marked"]].sort_values(["year_num", "city"])[
        [col for col in case_cols if col in elements.columns]
    ]
    cases.to_csv(OUT / "jesuit_marked_elements_cases.csv", index=False)

    # Join to observed print-reach profiles where available.
    if PROFILE.exists():
        profile = pd.read_csv(PROFILE)
        profile = profile[["key", "author_canonical", "geographic_reach_type"]].drop_duplicates("key")
        reach = elements[elements["year_num"].le(1705)].merge(profile, left_on="classification_key", right_on="key", how="left")
        reach["reach_known"] = reach["geographic_reach_type"].notna()
        reach_rows = []
        for group, sub in reach[reach["reach_known"]].groupby("group"):
            counts = sub["geographic_reach_type"].value_counts()
            for reach_type, count in counts.items():
                reach_rows.append(
                    {
                        "group": group,
                        "geographic_reach_type": reach_type,
                        "count": int(count),
                        "denominator": len(sub),
                        "share": count / len(sub),
                    }
                )
        pd.DataFrame(reach_rows).to_csv(OUT / "to_1705_jesuit_marked_elements_print_reach_comparison.csv", index=False)

    os.environ.setdefault("MPLCONFIGDIR", "/private/tmp/matplotlib-cache-elements-dh")
    os.environ.setdefault("XDG_CACHE_HOME", "/private/tmp/elements-dh-cache")
    try:
        import matplotlib.pyplot as plt
    except Exception:
        return

    to1705_bool = pd.read_csv(OUT / "to_1705_jesuit_marked_elements_bool_comparison.csv")
    markers = [
        "practical / utility",
        "access / pedagogy",
        "method / order",
        "augmentation / composition",
        "correction / revision",
        "translation / transfer",
        "visual aids",
        "students / schools",
        "universities / colleges",
        "institution mentioned",
    ]
    plot = to1705_bool[to1705_bool["marker"].isin(markers)].pivot(index="group", columns="marker", values="share")
    plot = plot.reindex(["Other Elements", "Jesuit-marked"])[markers]
    fig, ax = plt.subplots(figsize=(13.5, 4.6))
    im = ax.imshow(plot.values, vmin=0, vmax=1, cmap="YlGnBu", aspect="auto")
    ax.set_yticks(range(len(plot.index)), labels=["Other Elements", "Jesuit-marked"])
    ax.set_xticks(range(len(plot.columns)), labels=plot.columns, rotation=35, ha="right")
    ax.set_title("Jesuit-Marked Elements: Title-Page Markers (to 1705)", loc="left", fontsize=16, weight="bold")
    for i in range(len(plot.index)):
        for j in range(len(plot.columns)):
            ax.text(j, i, pct(float(plot.iat[i, j])), ha="center", va="center", fontsize=9)
    fig.colorbar(im, ax=ax, fraction=0.025, pad=0.02, label="Share of editions")
    fig.tight_layout()
    fig.savefig(FIG / "slide_jesuit_marked_elements_markers_to_1705.png", dpi=300)
    plt.close(fig)

    # Slide-oriented grouped bars for the clearest non-tautological contrasts.
    cat = pd.read_csv(OUT / "to_1705_jesuit_marked_elements_category_comparison.csv")
    reach_path = OUT / "to_1705_jesuit_marked_elements_print_reach_comparison.csv"
    reach = pd.read_csv(reach_path) if reach_path.exists() else pd.DataFrame()
    contrasts = []

    def add_contrast(label: str, jesuit: float, other: float) -> None:
        contrasts.append({"marker": label, "Jesuit-marked": jesuit, "Other Elements": other})

    def bool_share(marker: str, group: str) -> float:
        row = to1705_bool[to1705_bool["marker"].eq(marker) & to1705_bool["group"].eq(group)]
        return float(row["share"].iloc[0]) if len(row) else 0

    def cat_share(variable: str, category: str, group: str) -> float:
        row = cat[cat["variable"].eq(variable) & cat["category"].eq(category) & cat["group"].eq(group)]
        return float(row["share"].iloc[0]) if len(row) else 0

    def reach_share(category: str, group: str) -> float:
        if reach.empty:
            return 0
        row = reach[reach["geographic_reach_type"].eq(category) & reach["group"].eq(group)]
        return float(row["share"].iloc[0]) if len(row) else 0

    add_contrast("1-6 + solids", cat_share("book coverage", "books_1_6_plus_solids", "Jesuit-marked"), cat_share("book coverage", "books_1_6_plus_solids", "Other Elements"))
    add_contrast("practical", bool_share("practical / utility", "Jesuit-marked"), bool_share("practical / utility", "Other Elements"))
    add_contrast("method/order", bool_share("method / order", "Jesuit-marked"), bool_share("method / order", "Other Elements"))
    add_contrast("correction", bool_share("correction / revision", "Jesuit-marked"), bool_share("correction / revision", "Other Elements"))
    add_contrast("translation", bool_share("translation / transfer", "Jesuit-marked"), bool_share("translation / transfer", "Other Elements"))
    add_contrast("multi-region print", reach_share("pan_european_multi_region", "Jesuit-marked"), reach_share("pan_european_multi_region", "Other Elements"))

    contrast_df = pd.DataFrame(contrasts)
    contrast_df.to_csv(OUT / "to_1705_jesuit_marked_elements_slide_contrasts.csv", index=False)
    fig, ax = plt.subplots(figsize=(11, 5.7))
    x = range(len(contrast_df))
    width = 0.36
    ax.bar([i - width / 2 for i in x], contrast_df["Jesuit-marked"], width=width, color="#245c8f", label="Jesuit-marked")
    ax.bar([i + width / 2 for i in x], contrast_df["Other Elements"], width=width, color="#c9b79c", label="Other Elements")
    ax.set_xticks(list(x), labels=contrast_df["marker"], rotation=25, ha="right")
    ax.set_ylim(0, 1)
    ax.yaxis.set_major_formatter(lambda value, _pos: f"{int(value * 100)}%")
    ax.set_title("What Distinguishes Jesuit-Marked Elements? (to 1705)", loc="left", fontsize=16, weight="bold")
    ax.spines[["top", "right", "left"]].set_visible(False)
    ax.grid(axis="y", alpha=0.2)
    ax.legend(frameon=False, loc="upper left")
    for i, row in contrast_df.iterrows():
        for offset, col in [(-width / 2, "Jesuit-marked"), (width / 2, "Other Elements")]:
            value = row[col]
            ax.text(i + offset, value + 0.025, f"{value * 100:.0f}%", ha="center", va="bottom", fontsize=9)
    fig.tight_layout()
    fig.savefig(FIG / "slide_jesuit_marked_elements_distinctiveness.png", dpi=300)
    plt.close(fig)


if __name__ == "__main__":
    main()
