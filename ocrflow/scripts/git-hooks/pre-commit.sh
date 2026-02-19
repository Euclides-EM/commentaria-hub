#!/usr/bin/env bash

# location: .git/hooks/pre-commit

set -e

# ---- CONFIG ----
DB_PATH="ocrflow/store/ocrflow.db"
BACKUP_DIR="ocrflow/backups"
DATE="$(date +"%Y-%m-%d_%H-%M-%S")"
BACKUP_NAME="ocrflow_${DATE}.db"
BACKUP_PATH="${BACKUP_DIR}/${BACKUP_NAME}"
# ----------------

# Skip if DB does not exist
if [ ! -f "$DB_PATH" ]; then
  echo "pre-commit: ocrflow.db not found, skipping backup"
  exit 0
fi

mkdir -p "$BACKUP_DIR"

# Create backup
cp "$DB_PATH" "$BACKUP_PATH"

# Stage backup
git add "$BACKUP_PATH"

echo "pre-commit: database backup created at ${BACKUP_PATH}"
