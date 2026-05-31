#!/bin/bash
#SBATCH --job-name=yolo_train
#SBATCH --partition=gpu_v100
#SBATCH --gres=gpu:v100:1
#SBATCH --cpus-per-task=4
#SBATCH --mem=24G
#SBATCH --time=24:00:00
#SBATCH --output=logs/%x_%j.out
#SBATCH --error=logs/%x_%j.err

set -euo pipefail

if command -v module >/dev/null 2>&1; then
  module purge || true
fi
unset PYTHONHOME PYTHONPATH LD_LIBRARY_PATH

PROJECT_ROOT="${PROJECT_ROOT:-/pbs/home/m/mjoskowicz/jobs/train_yolo}"

cd "$SLURM_SUBMIT_DIR"

if [[ ! -f manifest.env ]]; then
  echo "Missing manifest.env in ${SLURM_SUBMIT_DIR}" >&2
  exit 1
fi

source manifest.env

mkdir -p "${LOGS_DIR}" "${OUTPUT_DIR}" artifacts

echo "host=$(hostname)"
echo "pwd=$(pwd)"
echo "PROJECT_ROOT=${PROJECT_ROOT}"
echo "RUN_DIR=${RUN_DIR}"
echo "CUDA_VISIBLE_DEVICES=${CUDA_VISIBLE_DEVICES:-}"
nvidia-smi || true

if [[ ! -d "${PROJECT_ROOT}/.venv" ]]; then
  echo "Missing virtualenv at ${PROJECT_ROOT}/.venv" >&2
  exit 1
fi

PROJECT="runs_yolo"
RUN_NAME="${RUN_ID}"
DATASET_NAME="dataset-${RUN_ID}"

echo "RUN_ID=${RUN_ID}"
echo "PROJECT=${PROJECT}"
echo "RUN_NAME=${RUN_NAME}"
echo "DATASET_NAME=${DATASET_NAME}"
echo "DATASET_ZIP_PATH=${DATASET_ZIP_PATH}"
echo "BASE_MODEL_PATH=${BASE_MODEL_PATH}"
echo "MODEL_UPLOAD_URL=${MODEL_UPLOAD_URL:-}"

if [[ ! -f "${DATASET_ZIP_PATH}" ]]; then
  echo "Dataset ZIP not found: ${DATASET_ZIP_PATH}" >&2
  exit 1
fi

if [[ -n "${BASE_MODEL_PATH}" && ! -f "${BASE_MODEL_PATH}" ]]; then
  echo "Base model not found: ${BASE_MODEL_PATH}" >&2
  exit 1
fi

source "${PROJECT_ROOT}/.venv/bin/activate"

python -c "import torch; print('torch', torch.__version__, 'cuda', torch.version.cuda, 'available', torch.cuda.is_available())"

TRAIN_ARGS=(
  --dataset-zip-path "${DATASET_ZIP_PATH}"
  --dataset-name "${DATASET_NAME}"
  --project "${RUN_DIR}/${PROJECT}"
  --name "${RUN_NAME}"
  --output-dir "${OUTPUT_DIR}"
  --work-dir "${WORK_DIR}"
  --imgsz 640
  --batch 16
  --workers 2
)

if [[ -n "${BASE_MODEL_PATH}" ]]; then
  TRAIN_ARGS+=(--model-path "${BASE_MODEL_PATH}")
fi

if [[ -n "${MODEL_UPLOAD_URL:-}" ]]; then
  TRAIN_ARGS+=(--model-upload-url "$MODEL_UPLOAD_URL")
  TRAIN_ARGS+=(--model-name "${MODEL_NAME:-}")
  TRAIN_ARGS+=(--model-description "${MODEL_DESCRIPTION:-}")
  TRAIN_ARGS+=(--model-base-annotations "${MODEL_BASE_ANNOTATIONS:-}")
  TRAIN_ARGS+=(--model-base-model-id "${MODEL_BASE_MODEL_ID:-}")
  if [[ -n "${MODEL_UPLOAD_TOKEN:-}" ]]; then
    TRAIN_ARGS+=(--model-upload-token "$MODEL_UPLOAD_TOKEN")
  fi
fi

if [[ -n "${TRAIN_EPOCHS:-}" ]]; then
  TRAIN_ARGS+=(--epochs "$TRAIN_EPOCHS")
else
  TRAIN_ARGS+=(--epochs 50)
fi

python "${PROJECT_ROOT}/script.py" "${TRAIN_ARGS[@]}"
