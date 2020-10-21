
compose-up:
	@echo "Starting Docker Compose..."
	DOCKER_BUILDKIT=1 COMPOSE_DOCKER_CLI_BUILD=1 docker-compose up -d --build

compose-rm:
	@echo "Stopping Docker Compose..."
	docker-compose stop && docker-compose rm

clean:
	@echo "Clean Container in cb-dragonfly module..."
	docker ps -aqf name="^cb-dragonfly" | xargs -I {} docker rm -f {}
