#!/usr/bin/env fish
# Stop and remove only the web frontend container.
docker compose -f docker/docker-compose.yml stop web
docker compose -f docker/docker-compose.yml rm -f web
