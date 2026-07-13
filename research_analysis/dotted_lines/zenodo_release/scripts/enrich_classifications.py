#!/usr/bin/env python3
"""Filter and enrich the dotted-lines classification table.

The script uses only Python's standard library. By default, paths are resolved
from the repository root, so it can be run from any working directory.
"""

from __future__ import annotations

import argparse
import csv
import json
import re
from pathlib import Path
from typing import Iterable


REPO_ROOT = Path(__file__).resolve().parents[4]
DOTTED_LINES_DIR = REPO_ROOT / "research_analysis/dotted_lines"
ZENODO_RELEASE_DIR = DOTTED_LINES_DIR / "zenodo_release"

DEFAULT_INPUT = DOTTED_LINES_DIR / "data/classifications.csv"
DEFAULT_METADATA_DIR = REPO_ROOT / "ocrflow/store/items_metadata"
DEFAULT_OUTPUT = ZENODO_RELEASE_DIR / "data/classifications_enriched_generated.csv"
DEFAULT_SCANS_OUTPUT = ZENODO_RELEASE_DIR / "data/classification_volumes_scans.csv"
DEFAULT_CROPS_JSON_DIR = ZENODO_RELEASE_DIR / "tmp/crops_json"
DEFAULT_CROPS_OUTPUT = ZENODO_RELEASE_DIR / "data/classification_crops.csv"

ARITHMETICAL_COLUMN = "Elements (Arithmetical Books)"
GEOMETRICAL_COLUMN = "Elements (Geometrical Books)"
CLASSIFICATION_COLUMN = "Has dotted diagrams in the wide sense"

EXCLUDED_OUTPUT_COLUMNS = {
    ARITHMETICAL_COLUMN,
    GEOMETRICAL_COLUMN,
    "Final Architecture",
    "Final Arithmetic",
    "Final Astronomy",
    "Final Cartography",
    "Final Commercial Mathematics",
    "Final Construction",
    "Final Cosmography",
    "Final Geography",
    "Final Gnomonics & Horology",
    "Final Instrument Construction",
    "Final Instrument Use",
    "Final Mechanics",
    "Final Military Engineering",
    "Final Music Theory",
    "Final Navigation",
    "Final Perspective",
    "Final Practical Geometry",
    "Final Surveying",
    "Final Theoretical Mathematics",
    "Final Trigonometry",
}

ENRICHED_OUTPUT_COLUMN_NAMES = {
    "Year": "year",
    "City": "city",
    "Language": "language",
    "Format": "format",
    "Elements (Arithmetical Books)": "elements_arithmetical_books",
    "Elements (Geometrical Books)": "elements_geometrical_books",
    "Geometrical Dotted Lines": "geometrical_dotted_lines",
    "Token Lines": "token_lines",
    'Other "Dotted" Lines': "other_dotted_lines",
    "Has dotted diagrams in the wide sense": "has_dotted_diagrams_wide_sense",
    "RAW Dotted lines": "raw_dotted_lines",
    "Lines classified": "lines_classified",
    "Notes": "notes",
    "Has diagrams": "has_diagrams",
    "Reprint of": "reprint_of",
    "Includes Elements Plane Geometry Books (I-VI)": (
        "includes_elements_plane_geometry_books_1_6"
    ),
    "Includes Elements Arithmetical Books (VII-IX)": (
        "includes_elements_arithmetical_books_7_9"
    ),
    "Includes Elements Book X": "includes_elements_book_10",
    "Includes Elements Solid Geometry Books (XI-XIII)": (
        "includes_elements_solid_geometry_books_11_13"
    ),
    "Reprint of (inferred from clusters)": "reprint_of_inferred_from_clusters",
    "Reprint cluster comparison": "reprint_cluster_comparison",
}

BOOK_GROUP_COLUMNS = {
    "Includes Elements Plane Geometry Books (I-VI)": set(range(1, 7)),
    "Includes Elements Arithmetical Books (VII-IX)": set(range(7, 10)),
    "Includes Elements Book X": {10},
    "Includes Elements Solid Geometry Books (XI-XIII)": set(range(11, 14)),
}

CROP_FILENAME_PATTERN = re.compile(
    r"^(?P<page>\d+)_(?P<type>.+)_(?P<sequence>\d+)\.[^.]+$"
)


def read_rows(path: Path) -> tuple[list[str], list[dict[str, str]]]:
    with path.open(encoding="utf-8-sig", newline="") as handle:
        reader = csv.DictReader(handle)
        if reader.fieldnames is None:
            raise ValueError(f"CSV has no header: {path}")
        return list(reader.fieldnames), list(reader)


def rows_by_unique_key(path: Path) -> dict[str, dict[str, str]]:
    _, rows = read_rows(path)
    keyed: dict[str, dict[str, str]] = {}
    for row in rows:
        key = row["key"].strip()
        if key in keyed:
            raise ValueError(f"Duplicate key {key!r} in {path}")
        keyed[key] = row
    return keyed


def write_rows(path: Path, fieldnames: list[str], rows: Iterable[dict[str, str]]) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    with path.open("w", encoding="utf-8", newline="") as handle:
        writer = csv.DictWriter(handle, fieldnames=fieldnames, extrasaction="ignore")
        writer.writeheader()
        writer.writerows(rows)


def parse_elements_books(value: str) -> tuple[list[int], list[str], str]:
    """Expand book ranges while retaining qualifiers and unknown tokens."""
    value = value.strip()
    if not value:
        return [], [], ""

    qualifier_match = re.search(r"\s*(\([^)]*\))\s*$", value)
    qualifier = qualifier_match.group(1) if qualifier_match else ""
    core = value[: qualifier_match.start()].strip() if qualifier_match else value

    books: list[int] = []
    unknown: list[str] = []
    for token in (part.strip() for part in core.split(",")):
        if not token:
            continue
        range_match = re.fullmatch(r"(\d+)\s*[-–]\s*(\d+)", token)
        if range_match:
            start, end = map(int, range_match.groups())
            books.extend(range(start, end + 1) if start <= end else range(start, end - 1, -1))
        elif token.isdigit():
            books.append(int(token))
        else:
            unknown.append(token)

    # Preserve first occurrence, which matters for unusually ordered source data.
    books = list(dict.fromkeys(books))
    normalized_parts = [str(book) for book in books] + unknown
    normalized = ", ".join(normalized_parts)
    if qualifier:
        normalized = f"{normalized} {qualifier}" if normalized else qualifier
    return books, unknown, normalized


def book_group_flags(books: list[int]) -> dict[str, str]:
    if books:
        present = set(books)
        return {
            column: "TRUE" if present & group else "FALSE"
            for column, group in BOOK_GROUP_COLUMNS.items()
        }

    return {column: "" for column in BOOK_GROUP_COLUMNS}


def compare_reprint(
    edition_key: str,
    classification_value: str,
    inferred_value: str,
    family_by_item: dict[str, frozenset[str]],
) -> str:
    classification_value = classification_value.strip()
    if classification_value:
        family = family_by_item.get(edition_key, frozenset())
        if classification_value in family:
            return "same reprint family"
        return "classification only" if not family else "conflict"
    if inferred_value.strip():
        return "clusters only"
    return "neither"


def reprint_families(
    metadata_dir: Path, years: dict[str, int]
) -> tuple[dict[str, frozenset[str]], dict[str, str]]:
    """Build connected reprint families and infer each family's earliest base."""
    _, cluster_rows = read_rows(metadata_dir / "clusters.csv")
    reprint_clusters = {
        row["key"].strip()
        for row in cluster_rows
        if row.get("type", "").strip().casefold() == "reprint"
    }
    _, membership_rows = read_rows(metadata_dir / "cluster_items.csv")
    members_by_cluster: dict[str, list[str]] = {}
    for row in membership_rows:
        cluster_key = row.get("cluster_key", "").strip()
        item_key = row.get("item_key", "").strip()
        if cluster_key in reprint_clusters and item_key:
            members_by_cluster.setdefault(cluster_key, []).append(item_key)

    adjacency: dict[str, set[str]] = {}
    for members in members_by_cluster.values():
        member_set = set(members)
        for member in member_set:
            adjacency.setdefault(member, set()).update(member_set - {member})

    family_by_item: dict[str, frozenset[str]] = {}
    inferred_base: dict[str, str] = {}
    unseen = set(adjacency)
    while unseen:
        seed = unseen.pop()
        component = {seed}
        stack = [seed]
        while stack:
            current = stack.pop()
            new_members = adjacency[current] - component
            component.update(new_members)
            unseen.difference_update(new_members)
            stack.extend(new_members)
        family = frozenset(component)
        for member in family:
            family_by_item[member] = family
        dated = [(years[item], item) for item in family if item in years]
        if dated:
            base = min(dated)[1]
            for member in family - {base}:
                inferred_base[member] = base
    return family_by_item, inferred_base


def corpus_for(row: dict[str, str]) -> str:
    is_main = (
        row.get(ARITHMETICAL_COLUMN, "").strip().upper() == "V"
        or row.get(GEOMETRICAL_COLUMN, "").strip().upper() == "V"
    )
    return "Main Elements" if is_main else "Additional Work"


def crop_image_names(path: Path) -> Iterable[str]:
    """Yield image names from either a single- or multi-volume crop JSON."""
    with path.open(encoding="utf-8") as handle:
        data = json.load(handle)

    if "images" in data:
        groups = [data]
    elif "volumes" in data:
        groups = data["volumes"]
    else:
        raise ValueError(f"Crop JSON has neither images nor volumes: {path}")

    for group in groups:
        images = group.get("images", [])
        if not isinstance(images, list):
            raise ValueError(f"Crop JSON images must be a list: {path}")
        yield from images


def crop_rows(crops_json_dir: Path, edition_keys: Iterable[str]) -> list[dict[str, str]]:
    rows: list[dict[str, str]] = []
    for edition_key in edition_keys:
        path = crops_json_dir / f"{edition_key}.json"
        if not path.is_file():
            continue

        for image_name in crop_image_names(path):
            match = CROP_FILENAME_PATTERN.fullmatch(Path(image_name).name)
            if match is None:
                raise ValueError(f"Unexpected crop image filename in {path}: {image_name!r}")
            rows.append(
                {
                    "edition_key": edition_key,
                    "page_in_facsimile": str(int(match.group("page"))),
                    "sequence_on_page": str(int(match.group("sequence"))),
                    "automatically_identified_type": match.group("type").replace("_", " "),
                }
            )
    return rows


def enrich(args: argparse.Namespace) -> tuple[int, int, int, int]:
    input_fields, classification_rows = read_rows(args.input)
    items = rows_by_unique_key(args.metadata_dir / "items_print.csv")
    elements = rows_by_unique_key(args.metadata_dir / "metadata_elements_print.csv")
    _, shelfmarks = read_rows(args.metadata_dir / "shelfmarks.csv")

    years: dict[str, int] = {}
    for key, item in items.items():
        value = item.get("year", "").strip()
        if value.isdigit():
            years[key] = int(value)
    for row in classification_rows:
        value = row.get("Year", "").strip()
        if value.isdigit():
            years.setdefault(row["edition_key"].strip(), int(value))
    family_by_item, inferred_reprints = reprint_families(args.metadata_dir, years)

    enrichment_fields = [
        "corpus",
        "short_title",
        "author_or_editor",
        "publisher",
        "volumes",
        "ustc_id",
        "elements_books",
        *BOOK_GROUP_COLUMNS.keys(),
        "additional_content",
        "Reprint of (inferred from clusters)",
        "Reprint cluster comparison",
    ]
    output_fields = ["edition_key", "corpus"] + [
        field
        for field in input_fields
        if field not in {"edition_key", "corpus"} | EXCLUDED_OUTPUT_COLUMNS
    ]
    output_fields.extend(field for field in enrichment_fields if field not in output_fields)

    output_rows: list[dict[str, str]] = []
    retained_keys: set[str] = set()
    removed = 0
    for source_row in classification_rows:
        if source_row.get(CLASSIFICATION_COLUMN, "").strip().casefold() == "unclassified":
            removed += 1
            continue

        row = dict(source_row)
        key = row["edition_key"].strip()
        retained_keys.add(key)
        item = items.get(key, {})
        element = elements.get(key, {})
        books, _, normalized_books = parse_elements_books(element.get("elements_books", ""))

        row["corpus"] = corpus_for(row)
        for field in ("short_title", "author_or_editor", "publisher", "volumes", "ustc_id"):
            row[field] = item.get(field, "")
        row["elements_books"] = normalized_books
        row.update(book_group_flags(books))
        row["additional_content"] = element.get("additional_content", "")
        inferred_reprint = inferred_reprints.get(key, "")
        row["Reprint of (inferred from clusters)"] = inferred_reprint
        row["Reprint cluster comparison"] = compare_reprint(
            key, row.get("Reprint of", ""), inferred_reprint, family_by_item
        )
        output_rows.append(row)

    normalized_output_fields = [
        ENRICHED_OUTPUT_COLUMN_NAMES.get(field, field) for field in output_fields
    ]
    if len(normalized_output_fields) != len(set(normalized_output_fields)):
        raise ValueError("Normalized enriched output column names are not unique")
    normalized_output_rows = [
        {
            ENRICHED_OUTPUT_COLUMN_NAMES.get(field, field): row.get(field, "")
            for field in output_fields
        }
        for row in output_rows
    ]
    write_rows(args.output, normalized_output_fields, normalized_output_rows)

    scan_rows = [
        {
            "edition_key": row.get("key", ""),
            "volume": row.get("volume", ""),
            "scan": row.get("scan", ""),
        }
        for row in shelfmarks
        if row.get("key", "").strip() in retained_keys
    ]
    write_rows(args.scans_output, ["edition_key", "volume", "scan"], scan_rows)

    crops = crop_rows(
        args.crops_json_dir, (row["edition_key"].strip() for row in output_rows)
    )
    write_rows(
        args.crops_output,
        [
            "edition_key",
            "page_in_facsimile",
            "sequence_on_page",
            "automatically_identified_type",
        ],
        crops,
    )
    return len(output_rows), removed, len(scan_rows), len(crops)


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--input", type=Path, default=DEFAULT_INPUT)
    parser.add_argument("--metadata-dir", type=Path, default=DEFAULT_METADATA_DIR)
    parser.add_argument("--output", type=Path, default=DEFAULT_OUTPUT)
    parser.add_argument("--scans-output", type=Path, default=DEFAULT_SCANS_OUTPUT)
    parser.add_argument("--crops-json-dir", type=Path, default=DEFAULT_CROPS_JSON_DIR)
    parser.add_argument("--crops-output", type=Path, default=DEFAULT_CROPS_OUTPUT)
    return parser


def main() -> None:
    args = build_parser().parse_args()
    retained, removed, scans, crops = enrich(args)
    print(f"Wrote {retained} enriched rows to {args.output}")
    print(f"Removed {removed} unclassified rows")
    print(f"Wrote {scans} volume/scan rows to {args.scans_output}")
    print(f"Wrote {crops} crop rows to {args.crops_output}")


if __name__ == "__main__":
    main()
