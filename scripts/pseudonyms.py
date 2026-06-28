#!/usr/bin/env python3
from __future__ import annotations

import argparse
import csv
import re
import zipfile
from dataclasses import dataclass
from pathlib import Path
from typing import Iterable
from xml.etree import ElementTree


XML_NS = {"main": "http://schemas.openxmlformats.org/spreadsheetml/2006/main"}
TARGET_COLUMNS = {
    "names unsorted": ("other", True),
    "last name": ("last", False),
    "first name": ("first", False),
    "location name": ("other", False),
}


@dataclass(frozen=True)
class MappingRow:
    name: str
    pseudonym: str
    position: str
    source: str


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Create a pseudonyms mapping CSV from every sheet in an XLSX workbook."
    )
    parser.add_argument("xlsx_path", help="Path to the input XLSX file.")
    parser.add_argument(
        "-o",
        "--output",
        help="Path to the output CSV. Defaults to pseudonyms.csv in the current working directory.",
    )
    return parser.parse_args()


def normalize_header(value: str) -> str:
    return " ".join(value.replace("\xa0", " ").split()).strip().lower()


def sanitize_value(value: str) -> str:
    without_ellipsis = value.replace("...", "")
    return re.sub(r"\([^()]*\)", "", without_ellipsis)


def normalize_value(value: str) -> str:
    return " ".join(sanitize_value(value).replace("\xa0", " ").split()).strip()


def normalize_cell_text(value: str) -> str:
    normalized = sanitize_value(value).replace("\xa0", " ")
    normalized = normalized.replace("\r\n", "\n").replace("\r", "\n")
    return "\n".join(" ".join(part.split()).strip() for part in normalized.split("\n"))


def is_skippable_value(value: str) -> bool:
    normalized = normalize_value(value)
    return normalized == "" or normalized == "-"


def split_multiline_values(value: str) -> list[str]:
    return [part.strip() for part in value.replace("\r\n", "\n").replace("\r", "\n").split("\n")]


def column_index_from_ref(cell_ref: str) -> int:
    letters = "".join(char for char in cell_ref if char.isalpha())
    index = 0
    for char in letters:
        index = index * 26 + (ord(char.upper()) - ord("A") + 1)
    return index - 1


def split_comma_name(value: str) -> list[tuple[str, str]]:
    parts = [part.strip() for part in value.split(",")]
    if len(parts) < 2:
        return []
    last = parts[0]
    first = parts[1]
    if not last or not first:
        return []
    return [(last, "last"), (first, "first")]


def split_unsorted_name(value: str) -> list[tuple[str, str]]:
    parts = [part.strip() for part in value.split(",")]
    if len(parts) < 2:
        return []
    last = parts[0]
    first = parts[1]
    if not last or not first:
        return []
    return [(last, "last"), (first, "first")]


def split_name_parts(name: str) -> tuple[str, str]:
    parts = name.rsplit(" ", 1)
    if len(parts) != 2:
        return ("", name)
    return (parts[0], parts[1])


def first_name_initials(name: str) -> list[str]:
    first_part, _ = split_name_parts(name)
    if not first_part:
        return []

    initials: list[str] = []
    for token in first_part.split():
        cleaned = token.strip(".")
        if not cleaned:
            continue
        initial = f"{cleaned[0]}."
        if initial not in initials:
            initials.append(initial)
    if len(initials) > 1:
        combined = "".join(initials)
        if combined not in initials:
            initials.append(combined)
    return initials


def load_shared_strings(archive: zipfile.ZipFile) -> list[str]:
    try:
        data = archive.read("xl/sharedStrings.xml")
    except KeyError:
        return []

    root = ElementTree.fromstring(data)
    strings: list[str] = []
    for item in root.findall("main:si", XML_NS):
        text = "".join(item.itertext())
        strings.append(normalize_cell_text(text))
    return strings


def load_sheet_names(archive: zipfile.ZipFile) -> list[tuple[str, str]]:
    workbook_root = ElementTree.fromstring(archive.read("xl/workbook.xml"))
    rels_root = ElementTree.fromstring(archive.read("xl/_rels/workbook.xml.rels"))
    relationships = {
        rel.attrib["Id"]: rel.attrib["Target"]
        for rel in rels_root
        if rel.tag.endswith("Relationship")
    }

    sheets: list[tuple[str, str]] = []
    for sheet in workbook_root.findall("main:sheets/main:sheet", XML_NS):
        rel_id = sheet.attrib.get("{http://schemas.openxmlformats.org/officeDocument/2006/relationships}id")
        if not rel_id or rel_id not in relationships:
            continue
        target = relationships[rel_id]
        if not target.startswith("xl/"):
            target = f"xl/{target.lstrip('/')}"
        sheets.append((sheet.attrib.get("name", "Sheet"), target))
    return sheets


def cell_value(cell: ElementTree.Element, shared_strings: list[str]) -> str:
    cell_type = cell.attrib.get("t")
    if cell_type == "inlineStr":
        return normalize_cell_text("".join(cell.itertext()))

    value_node = cell.find("main:v", XML_NS)
    if value_node is None or value_node.text is None:
        return ""

    raw_value = value_node.text
    if cell_type == "s":
        try:
            return shared_strings[int(raw_value)]
        except (ValueError, IndexError):
            return ""
    return normalize_cell_text(raw_value)


def load_sheet_rows(
    archive: zipfile.ZipFile, sheet_path: str, shared_strings: list[str]
) -> list[list[str]]:
    sheet_root = ElementTree.fromstring(archive.read(sheet_path))
    rows: list[list[str]] = []
    for row in sheet_root.findall("main:sheetData/main:row", XML_NS):
        values: dict[int, str] = {}
        max_index = -1
        for cell in row.findall("main:c", XML_NS):
            ref = cell.attrib.get("r", "")
            if not ref:
                continue
            index = column_index_from_ref(ref)
            values[index] = cell_value(cell, shared_strings)
            if index > max_index:
                max_index = index
        if max_index < 0:
            rows.append([])
            continue
        rows.append([values.get(i, "") for i in range(max_index + 1)])
    return rows


def iter_mapping_rows(
    name: str, value: str, default_position: str, split_unsorted: bool, source: str
) -> Iterable[MappingRow]:
    normalized_name = normalize_value(name)
    if is_skippable_value(normalized_name):
        return

    for part in split_multiline_values(value):
        normalized_value = normalize_value(part)
        if is_skippable_value(normalized_value):
            continue

        split_rows = split_unsorted_name(normalized_value) if split_unsorted else split_comma_name(normalized_value)
        if split_rows:
            for pseudonym, position in split_rows:
                if normalized_name == pseudonym:
                    continue
                yield MappingRow(normalized_name, pseudonym, position, source)
                if position == "first":
                    for initial in first_name_initials(normalized_name):
                        if initial != pseudonym and normalized_name != initial:
                            yield MappingRow(normalized_name, initial, position, source)
            continue

        if normalized_name == normalized_value:
            continue
        yield MappingRow(normalized_name, normalized_value, default_position, source)
        if default_position == "first":
            for initial in first_name_initials(normalized_name):
                if initial != normalized_value and normalized_name != initial:
                    yield MappingRow(normalized_name, initial, default_position, source)


def collect_rows(xlsx_path: Path) -> list[MappingRow]:
    rows: list[MappingRow] = []
    seen: set[MappingRow] = set()

    with zipfile.ZipFile(xlsx_path) as archive:
        shared_strings = load_shared_strings(archive)
        for sheet_name, sheet_path in load_sheet_names(archive):
            sheet_rows = load_sheet_rows(archive, sheet_path, shared_strings)
            if not sheet_rows:
                continue

            headers = [normalize_header(value) for value in sheet_rows[0]]
            target_indexes = {
                index: TARGET_COLUMNS[header]
                for index, header in enumerate(headers)
                if header in TARGET_COLUMNS
            }

            for row in sheet_rows[1:]:
                if not row:
                    continue
                proper_name = normalize_value(row[0] if len(row) > 0 else "")
                if not proper_name:
                    continue
                for index, (default_position, split_unsorted) in target_indexes.items():
                    if index >= len(row):
                        continue
                    for mapping_row in iter_mapping_rows(
                        proper_name,
                        row[index],
                        default_position,
                        split_unsorted,
                        sheet_name,
                    ):
                        if mapping_row not in seen:
                            seen.add(mapping_row)
                            rows.append(mapping_row)

    return rows


def default_output_path() -> Path:
    return Path.cwd() / "../ocrflow/store/items_metadata/pseudonyms.csv"


def write_csv(output_path: Path, rows: list[MappingRow]) -> None:
    with output_path.open("w", newline="", encoding="utf-8") as handle:
        writer = csv.writer(handle)
        writer.writerow(["name", "pseudonym", "position", "source"])
        for row in rows:
            writer.writerow([row.name, row.pseudonym, row.position, row.source])


def main() -> int:
    args = parse_args()
    xlsx_path = Path(args.xlsx_path).expanduser().resolve()
    output_path = (
        Path(args.output).expanduser().resolve()
        if args.output
        else default_output_path().resolve()
    )

    rows = collect_rows(xlsx_path)
    write_csv(output_path, rows)
    print(f"Wrote {len(rows)} rows to {output_path}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
