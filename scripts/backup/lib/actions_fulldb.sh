# scripts/backup/lib/actions_fulldb.sh
#
# Full database operations

# --- Actions: Full DB ---------------------------------------------------------
do_save() {
  need_container_running
  local ts file tmpfile checksum
  ts="$(timestamp)"
  file="$BACKUP_ROOT/socialpredict_backup_${APP_ENV}_${ts}.dump.gz"
  tmpfile="${file}.partial"

  echo "Creating backup: $file"
  if ! psql_smart -c "SELECT current_user, current_database();" >/dev/null 2>&1; then
    echo "ERROR: Cannot connect as POSTGRES_USER='${POSTGRES_USER}'. Check your .env."
    exit 1
  fi

  if ! docker exec -i "$POSTGRES_CONTAINER_NAME" bash -c "$(pg_dump_cmd "$POSTGRES_DATABASE") " | gzip -c > "$tmpfile"; then
    echo "ERROR: pg_dump failed."
    rm -f "$tmpfile"; exit 1
  fi

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
  local latest; latest="$(latest_backup_file)"
  [ -z "$latest" ] && { echo "(none)"; return; }
  echo "$latest"
}

confirm_restore() {
  local file="$1" target_db="$2"
  echo "About to RESTORE into database '${target_db}' inside container '${POSTGRES_CONTAINER_NAME}'."
  echo "This will overwrite existing data for that database."
  echo
  echo "Backup file: $file"
  echo
  read -r -p "Type EXACT database name (${target_db}) to proceed, or anything else to abort: " answer
  [ "$answer" = "$target_db" ] || { echo "Aborted."; exit 1; }
}

do_restore_file() {
  local file="$1"
  [ -f "$file" ] || { echo "ERROR: Backup file not found: $file"; exit 1; }
  need_container_running

  local target_db="${RESTORE_DB:-$POSTGRES_DATABASE}"
  confirm_restore "$file" "$target_db"

  local sumfile="${file}.sha256"
  if [ -f "$sumfile" ]; then
    echo "Verifying checksum..."
    local expected actual
    expected="$(awk '{print $1}' "$sumfile")"
    actual="$(sha256_file "$file")"
    [ "$expected" = "$actual" ] || { echo "ERROR: Checksum mismatch!"; exit 1; }
    echo "Checksum OK."
  fi

  if ! psql_smart -c "SELECT current_user;" >/dev/null 2>&1; then
    echo "ERROR: Cannot connect to Postgres as POSTGRES_USER='${POSTGRES_USER}'."
    exit 1
  fi

  ensure_db_exists "$target_db"

  echo "Restoring (owners/privileges $( [ "${PRESERVE_OWNERS:-0}" = "1" ] && echo "PRESERVED" || echo "IGNORED" ))..."
  if ! gunzip -c "$file" | docker exec -i "$POSTGRES_CONTAINER_NAME" bash -c "$(pg_restore_cmd "$target_db") "; then
    echo "ERROR: pg_restore failed." ; exit 1
  fi
  echo "Restore complete."
}

do_restore_latest() {
  local latest; latest="$(latest_backup_file)"
  [ -z "$latest" ] && { echo "ERROR: No backups found in $BACKUP_ROOT"; exit 1; }
  do_restore_file "$latest"
}

do_inspect() {
  local file="$1"
  [ -f "$file" ] || { echo "ERROR: Backup file not found: $file"; exit 1; }
  local tmp; tmp="$(mktemp -t sp_dump_XXXXXX.dump)"
  trap 'rm -f "$tmp"' EXIT
  gunzip -c "$file" > "$tmp"

  echo "== Dump inspection =="; echo "File: $file"; echo
  echo "-- Owners referenced --"
  (pg_restore -l "$tmp" | grep -Eo 'OWNER TO [^;]+' | sort -u) || echo "(none)"
  echo
  echo "-- Privilege statements (GRANT/REVOKE) --"
  (pg_restore -l "$tmp" | grep -E 'GRANT |REVOKE ' | sort -u | sed -n '1,200p') || echo "(none)"
  echo
  echo "-- Extensions requested --"
  (pg_restore -l "$tmp" | grep -i 'EXTENSION - ' | sed -n '1,200p') || echo "(none)"
}
