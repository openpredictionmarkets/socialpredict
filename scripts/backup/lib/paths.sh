# scripts/backup/lib/paths.sh
#
# Path resolution and filesystem helpers

# Determine paths
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
SCRIPT_DIR="$(dirname "$SCRIPT_DIR")"  # Go back to scripts/backup from lib/
ROOT_DIR="$(dirname "$SCRIPT_DIR")"    # .../scripts
ROOT_DIR="$(dirname "$ROOT_DIR")"      # repo root

abs_path() { (cd "$1" 2>/dev/null && pwd -P) || return 1; }

PARENT_OF_ROOT="$(dirname "$(abs_path "$ROOT_DIR")")"
BACKUP_ROOT="$PARENT_OF_ROOT/backups"
mkdir -p "$BACKUP_ROOT"

# Filesystem utilities
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
