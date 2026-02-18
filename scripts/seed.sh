#!/usr/bin/env bash
set -euo pipefail

DB_DSN="${NEXUSCLAW_DATABASE_DSN:-postgres://nexusclaw:nexusclaw@localhost:5432/nexusclaw?sslmode=disable}"

echo "Seeding database..."
# TODO: Add seed data
echo "Seeding complete."
