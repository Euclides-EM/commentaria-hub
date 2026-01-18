import os
import roboflow

api_key = os.environ["ROBOFLOW_API_KEY"]
workspace_id = os.environ["ROBOFLOW_WORKSPACE_ID"]
dataset_path = os.environ["ROBOFLOW_DATASET_PATH"]
project_id = os.environ["ROBOFLOW_PROJECT_ID"]
is_not_ground_truth = os.environ.get("ROBOFLOW_IS_NOT_GROUND_TRUTH", "False") == "True"

rf = roboflow.Roboflow(api_key=api_key)

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