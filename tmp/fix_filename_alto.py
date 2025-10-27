#!/usr/bin/env python3
import argparse
import os
import re
from pathlib import Path
import shutil

# Regex to capture <fileName> with optional namespace prefix and attributes, preserving tags
# Groups:
#  1: opening tag up to '>'
#  2: current value
#  3: closing tag
FILE_NAME_TAG_RE = re.compile(
    r"(<(?P<prefix>[A-Za-z_][\w\.-]*:)?fileName\b[^>]*>)(?P<val>[^<]*)(</(?P=prefix)?fileName>)",
    re.DOTALL
)

ENCODING_RE = re.compile(rb'<\?xml[^>]*encoding=["\']([A-Za-z0-9_\-]+)["\']', re.I)

def detect_encoding(p: Path, default="utf-8"):
    with p.open("rb") as f:
        head = f.read(512)
    m = ENCODING_RE.search(head)
    if not m:
        return default
    enc = m.group(1).decode("ascii", errors="ignore")
    # Common XML encodings; fall back to default if Python does not know it
    try:
        "".encode(enc)
        return enc
    except LookupError:
        return default

def replace_fileName_basenames(text: str):
    def _sub(m: re.Match):
        val = m.group("val") or ""
        # Only change if value looks path-like
        if ("/" in val) or ("\\" in val):
            base = os.path.basename(val.strip())
            return f"{m.group(1)}{base}{m.group(4)}"
        return m.group(0)
    return FILE_NAME_TAG_RE.sub(_sub, text)

def process_dir(in_dir: Path, out_dir: Path, recursive: bool, dry_run: bool):
    if not in_dir.is_dir():
        raise SystemExit(f"Input path is not a directory: {in_dir}")
    out_dir.mkdir(parents=True, exist_ok=True)

    pattern = "**/*.xml" if recursive else "*.xml"
    for src in in_dir.glob(pattern):
        rel = src.relative_to(in_dir)
        dst = out_dir / rel
        dst.parent.mkdir(parents=True, exist_ok=True)

        enc = detect_encoding(src)
        original = src.read_text(encoding=enc, errors="strict")
        fixed = replace_fileName_basenames(original)

        if fixed != original:
            if dry_run:
                print(f"[dry-run] would change: {src}")
                # copy original to keep structure visible
                if not dst.exists():
                    shutil.copy2(src, dst)
            else:
                dst.write_text(fixed, encoding=enc)
                print(f"fixed: {src} -> {dst}")
        else:
            # no change, copy as is
            if dry_run:
                print(f"[dry-run] no change: {src}")
            if not dst.exists():
                shutil.copy2(src, dst)

def main():
    ap = argparse.ArgumentParser(
        description="Rewrite only the inner text of <fileName> to its basename, preserving XML bytes elsewhere"
    )
    ap.add_argument("folder", type=Path, help="Input folder with XML files")
    ap.add_argument("-o", "--out", type=Path, help="Output folder, default <input>_fixed")
    ap.add_argument("-r", "--recursive", action="store_true", help="Recurse into subfolders")
    ap.add_argument("--dry-run", action="store_true", help="Show what would change")
    args = ap.parse_args()

    in_dir = args.folder.resolve()
    out_dir = (args.out or Path(str(in_dir) + "_fixed")).resolve()

    process_dir(in_dir, out_dir, recursive=args.recursive, dry_run=args.dry_run)
    if not args.dry_run:
        print(f"Done. Output at: {out_dir}")

if __name__ == "__main__":
    main()
