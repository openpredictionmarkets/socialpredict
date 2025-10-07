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

do_list_all() {
  echo "Backups in $BACKUP_ROOT (newest first):"
  echo
  echo "Full database backups:"
  ls -1t "$BACKUP_ROOT"/socialpredict_backup_"${APP_ENV}"_*.dump.gz 2>/dev/null || echo "  (none found)"
  echo
  echo "Users-only backups (current balances):"
  ls -1t "$BACKUP_ROOT"/socialpredict_users_"${APP_ENV}"_*.csv.gz 2>/dev/null || echo "  (none found)"
  echo
  echo "Users-only backups (reset balances):"
  ls -1t "$BACKUP_ROOT"/socialpredict_users_reset_"${APP_ENV}"_*.csv.gz 2>/dev/null || echo "  (none found)"
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

  need_container_running

  echo "== Dump inspection =="
  echo "File: $file"
  echo

  # 1) Decompress to a seekable temp file on host
  local tmp_host
  tmp_host="$(mktemp -t sp_dump_XXXXXX.dump)"
  # shellcheck disable=SC2064
  trap "rm -f '$tmp_host'" RETURN
  if ! gunzip -c "$file" > "$tmp_host"; then
    echo "ERROR: Failed to decompress dump."
    exit 1
  fi

  # 2) Try inside the running PG container (copy in, run, clean up)
  local toc="" tmp_in
  tmp_in="/tmp/$(basename "$tmp_host")"
  if docker cp "$tmp_host" "$POSTGRES_CONTAINER_NAME:$tmp_in" >/dev/null 2>&1; then
    if toc="$(docker exec "$POSTGRES_CONTAINER_NAME" sh -lc 'command -v pg_restore >/dev/null 2>&1 && pg_restore -l '"$tmp_in"' 2>/dev/null' )" && [ -n "$toc" ]; then
      docker exec "$POSTGRES_CONTAINER_NAME" sh -lc 'rm -f '"$tmp_in"'' >/dev/null 2>&1 || true
    else
      docker exec "$POSTGRES_CONTAINER_NAME" sh -lc 'rm -f '"$tmp_in"'' >/dev/null 2>&1 || true
      toc=""
    fi
  fi

  # 3) Fallback: ephemeral container using the SAME image as the running PG
  if [ -z "$toc" ]; then
    local img
    img="$(docker inspect "$POSTGRES_CONTAINER_NAME" --format '{{.Config.Image}}' 2>/dev/null || true)"
    [ -n "$img" ] || img="postgres:16-alpine"
    toc="$(docker run --rm -v "$tmp_host":/dump.dump:ro "$img" sh -lc 'pg_restore -l /dump.dump 2>/dev/null' || true)"
    if [ -z "$toc" ]; then
      echo "ERROR: Failed to run 'pg_restore -l' (checked running container and fallback image: $img)."
      echo "Optionally install pg_restore on host: macOS 'brew install libpq && brew link --force libpq', Debian/Ubuntu 'apt-get install postgresql-client'."
      exit 1
    fi
  fi

  echo "-- Owners referenced --"
  if ! echo "$toc" | grep -Eo 'OWNER TO [^;]+' | sort -u | sed -n '1,200p'; then
    echo "(none)"
  fi
  echo

  echo "-- Privilege statements (GRANT/REVOKE) --"
  if ! echo "$toc" | grep -E 'GRANT |REVOKE ' | sort -u | sed -n '1,200p'; then
    echo "(none)"
  fi
  echo

  echo "-- Extensions requested --"
  if ! echo "$toc" | grep -i 'EXTENSION - ' | sed -n '1,200p'; then
    echo "(none)"
  fi
}
