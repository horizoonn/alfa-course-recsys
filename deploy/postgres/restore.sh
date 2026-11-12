#!/usr/bin/env bash
set -euo pipefail

: "${PGHOST:?PGHOST is required}"
: "${PGUSER:?PGUSER is required}"
: "${PGPASSWORD:?PGPASSWORD is required}"
: "${PGDATABASE:?PGDATABASE is required}"

cd -- "$(dirname -- "${BASH_SOURCE[0]}")"
export PGAPPNAME=alfa-course-restore
export RESTORE_BACKUP=/backups/source.backup

psql -X --no-password --dbname="$PGDATABASE" \
  --set=ON_ERROR_STOP=1 --file=restore.sql
