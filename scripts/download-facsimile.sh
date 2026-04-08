#!/usr/bin/env bash
set -euo pipefail

if [ "$#" -lt 1 ]; then
  echo "usage: $0 <edition_id> [edition_id ...] [existing_output_dir]" >&2
  exit 1
fi

out_dir="."

if [ "$#" -gt 1 ] && [ -d "${!#}" ]; then
  out_dir="${!#}"
  set -- "${@:1:$(($#-1))}"
fi

repo_url="${FACSIMILES_GITHUB_REPO_URL:-https://github.com/Euclides-EM/elements-facsimile}"

case "$repo_url" in
  https://github.com/*/*)
    repo_path="${repo_url#https://github.com/}"
    owner="${repo_path%%/*}"
    repo="${repo_path#*/}"
    repo="${repo%.git}"
    ;;
  *)
    echo "unsupported FACSIMILES_GITHUB_REPO_URL: $repo_url" >&2
    exit 1
    ;;
esac

mkdir -p "$out_dir"

extract_json_string() {
  local key="$1"
  sed -n "s/.*\"${key}\":[[:space:]]*\"\\([^\"]*\\)\".*/\\1/p" | head -n 1
}

download_one() {
  local edition_id="$1"
  local out="${out_dir}/${edition_id}.pdf"
  local api_url="https://api.github.com/repos/${owner}/${repo}/contents/docs/${edition_id}.pdf?ref=main"
  local meta
  local download_url
  local raw
  local oid
  local size
  local batch_url
  local batch_payload
  local batch_response
  local lfs_url

  meta="$(curl -fsSL -H "Accept: application/vnd.github+json" "$api_url")"
  download_url="$(printf '%s\n' "$meta" | extract_json_string download_url)"

  if [ -z "$download_url" ]; then
    echo "failed to resolve download_url for ${edition_id}" >&2
    exit 1
  fi

  raw="$(mktemp)"
  trap 'rm -f "$raw"' RETURN

  curl -fsSL -H "Accept: application/vnd.github.v3.raw" "$download_url" -o "$raw"

  if head -n 1 "$raw" | grep -qx 'version https://git-lfs.github.com/spec/v1'; then
    oid="$(sed -n 's/^oid sha256:\([a-f0-9]\{64\}\)$/\1/p' "$raw" | head -n 1)"
    size="$(sed -n 's/^size \([0-9][0-9]*\)$/\1/p' "$raw" | head -n 1)"

    if [ -z "$oid" ] || [ -z "$size" ]; then
      echo "failed to parse LFS pointer for ${edition_id}" >&2
      exit 1
    fi

    batch_url="https://github.com/${owner}/${repo}.git/info/lfs/objects/batch"
    batch_payload=$(printf '{"operation":"download","objects":[{"oid":"%s","size":%s}]}' "$oid" "$size")
    batch_response="$(curl -fsSL \
      -H "Accept: application/vnd.git-lfs+json" \
      -H "Content-Type: application/vnd.git-lfs+json" \
      -d "$batch_payload" \
      "$batch_url")"

    lfs_url="$(printf '%s\n' "$batch_response" | extract_json_string href)"
    if [ -z "$lfs_url" ]; then
      echo "failed to resolve LFS download URL for ${edition_id}" >&2
      exit 1
    fi

    curl -fL "$lfs_url" -o "$out"
  else
    mv "$raw" "$out"
    raw=""
  fi

  if [ -n "$raw" ]; then
    rm -f "$raw"
  fi

  echo "saved $out"
}

for edition_id in "$@"; do
  download_one "$edition_id"
done
