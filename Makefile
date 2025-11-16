BINARY_NAME := pr-service
CMD_PATH    := ./cmd/main.go
DOCKER_IMAGE := pr-service:local

.PHONY: build run lint test test-integration docker-build docker-up docker-down docker-logs migrate-up loadtest

build:
	mkdir -p bin
	go build -o ./bin/$(BINARY_NAME) $(CMD_PATH)

run: docker-up

lint:
	golangci-lint run ./...

test:
	go test ./...

test-integration:
	@echo "Running integration tests with local Postgres on localhost:5432..."
	TEST_DATABASE_URL="postgres://postgres:password@localhost:5432/postgres?sslmode=disable" \
	  go test ./internal/domain/repository/postgres -tags=integration -v

docker-build:
	docker build -t $(DOCKER_IMAGE) .

docker-up:
	docker-compose up --build

docker-down:
	docker-compose down -v

docker-logs:
	docker-compose logs -f pr-service pr-db

migrate-up:
	docker-compose run --rm pr-migrate

loadtest:
	k6 run -e BASE_URL=http://localhost:8080 loadtest/k6_pr_service.js
