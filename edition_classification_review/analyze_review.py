#!/usr/bin/env python3
import csv
import math
from collections import Counter, defaultdict
from pathlib import Path


ROOT = Path(__file__).resolve().parent
INPUT_CANDIDATES = [
    ROOT / "features-comparison-May-23 - editions_results_with_v7.csv",
    ROOT / "features-comparison-May-23 - editions_results_with_v6.csv",
    ROOT / "features-comparison-May-23 - editions_results.csv",
]
FULL = ROOT / "full_row_suggestions.csv"
SUMMARY = ROOT / "kpi_summary.csv"
REPORT = ROOT / "analysis_report.md"

BASE_LLM_COLS = ["llm_Value_1", "llm_Value_3", "llm_Value_4", "llm_Value_6", "llm_Value_7"]
LABELS = {"primary", "secondary", "unrelated", "unknown", ""}
MANUAL_LABELS = {"primary", "secondary", "unrelated"}
RELATED = {"primary", "secondary"}
VALUE_RANK = {"primary": 4, "secondary": 3, "unrelated": 2, "unknown": 1, "": 0}
BASE_FIELDS = ["Page/Key", "language", "year", "city", "Classification", "Orig_Value"]


def input_path():
    for path in INPUT_CANDIDATES:
        if path.exists():
            return path
    raise FileNotFoundError("no edition classification comparison CSV found")


def clean(row, col):
    return (row.get(col) or "").strip()


def pct(x):
    return "" if x is None or math.isnan(x) else f"{x:.1%}"


def pp(x):
    return "" if x is None or math.isnan(x) else f"{x * 100:+.1f} pp"


def prompt_label(col):
    return col.replace("llm_Value_", "v")


def review_output_path(latest_col):
    if latest_col == "llm_Value_6":
        return ROOT / "v6_diagnostic_review_regenerated.csv"
    return ROOT / f"{prompt_label(latest_col)}_diagnostic_review.csv"


def majority_threshold(llm_cols):
    return len(llm_cols) // 2 + 1


def read_rows(path):
    with path.open(newline="", encoding="utf-8-sig") as f:
        rows = list(csv.DictReader(f))
    llm_cols = [col for col in BASE_LLM_COLS if col in rows[0]]
    bad = [(i + 2, c, clean(r, c)) for i, r in enumerate(rows) for c in ["Orig_Value", *llm_cols] if clean(r, c) not in LABELS]
    if bad:
        raise ValueError(f"unexpected labels: {bad[:10]}")
    return rows, llm_cols


def metrics(rows, col):
    eval_rows = [r for r in rows if clean(r, "Orig_Value") in MANUAL_LABELS]
    n = len(eval_rows)
    if not n:
        return {}
    exact = sum(clean(r, col) == clean(r, "Orig_Value") for r in eval_rows)
    unknown = sum(clean(r, col) == "unknown" for r in eval_rows)
    covered = [r for r in eval_rows if clean(r, col) != "unknown"]
    covered_exact = sum(clean(r, col) == clean(r, "Orig_Value") for r in covered)
    tp = sum(clean(r, "Orig_Value") in RELATED and clean(r, col) in RELATED for r in eval_rows)
    fp = sum(clean(r, "Orig_Value") == "unrelated" and clean(r, col) in RELATED for r in eval_rows)
    fn = sum(clean(r, "Orig_Value") in RELATED and clean(r, col) == "unrelated" for r in eval_rows)
    unk_pos = sum(clean(r, "Orig_Value") in RELATED and clean(r, col) == "unknown" for r in eval_rows)
    return {
        "manual_rows": n,
        "exact": exact / n,
        "covered_exact": covered_exact / len(covered) if covered else None,
        "unknown": unknown / n,
        "related_precision": tp / (tp + fp) if tp + fp else None,
        "related_recall": tp / (tp + fn + unk_pos) if tp + fn + unk_pos else None,
        "true_related": tp,
        "false_related": fp,
        "missed_related": fn,
        "unknown_when_related": unk_pos,
    }


def consensus(row, llm_cols):
    votes = [clean(row, c) for c in llm_cols]
    counts = Counter(votes)
    value, count = counts.most_common(1)[0]
    if count >= majority_threshold(llm_cols):
        return value, count
    return max(votes, key=lambda v: VALUE_RANK.get(v, 0)), 1


def status_for(row, llm_cols):
    manual = clean(row, "Orig_Value")
    votes = [clean(row, c) for c in llm_cols]
    positives = [v for v in votes if v in RELATED]
    cons, count = consensus(row, llm_cols)

    if manual in RELATED:
        if positives:
            return manual, "keep_manual_positive", "manual positive; at least one LLM also says related"
        return manual, "review_manual_positive", "manual positive but no LLM confirms"
    if manual == "unrelated":
        if len(positives) >= majority_threshold(llm_cols):
            return cons, "majority_llm_related_manual_unrelated", "manual unrelated but an LLM majority says related"
        if len(positives) > 1:
            return "unrelated", "split_llm_related_manual_unrelated", "manual unrelated but multiple LLMs say related without a majority"
        if len(positives) == 1:
            return "unrelated", "single_llm_related_manual_unrelated", "manual unrelated but one LLM says related"
        return "unrelated", "stable_unrelated", "manual unrelated and LLMs do not give a related majority"
    if count >= 2 and cons in RELATED:
        return cons, "blank_with_llm_related_consensus", "blank manual value; 2-3 LLMs say related"
    if count >= 2 and cons == "unrelated":
        return "unrelated", "blank_with_llm_unrelated_consensus", "blank manual value; LLM majority says unrelated"
    if votes.count("unknown") >= 2:
        return "unknown", "blank_needs_evidence", "blank manual value; LLM majority abstains"
    return cons, "blank_no_majority", "blank manual value; no LLM majority"


def review_question(status):
    if status == "majority_llm_related_manual_unrelated":
        return "Is the manual unrelated label wrong, or are the LLMs over-classifying? If related, primary or secondary?"
    if status == "review_manual_positive":
        return "Is the manual positive label truly supported by the metadata? If yes, what clue should V6 notice?"
    if status == "blank_with_llm_related_consensus":
        return "Can we fill this blank as related? If yes, is the value primary or secondary?"
    if status == "single_llm_related_manual_unrelated":
        return "Is this a useful weak signal or a false positive pattern V6 should suppress?"
    if status == "split_llm_related_manual_unrelated":
        return "Is the split related signal a manual miss or lingering over-classification?"
    if status == "blank_no_majority":
        return "Which final value is defensible from the metadata?"
    return ""


def priority_score(row, status):
    classification = clean(row, "Classification")
    noisy = {"Theoretical Mathematics", "Practical Geometry", "Instrument Use", "Instrument Construction", "Construction", "Trigonometry"}
    score = {
        "majority_llm_related_manual_unrelated": 100,
        "blank_with_llm_related_consensus": 90,
        "review_manual_positive": 85,
        "single_llm_related_manual_unrelated": 40,
        "blank_no_majority": 35,
        "blank_needs_evidence": 20,
    }.get(status, 0)
    if classification in noisy:
        score += 10
    latest_col = "llm_Value_7" if "llm_Value_7" in row else "llm_Value_6"
    if clean(row, latest_col) in RELATED:
        score += 3
    llm_cols = [col for col in BASE_LLM_COLS if col in row]
    if len({clean(row, c) for c in llm_cols}) == 1:
        score += 2
    return score


def write_csv(path, rows, fieldnames):
    with path.open("w", newline="", encoding="utf-8") as f:
        w = csv.DictWriter(f, fieldnames=fieldnames)
        w.writeheader()
        w.writerows(rows)


def md_table(headers, rows):
    out = ["| " + " | ".join(headers) + " |", "| " + " | ".join(["---"] * len(headers)) + " |"]
    for row in rows:
        out.append("| " + " | ".join(str(v) for v in row) + " |")
    return "\n".join(out)


def main():
    current_input = input_path()
    rows, llm_cols = read_rows(current_input)
    latest_col = llm_cols[-1]
    full_rows = []
    for r in rows:
        cons, count = consensus(r, llm_cols)
        suggestion, status, reason = status_for(r, llm_cols)
        out = {
            **{field: clean(r, field) for field in BASE_FIELDS},
            **{col: clean(r, col) for col in llm_cols},
            "llm_consensus": cons,
            "llm_consensus_count": count,
            "suggested_final_value": suggestion,
            "decision_status": status,
            "decision_reason": reason,
        }
        full_rows.append(out)

    write_csv(FULL, full_rows, list(full_rows[0].keys()))

    review_candidates = [
        {
            **r,
            "review_priority": priority_score(r, r["decision_status"]),
            "review_question": review_question(r["decision_status"]),
            "your_final_value": "",
            "your_error_type": "",
            "your_notes_for_v6": "",
        }
        for r in full_rows
        if r["decision_status"] in {
            "majority_llm_related_manual_unrelated",
            "blank_with_llm_related_consensus",
            "review_manual_positive",
            "single_llm_related_manual_unrelated",
            "split_llm_related_manual_unrelated",
            "blank_no_majority",
        }
    ]
    review_candidates.sort(key=lambda r: (-int(r["review_priority"]), r["Classification"], r["Page/Key"]))

    selected = []
    counts_by_status = Counter()
    counts_by_class = Counter()
    caps = {
        "majority_llm_related_manual_unrelated": 75,
        "blank_with_llm_related_consensus": 30,
        "review_manual_positive": 20,
        "split_llm_related_manual_unrelated": 35,
        "single_llm_related_manual_unrelated": 25,
        "blank_no_majority": 10,
    }
    class_cap = 12
    for r in review_candidates:
        status = r["decision_status"]
        classification = r["Classification"]
        if counts_by_status[status] >= caps.get(status, 0):
            continue
        if counts_by_class[classification] >= class_cap and status != "review_manual_positive":
            continue
        selected.append(r)
        counts_by_status[status] += 1
        counts_by_class[classification] += 1
    for i, r in enumerate(selected, 1):
        r["review_id"] = i

    review_fields = ["review_id"] + [f for f in selected[0].keys() if f != "review_id"]
    review_output = review_output_path(latest_col)
    write_csv(review_output, selected, review_fields)

    summary_rows = []
    for scope_type, groups in [
        ("overall", {"all": rows}),
        ("language", defaultdict(list)),
        ("classification", defaultdict(list)),
    ]:
        if scope_type == "language":
            for r in rows:
                groups[clean(r, "language")].append(r)
        if scope_type == "classification":
            for r in rows:
                groups[clean(r, "Classification")].append(r)
        for scope, subrows in groups.items():
            for col in llm_cols:
                m = metrics(subrows, col)
                if not m:
                    continue
                summary_rows.append({
                    "scope_type": scope_type,
                    "scope_value": scope,
                    "prompt": col,
                    "manual_rows": m["manual_rows"],
                    "exact": pct(m["exact"]),
                    "covered_exact": pct(m["covered_exact"]),
                    "unknown": pct(m["unknown"]),
                    "related_precision": pct(m["related_precision"]),
                    "related_recall": pct(m["related_recall"]),
                    "true_related": m["true_related"],
                    "false_related": m["false_related"],
                    "missed_related": m["missed_related"],
                    "unknown_when_related": m["unknown_when_related"],
                })
    write_csv(SUMMARY, summary_rows, list(summary_rows[0].keys()))

    manual_counts = Counter(clean(r, "Orig_Value") or "blank" for r in rows)
    status_counts = Counter(r["decision_status"] for r in full_rows)
    overall = [(col, metrics(rows, col)) for col in llm_cols]
    v4_metrics = metrics(rows, "llm_Value_4")
    latest_metrics = metrics(rows, latest_col)
    report = [
        "# Edition Classification Analysis Report",
        "",
        f"This uses `{current_input.relative_to(ROOT.parent)}`. `Orig_Value` is treated as an imperfect reference, not ground truth.",
        "",
        "## Dataset",
        "",
        f"- Rows: {len(rows):,}",
        f"- Editions: {len({clean(r, 'Page/Key') for r in rows}):,}",
        f"- Classifications per edition: {sorted(set(Counter(clean(r, 'Page/Key') for r in rows).values()))}",
        f"- Manual labels: " + ", ".join(f"`{k}` {v:,}" for k, v in manual_counts.most_common()),
        "",
        "## KPI Snapshot",
        "",
        md_table(
            ["Prompt", "Exact", "Covered exact", "Unknown", "Related precision", "Related recall", "False related"],
            [
                [prompt_label(col), pct(m["exact"]), pct(m["covered_exact"]), pct(m["unknown"]), pct(m["related_precision"]), pct(m["related_recall"]), m["false_related"]]
                for col, m in overall
            ],
        ),
        "",
        "Interpretation: exact accuracy is limited by noisy manual labels and the dominance of `unrelated`. Related precision, recall, false-related count, and unknown rate are the more useful prompt-comparison signals.",
        "",
        f"## Latest Run Delta vs V4",
        "",
        md_table(
            ["Metric", "v4", prompt_label(latest_col), "Delta"],
            [
                ["Exact", pct(v4_metrics["exact"]), pct(latest_metrics["exact"]), pp(latest_metrics["exact"] - v4_metrics["exact"])],
                ["Covered exact", pct(v4_metrics["covered_exact"]), pct(latest_metrics["covered_exact"]), pp(latest_metrics["covered_exact"] - v4_metrics["covered_exact"])],
                ["Unknown", pct(v4_metrics["unknown"]), pct(latest_metrics["unknown"]), pp(latest_metrics["unknown"] - v4_metrics["unknown"])],
                ["Related precision", pct(v4_metrics["related_precision"]), pct(latest_metrics["related_precision"]), pp(latest_metrics["related_precision"] - v4_metrics["related_precision"])],
                ["Related recall", pct(v4_metrics["related_recall"]), pct(latest_metrics["related_recall"]), pp(latest_metrics["related_recall"] - v4_metrics["related_recall"])],
                ["False related", v4_metrics["false_related"], latest_metrics["false_related"], f"{latest_metrics['false_related'] - v4_metrics['false_related']:+d}"],
            ],
        ),
        "",
        "## Adjudication Buckets",
        "",
        md_table(["Bucket", "Rows"], [[k, v] for k, v in status_counts.most_common()]),
        "",
        "## What To Review Now",
        "",
        f"`{review_output.name}` contains {len(selected)} rows. Fill `your_final_value`, `your_error_type`, and `your_notes_for_v6`.",
        "",
        "Suggested `your_final_value`: `primary`, `secondary`, `unrelated`, `unknown`, or `unsure`.",
        "",
        "Suggested `your_error_type`: `manual_missed_related`, `llm_overclassified`, `primary_secondary_wrong`, `needs_more_metadata`, `definition_unclear`, or `other`.",
        "",
        "## Current DB Run Provenance",
        "",
        "- `llm_Value_6` was merged from the current `ocrflow/store/ocrflow.db` values for `scope='editions'` and `feature_id='m_classifier'`.",
        "- The active DB result metadata reports `source_revision='f96afd86-79f0-4736-a91a-d58e37b6db65'`, which is the stored `v1` revision.",
        "- The intended V6 prompt revision in migrations is `6f4aafde-a8d9-4a50-8cae-7947b470c6f6`; the intended V7 prompt revision is `7a3f3e5a-8f8a-4c47-b1e7-56c31c9ab7d0`.",
        "",
    ]
    REPORT.write_text("\n".join(report), encoding="utf-8")


if __name__ == "__main__":
    main()
