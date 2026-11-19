import hashlib
import math
import os
from pathlib import Path
from urllib.parse import urlsplit

import requests
import roboflow
from PIL import Image

api_key = os.environ["ROBOFLOW_API_KEY"]
workspace_id = os.environ["ROBOFLOW_WORKSPACE_ID"]
dataset_path = os.environ["ROBOFLOW_DATASET_PATH"]
project_id = os.environ["ROBOFLOW_PROJECT_ID"]
is_not_ground_truth = os.environ.get("ROBOFLOW_IS_NOT_GROUND_TRUTH", "False") == "True"
max_upload_pixels = int(os.environ.get("ROBOFLOW_MAX_UPLOAD_PIXELS", "80000000"))


def redact_secret(value, secret):
    return value.replace(secret, "<redacted>") if secret else value


def response_body_for_log(response):
    body = redact_secret(response.text, api_key)
    if not body.strip():
        return "<empty response body>"
    return body


def logged_post(url, *args, **kwargs):
    """Expose 5xx bodies that the Roboflow SDK otherwise hides."""
    parsed_url = urlsplit(str(url))
    safe_url = f"{parsed_url.scheme}://{parsed_url.netloc}{parsed_url.path}"

    try:
        response = original_post(url, *args, **kwargs)
    except Exception as exc:
        print(
            f"Roboflow POST request failed: {type(exc).__name__}: "
            f"{redact_secret(str(exc), api_key)}",
            flush=True,
        )
        raise

    if response.status_code >= 500:
        print(
            f"Roboflow HTTP error: POST {safe_url} returned "
            f"{response.status_code}; response body: {response_body_for_log(response)}",
            flush=True,
        )

    return response


def upload_images(dataset_dir):
    images_dir = Path(dataset_dir) / "images"
    if not images_dir.is_dir():
        return []
    return [
        path
        for path in images_dir.iterdir()
        if path.is_file() and path.suffix.lower() in {".jpg", ".jpeg", ".png"}
    ]


def resize_oversized_images(dataset_dir, pixel_limit):
    if pixel_limit <= 0:
        raise ValueError("ROBOFLOW_MAX_UPLOAD_PIXELS must be positive")

    # These are trusted, locally generated YOLO images. Disable Pillow's lower
    # decompression-bomb threshold so that they can be opened and reduced.
    Image.MAX_IMAGE_PIXELS = None
    for image_path in upload_images(dataset_dir):
        with Image.open(image_path) as source:
            width, height = source.size
            pixels = width * height
            if pixels <= pixel_limit:
                continue

            scale = math.sqrt(pixel_limit / pixels)
            resized_width = max(1, math.floor(width * scale))
            resized_height = max(1, math.floor(height * scale))
            while resized_width * resized_height > pixel_limit:
                if resized_width >= resized_height:
                    resized_width -= 1
                else:
                    resized_height -= 1

            resized = source.resize(
                (resized_width, resized_height), Image.Resampling.LANCZOS
            )
            image_format = source.format or (
                "PNG" if image_path.suffix.lower() == ".png" else "JPEG"
            )
            if image_format.upper() == "JPEG" and resized.mode not in ("RGB", "L"):
                resized = resized.convert("RGB")

            temporary_path = image_path.with_name(image_path.name + ".resizing")
            save_options = {"quality": 90} if image_format.upper() == "JPEG" else {}
            try:
                resized.save(temporary_path, format=image_format, **save_options)
                os.replace(temporary_path, image_path)
            finally:
                if temporary_path.exists():
                    temporary_path.unlink()

            print(
                f"Reduced Roboflow upload image {image_path}: "
                f"{width}x{height} -> {resized_width}x{resized_height}. "
                "YOLO labels remain unchanged because their coordinates are normalized.",
                flush=True,
            )


def main():
    global original_post

    key_fingerprint = hashlib.sha256(api_key.encode()).hexdigest()[:12]
    print(
        "Starting Roboflow upload: "
        f"workspace={workspace_id!r}, project={project_id!r}, "
        f"api_key_length={len(api_key)}, api_key_sha256={key_fingerprint}",
        flush=True,
    )

    original_post = requests.post
    requests.post = logged_post
    try:
        rf = roboflow.Roboflow(api_key=api_key)
        resize_oversized_images(dataset_path, max_upload_pixels)
        workspace = rf.workspace(workspace_id)
        workspace.upload_dataset(
            dataset_path,
            project_id,
            num_workers=10,
            project_license="MIT",
            project_type="object-detection",
            batch_name=None,
            num_retries=0,
            is_prediction=is_not_ground_truth,
        )
    finally:
        requests.post = original_post


if __name__ == "__main__":
    main()
