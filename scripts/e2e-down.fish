#!/usr/bin/env fish
# Tear down the app container after tests.
docker compose -f docker/docker-compose.yml down
