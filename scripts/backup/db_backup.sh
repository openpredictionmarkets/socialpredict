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

pg_dump_users_cmd() {
  # pg_dump for users table only; data-only with inserts format for easier manipulation
  echo "PGPASSWORD='${POSTGRES_PASSWORD}' pg_dump -U '${POSTGRES_USER}' -h 'localhost' -p '${POSTGRES_PORT}' -d '${POSTGRES_DATABASE}' --table=users --data-only --column-inserts"
}

pg_dump_users_reset_cmd() {
  # pg_dump for users table with balance reset - uses a query to reset account_balance to initial_account_balance
  local query="SELECT id, username, display_name, user_type, email, password, initial_account_balance, initial_account_balance as account_balance, personal_emoji, description, personal_link1, personal_link2, personal_link3, personal_link4, created_at, updated_at FROM users"
  echo "PGPASSWORD='${POSTGRES_PASSWORD}' psql -U '${POSTGRES_USER}' -h 'localhost' -p '${POSTGRES_PORT}' -d '${POSTGRES_DATABASE}' -c \"COPY ($query) TO STDOUT WITH CSV HEADER;\""
}

pg_restore_cmd() {
  # pg_restore inside container; restore into same database; clean objects first
  echo "PGPASSWORD='${POSTGRES_PASSWORD}' pg_restore -U '${POSTGRES_USER}' -h 'localhost' -p '${POSTGRES_PORT}' -d '${POSTGRES_DATABASE}' --clean --if-exists"
}

psql_cmd() {
  # psql command for direct SQL execution
  echo "PGPASSWORD='${POSTGRES_PASSWORD}' psql -U '${POSTGRES_USER}' -h 'localhost' -p '${POSTGRES_PORT}' -d '${POSTGRES_DATABASE}'"
}

latest_backup_file() {
  ls -1t "$BACKUP_ROOT"/socialpredict_backup_"${APP_ENV}"_*.dump.gz 2>/dev/null | head -n 1 || true
}

latest_users_backup_file() {
  ls -1t "$BACKUP_ROOT"/socialpredict_users_"${APP_ENV}"_*.dump.gz 2>/dev/null | head -n 1 || true
}

print_usage() {
  cat <<EOF
Usage: ./SocialPredict backup [options]

DATABASE BACKUP OPERATIONS:
  --save                      Create a new full database backup
  --list                      List available full database backups (newest first)
  --latest                    Print the path to the newest full database backup
  --restore <file>            Restore from specific full database backup file (with confirmation)
  --restore-latest            Restore from the newest full database backup (with confirmation)

USER-ONLY BACKUP OPERATIONS:
  --save-users                Create a users-only backup (preserves current balances)
  --save-users-reset          Create a users-only backup with balances reset to initial values (requires confirmation)
  --list-users                List available users-only backups (newest first)
  --latest-users              Print the path to the newest users-only backup
  --restore-users <file>      Restore from specific users-only backup file (with confirmation)
  --restore-users-latest      Restore from the newest users-only backup (with confirmation)

OTHER:
  --help, -h                  Show this help

Backup Files Created:
  Full Database:
    socialpredict_backup_${APP_ENV}_YYYYmmdd_HHMMSS.dump.gz
    socialpredict_backup_${APP_ENV}_YYYYmmdd_HHMMSS.dump.gz.sha256

  Users Only (current balances):
    socialpredict_users_${APP_ENV}_YYYYmmdd_HHMMSS.dump.gz
    socialpredict_users_${APP_ENV}_YYYYmmdd_HHMMSS.dump.gz.sha256

  Users Only (reset balances):
    socialpredict_users_reset_${APP_ENV}_YYYYmmdd_HHMMSS.csv.gz
    socialpredict_users_reset_${APP_ENV}_YYYYmmdd_HHMMSS.csv.gz.sha256

All backups are stored in: $BACKUP_ROOT

Examples:
  ./SocialPredict backup --save-users          # Backup users with current balances
  ./SocialPredict backup --save-users-reset    # Backup users with reset balances (requires 'RESET' confirmation)
  ./SocialPredict backup --list-users          # List all users-only backups
  ./SocialPredict backup --restore-users-latest # Restore newest users backup (requires 'RESTORE USERS' confirmation)
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

do_save_users() {
  need_container_running
  local ts file tmpfile checksum
  ts="$(timestamp)"
  file="$BACKUP_ROOT/socialpredict_users_${APP_ENV}_${ts}.dump.gz"
  tmpfile="${file}.partial"

  echo "Creating users-only backup: $file"
  echo "NOTE: This backup preserves current account balances."
  
  # Run pg_dump for users table only; stream to host; compress
  if ! docker exec -i "$POSTGRES_CONTAINER_NAME" bash -c "$(pg_dump_users_cmd)" | gzip -c > "$tmpfile"; then
    echo "ERROR: pg_dump failed."
    rm -f "$tmpfile"
    exit 1
  fi

  # Finalize: checksum + move
  checksum="$(sha256_file "$tmpfile")"
  echo "$checksum  $(basename "$file")" > "${file}.sha256"
  mv "$tmpfile" "$file"

  echo "Users backup complete:"
  echo "  File: $file"
  echo "  SHA256: $checksum"
}

do_save_users_reset() {
  need_container_running
  
  echo "WARNING: This will create a users backup with all account balances RESET to initial values."
  echo "Current balances will be lost in this backup."
  echo
  read -r -p "Are you sure you want to reset balances in this backup? (type 'RESET' to confirm): " answer
  if [ "$answer" != "RESET" ]; then
    echo "Aborted."
    exit 1
  fi

  local ts file tmpfile checksum
  ts="$(timestamp)"
  file="$BACKUP_ROOT/socialpredict_users_reset_${APP_ENV}_${ts}.csv.gz"
  tmpfile="${file}.partial"

  echo "Creating users backup with reset balances: $file"
  
  # Run the CSV export with balance reset; stream to host; compress
  if ! docker exec -i "$POSTGRES_CONTAINER_NAME" bash -c "$(pg_dump_users_reset_cmd)" | gzip -c > "$tmpfile"; then
    echo "ERROR: CSV export failed."
    rm -f "$tmpfile"
    exit 1
  fi

  # Finalize: checksum + move
  checksum="$(sha256_file "$tmpfile")"
  echo "$checksum  $(basename "$file")" > "${file}.sha256"
  mv "$tmpfile" "$file"

  echo "Users backup with reset balances complete:"
  echo "  File: $file"
  echo "  SHA256: $checksum"
}

do_list_users() {
  echo "User backups in $BACKUP_ROOT (newest first):"
  echo
  echo "Regular users backups (with current balances):"
  ls -1t "$BACKUP_ROOT"/socialpredict_users_"${APP_ENV}"_*.dump.gz 2>/dev/null || echo "  (none found)"
  echo
  echo "Reset balance users backups:"
  ls -1t "$BACKUP_ROOT"/socialpredict_users_reset_"${APP_ENV}"_*.csv.gz 2>/dev/null || echo "  (none found)"
}

do_latest_users() {
  local latest
  latest="$(latest_users_backup_file)"
  if [ -z "$latest" ]; then
    echo "(none)"
  else
    echo "$latest"
  fi
}

confirm_users_restore() {
  local file="$1"
  echo "WARNING: About to RESTORE USERS into database '${POSTGRES_DATABASE}' inside container '${POSTGRES_CONTAINER_NAME}'."
  echo "This will overwrite ALL EXISTING USER DATA including balances, profiles, etc."
  echo
  echo "Backup file: $file"
  echo
  read -r -p "Type 'RESTORE USERS' to proceed, or anything else to abort: " answer
  if [ "$answer" != "RESTORE USERS" ]; then
    echo "Aborted."
    exit 1
  fi
}

do_restore_users() {
  local file="$1"
  [ -f "$file" ] || { echo "ERROR: Users backup file not found: $file"; exit 1; }
  need_container_running
  confirm_users_restore "$file"

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

  echo "Restoring users data..."
  
  if [[ "$file" == *.csv.gz ]]; then
    # Handle CSV format (reset balances)
    echo "ERROR: CSV restore not yet implemented. Please use regular users backup format."
    exit 1
  else
    # Handle SQL dump format with proper foreign key constraint handling
    local restore_sql
    restore_sql=$(cat <<'EOF'
BEGIN;

-- Set constraints to deferred for this transaction
SET CONSTRAINTS ALL DEFERRED;

-- Clear existing users table
TRUNCATE users RESTART IDENTITY CASCADE;

-- The restore data will be inserted here via stdin

COMMIT;
EOF
)

    # Execute the restore in a transaction with deferred constraints
    {
      echo "$restore_sql" | sed '/-- The restore data will be inserted here via stdin/d'
      echo "-- Inserting users data:"
      gunzip -c "$file"
      echo "COMMIT;"
    } | docker exec -i "$POSTGRES_CONTAINER_NAME" bash -c "$(psql_cmd)"

    if [ $? -ne 0 ]; then
      echo "ERROR: Users restore failed."
      exit 1
    fi
  fi

  echo "Users restore complete."
}

do_restore_users_latest() {
  local latest
  latest="$(latest_users_backup_file)"
  if [ -z "$latest" ]; then
    echo "ERROR: No users backups found in $BACKUP_ROOT"
    exit 1
  fi
  do_restore_users "$latest"
}

# --- Main ---------------------------------------------------------------------
ACTION="${1:-"--help"}"
case "$ACTION" in
  --save)
    do_save
    ;;
  --save-users)
    do_save_users
    ;;
  --save-users-reset)
    do_save_users_reset
    ;;
  --list)
    do_list
    ;;
  --list-users)
    do_list_users
    ;;
  --latest)
    do_latest
    ;;
  --latest-users)
    do_latest_users
    ;;
  --restore)
    shift || true
    [ -n "${1:-}" ] || { echo "ERROR: --restore requires a file path"; exit 1; }
    do_restore_file "$1"
    ;;
  --restore-latest)
    do_restore_latest
    ;;
  --restore-users)
    shift || true
    [ -n "${1:-}" ] || { echo "ERROR: --restore-users requires a file path"; exit 1; }
    do_restore_users "$1"
    ;;
  --restore-users-latest)
    do_restore_users_latest
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
