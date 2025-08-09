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
#   ./SocialPredict backup --help
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

# Determine SCRIPT_DIR (root/scripts/backup) and BACKUP_ROOT (../backups relative to root)
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"         # .../scripts
ROOT_DIR="$(dirname "$ROOT_DIR")"           # repo root
# readlink -f is not available on macOS by default; do a portable resolution:
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

pg_dump_cmd() {
  # pg_dump inside container; output to stdout; custom format; gzip on host
  # Use in-container localhost & port; auth via PGPASSWORD env
  echo "PGPASSWORD='${POSTGRES_PASSWORD}' pg_dump -U '${POSTGRES_USER}' -h 'localhost' -p '${POSTGRES_PORT}' -d '${POSTGRES_DATABASE}' -Fc"
}

pg_restore_cmd() {
  # pg_restore inside container; restore into same database; clean objects first
  echo "PGPASSWORD='${POSTGRES_PASSWORD}' pg_restore -U '${POSTGRES_USER}' -h 'localhost' -p '${POSTGRES_PORT}' -d '${POSTGRES_DATABASE}' --clean --if-exists"
}

latest_backup_file() {
  ls -1t "$BACKUP_ROOT"/socialpredict_backup_"${APP_ENV}"_*.dump.gz 2>/dev/null | head -n 1 || true
}

print_usage() {
  cat <<EOF
Usage: ./SocialPredict backup [--save | --list | --latest | --restore <file> | --restore-latest | --help]

  --save             Create a new backup in $BACKUP_ROOT
  --list             List available backups (newest first)
  --latest           Print the path to the newest backup
  --restore <file>   Restore the given .dump.gz into the running DB (with confirmation)
  --restore-latest   Restore the newest backup (with confirmation)
  --help             Show this help

Backups:
  socialpredict_backup_${APP_ENV}_YYYYmmdd_HHMMSS.dump.gz
  socialpredict_backup_${APP_ENV}_YYYYmmdd_HHMMSS.dump.gz.sha256
EOF
}

# --- Actions ------------------------------------------------------------------
do_save() {
  need_container_running
  local ts file tmpfile checksum
  ts="$(timestamp)"
  file="$BACKUP_ROOT/socialpredict_backup_${APP_ENV}_${ts}.dump.gz"
  tmpfile="${file}.partial"

  echo "Creating backup: $file"
  # Run pg_dump inside the container; stream to host; compress
  if ! docker exec -i "$POSTGRES_CONTAINER_NAME" bash -c "$(pg_dump_cmd) " | gzip -c > "$tmpfile"; then
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
  local file="$1"
  echo "About to RESTORE into database '${POSTGRES_DATABASE}' inside container '${POSTGRES_CONTAINER_NAME}'."
  echo "This will overwrite existing data for that database."
  echo
  echo "Backup file: $file"
  echo
  read -r -p "Type EXACT database name (${POSTGRES_DATABASE}) to proceed, or anything else to abort: " answer
  if [ "$answer" != "$POSTGRES_DATABASE" ]; then
    echo "Aborted."
    exit 1
  fi
}

do_restore_file() {
  local file="$1"
  [ -f "$file" ] || { echo "ERROR: Backup file not found: $file"; exit 1; }
  need_container_running
  confirm_restore "$file"

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

  echo "Restoring..."
  # Decompress on host; stream into pg_restore in container
  if ! gunzip -c "$file" | docker exec -i "$POSTGRES_CONTAINER_NAME" bash -c "$(pg_restore_cmd) "; then
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
