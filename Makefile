up:
	@echo Starting Docker images...
	POSTGRES_USER=${POSTGRES_USER} POSTGRES_PASSWORD=${POSTGRES_PASSWORD} docker compose up -d
	@echo Docker images started!

down:
	@echo Stopping docker compose...
	docker-compose down
	@echo Done!

build:
	@echo Building Docker images...
	docker compose up --build
	@echo Done!
	