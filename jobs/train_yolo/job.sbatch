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

cd "$SLURM_SUBMIT_DIR"
mkdir -p logs artifacts

echo "host=$(hostname)"
echo "pwd=$(pwd)"
echo "CUDA_VISIBLE_DEVICES=${CUDA_VISIBLE_DEVICES:-}"
nvidia-smi || true

# ---- generate short unique id ----
RUN_ID="$(date +%y%m%d-%H%M)-$RANDOM"

PROJECT="runs_yolo"
RUN_NAME="run-${RUN_ID}"
DATASET_NAME="dataset-${RUN_ID}"

echo "RUN_ID=${RUN_ID}"
echo "PROJECT=${PROJECT}"
echo "RUN_NAME=${RUN_NAME}"
echo "DATASET_NAME=${DATASET_NAME}"

source .venv/bin/activate

python train_yolo.py \
  --dataset-url "https://app.roboflow.com/ds/06eW0gCNoL?key=HbRQg7ERhH" \
  --dataset-name "${DATASET_NAME}" \
  --project "${PROJECT}" \
  --name "${RUN_NAME}" \
  --epochs 50 \
  --imgsz 640 \
  --batch 16 \
  --workers 2
