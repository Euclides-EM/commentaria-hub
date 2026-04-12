#!/usr/bin/env bash
set -euo pipefail

# Usage:
# ./submit_remote.sh \
#   ~/Downloads/export_doc7_paris_1598a_alto_20260411123937.zip \
#   ~/Downloads/export_doc7_paris_1598a_alto_20260411123937.zip \
#   ~/Downloads/export_doc31_paris_1615_manually_corrected_alto_202604111236.zip \
#   ~/Downloads/export_doc43_1598_manually_corrected_alto_202604111236.zip

REMOTE_HOST="mjoskowicz@cca.in2p3.fr"
REMOTE_ROOT="/pbs/home/m/mjoskowicz/jobs/train_ocr"
REMOTE_ASSETS_MODELS="${REMOTE_ROOT}/assets/models"
REMOTE_ASSETS_ZIPS="${REMOTE_ROOT}/assets/zips"
REMOTE_RUNS="${REMOTE_ROOT}/runs"

SSH_OPTS=(
  -o StrictHostKeyChecking=accept-new
)

ssh_remote() {
  ssh "${SSH_OPTS[@]}" "$@"
}

scp_remote() {
  scp "${SSH_OPTS[@]}" "$@"
}

log() {
  printf '[%s] %s\n' "$(date '+%Y-%m-%d %H:%M:%S')" "$*"
}

LOCAL_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${LOCAL_DIR}/../.." && pwd)"

LOCAL_SCRIPT_PY="${LOCAL_DIR}/script.py"
LOCAL_REQUIREMENTS="${LOCAL_DIR}/requirements.txt"
LOCAL_JOB_SBATCH="${LOCAL_DIR}/job.sbatch"
LOCAL_BASE_MODEL="${REPO_ROOT}/ocrflow/store/models/Gallicorpor.mlmodel"

if [[ "$#" -lt 1 ]]; then
  echo "Pass at least one ZIP file."
  exit 1
fi

ZIP_FILES=("$@")

RUN_ID="$(date +%y%m%d-%H%M)-$RANDOM"
RUN_NAME="kraken_train_${RUN_ID}"
REMOTE_RUN_DIR="${REMOTE_RUNS}/${RUN_NAME}"
REMOTE_MODEL_NAME="$(basename "${LOCAL_BASE_MODEL}")"
REMOTE_MODEL_PATH="${REMOTE_ASSETS_MODELS}/${REMOTE_MODEL_NAME}"

log "==> Preparing remote directories..."
ssh_remote "${REMOTE_HOST}" bash <<EOF
set -euo pipefail

PROJECT_ROOT=${REMOTE_ROOT}

mkdir -p "\${PROJECT_ROOT}" \
         "\${PROJECT_ROOT}/assets/models" \
         "\${PROJECT_ROOT}/assets/zips" \
         "\${PROJECT_ROOT}/runs" \
         "\${PROJECT_ROOT}/logs" \
         "\${PROJECT_ROOT}/artifacts"
EOF

log "==> Syncing job files..."
scp_remote "${LOCAL_SCRIPT_PY}" "${REMOTE_HOST}:${REMOTE_ROOT}/script.py"
scp_remote "${LOCAL_REQUIREMENTS}" "${REMOTE_HOST}:${REMOTE_ROOT}/requirements.txt"
scp_remote "${LOCAL_JOB_SBATCH}" "${REMOTE_HOST}:${REMOTE_ROOT}/job.sbatch"

log "==> Ensuring remote Python environment..."
ssh_remote "${REMOTE_HOST}" bash <<EOF
set -euo pipefail

PROJECT_ROOT=${REMOTE_ROOT}

if [[ ! -d "\${PROJECT_ROOT}/.venv" ]]; then
  log "Creating Python virtual environment..."
  python3 -m venv "\${PROJECT_ROOT}/.venv"
fi

source "\${PROJECT_ROOT}/.venv/bin/activate"
pip install -U pip wheel
pip install -r "\${PROJECT_ROOT}/requirements.txt"
deactivate
EOF

log "Checking whether base model already exists remotely..."
if ssh_remote "${REMOTE_HOST}" "[[ -f '${REMOTE_MODEL_PATH}' ]]"; then
  log "Remote base model already exists at ${REMOTE_MODEL_PATH}"
else
  model_size="$(du -h "${LOCAL_BASE_MODEL}" | awk '{print $1}')"
  log "Remote base model missing. Starting upload of ${REMOTE_MODEL_NAME} (${model_size})"

  rsync -avh --progress --stats --partial \
    -e "ssh -o StrictHostKeyChecking=accept-new" \
    "${LOCAL_BASE_MODEL}" \
    "${REMOTE_HOST}:${REMOTE_ASSETS_MODELS}/"

  log "Finished base model sync."
fi


REMOTE_ZIP_PATHS=()
for local_zip in "${ZIP_FILES[@]}"; do
  zip_name="$(basename "${local_zip}")"
  remote_zip="${REMOTE_ASSETS_ZIPS}/${zip_name}"

  log "==> Uploading ZIP: ${zip_name}"
  scp_remote "${local_zip}" "${REMOTE_HOST}:${remote_zip}"
  REMOTE_ZIP_PATHS+=("${remote_zip}")
done

log "==> Creating remote run directory..."
ssh_remote "${REMOTE_HOST}" "mkdir -p ${REMOTE_RUN_DIR}"

log "==> Writing manifest.env..."
tmp_manifest="$(mktemp)"
{
  echo "export RUN_ID=\"${RUN_ID}\""
  echo "export RUN_NAME=\"${RUN_NAME}\""
  echo "export RUN_DIR=\"${REMOTE_RUN_DIR}\""
  echo "export BASE_MODEL_PATH=\"${REMOTE_MODEL_PATH}\""
  echo "export MODEL_PREFIX=\"kraken_model_${RUN_ID}\""
  echo "export WORK_DIR=\"${REMOTE_RUN_DIR}/work_kraken_${RUN_ID}\""
  echo "export OUTPUT_DIR=\"${REMOTE_RUN_DIR}/trained_models\""
  echo "export ARTIFACTS_DIR=\"${REMOTE_RUN_DIR}/artifacts\""
  echo "export LOGS_DIR=\"${REMOTE_RUN_DIR}/logs\""
  echo "export ZIP_PATHS=("
  for rz in "${REMOTE_ZIP_PATHS[@]}"; do
    echo "  \"${rz}\""
  done
  echo ")"
} > "${tmp_manifest}"

scp_remote "${tmp_manifest}" "${REMOTE_HOST}:${REMOTE_RUN_DIR}/manifest.env"
rm -f "${tmp_manifest}"

log "==> Submitting job..."
ssh_remote "${REMOTE_HOST}" "
  cd ${REMOTE_RUN_DIR}
  sbatch ${REMOTE_ROOT}/job.sbatch
"

echo
log "Job submitted successfully."
echo "Run name: ${RUN_NAME}"
echo "Remote directory: ${REMOTE_RUN_DIR}"