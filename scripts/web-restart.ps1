# Recreate the web container and ensure redis is up.
docker compose -f docker/docker-compose.yml up --build -d --force-recreate web redis
docker image prune -f
