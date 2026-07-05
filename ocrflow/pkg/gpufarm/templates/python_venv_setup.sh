set -euo pipefail
if command -v module >/dev/null 2>&1; then
  module purge || true
fi
unset PYTHONHOME PYTHONPATH LD_LIBRARY_PATH
PROJECT_ROOT=%s
if [[ ! -d "${PROJECT_ROOT}/.venv" ]]; then
  python3 -m venv "${PROJECT_ROOT}/.venv"
fi
source "${PROJECT_ROOT}/.venv/bin/activate"
python -m pip install -U pip wheel
python -m pip install -r "${PROJECT_ROOT}/requirements.txt"
deactivate