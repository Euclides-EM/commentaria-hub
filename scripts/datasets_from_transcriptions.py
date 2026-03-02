#!/usr/bin/env python3

from __future__ import annotations

import json
import os
import re
import sys
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


def find_matching_dataset(
    datasets: list[dict[str, Any]],
    facsimile_id: str,
    pages_spec: str,
) -> dict[str, Any] | None:
    wanted_pages = parse_pages_spec(pages_spec)
    for ds in datasets:
        if ds.get("facsimile_id") != facsimile_id:
            continue
        try:
            ds_pages = parse_pages_spec(ds.get("pages"))
        except (TypeError, ValueError):
            continue
        if ds_pages == wanted_pages:
            return ds
    return None


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

    for transcription_dir in transcription_dirs:
        edition_id = transcription_dir.name
        matching_facsimiles = facsimiles_by_edition.get(edition_id, [])
        if not matching_facsimiles:
            print(f"ERROR: no facsimile found for transcription dir '{edition_id}'", file=sys.stderr)
            continue

        matching_facsimiles = sorted(matching_facsimiles, key=lambda f: str(f.get("id", "")))
        if len(matching_facsimiles) > 1:
            ids = ",".join(str(f.get("id", "")) for f in matching_facsimiles)
            print(
                f"WARN: multiple facsimiles found for '{edition_id}' ({ids}), using first",
                file=sys.stderr,
            )

        facsimile_id = str(matching_facsimiles[0].get("id", ""))
        if not facsimile_id:
            print(f"ERROR: facsimile for '{edition_id}' has no id", file=sys.stderr)
            continue

        print(f"{edition_id}: facsimile_id={facsimile_id}")

        pages = pages_from_transcription_dir(transcription_dir)
        if not pages:
            print(f"ERROR: no page-* dirs found in '{edition_id}'", file=sys.stderr)
            continue

        pages_spec = pages_to_spec(pages)
        existing = find_matching_dataset(datasets, facsimile_id, pages_spec)
        if existing:
            print(
                f"{edition_id}: dataset already exists id={existing.get('id')} pages={pages_spec}"
            )
            continue

        dataset_payload = {
            "name": f"{edition_id}_{DATASET_NAME_SUFFIX}",
            "edition_id": edition_id,
            "facsimile_id": facsimile_id,
            "pages": pages_spec,
        }
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
        print(f"{edition_id}: created dataset id={created.get('id')} pages={pages_spec}")
        datasets.append(created)

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
