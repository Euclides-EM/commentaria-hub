#!/usr/bin/env python3
"""Check frozen research memberships against retained data and live metadata."""

from __future__ import annotations

import csv
import hashlib
from pathlib import Path


ROOT = Path(__file__).resolve().parents[3]
ANALYSIS = ROOT / "research_analysis"
TITLE_BASE = ANALYSIS / "title_pages" / "data" / "base_keys"


def key_file(path: Path) -> set[str]:
    return {line.strip() for line in path.read_text(encoding="utf-8").splitlines() if line.strip()}


def csv_keys(path: Path, column: str) -> set[str]:
    with path.open(encoding="utf-8-sig", newline="") as handle:
        return {row[column].strip() for row in csv.DictReader(handle) if row[column].strip()}


def compare(label: str, frozen: set[str], actual: set[str], allow_new: bool = False) -> list[str]:
    errors = []
    missing = sorted(frozen - actual)
    new = sorted(actual - frozen)
    if missing:
        errors.append(f"{label}: missing frozen keys: {', '.join(missing)}")
    if new:
        prefix = "live keys outside frozen set (not included)" if allow_new else "unexpected keys"
        print(f"{label}: {prefix}: {', '.join(new)}")
        if not allow_new:
            errors.append(f"{label}: retained data and frozen keys differ")
    print(f"{label}: frozen={len(frozen)} actual={len(actual)}")
    return errors


def main() -> None:
    errors = []
    errors += compare(
        "edition classification",
        key_file(ANALYSIS / "edition_classification/data/edition_classification_keys.txt"),
        csv_keys(ANALYSIS / "edition_classification/data/edition_subject_classifications_reviewed.csv", "Page/Key"),
    )
    errors += compare(
        "dotted lines",
        key_file(ANALYSIS / "dotted_lines/data/dotted_lines_keys.txt"),
        csv_keys(ANALYSIS / "dotted_lines/data/classifications.csv", "edition_key"),
    )
    title_rows = ANALYSIS / "title_pages/data/analysis_ready/title_page_corpus.csv"
    errors += compare("title pages", key_file(TITLE_BASE / "title_page_keys.txt"), csv_keys(title_rows, "title_page_key"))
    errors += compare(
        "title-page representatives",
        key_file(TITLE_BASE / "title_page_representative_keys.txt"),
        csv_keys(title_rows, "classification_key"),
    )
    errors += compare(
        "Elements representatives",
        key_file(TITLE_BASE / "title_page_elements_representative_keys.txt"),
        csv_keys(ANALYSIS / "title_pages/data/analysis_ready/elements_modes.csv", "classification_key"),
    )
    errors += compare(
        "Elements print geography",
        key_file(TITLE_BASE / "title_page_elements_print_geography_keys.txt"),
        csv_keys(ROOT / "ocrflow/store/items_metadata/metadata_elements_print.csv", "key"),
        allow_new=True,
    )

    checksum_path = TITLE_BASE / "metadata_checksums.sha256"
    for line in checksum_path.read_text(encoding="utf-8").splitlines():
        expected, relative = line.split("  ", 1)
        path = ROOT / relative
        actual = hashlib.sha256(path.read_bytes()).hexdigest()
        if actual != expected:
            print(f"metadata changed: {relative}")

    if errors:
        raise SystemExit("\n".join(errors))
    print("Base-key integrity check passed.")


if __name__ == "__main__":
    main()
