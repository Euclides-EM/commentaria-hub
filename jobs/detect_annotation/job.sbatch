#!/bin/bash
#SBATCH --job-name=annotation_detect
#SBATCH --partition=gpu_v100
#SBATCH --gres=gpu:v100:1
#SBATCH --cpus-per-task=4
#SBATCH --mem=24G
#SBATCH --time=12:00:00
#SBATCH --output=logs/%x_%j.out
#SBATCH --error=logs/%x_%j.err

set -euo pipefail

if command -v module >/dev/null 2>&1; then
  module purge || true
fi
unset PYTHONHOME PYTHONPATH LD_LIBRARY_PATH

PROJECT_ROOT="${PROJECT_ROOT:-/pbs/home/m/mjoskowicz/jobs/detect_annotation}"

cd "$SLURM_SUBMIT_DIR"

if [[ ! -f manifest.env ]]; then
  echo "Missing manifest.env in ${SLURM_SUBMIT_DIR}" >&2
  exit 1
fi

source manifest.env

mkdir -p "${LOGS_DIR}" "${OUTPUT_DIR}" "${ARTIFACTS_DIR}"

echo "host=$(hostname)"
echo "pwd=$(pwd)"
echo "PROJECT_ROOT=${PROJECT_ROOT}"
echo "RUN_DIR=${RUN_DIR}"
echo "MODE=${MODE}"
echo "RESULT_UPLOAD_URL=${RESULT_UPLOAD_URL:-}"
echo "CUDA_VISIBLE_DEVICES=${CUDA_VISIBLE_DEVICES:-}"
nvidia-smi || true

if [[ ! -d "${PROJECT_ROOT}/.venv" ]]; then
  echo "Missing virtualenv at ${PROJECT_ROOT}/.venv" >&2
  exit 1
fi

source "${PROJECT_ROOT}/.venv/bin/activate"

# Each page worker owns one CPU. Prevent PyTorch/OpenMP inside every worker
# from multiplying the four allocated Slurm CPUs into dozens of host threads.
export OMP_NUM_THREADS="${OMP_NUM_THREADS:-1}"
export MKL_NUM_THREADS="${MKL_NUM_THREADS:-1}"

python -c "import torch; print('torch', torch.__version__, 'cuda', torch.version.cuda, 'available', torch.cuda.is_available())"

set +e
python "${PROJECT_ROOT}/script.py"
status=$?
set -e

if [[ $status -eq 0 ]]; then
  touch "${ARTIFACTS_DIR}/done"
else
  echo "annotation detection failed with exit ${status}" > "${ARTIFACTS_DIR}/failed"
  exit "$status"
fi
