#!/usr/bin/env python3
"""Classify named Elements lines by geographic reach and test correlates."""

from __future__ import annotations

from pathlib import Path
import os
import re

import pandas as pd


ROOT = Path(__file__).resolve().parents[3]
OUT = ROOT / "exploration" / "old_presentation_audit" / "experiment_outputs" / "local_circulation_network"
FIG = OUT / "named_adapter_figures"
FIG.mkdir(parents=True, exist_ok=True)
MATRIX = ROOT / "derived_data" / "metadata_elements_corpus_ecology_matrix.csv"


CITY_REGION = {
    "Alcala": "Iberia",
    "Amsterdam": "Low Countries",
    "Arnhem": "Low Countries",
    "Antwerp": "Low Countries",
    "Basel": "Swiss/Upper Rhine",
    "Bamberg": "German lands",
    "Bologna": "Italy",
    "Cambridge": "England",
    "Cologne": "German lands",
    "Douai": "Low Countries",
    "Ferrara": "Italy",
    "Frankfurt": "German lands",
    "Hamburg": "German lands",
    "Leiden": "Low Countries",
    "Leuven": "Low Countries",
    "London": "England",
    "Lausanne": "Swiss/Upper Rhine",
    "Lyon": "France",
    "Mainz": "German lands",
    "Osnabrück": "German lands",
    "Oxford": "England",
    "Padua": "Italy",
    "Paris": "France",
    "Pesaro": "Italy",
    "Rome": "Italy",
    "Rotterdam": "Low Countries",
    "Rouen": "France",
    "Strasbourg": "Swiss/Upper Rhine",
    "Urbino": "Italy",
    "Utrecht": "Low Countries",
    "Venice": "Italy",
    "Wittenberg": "German lands",
    "Würzburg": "German lands",
}


def clean(value: object) -> str:
    if pd.isna(value):
        return ""
    return str(value).strip()


def split_city_labels(value: object) -> list[str]:
    text = clean(value)
    if not text:
        return []
    parts = [p.strip() for p in re.split(r",| and |/|;", text) if p.strip()]
    return parts or [text]


def bool_col(df: pd.DataFrame, name: str) -> pd.Series:
    if name not in df.columns:
        return pd.Series(False, index=df.index)
    return df[name].map(lambda value: bool(value) if not pd.isna(value) else False)


def first_non_null(df: pd.DataFrame, *cols: str) -> pd.Series:
    out = pd.Series(pd.NA, index=df.index)
    for col in cols:
        if col in df.columns:
            out = out.fillna(df[col])
    return out


def classify_reach(cities: set[str], regions: set[str]) -> str:
    if len(cities) == 1:
        return "single_city"
    if len(regions) == 1:
        return "single_region"
    if len(regions) == 2 and len(cities) <= 3:
        return "interregional_route"
    return "pan_european_multi_region"


def main() -> None:
    profiles = pd.read_csv(OUT / "to_1705_named_adapter_profiles_enriched.csv")
    items = pd.read_csv(OUT / "to_1705_named_adapter_item_long.csv").drop_duplicates(["author_canonical", "key"])
    items = items[items["author_canonical"].isin(set(profiles["author_canonical"]))].copy()
    matrix = pd.read_csv(MATRIX)
    matrix = matrix[matrix["is_metadata_elements_representative"].fillna(False).astype(bool)].copy()

    expanded_city_rows = []
    for _, row in items.iterrows():
        for city in split_city_labels(row["city"]):
            expanded_city_rows.append(
                {
                    "author_canonical": row["author_canonical"],
                    "key": row["key"],
                    "city": city,
                    "region": CITY_REGION.get(city, "Unmapped"),
                }
            )
    city_rows = pd.DataFrame(expanded_city_rows)
    city_rows.to_csv(OUT / "to_1705_named_adapter_city_region_rows.csv", index=False)

    reach_rows = []
    for author, sub in city_rows.groupby("author_canonical"):
        cities = set(sub["city"])
        regions = set(sub["region"])
        reach_rows.append(
            {
                "author_canonical": author,
                "geographic_reach_type": classify_reach(cities, regions),
                "city_count_split": len(cities),
                "cities_split": " | ".join(sorted(cities)),
                "region_count": len(regions),
                "regions": " | ".join(sorted(regions)),
            }
        )
    reach = pd.DataFrame(reach_rows)
    reach = profiles.merge(reach, on="author_canonical", how="left")
    reach.to_csv(OUT / "to_1705_named_adapter_geographic_reach.csv", index=False)

    reach_counts = (
        reach.groupby("geographic_reach_type")
        .agg(author_line_count=("author_canonical", "count"), edition_count=("edition_count", "sum"))
        .reset_index()
        .sort_values("author_line_count", ascending=False)
    )
    total_authors = reach_counts["author_line_count"].sum()
    total_editions = reach_counts["edition_count"].sum()
    reach_counts["author_line_share"] = reach_counts["author_line_count"] / total_authors
    reach_counts["edition_share"] = reach_counts["edition_count"] / total_editions
    reach_counts.to_csv(OUT / "to_1705_named_adapter_geographic_reach_counts.csv", index=False)

    features = items.merge(reach[["author_canonical", "geographic_reach_type"]], on="author_canonical", how="left").merge(
        matrix, left_on="key", right_on="classification_key", how="left", indicator=True
    )
    features["has_feature_row"] = features["_merge"].eq("both")
    lang = first_non_null(features, "language_first", "language_x", "language")
    fmt = first_non_null(features, "format_x", "format")

    features["practical"] = bool_col(features, "claim_utility_practice_application") | bool_col(features, "ival_utility_application_practice")
    features["books_1_6_or_solids"] = features["elements_books_group"].isin(["books_1_6", "books_1_6_plus_solids"])
    features["vernacular_non_latin"] = lang.ne("LATIN")
    features["small_format_8_or_smaller"] = pd.to_numeric(fmt, errors="coerce").ge(8)
    features["large_format_folio_quarto"] = pd.to_numeric(fmt, errors="coerce").isin([2, 4])
    features["institution"] = (
        bool_col(features, "institutions_has")
        | bool_col(features, "inst_universities_academies_colleges")
        | bool_col(features, "inst_jesuit")
        | bool_col(features, "inst_other_religious_orders")
        | bool_col(features, "inst_royal_court_institution")
        | bool_col(features, "inst_civic_state_institution")
        | bool_col(features, "inst_libraries_collections")
        | bool_col(features, "inst_engineering_military_schools")
    )
    features["jesuit"] = bool_col(features, "inst_jesuit") | bool_col(features, "jesuit_religious")
    features["pedagogy_access"] = bool_col(features, "claim_accessibility_clarity_pedagogy") | bool_col(
        features, "ival_ease_clarity_accessibility"
    )
    features["method_order"] = bool_col(features, "claim_method_demonstration_order") | bool_col(
        features, "ival_demonstration_method_order"
    )
    features["augmentation"] = bool_col(features, "claim_augmentation_enrichment_composition") | bool_col(
        features, "ival_augmentation_enrichment_completeness"
    )
    features["restoration_ancient"] = bool_col(features, "claim_ancient_authority_restoration") | bool_col(
        features, "ival_ancient_restoration_humanist"
    )
    features["correction"] = bool_col(features, "claim_correction_revision_accuracy") | bool_col(
        features, "ival_correction_revision_accuracy"
    )
    features["visual_aids"] = bool_col(features, "claim_visual_material_aids") | bool_col(features, "ival_visual_materiality_diagrams")

    factor_cols = [
        "practical",
        "books_1_6_or_solids",
        "vernacular_non_latin",
        "small_format_8_or_smaller",
        "large_format_folio_quarto",
        "institution",
        "jesuit",
        "pedagogy_access",
        "method_order",
        "augmentation",
        "restoration_ancient",
        "correction",
        "visual_aids",
    ]
    matched = features[features["has_feature_row"]].copy()
    matched.to_csv(OUT / "to_1705_named_adapter_geographic_reach_item_features.csv", index=False)
    missing = features[~features["has_feature_row"]][["author_canonical", "key", "year_x", "city_x", "language_x"]]
    missing.to_csv(OUT / "to_1705_named_adapter_geographic_reach_missing_feature_rows.csv", index=False)

    rows = []
    for reach_type, sub in matched.groupby("geographic_reach_type"):
        row = {"geographic_reach_type": reach_type, "matched_edition_count": len(sub), "author_line_count": sub["author_canonical"].nunique()}
        for col in factor_cols:
            row[f"{col}_count"] = int(sub[col].sum())
            row[f"{col}_share"] = sub[col].mean()
        rows.append(row)
    factor_summary = pd.DataFrame(rows).sort_values("geographic_reach_type")
    factor_summary.to_csv(OUT / "to_1705_named_adapter_geographic_reach_factor_summary.csv", index=False)

    author_rows = []
    for author, sub in matched.groupby("author_canonical"):
        row = {
            "author_canonical": author,
            "geographic_reach_type": sub["geographic_reach_type"].iloc[0],
            "matched_edition_count": len(sub),
        }
        for col in factor_cols:
            row[f"any_{col}"] = bool(sub[col].any())
            row[f"majority_{col}"] = bool(sub[col].mean() >= 0.5)
            row[f"{col}_share"] = sub[col].mean()
        author_rows.append(row)
    author_factor = pd.DataFrame(author_rows).sort_values(["geographic_reach_type", "matched_edition_count"], ascending=[True, False])
    author_factor.to_csv(OUT / "to_1705_named_adapter_geographic_reach_author_factors.csv", index=False)

    author_profile_rows = []
    for reach_type, sub in author_factor.groupby("geographic_reach_type"):
        row = {"geographic_reach_type": reach_type, "author_line_count": len(sub), "matched_edition_count": sub["matched_edition_count"].sum()}
        for col in factor_cols:
            row[f"any_{col}_author_count"] = int(sub[f"any_{col}"].sum())
            row[f"any_{col}_author_share"] = sub[f"any_{col}"].mean()
            row[f"majority_{col}_author_count"] = int(sub[f"majority_{col}"].sum())
            row[f"majority_{col}_author_share"] = sub[f"majority_{col}"].mean()
        author_profile_rows.append(row)
    pd.DataFrame(author_profile_rows).sort_values("geographic_reach_type").to_csv(
        OUT / "to_1705_named_adapter_geographic_reach_author_factor_summary.csv", index=False
    )

    os.environ.setdefault("MPLCONFIGDIR", "/private/tmp/matplotlib-cache-elements-dh")
    os.environ.setdefault("XDG_CACHE_HOME", "/private/tmp/elements-dh-cache")
    try:
        import matplotlib.pyplot as plt
    except Exception:
        return

    heat_cols = [
        "practical",
        "books_1_6_or_solids",
        "vernacular_non_latin",
        "small_format_8_or_smaller",
        "institution",
        "jesuit",
        "pedagogy_access",
        "method_order",
        "augmentation",
        "restoration_ancient",
        "correction",
    ]
    reach_order = ["single_city", "single_region", "interregional_route", "pan_european_multi_region"]
    factor_summary = factor_summary.set_index("geographic_reach_type").reindex(reach_order).reset_index()
    heat = factor_summary.set_index("geographic_reach_type")[[f"{c}_share" for c in heat_cols]]
    heat.columns = [
        "practical",
        "1-6/solids",
        "vernacular",
        "8mo+",
        "institution",
        "Jesuit",
        "pedagogy",
        "method",
        "augmentation",
        "restoration",
        "correction",
    ]
    row_counts = factor_summary.set_index("geographic_reach_type")["matched_edition_count"].to_dict()
    row_labels = [
        f"Single city\nn={int(row_counts['single_city'])}",
        f"Single region\nn={int(row_counts['single_region'])}",
        f"Interregional route\nn={int(row_counts['interregional_route'])}",
        f"Multi-region\nn={int(row_counts['pan_european_multi_region'])}",
    ]
    fig, ax = plt.subplots(figsize=(14, 6.2))
    im = ax.imshow(heat.values, cmap="YlGnBu", vmin=0, vmax=1, aspect="auto")
    ax.set_yticks(range(len(heat.index)), labels=row_labels)
    ax.set_xticks(range(len(heat.columns)), labels=heat.columns, rotation=35, ha="right")
    ax.set_title("Title-Page Markers by Geographic Print Profile (matched editions, to 1705)", loc="left", fontsize=16, weight="bold", pad=12)
    for i in range(len(heat.index)):
        for j in range(len(heat.columns)):
            ax.text(j, i, f"{heat.iat[i, j]*100:.0f}%", ha="center", va="center", fontsize=8, color="#111111")
    fig.colorbar(im, ax=ax, fraction=0.025, pad=0.02, label="Share of matched editions")
    fig.tight_layout()
    fig.savefig(FIG / "geographic_reach_correlates_heatmap.png", dpi=220)
    fig.savefig(FIG / "slide_title_page_markers_by_print_profile.png", dpi=300)
    plt.close(fig)


if __name__ == "__main__":
    main()
