#!/usr/bin/env bash
set -e

echo "Starting NexusClaw stack..."
docker compose -f deployments/docker-compose.yaml up -d

echo "Waiting for PostgreSQL to be ready..."
until docker compose -f deployments/docker-compose.yaml exec -T postgres pg_isready -U nexusclaw > /dev/null 2>&1; do
  sleep 2
done

echo "Running migrations..."
for f in internal/platform/database/migrations/*.up.sql; do
    echo "Applying $(basename "$f")..."
    docker compose -f deployments/docker-compose.yaml exec -T postgres psql -U nexusclaw -d nexusclaw < "$f"
done

echo "Migrations complete."
echo "Done! NexusClaw is running."
echo "API is at http://localhost:8080"
echo "Web UI is at http://localhost:3000"
