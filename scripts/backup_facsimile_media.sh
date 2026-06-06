#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Usage:
  backup_facsimile_media.sh --source <rclone-source-root> --dest <rclone-dest-root> [options]

Examples:
  backup_facsimile_media.sh \
    --source commentaria-server:/data/euclides/commentaria-hub/facsimiles \
    --dest /Volumes/CommentariaBackup/commentaria-hub/facsimiles

  backup_facsimile_media.sh \
    --source commentaria-server:/data/euclides/commentaria-hub/facsimiles \
    --dest gdrive:commentaria-hub-media-backup/facsimiles

Options:
  --source PATH       rclone source root containing pdfs/ and diagrams/
  --dest PATH         rclone destination root where pdfs/ and diagrams/ will be mirrored
  --log-dir PATH      directory for logs; default: <dest>/_backup_logs for local destinations,
                     otherwise ./logs/facsimile-media-backup
  --dry-run           show what would change without copying
  --transfers N       concurrent transfers; default: 4
  --checkers N        concurrent checks; default: 8
  --stats SECONDS     rclone stats interval; default: 30s
  --delete-extra      delete destination files that no longer exist on the server
  --no-inplace        do not write files directly at the destination path
  -h, --help          show this help

The copy is resumable and idempotent. Existing matching files are checked and skipped.
Interrupted destination files are kept, then corrected on the next run.
USAGE
}

source_root=""
dest_root=""
log_dir=""
dry_run=0
delete_extra=0
inplace=1
transfers=4
checkers=8
stats_interval="30s"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --source)
      source_root="${2:-}"
      shift 2
      ;;
    --dest)
      dest_root="${2:-}"
      shift 2
      ;;
    --log-dir)
      log_dir="${2:-}"
      shift 2
      ;;
    --dry-run)
      dry_run=1
      shift
      ;;
    --delete-extra)
      delete_extra=1
      shift
      ;;
    --no-inplace)
      inplace=0
      shift
      ;;
    --transfers)
      transfers="${2:-}"
      shift 2
      ;;
    --checkers)
      checkers="${2:-}"
      shift 2
      ;;
    --stats)
      stats_interval="${2:-}"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown argument: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
done

if [[ -z "$source_root" || -z "$dest_root" ]]; then
  echo "Both --source and --dest are required." >&2
  usage >&2
  exit 2
fi

if ! command -v rclone >/dev/null 2>&1; then
  echo "rclone is required. Install it from https://rclone.org/install/ and configure the needed remotes." >&2
  exit 127
fi

source_root="${source_root%/}"
dest_root="${dest_root%/}"

if [[ -z "$log_dir" ]]; then
  if [[ "$dest_root" == /* || "$dest_root" == ./* || "$dest_root" == ../* ]]; then
    log_dir="${dest_root}/_backup_logs"
  else
    log_dir="./logs/facsimile-media-backup"
  fi
fi

mkdir -p "$log_dir"

timestamp="$(date -u +%Y%m%dT%H%M%SZ)"
main_log="${log_dir}/facsimile-media-backup-${timestamp}.log"
combined_log="${log_dir}/facsimile-media-backup-${timestamp}.combined.log"
summary_log="${log_dir}/facsimile-media-backup-${timestamp}.summary.txt"

copy_command="copy"
if [[ "$delete_extra" -eq 1 ]]; then
  copy_command="sync"
fi

rclone_args=(
  "--log-file=${main_log}"
  "--log-level=INFO"
  "--stats=${stats_interval}"
  "--stats-one-line"
  "--human-readable"
  "--transfers=${transfers}"
  "--checkers=${checkers}"
  "--create-empty-src-dirs"
  "--retries=10"
  "--low-level-retries=20"
  "--retries-sleep=30s"
  "--contimeout=60s"
  "--timeout=5m"
)

if [[ "$dry_run" -eq 1 ]]; then
  rclone_args+=("--dry-run")
fi

if [[ "$inplace" -eq 1 ]]; then
  rclone_args+=("--inplace")
fi

run_one() {
  local name="$1"
  local src="${source_root}/${name}"
  local dst="${dest_root}/${name}"

  echo "== ${name} =="
  echo "Source: ${src}"
  echo "Dest:   ${dst}"
  rclone "${rclone_args[@]}" "$copy_command" "$src" "$dst"
}

check_one() {
  local name="$1"
  local src="${source_root}/${name}"
  local dst="${dest_root}/${name}"

  {
    echo
    echo "== ${name} verification =="
    echo "# rclone check combined report"
    echo "# = same, * differs, + missing from destination, - extra on destination, ! error"
  } >> "$combined_log"

  if rclone check "$src" "$dst" --one-way --size-only --combined - >> "$combined_log"; then
    echo "Verified ${name}: source files are present at destination."
    return 0
  fi

  if [[ "$dry_run" -eq 1 ]]; then
    echo "Dry run verification for ${name} found expected differences; see ${combined_log}."
    return 0
  fi

  echo "Verification failed for ${name}; see ${combined_log}." >&2
  return 1
}

{
  echo "started_at=${timestamp}"
  echo "source=${source_root}"
  echo "dest=${dest_root}"
  echo "mode=${copy_command}"
  echo "dry_run=${dry_run}"
  echo "inplace=${inplace}"
  echo "transfers=${transfers}"
  echo "checkers=${checkers}"
  echo "main_log=${main_log}"
  echo "combined_log=${combined_log}"
  echo
} | tee "$summary_log"

run_one "pdfs" 2>&1 | tee -a "$summary_log"
run_one "diagrams" 2>&1 | tee -a "$summary_log"
check_one "pdfs" 2>&1 | tee -a "$summary_log"
check_one "diagrams" 2>&1 | tee -a "$summary_log"

{
  echo
  echo "finished_at=$(date -u +%Y%m%dT%H%M%SZ)"
  echo "Logs:"
  echo "  ${summary_log}"
  echo "  ${main_log}"
  echo "  ${combined_log}"
} | tee -a "$summary_log"
