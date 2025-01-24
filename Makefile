up:
	@echo Starting Docker images...
	POSTGRES_USER=userUser123 POSTGRES_PASSWORD=passwordPsw123 docker compose up -d
	@echo Docker images started!

down:
	@echo Stopping docker compose...
	docker-compose down
	@echo Done!