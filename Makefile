.PHONY: run build test test-cover test-integration lint clean migrate-up migrate-down migrate-create docker-build docker-up docker-down docker-logs

APP_NAME := aifa-api
BUILD_DIR := bin

# Development
run:
	go run ./cmd/api

# Build
build:
	CGO_ENABLED=0 go build -ldflags="-w -s" -o $(BUILD_DIR)/$(APP_NAME) ./cmd/api

# Testing
test:
	go test -v -race ./...

test-cover:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

test-integration:
	go test -v -race -tags=integration ./...

# Linting
lint:
	golangci-lint run ./...

# Cleanup
clean:
	rm -rf $(BUILD_DIR) coverage.out coverage.html

# Docker
docker-build:
	docker build -t $(APP_NAME) .

docker-up:
	docker compose up -d

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f

# Database migrations (requires golang-migrate CLI)
migrate-up:
	migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path migrations -database "$(DATABASE_URL)" down 1

migrate-create:
	migrate create -ext sql -dir migrations -seq $(name)
