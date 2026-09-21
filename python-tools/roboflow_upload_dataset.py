import hashlib
import os
from urllib.parse import urlsplit

import requests
import roboflow

api_key = os.environ["ROBOFLOW_API_KEY"]
workspace_id = os.environ["ROBOFLOW_WORKSPACE_ID"]
dataset_path = os.environ["ROBOFLOW_DATASET_PATH"]
project_id = os.environ["ROBOFLOW_PROJECT_ID"]
is_not_ground_truth = os.environ.get("ROBOFLOW_IS_NOT_GROUND_TRUTH", "False") == "True"


def redact_secret(value, secret):
    return value.replace(secret, "<redacted>") if secret else value


def logged_post(url, *args, **kwargs):
    """Expose the auth response that the Roboflow SDK otherwise hides."""
    parsed_url = urlsplit(str(url))
    safe_url = f"{parsed_url.scheme}://{parsed_url.netloc}{parsed_url.path}"
    print(f"Roboflow auth request: POST {safe_url}", flush=True)

    try:
        response = original_post(url, *args, **kwargs)
    except Exception as exc:
        print(
            f"Roboflow auth request failed: {type(exc).__name__}: "
            f"{redact_secret(str(exc), api_key)}",
            flush=True,
        )
        raise

    response_body = redact_secret(response.text[:1000], api_key)
    print(f"Roboflow auth response status: {response.status_code}", flush=True)
    print(f"Roboflow auth response body: {response_body}", flush=True)

    return response


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
finally:
    requests.post = original_post

workspace = rf.workspace(workspace_id)

workspace.upload_dataset(
    dataset_path,
    project_id,
    num_workers=10,
    project_license="MIT",
    project_type="object-detection",
    batch_name=None,
    num_retries=0,
    is_prediction=is_not_ground_truth
)
