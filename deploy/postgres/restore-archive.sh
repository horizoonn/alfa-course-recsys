#!/usr/bin/env bash
set -euo pipefail

: "${PGDATABASE:?PGDATABASE is required}"
: "${RESTORE_BACKUP:?RESTORE_BACKUP is required}"

if [[ ! -f "$RESTORE_BACKUP" || ! -r "$RESTORE_BACKUP" || ! -s "$RESTORE_BACKUP" ]]; then
  printf 'A nonempty, readable backup file is required: %s\n' "$RESTORE_BACKUP" >&2
  exit 1
fi

pg_restore --version
printf 'Restoring archive into database %s.\n' "$PGDATABASE"
pg_restore --list "$RESTORE_BACKUP" > /dev/null
pg_restore --no-password --dbname="$PGDATABASE" \
  --no-owner --no-acl --single-transaction --verbose "$RESTORE_BACKUP"
