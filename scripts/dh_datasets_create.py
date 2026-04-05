#!/usr/bin/env python3

import csv
import concurrent.futures
import json
import os
import random
import re
import sys
import threading
import urllib.error
import urllib.parse
import urllib.request
from pathlib import Path
from typing import Any, Dict, Iterable, List, Optional, Set, Tuple


BASE_URL = "https://euclides.huma-num.fr/commentaria"
PAGE_COUNTS_CSV_URL = "https://raw.githubusercontent.com/Euclides-EM/elements-facsimile/refs/heads/main/page_counts.csv"
REPO_ROOT = Path(__file__).resolve().parents[1]
CORPUSES_CSV_PATH = REPO_ROOT / "ocrflow" / "store" / "items_metadata" / "corpuses.csv"
RESUME_STATE_PATH = Path(__file__).resolve().parent / "dh_datasets_create.state.json"
DRY_RUN = False
DEFAULT_CONCURRENCY = 2
HTTP_TIMEOUT_SECONDS = 60
DATASET_CREATE_TIMEOUT_SECONDS = 60 * 60
DEFAULT_DPI = 300
SLICE_PAGE_COUNT = 50
AUTH_TOKEN_ENV_VAR = "GITHUB_TOKEN"


def main() -> int:
    require_auth_token()
    page_counts_by_edition_id = load_page_counts()
    tasks = collect_tasks()
    if not tasks:
        print("No matching facsimiles found for dh corpuses.")
        return 0

    pending_tasks = get_pending_tasks(tasks)
    if not pending_tasks:
        print(f"Collected {len(tasks)} task(s). All tasks are already completed.")
        return 0

    print(
        f"Collected {len(tasks)} task(s). Running {len(pending_tasks)} pending task(s) with concurrency {DEFAULT_CONCURRENCY}."
    )

    if DRY_RUN:
        failed_tasks = 0
        for task_number, total_tasks, task in pending_tasks:
            if not process_task(task_number, total_tasks, task, page_counts_by_edition_id, None):
                failed_tasks += 1
        return report_run_outcome(len(pending_tasks), failed_tasks)

    state_lock = threading.Lock()
    max_workers = min(DEFAULT_CONCURRENCY, len(pending_tasks))
    failed_tasks = 0
    with concurrent.futures.ThreadPoolExecutor(max_workers=max_workers) as executor:
        futures = [
            executor.submit(process_task, task_number, total_tasks, task, page_counts_by_edition_id, state_lock)
            for task_number, total_tasks, task in pending_tasks
        ]
        for future in concurrent.futures.as_completed(futures):
            if not future.result():
                failed_tasks += 1

    return report_run_outcome(len(pending_tasks), failed_tasks)


def collect_tasks() -> List[Dict[str, str]]:
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

    tasks: List[Dict[str, str]] = []
    for key in csv_keys:
        matches = [
            facsimiles_by_edition_id[edition_id]
            for edition_id in sorted_matching_edition_ids(key, facsimiles_by_edition_id.keys())
        ]
        if not matches:
            print(f"Warning: no facsimile found for key {key}", file=sys.stderr)
            continue
        tasks.extend(matches)

    return tasks


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


def get_pending_tasks(tasks: List[Dict[str, str]]) -> List[Tuple[int, int, Dict[str, str]]]:
    completed_facsimile_ids = load_completed_facsimile_ids()
    pending_tasks: List[Tuple[int, int, Dict[str, str]]] = []
    total_tasks = len(tasks)
    for index, task in enumerate(tasks):
        if task["facsimile_id"] in completed_facsimile_ids:
            continue
        pending_tasks.append((index + 1, total_tasks, task))
    return pending_tasks


def load_completed_facsimile_ids() -> Set[str]:
    state = load_resume_state()
    completed_facsimile_ids = set(state.keys())
    if completed_facsimile_ids:
        print(f"Resume state found with {len(completed_facsimile_ids)} completed facsimile(s).")
    return completed_facsimile_ids


def load_resume_state() -> Dict[str, Any]:
    if not RESUME_STATE_PATH.exists():
        return {}
    try:
        state = json.loads(RESUME_STATE_PATH.read_text(encoding="utf-8"))
    except json.JSONDecodeError as exc:
        fail(f"Failed to parse resume state file {RESUME_STATE_PATH}: {exc}")
    if not isinstance(state, dict):
        fail(f"Resume state file {RESUME_STATE_PATH} must contain a JSON object.")
    for facsimile_id, entry in state.items():
        if not isinstance(entry, dict):
            fail(
                f"Resume state entry for facsimile {facsimile_id!r} must be an object with dataset_id and pages."
            )
        for field_name in ("dataset_id", "pages"):
            if field_name not in entry:
                fail(
                    f"Resume state entry for facsimile {facsimile_id!r} is missing required field {field_name!r}."
                )
    return state


def save_resume_state(
    facsimile_id: str,
    dataset_id: str,
    pages: str,
) -> None:
    state = load_resume_state()
    state[facsimile_id] = {
        "dataset_id": dataset_id,
        "pages": pages,
    }
    RESUME_STATE_PATH.write_text(json.dumps(state, indent=2) + "\n", encoding="utf-8")


def process_task(
    task_number: int,
    total_tasks: int,
    task: Dict[str, str],
    page_counts_by_edition_id: Dict[str, int],
    state_lock: Optional[Any],
) -> bool:
    facsimile_id = task["facsimile_id"]
    edition_id = task["edition_id"]
    print(f"[{task_number}/{total_tasks}] Processing {edition_id} ({facsimile_id})")
    try:
        if DRY_RUN:
            dry_run_task(task, edition_id, page_counts_by_edition_id)
            return True

        dataset = create_dataset(task, page_counts_by_edition_id)
        dataset_id = require_field(dataset, "id", "dataset create response")
        dataset_pages = str(dataset.get("pages") or "")
        print(f"[{task_number}/{total_tasks}] Created dataset_id={dataset_id} pages={dataset_pages!r}")

        if state_lock is None:
            save_resume_state(facsimile_id, dataset_id, dataset_pages)
        else:
            with state_lock:
                save_resume_state(facsimile_id, dataset_id, dataset_pages)
        print(f"[{task_number}/{total_tasks}] Completed {edition_id} ({facsimile_id})")
        return True
    except (Exception, SystemExit) as exc:
        error_message = str(exc)
        print(
            f"[{task_number}/{total_tasks}] Error for {edition_id} ({facsimile_id}): {error_message}",
            file=sys.stderr,
        )
        return False


def report_run_outcome(total_pending_tasks: int, failed_tasks: int) -> int:
    if failed_tasks:
        print(
            f"Finished processing {total_pending_tasks} pending task(s) with {failed_tasks} failure(s).",
            file=sys.stderr,
        )
        return 1
    print("All tasks completed.")
    return 0


def choose_pages(edition_id: str, page_counts_by_edition_id: Dict[str, int]) -> str:
    total_pages = page_counts_by_edition_id.get(edition_id)
    if total_pages is None:
        fail(f"Missing page count for edition_id {edition_id!r} in {PAGE_COUNTS_CSV_URL}.")
    if total_pages <= SLICE_PAGE_COUNT:
        return f"1-{total_pages}"
    start_page = random.randint(1, total_pages - SLICE_PAGE_COUNT + 1)
    end_page = start_page + SLICE_PAGE_COUNT - 1
    return f"{start_page}-{end_page}"


def create_dataset(task: Dict[str, str], page_counts_by_edition_id: Dict[str, int]) -> Dict[str, Any]:
    pages = choose_pages(task["edition_id"], page_counts_by_edition_id)
    payload = {
        "name": f'DH_Sliced_{task["edition_id"]}',
        "dpi": DEFAULT_DPI,
        "facsimile_id": task["facsimile_id"],
        "deskewed": True,
        "denoised": True,
        "pages": pages,
    }
    query = urllib.parse.urlencode({"create_default_annotation": "true"})
    response = fetch_json("POST", f"/api/v1/datasets?{query}", payload, timeout_seconds=DATASET_CREATE_TIMEOUT_SECONDS)
    if not isinstance(response, dict):
        fail("Expected dataset create response to be a JSON object.")
    return response


def dry_run_task(task: Dict[str, str], edition_id: str, page_counts_by_edition_id: Dict[str, int]) -> None:
    pages = choose_pages(edition_id, page_counts_by_edition_id)
    dataset_payload = {
        "name": f'DH_Sliced_{task["edition_id"]}',
        "dpi": DEFAULT_DPI,
        "facsimile_id": task["facsimile_id"],
        "deskewed": True,
        "denoised": True,
        "pages": pages,
    }
    print(
        "DRY RUN POST "
        f"{BASE_URL}/api/v1/datasets?create_default_annotation=true "
        f"payload={json.dumps(dataset_payload, sort_keys=True)}"
    )


def fetch_json(
    method: str,
    path: str,
    payload: Optional[Dict[str, Any]] = None,
    timeout_seconds: int = HTTP_TIMEOUT_SECONDS,
) -> Any:
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


def require_field(data: Dict[str, Any], field_name: str, context: str) -> str:
    value = data.get(field_name)
    if not value:
        fail(f"Missing field {field_name!r} in {context}.")
    return value


def require_auth_token() -> str:
    token = os.environ.get(AUTH_TOKEN_ENV_VAR, "").strip()
    if not token:
        fail(f"Environment variable {AUTH_TOKEN_ENV_VAR} is required.")
    return token


def fail(message: str) -> None:
    raise SystemExit(f"Error: {message}")


if __name__ == "__main__":
    raise SystemExit(main())
