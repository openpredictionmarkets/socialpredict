# scripts/backup/lib/actions_users.sh
#
# Users-only CSV operations with FK-safe UPSERT

# --- Actions: Users CSV -------------------------------------------------------
do_save_users_common() {
  local label="$1" sql_fn="$2"
  need_container_running
  local ts file tmpfile checksum
  ts="$(timestamp)"
  file="$BACKUP_ROOT/${label}_${APP_ENV}_${ts}.csv.gz"
  tmpfile="${file}.partial"
  echo "Creating users CSV backup: $file"

  # run COPY (SELECT...) TO STDOUT WITH CSV HEADER; gzip on host
  local copy_sql; copy_sql="$($sql_fn)"
  if ! docker exec -e PGPASSWORD="$POSTGRES_PASSWORD" -i "$POSTGRES_CONTAINER_NAME" \
       psql -v ON_ERROR_STOP=1 -h localhost -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DATABASE" -c \
       "\COPY (${copy_sql}) TO STDOUT WITH CSV HEADER" | gzip -c > "$tmpfile"; then
    echo "ERROR: Users CSV export failed." ; rm -f "$tmpfile"; exit 1
  fi

  checksum="$(sha256_file "$tmpfile")"
  echo "$checksum  $(basename "$file")" > "${file}.sha256"
  mv "$tmpfile" "$file"
  echo "Users CSV backup complete:"
  echo "  File: $file"
  echo "  SHA256: $checksum"
}

do_save_users() {
  do_save_users_common "socialpredict_users" users_copy_sql
}
do_save_users_reset() {
  echo "WARNING: This CSV backup will set account_balance to initial_account_balance."
  read -r -p "Type 'RESET' to proceed: " a; [ "$a" = "RESET" ] || { echo "Aborted."; exit 1; }
  do_save_users_common "socialpredict_users_reset" users_copy_reset_sql
}

do_list_users() {
  echo "User CSV backups in $BACKUP_ROOT (newest first):"; echo
  echo "  Current-balance CSVs:"
  ls -1t "$BACKUP_ROOT"/socialpredict_users_"${APP_ENV}"_*.csv.gz 2>/dev/null || echo "    (none found)"
  echo
  echo "  Reset-balance CSVs:"
  ls -1t "$BACKUP_ROOT"/socialpredict_users_reset_"${APP_ENV}"_*.csv.gz 2>/dev/null || echo "    (none found)"
}

do_latest_users() {
  local latest; latest="$(latest_users_backup_file)"
  [ -z "$latest" ] && { echo "(none)"; return; }
  echo "$latest"
}

confirm_users_restore() {
  local file="$1"
  echo "WARNING: About to RESTORE USERS (MERGE/UPSERT by username) into database '${POSTGRES_DATABASE}'."
  echo "This will overwrite matching users' fields (balances included). Non-listed users remain."
  echo
  echo "Backup file: $file"
  echo
  read -r -p "Type 'RESTORE USERS' to proceed: " answer
  [ "$answer" = "RESTORE USERS" ] || { echo "Aborted."; exit 1; }
}

do_restore_users() {
  local file="$1"
  [ -f "$file" ] || { echo "ERROR: Users backup file not found: $file"; exit 1; }
  need_container_running
  confirm_users_restore "$file"

  local sumfile="${file}.sha256"
  if [ -f "$sumfile" ]; then
    echo "Verifying checksum..."
    local expected actual
    expected="$(awk '{print $1}' "$sumfile")"
    actual="$(sha256_file "$file")"
    [ "$expected" = "$actual" ] || { echo "ERROR: Checksum mismatch!"; exit 1; }
    echo "Checksum OK."
  fi

  if [[ "$file" == *.dump.gz ]]; then
    echo "ERROR: This tool now expects users backups as CSV (*.csv.gz)."
    echo "Old *.dump.gz users backups used raw INSERTs and can conflict with existing IDs/uniques."
    echo "Create a new CSV users backup via: ./SocialPredict backup --save-users"
    exit 1
  fi

  # Build the UPSERT SQL once. We do not delete from users; we only MERGE by username.
  # NOTE: We intentionally exclude 'id' to avoid PK/sequence conflicts.
  read -r -d '' UPSERT_SQL <<'EOSQL'
BEGIN;

CREATE TEMP TABLE users_stage (
  username TEXT PRIMARY KEY,
  display_name TEXT,
  user_type TEXT,
  email TEXT,
  password TEXT,
  initial_account_balance BIGINT,
  account_balance BIGINT,
  personal_emoji TEXT,
  description TEXT,
  personal_link1 TEXT,
  personal_link2 TEXT,
  personal_link3 TEXT,
  personal_link4 TEXT,
  created_at TIMESTAMPTZ,
  updated_at TIMESTAMPTZ
) ON COMMIT DROP;

-- COPY will come from STDIN next

-- Upsert into real users table by username
INSERT INTO users (
  username, display_name, user_type, email, password,
  initial_account_balance, account_balance,
  personal_emoji, description, personal_link1, personal_link2, personal_link3, personal_link4,
  created_at, updated_at
)
SELECT
  s.username, s.display_name, s.user_type, s.email, s.password,
  s.initial_account_balance, s.account_balance,
  s.personal_emoji, s.description, s.personal_link1, s.personal_link2, s.personal_link3, s.personal_link4,
  s.created_at, s.updated_at
FROM users_stage s
ON CONFLICT (username) DO UPDATE SET
  display_name = EXCLUDED.display_name,
  user_type    = EXCLUDED.user_type,
  email       = EXCLUDED.email,
  password    = EXCLUDED.password,
  initial_account_balance = EXCLUDED.initial_account_balance,
  account_balance        = EXCLUDED.account_balance,
  personal_emoji = EXCLUDED.personal_emoji,
  description   = EXCLUDED.description,
  personal_link1= EXCLUDED.personal_link1,
  personal_link2= EXCLUDED.personal_link2,
  personal_link3= EXCLUDED.personal_link3,
  personal_link4= EXCLUDED.personal_link4,
  created_at    = LEAST(users.created_at, EXCLUDED.created_at),
  updated_at    = GREATEST(users.updated_at, EXCLUDED.updated_at);

COMMIT;
EOSQL

  # Open one docker exec psql session and feed:
  #  1) BEGIN + temp table
  #  2) \COPY users_stage FROM STDIN CSV HEADER  (the CSV we gunzip)
  #  3) The INSERT ... ON CONFLICT ... MERGE + COMMIT
  {
    echo "BEGIN;"
    echo "$UPSERT_SQL" | sed '1,/CREATE TEMP TABLE users_stage/d' | sed '/-- COPY will come from STDIN next/,$d'
    echo "\COPY users_stage ($USERS_CSV_COLUMNS) FROM STDIN WITH CSV HEADER"
    gunzip -c "$file"
    echo "\."
    echo "$UPSERT_SQL" | sed '1,/-- Upsert into real users table by username/d'
  } | docker exec -e PGPASSWORD="$POSTGRES_PASSWORD" -i "$POSTGRES_CONTAINER_NAME" \
        psql -v ON_ERROR_STOP=1 -h localhost -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DATABASE"

  echo "Users restore (MERGE/UPSERT) complete."
}

do_restore_users_latest() {
  local latest; latest="$(latest_users_backup_file)"
  [ -z "$latest" ] && { echo "ERROR: No users backups found in $BACKUP_ROOT"; exit 1; }
  do_restore_users "$latest"
}
