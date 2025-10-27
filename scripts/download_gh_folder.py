#!/usr/bin/env python3
import os
import sys
import time
import argparse
import requests
from urllib.parse import urlparse

API_BASE = "https://api.github.com"

def github_headers():
    hdrs = {
        "Accept": "application/vnd.github+json",
        "User-Agent": "gh-folder-downloader"
    }
    token = ""
    if token:
        hdrs["Authorization"] = f"Bearer {token}"
    return hdrs

def parse_github_tree_url(url: str):
    """
    Expected form:
    https://github.com/<owner>/<repo>/tree/<ref>/<path...>
    """
    p = urlparse(url)
    parts = [s for s in p.path.split("/") if s]
    if len(parts) < 5 or parts[2] != "tree":
        raise ValueError("URL must look like https://github.com/<owner>/<repo>/tree/<ref>/<path>")
    owner, repo, _, ref = parts[:4]
    path = "/".join(parts[4:])
    return owner, repo, ref, path

def safe_join(dest_root: str, rel_path: str) -> str:
    dest = os.path.abspath(os.path.join(dest_root, rel_path))
    root = os.path.abspath(dest_root)
    if not dest.startswith(root + os.sep) and dest != root:
        raise ValueError(f"Refusing to write outside destination: {dest}")
    return dest

def mkdirs(path: str):
    os.makedirs(path, exist_ok=True)

def download_file(download_url: str, dest_path: str, retry=3):
    for attempt in range(retry):
        r = requests.get(download_url, stream=True, headers=github_headers(), timeout=60)
        if r.status_code == 200:
            mkdirs(os.path.dirname(dest_path))
            with open(dest_path, "wb") as f:
                for chunk in r.iter_content(chunk_size=1 << 20):
                    if chunk:
                        f.write(chunk)
            return
        # gentle backoff on transient errors
        if r.status_code in (429, 502, 503, 504):
            time.sleep(2 ** attempt)
            continue
        if r.status_code == 403:
            rem = r.headers.get("X-RateLimit-Remaining")
            if rem == "0":
                raise RuntimeError("GitHub rate limit hit. Set GITHUB_TOKEN to increase limits")
        raise RuntimeError(f"Failed to download {download_url} (HTTP {r.status_code})")
    raise RuntimeError(f"Failed to download after retries: {download_url}")

def list_dir(owner: str, repo: str, path: str, ref: str):
    url = f"{API_BASE}/repos/{owner}/{repo}/contents/{path}?ref={ref}"
    r = requests.get(url, headers=github_headers(), timeout=60)
    if r.status_code == 404:
        raise RuntimeError(f"Not found: {url}")
    if r.status_code == 403 and r.headers.get("X-RateLimit-Remaining") == "0":
        raise RuntimeError("GitHub rate limit hit. Set GITHUB_TOKEN")
    r.raise_for_status()
    return r.json()

def recurse_download(owner: str, repo: str, ref: str, root_path: str, dest_root: str, rel_path: str = ""):
    api_path = f"{root_path}/{rel_path}".strip("/")
    entries = list_dir(owner, repo, api_path, ref)

    # GitHub returns either a list for directories or a file object for single files
    if isinstance(entries, dict) and entries.get("type") == "file":
        # single file path given
        entry = entries
        rel = entry["path"][len(root_path):].lstrip("/")
        dest = safe_join(dest_root, rel)
        print(f"File: {entry['path']} -> {dest}")
        download_file(entry["download_url"], dest)
        return

    for e in entries:
        e_type = e.get("type")
        if e_type == "file":
            rel = e["path"][len(root_path):].lstrip("/")
            dest = safe_join(dest_root, rel)
            print(f"File: {e['path']} -> {dest}")
            download_file(e["download_url"], dest)
        elif e_type == "dir":
            sub_rel = e["path"][len(root_path):].lstrip("/")
            print(f"Dir:  {e['path']}")
            recurse_download(owner, repo, ref, root_path, dest_root, sub_rel)
        else:
            # skip symlinks or submodules
            print(f"Skip: {e.get('path')} (type {e_type})")

def main():
    ap = argparse.ArgumentParser(description="Download all files from a GitHub folder recursively")
    ap.add_argument("url", help="GitHub folder URL, e.g. https://github.com/OCR-D/gt_structure_text/tree/main/data/alberti_pictura_1540")
    ap.add_argument("-o", "--output", default="downloaded",
                    help="Destination directory (default: downloaded)")
    args = ap.parse_args()

    owner, repo, ref, path = parse_github_tree_url(args.url)
    dest_root = os.path.abspath(args.output)
    print(f"Repo: {owner}/{repo}  Ref: {ref}  Path: {path}")
    print(f"Dest: {dest_root}")
    mkdirs(dest_root)

    recurse_download(owner, repo, ref, path, dest_root)

if __name__ == "__main__":
    main()
