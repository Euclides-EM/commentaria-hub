#!/usr/bin/env python3

from __future__ import annotations

import json
import os
import re
import sys
import threading
from concurrent.futures import ThreadPoolExecutor, as_completed
from pathlib import Path
from typing import Any
from urllib import error, parse, request

BASE_URL = "http://localhost:8085/api/v1"
DATASET_NAME_SUFFIX = "Book_X_Transcriptions"
TRANSCRIPTIONS_ROOT = (
        Path(__file__).resolve().parent.parent / "ocrflow" / "store" / "data" / "transcriptions"
)
REQUEST_TIMEOUT_SECONDS = 300
GITHUB_TOKEN_ENV_VAR = "GITHUB_TOKEN"
CREATE_ASYNC = False
CREATE_DEFAULT_ANNOTATION = True
DIR_PROCESSING_CONCURRENCY = 4

PAGE_DIR_PATTERN = re.compile(r"^page-(\d+)$")


def api_request(
        method: str,
        path: str,
        auth_token: str,
        query: dict[str, str] | None = None,
        body: dict[str, Any] | None = None,
) -> Any:
    url = f"{BASE_URL.rstrip('/')}/{path.lstrip('/')}"
    if query:
        url = f"{url}?{parse.urlencode(query)}"

    headers = {"Accept": "application/json"}
    headers["Authorization"] = f"Bearer {auth_token}"

    payload = None
    if body is not None:
        payload = json.dumps(body).encode("utf-8")
        headers["Content-Type"] = "application/json"

    req = request.Request(url=url, data=payload, headers=headers, method=method)
    try:
        with request.urlopen(req, timeout=REQUEST_TIMEOUT_SECONDS) as resp:
            raw = resp.read().decode("utf-8").strip()
    except error.HTTPError as exc:
        details = exc.read().decode("utf-8", errors="replace").strip()
        raise RuntimeError(f"{method} {url} failed: {exc.code} {exc.reason} {details}") from exc
    except error.URLError as exc:
        raise RuntimeError(f"{method} {url} failed: {exc.reason}") from exc

    if not raw:
        return None
    return json.loads(raw)


def parse_pages_spec(spec: str | None) -> tuple[int, ...]:
    if not spec:
        return ()
    pages: set[int] = set()
    for chunk in spec.split(","):
        chunk = chunk.strip()
        if not chunk:
            continue
        if "-" in chunk:
            start_s, end_s = chunk.split("-", 1)
            start = int(start_s)
            end = int(end_s)
            if end < start:
                start, end = end, start
            pages.update(range(start, end + 1))
        else:
            pages.add(int(chunk))
    return tuple(sorted(pages))


def pages_to_spec(pages: list[int]) -> str:
    if not pages:
        return ""
    parts: list[str] = []
    start = pages[0]
    end = pages[0]
    for page in pages[1:]:
        if page == end + 1:
            end = page
            continue
        parts.append(f"{start}-{end}" if start != end else str(start))
        start = page
        end = page
    parts.append(f"{start}-{end}" if start != end else str(start))
    return ",".join(parts)


def pages_from_transcription_dir(transcription_dir: Path) -> list[int]:
    pages: set[int] = set()
    for child in transcription_dir.iterdir():
        if not child.is_dir():
            continue
        match = PAGE_DIR_PATTERN.fullmatch(child.name)
        if not match:
            continue
        pages.add(int(match.group(1)))
    return sorted(pages)


def process_transcription_dir(
        transcription_dir: Path,
        facsimiles_by_edition: dict[str, list[dict[str, Any]]],
        github_token: str,
        existing_dataset_keys: set[tuple[str, tuple[int, ...]]],
        state_lock: threading.Lock,
        print_lock: threading.Lock,
) -> None:
    edition_id = transcription_dir.name
    matching_facsimiles = facsimiles_by_edition.get(edition_id, [])
    if not matching_facsimiles:
        with print_lock:
            print(f"ERROR: no facsimile found for transcription dir '{edition_id}'", file=sys.stderr)
        return

    matching_facsimiles = sorted(matching_facsimiles, key=lambda f: str(f.get("id", "")))
    if len(matching_facsimiles) > 1:
        ids = ",".join(str(f.get("id", "")) for f in matching_facsimiles)
        with print_lock:
            print(
                f"WARN: multiple facsimiles found for '{edition_id}' ({ids}), using first",
                file=sys.stderr,
            )

    facsimile_id = str(matching_facsimiles[0].get("id", ""))
    if not facsimile_id:
        with print_lock:
            print(f"ERROR: facsimile for '{edition_id}' has no id", file=sys.stderr)
        return

    with print_lock:
        print(f"{edition_id}: facsimile_id={facsimile_id}")

    pages = pages_from_transcription_dir(transcription_dir)
    if not pages:
        with print_lock:
            print(f"ERROR: no page-* dirs found in '{edition_id}'", file=sys.stderr)
        return

    pages_spec = pages_to_spec(pages)
    dataset_key = (facsimile_id, tuple(pages))

    with state_lock:
        if dataset_key in existing_dataset_keys:
            with print_lock:
                print(f"{edition_id}: dataset already exists pages={pages_spec}")
            return
        existing_dataset_keys.add(dataset_key)

    dataset_payload = {
        "name": f"{edition_id}_{DATASET_NAME_SUFFIX}",
        "edition_id": edition_id,
        "facsimile_id": facsimile_id,
        "pages": pages_spec,
    }
    try:
        created = api_request(
            "POST",
            "/datasets",
            auth_token=github_token,
            query={
                "async": str(CREATE_ASYNC).lower(),
                "create_default_annotation": str(CREATE_DEFAULT_ANNOTATION).lower(),
            },
            body=dataset_payload,
        )
    except Exception:
        with state_lock:
            existing_dataset_keys.discard(dataset_key)
        raise

    with print_lock:
        print(f"{edition_id}: created dataset id={created.get('id')} pages={pages_spec}")


def main() -> int:
    github_token = os.environ.get(GITHUB_TOKEN_ENV_VAR, "").strip()
    if not github_token:
        print(f"ERROR: {GITHUB_TOKEN_ENV_VAR} is not set", file=sys.stderr)
        return 1

    if not TRANSCRIPTIONS_ROOT.is_dir():
        print(f"ERROR: transcriptions dir not found: {TRANSCRIPTIONS_ROOT}", file=sys.stderr)
        return 1

    transcription_dirs = sorted(p for p in TRANSCRIPTIONS_ROOT.iterdir() if p.is_dir())
    facsimiles = api_request("GET", "/facsimilies", auth_token=github_token)
    datasets = api_request("GET", "/datasets", auth_token=github_token)

    facsimiles_by_edition: dict[str, list[dict[str, Any]]] = {}
    for fac in facsimiles:
        edition_id = fac.get("edition_id")
        if not edition_id:
            continue
        facsimiles_by_edition.setdefault(edition_id, []).append(fac)

    existing_dataset_keys: set[tuple[str, tuple[int, ...]]] = set()
    for ds in datasets:
        facsimile_id = ds.get("facsimile_id")
        if not facsimile_id:
            continue
        try:
            parsed_pages = parse_pages_spec(ds.get("pages"))
        except (TypeError, ValueError):
            continue
        existing_dataset_keys.add((str(facsimile_id), parsed_pages))

    state_lock = threading.Lock()
    print_lock = threading.Lock()

    with ThreadPoolExecutor(max_workers=DIR_PROCESSING_CONCURRENCY) as executor:
        futures = [
            executor.submit(
                process_transcription_dir,
                transcription_dir,
                facsimiles_by_edition,
                github_token,
                existing_dataset_keys,
                state_lock,
                print_lock,
            )
            for transcription_dir in transcription_dirs
        ]
        for future in as_completed(futures):
            future.result()

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
