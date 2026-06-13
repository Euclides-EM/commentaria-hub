#!/usr/bin/env python3
"""Build report-ready tables and first-pass visualizations.

The exploratory phase files contain many useful outputs, but the report needs
compact, integrated tables. This script creates those tables in one place.
"""

from __future__ import annotations

import os
from pathlib import Path

os.environ.setdefault("MPLCONFIGDIR", "/tmp/elements_dh_matplotlib")

import matplotlib.pyplot as plt
import numpy as np
import pandas as pd
import seaborn as sns
from sklearn.decomposition import PCA
from sklearn.preprocessing import StandardScaler


ROOT = Path(__file__).resolve().parents[2]
DERIVED = ROOT / "derived_data"
REPORT = ROOT / "report"
TABLES = REPORT / "tables"
FIGURES = REPORT / "figures"

TABLES.mkdir(parents=True, exist_ok=True)
FIGURES.mkdir(parents=True, exist_ok=True)


FULL = DERIVED / "metadata_elements_corpus_ecology_matrix.csv"
ELEMENTS = DERIVED / "metadata_elements_natural_modes_matrix_with_format.csv"
DEDUCTIVE = DERIVED / "deductive_parts_cases.csv"


SUBJECTS = [
    "Arithmetic/Commerce",
    "Geometry/Theory",
    "Visual/Spatial Arts",
    "Instruments/Measurement",
    "Cosmos/Earth",
    "Applied Mechanics/Military",
    "Music",
]

SOCIAL_MARKERS = {
    "Students/learners": "aud_students_learners",
    "General readers/lovers": "aud_general_readers_lovers",
    "Mathematicians/scholars": "aud_mathematicians_scholars",
    "Artisans/visual trades": "aud_artisans_visual_trades",
    "Architects/builders": "aud_architects_builders",
    "Surveyors/engineers": "aud_surveyors_geometers_engineers",
    "Military users": "aud_military_users",
    "Merchants/commercial": "aud_merchants_commercial_users",
    "Nobility/court audience": "aud_nobility_court_as_audience",
    "Universities/academies": "inst_universities_academies_colleges",
    "Jesuit": "inst_jesuit",
    "Other religious orders": "inst_other_religious_orders",
    "Royal/court institution": "inst_royal_court_institution",
    "Civic/state institution": "inst_civic_state_institution",
    "Engineering/military schools": "inst_engineering_military_schools",
    "Math professor/lecturer": "role_mathematics_professor_lecturer",
    "Royal official": "role_royal_official",
    "Religious identity": "role_religious_identity",
    "Engineer/practitioner": "role_engineer_practitioner",
    "Royal/princely patron": "pat_royal_princely_patron",
    "Noble patron": "pat_noble_patron",
    "Ecclesiastical patron": "pat_ecclesiastical_patron",
    "Civic/collective patron": "pat_civic_collective_patron",
    "Named private patron": "pat_named_private_patron",
}

INTELLECTUAL_MARKERS = {
    "Ease/clarity": "ival_ease_clarity_accessibility",
    "Utility/application": "ival_utility_application_practice",
    "Correction/revision": "ival_correction_revision_accuracy",
    "Augmentation/enrichment": "ival_augmentation_enrichment_completeness",
    "Novelty/modernity": "ival_novelty_invention_modernity",
    "Translation/vernacular": "ival_translation_vernacularization",
    "Demonstration/method": "ival_demonstration_method_order",
    "Ancient restoration": "ival_ancient_restoration_humanist",
    "Visual materiality": "ival_visual_materiality_diagrams",
    "Abridgment/selection": "ival_abridgment_contraction_selection",
}

CLAIM_MARKERS = {
    "Canonical/textual identity": "claim_canonical_textual_identity",
    "Method/demonstration/order": "claim_method_demonstration_order",
    "Access/clarity/pedagogy": "claim_accessibility_clarity_pedagogy",
    "Utility/practice/application": "claim_utility_practice_application",
    "Correction/revision": "claim_correction_revision_accuracy",
    "Augmentation/composition": "claim_augmentation_enrichment_composition",
    "Translation/transfer": "claim_translation_vernacularization_transfer",
    "Ancient authority/restoration": "claim_ancient_authority_restoration",
    "Novelty/modernity": "claim_novelty_modernity_invention",
    "Visual aids": "claim_visual_material_aids",
    "Completeness/system": "claim_completeness_totality_system",
    "Selection/extraction": "claim_selection_extraction_abridgment",
}

DEDUCTIVE_PARTS = {
    "Demonstrations/proofs": "part_demonstrations_proofs",
    "Propositions": "part_propositions",
    "Theorems": "part_theorems",
    "Problems": "part_problems",
    "Figures/diagrams": "part_figures_diagrams",
    "Scholia/commentary": "part_scholia_commentary",
    "Notes/observations": "part_notes_observations",
    "Corollaries": "part_corollaries",
    "Definitions": "part_definitions",
    "Axioms/common notions": "part_axioms_common_notions",
    "Enunciations": "part_enunciations",
    "Examples": "part_examples",
    "Operations/constructions": "part_operations_constructions",
    "Principles": "part_principles",
    "Paradoxes": "part_paradoxes",
}

CORE_REPORT_MARKERS = {
    **CLAIM_MARKERS,
    **INTELLECTUAL_MARKERS,
    **{
        "Students/learners": "aud_students_learners",
        "General readers/lovers": "aud_general_readers_lovers",
        "Surveyors/engineers": "aud_surveyors_geometers_engineers",
        "Military users": "aud_military_users",
        "Universities/academies": "inst_universities_academies_colleges",
        "Jesuit": "inst_jesuit",
        "Engineering/military schools": "inst_engineering_military_schools",
        "Math professor/lecturer": "role_mathematics_professor_lecturer",
        "Engineer/practitioner": "role_engineer_practitioner",
        "Royal/princely patron": "pat_royal_princely_patron",
    },
}


def bool_series(df: pd.DataFrame, col: str) -> pd.Series:
    if col not in df:
        return pd.Series(False, index=df.index)
    s = df[col]
    if s.dtype == bool:
        return s.fillna(False)
    return s.astype(str).str.lower().isin(["true", "1", "yes", "primary", "secondary"])


def rate_table(df: pd.DataFrame, group_col: str, markers: dict[str, str], min_den: int = 1) -> pd.DataFrame:
    rows = []
    for group, sub in df.groupby(group_col, dropna=False):
        group = "missing/unknown" if pd.isna(group) or group == "" else group
        den = len(sub)
        if den < min_den:
            continue
        for label, col in markers.items():
            n = int(bool_series(sub, col).sum())
            rows.append(
                {
                    "group": group,
                    "marker": label,
                    "count": n,
                    "denominator": den,
                    "pct": round(100 * n / den, 1) if den else 0,
                }
            )
    return pd.DataFrame(rows)


def pivot_rates(rate_df: pd.DataFrame) -> pd.DataFrame:
    return rate_df.pivot(index="group", columns="marker", values="pct").fillna(0)


def save_heatmap(matrix: pd.DataFrame, path: Path, title: str, figsize=(12, 7), vmin=0, vmax=None) -> None:
    plt.figure(figsize=figsize)
    sns.heatmap(
        matrix,
        cmap="YlGnBu",
        annot=True,
        fmt=".1f",
        linewidths=0.4,
        linecolor="#eeeeee",
        cbar_kws={"label": "% of group"},
        vmin=vmin,
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


def subject_long_table(df: pd.DataFrame) -> pd.DataFrame:
    rows = []
    for subject in SUBJECTS:
        vals = df[subject].fillna("unrelated").astype(str)
        rows.append(
            {
                "subject": subject,
                "primary_count": int((vals == "primary").sum()),
                "secondary_count": int((vals == "secondary").sum()),
                "unknown_count": int((vals == "unknown").sum()),
                "any_primary_or_secondary_count": int(vals.isin(["primary", "secondary"]).sum()),
                "total_rows": len(df),
                "primary_pct": round(100 * (vals == "primary").sum() / len(df), 1),
                "any_primary_or_secondary_pct": round(
                    100 * vals.isin(["primary", "secondary"]).sum() / len(df), 1
                ),
            }
        )
    return pd.DataFrame(rows)


def subject_marker_rates(df: pd.DataFrame, markers: dict[str, str]) -> pd.DataFrame:
    rows = []
    for subject in SUBJECTS:
        sub = df[df[subject].fillna("").astype(str).eq("primary")]
        den = len(sub)
        for label, col in markers.items():
            n = int(bool_series(sub, col).sum())
            rows.append(
                {
                    "subject": subject,
                    "marker": label,
                    "count": n,
                    "denominator": den,
                    "pct": round(100 * n / den, 1) if den else 0,
                }
            )
    return pd.DataFrame(rows)


def top_contrasts(df: pd.DataFrame, markers: dict[str, str]) -> pd.DataFrame:
    is_el = bool_series(df, "is_metadata_elements_representative")
    rows = []
    for label, col in markers.items():
        e = df[is_el]
        n = df[~is_el]
        e_count = int(bool_series(e, col).sum())
        n_count = int(bool_series(n, col).sum())
        e_rate = 100 * e_count / len(e)
        n_rate = 100 * n_count / len(n)
        rows.append(
            {
                "marker": label,
                "metadata_elements_count": e_count,
                "metadata_elements_denominator": len(e),
                "metadata_elements_pct": round(e_rate, 1),
                "non_elements_count": n_count,
                "non_elements_denominator": len(n),
                "non_elements_pct": round(n_rate, 1),
                "elements_minus_non_elements_pct_points": round(e_rate - n_rate, 1),
            }
        )
    out = pd.DataFrame(rows)
    return out.sort_values("elements_minus_non_elements_pct_points", ascending=False)


def add_deductive_flags(df: pd.DataFrame, deductive: pd.DataFrame) -> pd.DataFrame:
    out = df.copy()
    keyed = deductive.set_index("classification_key")
    for label, col in DEDUCTIVE_PARTS.items():
        values = keyed[col] if col in keyed else pd.Series(dtype=bool)
        out[col] = out["classification_key"].map(values).astype("boolean").fillna(False).astype(bool)
    out["has_any_deductive_part"] = out[[*DEDUCTIVE_PARTS.values()]].any(axis=1)
    return out


def pca_plot(df: pd.DataFrame, feature_cols: list[str], color_col: str, path: Path, title: str) -> pd.DataFrame:
    work = df.copy()
    X = pd.DataFrame({c: bool_series(work, c).astype(float) for c in feature_cols})
    X = X.loc[:, X.sum(axis=0) > 1]
    scaled = StandardScaler().fit_transform(X)
    pca = PCA(n_components=2, random_state=0)
    coords = pca.fit_transform(scaled)
    plot_df = pd.DataFrame(
        {
            "PC1": coords[:, 0],
            "PC2": coords[:, 1],
            color_col: work[color_col].fillna("missing/unknown").astype(str).values,
        }
    )
    plt.figure(figsize=(10, 7))
    sns.scatterplot(
        data=plot_df,
        x="PC1",
        y="PC2",
        hue=color_col,
        s=38,
        alpha=0.78,
        linewidth=0,
        palette="tab10",
    )
    plt.axhline(0, color="#cccccc", linewidth=0.8)
    plt.axvline(0, color="#cccccc", linewidth=0.8)
    plt.title(
        f"{title}\nPC1 {pca.explained_variance_ratio_[0]*100:.1f}%, "
        f"PC2 {pca.explained_variance_ratio_[1]*100:.1f}%",
        fontsize=12,
    )
    plt.legend(bbox_to_anchor=(1.02, 1), loc="upper left", borderaxespad=0)
    plt.tight_layout()
    plt.savefig(path, dpi=220)
    plt.close()

    loadings = pd.DataFrame(
        {
            "feature": X.columns,
            "pc1_loading": pca.components_[0],
            "pc2_loading": pca.components_[1],
        }
    )
    loadings["pc1_abs"] = loadings["pc1_loading"].abs()
    loadings["pc2_abs"] = loadings["pc2_loading"].abs()
    return loadings.sort_values(["pc1_abs", "pc2_abs"], ascending=False)


def main() -> None:
    full = pd.read_csv(FULL)
    elements = pd.read_csv(ELEMENTS)
    deductive = pd.read_csv(DEDUCTIVE)

    full = add_deductive_flags(full, deductive)
    elements = add_deductive_flags(elements, deductive)

    full["corpus_group"] = np.where(
        bool_series(full, "is_metadata_elements_representative"),
        "Metadata Elements",
        "Non-Elements",
    )
    full["pca_group"] = np.select(
        [
            bool_series(full, "is_metadata_elements_representative"),
            full["primary_subject_family"].fillna("").str.contains("Practical Geometry", regex=False),
            full["primary_subject_family"].fillna("").str.contains("Instruments/Measurement", regex=False),
            full["primary_subject_family"].fillna("").str.contains("Arithmetic/Commerce", regex=False),
            full["primary_subject_family"].fillna("").str.contains("Geometry/Theory", regex=False),
        ],
        [
            "Metadata Elements",
            "Practical Geometry",
            "Instruments/Measurement",
            "Arithmetic/Commerce",
            "Geometry/Theory non-Elements",
        ],
        default="Other non-Elements",
    )

    # Corpus accounting and subject terrain.
    corpus_rows = [
        {"metric": "representative_rows", "count": len(full)},
        {
            "metric": "metadata_elements_representatives",
            "count": int(bool_series(full, "is_metadata_elements_representative").sum()),
        },
        {
            "metric": "non_elements_representatives",
            "count": int((~bool_series(full, "is_metadata_elements_representative")).sum()),
        },
        {
            "metric": "elements_with_natural_mode_rows",
            "count": len(elements),
        },
        {
            "metric": "rows_with_named_deductive_parts",
            "count": int(full["has_any_deductive_part"].sum()),
        },
    ]
    pd.DataFrame(corpus_rows).to_csv(TABLES / "report_corpus_accounting.csv", index=False)

    subject_counts = subject_long_table(full)
    subject_counts.to_csv(TABLES / "report_subject_zone_counts.csv", index=False)

    # Subject social/intellectual heatmaps.
    subject_social = subject_marker_rates(full, SOCIAL_MARKERS)
    subject_social.to_csv(TABLES / "report_subject_social_rates_long.csv", index=False)
    subject_social_matrix = subject_social.pivot(index="subject", columns="marker", values="pct").fillna(0)
    subject_social_matrix.to_csv(TABLES / "report_subject_social_rates_matrix.csv")
    save_heatmap(
        subject_social_matrix,
        FIGURES / "heatmap_subject_social_rates.png",
        "Primary Subject Zones x Social Markers",
        figsize=(16, 5.8),
    )

    subject_int = subject_marker_rates(full, INTELLECTUAL_MARKERS)
    subject_int.to_csv(TABLES / "report_subject_intellectual_rates_long.csv", index=False)
    subject_int_matrix = subject_int.pivot(index="subject", columns="marker", values="pct").fillna(0)
    subject_int_matrix.to_csv(TABLES / "report_subject_intellectual_rates_matrix.csv")
    save_heatmap(
        subject_int_matrix,
        FIGURES / "heatmap_subject_intellectual_rates.png",
        "Primary Subject Zones x Intellectual Values",
        figsize=(12, 5.8),
    )

    # Elements/non-Elements contrasts.
    contrast = top_contrasts(full, CORE_REPORT_MARKERS | {"Any named deductive part": "has_any_deductive_part"})
    contrast.to_csv(TABLES / "report_elements_vs_non_elements_core_contrasts.csv", index=False)

    top_abs = contrast.reindex(contrast["elements_minus_non_elements_pct_points"].abs().sort_values(ascending=False).index).head(30)
    plt.figure(figsize=(9, 9))
    sns.barplot(
        data=top_abs,
        y="marker",
        x="elements_minus_non_elements_pct_points",
        color="#4C78A8",
    )
    plt.axvline(0, color="#333333", linewidth=0.8)
    plt.xlabel("Metadata Elements minus non-Elements, percentage points")
    plt.ylabel("")
    plt.title("Largest Elements vs Non-Elements Title-Page Contrasts")
    plt.tight_layout()
    plt.savefig(FIGURES / "bar_elements_vs_non_elements_top_contrasts.png", dpi=220)
    plt.close()

    # Elements internal tables.
    selected_elements_markers = {
        **CLAIM_MARKERS,
        "Students/learners": "aud_students_learners",
        "General readers/lovers": "aud_general_readers_lovers",
        "Surveyors/engineers": "aud_surveyors_geometers_engineers",
        "Military users": "aud_military_users",
        "Universities/academies": "inst_universities_academies_colleges",
        "Jesuit": "inst_jesuit",
        "Math professor/lecturer": "role_mathematics_professor_lecturer",
        "Any named deductive part": "has_any_deductive_part",
        "Demonstrations/proofs": "part_demonstrations_proofs",
        "Propositions": "part_propositions",
        "Scholia/commentary": "part_scholia_commentary",
        "Figures/diagrams": "part_figures_diagrams",
        "Problems": "part_problems",
        "Operations/constructions": "part_operations_constructions",
    }
    mode_rates = rate_table(elements, "natural_dominant_mode", selected_elements_markers, min_den=5)
    mode_rates.to_csv(TABLES / "report_elements_mode_marker_rates_long.csv", index=False)
    mode_matrix = pivot_rates(mode_rates)
    mode_matrix.to_csv(TABLES / "report_elements_mode_marker_rates_matrix.csv")
    save_heatmap(
        mode_matrix,
        FIGURES / "heatmap_elements_mode_marker_rates.png",
        "Elements Natural Modes x Report Markers",
        figsize=(16, 7),
    )

    book_rates = rate_table(elements, "elements_books_group", selected_elements_markers, min_den=5)
    book_rates.to_csv(TABLES / "report_elements_bookgroup_marker_rates_long.csv", index=False)
    book_matrix = pivot_rates(book_rates)
    book_matrix.to_csv(TABLES / "report_elements_bookgroup_marker_rates_matrix.csv")
    save_heatmap(
        book_matrix,
        FIGURES / "heatmap_elements_bookgroup_marker_rates.png",
        "Elements Book Groups x Report Markers",
        figsize=(16, 7.8),
    )

    # Deductive parts.
    deductive_markers = {"Any named deductive part": "has_any_deductive_part", **DEDUCTIVE_PARTS}
    deductive_corpus = rate_table(full, "corpus_group", deductive_markers)
    deductive_corpus.to_csv(TABLES / "report_deductive_parts_by_corpus_long.csv", index=False)
    deductive_corpus_matrix = pivot_rates(deductive_corpus)
    deductive_corpus_matrix.to_csv(TABLES / "report_deductive_parts_by_corpus_matrix.csv")
    save_heatmap(
        deductive_corpus_matrix,
        FIGURES / "heatmap_deductive_parts_by_corpus.png",
        "Named Mathematical Parts: Elements vs Non-Elements",
        figsize=(14, 3),
    )

    deductive_book = rate_table(elements, "elements_books_group", deductive_markers, min_den=5)
    deductive_book.to_csv(TABLES / "report_deductive_parts_by_bookgroup_long.csv", index=False)
    deductive_book_matrix = pivot_rates(deductive_book)
    deductive_book_matrix.to_csv(TABLES / "report_deductive_parts_by_bookgroup_matrix.csv")
    save_heatmap(
        deductive_book_matrix,
        FIGURES / "heatmap_deductive_parts_by_bookgroup.png",
        "Named Mathematical Parts By Elements Book Group",
        figsize=(15, 7.8),
    )

    # Period/language/book group accounting for historical trend sections.
    for group_col in ["period", "language_first", "format_group", "elements_books_group", "natural_dominant_mode"]:
        rate_table(elements, group_col, selected_elements_markers, min_den=3).to_csv(
            TABLES / f"report_elements_{group_col}_marker_rates_long.csv",
            index=False,
        )

    # PCA prototypes.
    feature_cols = []
    for subject in SUBJECTS:
        full[f"subject_primary_{subject}"] = full[subject].fillna("").astype(str).eq("primary")
        feature_cols.append(f"subject_primary_{subject}")
    feature_cols.extend(list(CORE_REPORT_MARKERS.values()))
    feature_cols.extend(["has_any_deductive_part", *DEDUCTIVE_PARTS.values()])
    pca_loadings = pca_plot(
        full,
        feature_cols,
        "pca_group",
        FIGURES / "pca_full_corpus_report_features.png",
        "Full Corpus PCA: Title-Page Feature Space",
    )
    pca_loadings.to_csv(TABLES / "report_pca_full_corpus_loadings.csv", index=False)

    element_feature_cols = list(selected_elements_markers.values())
    for col in [
        "mode_sparse_canonical",
        "mode_pedagogical_method",
        "mode_vernacular_transfer",
        "mode_institutional_authority",
        "mode_composite_apparatus",
        "mode_practical_public",
        "mode_corrected_updated",
        "mode_humanist_ancient",
    ]:
        if col in elements:
            element_feature_cols.append(col)
    pca_el_loadings = pca_plot(
        elements,
        element_feature_cols,
        "natural_dominant_mode",
        FIGURES / "pca_elements_only_report_features.png",
        "Metadata Elements PCA: Internal Feature Space",
    )
    pca_el_loadings.to_csv(TABLES / "report_pca_elements_only_loadings.csv", index=False)

    summary = [
        "# Report Infrastructure Outputs",
        "",
        "Generated tables and first-pass figures for the report skeleton.",
        "",
        "## Key Tables",
        "",
        "- `tables/report_corpus_accounting.csv`",
        "- `tables/report_subject_zone_counts.csv`",
        "- `tables/report_subject_social_rates_matrix.csv`",
        "- `tables/report_subject_intellectual_rates_matrix.csv`",
        "- `tables/report_elements_vs_non_elements_core_contrasts.csv`",
        "- `tables/report_elements_mode_marker_rates_matrix.csv`",
        "- `tables/report_elements_bookgroup_marker_rates_matrix.csv`",
        "- `tables/report_deductive_parts_by_corpus_matrix.csv`",
        "- `tables/report_deductive_parts_by_bookgroup_matrix.csv`",
        "",
        "## Key Figures",
        "",
        "- `figures/heatmap_subject_social_rates.png`",
        "- `figures/heatmap_subject_intellectual_rates.png`",
        "- `figures/bar_elements_vs_non_elements_top_contrasts.png`",
        "- `figures/heatmap_elements_mode_marker_rates.png`",
        "- `figures/heatmap_elements_bookgroup_marker_rates.png`",
        "- `figures/heatmap_deductive_parts_by_corpus.png`",
        "- `figures/heatmap_deductive_parts_by_bookgroup.png`",
        "- `figures/pca_full_corpus_report_features.png`",
        "- `figures/pca_elements_only_report_features.png`",
        "",
        "Use PCA cautiously: it is exploratory and only useful if the plotted gradients are historically interpretable.",
    ]
    (REPORT / "REPORT_INFRASTRUCTURE_OUTPUTS.md").write_text("\n".join(summary) + "\n", encoding="utf-8")

    print(f"Wrote tables to {TABLES}")
    print(f"Wrote figures to {FIGURES}")
    print(f"Full corpus rows: {len(full)}")
    print(f"Elements rows: {len(elements)}")


if __name__ == "__main__":
    main()
