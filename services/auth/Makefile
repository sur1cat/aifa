.PHONY: run test build docker docker-run tidy fmt vet lint clean

BIN := bin/auth-service
IMAGE := aifa/auth-service:dev

run:
	go run ./cmd/server

test:
	go test -race -count=1 ./...

build:
	CGO_ENABLED=0 go build -ldflags="-w -s" -trimpath -o $(BIN) ./cmd/server

docker:
	docker build -t $(IMAGE) .

docker-run: docker
	docker run --rm --env-file .env -p 8001:8080 $(IMAGE)

tidy:
	go mod tidy

fmt:
	gofmt -s -w .

vet:
	go vet ./...

clean:
	rm -rf bin
