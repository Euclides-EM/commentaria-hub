#!/bin/bash
#SBATCH --job-name=kraken_train
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

PROJECT_ROOT="${PROJECT_ROOT:-/pbs/home/m/mjoskowicz/jobs/train_ocr}"

cd "$SLURM_SUBMIT_DIR"

if [[ ! -f manifest.env ]]; then
  echo "Missing manifest.env in ${SLURM_SUBMIT_DIR}" >&2
  exit 1
fi

source manifest.env

mkdir -p "${LOGS_DIR}" "${OUTPUT_DIR}"

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

source "${PROJECT_ROOT}/.venv/bin/activate"

python -c "import torch; print('torch', torch.__version__, 'cuda', torch.version.cuda, 'available', torch.cuda.is_available())"

echo "WORK_DIR=${WORK_DIR}"
echo "OUTPUT_DIR=${OUTPUT_DIR}"
echo "MODEL_PREFIX=${MODEL_PREFIX}"
echo "BASE_MODEL_PATH=${BASE_MODEL_PATH}"
echo "MODEL_UPLOAD_URL=${MODEL_UPLOAD_URL:-}"
echo "ZIP_PATHS=${ZIP_PATHS[*]}"

if [[ -n "${BASE_MODEL_PATH}" && ! -f "${BASE_MODEL_PATH}" ]]; then
  echo "Base model not found: ${BASE_MODEL_PATH}" >&2
  exit 1
fi

for z in "${ZIP_PATHS[@]}"; do
  if [[ ! -f "${z}" ]]; then
    echo "ZIP not found: ${z}" >&2
    exit 1
  fi
done

TRAIN_ARGS=()
for z in "${ZIP_PATHS[@]}"; do
  TRAIN_ARGS+=(--zip-path "$z")
done

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

TRAIN_ARGS+=(--base-model-path "${BASE_MODEL_PATH}")
TRAIN_ARGS+=(--output-dir "${OUTPUT_DIR}")
TRAIN_ARGS+=(--work-dir "${WORK_DIR}")
TRAIN_ARGS+=(--model-file-prefix "${MODEL_PREFIX}")
TRAIN_ARGS+=(--unicode-norm "NFD")
TRAIN_ARGS+=(--batch-size 2)
TRAIN_ARGS+=(--learning-rate 1e-4)
TRAIN_ARGS+=(--device "cuda:0")
if [[ -n "${TRAIN_EPOCHS:-}" ]]; then
  TRAIN_ARGS+=(--epochs "$TRAIN_EPOCHS")
fi

python "${PROJECT_ROOT}/script.py" "${TRAIN_ARGS[@]}"
