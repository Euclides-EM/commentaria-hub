#!/usr/bin/env python3
from __future__ import annotations

import argparse
import os
from concurrent.futures import ProcessPoolExecutor
from pathlib import Path

from PIL import Image


IMAGE_EXTENSIONS = {".png", ".webp", ".tif", ".tiff", ".gif", ".jpeg", ".jpg"}
DEFAULT_DIRECTORY = (Path(__file__).resolve().parent / "../ocrflow/store/data/tps/imgs").resolve()


def has_alpha(mode: str, image: Image.Image) -> bool:
    if "A" in mode:
        return True
    return mode == "P" and "transparency" in image.info


def alpha_bbox(image_path: Path) -> tuple[str, bool, tuple[int, int, int, int] | None, tuple[int, int]]:
    with Image.open(image_path) as image:
        if not has_alpha(image.mode, image):
            return (image_path.name, False, None, image.size)

        alpha = image.getchannel("A") if "A" in image.mode else image.convert("RGBA").getchannel("A")
        bbox = alpha.getbbox()
        return (image_path.name, True, bbox, image.size)


def find_transparent_border(image_path: Path) -> tuple[str, bool] | None:
    name, has_alpha_channel, bbox, size = alpha_bbox(image_path)
    if not has_alpha_channel or bbox is None:
        return None

    full_bbox = (0, 0, size[0], size[1])
    if bbox != full_bbox:
        return (name, True)
    return None


def iter_images(directory: Path) -> list[Path]:
    return sorted(
        path
        for path in directory.iterdir()
        if path.is_file() and path.suffix.lower() in IMAGE_EXTENSIONS
    )


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Find images whose opaque content is surrounded by transparent padding."
    )
    parser.add_argument(
        "directory",
        nargs="?",
        default=str(DEFAULT_DIRECTORY),
        help="Directory to scan. Defaults to the TPS image directory relative to this script.",
    )
    parser.add_argument(
        "--workers",
        type=int,
        default=os.cpu_count() or 4,
        help="Number of worker processes to use. Defaults to CPU count.",
    )
    return parser.parse_args()


def main() -> int:
    args = parse_args()
    directory = Path(args.directory).resolve()
    image_paths = iter_images(directory)

    if not image_paths:
        print("No images found.")
        return 0

    with ProcessPoolExecutor(max_workers=args.workers) as executor:
        results = executor.map(find_transparent_border, image_paths, chunksize=16)
        matches = [result[0] for result in results if result is not None]

    for match in matches:
        print(match)

    print(f"\nFound {len(matches)} image(s) with transparent border padding out of {len(image_paths)} scanned.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
