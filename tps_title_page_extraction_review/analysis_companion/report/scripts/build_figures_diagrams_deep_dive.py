#!/usr/bin/env python3
"""Split figure/diagram language into report-ready historical functions."""

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
TITLE_MATRIX = DERIVED / "title_page_analysis_matrix.csv"
DEDUCTIVE = DERIVED / "deductive_parts_cases.csv"


def bool_series(df: pd.DataFrame, col: str) -> pd.Series:
    if col not in df:
        return pd.Series(False, index=df.index)
    s = df[col]
    if s.dtype == bool:
        return s.fillna(False)
    return s.astype(str).str.lower().isin(["true", "1", "yes", "primary", "secondary"])


def contains_any(text: pd.Series, needles: list[str]) -> pd.Series:
    out = pd.Series(False, index=text.index)
    lowered = text.fillna("").astype(str).str.lower()
    for needle in needles:
        out = out | lowered.str.contains(needle.lower(), regex=False)
    return out


def rate_table(df: pd.DataFrame, group_col: str, markers: dict[str, str], min_den: int = 1) -> pd.DataFrame:
    rows = []
    for group, sub in df.groupby(group_col, dropna=False):
        den = len(sub)
        if den < min_den:
            continue
        group = "missing/unknown" if pd.isna(group) or group == "" else group
        for label, col in markers.items():
            n = int(bool_series(sub, col).sum())
            rows.append(
                {
                    "group": group,
                    "marker": label,
                    "count": n,
                    "denominator": den,
                    "pct": round(100 * n / den, 1) if den else 0.0,
                }
            )
    return pd.DataFrame(rows)


def save_heatmap(matrix: pd.DataFrame, path: Path, title: str) -> None:
    plt.figure(figsize=(12, 6.5))
    sns.heatmap(
        matrix,
        cmap="YlGnBu",
        annot=True,
        fmt=".1f",
        linewidths=0.4,
        linecolor="#eeeeee",
        cbar_kws={"label": "% of group"},
        vmin=0,
        vmax=max(25, min(100, matrix.to_numpy().max() + 5)),
    )
    plt.title(title, fontsize=13, pad=12)
    plt.xlabel("")
    plt.ylabel("")
    plt.xticks(rotation=35, ha="right")
    plt.yticks(rotation=0)
    plt.tight_layout()
    plt.savefig(path, dpi=220)
    plt.close()


def build_metadata_by_key() -> pd.DataFrame:
    meta = pd.read_csv(TITLE_MATRIX)
    meta["format_numeric"] = pd.to_numeric(meta.get("format"), errors="coerce")

    def first_nonblank(values: pd.Series) -> str:
        cleaned = values.dropna().astype(str)
        cleaned = cleaned[cleaned.str.len() > 0]
        return cleaned.iloc[0] if len(cleaned) else ""

    grouped = (
        meta.groupby("classification_key", dropna=False)
        .agg(
            format_first=("format_numeric", "first"),
            formats_all=("format_numeric", lambda s: " | ".join(str(int(x)) for x in sorted(s.dropna().unique()))),
            has_diagrams_metadata=("has_diagrams", "max"),
            title_has_diagrams_metadata=("title_has_diagrams", "max"),
            title_page_rows=("title_page_key", "count"),
        )
        .reset_index()
    )

    def format_group(value: object) -> str:
        if pd.isna(value):
            return "missing/unknown"
        value = int(value)
        return {
            2: "folio",
            4: "quarto",
            8: "octavo",
            12: "duodecimo",
        }.get(value, str(value))

    grouped["format_group"] = grouped["format_first"].map(format_group)
    return grouped


def main() -> None:
    df = pd.read_csv(FULL)
    meta = build_metadata_by_key()
    ded = pd.read_csv(DEDUCTIVE)[
        [
            "classification_key",
            "part_figures_diagrams",
            "part_demonstrations_proofs",
            "part_propositions",
            "part_theorems",
            "part_problems",
            "part_operations_constructions",
            "part_scholia_commentary",
            "part_corollaries",
            "part_notes_observations",
            "deductive_part_evidence",
        ]
    ]
    df = df.merge(meta, on="classification_key", how="left")
    df = df.merge(ded, on="classification_key", how="left")

    text = (
        df.get("rich_claim_text_raw", "").fillna("").astype(str)
        + " | "
        + df.get("content_description", "").fillna("").astype(str)
        + " | "
        + df.get("enriched_with", "").fillna("").astype(str)
        + " | "
        + df.get("deductive_part_evidence", "").fillna("").astype(str)
    )

    visual_title_claim = (
        bool_series(df, "claim_visual_material_aids")
        | bool_series(df, "ival_visual_materiality_diagrams")
        | bool_series(df, "part_figures_diagrams")
        | contains_any(text, ["figure", "figures", "diagram", "diagrams", "figuren", "figueren", "figura", "figuras", "schemata"])
    )

    proof_apparatus = visual_title_claim & (
        bool_series(df, "part_demonstrations_proofs")
        | bool_series(df, "part_propositions")
        | bool_series(df, "part_theorems")
        | bool_series(df, "part_corollaries")
        | bool_series(df, "claim_method_demonstration_order")
        | contains_any(text, ["demonstration", "demonstrations", "proof", "theorem", "proposition", "corollar", "demonstratis"])
    )
    practical_operation = visual_title_claim & (
        bool_series(df, "part_operations_constructions")
        | bool_series(df, "part_problems")
        | bool_series(df, "claim_utility_practice_application")
        | bool_series(df, "aud_surveyors_geometers_engineers")
        | bool_series(df, "role_engineer_practitioner")
        | contains_any(
            text,
            [
                "operation",
                "construction",
                "measuring",
                "mensuration",
                "survey",
                "instrument",
                "maken",
                "veranderen",
                "t'samenvoe",
                "aftrecken",
                "vermenigvuldig",
                "deelen",
                "practical",
            ],
        )
    )
    visual_pedagogy = visual_title_claim & (
        bool_series(df, "claim_accessibility_clarity_pedagogy")
        | bool_series(df, "ival_ease_clarity_accessibility")
        | bool_series(df, "aud_students_learners")
        | bool_series(df, "aud_general_readers_lovers")
        | contains_any(text, ["explain", "explique", "verkla", "easy", "easie", "clear", "facile", "intelligible", "learn"])
    )
    edition_furnishing = visual_title_claim & (
        bool_series(df, "claim_augmentation_enrichment_composition")
        | bool_series(df, "ival_augmentation_enrichment_completeness")
        | bool_series(df, "metadata_has_additional_content")
        | contains_any(text, ["added", "adiect", "auct", "augment", "enriched", "engraved", "copper", "plates", "tables", "furnished"])
    )
    ancient_learned_apparatus = visual_title_claim & (
        bool_series(df, "claim_ancient_authority_restoration")
        | bool_series(df, "ival_ancient_restoration_humanist")
        | bool_series(df, "part_scholia_commentary")
        | contains_any(text, ["theon", "proclus", "scholia", "greek", "graec", "ancient", "veter"])
    )

    df["visual_title_claim"] = visual_title_claim
    df["fig_proof_apparatus"] = proof_apparatus
    df["fig_practical_operation"] = practical_operation
    df["fig_visual_pedagogy"] = visual_pedagogy
    df["fig_edition_furnishing"] = edition_furnishing
    df["fig_ancient_learned_apparatus"] = ancient_learned_apparatus
    df["metadata_diagrams_any"] = bool_series(df, "has_diagrams_metadata") | bool_series(df, "title_has_diagrams_metadata")
    df["visual_claim_but_no_metadata_diagrams"] = df["visual_title_claim"] & ~df["metadata_diagrams_any"]
    df["metadata_diagrams_but_no_visual_claim"] = df["metadata_diagrams_any"] & ~df["visual_title_claim"]
    df["corpus_group"] = df["is_metadata_elements_representative"].map({True: "metadata Elements", False: "non-Elements"})

    markers = {
        "visual title claim": "visual_title_claim",
        "metadata diagrams": "metadata_diagrams_any",
        "figures as proof apparatus": "fig_proof_apparatus",
        "figures as practical operation": "fig_practical_operation",
        "figures as visual pedagogy": "fig_visual_pedagogy",
        "figures as edition furnishing": "fig_edition_furnishing",
        "figures as ancient/learned apparatus": "fig_ancient_learned_apparatus",
        "visual claim but no metadata diagrams": "visual_claim_but_no_metadata_diagrams",
        "metadata diagrams but no visual claim": "metadata_diagrams_but_no_visual_claim",
    }

    groups = [
        ("corpus", "corpus_group"),
        ("subject", "primary_subject_family"),
        ("period", "period"),
        ("language", "language_first"),
        ("format", "format_group"),
        ("elements_bookgroup", "elements_books_group"),
        ("elements_mode", "euclid_elements_dominant_mode"),
    ]
    for name, col in groups:
        if col not in df:
            continue
        rates = rate_table(df, col, markers)
        rates.to_csv(TABLES / f"report_figures_diagrams_by_{name}_long.csv", index=False)
        matrix = rates.pivot(index="group", columns="marker", values="pct").fillna(0)
        matrix.to_csv(TABLES / f"report_figures_diagrams_by_{name}_matrix.csv")
        if name in {"corpus", "subject", "elements_bookgroup", "elements_mode", "format"}:
            save_heatmap(
                matrix,
                FIGURES / f"heatmap_figures_diagrams_by_{name}.png",
                f"Figure/Diagram Functions By {name.replace('_', ' ').title()}",
            )

    case_fields = [
        "classification_key",
        "year",
        "city",
        "language",
        "format_group",
        "primary_subject_family",
        "is_metadata_elements_representative",
        "elements_books_group",
        "euclid_elements_dominant_mode",
        "visual_title_claim",
        "metadata_diagrams_any",
        "fig_proof_apparatus",
        "fig_practical_operation",
        "fig_visual_pedagogy",
        "fig_edition_furnishing",
        "fig_ancient_learned_apparatus",
        "rich_claim_text_raw",
        "rich_social_text_raw",
        "deductive_part_evidence",
    ]
    case_fields = [c for c in case_fields if c in df.columns]
    cases = df[df["visual_title_claim"] | df["metadata_diagrams_any"]].copy()
    cases.sort_values(["is_metadata_elements_representative", "visual_title_claim", "year"], ascending=[False, False, True])[
        case_fields
    ].to_csv(TABLES / "report_figures_diagrams_cases.csv", index=False)

    shortlist = []
    for route in [
        "fig_proof_apparatus",
        "fig_practical_operation",
        "fig_visual_pedagogy",
        "fig_edition_furnishing",
        "fig_ancient_learned_apparatus",
    ]:
        sub = df[df[route]].copy()
        shortlist.append(sub.sort_values(["is_metadata_elements_representative", "year"], ascending=[False, True]).head(12))
    pd.concat(shortlist).drop_duplicates("classification_key")[case_fields].to_csv(
        TABLES / "report_figures_diagrams_close_reading_shortlist.csv", index=False
    )

    df.to_csv(TABLES / "report_figures_diagrams_scored_matrix.csv", index=False)

    readme = [
        "# Figures And Diagrams Outputs",
        "",
        "Generated by `report/scripts/build_figures_diagrams_deep_dive.py`.",
        "",
        "Core outputs:",
        "",
        "- `tables/report_figures_diagrams_by_corpus_matrix.csv`",
        "- `tables/report_figures_diagrams_by_subject_matrix.csv`",
        "- `tables/report_figures_diagrams_by_elements_bookgroup_matrix.csv`",
        "- `tables/report_figures_diagrams_by_elements_mode_matrix.csv`",
        "- `tables/report_figures_diagrams_by_format_matrix.csv`",
        "- `tables/report_figures_diagrams_cases.csv`",
        "- `tables/report_figures_diagrams_close_reading_shortlist.csv`",
        "- `figures/heatmap_figures_diagrams_by_corpus.png`",
        "- `figures/heatmap_figures_diagrams_by_subject.png`",
        "- `figures/heatmap_figures_diagrams_by_elements_bookgroup.png`",
        "- `figures/heatmap_figures_diagrams_by_elements_mode.png`",
        "- `figures/heatmap_figures_diagrams_by_format.png`",
    ]
    (REPORT / "REPORT_FIGURES_DIAGRAMS_OUTPUTS.md").write_text("\n".join(readme) + "\n")

    print("Wrote figures/diagrams deep-dive outputs")


if __name__ == "__main__":
    main()
