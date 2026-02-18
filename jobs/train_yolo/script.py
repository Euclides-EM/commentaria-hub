#!/usr/bin/env python3

"""
Train a YOLO model on a Roboflow dataset zip.

Key points:
- Dataset URL is REQUIRED
- Downloads dataset + model only if missing
- Summarizes dataset splits
- Trains YOLO model
- Writes artifacts/result.json with best model path
"""

from __future__ import annotations

import argparse
import json
import os
import subprocess
import sys
import time
from pathlib import Path
from typing import Any, Dict, Optional
import zipfile
import yaml

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


def download_if_missing(url: str, dst: Path) -> None:
    if dst.exists() and dst.stat().st_size > 0:
        log(f"Exists, skip download: {dst}")
        return
    dst.parent.mkdir(parents=True, exist_ok=True)
    run(["wget", "-O", str(dst), url])


def extract_zip_if_missing(zip_path: Path, out_dir: Path) -> None:
    if out_dir.exists() and any(out_dir.iterdir()):
        log(f"Exists, skip extract: {out_dir}")
        return

    out_dir.mkdir(parents=True, exist_ok=True)
    log(f"Extracting {zip_path}")
    with zipfile.ZipFile(zip_path, "r") as z:
        z.extractall(out_dir)


def count_images(folder: Path) -> int:
    exts = (".jpg", ".jpeg", ".png")
    return sum(1 for p in folder.rglob("*") if p.suffix.lower() in exts)


def summarize_dataset(data_yaml: Path) -> Dict[str, int]:
    cfg = yaml.safe_load(data_yaml.read_text())
    base = data_yaml.parent
    counts = {}

    for split in ("train", "val", "test"):
        if split not in cfg:
            continue
        p = (base / cfg[split]).resolve()
        counts[split] = count_images(p) if p.exists() else 0
        log(f"{split}: {p} ({counts[split]} images)")

    return counts


# ---------- Main ----------

def main() -> int:
    parser = argparse.ArgumentParser(description="Train YOLO model")
    parser.add_argument(
        "--dataset-url",
        required=True,
        help="Roboflow dataset zip URL (required)",
    )
    parser.add_argument(
        "--dataset-name",
        help="Dataset folder name (default derived from URL)",
    )
    parser.add_argument("--model-url", default=DEFAULT_MODEL_URL)
    parser.add_argument("--model-path", default=DEFAULT_MODEL_PATH)
    parser.add_argument("--project", default=DEFAULT_PROJECT)
    parser.add_argument("--name", help="Run name")
    parser.add_argument("--epochs", type=int, default=50)
    parser.add_argument("--imgsz", type=int, default=640)
    parser.add_argument("--batch", type=int, default=16)
    parser.add_argument("--workers", type=int, default=2)

    args = parser.parse_args()

    os.environ["TORCHDYNAMO_DISABLE"] = "1"

    dataset_name = args.dataset_name or Path(args.dataset_url).stem
    run_name = args.name or f"{dataset_name}_finetune"

    dataset_zip = Path(f"{dataset_name}.zip")
    dataset_dir = Path(dataset_name)
    data_yaml = dataset_dir / "data.yaml"
    model_path = Path(args.model_path)

    # Download dataset
    log("Downloading dataset")
    download_if_missing(args.dataset_url, dataset_zip)
    extract_zip_if_missing(dataset_zip, dataset_dir)

    if not data_yaml.exists():
        raise FileNotFoundError(f"{data_yaml} missing after extraction")

    counts = summarize_dataset(data_yaml)

    # Download model
    log("Downloading base model")
    download_if_missing(args.model_url, model_path)

    # Train
    log("Starting YOLO training")
    model = YOLO(str(model_path))

    results = model.train(
        data=str(data_yaml.resolve()),
        epochs=args.epochs,
        imgsz=args.imgsz,
        batch=args.batch,
        workers=args.workers,
        project=args.project,
        name=run_name,
        exist_ok=True,
    )

    run_dir = Path(args.project) / run_name
    best_pt = run_dir / "weights" / "best.pt"

    artifact = {
        "dataset_url": args.dataset_url,
        "dataset_name": dataset_name,
        "best_model": str(best_pt),
        "split_counts": counts,
    }

    Path("artifacts").mkdir(exist_ok=True)
    Path("artifacts/result.json").write_text(json.dumps(artifact, indent=2))

    log(f"Best model: {best_pt}")
    log("Done")

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
