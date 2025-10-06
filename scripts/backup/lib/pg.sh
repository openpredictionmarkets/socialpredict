# scripts/backup/lib/pg.sh
#
# Postgres helpers and command builders

# psql wrapper; tries DB, then postgres, then template1
psql_try() {
  local db="$1"; shift
  docker exec -e PGPASSWORD="$POSTGRES_PASSWORD" -i "$POSTGRES_CONTAINER_NAME" \
    psql -v ON_ERROR_STOP=1 -h localhost -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$db" "$@"
}

psql_smart() {
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
  local db="$1"
  echo "PGPASSWORD='${POSTGRES_PASSWORD}' pg_dump -U '${POSTGRES_USER}' -h 'localhost' -p '${POSTGRES_PORT}' -d '${db}' -Fc"
}

pg_restore_cmd() {
  local target_db="$1"; shift || true
  local flags=( --clean --if-exists )
  if [ "${PRESERVE_OWNERS:-0}" != "1" ]; then
    flags+=( --no-owner --no-privileges )
  fi
  if [ -n "${RESTORE_JOBS:-}" ]; then
    flags+=( -j "$RESTORE_JOBS" )
  fi
  echo "PGPASSWORD='${POSTGRES_PASSWORD}' pg_restore -U '${POSTGRES_USER}' -h 'localhost' -p '${POSTGRES_PORT}' -d '${target_db}' ${flags[*]}"
}

ensure_db_exists() {
  local db="$1"
  if psql_smart -tAc "SELECT 1 FROM pg_database WHERE datname='${db}'" | grep -q 1; then
    return
  fi
  echo "Target database '${db}' not found. Creating it (OWNER=${POSTGRES_USER})..."
  psql_smart -c "CREATE DATABASE \"${db}\" OWNER \"${POSTGRES_USER}\";"
}
