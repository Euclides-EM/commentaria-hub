#!/bin/bash
#SBATCH --job-name=crops_gen
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

DRIVE_REMOTE_NAME="drive"

curl https://euclides.huma-num.fr/commentaria/api/v1/facsimilies -o facsimilies.json

jq -r '.[] | select(.diagram_crops_available == false) | .edition_id' facsimilies.json > target_keys.txt

if [ ! -s target_keys.txt ]; then
    echo "No editions found without diagram crops. Exiting."
    exit 0
fi
echo "Found $(wc -l < target_keys.txt) editions without diagram crops."

rm -rf docs || true
mkdir -p docs

for key in $(cat target_keys.txt); do
    curl -fL \
      -H "Authorization: Bearer ${GITHUB_TOKEN}" \
      -o docs/${key}.pdf \
      https://euclides.huma-num.fr/commentaria/api/v1/editions/${key}/facsimile.pdf
done

[ -d "sphaera-cordeep" ] || git clone https://github.com/Euclides-EM/sphaera-cordeep.git
cd sphaera-cordeep
git pull
uv sync

PYTHONPATH=. uv run detect_batch.py ../docs

cd ../docs/diagrams

tar -czf commentaria-diagram-crops.tar.gz */

rclone copyto --drive-root-folder-id=1sx4altDCtTaM1XAIedY8k3yq9mFrJ6jM commentaria-diagram-crops.tar.gz ${DRIVE_REMOTE_NAME}:commentaria-diagram-crops.tar.gz

cd ../..
rm -rf docs
