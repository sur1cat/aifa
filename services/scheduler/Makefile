.PHONY: run test build docker docker-run tidy fmt vet clean

BIN := bin/scheduler-worker
IMAGE := aifa/scheduler-worker:dev

run:
	go run ./cmd/worker

test:
	go test -race -count=1 ./...

build:
	CGO_ENABLED=0 go build -ldflags="-w -s" -trimpath -o $(BIN) ./cmd/worker

docker:
	docker build -t $(IMAGE) .

docker-run: docker
	docker run --rm --env-file .env -p 8009:8080 $(IMAGE)

tidy:
	go mod tidy

fmt:
	gofmt -s -w .

vet:
	go vet ./...

clean:
	rm -rf bin
