#!/usr/bin/env fish
# Build and run the web frontend with redis.
docker compose -f docker/docker-compose.yml up --build -d web redis
