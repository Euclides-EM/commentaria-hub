#!/usr/bin/env python3
"""Open every page of matched Transkribus documents and wait for its download."""

import argparse
import csv
import re
import shutil
import subprocess
import time
import webbrowser
from pathlib import Path


DEFAULT_CSV = Path(__file__).with_name("index.csv")
DEFAULT_DOWNLOADS = Path("/mnt/c/Users/reall/Downloads")
DEFAULT_OUTPUT_ROOT = Path.cwd() / "pages"
DEFAULT_DOWNLOAD_TIMEOUT_SECONDS = 10.0
PAGE_DELAY_SECONDS = 0.2
MAX_BROWSER_OPEN_RETRIES = 1
MAX_BROWSER_OPEN_ATTEMPTS = MAX_BROWSER_OPEN_RETRIES + 1
MAX_DOWNLOAD_RETRIES = 1
MAX_DOWNLOAD_ATTEMPTS = MAX_DOWNLOAD_RETRIES + 1
DOWNLOAD_POLL_INTERVAL_SECONDS = 0.25
DOWNLOAD_STABLE_OBSERVATIONS = 2
FIRST_PAGE = 1

TRANSKRIBUS_PAGE_URL = (
    "https://app.transkribus.org/sites/noscemus/doc/"
    "{doc_id}/detail?pageid={page}"
)
DOWNLOAD_FILENAME = "{doc_id}_{page}.txt"
PAGE_FILENAME = "{page}.txt"
DOC_ID_PATTERN = re.compile(r"/doc/(\d+)(?:[/?]|$)")
REQUIRED_CSV_COLUMNS = frozenset(
    {"href", "page_count", "exists_in_commentaria"}
)
CSV_ENCODING = "utf-8"
WSL_VERSION_FILE = Path("/proc/version")
WSL_VERSION_MARKER = "microsoft"
WINDOWS_COMMAND = "cmd.exe"
WINDOWS_START_ARGUMENTS = ("/c", "start", "")
BROWSER_NEW_TAB = 2


def parse_args():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--csv", type=Path, default=DEFAULT_CSV)
    parser.add_argument("--downloads", type=Path, default=DEFAULT_DOWNLOADS)
    parser.add_argument(
        "--output-root",
        type=Path,
        default=DEFAULT_OUTPUT_ROOT,
        help="Destination root (default: ./pages).",
    )
    parser.add_argument(
        "--timeout",
        type=float,
        default=DEFAULT_DOWNLOAD_TIMEOUT_SECONDS,
    )
    parser.add_argument(
        "--start-doc-id",
        help="Skip rows until this Transkribus document ID (useful when resuming).",
    )
    parser.add_argument(
        "--start-page",
        type=int,
        default=FIRST_PAGE,
        help="First page for --start-doc-id; subsequent documents start at page 1.",
    )
    return parser.parse_args()


def extract_doc_id(href):
    match = DOC_ID_PATTERN.search(href)
    if not match:
        raise ValueError(f"Could not extract document ID from href: {href!r}")
    return match.group(1)


def file_state(path):
    try:
        stat = path.stat()
    except (FileNotFoundError, PermissionError, OSError):
        return None
    return stat.st_size, stat.st_mtime_ns, stat.st_ctime_ns


def wait_for_download(expected_file, before_state, timeout):
    """Return the expected new/modified file once its size remains stable."""
    deadline = time.monotonic() + timeout
    previous_state = None
    stable_observations = 0

    while time.monotonic() < deadline:
        time.sleep(DOWNLOAD_POLL_INTERVAL_SECONDS)
        state = file_state(expected_file)
        if state is None or state == before_state:
            continue

        if previous_state == state:
            stable_observations += 1
        else:
            stable_observations = 0
        previous_state = state

        # Repeated unchanged observations avoid advancing while a browser is
        # still writing the final file without a temporary suffix.
        if stable_observations >= DOWNLOAD_STABLE_OBSERVATIONS:
            return expected_file

    return None


def move_download(source, destination):
    destination.parent.mkdir(parents=True, exist_ok=True)
    if destination.exists():
        raise FileExistsError(f"Refusing to overwrite existing page: {destination}")
    return Path(shutil.move(str(source), str(destination)))


def running_under_wsl():
    try:
        return WSL_VERSION_MARKER in WSL_VERSION_FILE.read_text().lower()
    except OSError:
        return False


def open_in_browser(url):
    # Python's webbrowser module often cannot find the Windows browser from WSL.
    if running_under_wsl() and shutil.which(WINDOWS_COMMAND):
        subprocess.Popen(
            [WINDOWS_COMMAND, *WINDOWS_START_ARGUMENTS, url],
            stdout=subprocess.DEVNULL,
            stderr=subprocess.DEVNULL,
        )
        return

    if not webbrowser.open(url, new=BROWSER_NEW_TAB):
        raise RuntimeError("No usable web browser was found")


def load_matched_rows(csv_path):
    with csv_path.open(encoding=CSV_ENCODING, newline="") as source:
        reader = csv.DictReader(source)
        missing = REQUIRED_CSV_COLUMNS - set(reader.fieldnames or ())
        if missing:
            raise ValueError(f"{csv_path} is missing columns: {sorted(missing)}")
        return [row for row in reader if row["exists_in_commentaria"].strip()]


def main():
    args = parse_args()
    if args.timeout <= 0:
        raise ValueError("--timeout must be greater than zero")
    if args.start_page < FIRST_PAGE:
        raise ValueError("--start-page must be at least 1")
    if not args.downloads.is_dir():
        raise FileNotFoundError(f"Downloads directory not found: {args.downloads}")

    rows = load_matched_rows(args.csv)
    waiting_for_start = bool(args.start_doc_id)
    browser_open_attempts = browser_open_failures = 0
    downloaded = download_timeouts = skipped = 0

    for row_number, row in enumerate(rows, start=1):
        doc_id = extract_doc_id(row["href"])
        if waiting_for_start:
            if doc_id != args.start_doc_id:
                continue
            waiting_for_start = False

        try:
            page_count = int(row["page_count"])
        except ValueError as error:
            raise ValueError(
                f"Invalid page_count for document {doc_id}: {row['page_count']!r}"
            ) from error

        first_page = (
            args.start_page if doc_id == args.start_doc_id else FIRST_PAGE
        )
        if first_page > page_count:
            print(
                f"[document {row_number}/{len(rows)}] {doc_id}: "
                f"start page {first_page} exceeds page count {page_count}; skipped",
                flush=True,
            )
            continue

        print(
            f"[document {row_number}/{len(rows)}] {doc_id}: "
            f"{row['title']} ({page_count} pages)",
            flush=True,
        )

        for page in range(first_page, page_count + 1):
            target_file = (
                args.output_root
                / doc_id
                / PAGE_FILENAME.format(page=page)
            )
            if target_file.is_file():
                skipped += 1
                print(
                    f"  [page {page}/{page_count}] already exists; "
                    f"skipping {target_file}",
                    flush=True,
                )
                continue

            url = TRANSKRIBUS_PAGE_URL.format(doc_id=doc_id, page=page)
            expected_download = args.downloads / DOWNLOAD_FILENAME.format(
                doc_id=doc_id,
                page=page,
            )
            for download_attempt in range(1, MAX_DOWNLOAD_ATTEMPTS + 1):
                before_state = file_state(expected_download)
                opened_successfully = False

                for open_attempt in range(1, MAX_BROWSER_OPEN_ATTEMPTS + 1):
                    if browser_open_attempts:
                        time.sleep(PAGE_DELAY_SECONDS)

                    browser_open_attempts += 1
                    print(
                        f"  [page {page}/{page_count}] "
                        f"[download attempt "
                        f"{download_attempt}/{MAX_DOWNLOAD_ATTEMPTS}] "
                        f"[open attempt "
                        f"{open_attempt}/{MAX_BROWSER_OPEN_ATTEMPTS}] "
                        f"opening {url}",
                        flush=True,
                    )
                    try:
                        open_in_browser(url)
                    except (
                        OSError,
                        RuntimeError,
                        subprocess.SubprocessError,
                    ) as error:
                        print(
                            f"  [page {page}/{page_count}] browser open "
                            f"failed: {error}",
                            flush=True,
                        )
                    else:
                        opened_successfully = True
                        break

                if not opened_successfully:
                    browser_open_failures += 1
                    print(
                        f"  [page {page}/{page_count}] browser open failed "
                        f"after {MAX_BROWSER_OPEN_ATTEMPTS} attempts; continuing",
                        flush=True,
                    )
                    break

                downloaded_file = wait_for_download(
                    expected_download, before_state, args.timeout
                )
                if downloaded_file is not None:
                    moved_file = move_download(downloaded_file, target_file)
                    downloaded += 1
                    print(
                        f"  [page {page}/{page_count}] downloaded and moved "
                        f"to {moved_file}",
                        flush=True,
                    )
                    break

                if download_attempt < MAX_DOWNLOAD_ATTEMPTS:
                    print(
                        f"  [page {page}/{page_count}] timeout after "
                        f"{args.timeout:g}s "
                        f"(download attempt "
                        f"{download_attempt}/{MAX_DOWNLOAD_ATTEMPTS}); "
                        f"reopening page",
                        flush=True,
                    )
                else:
                    download_timeouts += 1
                    print(
                        f"  [page {page}/{page_count}] timeout after "
                        f"{MAX_DOWNLOAD_ATTEMPTS} download attempts; continuing",
                        flush=True,
                    )

    if waiting_for_start:
        raise ValueError(f"--start-doc-id {args.start_doc_id!r} was not found")

    print(
        f"Finished: browser_open_attempts={browser_open_attempts}, "
        f"browser_open_failures={browser_open_failures}, "
        f"downloaded_and_moved={downloaded}, "
        f"download_timeouts={download_timeouts}, "
        f"skipped_existing={skipped}",
        flush=True,
    )


if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt:
        print("\nInterrupted by user.", flush=True)
        raise SystemExit(130)
