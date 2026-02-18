.PHONY: build run test lint migrate docker-up docker-down ci clean

BINARY=bin/nexusclaw
MAIN=./cmd/nexusclaw

build:
	go build -o $(BINARY) $(MAIN)

run:
	go run $(MAIN) serve

test:
	go test ./... -race -count=1

lint:
	golangci-lint run ./...

migrate:
	bash scripts/migrate.sh

docker-up:
	docker-compose -f deployments/docker-compose.yaml up -d

docker-down:
	docker-compose -f deployments/docker-compose.yaml down

ci: lint test build

clean:
	rm -rf bin/
