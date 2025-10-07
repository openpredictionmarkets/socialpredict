# scripts/backup/lib/users_csv.sh
#
# Users CSV schema and SELECT builders

# Users CSV column list (NO id: avoids PK collisions & sequence headaches)
USERS_CSV_COLUMNS="username,display_name,user_type,email,password,initial_account_balance,account_balance,personal_emoji,description,personal_link1,personal_link2,personal_link3,personal_link4,created_at,updated_at"

# Build COPY SELECT for users (current balances)
users_copy_sql() {
  cat <<'SQL'
SELECT
  username,
  display_name,
  user_type,
  email,
  password,
  initial_account_balance,
  account_balance,
  personal_emoji,
  description,
  personal_link1,
  personal_link2,
  personal_link3,
  personal_link4,
  created_at,
  updated_at
FROM users
ORDER BY username
SQL
}

users_copy_reset_sql() {
  cat <<'SQL'
SELECT
  username,
  display_name,
  user_type,
  email,
  password,
  initial_account_balance,
  initial_account_balance AS account_balance,
  personal_emoji,
  description,
  personal_link1,
  personal_link2,
  personal_link3,
  personal_link4,
  created_at,
  updated_at
FROM users
ORDER BY username
SQL
}

latest_backup_file() {
  ls -1t "$BACKUP_ROOT"/socialpredict_backup_"${APP_ENV}"_*.dump.gz 2>/dev/null | head -n 1 || true
}

latest_users_backup_file() {
  ls -1t "$BACKUP_ROOT"/socialpredict_users_"${APP_ENV}"_*.csv.gz 2>/dev/null | head -n 1 || true
}
