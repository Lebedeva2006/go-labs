.PHONY: migrate run generate build test lint

migrate:
	go tool goose -dir migrations postgres "$$DATABASE_URL" up

migrate-down:
	go tool goose -dir migrations postgres "$$DATABASE_URL" down

run:
	go run ./cmd/trip-service

generate:
	go tool oapi-codegen \
	 -generate types,chi-server \
	 -include-operation-ids createTrip,getTrip,finishTrip,health,ready \
	 -package api \
	 -o internal/generated/api.gen.go \
 	 contracts/openapi/trip-service.openapi.yaml

build:
	go build -o bin/trip-service ./cmd/trip-service

test:
	go test -race ./...
