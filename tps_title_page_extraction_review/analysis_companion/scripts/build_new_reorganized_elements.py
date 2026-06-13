#!/usr/bin/env python3
"""Find Elements title pages that claim novelty, new order, or reconstruction."""

from __future__ import annotations

import csv
import re
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
DERIVED = ROOT / "derived_data"
MATRIX = DERIVED / "metadata_elements_natural_modes_matrix_with_format.csv"


PATTERNS = {
    "new_elements_or_new_geometry": [
        r"\bnouveaux?\s+elem", r"\bnew\s+elements?", r"\beuclides?\s+reform", r"\bnew\s+geometry",
        r"\bnov[aoui]s?\s+element", r"\bnuov[oa]\s+element",
    ],
    "new_order_or_arrangement": [
        r"ordre tout nouveau", r"nouvel ordre", r"new order", r"nuov.?ordine", r"nova methodo",
        r"nouvelle methode", r"nouvelle méthode", r"new.*method", r"neue.*art", r"sonderbahre.*art",
        r"ita ordinat", r"ordinat[ae]", r"digesta", r"disposita", r"reduit", r"réduit",
    ],
    "new_demonstrations": [
        r"nouvelles? demonstr", r"new.*demonstr", r"novis demonstration", r"novis vbique.*demonstr",
        r"novis.*demonstrationibus", r"neue.*demonstr", r"nuov.*dimostr", r"demonstr.*new",
    ],
    "pedagogical_ease": [
        r"facile", r"tres[- ]facile", r"easie", r"easy", r"leichte", r"faciliorem captum",
        r"brevi", r"succinct", r"courte", r"commodiorem formam", r"commodiùs",
    ],
    "selection_reduction_contraction": [
        r"select", r"contract", r"abr[ée]g", r"reduit.*propositions?", r"propositiones select",
        r"extraction", r"extraict", r"excerpt", r"succinct",
    ],
    "analytic_or_symbolic_retooling": [
        r"algebra", r"notes reelles", r"notas reales", r"symbolic", r"symboliq", r"zeichen",
        r"løse-kunst", r"analyse", r"analysis",
    ],
    "anti_ancient_or_modern_comparison": [
        r"antiquos.*recentiores", r"recentiores", r"modern", r"antea ab aliis", r"nunquam ante",
        r"point encore avis", r"nouveaux moyens", r"nouvelles mesures", r"nouvelles manières",
    ],
}

PHILOLOGY_PATTERNS = [
    r"restitu", r"restore", r"restored", r"ancient", r"antiqu", r"veter", r"theon", r"gr[æe]c",
    r"commandini", r"correct", r"corrig", r"emend", r"castig", r"recognit",
]


def read_csv(path: Path) -> list[dict[str, str]]:
    with path.open(encoding="utf-8-sig", newline="") as f:
        return list(csv.DictReader(f))


def write_csv(path: Path, rows: list[dict[str, object]], fields: list[str]) -> None:
    with path.open("w", encoding="utf-8", newline="") as f:
        w = csv.DictWriter(f, fieldnames=fields, extrasaction="ignore")
        w.writeheader()
        w.writerows(rows)


def hit_patterns(text: str, patterns: list[str]) -> list[str]:
    hits = []
    for pat in patterns:
        if re.search(pat, text, flags=re.I):
            hits.append(pat)
    return hits


def truthy(v: str | None) -> bool:
    return str(v or "").strip().lower() in {"1", "true", "yes", "y"}


def main() -> None:
    rows = read_csv(MATRIX)
    out = []
    for r in rows:
        text = " | ".join([
            r.get("short_title", ""),
            r.get("action_verbs", ""),
            r.get("base_content", ""),
            r.get("content_description", ""),
            r.get("edition_details", ""),
            r.get("elements_designation", ""),
            r.get("enriched_with", ""),
            r.get("rich_claim_text", ""),
        ])
        text_l = text.lower()
        motif_hits: dict[str, list[str]] = {}
        for motif, pats in PATTERNS.items():
            hits = hit_patterns(text_l, pats)
            if hits:
                motif_hits[motif] = hits
        if not motif_hits:
            continue
        philology_hits = hit_patterns(text_l, PHILOLOGY_PATTERNS)
        novelty_score = len(motif_hits)
        philology_score = len(set(philology_hits))
        if motif_hits.get("new_elements_or_new_geometry"):
            pole = "strong_new_elements"
        elif motif_hits.get("new_order_or_arrangement") and motif_hits.get("new_demonstrations"):
            pole = "new_order_and_demonstration"
        elif motif_hits.get("new_order_or_arrangement") or motif_hits.get("new_demonstrations"):
            pole = "reordered_or_new_demonstrations"
        elif motif_hits.get("selection_reduction_contraction"):
            pole = "selected_reduced_contracted"
        else:
            pole = "method_ease_or_retooling"
        if philology_score and novelty_score:
            relation = "hybrid_with_philological_or_correction_rhetoric"
        else:
            relation = "mostly_reconstructive_rhetoric"
        out.append({
            "classification_key": r["classification_key"],
            "year": r.get("year", ""),
            "period": r.get("period", ""),
            "city": r.get("city", ""),
            "language_first": r.get("language_first", ""),
            "author_or_editor": r.get("author_or_editor", ""),
            "elements_books": r.get("elements_books", ""),
            "elements_books_group": r.get("elements_books_group", ""),
            "primary_subject_family": r.get("primary_subject_family", ""),
            "natural_dominant_mode": r.get("natural_dominant_mode", ""),
            "reconstructive_pole": pole,
            "philology_relation": relation,
            "motif_count": novelty_score,
            "motifs": " | ".join(motif_hits.keys()),
            "matched_patterns": " | ".join(f"{k}: {', '.join(v)}" for k, v in motif_hits.items()),
            "philology_patterns": " | ".join(sorted(set(philology_hits))),
            "claim_novelty_modernity_invention": r.get("claim_novelty_modernity_invention", ""),
            "claim_method_demonstration_order": r.get("claim_method_demonstration_order", ""),
            "claim_accessibility_clarity_pedagogy": r.get("claim_accessibility_clarity_pedagogy", ""),
            "claim_correction_revision_accuracy": r.get("claim_correction_revision_accuracy", ""),
            "claim_ancient_authority_restoration": r.get("claim_ancient_authority_restoration", ""),
            "rich_claim_text": r.get("rich_claim_text", ""),
        })
    out.sort(key=lambda r: (r["reconstructive_pole"], r["year"], r["classification_key"]))
    fields = list(out[0].keys()) if out else []
    write_csv(DERIVED / "new_reorganized_elements_cases.csv", out, fields)

    # Summary by pole.
    summary = []
    for pole in sorted({r["reconstructive_pole"] for r in out}):
        subset = [r for r in out if r["reconstructive_pole"] == pole]
        summary.append({
            "reconstructive_pole": pole,
            "n": len(subset),
            "keys": " | ".join(r["classification_key"] for r in subset),
            "periods": " | ".join(sorted({r["period"] for r in subset if r["period"]})),
            "languages": " | ".join(sorted({r["language_first"] for r in subset if r["language_first"]})),
            "claim_novelty_pct": round(100 * sum(truthy(r["claim_novelty_modernity_invention"]) for r in subset) / len(subset), 1),
            "method_pct": round(100 * sum(truthy(r["claim_method_demonstration_order"]) for r in subset) / len(subset), 1),
            "ancient_restoration_pct": round(100 * sum(truthy(r["claim_ancient_authority_restoration"]) for r in subset) / len(subset), 1),
            "hybrid_philology_count": sum(r["philology_relation"].startswith("hybrid") for r in subset),
        })
    write_csv(DERIVED / "new_reorganized_elements_summary.csv", summary, list(summary[0].keys()) if summary else [])

    print(f"New/reorganized candidates: {len(out)}")
    for s in summary:
        print(f"{s['reconstructive_pole']}: {s['n']}")


if __name__ == "__main__":
    main()
