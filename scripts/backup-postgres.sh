#!/usr/bin/env sh
set -e

timestamp="$(date +%Y%m%d-%H%M%S)"
backup_dir="backups"
mkdir -p "$backup_dir"

db_user="${DB_USER:-postgres}"
db_name="${DB_NAME:-imp101}"
output_file="$backup_dir/imp101-$timestamp.sql"

docker exec -e PGPASSWORD="${DB_PASSWORD:-postgres}" imp101-postgres \
  pg_dump -U "$db_user" "$db_name" > "$output_file"

echo "Backup created at $output_file"
