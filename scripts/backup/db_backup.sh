#!/usr/bin/env bash
# scripts/backup/db_backup.sh
#
# Database backup & restore utilities for SocialPredict
#
# Full DB:
#   ./SocialPredict backup --save
#   ./SocialPredict backup --list
#   ./SocialPredict backup --latest
#   ./SocialPredict backup --restore </path/to/backup.dump.gz>
#   ./SocialPredict backup --restore-latest
#   ./SocialPredict backup --inspect </path/to/backup.dump.gz>
#
# Users only:
#   ./SocialPredict backup --save-users
#   ./SocialPredict backup --save-users-reset
#   ./SocialPredict backup --list-users
#   ./SocialPredict backup --latest-users
#   ./SocialPredict backup --restore-users </path/to/users.csv.gz>
#   ./SocialPredict backup --restore-users-latest
#
# Environment overrides (optional):
#   PRESERVE_OWNERS=1      # keep original owners/privileges on restore (default: off)
#   RESTORE_DB=<name>      # restore into this DB instead of POSTGRES_DATABASE (full DB only)
#   RESTORE_JOBS=<n>       # pg_restore parallel jobs (-j), e.g. 4
#
# Notes:
# - Must be called via ./SocialPredict (enforced by CALLED_FROM_SOCIALPREDICT guard)
# - Uses .env for DB container/name/user/pass/ports
# - Writes backups to a directory PARALLEL to repo root: ../backups/
# - Filenames:
#     Full DB: socialpredict_backup_${APP_ENV}_${YYYYmmdd_HHMMSS}.dump.gz + .sha256
#     Users:   socialpredict_users_${APP_ENV}_${YYYYmmdd_HHMMSS}.csv.gz     + .sha256
#     UsersR:  socialpredict_users_reset_${APP_ENV}_${YYYYmmdd_HHMMSS}.csv.gz + .sha256

set -euo pipefail

# Source library modules in dependency order
source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/lib/paths.sh"
source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/lib/env.sh"
source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/lib/pg.sh"
source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/lib/users_csv.sh"
source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/lib/actions_fulldb.sh"
source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/lib/actions_users.sh"

print_usage() {
  cat <<EOF
Usage: ./SocialPredict backup [options]

DATABASE BACKUP OPERATIONS:
  --save                      Create a new full database backup
  --list                      List available full database backups (newest first)
  --latest                    Print the path to the newest full database backup
  --restore <file>            Restore from specific full database backup file (with confirmation)
  --restore-latest            Restore from the newest full database backup (with confirmation)
  --inspect <file>            Show owner/privilege hints & extensions in a dump

USER-ONLY BACKUP OPERATIONS:
  --save-users                Create a users-only backup (preserves current balances)
  --save-users-reset          Create a users-only backup with balances reset to initial values (requires confirmation)
  --list-users                List available users-only backups (newest first)
  --latest-users              Print the path to the newest users-only backup
  --restore-users <file>      Restore users from CSV backup (MERGE/UPSERT by username)
  --restore-users-latest      Restore from the newest users-only backup (MERGE/UPSERT)

OTHER:
  --help, -h                  Show this help

Backup Files Created:
  Full Database:
    socialpredict_backup_${APP_ENV}_YYYYmmdd_HHMMSS.dump.gz
    socialpredict_backup_${APP_ENV}_YYYYmmdd_HHMMSS.dump.gz.sha256

  Users Only (current balances):
    socialpredict_users_${APP_ENV}_YYYYmmdd_HHMMSS.csv.gz
    socialpredict_users_${APP_ENV}_YYYYmmdd_HHMMSS.csv.gz.sha256

  Users Only (reset balances):
    socialpredict_users_reset_${APP_ENV}_YYYYmmdd_HHMMSS.csv.gz
    socialpredict_users_reset_${APP_ENV}_YYYYmmdd_HHMMSS.csv.gz.sha256

All backups are stored in: $BACKUP_ROOT

Environment overrides (full DB restore):
  PRESERVE_OWNERS=1      Keep original owners/privileges (default: skip owners/privs)
  RESTORE_DB=<name>      Restore into this database instead of \$POSTGRES_DATABASE
  RESTORE_JOBS=<n>       Use pg_restore -j <n> parallel jobs
EOF
}

# --- Main ---------------------------------------------------------------------
ACTION="${1:-"--help"}"
case "$ACTION" in
  --save)                    do_save ;;
  --list)                    do_list ;;
  --latest)                  do_latest ;;
  --restore)                 shift || true; [ -n "${1:-}" ] || { echo "ERROR: --restore requires a file path"; exit 1; }; do_restore_file "$1" ;;
  --restore-latest)          do_restore_latest ;;
  --inspect)                 shift || true; [ -n "${1:-}" ] || { echo "ERROR: --inspect requires a file path"; exit 1; }; do_inspect "$1" ;;

  --save-users)              do_save_users ;;
  --save-users-reset)        do_save_users_reset ;;
  --list-users)              do_list_users ;;
  --latest-users)            do_latest_users ;;
  --restore-users)           shift || true; [ -n "${1:-}" ] || { echo "ERROR: --restore-users requires a file path"; exit 1; }; do_restore_users "$1" ;;
  --restore-users-latest)    do_restore_users_latest ;;

  --help|-h)                 print_usage ;;
  *)                         echo "Unknown action: $ACTION"; echo; print_usage; exit 1 ;;
esac
