#!/usr/bin/env bash
set -euo pipefail

# Usage:
# ./submit_remote.sh [--model /path/to/model.mlmodel] \
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
DEFAULT_BASE_MODEL="${REPO_ROOT}/ocrflow/store/models/Gallicorpor.mlmodel"
LOCAL_BASE_MODEL="${DEFAULT_BASE_MODEL}"

POSITIONAL_ARGS=()
while [[ $# -gt 0 ]]; do
  case "$1" in
    --model)
      if [[ $# -lt 2 ]]; then
        echo "Missing value after --model" >&2
        exit 1
      fi
      LOCAL_BASE_MODEL="$2"
      shift 2
      ;;
    *)
      POSITIONAL_ARGS+=("$1")
      shift
      ;;
  esac
done

set -- "${POSITIONAL_ARGS[@]}"

if [[ "$#" -lt 1 ]]; then
  echo "Pass at least one ZIP file." >&2
  exit 1
fi

ZIP_FILES=("$@")

if [[ ! -f "${LOCAL_SCRIPT_PY}" ]]; then
  echo "Missing script.py: ${LOCAL_SCRIPT_PY}" >&2
  exit 1
fi

if [[ ! -f "${LOCAL_REQUIREMENTS}" ]]; then
  echo "Missing requirements.txt: ${LOCAL_REQUIREMENTS}" >&2
  exit 1
fi

if [[ ! -f "${LOCAL_JOB_SBATCH}" ]]; then
  echo "Missing job.sbatch: ${LOCAL_JOB_SBATCH}" >&2
  exit 1
fi

if [[ ! -f "${LOCAL_BASE_MODEL}" ]]; then
  echo "Missing base model: ${LOCAL_BASE_MODEL}" >&2
  exit 1
fi

for local_zip in "${ZIP_FILES[@]}"; do
  if [[ ! -f "${local_zip}" ]]; then
    echo "Missing ZIP file: ${local_zip}" >&2
    exit 1
  fi
done

RUN_ID="$(date +%y%m%d-%H%M)-$RANDOM"
RUN_NAME="kraken_train_${RUN_ID}"
REMOTE_RUN_DIR="${REMOTE_RUNS}/${RUN_NAME}"
REMOTE_MODEL_NAME="$(basename "${LOCAL_BASE_MODEL}")"
REMOTE_MODEL_PATH="${REMOTE_ASSETS_MODELS}/${REMOTE_MODEL_NAME}"

log "Using base model: ${LOCAL_BASE_MODEL}"

log "==> Preparing remote directories..."
ssh_remote "${REMOTE_HOST}" bash <<EOF
set -euo pipefail
if command -v module >/dev/null 2>&1; then
  module purge || true
fi
unset PYTHONHOME PYTHONPATH LD_LIBRARY_PATH

PROJECT_ROOT=${REMOTE_ROOT}

mkdir -p "\${PROJECT_ROOT}" \
	         "\${PROJECT_ROOT}/assets/models" \
	         "\${PROJECT_ROOT}/assets/zips" \
	         "\${PROJECT_ROOT}/runs" \
	         "\${PROJECT_ROOT}/logs"
EOF

log "==> Syncing job files..."
scp_remote "${LOCAL_SCRIPT_PY}" "${REMOTE_HOST}:${REMOTE_ROOT}/script.py"
scp_remote "${LOCAL_REQUIREMENTS}" "${REMOTE_HOST}:${REMOTE_ROOT}/requirements.txt"
scp_remote "${LOCAL_JOB_SBATCH}" "${REMOTE_HOST}:${REMOTE_ROOT}/job.sbatch"

log "==> Ensuring remote Python environment..."
ssh_remote "${REMOTE_HOST}" bash <<EOF
set -euo pipefail
if command -v module >/dev/null 2>&1; then
  module purge || true
fi
unset PYTHONHOME PYTHONPATH LD_LIBRARY_PATH

PROJECT_ROOT=${REMOTE_ROOT}

if [[ ! -d "\${PROJECT_ROOT}/.venv" ]]; then
  log "Creating Python virtual environment..."
  python3 -m venv "\${PROJECT_ROOT}/.venv"
fi

source "\${PROJECT_ROOT}/.venv/bin/activate"
python -m pip install -U pip wheel
python -m pip install -r "\${PROJECT_ROOT}/requirements.txt"
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
  echo "export MODEL_PREFIX=\"kraken_model\""
  echo "export WORK_DIR=\"${REMOTE_RUN_DIR}/workspace\""
  echo "export OUTPUT_DIR=\"${REMOTE_RUN_DIR}/trained_models\""
  echo "export LOGS_DIR=\"${REMOTE_RUN_DIR}/logs\""
  echo "export ZIP_PATHS=("
  for rz in "${REMOTE_ZIP_PATHS[@]}"; do
    echo "  \"${rz}\""
  done
  echo ")"
} > "${tmp_manifest}"

scp_remote "${tmp_manifest}" "${REMOTE_HOST}:${REMOTE_RUN_DIR}/manifest.env"
rm -f "${tmp_manifest}"

log "Submitting job..."
JOB_ID=$(ssh_remote "${REMOTE_HOST}" "
  cd ${REMOTE_RUN_DIR} && sbatch ${REMOTE_ROOT}/job.sbatch
" | awk '{print $4}')

echo
log "Job submitted successfully."
echo "Run name: ${RUN_NAME}"
echo "Remote directory: ${REMOTE_RUN_DIR}"
echo "SLURM Job ID: ${JOB_ID}"
echo

TAIL_CMD="tail -f ${REMOTE_RUN_DIR}/logs/kraken_train_${JOB_ID}.*"
SSH_TAIL_CMD="ssh ${REMOTE_HOST} \"${TAIL_CMD}\""

echo "To monitor logs locally:"
echo "  ${SSH_TAIL_CMD}"
echo
echo "To monitor logs after logging in:"
echo "  ${TAIL_CMD}"
