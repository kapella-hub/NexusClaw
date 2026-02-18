#!/usr/bin/env bash
set -euo pipefail

MIGRATIONS_DIR="internal/platform/database/migrations"
DB_DSN="${NEXUSCLAW_DATABASE_DSN:-postgres://nexusclaw:nexusclaw@localhost:5432/nexusclaw?sslmode=disable}"

echo "Running migrations from ${MIGRATIONS_DIR}..."
for f in "${MIGRATIONS_DIR}"/*.up.sql; do
    echo "Applying $(basename "$f")..."
    psql "${DB_DSN}" -f "$f"
done
echo "Migrations complete."
