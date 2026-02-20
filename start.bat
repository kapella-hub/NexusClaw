@echo off
echo Starting NexusClaw stack...

docker compose -f deployments\docker-compose.yaml up -d

echo Waiting for PostgreSQL to be ready...
:loop
docker compose -f deployments\docker-compose.yaml exec -T postgres pg_isready -U nexusclaw >nul 2>&1
if errorlevel 1 (
  timeout /t 2 /nobreak >nul
  goto loop
)

echo Running migrations...
for %%f in (internal\platform\database\migrations\*.up.sql) do (
    echo Applying %%~nxf...
    docker compose -f deployments\docker-compose.yaml exec -T postgres psql -U nexusclaw -d nexusclaw < "%%f"
)

echo Migrations complete.
echo Done! NexusClaw is running.
echo API is at http://localhost:8080
echo Web UI is at http://localhost:3000
