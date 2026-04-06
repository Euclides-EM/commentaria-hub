#!/usr/bin/env python3

import csv
import json
import os
import random
import re
import shlex
import sys
import urllib.error
import urllib.request
from pathlib import Path
from typing import Any, Dict, Iterable, List, Optional, Tuple

BASE_URL = "https://euclides.huma-num.fr/commentaria"
PAGE_COUNTS_CSV_URL = "https://raw.githubusercontent.com/Euclides-EM/elements-facsimile/refs/heads/main/page_counts.csv"
REPO_ROOT = Path(__file__).resolve().parents[1]
CORPUSES_CSV_PATH = REPO_ROOT / "ocrflow" / "store" / "items_metadata" / "corpuses.csv"
HTTP_TIMEOUT_SECONDS = 60
DEFAULT_DPI = 300
SLICE_PAGE_COUNT = 50
AUTH_TOKEN_ENV_VAR = "GITHUB_TOKEN"


def main() -> int:
    require_auth_token()
    page_counts_by_edition_id = load_page_counts()
    print_shell_method()
    task_count = collect_tasks(page_counts_by_edition_id)
    if not task_count:
        print("No matching facsimiles found for dh corpuses.")
        return 0
    print(f"Collected {task_count} task(s).")
    return 0


def collect_tasks(page_counts_by_edition_id: Dict[str, int]) -> int:
    csv_keys = load_dh_keys()
    facsimiles = fetch_json("GET", "/api/v1/facsimilies")
    if not isinstance(facsimiles, list):
        fail("Expected facsimiles response to be a JSON array.")

    facsimiles_by_edition_id: Dict[str, Dict[str, str]] = {}
    for facsimile in facsimiles:
        if not isinstance(facsimile, dict):
            fail("Expected each facsimile to be a JSON object.")
        edition_id = facsimile.get("edition_id")
        facsimile_id = facsimile.get("id")
        if not edition_id or not facsimile_id:
            continue
        facsimiles_by_edition_id[edition_id] = {
            "edition_id": edition_id,
            "facsimile_id": facsimile_id,
        }

    task_count = 0
    for key in csv_keys:
        matches = [
            facsimiles_by_edition_id[edition_id]
            for edition_id in sorted_matching_edition_ids(key, facsimiles_by_edition_id.keys())
        ]
        if not matches:
            print(f"Warning: no facsimile found for key {key}", file=sys.stderr)
            continue
        for task in matches:
            print_task_command(task, page_counts_by_edition_id)
            task_count += 1

    return task_count


def load_dh_keys() -> List[str]:
    keys: List[str] = []
    with CORPUSES_CSV_PATH.open(newline="", encoding="utf-8") as handle:
        reader = csv.DictReader(handle)
        for row in reader:
            study = (row.get("study") or "").strip()
            if has_dh_study(study):
                key = (row.get("key") or "").strip()
                if key:
                    keys.append(key)
    return keys


def load_page_counts() -> Dict[str, int]:
    request = urllib.request.Request(PAGE_COUNTS_CSV_URL, headers={"Accept": "text/csv"})
    try:
        with urllib.request.urlopen(request, timeout=HTTP_TIMEOUT_SECONDS) as response:
            charset = response.headers.get_content_charset("utf-8")
            raw_csv = response.read().decode(charset)
    except urllib.error.HTTPError as exc:
        error_body = exc.read().decode("utf-8", errors="replace")
        fail(f"GET {PAGE_COUNTS_CSV_URL} failed with HTTP {exc.code}: {error_body}")
    except urllib.error.URLError as exc:
        fail(f"GET {PAGE_COUNTS_CSV_URL} failed: {exc}")

    page_counts_by_edition_id: Dict[str, int] = {}
    reader = csv.DictReader(raw_csv.splitlines())
    if not reader.fieldnames:
        fail(f"{PAGE_COUNTS_CSV_URL} returned an empty CSV.")
    if "filename" not in reader.fieldnames or "pages" not in reader.fieldnames:
        fail(f"{PAGE_COUNTS_CSV_URL} must contain filename and pages columns.")

    for row in reader:
        filename = (row.get("filename") or "").strip()
        pages_raw = (row.get("pages") or "").strip()
        if not filename or not pages_raw:
            continue
        if not filename.endswith(".pdf"):
            fail(f"Unexpected filename in {PAGE_COUNTS_CSV_URL}: {filename!r}")
        try:
            pages = int(pages_raw)
        except ValueError:
            fail(f"Invalid pages value in {PAGE_COUNTS_CSV_URL} for {filename!r}: {pages_raw!r}")
        if pages < 1:
            fail(f"Invalid page count in {PAGE_COUNTS_CSV_URL} for {filename!r}: {pages}")
        page_counts_by_edition_id[filename[:-4]] = pages

    if not page_counts_by_edition_id:
        fail(f"{PAGE_COUNTS_CSV_URL} did not contain any page counts.")
    return page_counts_by_edition_id


def has_dh_study(study: str) -> bool:
    return "dh" in {part.strip().lower() for part in study.split(",") if part.strip()}


def sorted_matching_edition_ids(key: str, edition_ids: Iterable[str]) -> List[str]:
    pattern = re.compile(rf"^{re.escape(key)}(?:_vol(\d+))?$")
    matches: List[Tuple[int, str]] = []
    for edition_id in edition_ids:
        match = pattern.fullmatch(edition_id)
        if not match:
            continue
        volume = match.group(1)
        sort_index = 0 if volume is None else int(volume)
        matches.append((sort_index, edition_id))
    matches.sort(key=lambda item: (item[0], item[1]))
    return [edition_id for _, edition_id in matches]


def choose_pages(edition_id: str, page_counts_by_edition_id: Dict[str, int]) -> str:
    total_pages = page_counts_by_edition_id.get(edition_id)
    if total_pages is None:
        fail(f"Missing page count for edition_id {edition_id!r} in {PAGE_COUNTS_CSV_URL}.")
    if total_pages <= SLICE_PAGE_COUNT:
        return f"1-{total_pages}"
    selected_pages = sorted(random.sample(range(1, total_pages + 1), SLICE_PAGE_COUNT))
    return ",".join(str(page) for page in selected_pages)


def print_shell_method() -> None:
    print("create_dataset() {")
    print("  local name=\"$1\"")
    print("  local facsimile_id=\"$2\"")
    print("  local pages=\"$3\"")
    print("  curl -X POST \\")
    print("    -H 'Accept: application/json' \\")
    print("    -H 'Content-Type: application/json' \\")
    print(f'    -H "Authorization: Bearer ${AUTH_TOKEN_ENV_VAR}" \\')
    print(f"    '{BASE_URL}/api/v1/datasets?create_default_annotation=true' \\")
    print("    -d @- <<EOF")
    print("{")
    print('  "name": "'"'"'${name}'"'"'",')
    print(f'  "dpi": {DEFAULT_DPI},')
    print('  "facsimile_id": "'"'"'${facsimile_id}'"'"'",')
    print('  "deskewed": true,')
    print('  "denoised": true,')
    print('  "pages": "'"'"'${pages}'"'"'"')
    print("}")
    print("EOF")
    print("}")
    print()


def print_task_command(task: Dict[str, str], page_counts_by_edition_id: Dict[str, int]) -> None:
    pages = choose_pages(task["edition_id"], page_counts_by_edition_id)
    dataset_name = shell_quote(f'DH_Sliced_{task["edition_id"]}')
    facsimile_id = shell_quote(task["facsimile_id"])
    pages_arg = shell_quote(pages)
    print(f"create_dataset \"{dataset_name}\" \"{facsimile_id}\" \"{pages_arg}\"")


def fetch_json(
        method: str,
        path: str,
        payload: Optional[Dict[str, Any]] = None,
        timeout_seconds: int = HTTP_TIMEOUT_SECONDS,
) -> object:
    url = f"{BASE_URL}{path}"
    body = None
    headers = {
        "Accept": "application/json",
        "Authorization": f"Bearer {require_auth_token()}",
    }

    if payload is not None:
        body = json.dumps(payload).encode("utf-8")
        headers["Content-Type"] = "application/json"

    request = urllib.request.Request(url, data=body, headers=headers, method=method)
    try:
        with urllib.request.urlopen(request, timeout=timeout_seconds) as response:
            charset = response.headers.get_content_charset("utf-8")
            raw_body = response.read().decode(charset)
    except urllib.error.HTTPError as exc:
        error_body = exc.read().decode("utf-8", errors="replace")
        fail(f"{method} {url} failed with HTTP {exc.code}: {error_body}")
    except urllib.error.URLError as exc:
        fail(f"{method} {url} failed: {exc}")

    if not raw_body.strip():
        return None

    try:
        return json.loads(raw_body)
    except json.JSONDecodeError as exc:
        fail(f"{method} {url} returned invalid JSON: {exc}")


def shell_quote(value: str) -> str:
    return shlex.quote(value)


def require_auth_token() -> str:
    token = os.environ.get(AUTH_TOKEN_ENV_VAR, "").strip()
    if not token:
        fail(f"Environment variable {AUTH_TOKEN_ENV_VAR} is required.")
    return token


def fail(message: str) -> None:
    raise SystemExit(f"Error: {message}")


if __name__ == "__main__":
    raise SystemExit(main())
