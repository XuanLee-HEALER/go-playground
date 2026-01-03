#!/usr/bin/env fish
# Build and run the app container for local e2e tests.
docker compose -f docker/docker-compose.yml up --build -d
