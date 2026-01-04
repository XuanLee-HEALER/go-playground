# Start the full LTM lab stack with Docker Compose.
$compose = Join-Path $PSScriptRoot "..\docker-compose.yml"

docker compose -f $compose up -d --build
