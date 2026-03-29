.PHONY: generate run test lint migrate-up migrate-down seed

generate:
	go run github.com/99designs/gqlgen generate

run:
	go run cmd/server/main.go

test:
	go test ./... -v -race

lint:
	golangci-lint run ./...

migrate-up:
	go run -tags migrate cmd/migrate/main.go up

migrate-down:
	go run -tags migrate cmd/migrate/main.go down

seed:
	go run cmd/seed/main.go

docker-up:
	docker compose up -d

docker-down:
	docker compose down
