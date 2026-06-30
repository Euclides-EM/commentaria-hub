#!/usr/bin/env python3
"""Split commentary/scholia language into report-ready subtypes."""

from __future__ import annotations

import re
import unicodedata
from pathlib import Path

import os
os.environ.setdefault("MPLCONFIGDIR", "/tmp/elements_dh_matplotlib")

import pandas as pd
import matplotlib.pyplot as plt
import seaborn as sns


ROOT = Path(__file__).resolve().parents[2]
DERIVED = ROOT / "data" / "analysis_ready"
REPORT = ROOT / "results" / "report"
TABLES = REPORT / "tables"
FIGURES = REPORT / "figures"

ELEMENTS = DERIVED / "elements_modes.csv"
FULL = DERIVED / "elements_ecology.csv"

TABLES.mkdir(parents=True, exist_ok=True)
FIGURES.mkdir(parents=True, exist_ok=True)


COMMENTARY_TYPES = {
    "ancient_humanist_scholia": [
        r"\bscholi\w*\s+antiqu\w*\b",
        r"\bantiqu\w*\s+scholi\w*\b",
        r"\be scriptis veterum\b",
        r"\bveterum ac recentiorum\b",
        r"\btheon\w*\b",
        r"\bprocl\w*\b",
        r"\bgreek.*latin\b",
        r"\bgraec\w*.{0,40}latin\w*\b",
        r"\bgr[ae]c\w*.{0,40}latin\w*\b",
        r"\blatin\w*.{0,40}gr[ae]c\w*\b",
        r"\bde graeca lingua in latinam\b",
    ],
    "clavius_jesuit_commentary": [
        r"\bclavi\w*\b",
        r"\bclavij\b",
        r"\bcommentari\w*\b.{0,80}\bsocietat\w* ies\w*\b",
        r"\bsocietat\w* ies\w*.{0,80}\bcommentari\w*\b",
        r"\baccurat\w* scholi\w*\b",
        r"\bscholi\w*.{0,80}\bjesu\b",
    ],
    "pedagogical_explanation": [
        r"\bexplain\w*\b",
        r"\bexpliq\w*\b",
        r"\bexplic\w*\b",
        r"\bexplica\w*\b",
        r"\bspiegat\w*\b",
        r"\bverklaring\w*\b",
        r"\bverclaring\w*\b",
        r"\bverkla\w*\b",
        r"\berkla\w*\b",
        r"\bcomment[eé]s?\b",
        r"\bcommentato\b",
        r"\bcommentati\b",
        r"\bcommentary\b",
        r"\bcommentario\b",
        r"\bcommentaire\b",
        r"\bcommentarius\b",
    ],
    "notes_annotations_observations": [
        r"\bnotes?\b",
        r"\bnotis\b",
        r"\bnota\w*\b",
        r"\bannotation\w*\b",
        r"\banmerk\w*\b",
        r"\bobservation\w*\b",
        r"\bobservat\w*\b",
        r"\bremarq\w*\b",
        r"\bremarks?\b",
        r"\banimadversion\w*\b",
        r"\banimaduersion\w*\b",
    ],
    "contracted_or_extracted_commentary": [
        r"\bex maioribus\b.{0,80}\bcommentari\w*\b",
        r"\bex majoribus\b.{0,80}\bcommentari\w*\b",
        r"\bcommentari\w*.{0,80}\bcontract\w*\b",
        r"\bcontract\w*.{0,80}\bcommentari\w*\b",
        r"\bin commodiorem formam contract\w*\b",
        r"\babbrev\w*.{0,80}\bcommentari\w*\b",
        r"\bextract\w*.{0,80}\bcommentari\w*\b",
    ],
    "scholia_general": [
        r"\bscholi\w*\b",
        r"\bscholies?\b",
        r"\bscolii\b",
        r"\bscolies?\b",
    ],
}


def norm(text: str) -> str:
    text = unicodedata.normalize("NFKD", text or "")
    text = "".join(ch for ch in text if not unicodedata.combining(ch))
    text = text.lower()
    text = text.replace("ſ", "s").replace("ß", "ss").replace("æ", "ae").replace("œ", "oe")
    text = re.sub(r"[^a-z0-9]+", " ", text)
    return re.sub(r"\s+", " ", text).strip()


def has_any(text: str, patterns: list[str]) -> bool:
    return any(re.search(pattern, text) for pattern in patterns)


def bool_series(df: pd.DataFrame, col: str) -> pd.Series:
    if col not in df:
        return pd.Series(False, index=df.index)
    if df[col].dtype == bool:
        return df[col].fillna(False)
    return df[col].astype(str).str.lower().isin(["true", "1", "yes", "primary", "secondary"])


def add_commentary_types(df: pd.DataFrame) -> pd.DataFrame:
    evidence_cols = [
        "short_title",
        "rich_claim_text",
        "int_text",
        "value_text",
        "content_description",
        "base_content",
        "enriched_with",
        "edition_details",
        "bound_with",
        "bound_with_minimal",
        "additional_content",
        "author_or_editor",
        "editor_name",
        "editor_description",
        "references_to_euclid",
        "description_of_euclid",
        "elements_designation",
    ]
    cols = [c for c in evidence_cols if c in df]
    out = df.copy()
    out["commentary_evidence"] = out[cols].fillna("").agg(" | ".join, axis=1)
    out["commentary_norm"] = out["commentary_evidence"].map(norm)

    for name, patterns in COMMENTARY_TYPES.items():
        out[name] = out["commentary_norm"].map(lambda text, p=patterns: has_any(text, p))

    # A broader flag that excludes pure "illustrated" unless it is attached to a
    # commentary/scholia/note/explanation term.
    out["any_commentary_split"] = out[list(COMMENTARY_TYPES)].any(axis=1)
    out["commentary_type_count"] = out[list(COMMENTARY_TYPES)].sum(axis=1)
    out["commentary_types"] = out.apply(
        lambda row: ";".join(name for name in COMMENTARY_TYPES if row[name]),
        axis=1,
    )
    return out


def rate_table(df: pd.DataFrame, group_col: str, type_cols: list[str], min_den: int = 1) -> pd.DataFrame:
    rows = []
    for group, sub in df.groupby(group_col, dropna=False):
        group = "missing/unknown" if pd.isna(group) or group == "" else group
        den = len(sub)
        if den < min_den:
            continue
        for col in type_cols:
            count = int(sub[col].sum())
            rows.append(
                {
                    "group": group,
                    "commentary_type": col,
                    "count": count,
                    "denominator": den,
                    "pct": round(100 * count / den, 1) if den else 0,
                }
            )
    return pd.DataFrame(rows)


def pivot_rates(df: pd.DataFrame) -> pd.DataFrame:
    return df.pivot(index="group", columns="commentary_type", values="pct").fillna(0)


def save_heatmap(matrix: pd.DataFrame, path: Path, title: str, figsize=(11, 5)) -> None:
    plt.figure(figsize=figsize)
    sns.heatmap(
        matrix,
        cmap="YlGnBu",
        annot=True,
        fmt=".1f",
        linewidths=0.4,
        linecolor="#eeeeee",
        cbar_kws={"label": "% of group"},
        vmin=0,
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
    elements = add_commentary_types(pd.read_csv(ELEMENTS))
    full = add_commentary_types(pd.read_csv(FULL))
    full["corpus_group"] = full["is_metadata_elements_representative"].astype(str).str.lower().eq("true").map(
        {True: "Metadata Elements", False: "Non-Elements"}
    )

    type_cols = list(COMMENTARY_TYPES) + ["any_commentary_split"]

    case_cols = [
        "classification_key",
        "short_title",
        "year",
        "period",
        "city",
        "language_first",
        "format_group",
        "elements_books_group",
        "natural_dominant_mode",
        "author_or_editor",
        "commentary_type_count",
        "commentary_types",
        "commentary_evidence",
    ] + type_cols

    elements[elements["any_commentary_split"]][case_cols].to_csv(
        TABLES / "report_commentary_split_elements_cases.csv",
        index=False,
    )

    # Corpus comparison.
    corpus_long = rate_table(full, "corpus_group", type_cols, min_den=1)
    corpus_long.to_csv(TABLES / "report_commentary_split_by_corpus_long.csv", index=False)
    corpus_matrix = pivot_rates(corpus_long)
    corpus_matrix.to_csv(TABLES / "commentary_split_by_corpus_matrix.csv")
    save_heatmap(
        corpus_matrix,
        FIGURES / "heatmap_commentary_split_by_corpus.png",
        "Commentary / Explanation Types: Elements vs Non-Elements",
        figsize=(10, 2.8),
    )

    # Elements-internal comparisons.
    for group_col, min_den in [
        ("elements_books_group", 5),
        ("natural_dominant_mode", 5),
        ("period", 3),
        ("language_first", 3),
        ("format_group", 3),
    ]:
        long = rate_table(elements, group_col, type_cols, min_den=min_den)
        long.to_csv(TABLES / f"commentary_split_by_{group_col}_long.csv", index=False)
        matrix = pivot_rates(long)
        matrix_name = f"commentary_split_by_{group_col}_matrix.csv"
        figure_name = f"heatmap_commentary_split_by_{group_col}.png"
        if group_col == "elements_books_group":
            matrix_name = "commentary_by_elements_book_group_matrix.csv"
            figure_name = "heatmap_commentary_by_elements_book_group.png"
        matrix.to_csv(TABLES / matrix_name)
        if group_col in {"elements_books_group", "natural_dominant_mode", "period", "language_first"}:
            save_heatmap(
                matrix,
                FIGURES / figure_name,
                f"Commentary / Explanation Types By {group_col}",
                figsize=(12, 6),
            )

    # Pair/co-occurrence matrix among commentary types.
    pair_rows = []
    for a in type_cols:
        for b in type_cols:
            pair_rows.append(
                {
                    "row_type": a,
                    "col_type": b,
                    "count": int((elements[a] & elements[b]).sum()),
                    "row_denominator": int(elements[a].sum()),
                    "pct_of_row_type": round(100 * (elements[a] & elements[b]).sum() / elements[a].sum(), 1)
                    if elements[a].sum()
                    else 0,
                }
            )
    pair_df = pd.DataFrame(pair_rows)
    pair_df.to_csv(TABLES / "report_commentary_split_cooccurrence_long.csv", index=False)
    pair_df.pivot(index="row_type", columns="col_type", values="pct_of_row_type").fillna(0).to_csv(
        TABLES / "report_commentary_split_cooccurrence_matrix.csv"
    )

    # Close-reading shortlist.
    elements[elements["any_commentary_split"]].sort_values(
        ["commentary_type_count", "year", "classification_key"],
        ascending=[False, True, True],
    )[case_cols].head(80).to_csv(TABLES / "report_commentary_split_close_reading_shortlist.csv", index=False)

    lines = [
        "# Commentary Split Outputs",
        "",
        f"Metadata Elements representatives: {len(elements)}",
        f"Elements rows with commentary/scholia/explanation/notes subtype: {int(elements['any_commentary_split'].sum())}",
        f"Full corpus rows with subtype: {int(full['any_commentary_split'].sum())}",
        "",
        "Outputs:",
        "",
        "- `tables/report_commentary_split_elements_cases.csv`",
        "- `tables/report_commentary_split_by_corpus_matrix.csv`",
        "- `tables/report_commentary_split_by_elements_books_group_matrix.csv`",
        "- `tables/report_commentary_split_by_natural_dominant_mode_matrix.csv`",
        "- `tables/report_commentary_split_by_period_matrix.csv`",
        "- `tables/report_commentary_split_by_language_first_matrix.csv`",
        "- `tables/report_commentary_split_by_format_group_matrix.csv`",
        "- `tables/report_commentary_split_cooccurrence_matrix.csv`",
        "- `tables/report_commentary_split_close_reading_shortlist.csv`",
        "- `figures/heatmap_commentary_split_by_corpus.png`",
        "- `figures/heatmap_commentary_split_by_elements_books_group.png`",
        "- `figures/heatmap_commentary_split_by_natural_dominant_mode.png`",
        "- `figures/heatmap_commentary_split_by_period.png`",
        "- `figures/heatmap_commentary_split_by_language_first.png`",
        "",
        "Pattern matching is broad and for navigation. Final claims need close-reading examples.",
    ]
    (REPORT / "REPORT_COMMENTARY_SPLIT_OUTPUTS.md").write_text("\n".join(lines) + "\n", encoding="utf-8")

    print(f"Elements rows: {len(elements)}")
    print(f"Elements commentary rows: {int(elements['any_commentary_split'].sum())}")
    print(f"Full corpus commentary rows: {int(full['any_commentary_split'].sum())}")


if __name__ == "__main__":
    main()
