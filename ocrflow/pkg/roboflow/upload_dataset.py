import roboflow

rf = roboflow.Roboflow(api_key="API_KEY")

workspace = rf.workspace("WORKSPACE_ID")

workspace.upload_dataset(
    "DATASET_PATH",
    "PROJECT_ID",
    num_workers=10,
    project_license="MIT",
    project_type="object-detection",
    batch_name=None,
    num_retries=0,
    is_prediction=IS_NOT_GROUND_TRUTH
)