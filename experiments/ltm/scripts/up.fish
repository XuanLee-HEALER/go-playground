#!/usr/bin/env fish
# Start the full LTM lab stack with Docker Compose.
set compose (string join "" (status dirname) "/../docker-compose.yml")

docker compose -f $compose up -d --build
