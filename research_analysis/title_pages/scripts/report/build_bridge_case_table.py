#!/usr/bin/env python3
"""Build report-ready bridge-case tables for Elements/ecology boundaries."""

from __future__ import annotations

import os
from pathlib import Path

os.environ.setdefault("MPLCONFIGDIR", "/tmp/elements_dh_matplotlib")

import matplotlib.pyplot as plt
import pandas as pd
import seaborn as sns


ROOT = Path(__file__).resolve().parents[2]
DERIVED = ROOT / "data" / "analysis_ready"
REPORT = ROOT / "results" / "report"
TABLES = REPORT / "tables"
FIGURES = REPORT / "figures"

TABLES.mkdir(parents=True, exist_ok=True)
FIGURES.mkdir(parents=True, exist_ok=True)

FULL = DERIVED / "elements_ecology.csv"


def bool_series(df: pd.DataFrame, col: str) -> pd.Series:
    if col not in df:
        return pd.Series(False, index=df.index)
    s = df[col]
    if s.dtype == bool:
        return s.fillna(False)
    return s.astype(str).str.lower().isin(["true", "1", "yes", "primary", "secondary"])


def text_has(df: pd.DataFrame, col: str, patterns: list[str]) -> pd.Series:
    if col not in df:
        return pd.Series(False, index=df.index)
    text = df[col].fillna("").astype(str).str.lower()
    out = pd.Series(False, index=df.index)
    for pattern in patterns:
        out = out | text.str.contains(pattern.lower(), regex=False)
    return out


def score(df: pd.DataFrame, parts: list[pd.Series]) -> pd.Series:
    out = pd.Series(0, index=df.index, dtype=int)
    for part in parts:
        out += part.fillna(False).astype(int)
    return out


def subject_any(df: pd.DataFrame, subjects: list[str]) -> pd.Series:
    out = pd.Series(False, index=df.index)
    for subject in subjects:
        if subject in df:
            out = out | df[subject].fillna("").astype(str).isin(["primary", "secondary"])
    return out


def export_top_cases(df: pd.DataFrame, membership: pd.DataFrame, path: Path) -> pd.DataFrame:
    fields = [
        "bridge_route",
        "bridge_score",
        "classification_key",
        "year",
        "city",
        "language",
        "format",
        "primary_subject_family",
        "primary_classes",
        "elements_books_group",
        "euclid_elements_dominant_mode",
        "author_or_editor",
        "rich_claim_text_raw",
        "rich_social_text_raw",
        "metadata_elements_title_page_keys",
    ]
    joined = membership.merge(df, left_on="row_index", right_index=True, how="left")
    fields = [f for f in fields if f in joined.columns]
    top = (
        joined
        .sort_values(["bridge_route", "bridge_score", "year", "classification_key"], ascending=[True, False, True, True])
        .groupby("bridge_route", group_keys=False)
        .head(14)[fields]
    )
    top.to_csv(path, index=False)
    return top


def pct(n: int, d: int) -> float:
    return round(100 * n / d, 1) if d else 0.0


def save_heatmap(matrix: pd.DataFrame, path: Path, title: str) -> None:
    plt.figure(figsize=(13, 5.5))
    sns.heatmap(
        matrix,
        cmap="YlGnBu",
        annot=True,
        fmt=".1f",
        linewidths=0.4,
        linecolor="#eeeeee",
        cbar_kws={"label": "% of route"},
        vmin=0,
        vmax=100,
    )
    plt.title(title, fontsize=13, pad=12)
    plt.xlabel("")
    plt.ylabel("")
    plt.xticks(rotation=35, ha="right")
    plt.yticks(rotation=0)
    plt.tight_layout()
    plt.savefig(path, dpi=220)
    plt.close()


def main() -> None:
    df = pd.read_csv(FULL)
    year_numeric = pd.to_numeric(df["year"], errors="coerce")

    is_elements = bool_series(df, "is_metadata_elements_representative")
    mentions_euclid = bool_series(df, "euclid_or_elements")
    practical_subject = subject_any(
        df,
        [
            "Practical Geometry",
            "Surveying",
            "Instrument Construction",
            "Instrument Use",
            "Mechanics",
            "Military Engineering",
            "Architecture",
            "Perspective",
            "Commercial Mathematics",
            "Navigation",
        ],
    )
    geometry_subject = subject_any(df, ["Theoretical Mathematics", "Practical Geometry", "Surveying"])
    instrument_or_art_subject = subject_any(
        df,
        ["Instrument Construction", "Instrument Use", "Architecture", "Perspective", "Construction"],
    )

    canonical_score = score(
        df,
        [
            is_elements,
            bool_series(df, "claim_canonical_textual_identity"),
            bool_series(df, "claim_ancient_authority_restoration"),
            bool_series(df, "ival_ancient_restoration_humanist"),
            bool_series(df, "claim_correction_revision_accuracy"),
            bool_series(df, "claim_translation_vernacularization_transfer"),
            bool_series(df, "claim_augmentation_enrichment_composition"),
            ~bool_series(df, "claim_utility_practice_application"),
            ~bool_series(df, "aud_surveyors_geometers_engineers"),
            ~bool_series(df, "aud_military_users"),
        ],
    )
    usable_score = score(
        df,
        [
            is_elements,
            bool_series(df, "claim_method_demonstration_order"),
            bool_series(df, "claim_accessibility_clarity_pedagogy"),
            bool_series(df, "claim_utility_practice_application"),
            bool_series(df, "ival_utility_application_practice"),
            bool_series(df, "claim_selection_extraction_abridgment"),
            bool_series(df, "aud_general_readers_lovers") | bool_series(df, "aud_students_learners"),
            practical_subject,
            text_has(df, "elements_books_group", ["1_6_plus_solids", "books_1_6"]),
        ],
    )
    euclidean_practical_score = score(
        df,
        [
            ~is_elements,
            mentions_euclid,
            geometry_subject | practical_subject,
            bool_series(df, "claim_method_demonstration_order"),
            bool_series(df, "claim_utility_practice_application"),
            bool_series(df, "ival_visual_materiality_diagrams"),
            bool_series(df, "aud_surveyors_geometers_engineers") | bool_series(df, "role_engineer_practitioner"),
            text_has(df, "rich_claim_text_raw", ["ex demonstratis", "foundation", "fondament", "fundament"]),
        ],
    )
    material_practical_score = score(
        df,
        [
            ~is_elements,
            ~mentions_euclid,
            practical_subject | instrument_or_art_subject,
            bool_series(df, "claim_utility_practice_application"),
            bool_series(df, "ival_visual_materiality_diagrams"),
            bool_series(df, "aud_surveyors_geometers_engineers")
            | bool_series(df, "aud_military_users")
            | bool_series(df, "aud_artisans_visual_trades")
            | bool_series(df, "aud_merchants_commercial_users"),
            bool_series(df, "role_engineer_practitioner"),
            bool_series(df, "claim_visual_material_aids"),
        ],
    )

    df["canonical_elements_score"] = canonical_score
    df["usable_elements_score"] = usable_score
    df["euclidean_practical_geometry_score"] = euclidean_practical_score
    df["professional_material_practical_arts_score"] = material_practical_score

    route_defs = [
        ("canonical_elements", canonical_score, is_elements, 7),
        ("usable_elements", usable_score, is_elements, 5),
        ("euclidean_practical_geometry", euclidean_practical_score, ~is_elements, 4),
        ("professional_material_practical_arts", material_practical_score, ~is_elements, 4),
    ]

    membership_rows = []
    for route, route_score, mask, threshold in route_defs:
        for idx in df.index[mask & (route_score >= threshold)]:
            membership_rows.append(
                {
                    "row_index": idx,
                    "bridge_route": route,
                    "bridge_score": int(route_score.loc[idx]),
                    "threshold": threshold,
                }
            )
    membership = pd.DataFrame(membership_rows)
    membership.to_csv(TABLES / "report_bridge_case_membership_long.csv", index=False)

    route_names = [name for name, _, _, _ in route_defs]
    overlap_rows = []
    membership_sets = {
        route: set(membership.loc[membership["bridge_route"].eq(route), "row_index"])
        for route in route_names
    }
    for route_a in route_names:
        row = {"bridge_route": route_a}
        for route_b in route_names:
            row[route_b] = len(membership_sets[route_a] & membership_sets[route_b])
        overlap_rows.append(row)
    pd.DataFrame(overlap_rows).to_csv(TABLES / "bridge_case_route_overlap_matrix.csv", index=False)

    scored_path = TABLES / "report_bridge_case_scored_matrix.csv"
    df.to_csv(scored_path, index=False)

    top = export_top_cases(
        df,
        path=TABLES / "report_bridge_case_top_cases.csv",
        membership=membership,
    )

    marker_cols = {
        "canonical/textual": "claim_canonical_textual_identity",
        "ancient/restoration": "claim_ancient_authority_restoration",
        "method/demo/order": "claim_method_demonstration_order",
        "access/pedagogy": "claim_accessibility_clarity_pedagogy",
        "utility/practice": "claim_utility_practice_application",
        "translation": "claim_translation_vernacularization_transfer",
        "correction": "claim_correction_revision_accuracy",
        "augmentation": "claim_augmentation_enrichment_composition",
        "selection": "claim_selection_extraction_abridgment",
        "visual aids": "claim_visual_material_aids",
        "general readers/lovers": "aud_general_readers_lovers",
        "students": "aud_students_learners",
        "surveyors/engineers": "aud_surveyors_geometers_engineers",
        "military users": "aud_military_users",
        "universities/academies": "inst_universities_academies_colleges",
        "Jesuit": "inst_jesuit",
        "engineer/practitioner role": "role_engineer_practitioner",
    }

    rows = []
    joined = membership.merge(df, left_on="row_index", right_index=True, how="left")
    for route, sub in joined.groupby("bridge_route"):
        den = len(sub)
        rows.append({"bridge_route": route, "marker": "n", "count": den, "denominator": den, "pct": 100.0})
        rows.append(
            {
                "bridge_route": route,
                "marker": "metadata Elements",
                "count": int(bool_series(sub, "is_metadata_elements_representative").sum()),
                "denominator": den,
                "pct": pct(int(bool_series(sub, "is_metadata_elements_representative").sum()), den),
            }
        )
        rows.append(
            {
                "bridge_route": route,
                "marker": "mentions Euclid/Elements",
                "count": int(bool_series(sub, "euclid_or_elements").sum()),
                "denominator": den,
                "pct": pct(int(bool_series(sub, "euclid_or_elements").sum()), den),
            }
        )
        for label, col in marker_cols.items():
            n = int(bool_series(sub, col).sum())
            rows.append({"bridge_route": route, "marker": label, "count": n, "denominator": den, "pct": pct(n, den)})

    summary = pd.DataFrame(rows)
    summary.to_csv(TABLES / "report_bridge_case_route_marker_rates_long.csv", index=False)
    matrix = summary.pivot(index="bridge_route", columns="marker", values="pct").fillna(0)
    matrix.to_csv(TABLES / "bridge_case_route_marker_rates_matrix.csv")
    heatmap_cols = [
        "canonical/textual",
        "ancient/restoration",
        "method/demo/order",
        "access/pedagogy",
        "utility/practice",
        "translation",
        "correction",
        "augmentation",
        "selection",
        "visual aids",
        "general readers/lovers",
        "students",
        "surveyors/engineers",
        "military users",
        "Jesuit",
        "engineer/practitioner role",
    ]
    save_heatmap(
        matrix[[c for c in heatmap_cols if c in matrix.columns]],
        FIGURES / "heatmap_bridge_route_marker_rates.png",
        "Bridge Routes: Canonical, Usable, Euclidean Practical, Material Practical",
    )

    coverage_rows = []
    for route, sub in joined.groupby("bridge_route"):
        years = year_numeric.loc[sub["row_index"]].dropna()
        coverage_rows.append(
            {
                "bridge_route": route,
                "count": len(sub),
                "earliest_year": int(years.min()) if len(years) else "",
                "latest_year": int(years.max()) if len(years) else "",
                "top_languages": " | ".join(sub["language_first"].fillna("unknown").value_counts().head(5).index),
                "top_book_groups": " | ".join(sub["elements_books_group"].fillna("not metadata Elements").value_counts().head(5).index),
                "top_primary_subjects": " | ".join(sub["primary_subject_family"].fillna("unknown").value_counts().head(5).index),
            }
        )
    pd.DataFrame(coverage_rows).to_csv(TABLES / "report_bridge_case_route_overview.csv", index=False)

    readme = [
        "# Bridge Case Outputs",
        "",
        "Generated by `report/scripts/build_bridge_case_table.py`.",
        "",
        "Core outputs:",
        "",
        "- `tables/report_bridge_case_route_overview.csv`",
        "- `tables/report_bridge_case_route_marker_rates_matrix.csv`",
        "- `tables/report_bridge_case_route_marker_rates_long.csv`",
        "- `tables/report_bridge_case_route_overlap_matrix.csv`",
        "- `tables/report_bridge_case_top_cases.csv`",
        "- `tables/report_bridge_case_scored_matrix.csv`",
        "- `figures/heatmap_bridge_route_marker_rates.png`",
        "",
        "Routes are heuristic and intended for report integration and close-reading selection:",
        "",
        "- `canonical_elements`: metadata Elements framed through canonical identity, ancient authority, correction, translation, or apparatus, with less direct professional utility.",
        "- `usable_elements`: metadata Elements framed through method, explanation, utility, access, selection, or practical/social address.",
        "- `euclidean_practical_geometry`: non-Elements books that explicitly invoke Euclid/Elements while presenting practical geometry, measurement, problems, or utility.",
        "- `professional_material_practical_arts`: non-Elements practical/instrumental/visual books with professional or material-use signals and no explicit Euclid/Elements dependence.",
        "",
        f"Top-case table rows: {len(top)}",
    ]
    (REPORT / "REPORT_BRIDGE_CASE_OUTPUTS.md").write_text("\n".join(readme) + "\n")

    print(f"Wrote {scored_path}")
    print(f"Wrote {TABLES / 'report_bridge_case_top_cases.csv'}")
    print(f"Wrote {REPORT / 'REPORT_BRIDGE_CASE_OUTPUTS.md'}")


if __name__ == "__main__":
    main()
