.PHONY: local-init local-docker-start local-docker-build-and-start local-docker-post-start \
        local-complete-first-start local-docker-stop local-docker-restart local-docker-restart-and-build \
        docker-start docker-start-build docker-stop docker-restart docker-pull-model docker-up \
        test-unit test-integration test test-coverage help

help:
	@echo "Available commands:"
	@echo "  make local-complete-first-start  - Init .env, start all Docker services and pull the models"
	@echo "  make docker-start                - Start Docker compose services in background"
	@echo "  make docker-stop                 - Stop Docker compose services"
	@echo "  make docker-pull-model           - Pull llama3 and nomic-embed-text into Ollama container"
	@echo "  make test-unit                   - Run unit tests (with //go:build unit)"
	@echo "  make test-integration            - Run integration tests (rebuilds the API and runs //go:build integration against Docker)"
	@echo "  make test                        - Run all tests (unit + integration)"
	@echo "  make test-coverage               - Run unit tests with coverage report"

local-init:
	cp -n .env.example .env 2>/dev/null || true

local-docker-start: docker-start
local-docker-build-and-start: docker-start-build
local-docker-post-start: docker-pull-model
local-complete-first-start: local-init docker-start docker-pull-model
local-docker-stop: docker-stop
local-docker-restart: docker-restart
local-docker-restart-and-build: docker-stop docker-start-build

docker-start: local-init
	docker compose up -d

docker-start-build: local-init
	docker compose up --build -d

docker-pull-model:
	docker exec ollama ollama pull llama3
	docker exec ollama ollama pull nomic-embed-text

docker-up: local-init docker-start docker-pull-model

docker-stop:
	docker compose down

docker-restart: docker-stop docker-start

test-unit:
	go test -v -tags=unit ./...

test-integration: docker-start-build
	go test -v -tags=integration ./...

test: docker-start-build
	go test -v -tags="unit,integration" ./...

test-coverage:
	go test -tags=unit -covermode=count -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
