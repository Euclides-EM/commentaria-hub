#!/usr/bin/env python3

from __future__ import annotations

import json
import mimetypes
import os
import shutil
import shlex
import subprocess
import sys
import time
import uuid
import zipfile
from pathlib import Path

from lxml import etree
from PIL import Image, ImageDraw


def env(name: str, default: str = "") -> str:
    return os.environ.get(name, default)


def env_list(name: str) -> list[str]:
    return [x.strip() for x in env(name).splitlines() if x.strip()]


def log(msg: str) -> None:
    print(f"[{time.strftime('%Y-%m-%d %H:%M:%S')}] {msg}", flush=True)


def run(cmd: list[str]) -> None:
    log("+ " + " ".join(shlex.quote(x) for x in cmd))
    subprocess.check_call(cmd)


def xpath(el: etree._ElementTree | etree._Element, expr: str) -> list[etree._Element]:
    return el.xpath(expr)


def local_name(el: etree._Element) -> str:
    return etree.QName(el).localname


def copy_alto(alto_dir: Path, output_dir: Path) -> None:
    output_dir.mkdir(parents=True, exist_ok=True)
    if alto_dir.exists():
        for p in alto_dir.glob("*.xml"):
            shutil.copyfile(p, output_dir / p.name)


def zip_dir(src_dir: Path, dst_zip: Path) -> None:
    dst_zip.parent.mkdir(parents=True, exist_ok=True)
    with zipfile.ZipFile(dst_zip, "w", zipfile.ZIP_DEFLATED) as zf:
        for p in sorted(src_dir.rglob("*")):
            if p.is_file():
                zf.write(p, p.relative_to(src_dir))


def encode_multipart(fields: dict[str, str], file_field: str, file_path: Path) -> tuple[bytes, str]:
    boundary = "----commentaria-detect-" + uuid.uuid4().hex
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


def upload_result(upload_url: str, upload_token: str, mode: str, zip_path: Path) -> None:
    if not upload_url:
        raise RuntimeError("RESULT_UPLOAD_URL is not configured")
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
            f"file=@{zip_path};type=application/zip",
            "-F",
            f"mode={mode}",
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
            raise RuntimeError(f"Detection result upload failed with curl exit {result.returncode}: {detail}")
        if result.stdout.strip():
            log(f"Upload response: {result.stdout.strip()}")
        return

    import urllib.error
    import urllib.request

    body, content_type = encode_multipart({"mode": mode}, "file", zip_path)
    headers = {"Content-Type": content_type}
    if upload_token:
        headers["Authorization"] = f"Bearer {upload_token}"
    request = urllib.request.Request(upload_url, data=body, headers=headers, method="POST")
    try:
        with urllib.request.urlopen(request, timeout=300) as response:
            log(f"Uploaded detection result: HTTP {response.status}")
    except urllib.error.HTTPError as exc:
        detail = exc.read().decode("utf-8", errors="replace")
        raise RuntimeError(f"Detection result upload failed with HTTP {exc.code}: {detail}") from exc


def tag_label_map(tree: etree._ElementTree) -> dict[str, str]:
    out: dict[str, str] = {}
    for tag in xpath(tree, "//*[local-name()='OtherTag']"):
        tag_id = tag.get("ID")
        label = tag.get("LABEL")
        if tag_id and label:
            out[tag_id] = label
    return out


def has_label(tagrefs: str | None, labels: set[str], id_to_label: dict[str, str]) -> bool:
    if not tagrefs:
        return False
    return any(id_to_label.get(ref) in labels for ref in tagrefs.split())


def page_size(tree: etree._ElementTree) -> tuple[int, int]:
    pages = xpath(tree, "//*[local-name()='Page']")
    if not pages:
        raise RuntimeError("ALTO has no Page element")
    page = pages[0]
    return int(float(page.get("WIDTH", "0"))), int(float(page.get("HEIGHT", "0")))


def create_mask(alto_path: Path, mask_path: Path, main_labels: list[str], ignore_labels: list[str]) -> bool:
    tree = etree.parse(str(alto_path))
    width, height = page_size(tree)
    id_to_label = tag_label_map(tree)
    main_set = set(main_labels)
    ignore_set = set(ignore_labels)
    img = Image.new("P", (width, height), color=0)
    img.putpalette([255, 255, 255, 0, 0, 0])
    draw = ImageDraw.Draw(img)

    has_regions = False
    for tb in xpath(tree, "//*[local-name()='TextBlock']"):
        if not has_label(tb.get("TAGREFS"), main_set, id_to_label):
            continue
        if has_label(tb.get("TAGREFS"), ignore_set, id_to_label):
            continue
        x = float(tb.get("HPOS", "0"))
        y = float(tb.get("VPOS", "0"))
        w = float(tb.get("WIDTH", "0"))
        h = float(tb.get("HEIGHT", "0"))
        if w <= 0 or h <= 0:
            continue
        draw.rectangle([x, y, x + w, y + h], fill=1)
        has_regions = True

    for tb in xpath(tree, "//*[local-name()='TextBlock']"):
        if not has_label(tb.get("TAGREFS"), ignore_set, id_to_label):
            continue
        x = float(tb.get("HPOS", "0"))
        y = float(tb.get("VPOS", "0"))
        w = float(tb.get("WIDTH", "0"))
        h = float(tb.get("HEIGHT", "0"))
        draw.rectangle([x, y, x + w, y + h], fill=0)

    mask_path.parent.mkdir(parents=True, exist_ok=True)
    img.save(mask_path)
    return has_regions


def delete_lines(alto_path: Path) -> etree._ElementTree:
    tree = etree.parse(str(alto_path))
    for line in xpath(tree, "//*[local-name()='TextLine']"):
        parent = line.getparent()
        if parent is not None:
            parent.remove(line)
    return tree


def bbox(points: list[list[float]]) -> tuple[float, float, float, float]:
    xs = [p[0] for p in points if len(p) == 2]
    ys = [p[1] for p in points if len(p) == 2]
    if not xs or not ys:
        return 0, 0, 0, 0
    return min(xs), min(ys), max(xs), max(ys)


def points_to_string(points: list[list[float]]) -> str:
    return " ".join(f"{round(p[0])} {round(p[1])}" for p in points if len(p) == 2)


def rect_points(min_x: float, min_y: float, max_x: float, max_y: float) -> str:
    vals = [min_x, min_y, max_x, min_y, max_x, max_y, min_x, max_y, min_x, min_y]
    return " ".join(str(round(v)) for v in vals)


def ensure_line_tag(tree: etree._ElementTree) -> None:
    tags_nodes = xpath(tree, "//*[local-name()='Tags']")
    if not tags_nodes:
        root = tree.getroot()
        tags = etree.Element("Tags")
        root.insert(0, tags)
    else:
        tags = tags_nodes[0]
    for tag in xpath(tags, ".//*[local-name()='OtherTag']"):
        if tag.get("ID") == "LT_default":
            return
    etree.SubElement(tags, "OtherTag", ID="LT_default", LABEL="DefaultLine")


def glue_lines(alto_path: Path, baselines_json: Path) -> None:
    if not baselines_json.exists() or baselines_json.stat().st_size == 0:
        return
    data = json.loads(baselines_json.read_text())
    tree = etree.parse(str(alto_path))
    ensure_line_tag(tree)
    blocks = xpath(tree, "//*[local-name()='TextBlock']")
    if not blocks:
        return
    for line in data.get("lines", []):
        points = line.get("boundary") or line.get("baseline") or []
        min_x, min_y, max_x, max_y = bbox(points)
        cx = (min_x + max_x) / 2
        cy = (min_y + max_y) / 2
        target = blocks[0]
        for block in blocks:
            x = float(block.get("HPOS", "0"))
            y = float(block.get("VPOS", "0"))
            w = float(block.get("WIDTH", "0"))
            h = float(block.get("HEIGHT", "0"))
            if x <= cx <= x + w and y <= cy <= y + h:
                target = block
                break
        tl = etree.SubElement(target, "TextLine")
        tl.set("ID", "eSc_line_" + uuid.uuid4().hex[:12])
        tl.set("TAGREFS", "LT_default")
        tl.set("BASELINE", points_to_string(line.get("baseline") or []))
        tl.set("HPOS", str(round(min_x)))
        tl.set("VPOS", str(round(min_y)))
        tl.set("WIDTH", str(round(max_x - min_x)))
        tl.set("HEIGHT", str(round(max_y - min_y)))
        tl.set("POINTS", rect_points(min_x, min_y, max_x, max_y))
    tree.write(str(alto_path), encoding="UTF-8", xml_declaration=True, pretty_print=True)


def detect_lines(image_dir: Path, alto_dir: Path, output_dir: Path) -> None:
    include = env_list("INCLUDE_CATEGORIES")
    ignore = env_list("IGNORE_CATEGORIES")
    copy_alto(alto_dir, output_dir)
    for alto_path in sorted(output_dir.glob("*.xml")):
        tree = delete_lines(alto_path)
        tree.write(str(alto_path), encoding="UTF-8", xml_declaration=True, pretty_print=True)
        img_path = image_dir / f"{alto_path.stem}.png"
        if not img_path.exists():
            raise FileNotFoundError(f"image not found for {alto_path.name}: {img_path}")
        categories = include or [""]
        for category in categories:
            effective_include = [category] if category else []
            effective_ignore = ignore + [c for c in include if c != category]
            mask_path = output_dir / f"{alto_path.stem}-{category or 'all'}-mask.png"
            if not create_mask(alto_path, mask_path, effective_include, effective_ignore):
                continue
            json_path = output_dir / f"{alto_path.stem}-{category or 'all'}-baselines.json"
            run(["kraken", "-d", "cuda:0", "-i", str(img_path), str(json_path), "segment", "-bl", "--mask", str(mask_path), "--pad", "2", "2"])
            glue_lines(alto_path, json_path)


def model_segment(image_dir: Path, output_dir: Path, model_path: Path) -> None:
    output_dir.mkdir(parents=True, exist_ok=True)
    pairs: list[str] = []
    for img in sorted(image_dir.glob("*.png")):
        pairs.extend(["-i", str(img), str(output_dir / f"{img.stem}.xml")])
    if pairs:
        run(["yaltai", "kraken", "--alto", "-d", "cuda:0", *pairs, "segment", "--yolo", str(model_path)])


def model_ocr(image_dir: Path, alto_dir: Path, output_dir: Path, model_path: Path) -> None:
    copy_alto(alto_dir, output_dir)
    pairs: list[str] = []
    for alto_path in sorted(output_dir.glob("*.xml")):
        img_path = image_dir / f"{alto_path.stem}.png"
        pairs.extend(["-i", str(img_path), str(alto_path) + ".ocr.tmp"])
    if pairs:
        run(["kraken", "--alto", "-d", "cuda:0", *pairs, "segment", "-t", "alto", "ocr", "-m", str(model_path)])
        for tmp in output_dir.glob("*.xml.ocr.tmp"):
            tmp.replace(output_dir / tmp.name.removesuffix(".ocr.tmp"))


def main() -> int:
    mode = env("MODE")
    image_dir = Path(env("IMAGE_DIR"))
    alto_dir = Path(env("ALTO_DIR"))
    output_dir = Path(env("OUTPUT_DIR"))
    artifacts_dir = Path(env("ARTIFACTS_DIR"))
    model_path = Path(env("MODEL_PATH")) if env("MODEL_PATH") else Path()
    try:
        if mode == "lines":
            detect_lines(image_dir, alto_dir, output_dir)
        elif mode == "model_segment":
            model_segment(image_dir, output_dir, model_path)
        elif mode == "model_ocr":
            model_ocr(image_dir, alto_dir, output_dir, model_path)
        else:
            raise ValueError(f"unsupported MODE: {mode}")
        result_zip = artifacts_dir / "alto-result.zip"
        zip_dir(output_dir, result_zip)
        upload_result(env("RESULT_UPLOAD_URL"), env("RESULT_UPLOAD_TOKEN"), mode, result_zip)
        return 0
    except Exception as exc:
        artifacts_dir.mkdir(parents=True, exist_ok=True)
        (artifacts_dir / "error.log").write_text(str(exc), encoding="utf-8")
        log(f"ERROR: {exc}")
        return 1


if __name__ == "__main__":
    sys.exit(main())
