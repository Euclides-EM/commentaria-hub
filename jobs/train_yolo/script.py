#!/usr/bin/env python3

"""
Train a YOLO model on a Roboflow dataset zip.

Key points:
- Dataset URL is REQUIRED
- Downloads dataset + model only if missing
- Summarizes dataset splits
- Trains YOLO model
- Writes artifacts/result.json with best model path
- --dry-run skips heavy operations (download + training)
"""

from __future__ import annotations

import argparse
import json
import os
import subprocess
import time
from pathlib import Path
from typing import Dict
import zipfile
import yaml
from urllib.parse import urlparse, unquote

from ultralytics import YOLO


DEFAULT_MODEL_URL = "https://zenodo.org/records/10972956/files/CapricciosaX.pt?download=1"
DEFAULT_MODEL_PATH = "CapricciosaX.pt"
DEFAULT_PROJECT = "runs_capricciosa"


# ---------- Utilities ----------

def log(msg: str) -> None:
    print(f"[{time.strftime('%Y-%m-%d %H:%M:%S')}] {msg}", flush=True)


def run(cmd: list[str]) -> None:
    log("+ " + " ".join(cmd))
    subprocess.check_call(cmd)


def download_if_missing(url: str, dst: Path, dry_run: bool) -> None:
    if dst.exists() and dst.stat().st_size > 0:
        log(f"Exists, skip download: {dst}")
        return
    dst.parent.mkdir(parents=True, exist_ok=True)
    if dry_run:
        run(["wget", "--spider", url])
    else:
        run(["wget", "-O", str(dst), url])


def extract_zip_if_missing(zip_path: Path, out_dir: Path, dry_run: bool) -> None:
    if out_dir.exists() and any(out_dir.iterdir()):
        log(f"Exists, skip extract: {out_dir}")
        return

    out_dir.mkdir(parents=True, exist_ok=True)

    if dry_run:
        log(f"[DRY-RUN] Would extract {zip_path}")
        return

    log(f"Extracting {zip_path}")
    with zipfile.ZipFile(zip_path, "r") as z:
        z.extractall(out_dir)


def count_images(folder: Path) -> int:
    exts = (".jpg", ".jpeg", ".png")
    return sum(1 for p in folder.rglob("*") if p.suffix.lower() in exts)


# ---------- YAML Fix ----------

def fix_data_yaml(original_yaml: Path) -> Path:
    """
    Make Roboflow YAML resilient to extraction location.

    Handles common cases:
    - train: ../train/images but YAML is already in dataset root
    - val points to ../valid/images (Roboflow folder is 'valid', key is 'val')
    - absolute paths or already-correct relative paths remain unchanged
    """
    cfg = yaml.safe_load(original_yaml.read_text())
    base = original_yaml.parent

    fixed = dict(cfg)

    def resolve_split_path(rel: str) -> tuple[str, Path]:
        # Keep absolute paths as-is
        p = Path(rel)
        if p.is_absolute():
            return rel, p

        # 1) As written, relative to YAML location
        cand1 = (base / rel).resolve()
        if cand1.exists():
            return rel, cand1

        # 2) Strip leading ../ segments (Roboflow export mismatch)
        rel2 = rel
        while rel2.startswith("../"):
            rel2 = rel2[3:]
        cand2 = (base / rel2).resolve()
        if cand2.exists():
            return rel2, cand2

        # 3) If it refers to "val/..." but folder is "valid/..."
        rel3 = rel2.replace("val/", "valid/", 1)
        cand3 = (base / rel3).resolve()
        if cand3.exists():
            return rel3, cand3

        # Give up, return the best-effort rel2 (more likely correct)
        return rel2, cand2

    # Some Roboflow exports use "valid" key instead of "val"
    if "val" not in fixed and "valid" in fixed:
        fixed["val"] = fixed["valid"]

    for split in ("train", "val", "test"):
        if split not in fixed:
            continue
        rel = str(fixed[split])
        new_rel, resolved = resolve_split_path(rel)
        log(f"YAML {split}: {rel} -> {new_rel} (resolved: {resolved})")
        fixed[split] = new_rel

    # Write next to original YAML so paths stay relative to dataset root
    fixed_yaml = base / "data.fixed.yaml"
    fixed_yaml.write_text(yaml.safe_dump(fixed, sort_keys=False))

    return fixed_yaml


def summarize_dataset(data_yaml: Path) -> Dict[str, int]:
    cfg = yaml.safe_load(data_yaml.read_text())
    base = data_yaml.parent
    counts = {}

    for split in ("train", "val", "test"):
        if split not in cfg:
            continue

        p = Path(cfg[split])
        p = p if p.is_absolute() else (base / p).resolve()
        counts[split] = count_images(p) if p.exists() else 0
        log(f"{split}: {p} ({counts[split]} images)")

    err =  False
    for split in ("train", "val", "test"):
        if split not in counts:
            log(f"{split}: missing from YAML")
            err = True
        if split in counts and counts[split] == 0:
            log(f"{split}: no images found at {base / cfg[split]}")
            err = True

    if err:
        raise ValueError("Dataset YAML is missing splits or images, see logs for details")

    return counts

def default_dataset_name_from_url(url: str) -> str:
    p = urlparse(url)
    filename = Path(unquote(p.path)).name  # last path component, without query string
    stem = Path(filename).stem
    return stem or "dataset"


# ---------- Main ----------

def main() -> int:
    parser = argparse.ArgumentParser(description="Train YOLO model")
    parser.add_argument("--dataset-url", required=True)
    parser.add_argument("--dataset-name")
    parser.add_argument("--model-url", default=DEFAULT_MODEL_URL)
    parser.add_argument("--model-path", default=DEFAULT_MODEL_PATH)
    parser.add_argument("--project", default=DEFAULT_PROJECT)
    parser.add_argument("--name")
    parser.add_argument("--epochs", type=int, default=50)
    parser.add_argument("--imgsz", type=int, default=640)
    parser.add_argument("--batch", type=int, default=16)
    parser.add_argument("--workers", type=int, default=2)
    parser.add_argument("--dry-run", action="store_true", help="Skip downloads and training")

    args = parser.parse_args()

    os.environ["TORCHDYNAMO_DISABLE"] = "1"

    dataset_name = args.dataset_name or default_dataset_name_from_url(args.dataset_url)
    run_name = args.name or f"{dataset_name}_finetune"

    dataset_zip = Path(f"{dataset_name}.zip")
    dataset_dir = Path(dataset_name)
    data_yaml = dataset_dir / "data.yaml"
    model_path = Path(args.model_path)

    log(f"Dry run: {args.dry_run}")

    # Dataset handling
    download_if_missing(args.dataset_url, dataset_zip, args.dry_run)
    extract_zip_if_missing(dataset_zip, dataset_dir, args.dry_run)

    fixed_yaml = None
    counts = {}
    if data_yaml.exists():
        fixed_yaml = fix_data_yaml(data_yaml)
        counts = summarize_dataset(fixed_yaml)
    else:
        if args.dry_run:
            log(f"[DRY-RUN] {data_yaml} not present (expected because extraction is skipped)")
        else:
            if fixed_yaml is None:
                raise RuntimeError("fixed_yaml is None, data.yaml was not found")


    # Model download
    download_if_missing(args.model_url, model_path, args.dry_run)

    best_pt = None

    # Training
    if args.dry_run:
        log("[DRY-RUN] Skipping training")
    else:
        if fixed_yaml is None:
            raise RuntimeError("fixed_yaml is None, cannot train without data.yaml")
        log("Starting YOLO training")
        model = YOLO(str(model_path))

        model.train(
            data=str(fixed_yaml.resolve()),
            epochs=args.epochs,
            imgsz=args.imgsz,
            batch=args.batch,
            workers=args.workers,
            project=args.project,
            name=run_name,
            exist_ok=True,
        )

        best_pt = Path(args.project) / run_name / "weights" / "best.pt"

    artifact = {
        "dataset_url": args.dataset_url,
        "dataset_name": dataset_name,
        "best_model": str(best_pt) if best_pt else None,
        "split_counts": counts,
        "dry_run": args.dry_run,
    }

    Path("artifacts").mkdir(exist_ok=True)
    Path("artifacts/result.json").write_text(json.dumps(artifact, indent=2))

    log("Done")

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
