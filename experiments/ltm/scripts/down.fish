#!/usr/bin/env fish
# Stop the full LTM lab stack.
set compose (string join "" (status dirname) "/../docker-compose.yml")

docker compose -f $compose down
