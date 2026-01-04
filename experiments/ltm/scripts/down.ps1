# Stop the full LTM lab stack.
$compose = Join-Path $PSScriptRoot "..\docker-compose.yml"

docker compose -f $compose down
