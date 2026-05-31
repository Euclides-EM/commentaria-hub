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
import mimetypes
import os
import shutil
import shlex
import subprocess
import time
import urllib.error
import urllib.request
import uuid
from pathlib import Path
from typing import Dict, Optional
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
    log("+ " + " ".join(shlex.quote(x) for x in cmd))
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


def copy_or_download_dataset(dataset_url: Optional[str], dataset_zip_path: Optional[str], dst: Path, dry_run: bool) -> None:
    if dataset_zip_path:
        src = Path(dataset_zip_path)
        if not src.exists():
            raise FileNotFoundError(f"Dataset ZIP not found: {src}")
        if dst.exists() and dst.stat().st_size > 0:
            log(f"Exists, skip dataset copy: {dst}")
            return
        dst.parent.mkdir(parents=True, exist_ok=True)
        if dry_run:
            log(f"[DRY-RUN] Would copy {src} to {dst}")
        else:
            shutil.copyfile(src, dst)
        return

    if not dataset_url:
        raise ValueError("--dataset-url or --dataset-zip-path is required")
    download_if_missing(dataset_url, dst, dry_run)


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
    for split in ("train", "val"):
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


def encode_multipart(fields: dict[str, str], file_field: str, file_path: Path) -> tuple[bytes, str]:
    boundary = "----commentaria-yolo-" + uuid.uuid4().hex
    chunks: list[bytes] = []

    for name, value in fields.items():
        chunks.append(f"--{boundary}\r\n".encode("utf-8"))
        chunks.append(f'Content-Disposition: form-data; name="{name}"\r\n\r\n'.encode("utf-8"))
        chunks.append(value.encode("utf-8"))
        chunks.append(b"\r\n")

    content_type = mimetypes.guess_type(file_path.name)[0] or "application/octet-stream"
    chunks.append(f"--{boundary}\r\n".encode("utf-8"))
    chunks.append(
        (
            f'Content-Disposition: form-data; name="{file_field}"; filename="{file_path.name}"\r\n'
            f"Content-Type: {content_type}\r\n\r\n"
        ).encode("utf-8")
    )
    chunks.append(file_path.read_bytes())
    chunks.append(b"\r\n")
    chunks.append(f"--{boundary}--\r\n".encode("utf-8"))

    return b"".join(chunks), f"multipart/form-data; boundary={boundary}"


def upload_model(
    upload_url: str,
    upload_token: str,
    model_path: Path,
    name: str,
    description: str,
    base_annotations: str,
    base_model_id: str,
) -> None:
    curl_path = shutil.which("curl")
    if curl_path:
        cmd = [
            curl_path,
            "--fail",
            "--show-error",
            "--silent",
            "--location",
            "-H",
            "Expect:",
            "-F",
            f"file=@{model_path};type=application/octet-stream",
            "-F",
            f"name={name}",
            "-F",
            f"description={description}",
            "-F",
            f"base_annotations={base_annotations}",
            "-F",
            f"base_model_id={base_model_id}",
            upload_url,
        ]
        if upload_token:
            cmd[7:7] = ["-H", f"Authorization: Bearer {upload_token}"]

        redacted_cmd = [
            "Authorization: Bearer <redacted>" if part.startswith("Authorization: Bearer ") else part
            for part in cmd
        ]
        log("+ " + " ".join(shlex.quote(x) for x in redacted_cmd))
        result = subprocess.run(cmd, text=True, capture_output=True, check=False)
        if result.returncode != 0:
            detail = "\n".join(part for part in [result.stdout.strip(), result.stderr.strip()] if part)
            raise RuntimeError(f"Model upload failed with curl exit {result.returncode}: {detail}")
        if result.stdout.strip():
            log(f"Import response: {result.stdout.strip()}")
        log(f"Uploaded trained model to {upload_url}")
        return

    fields = {
        "name": name,
        "description": description,
        "base_annotations": base_annotations,
        "base_model_id": base_model_id,
    }
    body, content_type = encode_multipart(fields, "file", model_path)
    headers = {"Content-Type": content_type}
    if upload_token:
        headers["Authorization"] = f"Bearer {upload_token}"

    request = urllib.request.Request(upload_url, data=body, headers=headers, method="POST")
    try:
        with urllib.request.urlopen(request, timeout=300) as response:
            response_body = response.read().decode("utf-8", errors="replace")
            log(f"Uploaded trained model to {upload_url}: HTTP {response.status}")
            if response_body:
                log(f"Import response: {response_body}")
    except urllib.error.HTTPError as exc:
        response_body = exc.read().decode("utf-8", errors="replace")
        raise RuntimeError(f"Model upload failed with HTTP {exc.code}: {response_body}") from exc


# ---------- Main ----------

def main() -> int:
    parser = argparse.ArgumentParser(description="Train YOLO model")
    parser.add_argument("--dataset-url")
    parser.add_argument("--dataset-zip-path")
    parser.add_argument("--dataset-name")
    parser.add_argument("--model-url", default=DEFAULT_MODEL_URL)
    parser.add_argument("--model-path", default=DEFAULT_MODEL_PATH)
    parser.add_argument("--project", default=DEFAULT_PROJECT)
    parser.add_argument("--name")
    parser.add_argument("--output-dir", default="")
    parser.add_argument("--work-dir", default=".")
    parser.add_argument("--epochs", type=int, default=50)
    parser.add_argument("--imgsz", type=int, default=640)
    parser.add_argument("--batch", type=int, default=16)
    parser.add_argument("--workers", type=int, default=2)
    parser.add_argument("--model-upload-url", default="")
    parser.add_argument("--model-upload-token", default="")
    parser.add_argument("--model-name", default="")
    parser.add_argument("--model-description", default="")
    parser.add_argument("--model-base-annotations", default="")
    parser.add_argument("--model-base-model-id", default="")
    parser.add_argument("--dry-run", action="store_true", help="Skip downloads and training")

    args = parser.parse_args()

    os.environ["TORCHDYNAMO_DISABLE"] = "1"

    dataset_name = args.dataset_name or (
        Path(args.dataset_zip_path).stem if args.dataset_zip_path else default_dataset_name_from_url(args.dataset_url)
    )
    run_name = args.name or f"{dataset_name}_finetune"

    work_dir = Path(args.work_dir)
    dataset_zip = work_dir / f"{dataset_name}.zip"
    dataset_dir = work_dir / dataset_name
    data_yaml = dataset_dir / "data.yaml"
    model_path = Path(args.model_path)

    log(f"Dry run: {args.dry_run}")

    # Dataset handling
    copy_or_download_dataset(args.dataset_url, args.dataset_zip_path, dataset_zip, args.dry_run)
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


    # Model download (use absolute path so fine-tuning always loads this file)
    if args.model_path and Path(args.model_path).exists():
        log(f"Using existing model path: {model_path}")
    else:
        download_if_missing(args.model_url, model_path, args.dry_run)
    model_path = model_path.resolve()
    log(f"Using pretrained weights: {model_path}")

    best_pt = None

    # Training (fine-tune: load .pt then train; YOLO keeps weights by default)
    if args.dry_run:
        log("[DRY-RUN] Skipping training")
    else:
        if fixed_yaml is None:
            raise RuntimeError("fixed_yaml is None, cannot train without data.yaml")
        log("Starting YOLO fine-tuning from pretrained model")
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
        if args.output_dir:
            output_dir = Path(args.output_dir)
            output_dir.mkdir(parents=True, exist_ok=True)
            copied_best = output_dir / "yolo_model_best.pt"
            shutil.copyfile(best_pt, copied_best)
            best_pt = copied_best

        if args.model_upload_url:
            upload_model(
                args.model_upload_url,
                args.model_upload_token,
                best_pt,
                args.model_name or run_name,
                args.model_description,
                args.model_base_annotations,
                args.model_base_model_id,
            )

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
