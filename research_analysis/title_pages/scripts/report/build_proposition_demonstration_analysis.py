#!/usr/bin/env python3
"""Deepen proposition-use and demonstration-use for report sections.

This script focuses only on metadata-defined Elements representatives. It
classifies proposition/demonstration motifs in title-page evidence, then
summarizes them by book group, period, language, format, natural mode, and
major edition clusters.
"""

from __future__ import annotations

import re
import unicodedata
from pathlib import Path

import pandas as pd


ROOT = Path(__file__).resolve().parents[2]
DERIVED = ROOT / "data" / "analysis_ready"
REPORT = ROOT / "results" / "report"
TABLES = REPORT / "tables"

ELEMENTS = DERIVED / "elements_modes.csv"

TABLES.mkdir(parents=True, exist_ok=True)


MOTIFS = {
    "proposition_any": [
        r"\bproposition\w*\b",
        r"\bpropos\w*\b",
        r"\bpropositiones\b",
        r"\bvoorstel\w*\b",
    ],
    "proposition_use_application": [
        r"\buse of (each|every|all) proposition\w*\b",
        r"\buses of (each|every|all) proposition\w*\b",
        r"\busage de chaque proposition\w*\b",
        r"\busage de toutes? les proposition\w*\b",
        r"\bvsage de chaqve proposition\w*\b",
        r"\buso d[ei] ogni proposition\w*\b",
        r"\busum earum concernentibus\b",
        r"\bapplication\w*.*proposition\w*\b",
        r"\bproposition\w*.*application\w*\b",
        r"\bproposition\w*.*use\w*\b",
        r"\bproposition\w*.*vsage\b",
    ],
    "proposition_selection_extraction": [
        r"\bproposition\w* select\w*\b",
        r"\bselect\w* proposition\w*\b",
        r"\bpropositiones select\w*\b",
        r"\bproposition\w*.*ex sequentibus libris\b",
        r"\breliquorum librorum.*proposition\w*\b",
        r"\bproposition\w*.*reliquorum librorum\b",
        r"\bextract\w*.*proposition\w*\b",
        r"\bproposition\w*.*extract\w*\b",
    ],
    "proposition_ordering_reduction": [
        r"\bproposition\w*.*ordinat\w*\b",
        r"\bordinat\w*.*proposition\w*\b",
        r"\breduit? a \d+\s*proposition\w*\b",
        r"\breduced? to \d+\s*proposition\w*\b",
        r"\b\d+\s*proposition\w*\b",
        r"\bnew order.*proposition\w*\b",
        r"\bproposition\w*.*new order\b",
    ],
    "proposition_explanation_commentary": [
        r"\bexpliq\w*.*proposition\w*\b",
        r"\bexplain\w*.*proposition\w*\b",
        r"\bproposition\w*.*explain\w*\b",
        r"\bverklaring\w*.*proposition\w*\b",
        r"\bproposition\w*.*verklaring\w*\b",
        r"\bcorte verclaringen.*proposition\w*\b",
        r"\bshort explanation\w*.*proposition\w*\b",
    ],
    "demonstration_any": [
        r"\bdemonstrat\w*\b",
        r"\bdimonstrat\w*\b",
        r"\bpreuve\w*\b",
        r"\bproofs?\b",
        r"\bproved?\b",
        r"\bprobatio\w*\b",
        r"\bbeweis\w*\b",
    ],
    "demonstration_easy_clear": [
        r"\bdemonstrat\w*.{0,60}\b(easy|easie|facile|faciles|facilior|faciliorem|plain|clear|claire|succinct|succinctes|brevi|breui)\b",
        r"\b(easy|easie|facile|faciles|facilior|faciliorem|plain|clear|claire|succinct|succinctes|brevi|breui).{0,60}\bdemonstrat\w*\b",
        r"\bpreuve\w*.{0,60}\b(facile|claire|succinct)\w*\b",
        r"\b(facile|claire|succinct)\w*.{0,60}\bpreuve\w*\b",
    ],
    "demonstration_new": [
        r"\bnew demonstrat\w*\b",
        r"\bnouvelle?s? demonstrat\w*\b",
        r"\bnouuelles? demonstrations?\b",
        r"\bnew proof\w*\b",
        r"\bdemonstrat\w*.{0,40}\bnew\b",
        r"\bdemonstrat\w*.{0,40}\bnouvelle?\b",
    ],
    "demonstration_order_method": [
        r"\bnew order\b",
        r"\bnouvel ordre\b",
        r"\bnouuelle? ordre\b",
        r"\bordine nuovo\b",
        r"\bmethod\w*.{0,50}\bdemonstrat\w*\b",
        r"\bdemonstrat\w*.{0,50}\bmethod\w*\b",
        r"\bmethode\w*.{0,50}\bdemonstrat\w*\b",
        r"\bdemonstrat\w*.{0,50}\bmethode\w*\b",
    ],
    "demonstration_adapted_to_capacity": [
        r"\bad facil\w* captum\b",
        r"\bfaciliorem captum\b",
        r"\baccommod\w*.{0,50}\bdemonstrat\w*\b",
        r"\bdemonstrat\w*.{0,50}\baccommod\w*\b",
        r"\bcapacity\b",
        r"\bcompréhension\b",
    ],
    "proposition_and_demonstration": [
        r"\bproposition\w*.{0,80}\bdemonstrat\w*\b",
        r"\bdemonstrat\w*.{0,80}\bproposition\w*\b",
    ],
}

CLUSTERS = {
    "Dechales/Reeve/Williams": [
        r"\bdechales\b",
        r"\bdes chales\b",
        r"\bmilliet\b",
        r"\breeve\b",
        r"\bwilliams\b",
    ],
    "Tacquet": [r"\btacquet\b"],
    "Dou": [r"\bdou\b", r"\bjan pietersz\b"],
    "Henrion": [r"\bhenrion\b"],
    "Clavius": [r"\bclavius\b", r"\bclavij\b", r"\bclavio\b"],
    "Commandino": [r"\bcommandino\b"],
    "Arnauld/Port-Royal": [r"\barnauld\b", r"\bport royal\b"],
    "Reyher/Kiel": [r"\breyher\b", r"\bkiel\b"],
}


def norm(text: str) -> str:
    text = unicodedata.normalize("NFKD", text or "")
    text = "".join(ch for ch in text if not unicodedata.combining(ch))
    text = text.lower()
    text = text.replace("ſ", "s").replace("ß", "ss").replace("æ", "ae").replace("œ", "oe")
    text = re.sub(r"[^a-z0-9]+", " ", text)
    return re.sub(r"\s+", " ", text).strip()


def matches(text: str, patterns: list[str]) -> bool:
    return any(re.search(pattern, text) for pattern in patterns)


def rate_table(df: pd.DataFrame, group_col: str, motif_cols: list[str], min_den: int = 1) -> pd.DataFrame:
    rows = []
    for group, sub in df.groupby(group_col, dropna=False):
        group = "missing/unknown" if pd.isna(group) or group == "" else group
        den = len(sub)
        if den < min_den:
            continue
        for motif in motif_cols:
            count = int(sub[motif].sum())
            rows.append(
                {
                    "group": group,
                    "motif": motif,
                    "count": count,
                    "denominator": den,
                    "pct": round(100 * count / den, 1) if den else 0,
                }
            )
    return pd.DataFrame(rows)


def pivot_rates(df: pd.DataFrame) -> pd.DataFrame:
    return df.pivot(index="group", columns="motif", values="pct").fillna(0)


def main() -> None:
    df = pd.read_csv(ELEMENTS)

    evidence_cols = [
        "short_title",
        "rich_claim_text",
        "int_text",
        "value_text",
        "content_description",
        "base_content",
        "enriched_with",
        "edition_details",
        "author_or_editor",
        "editor_name",
        "references_to_euclid",
        "description_of_euclid",
        "elements_designation",
    ]
    df["prop_demo_evidence"] = df[[c for c in evidence_cols if c in df]].fillna("").agg(" | ".join, axis=1)
    df["prop_demo_norm"] = df["prop_demo_evidence"].map(norm)

    for motif, patterns in MOTIFS.items():
        df[motif] = df["prop_demo_norm"].map(lambda text, p=patterns: matches(text, p))

    def cluster_for(text: str) -> str:
        found = [name for name, pats in CLUSTERS.items() if matches(text, pats)]
        return "; ".join(found) if found else "Other / no major cluster"

    df["major_edition_cluster"] = df["prop_demo_norm"].map(cluster_for)

    motif_cols = list(MOTIFS)
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
        "major_edition_cluster",
        "prop_demo_evidence",
    ] + motif_cols

    cases = df[df[motif_cols].any(axis=1)].copy()
    cases[case_cols].to_csv(TABLES / "report_prop_demo_motif_cases.csv", index=False)

    # Motif rates by core report strata.
    for group_col, min_den in [
        ("elements_books_group", 5),
        ("natural_dominant_mode", 5),
        ("period", 3),
        ("language_first", 3),
        ("format_group", 3),
        ("major_edition_cluster", 3),
    ]:
        long = rate_table(df, group_col, motif_cols, min_den=min_den)
        long.to_csv(TABLES / f"proposition_demonstration_motifs_by_{group_col}_long.csv", index=False)
        output_name = f"proposition_demonstration_motifs_by_{group_col}_matrix.csv"
        if group_col == "elements_books_group":
            output_name = "proposition_demonstration_motifs_by_elements_book_group.csv"
        pivot_rates(long).to_csv(TABLES / output_name)

    # Cluster controls for key book groups.
    controls = []
    for group in sorted(df["elements_books_group"].fillna("missing/unknown").unique()):
        sub = df[df["elements_books_group"].fillna("missing/unknown") == group]
        if len(sub) < 5:
            continue
        for excluded in [
            "Dechales/Reeve/Williams",
            "Tacquet",
            "Dou",
            "Henrion",
            "Clavius",
            "Dechales/Reeve/Williams; Tacquet",
        ]:
            filtered = sub[~sub["major_edition_cluster"].str.contains(re.escape(excluded), na=False)]
            if len(filtered) < 3:
                continue
            for motif in motif_cols:
                controls.append(
                    {
                        "elements_books_group": group,
                        "excluded_cluster_pattern": excluded,
                        "motif": motif,
                        "count_after_exclusion": int(filtered[motif].sum()),
                        "denominator_after_exclusion": len(filtered),
                        "pct_after_exclusion": round(100 * filtered[motif].sum() / len(filtered), 1),
                        "original_count": int(sub[motif].sum()),
                        "original_denominator": len(sub),
                        "original_pct": round(100 * sub[motif].sum() / len(sub), 1),
                    }
                )
    pd.DataFrame(controls).to_csv(TABLES / "report_prop_demo_cluster_exclusion_controls.csv", index=False)

    # A compact top-cases list for close reading.
    df["prop_demo_motif_count"] = df[motif_cols].sum(axis=1)
    top_cases = df[df["prop_demo_motif_count"] > 0].sort_values(
        ["prop_demo_motif_count", "year", "classification_key"],
        ascending=[False, True, True],
    )
    top_cases[case_cols + ["prop_demo_motif_count"]].head(80).to_csv(
        TABLES / "report_prop_demo_close_reading_shortlist.csv",
        index=False,
    )

    summary_lines = [
        "# Proposition And Demonstration Motif Deep Dive",
        "",
        f"Metadata Elements representatives: {len(df)}",
        f"Rows with any proposition/demonstration motif: {int(df[motif_cols].any(axis=1).sum())}",
        "",
        "## Outputs",
        "",
        "- `tables/report_prop_demo_motif_cases.csv`",
        "- `tables/report_prop_demo_motifs_by_elements_books_group_matrix.csv`",
        "- `tables/report_prop_demo_motifs_by_natural_dominant_mode_matrix.csv`",
        "- `tables/report_prop_demo_motifs_by_period_matrix.csv`",
        "- `tables/report_prop_demo_motifs_by_language_first_matrix.csv`",
        "- `tables/report_prop_demo_motifs_by_format_group_matrix.csv`",
        "- `tables/report_prop_demo_motifs_by_major_edition_cluster_matrix.csv`",
        "- `tables/report_prop_demo_cluster_exclusion_controls.csv`",
        "- `tables/report_prop_demo_close_reading_shortlist.csv`",
        "",
        "Pattern matching is broad and designed for navigation. Use case rows for close reading before final prose.",
    ]
    (REPORT / "REPORT_PROP_DEMO_OUTPUTS.md").write_text("\n".join(summary_lines) + "\n", encoding="utf-8")

    print(f"Elements rows: {len(df)}")
    print(f"Rows with any motif: {int(df[motif_cols].any(axis=1).sum())}")
    print(f"Wrote proposition/demonstration outputs to {TABLES}")


if __name__ == "__main__":
    main()
