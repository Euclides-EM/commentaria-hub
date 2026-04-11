#!/usr/bin/env python3

import csv
import json
import os
import random
import re
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
PAGE_SELECTION_EDGE_EXCLUSION_RATIO = 0.05
DATASET_READY_WAIT_INTERVAL_SECONDS = 10
DATASET_READY_TIMEOUT_SECONDS = 15 * 60
AUTH_TOKEN_ENV_VAR = "GITHUB_TOKEN"


def main() -> int:
    require_auth_token()
    page_counts_by_edition_id = load_page_counts()
    existing_base_pages_by_dataset_name = load_existing_base_pages_by_dataset_name()
    print_shell_method()
    task_count = collect_tasks(page_counts_by_edition_id, existing_base_pages_by_dataset_name)
    if not task_count:
        print("No matching facsimiles found for dh corpuses.")
        return 0
    print(f"Collected {task_count} task(s).")
    return 0


def collect_tasks(
        page_counts_by_edition_id: Dict[str, int],
        existing_base_pages_by_dataset_name: Dict[str, str],
) -> int:
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

    tasks_to_print: List[Dict[str, str]] = []
    for key in csv_keys:
        matches = [
            facsimiles_by_edition_id[edition_id]
            for edition_id in sorted_matching_edition_ids(key, facsimiles_by_edition_id.keys())
        ]
        if not matches:
            print(f"Warning: no facsimile found for key {key}", file=sys.stderr)
            continue
        for task in matches:
            tasks_to_print.append(task)

    for task in sorted(tasks_to_print, key=task_dataset_name):
        print_task_command(task, page_counts_by_edition_id, existing_base_pages_by_dataset_name)

    return len(tasks_to_print)


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

    edge_exclusion_count = int(total_pages * PAGE_SELECTION_EDGE_EXCLUSION_RATIO)
    start_page = 1 + edge_exclusion_count
    end_page = total_pages - edge_exclusion_count
    selectable_pages = list(range(start_page, end_page + 1))
    if len(selectable_pages) < SLICE_PAGE_COUNT:
        selectable_pages = list(range(1, total_pages + 1))

    selected_pages = sorted(random.sample(selectable_pages, SLICE_PAGE_COUNT))
    return ",".join(str(page) for page in selected_pages)


def print_shell_method() -> None:
    print("create_dataset() {")
    print("  local name=\"$1\"")
    print("  local facsimile_id=\"$2\"")
    print("  local pages=\"$3\"")
    print("  local response")
    print("  local dataset_id")
    print("  local started_at")
    print("  local now")
    print("  local status")
    print("  local creation_error")
    print("  echo \"Creating dataset ${name} with facsimile ${facsimile_id} and pages ${pages}\"")
    print("  response=$(curl -fsS -X POST \\")
    print("    -H 'Accept: application/json' \\")
    print("    -H 'Content-Type: application/json' \\")
    print(f'    -H "Authorization: Bearer ${AUTH_TOKEN_ENV_VAR}" \\')
    print(f"    '{BASE_URL}/api/v1/datasets?create_default_annotation=true&async=true' \\")
    print("    -d @- <<EOF")
    print("{")
    print('  "name": "${name}",')
    print(f'  "dpi": {DEFAULT_DPI},')
    print('  "facsimile_id": "${facsimile_id}",')
    print('  "deskewed": true,')
    print('  "denoised": true,')
    print('  "pages": "${pages}"')
    print("}")
    print("EOF")
    print("  ) || exit 1")
    print("  dataset_id=$(printf '%s' \"$response\" | python3 -c 'import json, sys; print(json.load(sys.stdin).get(\"id\", \"\"))') || exit 1")
    print("  if [ -z \"$dataset_id\" ]; then")
    print("    echo \"Failed to extract dataset id from create response\" >&2")
    print("    exit 1")
    print("  fi")
    print("  started_at=$(date +%s)")
    print("  while true; do")
    print("    response=$(curl -fsS \\")
    print("      -H 'Accept: application/json' \\")
    print(f'      -H "Authorization: Bearer ${AUTH_TOKEN_ENV_VAR}" \\')
    print(f"      '{BASE_URL}/api/v1/datasets') || exit 1")
    print("    status=$(printf '%s' \"$response\" | python3 -c 'import json, sys; dataset_id = sys.argv[1]; datasets = json.load(sys.stdin); match = next((dataset for dataset in datasets if dataset.get(\"id\") == dataset_id), None); print((match or {}).get(\"status\", \"\"))' \"$dataset_id\") || exit 1")
    print("    if [ \"$status\" = \"ready\" ]; then")
    print("      break")
    print("    fi")
    print("    if [ \"$status\" = \"failed\" ]; then")
    print("      creation_error=$(printf '%s' \"$response\" | python3 -c 'import json, sys; dataset_id = sys.argv[1]; datasets = json.load(sys.stdin); match = next((dataset for dataset in datasets if dataset.get(\"id\") == dataset_id), None); print((match or {}).get(\"creation_error\", \"\"))' \"$dataset_id\") || exit 1")
    print("      echo \"Dataset creation failed for ${name}: ${creation_error}\" >&2")
    print("      exit 1")
    print("    fi")
    print("    now=$(date +%s)")
    print(f"    if [ $((now - started_at)) -ge {DATASET_READY_TIMEOUT_SECONDS} ]; then")
    print(f"      echo \"Timed out waiting for dataset ${{name}} to become ready after {DATASET_READY_TIMEOUT_SECONDS} seconds\" >&2")
    print("      exit 1")
    print("    fi")
    print(f"    sleep {DATASET_READY_WAIT_INTERVAL_SECONDS}")
    print("  done")
    print("}")
    print()


def print_task_command(
        task: Dict[str, str],
        page_counts_by_edition_id: Dict[str, int],
        existing_base_pages_by_dataset_name: Dict[str, str],
) -> None:
    dataset_name_raw = task_dataset_name(task)
    pages = existing_base_pages_by_dataset_name.get(dataset_name_raw)
    if not pages:
        pages = choose_pages(task["edition_id"], page_counts_by_edition_id)
    dataset_name = shell_double_quote(dataset_name_raw)
    facsimile_id = shell_double_quote(task["facsimile_id"])
    pages_arg = shell_double_quote(pages)
    print(f"create_dataset \"{dataset_name}\" \"{facsimile_id}\" \"{pages_arg}\"")


def task_dataset_name(task: Dict[str, str]) -> str:
    return f'DH_Sliced_{task["edition_id"]}'


def load_existing_base_pages_by_dataset_name() -> Dict[str, str]:
    datasets = fetch_json("GET", "/api/v1/datasets")
    if not isinstance(datasets, list):
        fail("Expected datasets response to be a JSON array.")

    base_pages_by_dataset_name: Dict[str, str] = {}
    for dataset in datasets:
        if not isinstance(dataset, dict):
            fail("Expected each dataset to be a JSON object.")
        dataset_id = (dataset.get("id") or "").strip()
        dataset_name = (dataset.get("name") or "").strip()
        if not dataset_id or not dataset_name or dataset_name in base_pages_by_dataset_name:
            continue
        pages = load_base_annotation_pages(dataset_id)
        if pages:
            base_pages_by_dataset_name[dataset_name] = pages
    return base_pages_by_dataset_name


def load_base_annotation_pages(dataset_id: str) -> Optional[str]:
    annotations = fetch_json("GET", f"/api/v1/datasets/{dataset_id}/annotations")
    if not isinstance(annotations, list):
        return None

    for annotation in annotations:
        if not isinstance(annotation, dict):
            fail(f"Expected each annotation for dataset {dataset_id!r} to be a JSON object.")
        if (annotation.get("name") or "").strip() != "Base":
            continue
        pages = (annotation.get("pages") or "").strip()
        if pages:
            return pages
        return None
    return None


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


def shell_double_quote(value: str) -> str:
    return value.replace("\\", "\\\\").replace('"', '\\"')


def require_auth_token() -> str:
    token = os.environ.get(AUTH_TOKEN_ENV_VAR, "").strip()
    if not token:
        fail(f"Environment variable {AUTH_TOKEN_ENV_VAR} is required.")
    return token


def fail(message: str) -> None:
    raise SystemExit(f"Error: {message}")


if __name__ == "__main__":
    raise SystemExit(main())
