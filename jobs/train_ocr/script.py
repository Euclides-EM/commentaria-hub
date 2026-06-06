#!/usr/bin/env python3

from __future__ import annotations

import argparse
import mimetypes
import os
import shutil
import shlex
import subprocess
import sys
import time
import urllib.error
import urllib.request
import uuid
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


def find_latest_model(output_dir: Path) -> Path:
    models = sorted(output_dir.glob("*.mlmodel"), key=lambda p: p.stat().st_mtime, reverse=True)
    if not models:
        raise RuntimeError(f"No .mlmodel output found in {output_dir}")
    return models[0]


def encode_multipart(fields: dict[str, str], file_field: str, file_path: Path) -> tuple[bytes, str]:
    boundary = "----commentaria-ocr-" + uuid.uuid4().hex
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
            detail = "\n".join(
                part for part in [result.stdout.strip(), result.stderr.strip()] if part
            )
            raise RuntimeError(f"Model upload failed with curl exit {result.returncode}: {detail}")
        if result.stdout.strip():
            log(f"Import response: {result.stdout.strip()}")
        log(f"Uploaded trained model to {upload_url}")
        return

    log("curl not found; falling back to Python urllib upload")
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
        default="workspace",
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
        "--epochs",
        type=int,
        default=int(os.environ.get("TRAIN_EPOCHS", "0") or "0"),
        help="Maximum number of training epochs. Leave unset or 0 to use Kraken defaults.",
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
    parser.add_argument(
        "--model-upload-url",
        default=os.environ.get("MODEL_UPLOAD_URL", ""),
        help="Optional Commentaria /models_upload endpoint to import the trained model after training",
    )
    parser.add_argument(
        "--model-upload-token",
        default=os.environ.get("MODEL_UPLOAD_TOKEN", ""),
        help="Optional bearer token for the model upload endpoint",
    )
    parser.add_argument(
        "--model-name",
        default=os.environ.get("MODEL_NAME", ""),
        help="Name for the imported model",
    )
    parser.add_argument(
        "--model-description",
        default=os.environ.get("MODEL_DESCRIPTION", ""),
        help="Description for the imported model",
    )
    parser.add_argument(
        "--model-base-annotations",
        default=os.environ.get("MODEL_BASE_ANNOTATIONS", ""),
        help="Comma-separated <dataset_id>:<annotation_id> references for the imported model",
    )
    parser.add_argument(
        "--model-base-model-id",
        default=os.environ.get("MODEL_BASE_MODEL_ID", ""),
        help="Base model ID for the imported model",
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
    log(f"Epochs         = {args.epochs if args.epochs > 0 else '(Kraken default)'}")
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
    if args.epochs > 0:
        train_cmd.extend(["--quit", "fixed", "-N", str(args.epochs)])

    if base_model_path:
        if not base_model_path.exists():
            raise FileNotFoundError(f"Base model not found: {base_model_path}")
        train_cmd.extend(["-i", str(base_model_path), "--resize", "add"])

    run(train_cmd)

    log("Training finished successfully")
    trained_model_path = find_latest_model(output_dir)
    log(f"Latest trained model: {trained_model_path}")

    if args.model_upload_url:
        log(f"Importing trained model via {args.model_upload_url}")
        upload_model(
            args.model_upload_url,
            args.model_upload_token,
            trained_model_path,
            args.model_name,
            args.model_description,
            args.model_base_annotations,
            args.model_base_model_id,
        )
    else:
        log("MODEL_UPLOAD_URL is not configured; skipping trained model import")

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
