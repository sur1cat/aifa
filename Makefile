.PHONY: dev test lint build migrate-up migrate-down migrate-create clean help

# Variables
BINARY_NAME=habitflow-api
DOCKER_COMPOSE=docker-compose -f deploy/docker-compose.yml
MIGRATE=migrate -path backend/migrations -database "$(DATABASE_URL)"

# Default database URL for local development
DATABASE_URL ?= postgres://habitflow:habitflow@localhost:5432/habitflow?sslmode=disable

## dev: Start development environment (API + Postgres)
dev:
	$(DOCKER_COMPOSE) up --build

## dev-bg: Start development environment in background
dev-bg:
	$(DOCKER_COMPOSE) up -d --build

## stop: Stop development environment
stop:
	$(DOCKER_COMPOSE) down

## logs: View API logs
logs:
	$(DOCKER_COMPOSE) logs -f api

## test: Run all tests
test:
	cd backend && go test -v -race -cover ./...

## test-short: Run tests without race detector
test-short:
	cd backend && go test -v -cover ./...

## lint: Run linters
lint:
	cd backend && golangci-lint run ./...

## build: Build production binary
build:
	cd backend && CGO_ENABLED=0 GOOS=linux go build -o bin/$(BINARY_NAME) ./cmd/api

## build-local: Build for local OS
build-local:
	cd backend && go build -o bin/$(BINARY_NAME) ./cmd/api

## migrate-up: Apply all pending migrations
migrate-up:
	$(MIGRATE) up

## migrate-down: Rollback last migration
migrate-down:
	$(MIGRATE) down 1

## migrate-create: Create new migration (usage: make migrate-create name=add_users)
migrate-create:
	$(MIGRATE) create -ext sql -dir backend/migrations -seq $(name)

## migrate-force: Force migration version (usage: make migrate-force version=1)
migrate-force:
	$(MIGRATE) force $(version)

## db-shell: Connect to database shell
db-shell:
	psql $(DATABASE_URL)

## clean: Clean build artifacts
clean:
	rm -rf backend/bin
	$(DOCKER_COMPOSE) down -v --remove-orphans

## deps: Download Go dependencies
deps:
	cd backend && go mod download && go mod tidy

## generate: Run go generate
generate:
	cd backend && go generate ./...

## docker-build: Build Docker image
docker-build:
	docker build -t habitflow-api:latest -f backend/Dockerfile backend/

## help: Show this help
help:
	@echo "HabitFlow Development Commands"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@grep -E '^## ' Makefile | sed 's/## //' | column -t -s ':'
