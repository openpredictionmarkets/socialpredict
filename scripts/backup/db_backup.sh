#!/usr/bin/env bash
# scripts/backup/db_backup.sh
#
# Database backup & restore utilities for SocialPredict
# Usage:
#   ./SocialPredict backup --save
#   ./SocialPredict backup --list
#   ./SocialPredict backup --latest
#   ./SocialPredict backup --restore </path/to/backup.dump.gz>
#   ./SocialPredict backup --restore-latest
#   ./SocialPredict backup --inspect </path/to/backup.dump.gz>
#   ./SocialPredict backup --help
#
# Environment overrides (optional):
#   PRESERVE_OWNERS=1      # keep original owners/privileges on restore (default: off)
#   RESTORE_DB=<name>      # restore into this DB instead of POSTGRES_DATABASE
#   RESTORE_JOBS=<n>       # pg_restore parallel jobs (-j), e.g. 4
#
# Notes:
# - Must be called via ./SocialPredict (enforced by CALLED_FROM_SOCIALPREDICT guard)
# - Uses .env for DB container/name/user/pass/ports
# - Writes backups to a directory PARALLEL to repo root: ../backups/
# - Filenames: socialpredict_backup_${APP_ENV}_${YYYYmmdd_HHMMSS}.dump.gz + .sha256

set -euo pipefail

# --- Guards -------------------------------------------------------------------
[ -z "${CALLED_FROM_SOCIALPREDICT:-}" ] && { echo "Not called from SocialPredict"; exit 42; }

# .env is already sourced by the SocialPredict script before calling this file
: "${APP_ENV:?APP_ENV not set}"
: "${POSTGRES_CONTAINER_NAME:?POSTGRES_CONTAINER_NAME not set}"
: "${POSTGRES_USER:?POSTGRES_USER not set}"
: "${POSTGRES_PASSWORD:?POSTGRES_PASSWORD not set}"
: "${POSTGRES_DATABASE:?POSTGRES_DATABASE not set}"
: "${POSTGRES_PORT:=5432}" # default for in-container access

# Determine paths
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"         # .../scripts
ROOT_DIR="$(dirname "$ROOT_DIR")"           # repo root
abs_path() { (cd "$1" 2>/dev/null && pwd -P) || return 1; }
PARENT_OF_ROOT="$(dirname "$(abs_path "$ROOT_DIR")")"
BACKUP_ROOT="$PARENT_OF_ROOT/backups"
mkdir -p "$BACKUP_ROOT"

# --- Helpers ------------------------------------------------------------------
timestamp() { date +"%Y%m%d_%H%M%S"; }

sha256_file() {
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$1" | awk '{print $1}'
  elif command -v shasum >/dev/null 2>&1; then
    shasum -a 256 "$1" | awk '{print $1}'
  else
    echo "ERROR: Neither sha256sum nor shasum found." >&2
    return 1
  fi
}

container_running() {
  docker ps --format '{{.Names}}' | grep -qx "$POSTGRES_CONTAINER_NAME"
}

need_container_running() {
  if ! container_running; then
    echo "ERROR: Postgres container '$POSTGRES_CONTAINER_NAME' is not running."
    echo "Start it first: ./SocialPredict up"
    exit 1
  fi
}

# psql wrapper; tries DB, then postgres, then template1
psql_try() {
  local db="$1"; shift
  docker exec -e PGPASSWORD="$POSTGRES_PASSWORD" -i "$POSTGRES_CONTAINER_NAME" \
    psql -v ON_ERROR_STOP=1 -h localhost -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$db" "$@"
}

psql_smart() {
  # Try connecting to configured DB; if missing, fall back to postgres/template1 purely for control commands.
  if psql_try "$POSTGRES_DATABASE" -c "SELECT 1;" >/dev/null 2>&1; then
    psql_try "$POSTGRES_DATABASE" "$@"
    return
  fi
  if psql_try postgres -c "SELECT 1;" >/dev/null 2>&1; then
    psql_try postgres "$@"
    return
  fi
  psql_try template1 "$@"
}

pg_dump_cmd() {
  # Custom format; output to stdout; auth via env
  local db="$1"
  echo "PGPASSWORD='${POSTGRES_PASSWORD}' pg_dump -U '${POSTGRES_USER}' -h 'localhost' -p '${POSTGRES_PORT}' -d '${db}' -Fc"
}

pg_restore_cmd() {
  # Build pg_restore with safe defaults; accepts <target-db> as $1
  local target_db="$1"; shift || true
  local flags=( --clean --if-exists )
  # Unless preserving owners, ignore owners & privileges (portable across clusters/roles)
  if [ "${PRESERVE_OWNERS:-0}" != "1" ]; then
    flags+=( --no-owner --no-privileges )
  fi
  # Optional parallel jobs
  if [ -n "${RESTORE_JOBS:-}" ]; then
    flags+=( -j "$RESTORE_JOBS" )
  fi
  echo "PGPASSWORD='${POSTGRES_PASSWORD}' pg_restore -U '${POSTGRES_USER}' -h 'localhost' -p '${POSTGRES_PORT}' -d '${target_db}' ${flags[*]}"
}

latest_backup_file() {
  ls -1t "$BACKUP_ROOT"/socialpredict_backup_"${APP_ENV}"_*.dump.gz 2>/dev/null | head -n 1 || true
}

print_usage() {
  cat <<EOF
Usage: ./SocialPredict backup [--save | --list | --latest | --restore <file> | --restore-latest | --inspect <file> | --help]

  --save                 Create a new backup in $BACKUP_ROOT
  --list                 List available backups (newest first)
  --latest               Print the path to the newest backup
  --restore <file>       Restore the given .dump.gz into the running DB (with confirmation)
  --restore-latest       Restore the newest backup (with confirmation)
  --inspect <file>       Show owner/privilege hints and extensions present in the dump
  --help                 Show this help

Environment overrides:
  PRESERVE_OWNERS=1      Keep original owners/privileges (default: skip owners/privs)
  RESTORE_DB=<name>      Restore into this database instead of \$POSTGRES_DATABASE
  RESTORE_JOBS=<n>       Use pg_restore -j <n> parallel jobs

Backups:
  socialpredict_backup_${APP_ENV}_YYYYmmdd_HHMMSS.dump.gz
  socialpredict_backup_${APP_ENV}_YYYYmmdd_HHMMSS.dump.gz.sha256
EOF
}

ensure_db_exists() {
  local db="$1"
  # If DB exists, nothing to do
  if psql_smart -tAc "SELECT 1 FROM pg_database WHERE datname='${db}'" | grep -q 1; then
    return
  fi
  echo "Target database '${db}' not found. Creating it (OWNER=${POSTGRES_USER})..."
  psql_smart -c "CREATE DATABASE \"${db}\" OWNER \"${POSTGRES_USER}\";"
}

# --- Actions ------------------------------------------------------------------
do_save() {
  need_container_running
  local ts file tmpfile checksum
  ts="$(timestamp)"
  file="$BACKUP_ROOT/socialpredict_backup_${APP_ENV}_${ts}.dump.gz"
  tmpfile="${file}.partial"

  echo "Creating backup: $file"
  # Validate we can connect with provided user
  if ! psql_smart -c "SELECT current_user, current_database();" >/dev/null 2>&1; then
    echo "ERROR: Cannot connect as POSTGRES_USER='${POSTGRES_USER}'. Check your .env credentials/role."
    exit 1
  fi

  # Run pg_dump inside the container; stream to host; compress
  if ! docker exec -i "$POSTGRES_CONTAINER_NAME" bash -c "$(pg_dump_cmd "$POSTGRES_DATABASE") " | gzip -c > "$tmpfile"; then
    echo "ERROR: pg_dump failed."
    rm -f "$tmpfile"
    exit 1
  fi

  # Finalize: checksum + move
  checksum="$(sha256_file "$tmpfile")"
  echo "$checksum  $(basename "$file")" > "${file}.sha256"
  mv "$tmpfile" "$file"

  echo "Backup complete:"
  echo "  File: $file"
  echo "  SHA256: $checksum"
}

do_list() {
  echo "Backups in $BACKUP_ROOT (newest first):"
  ls -1t "$BACKUP_ROOT"/socialpredict_backup_"${APP_ENV}"_*.dump.gz 2>/dev/null || echo "(none found)"
}

do_latest() {
  local latest
  latest="$(latest_backup_file)"
  if [ -z "$latest" ]; then
    echo "(none)"
  else
    echo "$latest"
  fi
}

confirm_restore() {
  local file="$1" target_db="$2"
  echo "About to RESTORE into database '${target_db}' inside container '${POSTGRES_CONTAINER_NAME}'."
  echo "This will overwrite existing data for that database."
  echo
  echo "Backup file: $file"
  echo
  read -r -p "Type EXACT database name (${target_db}) to proceed, or anything else to abort: " answer
  if [ "$answer" != "$target_db" ]; then
    echo "Aborted."
    exit 1
  fi
}

do_restore_file() {
  local file="$1"
  [ -f "$file" ] || { echo "ERROR: Backup file not found: $file"; exit 1; }
  need_container_running

  # Decide target DB
  local target_db="${RESTORE_DB:-$POSTGRES_DATABASE}"
  confirm_restore "$file" "$target_db"

  # Verify checksum if present
  local sumfile="${file}.sha256"
  if [ -f "$sumfile" ]; then
    echo "Verifying checksum..."
    local expected actual
    expected="$(awk '{print $1}' "$sumfile")"
    actual="$(sha256_file "$file")"
    if [ "$expected" != "$actual" ]; then
      echo "ERROR: Checksum mismatch! Expected=$expected Actual=$actual"
      exit 1
    fi
    echo "Checksum OK."
  fi

  # Basic connection sanity
  if ! psql_smart -c "SELECT current_user;" >/dev/null 2>&1; then
    echo "ERROR: Cannot connect to Postgres as POSTGRES_USER='${POSTGRES_USER}'."
    echo "Hint: If you see 'role \"postgres\" does not exist', set POSTGRES_USER in .env to the actual superuser for this cluster."
    exit 1
  fi

  # Ensure target DB exists (pg_restore needs it)
  ensure_db_exists "$target_db"

  echo "Restoring (owners/privileges $( [ "${PRESERVE_OWNERS:-0}" = "1" ] && echo "PRESERVED" || echo "IGNORED" ))..."
  # Decompress on host; stream into pg_restore in container
  if ! gunzip -c "$file" | docker exec -i "$POSTGRES_CONTAINER_NAME" bash -c "$(pg_restore_cmd "$target_db") "; then
    echo "ERROR: pg_restore failed."
    exit 1
  fi

  echo "Restore complete."
}

do_restore_latest() {
  local latest
  latest="$(latest_backup_file)"
  if [ -z "$latest" ]; then
    echo "ERROR: No backups found in $BACKUP_ROOT"
    exit 1
  fi
  do_restore_file "$latest"
}

do_inspect() {
  local file="$1"
  [ -f "$file" ] || { echo "ERROR: Backup file not found: $file"; exit 1; }

  # Decompress to a temp (pg_restore -l needs a seekable archive file)
  local tmp
  tmp="$(mktemp -t sp_dump_XXXXXX.dump)"
  trap 'rm -f "$tmp"' EXIT
  gunzip -c "$file" > "$tmp"

  echo "== Dump inspection =="
  echo "File: $file"
  echo
  echo "-- Owners referenced --"
  if ! pg_restore -l "$tmp" | grep -Eo 'OWNER TO [^;]+' | sort -u; then
    echo "(none found or not applicable)"
  fi
  echo
  echo "-- Privilege statements (GRANT/REVOKE) --"
  if ! pg_restore -l "$tmp" | grep -E 'GRANT |REVOKE ' | sort -u | sed -n '1,200p'; then
    echo "(none found or not applicable)"
  fi
  echo
  echo "-- Extensions requested --"
  if ! pg_restore -l "$tmp" | grep -i 'EXTENSION - ' | sed -n '1,200p'; then
    echo "(none found)"
  fi
}

# --- Main ---------------------------------------------------------------------
ACTION="${1:-"--help"}"
case "$ACTION" in
  --save)
    do_save
    ;;
  --list)
    do_list
    ;;
  --latest)
    do_latest
    ;;
  --restore)
    shift || true
    [ -n "${1:-}" ] || { echo "ERROR: --restore requires a file path"; exit 1; }
    do_restore_file "$1"
    ;;
  --restore-latest)
    do_restore_latest
    ;;
  --inspect)
    shift || true
    [ -n "${1:-}" ] || { echo "ERROR: --inspect requires a file path"; exit 1; }
    do_inspect "$1"
    ;;
  --help|-h)
    print_usage
    ;;
  *)
    echo "Unknown action: $ACTION"
    echo
    print_usage
    exit 1
    ;;
esac
