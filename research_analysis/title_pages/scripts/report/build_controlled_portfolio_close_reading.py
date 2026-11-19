#!/usr/bin/env python3
"""Select controlled same-person Elements/non-Elements comparison pairs."""

from __future__ import annotations

import csv
from collections import Counter, defaultdict
from pathlib import Path


ROOT = Path(__file__).resolve().parents[2]
RESULTS = ROOT / "results" / "report" / "tables"
CASE_MATRIX = RESULTS / "author_editor_portfolio_case_matrix.csv"
PAIR_MATRIX = RESULTS / "author_editor_portfolio_elements_non_elements_pairs.csv"


CLAIM_COLS = [
    "claim_method_demonstration_order",
    "claim_accessibility_clarity_pedagogy",
    "claim_utility_practice_application",
    "claim_correction_revision_accuracy",
    "claim_augmentation_enrichment_composition",
    "claim_translation_vernacularization_transfer",
    "claim_ancient_authority_restoration",
    "claim_visual_material_aids",
    "claim_selection_extraction_abridgment",
]

SOCIAL_COLS = [
    "pedagogical_school_arena",
    "court_state_service_arena",
    "military_fortification_arena",
    "professional_practical_arena",
    "religious_institutional_arena",
    "learned_scholarly_arena",
    "general_public_lovers_arena",
    "patronage_prestige_arena",
    "no_visible_social_arena",
]


def read_csv(path: Path) -> list[dict[str, str]]:
    with path.open(encoding="utf-8-sig", newline="") as f:
        return list(csv.DictReader(f))


def write_csv(path: Path, rows: list[dict[str, object]], fieldnames: list[str]) -> None:
    with path.open("w", encoding="utf-8", newline="") as f:
        writer = csv.DictWriter(f, fieldnames=fieldnames)
        writer.writeheader()
        writer.writerows(rows)


def flags(text: str) -> set[str]:
    return {part.strip() for part in (text or "").split("; ") if part.strip()}


def pct(num: int, den: int) -> str:
    return "" if not den else f"{100 * num / den:.1f}"


def main() -> None:
    cases = {r["classification_key"]: r for r in read_csv(CASE_MATRIX)}
    pairs = read_csv(PAIR_MATRIX)

    controlled = []
    for pair in pairs:
        if int(pair["year_gap"]) > 5:
            continue
        if pair["same_city"] != "1" or pair["same_language_first"] != "1":
            continue
        e = cases[pair["elements_key"]]
        n = cases[pair["non_elements_key"]]
        e_claims = flags(e["claim_flags"])
        n_claims = flags(n["claim_flags"])
        e_social = flags(e["social_flags"])
        n_social = flags(n["social_flags"])
        row = dict(pair)
        row.update({
            "control_level": "same_person_city_language_within_5y",
            "same_format_strict": pair["same_format"],
            "elements_short_title": e["short_title"],
            "non_elements_short_title": n["short_title"],
            "elements_publisher": e["publisher"],
            "non_elements_publisher": n["publisher"],
            "elements_rich_claim_text": e["rich_claim_text"],
            "non_elements_rich_claim_text": n["rich_claim_text"],
            "elements_rich_social_text": e["rich_social_text"],
            "non_elements_rich_social_text": n["rich_social_text"],
            "elements_only_claims": "; ".join(sorted(e_claims - n_claims)),
            "non_elements_only_claims": "; ".join(sorted(n_claims - e_claims)),
            "shared_claims": "; ".join(sorted(e_claims & n_claims)),
            "elements_only_social": "; ".join(sorted(e_social - n_social)),
            "non_elements_only_social": "; ".join(sorted(n_social - e_social)),
            "shared_social": "; ".join(sorted(e_social & n_social)),
        })
        controlled.append(row)

    controlled.sort(key=lambda r: (r["person"], int(r["year_gap"]), r["elements_key"], r["non_elements_key"]))

    # Collapse repeated reprint/cluster-like pair explosions by person and element case.
    shortlist = []
    seen = set()
    for row in controlled:
        key = (row["person"], row["elements_key"])
        if key in seen:
            continue
        seen.add(key)
        shortlist.append(row)

    by_person = defaultdict(list)
    for row in controlled:
        by_person[row["person"]].append(row)

    summary_rows = []
    for person, rows in sorted(by_person.items()):
        e_keys = sorted({r["elements_key"] for r in rows})
        n_keys = sorted({r["non_elements_key"] for r in rows})
        e_claim_counter = Counter()
        n_claim_counter = Counter()
        e_social_counter = Counter()
        n_social_counter = Counter()
        strict_format = 0
        for r in rows:
            if r["same_format"] == "1":
                strict_format += 1
            e_claim_counter.update(flags(r["elements_claim_flags"]))
            n_claim_counter.update(flags(r["non_elements_claim_flags"]))
            e_social_counter.update(flags(r["elements_social_flags"]))
            n_social_counter.update(flags(r["non_elements_social_flags"]))
        claim_deltas = []
        for col in CLAIM_COLS:
            e_rate = e_claim_counter[col] / len(rows)
            n_rate = n_claim_counter[col] / len(rows)
            if abs(e_rate - n_rate) >= 0.25:
                claim_deltas.append(f"{col}:{100*(e_rate-n_rate):+.0f}pp")
        social_deltas = []
        for col in SOCIAL_COLS:
            e_rate = e_social_counter[col] / len(rows)
            n_rate = n_social_counter[col] / len(rows)
            if abs(e_rate - n_rate) >= 0.25:
                social_deltas.append(f"{col}:{100*(e_rate-n_rate):+.0f}pp")
        summary_rows.append({
            "person": person,
            "controlled_pairs": len(rows),
            "strict_same_format_pairs": strict_format,
            "elements_cases": " | ".join(e_keys),
            "non_elements_cases": " | ".join(n_keys),
            "elements_case_count": len(e_keys),
            "non_elements_case_count": len(n_keys),
            "claim_deltas_25pp_or_more": "; ".join(claim_deltas),
            "social_deltas_25pp_or_more": "; ".join(social_deltas),
        })

    fields = list(controlled[0].keys()) if controlled else []
    write_csv(RESULTS / "controlled_same_person_pairs.csv", controlled, fields)
    write_csv(RESULTS / "controlled_author_editor_close_pair_shortlist.csv", shortlist, fields)
    write_csv(RESULTS / "controlled_same_person_summary.csv", summary_rows, list(summary_rows[0].keys()) if summary_rows else [])

    print(f"Controlled pairs: {len(controlled)}")
    print(f"People: {len(by_person)}")
    print(f"Strict same-format controlled pairs: {sum(1 for r in controlled if r['same_format'] == '1')}")
    print("Top people:")
    for person, rows in sorted(by_person.items(), key=lambda kv: (-len(kv[1]), kv[0]))[:12]:
        e = len({r["elements_key"] for r in rows})
        n = len({r["non_elements_key"] for r in rows})
        print(f"  {person}: {len(rows)} pairs, {e} Elements cases, {n} non-Elements cases")


if __name__ == "__main__":
    main()
