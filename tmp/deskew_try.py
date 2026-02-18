#!/usr/bin/env python3
"""
deskew_try.py

Try multiple deskew strategies on an input path (file or directory) and write
outputs into subfolders under --out.

Methods included:
  - opencv_hough      (no external CLI, uses OpenCV)
  - opencv_projection (no external CLI, uses OpenCV; projection-score search)
  - imagick           (optional, uses `magick` or `convert -deskew`)

Install (recommended):
  pip install opencv-python numpy

Optional (better projection method):
  pip install scikit-image

Usage:
  python deskew_try.py /path/to/pngs --out /path/to/out
  python deskew_try.py page.png --out out --methods opencv_hough imagick
"""

from __future__ import annotations

import argparse
import math
import os
import shutil
import subprocess
from dataclasses import dataclass
from pathlib import Path
from typing import Iterable, List, Optional, Tuple

import numpy as np

try:
    import cv2  # type: ignore
except ImportError as e:
    raise SystemExit("Missing dependency: opencv-python. Install with: pip install opencv-python") from e

try:
    from skimage.transform import radon  # type: ignore
    HAVE_SKIMAGE = True
except Exception:
    HAVE_SKIMAGE = False


SUPPORTED_EXTS = {".png", ".jpg", ".jpeg", ".tif", ".tiff"}


@dataclass
class Params:
    downscale_max: int = 1600         # max width/height for angle estimation
    trim_border: bool = True          # crop to content bbox before angle estimation
    angle_limit: float = 6.0          # degrees, search range [-limit, +limit]
    angle_step: float = 0.25          # degrees
    min_rotate: float = 0.15          # degrees, skip tiny corrections
    bg: int = 255                     # background fill (white)


def list_images(p: Path) -> List[Path]:
    if p.is_file():
        return [p]
    out: List[Path] = []
    for f in sorted(p.iterdir()):
        if f.is_file() and f.suffix.lower() in SUPPORTED_EXTS:
            out.append(f)
    return out


def read_image(path: Path) -> np.ndarray:
    img = cv2.imdecode(np.fromfile(str(path), dtype=np.uint8), cv2.IMREAD_COLOR)
    if img is None:
        raise ValueError(f"Failed to read image: {path}")
    return img


def write_image(path: Path, img: np.ndarray) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    ext = path.suffix.lower()
    if ext not in {".png", ".jpg", ".jpeg", ".tif", ".tiff"}:
        ext = ".png"
        path = path.with_suffix(ext)
    ok, buf = cv2.imencode(ext, img)
    if not ok:
        raise ValueError(f"Failed to encode image for: {path}")
    buf.tofile(str(path))


def resize_for_estimation(img: np.ndarray, max_side: int) -> np.ndarray:
    h, w = img.shape[:2]
    m = max(h, w)
    if m <= max_side:
        return img
    scale = max_side / float(m)
    new_w = max(1, int(round(w * scale)))
    new_h = max(1, int(round(h * scale)))
    return cv2.resize(img, (new_w, new_h), interpolation=cv2.INTER_AREA)


def binarize(gray: np.ndarray) -> np.ndarray:
    # robust-ish binarization for dirty scans
    gray = cv2.GaussianBlur(gray, (3, 3), 0)
    thr = cv2.adaptiveThreshold(
        gray, 255,
        adaptiveMethod=cv2.ADAPTIVE_THRESH_GAUSSIAN_C,
        thresholdType=cv2.THRESH_BINARY,
        blockSize=41,
        C=15,
    )
    # invert: ink=1 (white) background=0 (black) sometimes easier, but keep as 0/255 here
    return thr


def content_bbox(binary: np.ndarray) -> Optional[Tuple[int, int, int, int]]:
    # expects 0/255 with text mostly 0 on white background (after adaptiveThreshold it's 0/255)
    # invert to get "ink" as 1
    inv = 255 - binary
    # remove tiny speckle
    k = cv2.getStructuringElement(cv2.MORPH_RECT, (3, 3))
    inv = cv2.morphologyEx(inv, cv2.MORPH_OPEN, k, iterations=1)
    # find largest connected component-ish area via contours
    contours, _ = cv2.findContours(inv, cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE)
    if not contours:
        return None
    # merge bounding boxes of large contours
    areas = [(cv2.contourArea(c), c) for c in contours]
    areas.sort(reverse=True, key=lambda x: x[0])

    h, w = binary.shape[:2]
    min_area = (h * w) * 0.001  # 0.1% of page
    xs, ys, xe, ye = w, h, 0, 0
    kept = 0
    for a, c in areas[:50]:
        if a < min_area:
            break
        x, y, ww, hh = cv2.boundingRect(c)
        xs = min(xs, x)
        ys = min(ys, y)
        xe = max(xe, x + ww)
        ye = max(ye, y + hh)
        kept += 1

    if kept == 0:
        return None

    # pad a bit
    pad = int(round(0.01 * max(h, w)))
    xs = max(0, xs - pad)
    ys = max(0, ys - pad)
    xe = min(w, xe + pad)
    ye = min(h, ye + pad)

    if xe - xs < 50 or ye - ys < 50:
        return None
    return xs, ys, xe, ye


def rotate_fullres(img: np.ndarray, angle_deg: float, bg: int) -> np.ndarray:
    h, w = img.shape[:2]
    center = (w / 2.0, h / 2.0)
    M = cv2.getRotationMatrix2D(center, angle_deg, 1.0)
    # compute new bounds to keep everything
    cos = abs(M[0, 0])
    sin = abs(M[0, 1])
    new_w = int(round((h * sin) + (w * cos)))
    new_h = int(round((h * cos) + (w * sin)))
    M[0, 2] += (new_w / 2) - center[0]
    M[1, 2] += (new_h / 2) - center[1]
    return cv2.warpAffine(img, M, (new_w, new_h), flags=cv2.INTER_CUBIC, borderValue=(bg, bg, bg))


def estimate_angle_hough(gray: np.ndarray) -> float:
    # pipeline: binarize -> morph close horizontally -> edges -> hough -> median angle
    b = binarize(gray)
    inv = 255 - b

    # emphasize textlines, suppress ornaments/diagrams a bit
    kernel = cv2.getStructuringElement(cv2.MORPH_RECT, (60, 1))
    inv = cv2.morphologyEx(inv, cv2.MORPH_CLOSE, kernel, iterations=1)

    edges = cv2.Canny(inv, 50, 150, apertureSize=3)
    lines = cv2.HoughLinesP(edges, 1, np.pi / 180.0, threshold=120, minLineLength=80, maxLineGap=10)
    if lines is None:
        return 0.0

    angles: List[float] = []
    for (x1, y1, x2, y2) in lines[:, 0]:
        dx = x2 - x1
        dy = y2 - y1
        if dx == 0:
            continue
        ang = math.degrees(math.atan2(dy, dx))
        # keep near horizontal
        if -30.0 <= ang <= 30.0:
            angles.append(ang)

    if not angles:
        return 0.0
    return float(np.median(np.array(angles, dtype=np.float32)))


def projection_score(binary_inv: np.ndarray) -> float:
    # binary_inv: ink ~255, bg ~0. Score: variance of row sums (more peaky == better alignment)
    rows = np.sum(binary_inv > 0, axis=1).astype(np.float32)
    return float(np.var(rows))


def estimate_angle_projection(gray: np.ndarray, limit: float, step: float) -> float:
    b = binarize(gray)
    inv = 255 - b

    best_a = 0.0
    best_s = -1.0
    angles = np.arange(-limit, limit + 1e-9, step, dtype=np.float32)
    for a in angles:
        rot = rotate_fullres(inv, float(a), bg=0)  # bg=0 because inv has ink as 255
        s = projection_score(rot)
        if s > best_s:
            best_s = s
            best_a = float(a)
    return best_a


def estimate_angle_radon(gray: np.ndarray, limit: float, step: float) -> float:
    if not HAVE_SKIMAGE:
        return 0.0
    b = binarize(gray)
    inv = (255 - b).astype(np.float32) / 255.0  # ink=1.0
    # radon expects 2D float
    theta = np.arange(90 - limit, 90 + limit + 1e-9, step, dtype=np.float32)
    sinogram = radon(inv, theta=theta, circle=False)
    # maximize energy: sharp peaks when aligned
    scores = np.var(sinogram, axis=0)
    best = int(np.argmax(scores))
    # convert back to "skew angle" relative to horizontal
    # theta around 90 means vertical projection; offset from 90 corresponds to skew
    best_theta = float(theta[best])
    return best_theta - 90.0


def run_imagick(in_path: Path, out_path: Path) -> bool:
    # Try `magick` first (IM7), then `convert` (IM6)
    magick = shutil.which("magick")
    convert = shutil.which("convert")
    if magick:
        cmd = [magick, str(in_path), "-trim", "+repage", "-colorspace", "Gray", "-deskew", "40%", str(out_path)]
    elif convert:
        cmd = [convert, str(in_path), "-trim", "+repage", "-colorspace", "Gray", "-deskew", "40%", str(out_path)]
    else:
        return False

    out_path.parent.mkdir(parents=True, exist_ok=True)
    subprocess.run(cmd, check=True, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    return True


def estimate_angle(img_small: np.ndarray, params: Params, method: str) -> float:
    gray = cv2.cvtColor(img_small, cv2.COLOR_BGR2GRAY)

    if params.trim_border:
        b = binarize(gray)
        bb = content_bbox(b)
        if bb is not None:
            xs, ys, xe, ye = bb
            gray = gray[ys:ye, xs:xe]

    if method == "opencv_hough":
        return estimate_angle_hough(gray)

    if method == "opencv_projection":
        return estimate_angle_projection(gray, params.angle_limit, params.angle_step)

    if method == "radon" and HAVE_SKIMAGE:
        return estimate_angle_radon(gray, params.angle_limit, params.angle_step)

    return 0.0


def process_one(img_path: Path, out_root: Path, methods: List[str], params: Params) -> None:
    img = read_image(img_path)
    img_small = resize_for_estimation(img, params.downscale_max)

    for m in methods:
        out_dir = out_root / m
        out_path = out_dir / img_path.name

        if m == "imagick":
            try:
                ok = run_imagick(img_path, out_path)
                if not ok:
                    # Skip silently if ImageMagick not present
                    continue
            except subprocess.CalledProcessError:
                continue
            continue

        angle = estimate_angle(img_small, params, m)
        if abs(angle) < params.min_rotate:
            # still write a copy for comparison
            write_image(out_path, img)
            continue

        rotated = rotate_fullres(img, angle, bg=params.bg)
        write_image(out_path, rotated)


def main() -> None:
    ap = argparse.ArgumentParser()
    ap.add_argument("path", type=str, help="Input image file or directory")
    ap.add_argument("--out", type=str, required=True, help="Output directory")
    ap.add_argument(
        "--methods",
        nargs="+",
        default=["opencv_hough", "opencv_projection", "radon", "imagick"],
        choices=["opencv_hough", "opencv_projection", "radon", "imagick"],
        help="Methods to try",
    )
    ap.add_argument("--no-trim", action="store_true", help="Disable content trimming before angle estimation")
    ap.add_argument("--downscale-max", type=int, default=1600, help="Max side for angle estimation")
    ap.add_argument("--angle-limit", type=float, default=6.0, help="Angle search limit in degrees (projection/radon)")
    ap.add_argument("--angle-step", type=float, default=0.25, help="Angle step in degrees (projection/radon)")
    ap.add_argument("--min-rotate", type=float, default=0.15, help="Skip rotation if abs(angle) smaller than this")
    args = ap.parse_args()

    in_path = Path(args.path)
    out_root = Path(args.out)
    out_root.mkdir(parents=True, exist_ok=True)

    params = Params(
        downscale_max=args.downscale_max,
        trim_border=not args.no_trim,
        angle_limit=args.angle_limit,
        angle_step=args.angle_step,
        min_rotate=args.min_rotate,
    )

    imgs = list_images(in_path)
    if not imgs:
        raise SystemExit(f"No images found at: {in_path}")

    if "radon" in args.methods and not HAVE_SKIMAGE:
        print("Note: method 'radon' requested but scikit-image not installed. Install with: pip install scikit-image")

    for p in imgs:
        process_one(p, out_root, args.methods, params)

    print(f"Done. Outputs under: {out_root}")


if __name__ == "__main__":
    main()
