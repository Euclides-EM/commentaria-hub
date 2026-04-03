#!/usr/bin/env python3

import csv
import concurrent.futures
import json
import os
import random
import re
import sys
import threading
import time
import urllib.error
import urllib.parse
import urllib.request
from pathlib import Path
from typing import Any, Dict, Iterable, List, Optional, Set, Tuple


BASE_URL = "http://localhost:8085"
REPO_ROOT = Path(__file__).resolve().parents[1]
CORPUSES_CSV_PATH = REPO_ROOT / "ocrflow" / "store" / "items_metadata" / "corpuses.csv"
RESUME_STATE_PATH = Path(__file__).resolve().parent / "dh_datasets_create.state.json"
DRY_RUN = False
DEFAULT_CONCURRENCY = 4
HTTP_TIMEOUT_SECONDS = 60
DATASET_CREATE_TIMEOUT_SECONDS = 60 * 60
ANNOTATION_WAIT_TIMEOUT_SECONDS = 60
ANNOTATION_WAIT_INTERVAL_SECONDS = 1
DEFAULT_DPI = 300
SLICED_ANNOTATION_NAME = "Sliced"
FALLBACK_SLICE_RANGE = "50-100"
SLICE_PAGE_COUNT = 50
AUTH_TOKEN_ENV_VAR = "GITHUB_TOKEN"


def main() -> int:
    require_auth_token()
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
        for task_number, total_tasks, task in pending_tasks:
            process_task(task_number, total_tasks, task, None)
        print("All tasks completed.")
        return 0

    state_lock = threading.Lock()
    max_workers = min(DEFAULT_CONCURRENCY, len(pending_tasks))
    with concurrent.futures.ThreadPoolExecutor(max_workers=max_workers) as executor:
        futures = [
            executor.submit(process_task, task_number, total_tasks, task, state_lock)
            for task_number, total_tasks, task in pending_tasks
        ]
        try:
            for future in concurrent.futures.as_completed(futures):
                future.result()
        except BaseException:
            for future in futures:
                future.cancel()
            raise

    print("All tasks completed.")
    return 0


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
    completed_facsimile_ids = load_completed_facsimile_ids(tasks)
    pending_tasks: List[Tuple[int, int, Dict[str, str]]] = []
    total_tasks = len(tasks)
    for index, task in enumerate(tasks):
        if task["facsimile_id"] in completed_facsimile_ids:
            continue
        pending_tasks.append((index + 1, total_tasks, task))
    return pending_tasks


def load_completed_facsimile_ids(tasks: List[Dict[str, str]]) -> Set[str]:
    state = load_resume_state()
    completed_facsimile_ids = {
        facsimile_id
        for facsimile_id in parse_completed_facsimile_ids(state)
        if facsimile_id
    }
    last_successful_facsimile_id = state.get("last_successful_facsimile_id")
    if last_successful_facsimile_id and not completed_facsimile_ids:
        for index, task in enumerate(tasks):
            completed_facsimile_ids.add(task["facsimile_id"])
            if task["facsimile_id"] == last_successful_facsimile_id:
                print(
                    f"Resume state found for facsimile {last_successful_facsimile_id}; skipping the first {index + 1} task(s)."
                )
                return completed_facsimile_ids
        print(
            f"Resume state facsimile {last_successful_facsimile_id} not found in current task list; starting from the beginning.",
            file=sys.stderr,
        )
        return set()

    if completed_facsimile_ids:
        print(f"Resume state found with {len(completed_facsimile_ids)} completed facsimile(s).")
    return completed_facsimile_ids


def load_resume_state() -> Dict[str, str]:
    if not RESUME_STATE_PATH.exists():
        return {}
    try:
        state = json.loads(RESUME_STATE_PATH.read_text(encoding="utf-8"))
    except json.JSONDecodeError as exc:
        fail(f"Failed to parse resume state file {RESUME_STATE_PATH}: {exc}")
    if not isinstance(state, dict):
        fail(f"Resume state file {RESUME_STATE_PATH} must contain a JSON object.")
    return {str(key): str(value) for key, value in state.items()}


def parse_completed_facsimile_ids(state: Dict[str, str]) -> List[str]:
    completed_raw = state.get("completed_facsimile_ids", "")
    if not completed_raw:
        return []
    return [facsimile_id for facsimile_id in completed_raw.split(",") if facsimile_id]


def save_resume_state(last_successful_facsimile_id: str) -> None:
    state = load_resume_state()
    completed_facsimile_ids = set(parse_completed_facsimile_ids(state))
    completed_facsimile_ids.add(last_successful_facsimile_id)
    state = {
        "last_successful_facsimile_id": last_successful_facsimile_id,
        "completed_facsimile_ids": ",".join(sorted(completed_facsimile_ids)),
    }
    RESUME_STATE_PATH.write_text(json.dumps(state, indent=2) + "\n", encoding="utf-8")


def process_task(
    task_number: int,
    total_tasks: int,
    task: Dict[str, str],
    state_lock: Optional[Any],
) -> None:
    facsimile_id = task["facsimile_id"]
    edition_id = task["edition_id"]
    print(f"[{task_number}/{total_tasks}] Processing {edition_id} ({facsimile_id})")
    if DRY_RUN:
        dry_run_task(task, edition_id)
        return

    dataset = create_dataset(task)
    dataset_id = require_field(dataset, "id", "dataset create response")
    annotation = wait_for_default_annotation(dataset_id)
    annotation_id = require_field(annotation, "id", f"default annotation for dataset {dataset_id}")
    annotation_pages = str(annotation.get("pages") or "")
    slice_range = choose_slice_range(annotation, edition_id)
    apply_sliced_annotation(dataset_id, annotation_id, slice_range)
    print(
        f"[{task_number}/{total_tasks}] Created dataset_id={dataset_id} annotation_id={annotation_id} "
        f"pages={annotation_pages!r} slice_range={slice_range}"
    )

    if state_lock is None:
        save_resume_state(facsimile_id)
    else:
        with state_lock:
            save_resume_state(facsimile_id)
    print(f"[{task_number}/{total_tasks}] Completed {edition_id} ({facsimile_id})")


def create_dataset(task: Dict[str, str]) -> Dict[str, Any]:
    payload = {
        "name": task["edition_id"],
        "dpi": DEFAULT_DPI,
        "facsimile_id": task["facsimile_id"],
        "deskewed": True,
        "denoised": True,
    }
    query = urllib.parse.urlencode({"create_default_annotation": "true"})
    response = fetch_json("POST", f"/api/v1/datasets?{query}", payload, timeout_seconds=DATASET_CREATE_TIMEOUT_SECONDS)
    if not isinstance(response, dict):
        fail("Expected dataset create response to be a JSON object.")
    return response


def dry_run_task(task: Dict[str, str], edition_id: str) -> None:
    print(
        f"Warning: DRY_RUN cannot inspect the default annotation for {edition_id}; using fallback {FALLBACK_SLICE_RANGE}",
        file=sys.stderr,
    )
    dataset_payload = {
        "name": task["edition_id"],
        "dpi": DEFAULT_DPI,
        "facsimile_id": task["facsimile_id"],
        "deskewed": True,
        "denoised": True,
    }
    print(
        "DRY RUN POST "
        f"{BASE_URL}/api/v1/datasets?create_default_annotation=true "
        f"payload={json.dumps(dataset_payload, sort_keys=True)}"
    )
    print(
        "DRY RUN PUT "
        f"{BASE_URL}/api/v1/datasets/$datasetId/annotations/$annotationId/apply "
        f"payload={json.dumps(build_sliced_rules_payload(FALLBACK_SLICE_RANGE), sort_keys=True)}"
    )


def wait_for_default_annotation(dataset_id: str) -> Dict[str, Any]:
    deadline = time.time() + ANNOTATION_WAIT_TIMEOUT_SECONDS
    last_seen_names: List[str] = []

    while time.time() < deadline:
        annotations = fetch_json("GET", f"/api/v1/datasets/{dataset_id}/annotations")
        if not isinstance(annotations, list):
            fail(f"Expected annotations response for dataset {dataset_id} to be a JSON array.")

        by_name = {}
        for annotation in annotations:
            if not isinstance(annotation, dict):
                fail(f"Expected annotation entries for dataset {dataset_id} to be JSON objects.")
            name = annotation.get("name")
            annotation_id = annotation.get("id")
            if annotation_id and name:
                by_name[name] = annotation

        if "Base" in by_name:
            return by_name["Base"]

        if len(annotations) == 1:
            annotation = annotations[0]
            annotation_id = annotation.get("id")
            if annotation_id:
                return annotation

        last_seen_names = sorted(name for name in by_name if name)
        time.sleep(ANNOTATION_WAIT_INTERVAL_SECONDS)

    fail(
        f"Timed out waiting for default annotation for dataset {dataset_id}. "
        f"Last seen annotation names: {last_seen_names}"
    )


def choose_slice_range(annotation: Dict[str, Any], edition_id: str) -> str:
    pages = str(annotation.get("pages") or "").strip()
    if not pages:
        print(
            f"Warning: default annotation for {edition_id} has no pages; using fallback {FALLBACK_SLICE_RANGE}",
            file=sys.stderr,
        )
        return FALLBACK_SLICE_RANGE

    match = re.fullmatch(r"(\d+)-(\d+)", pages)
    if not match:
        print(
            f"Warning: default annotation for {edition_id} has unsupported pages {pages!r}; using fallback {FALLBACK_SLICE_RANGE}",
            file=sys.stderr,
        )
        return FALLBACK_SLICE_RANGE

    start_page = int(match.group(1))
    end_page = int(match.group(2))
    if start_page > end_page:
        print(
            f"Warning: default annotation for {edition_id} has invalid pages {pages!r}; using fallback {FALLBACK_SLICE_RANGE}",
            file=sys.stderr,
        )
        return FALLBACK_SLICE_RANGE

    available_pages = end_page - start_page + 1
    if available_pages <= SLICE_PAGE_COUNT:
        return f"{start_page}-{end_page}"

    max_start = end_page - SLICE_PAGE_COUNT + 1
    slice_start = random.randint(start_page, max_start)
    slice_end = slice_start + SLICE_PAGE_COUNT - 1
    return f"{slice_start}-{slice_end}"


def build_sliced_rules_payload(slice_range: str) -> Dict[str, Any]:
    return {
        "action": "create_new",
        "copy_feature_results": False,
        "name": SLICED_ANNOTATION_NAME,
        "rules": [
            {
                "pages": slice_range,
                "type": "slice_pages",
            }
        ],
    }


def apply_sliced_annotation(dataset_id: str, annotation_id: str, slice_range: str) -> Dict[str, Any]:
    response = fetch_json(
        "PUT",
        f"/api/v1/datasets/{dataset_id}/annotations/{annotation_id}/apply",
        build_sliced_rules_payload(slice_range),
    )
    if not isinstance(response, dict):
        fail(
            f"Expected apply response for dataset {dataset_id} annotation {annotation_id} to be a JSON object."
        )
    return response


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
