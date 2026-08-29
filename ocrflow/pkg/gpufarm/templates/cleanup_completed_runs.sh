set -euo pipefail
root={{shellquote .JobRoot}}
keep={{.CompletedRunsToKeep}}

# Do not clean anything if Slurm cannot tell us which work directories are
# active. Skipping cleanup is safer than deleting a running job's files.
if ! active_output="$(squeue -h -u "$USER" -o '%i|%Z')"; then
  echo "status=skipped reason=squeue_failed root=$root"
  exit 0
fi

read -r used_before available_before < <(df -B1 --output=used,avail "$root" | tail -n 1)
echo "status=started root=$root used_bytes=$used_before available_bytes=$available_before keep_completed_per_job=$keep"

declare -A active_by_dir=()
while IFS='|' read -r job_id work_dir; do
  if [[ -n "$job_id" && -n "$work_dir" ]]; then
    active_by_dir["$work_dir"]="$job_id"
    echo "active_job id=$job_id work_dir=$work_dir"
  fi
done <<< "$active_output"

deleted_runs=0
deleted_bytes=0
deleted_abandoned=0
deleted_retention=0
mapfile -d '' -t job_dirs < <(
  find "$root" -mindepth 1 -maxdepth 1 -type d -print0 | sort -z
)
for job_dir in "${job_dirs[@]}"; do
  mapfile -d '' -t runs < <(
    find "$job_dir" -mindepth 1 -maxdepth 1 -type d -name 'run_*' -printf '%T@ %p\0' |
      sort -z -nr
  )

  retained=0
  for entry in "${runs[@]}"; do
    run_dir="${entry#* }"
    if [[ -n "${active_by_dir[$run_dir]:-}" ]]; then
	  echo "protected reason=active run_dir=$run_dir job_id=${active_by_dir[$run_dir]}"
      continue
    fi

	delete_reason=""
	if [[ ! -e "$run_dir/artifacts/done" && ! -e "$run_dir/artifacts/failed" ]] &&
	   find "$run_dir" -maxdepth 0 -mmin +1440 -print -quit | grep -q .; then
	  # Uploads that crash before sbatch never create an artifact marker and do
	  # not appear in squeue. Remove them after a safety window, independently
	  # of completed-run retention.
	  delete_reason="abandoned_upload_older_than_24h"
	  ((deleted_abandoned += 1))
	elif (( retained < keep )); then
	  ((retained += 1))
	  echo "retained reason=recent run_dir=$run_dir position=$retained keep=$keep"
	  continue
	else
	  delete_reason="retention_limit"
	  ((deleted_retention += 1))
	fi

	job_ids=""
	if [[ -d "$run_dir/logs" ]]; then
	  job_ids="$(
	    find "$run_dir/logs" -maxdepth 1 -type f -printf '%f\n' |
	      sed -nE 's/.*_([0-9]+)\.(out|err)$/\1/p' |
	      sort -u |
	      paste -sd, -
	  )"
	fi
	run_bytes="$(du -sb "$run_dir" | cut -f1)"
	rm -rf -- "$run_dir"
	((deleted_runs += 1))
	((deleted_bytes += run_bytes))
	echo "deleted reason=$delete_reason run_dir=$run_dir job_ids=${job_ids:-unknown} bytes=$run_bytes"
  done
done

read -r used_after available_after < <(df -B1 --output=used,avail "$root" | tail -n 1)
reclaimed_bytes=$((used_before - used_after))
echo "status=completed deleted_runs=$deleted_runs deleted_abandoned=$deleted_abandoned deleted_retention=$deleted_retention deleted_bytes=$deleted_bytes reclaimed_bytes=$reclaimed_bytes used_bytes=$used_after available_bytes=$available_after"
