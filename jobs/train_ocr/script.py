#!/usr/bin/env python3

from __future__ import annotations

import argparse
import os
import shlex
import subprocess
import sys
import time
from pathlib import Path
from typing import Optional
from zipfile import ZipFile

import torch


IMG_EXTS = [".jpg", ".jpeg", ".png", ".tif", ".tiff"]


def log(msg: str) -> None:
    print(f"[{time.strftime('%Y-%m-%d %H:%M:%S')}] {msg}", flush=True)


def run(cmd: list[str]) -> None:
    log("+ " + " ".join(shlex.quote(x) for x in cmd))
    subprocess.check_call(cmd)


def find_image_for(xml_path: Path) -> Optional[Path]:
    stemless = xml_path.with_suffix("")
    for ext in IMG_EXTS:
        p = stemless.with_suffix(ext)
        if p.exists():
            return p
    return None


def unzip_archives(zip_paths: list[Path], pages_dir: Path) -> None:
    pages_dir.mkdir(parents=True, exist_ok=True)

    for zp in zip_paths:
        if not zp.exists():
            raise FileNotFoundError(f"ZIP not found: {zp}")

        log(f"Unzipping {zp}")
        out_dir = pages_dir / zp.stem
        out_dir.mkdir(parents=True, exist_ok=True)

        with ZipFile(zp, "r") as zf:
            zf.extractall(out_dir)

    log(f"Extracted {len(zip_paths)} ZIP(s) into: {pages_dir}")


def build_manifest(pages_dir: Path, manifest_path: Path) -> int:
    xml_files: list[Path] = []

    for xml_path in pages_dir.rglob("*.xml"):
        if xml_path.name.lower() == "mets.xml":
            continue

        img_path = find_image_for(xml_path)
        if img_path is not None:
            xml_files.append(xml_path)
        else:
            log(f"Warning: no image found for {xml_path}")

    xml_files = sorted(xml_files)

    with manifest_path.open("w", encoding="utf-8") as f:
        for p in xml_files:
            f.write(str(p.resolve()) + "\n")

    log(f"Found {len(xml_files)} ALTO files with matching page images")
    log(f"Wrote manifest: {manifest_path}")

    if not xml_files:
        raise RuntimeError("No ALTO XML files with matching images were found")

    return len(xml_files)


def main() -> int:
    parser = argparse.ArgumentParser(description="Train Kraken OCR model from ALTO ZIP exports")

    parser.add_argument(
        "--zip-path",
        action="append",
        required=True,
        help="Path to a ZIP export. Repeat for multiple ZIPs.",
    )
    parser.add_argument(
        "--base-model-path",
        default="",
        help="Optional Kraken base model for fine-tuning. Leave empty to train from scratch.",
    )
    parser.add_argument(
        "--output-dir",
        default="trained_models",
        help="Directory for model outputs",
    )
    parser.add_argument(
        "--work-dir",
        default="work_kraken",
        help="Working directory for extracted pages, manifest, and dataset.arrow",
    )
    parser.add_argument(
        "--model-file-prefix",
        required=True,
        help="Prefix for model output files",
    )
    parser.add_argument(
        "--unicode-norm",
        default="NFD",
        choices=["NFD", "NFC", "NFKD", "NFKC"],
        help="Unicode normalization mode",
    )
    parser.add_argument(
        "--batch-size",
        type=int,
        default=2,
        help="Ketos batch size",
    )
    parser.add_argument(
        "--learning-rate",
        type=float,
        default=1e-4,
        help="Learning rate",
    )
    parser.add_argument(
        "--device",
        default="cuda:0",
        help="Torch/Ketos device, for example cuda:0 or cpu",
    )
    parser.add_argument(
        "--seed",
        type=int,
        default=42,
        help="Random seed for reproducibility where applicable",
    )
    parser.add_argument(
        "--overwrite",
        action="store_true",
        help="Allow reusing and overwriting work files",
    )

    args = parser.parse_args()

    os.environ.setdefault("PYTHONUNBUFFERED", "1")
    os.environ.setdefault("CUDA_DEVICE_ORDER", "PCI_BUS_ID")

    log(f"Python executable: {sys.executable}")
    log(f"CUDA available: {torch.cuda.is_available()}")
    log(f"CUDA version: {torch.version.cuda}")
    log(f"Torch version: {torch.__version__}")

    if args.device.startswith("cuda") and not torch.cuda.is_available():
        raise RuntimeError("CUDA device requested but no GPU is available")

    zip_paths = [Path(p).expanduser().resolve() for p in args.zip_path]
    base_model_path = Path(args.base_model_path).expanduser().resolve() if args.base_model_path else None
    output_dir = Path(args.output_dir).expanduser().resolve()
    work_dir = Path(args.work_dir).expanduser().resolve()
    pages_dir = work_dir / "pages_unzipped"
    manifest_path = work_dir / "alto_files.txt"
    dataset_path = work_dir / "dataset.arrow"
    model_prefix = output_dir / args.model_file_prefix

    log(f"ZIPs           = {zip_paths}")
    log(f"Base model     = {base_model_path if base_model_path else '(training from scratch)'}")
    log(f"Output dir     = {output_dir}")
    log(f"Work dir       = {work_dir}")
    log(f"Model prefix   = {model_prefix}")
    log(f"Unicode norm   = {args.unicode_norm}")
    log(f"Batch size     = {args.batch_size}")
    log(f"Learning rate  = {args.learning_rate}")
    log(f"Device         = {args.device}")

    output_dir.mkdir(parents=True, exist_ok=True)
    work_dir.mkdir(parents=True, exist_ok=True)

    if args.overwrite:
        if manifest_path.exists():
            manifest_path.unlink()
        if dataset_path.exists():
            dataset_path.unlink()

    unzip_archives(zip_paths, pages_dir)
    build_manifest(pages_dir, manifest_path)

    run([
        "ketos",
        "compile",
        "-F", str(manifest_path),
        "--random-split", "0.8", "0.1", "0.1",
        "-f", "alto",
        "-o", str(dataset_path),
    ])

    train_cmd = [
        "ketos",
        "-d", args.device,
        "train",
        "-f", "binary",
        "-B", str(args.batch_size),
        "-r", str(args.learning_rate),
        "-u", args.unicode_norm,
        "-o", str(model_prefix),
        str(dataset_path),
    ]

    if base_model_path:
        if not base_model_path.exists():
            raise FileNotFoundError(f"Base model not found: {base_model_path}")
        train_cmd.extend(["-i", str(base_model_path), "--resize", "add"])

    run(train_cmd)

    log("Training finished successfully")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())