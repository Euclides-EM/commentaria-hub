#!/usr/bin/env python3
"""Test hypotheses about named Elements circulation profiles."""

from __future__ import annotations

from pathlib import Path
import os

import pandas as pd


ROOT = Path(__file__).resolve().parents[3]
OUT = ROOT / "exploration" / "old_presentation_audit" / "experiment_outputs" / "local_circulation_network"
FIG = OUT / "named_adapter_figures"
FIG.mkdir(parents=True, exist_ok=True)
MATRIX = ROOT / "derived_data" / "metadata_elements_corpus_ecology_matrix.csv"


def circulation_profile(row: pd.Series) -> str:
    if row["city_count"] == 1:
        return "single_city_series"
    if row["language_count"] == 1 and row["primary_language"] != "LATIN":
        return "regional_vernacular_series"
    if row["language_count"] == 1 and row["primary_language"] == "LATIN":
        return "multi_city_latin_institutional"
    if row["language_count"] > 1:
        return "multi_language_mediation"
    return "other"


def bool_col(df: pd.DataFrame, name: str) -> pd.Series:
    if name not in df.columns:
        return pd.Series(False, index=df.index)
    return df[name].map(lambda value: bool(value) if not pd.isna(value) else False)


def pct(value: float) -> str:
    return f"{value * 100:.1f}%"


def main() -> None:
    profiles = pd.read_csv(OUT / "to_1705_named_adapter_profiles_enriched.csv")
    profiles["circulation_profile"] = profiles.apply(circulation_profile, axis=1)
    profiles.to_csv(OUT / "to_1705_named_adapter_profiles_classified.csv", index=False)

    items = pd.read_csv(OUT / "to_1705_named_adapter_item_long.csv").drop_duplicates(["author_canonical", "key"])
    items = items[items["author_canonical"].isin(set(profiles["author_canonical"]))].copy()
    matrix = pd.read_csv(MATRIX)
    matrix = matrix[matrix["is_metadata_elements_representative"].fillna(False).astype(bool)].copy()

    item_features = items.merge(
        profiles[["author_canonical", "circulation_profile", "edition_count", "city_count", "language_count", "format_count"]],
        on="author_canonical",
        how="left",
    ).merge(matrix, left_on="key", right_on="classification_key", how="left", indicator=True)

    item_features["has_feature_row"] = item_features["_merge"].eq("both")
    item_features["practical_title_page_claim"] = bool_col(item_features, "claim_utility_practice_application") | bool_col(
        item_features, "ival_utility_application_practice"
    )
    item_features["book_1_6_or_1_6_plus_solids"] = item_features["elements_books_group"].isin(
        ["books_1_6", "books_1_6_plus_solids"]
    )
    item_language = item_features["language_first"]
    if "language_x" in item_features.columns:
        item_language = item_language.fillna(item_features["language_x"])
    elif "language" in item_features.columns:
        item_language = item_language.fillna(item_features["language"])
    item_features["non_latin_language"] = item_language.ne("LATIN")
    item_features["mentions_institution"] = (
        bool_col(item_features, "institutions_has")
        | bool_col(item_features, "inst_universities_academies_colleges")
        | bool_col(item_features, "inst_jesuit")
        | bool_col(item_features, "inst_other_religious_orders")
        | bool_col(item_features, "inst_royal_court_institution")
        | bool_col(item_features, "inst_civic_state_institution")
        | bool_col(item_features, "inst_libraries_collections")
        | bool_col(item_features, "inst_engineering_military_schools")
    )
    item_features["jesuit_marker"] = bool_col(item_features, "inst_jesuit") | bool_col(item_features, "jesuit_religious")

    feature_cols = [
        "practical_title_page_claim",
        "book_1_6_or_1_6_plus_solids",
        "non_latin_language",
        "mentions_institution",
        "jesuit_marker",
    ]

    matched = item_features[item_features["has_feature_row"]].copy()
    item_summary_rows = []
    for profile, sub in matched.groupby("circulation_profile"):
        row = {"circulation_profile": profile, "matched_edition_count": len(sub), "author_line_count": sub["author_canonical"].nunique()}
        for col in feature_cols:
            row[f"{col}_count"] = int(sub[col].sum())
            row[f"{col}_share"] = sub[col].mean()
        item_summary_rows.append(row)
    item_summary = pd.DataFrame(item_summary_rows).sort_values("circulation_profile")
    item_summary.to_csv(OUT / "to_1705_profile_hypothesis_item_summary.csv", index=False)

    author_rows = []
    for author, sub in matched.groupby("author_canonical"):
        profile = sub["circulation_profile"].iloc[0]
        row = {
            "author_canonical": author,
            "circulation_profile": profile,
            "matched_edition_count": len(sub),
            "total_profile_editions": int(profiles.set_index("author_canonical").loc[author, "edition_count"]),
        }
        for col in feature_cols:
            row[f"{col}_edition_share"] = sub[col].mean()
            row[f"any_{col}"] = bool(sub[col].any())
            row[f"majority_{col}"] = bool(sub[col].mean() >= 0.5)
        author_rows.append(row)
    author_summary = pd.DataFrame(author_rows).sort_values(["circulation_profile", "matched_edition_count"], ascending=[True, False])
    author_summary.to_csv(OUT / "to_1705_profile_hypothesis_author_summary.csv", index=False)

    author_profile_summary_rows = []
    for profile, sub in author_summary.groupby("circulation_profile"):
        row = {"circulation_profile": profile, "author_line_count": len(sub), "matched_edition_count": int(sub["matched_edition_count"].sum())}
        for col in feature_cols:
            row[f"any_{col}_author_count"] = int(sub[f"any_{col}"].sum())
            row[f"any_{col}_author_share"] = sub[f"any_{col}"].mean()
            row[f"majority_{col}_author_count"] = int(sub[f"majority_{col}"].sum())
            row[f"majority_{col}_author_share"] = sub[f"majority_{col}"].mean()
        author_profile_summary_rows.append(row)
    author_profile_summary = pd.DataFrame(author_profile_summary_rows).sort_values("circulation_profile")
    author_profile_summary.to_csv(OUT / "to_1705_profile_hypothesis_author_profile_summary.csv", index=False)

    qa_cols = ["author_canonical", "key", "year_x", "city_x", "language_x"]
    qa_cols = [c for c in qa_cols if c in item_features.columns]
    qa = item_features[~item_features["has_feature_row"]][qa_cols]
    qa.to_csv(OUT / "to_1705_profile_hypothesis_missing_feature_rows.csv", index=False)

    os.environ.setdefault("MPLCONFIGDIR", "/private/tmp/matplotlib-cache-elements-dh")
    os.environ.setdefault("XDG_CACHE_HOME", "/private/tmp/elements-dh-cache")
    try:
        import matplotlib.pyplot as plt
    except Exception:
        return

    heat = item_summary.set_index("circulation_profile")[[f"{col}_share" for col in feature_cols]]
    heat.columns = ["practical", "books 1-6/solids", "non-Latin", "institution", "Jesuit"]
    fig, ax = plt.subplots(figsize=(12, 5.2))
    im = ax.imshow(heat.values, cmap="YlGnBu", vmin=0, vmax=1)
    ax.set_yticks(range(len(heat.index)), labels=[x.replace("_", " ") for x in heat.index])
    ax.set_xticks(range(len(heat.columns)), labels=heat.columns, rotation=30, ha="right")
    ax.set_title("Hypothesis Markers By Circulation Profile (matched editions, to 1705)", loc="left", fontsize=12, weight="bold", pad=10)
    for i in range(len(heat.index)):
        for j in range(len(heat.columns)):
            ax.text(j, i, pct(float(heat.iat[i, j])), ha="center", va="center", fontsize=8, color="#111111")
    fig.colorbar(im, ax=ax, fraction=0.03, pad=0.02, label="Share of matched editions")
    fig.tight_layout()
    fig.savefig(FIG / "profile_hypothesis_marker_heatmap.png", dpi=220)
    plt.close(fig)


if __name__ == "__main__":
    main()
