@echo off
echo Stopping NexusClaw stack...
docker compose -f deployments\docker-compose.yaml down
echo Done.
