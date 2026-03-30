local-docker-start:
	docker compose up -d

local-docker-build-and-start:
	docker compose up --build -d

local-docker-stop:
	docker compose down

local-docker-restart: local-docker-stop local-docker-start
local-docker-restart-and-build: local-docker-stop local-docker-build-and-start
